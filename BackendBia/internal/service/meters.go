package service

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/engine"
	"github.com/biaenergy/backend/internal/numeric"
)

const (
	hoursPerDay         = 24
	expectedRangeSigmas = 2
	dateLayout          = "2006-01-02"
)

type DailyConsumption struct {
	Date     string  `json:"date"`
	KWh      float64 `json:"kwh"`
	Baseline float64 `json:"baseline"`
}

type MeterAnomalyRef struct {
	ID           string             `json:"id"`
	Type         domain.AnomalyType `json:"type"`
	Severity     domain.Severity    `json:"severity"`
	Confidence   float64            `json:"confidence"`
	PriorityRank int                `json:"priority_rank"`
	Title        string             `json:"title"`
}

type MeterSummary struct {
	domain.Meter
	PeriodKWh        float64            `json:"period_kwh"`
	LastDayKWh       float64            `json:"last_day_kwh"`
	BaselineDailyKWh float64            `json:"baseline_daily_kwh"`
	VariationPct     float64            `json:"variation_pct"`
	AvgVoltage       float64            `json:"avg_voltage_v"`
	AvgCurrent       float64            `json:"avg_current_a"`
	AvgPowerFactor   float64            `json:"avg_power_factor"`
	LastReadingAt    time.Time          `json:"last_reading_at"`
	Daily            []DailyConsumption `json:"daily"`
	Anomaly          *MeterAnomalyRef   `json:"anomaly"`
	AnomalyCount     int                `json:"anomaly_count"`
}

type MeterDetail struct {
	Meter     MeterSummary     `json:"meter"`
	Baseline  engine.Baseline  `json:"baseline"`
	Events    []domain.Event   `json:"events"`
	Anomalies []domain.Anomaly `json:"anomalies"`
}

type ReadingPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	ConsumptionKWh float64   `json:"consumption_kwh"`
	ExpectedKWh    float64   `json:"expected_kwh"`
	ExpectedLow    float64   `json:"expected_low"`
	ExpectedHigh   float64   `json:"expected_high"`
	VoltageV       float64   `json:"voltage_v"`
	CurrentA       float64   `json:"current_a"`
	PowerFactor    float64   `json:"power_factor"`
	Status         string    `json:"status"`
}

type MeterFilter string

const (
	FilterAll      MeterFilter = "all"
	FilterNormal   MeterFilter = "normal"
	FilterAlert    MeterFilter = "alert"
	FilterCritical MeterFilter = "critical"
)

type MeterQuery struct {
	Filter     MeterFilter
	Search     string
	SortBy     string
	Descending bool
}

type ReadingQuery struct {
	From  time.Time
	To    time.Time
	Daily bool
}

type MeterService struct {
	meters    MeterRepository
	readings  ReadingRepository
	events    EventRepository
	anomalies AnomalyRepository
	baselines map[string]engine.Baseline
}

func NewMeterService(repository Repository, analyzer *engine.Engine) *MeterService {
	baselines := map[string]engine.Baseline{}
	for meterID, readings := range repository.ReadingsByMeter() {
		from, to := analyzer.ReferenceWindow(readings)
		baselines[meterID] = engine.BuildBaseline(meterID, readings, from, to)
	}
	return &MeterService{
		meters:    repository,
		readings:  repository,
		events:    repository,
		anomalies: repository,
		baselines: baselines,
	}
}

func (s *MeterService) Summaries() []MeterSummary {
	anomalies := s.anomalies.ListAnomalies()
	meters := s.meters.ListMeters()
	summaries := make([]MeterSummary, 0, len(meters))
	for _, meter := range meters {
		summaries = append(summaries, s.summarize(meter, anomalies))
	}
	return summaries
}

func (s *MeterService) List(query MeterQuery) []MeterSummary {
	search := strings.ToLower(strings.TrimSpace(query.Search))
	summaries := []MeterSummary{}
	for _, summary := range s.Summaries() {
		if !matchesFilter(summary.Status, query.Filter) {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(summary.MeterID+" "+summary.Name+" "+summary.Location), search) {
			continue
		}
		summaries = append(summaries, summary)
	}
	sortSummaries(summaries, query.SortBy, query.Descending)
	return summaries
}

