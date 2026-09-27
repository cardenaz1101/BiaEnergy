package engine

import (
	"math"
	"time"
)

const (
	minConfidence = 0.05
	maxConfidence = 0.97

	magnitudeBonusRate = 0.2
	maxMagnitudeBonus  = 0.2

	weightOutageEvent         = 0.45
	weightDurationMatches     = 0.20
	weightReturnToBaseline    = 0.12
	outageDurationToleranceHr = 3

	weightDeviationWithEvent      = 0.35
	weightUnexplainedElectrical   = 0.20
	weightPartialEventExplanation = 0.15
	weightOperationalEvent        = 0.40
	weightStableNewLevel          = 0.12
	weightCoherentElectrical      = 0.12

	weightDeviation          = 0.45
	weightPersistence        = 0.12
	weightTransient          = 0.04
	weightElectricalChange   = 0.12
	weightNoExplainingEvent  = 0.07
	weightConsistentReadings = 0.04

	weightIntermittentReadings = 0.50
	weightHighAffectedShare    = 0.15
	weightStableConsumption    = 0.12
	weightCorruptTelemetry     = 0.08
	weightDataQualityEvent     = 0.07

	timingExactHours  = 2
	timingCloseHours  = 6
	weightTimingExact = 0.20
	weightTimingClose = 0.12
	weightTimingLoose = 0.05
)

func magnitudeBonus(meanDeviation float64) float64 {
	return math.Min(maxMagnitudeBonus, math.Abs(meanDeviation)*magnitudeBonusRate)
}

func timingWeight(eventAt, onset time.Time) float64 {
	hours := math.Abs(eventAt.Sub(onset).Hours())
	switch {
	case hours <= timingExactHours:
		return weightTimingExact
	case hours <= timingCloseHours:
		return weightTimingClose
	default:
		return weightTimingLoose
	}
}

func clampConfidence(total float64) float64 {
	return math.Max(minConfidence, math.Min(maxConfidence, total))
}
