package engine

import "time"

type Config struct {
	BaselineDays       int
	DeviationThreshold float64
	MinTransientHours  int
	MinPersistentHours int
	OutlierZScore      float64
	EventWindow        time.Duration
}

func DefaultConfig() Config {
	return Config{
		BaselineDays:       7,
		DeviationThreshold: 0.25,
		MinTransientHours:  3,
		MinPersistentHours: 24,
		OutlierZScore:      4,
		EventWindow:        12 * time.Hour,
	}
}