func (s *MeterService) Get(meterID string) (MeterDetail, error) {
	meter, err := s.meters.FindMeter(meterID)
	if err != nil {
		return MeterDetail{}, err
	}
	allAnomalies := s.anomalies.ListAnomalies()
	meterAnomalies := []domain.Anomaly{}
	for _, anomaly := range allAnomalies {
		if anomaly.MeterID == meterID {
			meterAnomalies = append(meterAnomalies, anomaly)
		}
	}
	return MeterDetail{
		Meter:     s.summarize(meter, allAnomalies),
		Baseline:  s.baselines[meterID],
		Events:    s.events.ListEvents(meterID),
		Anomalies: meterAnomalies,
	}, nil
}

func (s *MeterService) Readings(meterID string, query ReadingQuery) ([]ReadingPoint, error) {
	if _, err := s.meters.FindMeter(meterID); err != nil {
		return nil, err
	}
	baseline := s.baselines[meterID]
	readings := s.readings.ListReadings(meterID, query.From, query.To)
	points := make([]ReadingPoint, 0, len(readings))
	for _, reading := range readings {
		expected := baseline.ExpectedConsumption(reading.Timestamp)
		spread := expectedRangeSigmas * baseline.ConsumptionStdDev(reading.Timestamp)
		points = append(points, ReadingPoint{
			Timestamp:      reading.Timestamp,
			ConsumptionKWh: reading.ConsumptionKWh,
			ExpectedKWh:    numeric.Round(expected, 1),
			ExpectedLow:    numeric.Round(math.Max(0, expected-spread), 1),
			ExpectedHigh:   numeric.Round(expected+spread, 1),
			VoltageV:       reading.VoltageV,
			CurrentA:       reading.CurrentA,
			PowerFactor:    reading.PowerFactor,
			Status:         reading.Status,
		})
	}
	if query.Daily {
		return aggregateDaily(points), nil
	}
	return points, nil
}

func (s *MeterService) Events(meterID string) []domain.Event {
	return s.events.ListEvents(meterID)
}

func (s *MeterService) summarize(meter domain.Meter, anomalies []domain.Anomaly) MeterSummary {
	readings := s.readings.ListReadings(meter.MeterID, time.Time{}, time.Time{})
	baseline := s.baselines[meter.MeterID]
	summary := MeterSummary{Meter: meter, BaselineDailyKWh: numeric.Round(baseline.DailyKWh, 1)}

	dailyTotals := map[string]float64{}
	var days []string
	var voltage, current, powerFactor float64
	lastDayCount := 0
	for i, reading := range readings {
		summary.PeriodKWh += reading.ConsumptionKWh
		day := reading.Timestamp.Format(dateLayout)
		if _, seen := dailyTotals[day]; !seen {
			days = append(days, day)
		}
		dailyTotals[day] += reading.ConsumptionKWh
		if i >= len(readings)-hoursPerDay {
			summary.LastDayKWh += reading.ConsumptionKWh
			voltage += reading.VoltageV
			current += reading.CurrentA
			powerFactor += reading.PowerFactor
			lastDayCount++
		}
		summary.LastReadingAt = reading.Timestamp
	}
	for _, day := range days {
		summary.Daily = append(summary.Daily, DailyConsumption{
			Date:     day,
			KWh:      numeric.Round(dailyTotals[day], 1),
			Baseline: numeric.Round(baseline.DailyKWh, 1),
		})
	}
	summary.PeriodKWh = numeric.Round(summary.PeriodKWh, 1)
	summary.LastDayKWh = numeric.Round(summary.LastDayKWh, 1)
	if baseline.DailyKWh > 0 {
		summary.VariationPct = numeric.Round(numeric.PercentChange(summary.LastDayKWh, baseline.DailyKWh), 1)
	}
	if lastDayCount > 0 {
		count := float64(lastDayCount)
		summary.AvgVoltage = numeric.Round(voltage/count, 1)
		summary.AvgCurrent = numeric.Round(current/count, 1)
		summary.AvgPowerFactor = numeric.Round(powerFactor/count, 3)
	}
	summary.Anomaly, summary.AnomalyCount = topAnomaly(meter.MeterID, anomalies)
	return summary
}

