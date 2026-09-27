package engine

import (
	"math"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/numeric"
)

const (
	hoursPerDay          = 24
	minHourlyStdRatio    = 0.03
	minVoltageStdDev     = 0.5
	minCurrentStdDev     = 1
	minPowerFactorStdDev = 0.005
	minPowerRatioStdDev  = 0.01
)

type Stat struct {
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
}

func robustStat(values []float64, minStdDev float64) Stat {
	return Stat{Mean: numeric.Median(values), Std: math.Max(numeric.RobustStdDev(values), minStdDev)}
}

type Baseline struct {
	MeterID       string               `json:"meter_id"`
	From          time.Time            `json:"from"`
	To            time.Time            `json:"to"`
	HourlyKWh     [hoursPerDay]float64 `json:"hourly_kwh"`
	HourlyStdDev  [hoursPerDay]float64 `json:"hourly_std"`
	HourlyCurrent [hoursPerDay]float64 `json:"hourly_current"`
	DailyKWh      float64              `json:"daily_kwh"`
	Voltage       Stat                 `json:"voltage"`
	Current       Stat                 `json:"current"`
	PowerFactor   Stat                 `json:"power_factor"`
	PowerRatio    Stat                 `json:"power_ratio"`
	Samples       int                  `json:"samples"`
}

func BuildBaseline(meterID string, readings []domain.Reading, from, to time.Time) Baseline {
	baseline := Baseline{MeterID: meterID, From: from, To: to}
	var consumptionByHour, currentByHour [hoursPerDay][]float64
	var voltages, currents, powerFactors, powerRatios []float64

	for _, reading := range readings {
		if reading.Timestamp.Before(from) || !reading.Timestamp.Before(to) {
			continue
		}
		hour := reading.Timestamp.Hour()
		consumptionByHour[hour] = append(consumptionByHour[hour], reading.ConsumptionKWh)
		currentByHour[hour] = append(currentByHour[hour], reading.CurrentA)
		voltages = append(voltages, reading.VoltageV)
		currents = append(currents, reading.CurrentA)
		powerFactors = append(powerFactors, reading.PowerFactor)
		if ratio := reading.PowerRatio(); !math.IsNaN(ratio) {
			powerRatios = append(powerRatios, ratio)
		}
		baseline.Samples++
	}

	for hour := range hoursPerDay {
		expected := numeric.Median(consumptionByHour[hour])
		baseline.HourlyKWh[hour] = expected
		baseline.HourlyStdDev[hour] = math.Max(numeric.RobustStdDev(consumptionByHour[hour]), minHourlyStdRatio*expected)
		baseline.HourlyCurrent[hour] = numeric.Median(currentByHour[hour])
		baseline.DailyKWh += expected
	}
	baseline.Voltage = robustStat(voltages, minVoltageStdDev)
	baseline.Current = robustStat(currents, minCurrentStdDev)
	baseline.PowerFactor = robustStat(powerFactors, minPowerFactorStdDev)
	baseline.PowerRatio = robustStat(powerRatios, minPowerRatioStdDev)
	return baseline
}

func (b Baseline) ExpectedConsumption(at time.Time) float64 {
	return b.HourlyKWh[at.Hour()]
}

func (b Baseline) ConsumptionStdDev(at time.Time) float64 {
	return b.HourlyStdDev[at.Hour()]
}

func (b Baseline) ExpectedCurrent(at time.Time) float64 {
	return b.HourlyCurrent[at.Hour()]
}
