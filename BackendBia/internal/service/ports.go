package service

import (
	"time"

	"github.com/biaenergy/backend/internal/domain"
)

type MeterRepository interface {
	ListMeters() []domain.Meter
	FindMeter(meterID string) (domain.Meter, error)
}

type ReadingRepository interface {
	ListReadings(meterID string, from, to time.Time) []domain.Reading
	ReadingsByMeter() map[string][]domain.Reading
}

type EventRepository interface {
	ListEvents(meterID string) []domain.Event
}

type AnomalyRepository interface {
	ListAnomalies() []domain.Anomaly
	FindAnomaly(id string) (domain.Anomaly, error)
	UpdateAnomaly(id string, status domain.AnomalyStatus, note *domain.Note) (domain.Anomaly, error)
}

type AnalysisRepository interface {
	SaveRun(run domain.AnalysisRun) error
	FindRun(id string) (domain.AnalysisRun, error)
	LatestRun() (domain.AnalysisRun, bool)
	ListRuns() []domain.AnalysisRun
	PublishResults(run domain.AnalysisRun, anomalies []domain.Anomaly, statuses map[string]domain.MeterStatus) error
}

type Repository interface {
	MeterRepository
	ReadingRepository
	EventRepository
	AnomalyRepository
	AnalysisRepository
}
