package engine

import (
	"math"
	"sort"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/numeric"
)

const (
	maxSeverityWeight  = 3.0
	confidenceShare    = 0.9
	impactBonusPoints  = 10.0
	impactReferenceKWh = 5000.0
	maxPriorityScore   = 100.0
)

func typeWeight(anomalyType domain.AnomalyType) float64 {
	switch anomalyType {
	case domain.RealAnomaly:
		return 1.0
	case domain.DataQuality:
		return 0.85
	case domain.ExplainableAnomaly:
		return 0.55
	default:
		return 0.15
	}
}

func priorityScore(anomaly domain.Anomaly) float64 {
	weight := typeWeight(anomaly.Type)
	impact := math.Min(1, math.Abs(anomaly.Metrics.ExcessKWh)/impactReferenceKWh)
	base := maxPriorityScore * weight * (float64(anomaly.Severity.Weight()) / maxSeverityWeight) * anomaly.Confidence
	return base*confidenceShare + impactBonusPoints*impact*weight
}

func prioritize(anomalies []domain.Anomaly) {
	for i := range anomalies {
		anomalies[i].PriorityScore = numeric.Round(priorityScore(anomalies[i]), 1)
	}
	sort.SliceStable(anomalies, func(i, j int) bool { return anomalies[i].PriorityScore > anomalies[j].PriorityScore })
	for i := range anomalies {
		anomalies[i].PriorityRank = i + 1
	}
}
