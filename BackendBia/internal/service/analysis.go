package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/engine"
	"github.com/biaenergy/backend/internal/numeric"
)

const analysisTimeout = 2 * time.Minute

type AnalysisService struct {
	repository Repository
	analyzer   *engine.Engine
	stepDelay  time.Duration
	logger     *slog.Logger

	mu        sync.Mutex
	activeRun string
}

func NewAnalysisService(repository Repository, analyzer *engine.Engine, stepDelay time.Duration, logger *slog.Logger) *AnalysisService {
	return &AnalysisService{repository: repository, analyzer: analyzer, stepDelay: stepDelay, logger: logger}
}

func (s *AnalysisService) Start() (domain.AnalysisRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeRun != "" {
		if run, err := s.repository.FindRun(s.activeRun); err == nil {
			return run, domain.ErrAnalysisInProgress
		}
	}
	run := newRun()
	if err := s.repository.SaveRun(run); err != nil {
		return run, err
	}
	s.activeRun = run.ID
	go s.execute(run)
	return run, nil
}

func (s *AnalysisService) Get(id string) (domain.AnalysisRun, error) {
	return s.repository.FindRun(id)
}

func (s *AnalysisService) Latest() (domain.AnalysisRun, bool) {
	return s.repository.LatestRun()
}

func (s *AnalysisService) List() []domain.AnalysisRun {
	return s.repository.ListRuns()
}

func newRun() domain.AnalysisRun {
	now := time.Now().UTC()
	run := domain.AnalysisRun{ID: newRunID(now), Status: domain.RunRunning, StartedAt: now}
	for _, step := range engine.Steps {
		run.Steps = append(run.Steps, domain.AnalysisStep{Key: step.Key, Label: step.Label, Status: domain.StepPending})
	}
	run.Steps[0].Status = domain.StepRunning
	run.Steps[0].StartedAt = &now
	return run
}

func newRunID(now time.Time) string {
	suffix := make([]byte, 4)
	_, _ = rand.Read(suffix)
	return fmt.Sprintf("RUN-%s-%s", now.Format("20060102-150405"), hex.EncodeToString(suffix))
}

func (s *AnalysisService) execute(run domain.AnalysisRun) {
	defer s.clearActiveRun()
	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()

	nextStep := 0
	reportStep := func(stepKey, detail string) error {
		time.Sleep(s.stepDelay)
		now := time.Now().UTC()
		for i := range run.Steps {
			if run.Steps[i].Key == stepKey {
				run.Steps[i].Status, run.Steps[i].Detail, run.Steps[i].FinishedAt = domain.StepDone, detail, &now
				nextStep = i + 1
			}
		}
		if nextStep < len(run.Steps) {
			run.Steps[nextStep].Status, run.Steps[nextStep].StartedAt = domain.StepRunning, &now
		}
		return s.repository.SaveRun(run)
	}

	input := engine.Input{
		Meters:   s.repository.ListMeters(),
		Readings: s.repository.ReadingsByMeter(),
		Events:   s.repository.ListEvents(""),
	}
	run.MetersAnalyzed = len(input.Meters)
	result, err := s.analyzer.Analyze(ctx, run.ID, input, reportStep)
	finishedAt := time.Now().UTC()
	run.FinishedAt = &finishedAt

	if err != nil {
		run.Status, run.Error = domain.RunFailed, err.Error()
		if nextStep < len(run.Steps) {
			run.Steps[nextStep].Status = domain.StepError
		}
		s.logger.Error("análisis fallido", "run", run.ID, "error", err)
		if saveErr := s.repository.SaveRun(run); saveErr != nil {
			s.logger.Error("guardando análisis", "run", run.ID, "error", saveErr)
		}
		return
	}

	completeRun(&run, result)
	if err := s.repository.PublishResults(run, result.Anomalies, result.MeterStatuses); err != nil {
		s.logger.Error("publicando resultados", "run", run.ID, "error", err)
		return
	}
	s.logger.Info("análisis completado", "run", run.ID, "summary", run.Summary)
}

func completeRun(run *domain.AnalysisRun, result engine.Result) {
	run.Status = domain.RunCompleted
	run.ReadingsAnalyzed = result.ReadingsAnalyzed
	run.AnomaliesFound = len(result.Anomalies)
	confidenceSum := 0.0
	for _, anomaly := range result.Anomalies {
		confidenceSum += anomaly.Confidence
		if anomaly.IsPriority() {
			run.PriorityCount++
		}
	}
	if run.AnomaliesFound > 0 {
		run.AvgConfidence = numeric.Round(confidenceSum/float64(run.AnomaliesFound), 2)
	}
	run.Summary = pluralize(run.AnomaliesFound, "anomalía detectada", "anomalías detectadas") + " · " +
		pluralize(run.PriorityCount, "requiere atención prioritaria", "requieren atención prioritaria")
}

func (s *AnalysisService) clearActiveRun() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeRun = ""
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", count, plural)
}
