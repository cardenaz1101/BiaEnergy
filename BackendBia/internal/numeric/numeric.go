package numeric

import (
	"math"
	"sort"
)

const madToStdDev = 1.4826

func Round(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}

func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func StdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := Mean(values)
	sum := 0.0
	for _, v := range values {
		sum += (v - mean) * (v - mean)
	}
	return math.Sqrt(sum / float64(len(values)-1))
}

func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func RobustStdDev(values []float64) float64 {
	median := Median(values)
	deviations := make([]float64, len(values))
	for i, v := range values {
		deviations[i] = math.Abs(v - median)
	}
	return madToStdDev * Median(deviations)
}

func PercentChange(observed, baseline float64) float64 {
	if baseline == 0 {
		return 0
	}
	return (observed/baseline - 1) * 100
}
