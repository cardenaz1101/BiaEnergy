package domain

import "time"

type MeterStatus string

const (
	MeterPending  MeterStatus = "PENDING"
	MeterOK       MeterStatus = "OK"
	MeterAlert    MeterStatus = "ALERT"
	MeterCritical MeterStatus = "CRITICAL"
)

func (s MeterStatus) Rank() int {
	switch s {
	case MeterCritical:
		return 3
	case MeterAlert:
		return 2
	case MeterOK:
		return 1
	default:
		return 0
	}
}

type Meter struct {
	ID        string      `json:"id"`
	MeterID   string      `json:"meter_id"`
	Name      string      `json:"name"`
	Location  string      `json:"location"`
	Status    MeterStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}
