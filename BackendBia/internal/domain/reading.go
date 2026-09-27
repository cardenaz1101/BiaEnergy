package domain

import (
	"math"
	"time"
)

type Reading struct {
	ID             int64     `json:"id"`
	MeterID        string    `json:"meter_id"`
	Timestamp      time.Time `json:"timestamp"`
	ConsumptionKWh float64   `json:"consumption_kwh"`
	VoltageV       float64   `json:"voltage_v"`
	CurrentA       float64   `json:"current_a"`
	PowerFactor    float64   `json:"power_factor"`
	Status         string    `json:"status"`
}

func (r Reading) PowerRatio() float64 {
	activePowerKW := r.VoltageV * r.CurrentA * r.PowerFactor / 1000
	if activePowerKW <= 0 {
		return math.NaN()
	}
	return r.ConsumptionKWh / activePowerKW
}

func (r Reading) IsPhysicallyInvalid() bool {
	return r.VoltageV <= 0 || r.CurrentA < 0 || r.PowerFactor < 0 || r.PowerFactor > 1 || r.ConsumptionKWh < 0
}