func topAnomaly(meterID string, anomalies []domain.Anomaly) (*MeterAnomalyRef, int) {
	var top *domain.Anomaly
	count := 0
	for i := range anomalies {
		if anomalies[i].MeterID != meterID {
			continue
		}
		count++
		if top == nil || anomalies[i].PriorityRank < top.PriorityRank {
			top = &anomalies[i]
		}
	}
	if top == nil {
		return nil, 0
	}
	return &MeterAnomalyRef{
		ID:           top.ID,
		Type:         top.Type,
		Severity:     top.Severity,
		Confidence:   top.Confidence,
		PriorityRank: top.PriorityRank,
		Title:        top.Title,
	}, count
}

func matchesFilter(status domain.MeterStatus, filter MeterFilter) bool {
	switch filter {
	case FilterNormal:
		return status == domain.MeterOK || status == domain.MeterPending
	case FilterAlert:
		return status == domain.MeterAlert
	case FilterCritical:
		return status == domain.MeterCritical
	default:
		return true
	}
}

func severityScore(summary MeterSummary) int {
	if summary.Anomaly == nil {
		return 0
	}
	score := summary.Anomaly.Severity.Weight() * 10
	if summary.Anomaly.Type == domain.FalsePositive {
		score = 1
	}
	switch summary.Status {
	case domain.MeterCritical:
		score += 5
	case domain.MeterAlert:
		score += 3
	}
	return score
}

func sortSummaries(summaries []MeterSummary, sortBy string, descending bool) {
	less := map[string]func(a, b MeterSummary) bool{
		"consumption": func(a, b MeterSummary) bool { return a.LastDayKWh < b.LastDayKWh },
		"variation":   func(a, b MeterSummary) bool { return a.VariationPct < b.VariationPct },
		"severity":    func(a, b MeterSummary) bool { return severityScore(a) < severityScore(b) },
	}[sortBy]
	if less == nil {
		sort.SliceStable(summaries, func(i, j int) bool { return summaries[i].MeterID < summaries[j].MeterID })
		return
	}
	sort.SliceStable(summaries, func(i, j int) bool {
		if descending {
			return !less(summaries[i], summaries[j])
		}
		return less(summaries[i], summaries[j])
	})
}

func aggregateDaily(points []ReadingPoint) []ReadingPoint {
	type accumulator struct {
		point ReadingPoint
		count int
	}
	byDay := map[string]*accumulator{}
	var days []string
	for _, point := range points {
		day := point.Timestamp.Format(dateLayout)
		acc, ok := byDay[day]
		if !ok {
			acc = &accumulator{point: ReadingPoint{Timestamp: point.Timestamp.Truncate(hoursPerDay * time.Hour), Status: "OK"}}
			byDay[day] = acc
			days = append(days, day)
		}
		acc.count++
		acc.point.ConsumptionKWh += point.ConsumptionKWh
		acc.point.ExpectedKWh += point.ExpectedKWh
		acc.point.ExpectedLow += point.ExpectedLow
		acc.point.ExpectedHigh += point.ExpectedHigh
		acc.point.VoltageV += point.VoltageV
		acc.point.CurrentA += point.CurrentA
		acc.point.PowerFactor += point.PowerFactor
	}
	sort.Strings(days)
	daily := make([]ReadingPoint, 0, len(days))
	for _, day := range days {
		acc := byDay[day]
		count := float64(acc.count)
		point := acc.point
		point.ConsumptionKWh = numeric.Round(point.ConsumptionKWh, 1)
		point.ExpectedKWh = numeric.Round(point.ExpectedKWh, 1)
		point.ExpectedLow = numeric.Round(point.ExpectedLow, 1)
		point.ExpectedHigh = numeric.Round(point.ExpectedHigh, 1)
		point.VoltageV = numeric.Round(point.VoltageV/count, 1)
		point.CurrentA = numeric.Round(point.CurrentA/count, 1)
		point.PowerFactor = numeric.Round(point.PowerFactor/count, 3)
		daily = append(daily, point)
	}
	return daily
}
