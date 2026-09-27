package domain

import "time"

const (
	EventScheduledOutage   = "SCHEDULED_OUTAGE"
	EventMaintenance       = "MAINTENANCE"
	EventOperationalChange = "OPERATIONAL_CHANGE"
	EventDataQuality       = "DATA_QUALITY"
	EventUnknown           = "UNKNOWN"
)

type Event struct {
	ID          string    `json:"id"`
	MeterID     string    `json:"meter_id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
}

func FindEventByType(events []Event, types ...string) *Event {
	for i := range events {
		for _, eventType := range types {
			if events[i].Type == eventType {
				return &events[i]
			}
		}
	}
	return nil
}
