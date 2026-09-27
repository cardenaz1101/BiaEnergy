package service

import (
	"sort"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/numeric"
)

const topAnomaliesLimit = 5

type Period struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type DashboardSummary struct {
	MetersTotal           int                        `json:"meters_total"`
	MetersByStatus        map[domain.MeterStatus]int `json:"meters_by_status"`
	ConsumptionTotalKWh   float64                    `json:"consumption_total_kwh"`
	ConsumptionLastDayKWh float64                    `json:"consumption_last_day_kwh"`
	BaselineDailyKWh      float64                    `json:"baseline_daily_kwh"`
	VariationPct          float64                    `json:"variation_pct"`
	Period                Period                     `json:"period"`
	AnomaliesDetected     int                        `json:"anomalies_detected"`
	AnomaliesActionable   int                        `json:"anomalies_actionable"`
	AnomaliesOpen         int                        `json:"anomalies_open"`
	HighPriority          int                        `json:"high_priority"`
	AvgConfidence         float64                    `json:"avg_confidence"`
	AnomaliesByType       map[domain.AnomalyType]int `json:"anomalies_by_type"`
	TopAnomalies          []domain.Anomaly           `json:"top_anomalies"`
	DailyConsumption      []DailyConsumption         `json:"daily_consumption"`
	LastAnalysis          *domain.AnalysisRun        `json:"last_analysis"`
}

type DashboardService struct {
	meters    *MeterService
	anomalies AnomalyRepository
	analyses  AnalysisRepository
}

func NewDashboardService(meters *MeterService, anomalies AnomalyRepository, analyses AnalysisRepository) *DashboardService {
	return &DashboardService{meters: meters, anomalies: anomalies, analyses: analyses}
}

func (s *DashboardService) Summary() DashboardSummary {
	summary := DashboardSummary{
		MetersByStatus:  map[domain.MeterStatus]int{domain.MeterOK: 0, domain.MeterAlert: 0, domain.MeterCritical: 0, domain.MeterPending: 0},
		AnomaliesByType: map[domain.AnomalyType]int{},
	}
	s.addConsumption(&summary)
	s.addAnomalies(&summary)
	if run, ok := s.analyses.LatestRun(); ok {
		summary.LastAnalysis = &run
	}
	return summary
}

func (s *DashboardService) addConsumption(summary *DashboardSummary) {
	meters := s.meters.Summaries()
	summary.MetersTotal = len(meters)
	dailyTotals := map[string]*DailyConsumption{}
	var totalKWh, lastDayKWh, baselineKWh float64

	for _, meter := range meters {
		summary.MetersByStatus[meter.Status]++
		totalKWh += meter.PeriodKWh
		lastDayKWh += meter.LastDayKWh
		baselineKWh += meter.BaselineDailyKWh
		for _, day := range meter.Daily {
			total, ok := dailyTotals[day.Date]
			if !ok {
				total = &DailyConsumption{Date: day.Date}
				dailyTotals[day.Date] = total
			}
			total.KWh += day.KWh
			total.Baseline += day.Baseline
		}
		if len(meter.Daily) > 0 {
			firstDay, _ := time.Parse(dateLayout, meter.Daily[0].Date)
			if summary.Period.From.IsZero() || firstDay.Before(summary.Period.From) {
				summary.Period.From = firstDay
			}
		}
		if meter.LastReadingAt.After(summary.Period.To) {
			summary.Period.To = meter.LastReadingAt
		}
	}

	summary.DailyConsumption = make([]DailyConsumption, 0, len(dailyTotals))
	for _, total := range dailyTotals {
		summary.DailyConsumption = append(summary.DailyConsumption, DailyConsumption{
			Date:     total.Date,
			KWh:      numeric.Round(total.KWh, 1),
			Baseline: numeric.Round(total.Baseline, 1),
		})
	}
	sort.Slice(summary.DailyConsumption, func(i, j int) bool {
		return summary.DailyConsumption[i].Date < summary.DailyConsumption[j].Date
	})
	summary.ConsumptionTotalKWh = numeric.Round(totalKWh, 1)
	summary.ConsumptionLastDayKWh = numeric.Round(lastDayKWh, 1)
	summary.BaselineDailyKWh = numeric.Round(baselineKWh, 1)
	if baselineKWh > 0 {
		summary.VariationPct = numeric.Round(numeric.PercentChange(lastDayKWh, baselineKWh), 1)
	}
}

func (s *DashboardService) addAnomalies(summary *DashboardSummary) {
	anomalies := s.anomalies.ListAnomalies()
	summary.AnomaliesDetected = len(anomalies)
	confidenceSum := 0.0
	for _, anomaly := range anomalies {
		summary.AnomaliesByType[anomaly.Type]++
		confidenceSum += anomaly.Confidence
		if anomaly.IsPriority() {
			summary.HighPriority++
		}
		if anomaly.IsAnomaly {
			summary.AnomaliesActionable++
		}
		if anomaly.Status.IsActive() {
			summary.AnomaliesOpen++
		}
	}
	if len(anomalies) > 0 {
		summary.AvgConfidence = numeric.Round(confidenceSum/float64(len(anomalies)), 2)
	}
	summary.TopAnomalies = anomalies[:min(len(anomalies), topAnomaliesLimit)]
}
