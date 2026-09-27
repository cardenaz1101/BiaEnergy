package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/biaenergy/backend/internal/domain"
)

type snapshot struct {
	Runs          []domain.AnalysisRun          `json:"runs"`
	Anomalies     []domain.Anomaly              `json:"anomalies"`
	MeterStatuses map[string]domain.MeterStatus `json:"meter_status"`
}

type Store struct {
	mu        sync.RWMutex
	statePath string
	meters    map[string]*domain.Meter
	readings  map[string][]domain.Reading
	events    []domain.Event
	runs      []domain.AnalysisRun
	anomalies []domain.Anomaly
}

func NewStore(readings []domain.Reading, events []domain.Event, catalog map[string]MeterInfo, statePath string) (*Store, error) {
	store := &Store{
		statePath: statePath,
		meters:    map[string]*domain.Meter{},
		readings:  map[string][]domain.Reading{},
		events:    events,
	}
	createdAt := time.Now().UTC()
	for _, reading := range readings {
		store.readings[reading.MeterID] = append(store.readings[reading.MeterID], reading)
		if _, exists := store.meters[reading.MeterID]; exists {
			continue
		}
		info, ok := catalog[reading.MeterID]
		if !ok {
			info = MeterInfo{Name: "Medidor " + reading.MeterID, Location: "Sin ubicación"}
		}
		store.meters[reading.MeterID] = &domain.Meter{
			ID:        reading.MeterID,
			MeterID:   reading.MeterID,
			Name:      info.Name,
			Location:  info.Location,
			Status:    domain.MeterPending,
			CreatedAt: createdAt,
		}
	}
	if err := store.restore(); err != nil {
		return nil, fmt.Errorf("restaurando estado: %w", err)
	}
	return store, nil
}

func (s *Store) restore() error {
	if s.statePath == "" {
		return nil
	}
	content, err := os.ReadFile(s.statePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var saved snapshot
	if err := json.Unmarshal(content, &saved); err != nil {
		return err
	}
	s.runs = saved.Runs
	s.anomalies = saved.Anomalies
	for meterID, status := range saved.MeterStatuses {
		if meter, ok := s.meters[meterID]; ok {
			meter.Status = status
		}
	}
	for i := range s.runs {
		if s.runs[i].Status == domain.RunRunning {
			s.runs[i].Status = domain.RunFailed
			s.runs[i].Error = "interrumpido por reinicio del servidor"
		}
	}
	return nil
}

func (s *Store) persist() error {
	if s.statePath == "" {
		return nil
	}
	saved := snapshot{Runs: s.runs, Anomalies: s.anomalies, MeterStatuses: map[string]domain.MeterStatus{}}
	for meterID, meter := range s.meters {
		saved.MeterStatuses[meterID] = meter.Status
	}
	content, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.statePath), 0o755); err != nil {
		return err
	}
	tempPath := s.statePath + ".tmp"
	if err := os.WriteFile(tempPath, content, 0o644); err != nil {
		return err
	}
	return os.Rename(tempPath, s.statePath)
}

func (s *Store) ListMeters() []domain.Meter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meters := make([]domain.Meter, 0, len(s.meters))
	for _, meter := range s.meters {
		meters = append(meters, *meter)
	}
	sort.Slice(meters, func(i, j int) bool { return meters[i].MeterID < meters[j].MeterID })
	return meters
}

func (s *Store) FindMeter(meterID string) (domain.Meter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meter, ok := s.meters[meterID]
	if !ok {
		return domain.Meter{}, domain.ErrNotFound
	}
	return *meter, nil
}

func (s *Store) ListReadings(meterID string, from, to time.Time) []domain.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.readings[meterID]
	readings := make([]domain.Reading, 0, len(all))
	for _, reading := range all {
		if !from.IsZero() && reading.Timestamp.Before(from) {
			continue
		}
		if !to.IsZero() && !reading.Timestamp.Before(to) {
			continue
		}
		readings = append(readings, reading)
	}
	return readings
}

func (s *Store) ReadingsByMeter() map[string][]domain.Reading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make(map[string][]domain.Reading, len(s.readings))
	for meterID, readings := range s.readings {
		copied[meterID] = append([]domain.Reading(nil), readings...)
	}
	return copied
}

func (s *Store) ListEvents(meterID string) []domain.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := []domain.Event{}
	for _, event := range s.events {
		if meterID == "" || event.MeterID == meterID {
			events = append(events, event)
		}
	}
	return events
}

func (s *Store) ListAnomalies() []domain.Anomaly {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Anomaly{}, s.anomalies...)
}

func (s *Store) FindAnomaly(id string) (domain.Anomaly, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, anomaly := range s.anomalies {
		if anomaly.ID == id {
			return anomaly, nil
		}
	}
	return domain.Anomaly{}, domain.ErrNotFound
}

func (s *Store) UpdateAnomaly(id string, status domain.AnomalyStatus, note *domain.Note) (domain.Anomaly, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.anomalies {
		if s.anomalies[i].ID != id {
			continue
		}
		if status != "" {
			s.anomalies[i].Status = status
		}
		if note != nil {
			s.anomalies[i].Notes = append(s.anomalies[i].Notes, *note)
		}
		return s.anomalies[i], s.persist()
	}
	return domain.Anomaly{}, domain.ErrNotFound
}

func (s *Store) SaveRun(run domain.AnalysisRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upsertRun(run)
	return s.persist()
}

func (s *Store) upsertRun(run domain.AnalysisRun) {
	for i := range s.runs {
		if s.runs[i].ID == run.ID {
			s.runs[i] = run
			return
		}
	}
	s.runs = append(s.runs, run)
}

func (s *Store) FindRun(id string) (domain.AnalysisRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, run := range s.runs {
		if run.ID == id {
			return run, nil
		}
	}
	return domain.AnalysisRun{}, domain.ErrNotFound
}

func (s *Store) LatestRun() (domain.AnalysisRun, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.runs) == 0 {
		return domain.AnalysisRun{}, false
	}
	return s.runs[len(s.runs)-1], true
}

func (s *Store) ListRuns() []domain.AnalysisRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	runs := append([]domain.AnalysisRun{}, s.runs...)
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt.After(runs[j].StartedAt) })
	return runs
}

func (s *Store) PublishResults(run domain.AnalysisRun, anomalies []domain.Anomaly, statuses map[string]domain.MeterStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := map[string]domain.Anomaly{}
	for _, anomaly := range s.anomalies {
		previous[anomaly.ContinuityKey()] = anomaly
	}
	for i := range anomalies {
		if earlier, ok := previous[anomalies[i].ContinuityKey()]; ok {
			anomalies[i].Status = earlier.Status
			anomalies[i].Notes = earlier.Notes
		}
	}
	s.anomalies = anomalies
	for meterID, meter := range s.meters {
		meter.Status = domain.MeterOK
		if status, ok := statuses[meterID]; ok {
			meter.Status = status
		}
	}
	s.upsertRun(run)
	return s.persist()
}
