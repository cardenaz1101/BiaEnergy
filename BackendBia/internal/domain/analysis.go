package domain

import "time"

type RunStatus string

const (
	RunRunning   RunStatus = "running"
	RunCompleted RunStatus = "completed"
	RunFailed    RunStatus = "failed"
)

type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepRunning StepStatus = "running"
	StepDone    StepStatus = "done"
	StepError   StepStatus = "error"
)

type AnalysisStep struct {
	Key        string     `json:"key"`
	Label      string     `json:"label"`
	Status     StepStatus `json:"status"`
	Detail     string     `json:"detail"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type AnalysisRun struct {
	ID               string         `json:"id"`
	Status           RunStatus      `json:"status"`
	StartedAt        time.Time      `json:"started_at"`
	FinishedAt       *time.Time     `json:"finished_at,omitempty"`
	Steps            []AnalysisStep `json:"steps"`
	MetersAnalyzed   int            `json:"meters_analyzed"`
	ReadingsAnalyzed int            `json:"readings_analyzed"`
	AnomaliesFound   int            `json:"anomalies_found"`
	PriorityCount    int            `json:"priority_count"`
	AvgConfidence    float64        `json:"avg_confidence"`
	Summary          string         `json:"summary"`
	Error            string         `json:"error,omitempty"`
}
