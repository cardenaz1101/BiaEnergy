package service

import (
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/domain"
)

type AnomalyQuery struct {
	Type     string
	Severity string
	Status   string
	MeterID  string
}

type AnomalyUpdate struct {
	Status string
	Note   string
	Author string
}

type AnomalyService struct {
	anomalies AnomalyRepository
}

func NewAnomalyService(anomalies AnomalyRepository) *AnomalyService {
	return &AnomalyService{anomalies: anomalies}
}

func (s *AnomalyService) List(query AnomalyQuery) []domain.Anomaly {
	matches := []domain.Anomaly{}
	for _, anomaly := range s.anomalies.ListAnomalies() {
		if matchesText(query.Type, string(anomaly.Type)) &&
			matchesText(query.Severity, string(anomaly.Severity)) &&
			matchesText(query.Status, string(anomaly.Status)) &&
			matchesText(query.MeterID, anomaly.MeterID) {
			matches = append(matches, anomaly)
		}
	}
	return matches
}

func (s *AnomalyService) Get(id string) (domain.Anomaly, error) {
	return s.anomalies.FindAnomaly(id)
}

func (s *AnomalyService) Update(id string, update AnomalyUpdate) (domain.Anomaly, error) {
	status := domain.AnomalyStatus(strings.ToUpper(strings.TrimSpace(update.Status)))
	if status != "" && !status.IsValid() {
		return domain.Anomaly{}, domain.ErrInvalidStatus
	}
	var note *domain.Note
	if text := strings.TrimSpace(update.Note); text != "" {
		note = &domain.Note{At: time.Now().UTC(), Author: update.Author, Text: text}
	}
	return s.anomalies.UpdateAnomaly(id, status, note)
}

func matchesText(filter, value string) bool {
	return filter == "" || strings.EqualFold(filter, value)
}
