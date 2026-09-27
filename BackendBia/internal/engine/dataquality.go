package engine

import (
	"math"
	"sort"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/numeric"
)

const (
	voltageOutlierSigmas       = 6.0
	powerFactorOutlierSigmas   = 5.0
	powerFactorOutlierMinDelta = 0.08
	powerRatioOutlierSigmas    = 5.0
	voltageJumpVolts           = 10.0
	denseClusterWindow         = 12
	denseClusterMinFlags       = 3
	trailingCleanReadings      = 6
	repeatedValueMinCount      = 3
	intermittentMinTransitions = 4
	intermittentMinFraction    = 0.08
	intermittentMaxFraction    = 0.85
	missingHoursIssueThreshold = 3
	powerFactorPrecision       = 3
)

const (
	reasonInvalidValue        = "valor físicamente imposible"
	reasonVoltageOutOfRange   = "voltaje fuera de rango"
	reasonAtypicalPowerFactor = "factor de potencia atípico"
	reasonIncoherentEnergy    = "energía incoherente con V·I·FP"
)

type FlaggedReading struct {
	Reading domain.Reading
	Reasons []string
}

type ReasonCount struct {
	Reason string
	Count  int
}

type DataQualityFinding struct {
	Flagged              []FlaggedReading
	First                time.Time
	Last                 time.Time
	SpanReadings         int
	Fraction             float64
	Transitions          int
	InvalidReadings      int
	MissingHours         int
	DuplicateReadings    int
	VoltageJumps         int
	RepeatedPowerFactors []float64
	Reasons              []ReasonCount
	Intermittent         bool
	MinVoltage           float64
	MaxVoltage           float64
	VoltageStdDev        float64
}

func (f DataQualityFinding) IsIssue() bool {
	return f.Intermittent || f.InvalidReadings > 0 || f.MissingHours >= missingHoursIssueThreshold || f.DuplicateReadings > 0
}

func (f DataQualityFinding) ReadingsPerFlag() int {
	return max(1, int(1/f.Fraction+0.5))
}

func flagReasons(baseline Baseline, reading domain.Reading) []string {
	var reasons []string
	if reading.IsPhysicallyInvalid() {
		reasons = append(reasons, reasonInvalidValue)
	}
	if math.Abs(reading.VoltageV-baseline.Voltage.Mean) > voltageOutlierSigmas*baseline.Voltage.Std {
		reasons = append(reasons, reasonVoltageOutOfRange)
	}
	if math.Abs(reading.PowerFactor-baseline.PowerFactor.Mean) > math.Max(powerFactorOutlierMinDelta, powerFactorOutlierSigmas*baseline.PowerFactor.Std) {
		reasons = append(reasons, reasonAtypicalPowerFactor)
	}
	if ratio := reading.PowerRatio(); !math.IsNaN(ratio) && math.Abs(ratio-baseline.PowerRatio.Mean) > powerRatioOutlierSigmas*baseline.PowerRatio.Std {
		reasons = append(reasons, reasonIncoherentEnergy)
	}
	return reasons
}

func scanDataQuality(baseline Baseline, readings []domain.Reading) DataQualityFinding {
	var finding DataQualityFinding
	flags := make([]bool, len(readings))
	reasonsByIndex := make([][]string, len(readings))
	powerFactorCounts := map[float64]int{}

	for i, reading := range readings {
		reasons := flagReasons(baseline, reading)
		if reading.IsPhysicallyInvalid() {
			finding.InvalidReadings++
		}
		if i > 0 {
			gap := reading.Timestamp.Sub(readings[i-1].Timestamp)
			switch {
			case gap == 0:
				finding.DuplicateReadings++
			case gap > time.Hour:
				finding.MissingHours += int(gap/time.Hour) - 1
			}
			if math.Abs(reading.VoltageV-readings[i-1].VoltageV) > voltageJumpVolts {
				finding.VoltageJumps++
			}
		}
		powerFactorCounts[numeric.Round(reading.PowerFactor, powerFactorPrecision)]++
		flags[i] = len(reasons) > 0
		reasonsByIndex[i] = reasons
	}

	first := firstDenseFlag(flags)
	if first < 0 {
		return finding
	}
	last := lastFlag(flags)
	for i := range first {
		flags[i] = false
	}

	reasonCounts := map[string]int{}
	var voltages []float64
	for i := first; i < len(readings); i++ {
		if !flags[i] {
			continue
		}
		finding.Flagged = append(finding.Flagged, FlaggedReading{Reading: readings[i], Reasons: reasonsByIndex[i]})
		for _, reason := range reasonsByIndex[i] {
			reasonCounts[reason]++
		}
		voltages = append(voltages, readings[i].VoltageV)
	}

	finding.First = readings[first].Timestamp
	finding.Last = readings[last].Timestamp
	finding.SpanReadings = last - first + 1
	if len(readings)-1-last < trailingCleanReadings {
		finding.SpanReadings = len(readings) - first
	}
	finding.Fraction = float64(len(finding.Flagged)) / float64(finding.SpanReadings)
	for i := first + 1; i < len(readings); i++ {
		if flags[i] != flags[i-1] {
			finding.Transitions++
		}
	}
	finding.Reasons = sortedReasons(reasonCounts)
	finding.RepeatedPowerFactors = repeatedPowerFactors(finding.Flagged, powerFactorCounts)
	finding.MinVoltage, finding.MaxVoltage = minMax(voltages)
	finding.VoltageStdDev = numeric.StdDev(voltages)
	finding.Intermittent = finding.Transitions >= intermittentMinTransitions &&
		finding.Fraction >= intermittentMinFraction &&
		finding.Fraction <= intermittentMaxFraction
	return finding
}

func firstDenseFlag(flags []bool) int {
	firstAny := -1
	for i, flagged := range flags {
		if !flagged {
			continue
		}
		if firstAny < 0 {
			firstAny = i
		}
		inWindow := 0
		for k := i; k < len(flags) && k < i+denseClusterWindow; k++ {
			if flags[k] {
				inWindow++
			}
		}
		if inWindow >= denseClusterMinFlags {
			return i
		}
	}
	return firstAny
}

func lastFlag(flags []bool) int {
	for i := len(flags) - 1; i >= 0; i-- {
		if flags[i] {
			return i
		}
	}
	return -1
}

func sortedReasons(counts map[string]int) []ReasonCount {
	reasons := make([]ReasonCount, 0, len(counts))
	for reason, count := range counts {
		reasons = append(reasons, ReasonCount{Reason: reason, Count: count})
	}
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Count != reasons[j].Count {
			return reasons[i].Count > reasons[j].Count
		}
		return reasons[i].Reason < reasons[j].Reason
	})
	return reasons
}

func repeatedPowerFactors(flagged []FlaggedReading, allCounts map[float64]int) []float64 {
	flaggedCounts := map[float64]int{}
	for _, item := range flagged {
		flaggedCounts[numeric.Round(item.Reading.PowerFactor, powerFactorPrecision)]++
	}
	var repeated []float64
	for value, count := range flaggedCounts {
		if count >= repeatedValueMinCount && allCounts[value] >= repeatedValueMinCount {
			repeated = append(repeated, value)
		}
	}
	sort.Float64s(repeated)
	return repeated
}

func minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	low, high := values[0], values[0]
	for _, v := range values {
		low, high = math.Min(low, v), math.Max(high, v)
	}
	return low, high
}
