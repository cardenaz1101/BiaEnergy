package engine

import (
	"math"
	"sort"
	"time"

	"github.com/biaenergy/backend/internal/domain"
)

const (
	SignalPersistentShift = "PERSISTENT_SHIFT"
	SignalTransientSpike  = "TRANSIENT_SPIKE"
	SignalTransientDrop   = "TRANSIENT_DROP"
	SignalDataQuality     = "DATA_QUALITY"
)

const (
	maxGapReadings      = 1
	ongoingTailReadings = 2
)

type Deviation struct {
	Kind          string
	Start         time.Time
	End           time.Time
	Hours         int
	Ongoing       bool
	MeanDeviation float64
	PeakZScore    float64
	ExcessKWh     float64
	Readings      []domain.Reading
}

func (d Deviation) Strength() float64 {
	return math.Abs(d.MeanDeviation) * float64(d.Hours)
}

func (d Deviation) IsDrop() bool {
	return d.MeanDeviation < 0
}

func detectDeviations(baseline Baseline, readings []domain.Reading, cfg Config) ([]Deviation, int) {
	count := len(readings)
	if count == 0 {
		return nil, 0
	}
	direction := make([]int, count)
	zScores := make([]float64, count)
	for i, reading := range readings {
		expected := baseline.ExpectedConsumption(reading.Timestamp)
		if expected <= 0 {
			continue
		}
		relative := reading.ConsumptionKWh/expected - 1
		zScores[i] = (reading.ConsumptionKWh - expected) / baseline.ConsumptionStdDev(reading.Timestamp)
		switch {
		case relative > cfg.DeviationThreshold:
			direction[i] = 1
		case relative < -cfg.DeviationThreshold:
			direction[i] = -1
		}
	}

	var deviations []Deviation
	outliers := 0
	for _, sign := range []int{-1, 1} {
		for i := 0; i < count; {
			if direction[i] != sign {
				i++
				continue
			}
			start, end := i, i
			for j := i + 1; j < count && j-end <= maxGapReadings+1; j++ {
				if direction[j] == sign {
					end = j
				}
			}
			if end-start+1 >= cfg.MinTransientHours {
				ongoing := end >= count-ongoingTailReadings
				deviations = append(deviations, buildDeviation(baseline, readings[start:end+1], zScores[start:end+1], sign, ongoing, cfg))
			} else {
				for k := start; k <= end; k++ {
					if math.Abs(zScores[k]) >= cfg.OutlierZScore {
						outliers++
					}
				}
			}
			i = end + 1
		}
	}
	sort.Slice(deviations, func(i, j int) bool { return deviations[i].Strength() > deviations[j].Strength() })
	return deviations, outliers
}

func buildDeviation(baseline Baseline, readings []domain.Reading, zScores []float64, sign int, ongoing bool, cfg Config) Deviation {
	deviation := Deviation{
		Start:    readings[0].Timestamp,
		End:      readings[len(readings)-1].Timestamp,
		Hours:    len(readings),
		Ongoing:  ongoing,
		Readings: readings,
	}
	var observed, expected float64
	for i, reading := range readings {
		observed += reading.ConsumptionKWh
		expected += baseline.ExpectedConsumption(reading.Timestamp)
		if math.Abs(zScores[i]) > math.Abs(deviation.PeakZScore) {
			deviation.PeakZScore = zScores[i]
		}
	}
	deviation.ExcessKWh = observed - expected
	if expected > 0 {
		deviation.MeanDeviation = observed/expected - 1
	}
	switch {
	case ongoing && deviation.Hours >= cfg.MinPersistentHours:
		deviation.Kind = SignalPersistentShift
	case sign > 0:
		deviation.Kind = SignalTransientSpike
	default:
		deviation.Kind = SignalTransientDrop
	}
	return deviation
}
