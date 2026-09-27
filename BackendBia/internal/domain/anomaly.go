package domain

import "time"

type AnomalyType string

const (
	RealAnomaly        AnomalyType = "REAL_ANOMALY"
	ExplainableAnomaly AnomalyType = "EXPLAINABLE_ANOMALY"
	DataQuality        AnomalyType = "DATA_QUALITY"
	FalsePositive      AnomalyType = "FALSE_POSITIVE"
)

func (t AnomalyType) Code() string {
	switch t {
	case RealAnomaly:
		return "RA"
	case ExplainableAnomaly:
		return "EX"
	case DataQuality:
		return "DQ"
	case FalsePositive:
		return "FP"
	default:
		return "XX"
	}
}

type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
	SeverityLow    Severity = "LOW"
)

func (s Severity) Weight() int {
	switch s {
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

type AnomalyStatus string

const (
	AnomalyOpen          AnomalyStatus = "OPEN"
	AnomalyInvestigating AnomalyStatus = "INVESTIGATING"
	AnomalyResolved      AnomalyStatus = "RESOLVED"
	AnomalyDismissed     AnomalyStatus = "DISMISSED"
)

func (s AnomalyStatus) IsValid() bool {
	switch s {
	case AnomalyOpen, AnomalyInvestigating, AnomalyResolved, AnomalyDismissed:
		return true
	default:
		return false
	}
}

func (s AnomalyStatus) IsActive() bool {
	return s == AnomalyOpen || s == AnomalyInvestigating
}

type EvidenceStance string

const (
	StanceSupports EvidenceStance = "supports"
	StanceAgainst  EvidenceStance = "against"
	StanceContext  EvidenceStance = "context"
)

type EvidenceKind string

const (
	EvidenceBaseline    EvidenceKind = "baseline"
	EvidenceElectrical  EvidenceKind = "electrical"
	EvidenceEvent       EvidenceKind = "event"
	EvidenceDataQuality EvidenceKind = "data_quality"
)

type Evidence struct {
	Kind        EvidenceKind   `json:"kind"`
	Stance      EvidenceStance `json:"stance"`
	Description string         `json:"description"`
}

type VariableChange struct {
	Variable    string  `json:"variable"`
	Label       string  `json:"label"`
	Unit        string  `json:"unit"`
	Baseline    float64 `json:"baseline"`
	Observed    float64 `json:"observed"`
	DeltaPct    float64 `json:"delta_pct"`
	ZScore      float64 `json:"z_score"`
	Significant bool    `json:"significant"`
}

type ConfidenceFactor struct {
	Label        string  `json:"label"`
	Contribution float64 `json:"contribution"`
}

type AnomalyMetrics struct {
	BaselineDailyKWh float64 `json:"baseline_daily_kwh"`
	CurrentDailyKWh  float64 `json:"current_daily_kwh"`
	VariationPct     float64 `json:"variation_pct"`
	ExcessKWh        float64 `json:"excess_kwh"`
	AffectedReadings int     `json:"affected_readings"`
}

type Note struct {
	At     time.Time `json:"at"`
	Author string    `json:"author"`
	Text   string    `json:"text"`
}

type Anomaly struct {
	ID                  string             `json:"id"`
	RunID               string             `json:"run_id"`
	MeterID             string             `json:"meter_id"`
	DetectedAt          time.Time          `json:"detected_at"`
	IsAnomaly           bool               `json:"anomaly"`
	Type                AnomalyType        `json:"type"`
	Signal              string             `json:"signal"`
	Severity            Severity           `json:"severity"`
	Confidence          float64            `json:"confidence"`
	PriorityScore       float64            `json:"priority_score"`
	PriorityRank        int                `json:"priority_rank"`
	Title               string             `json:"title"`
	Reason              string             `json:"reason"`
	Explanation         string             `json:"explanation"`
	RecommendedAction   string             `json:"recommended_action"`
	ActionSteps         []string           `json:"action_steps"`
	Status              AnomalyStatus      `json:"status"`
	StartedAt           time.Time          `json:"started_at"`
	EndedAt             *time.Time         `json:"ended_at,omitempty"`
	DurationHours       int                `json:"duration_hours"`
	Metrics             AnomalyMetrics     `json:"metrics"`
	ChangedVariables    []VariableChange   `json:"changed_variables"`
	RelatedEvents       []Event            `json:"related_events"`
	Evidence            []Evidence         `json:"evidence"`
	ConfidenceBreakdown []ConfidenceFactor `json:"confidence_breakdown"`
	Notes               []Note             `json:"notes"`
}

func (a Anomaly) IsPriority() bool {
	return a.IsAnomaly && a.Severity == SeverityHigh
}

func (a Anomaly) MeterStatus() MeterStatus {
	switch {
	case a.Type == FalsePositive:
		return MeterOK
	case a.Type == RealAnomaly && a.Severity == SeverityHigh:
		return MeterCritical
	default:
		return MeterAlert
	}
}

func (a Anomaly) ContinuityKey() string {
	return a.MeterID + "|" + string(a.Type)
}
