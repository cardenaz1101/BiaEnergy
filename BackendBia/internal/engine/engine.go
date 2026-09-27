package engine

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/biaenergy/backend/internal/domain"
)

type Step struct {
	Key   string
	Label string
}

var Steps = []Step{
	{Key: "readings", Label: "Lecturas"},
	{Key: "baseline", Label: "Baseline"},
	{Key: "detection", Label: "Detección"},
	{Key: "correlation", Label: "Correlación"},
	{Key: "events", Label: "Eventos"},
	{Key: "explanation", Label: "Explicación"},
	{Key: "recommendation", Label: "Recomendación"},
}

type StepReporter func(stepKey, detail string) error

type Input struct {
	Meters   []domain.Meter
	Readings map[string][]domain.Reading
	Events   []domain.Event
}

type Result struct {
	Anomalies        []domain.Anomaly
	MeterStatuses    map[string]domain.MeterStatus
	ReadingsAnalyzed int
}

type Engine struct {
	config Config
}

func New(config Config) *Engine {
	return &Engine{config: config}
}

func (e *Engine) ReferenceWindow(readings []domain.Reading) (time.Time, time.Time) {
	if len(readings) == 0 {
		return time.Time{}, time.Time{}
	}
	from := readings[0].Timestamp.Truncate(hoursPerDay * time.Hour)
	to := from.Add(time.Duration(e.config.BaselineDays) * hoursPerDay * time.Hour)
	return from, to
}

func (e *Engine) Analyze(ctx context.Context, runID string, input Input, report StepReporter) (Result, error) {
	if report == nil {
		report = func(string, string) error { return nil }
	}
	run := &pipeline{
		engine:     e,
		input:      input,
		runID:      runID,
		detectedAt: time.Now().UTC(),
		baselines:  map[string]Baseline{},
		evaluation: map[string][]domain.Reading{},
	}
	stages := []func() string{
		run.validateReadings,
		run.buildBaselines,
		run.detect,
		run.correlate,
		run.matchEvents,
		run.explain,
		run.recommend,
	}
	for i, stage := range stages {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if err := report(Steps[i].Key, stage()); err != nil {
			return Result{}, err
		}
	}
	return run.result, nil
}

type pipeline struct {
	engine     *Engine
	input      Input
	runID      string
	detectedAt time.Time
	meterIDs   []string
	baselines  map[string]Baseline
	evaluation map[string][]domain.Reading
	findings   []*finding
	result     Result
}

func (p *pipeline) validateReadings() string {
	for _, meter := range p.input.Meters {
		p.meterIDs = append(p.meterIDs, meter.MeterID)
		p.result.ReadingsAnalyzed += len(p.input.Readings[meter.MeterID])
	}
	sort.Strings(p.meterIDs)
	return fmt.Sprintf("%s lecturas horarias de %d medidores validadas",
		formatNumber(float64(p.result.ReadingsAnalyzed), 0), len(p.meterIDs))
}

func (p *pipeline) buildBaselines() string {
	var from, to time.Time
	for _, meterID := range p.meterIDs {
		readings := p.input.Readings[meterID]
		from, to = p.engine.ReferenceWindow(readings)
		p.baselines[meterID] = BuildBaseline(meterID, readings, from, to)
		for _, reading := range readings {
			if !reading.Timestamp.Before(to) {
				p.evaluation[meterID] = append(p.evaluation[meterID], reading)
			}
		}
	}
	return fmt.Sprintf("Perfil horario por medidor aprendido con %d días de referencia (%s – %s)",
		p.engine.config.BaselineDays, formatDay(from), formatDay(to.Add(-time.Hour)))
}

func (p *pipeline) detect() string {
	deviationCount, qualityCount, outlierCount := 0, 0, 0
	for _, meterID := range p.meterIDs {
		baseline := p.baselines[meterID]
		evaluation := p.evaluation[meterID]
		deviations, outliers := detectDeviations(baseline, evaluation, p.engine.config)
		outlierCount += outliers
		quality := scanDataQuality(baseline, evaluation)
		lastDay := lastDayConsumption(p.input.Readings[meterID])

		if quality.IsIssue() {
			p.findings = append(p.findings, &finding{meterID: meterID, baseline: baseline, dataQuality: &quality, lastDayKWh: lastDay})
			qualityCount++
		}
		if len(deviations) == 0 {
			continue
		}
		strongest := deviations[0]
		if quality.IsIssue() && !strongest.Start.Before(quality.First) {
			continue
		}
		p.findings = append(p.findings, &finding{meterID: meterID, baseline: baseline, deviation: &strongest, lastDayKWh: lastDay})
		deviationCount++
	}
	return fmt.Sprintf("%d desviaciones de consumo y %d problemas de calidad de datos detectados · %d outliers puntuales ignorados",
		deviationCount, qualityCount, outlierCount)
}

func (p *pipeline) correlate() string {
	electricalCount := 0
	for _, f := range p.findings {
		f.variables = compareVariables(f.baseline, p.findingWindow(f))
		if f.deviation != nil {
			f.electricalReasons = electricalChangeReasons(f.variables)
			if f.hasElectricalChange() {
				electricalCount++
			}
		}
	}
	return fmt.Sprintf("Consumo, voltaje, corriente y FP contrastados · %d hallazgos con cambio eléctrico asociado", electricalCount)
}

func (p *pipeline) findingWindow(f *finding) []domain.Reading {
	if f.deviation != nil {
		return f.deviation.Readings
	}
	var window []domain.Reading
	for _, reading := range p.evaluation[f.meterID] {
		if !reading.Timestamp.Before(f.dataQuality.First) {
			window = append(window, reading)
		}
	}
	return window
}

func (p *pipeline) matchEvents() string {
	matched := 0
	windowHours := p.engine.config.EventWindow.Hours()
	for _, f := range p.findings {
		onset := f.onset()
		for _, event := range p.input.Events {
			if event.MeterID == f.meterID && math.Abs(event.Timestamp.Sub(onset).Hours()) <= windowHours {
				f.events = append(f.events, event)
			}
		}
		if len(f.events) > 0 {
			matched++
		}
	}
	return fmt.Sprintf("%d eventos operativos cruzados · %d hallazgos con evento asociado", len(p.input.Events), matched)
}

func (p *pipeline) explain() string {
	for _, f := range p.findings {
		f.anomaly = classify(f, p.detectedAt, p.runID)
	}
	return "Clasificación, evidencia y confianza generadas para cada hallazgo"
}

func (p *pipeline) recommend() string {
	for _, f := range p.findings {
		describe(f)
		p.result.Anomalies = append(p.result.Anomalies, f.anomaly)
	}
	prioritize(p.result.Anomalies)

	p.result.MeterStatuses = map[string]domain.MeterStatus{}
	for _, meterID := range p.meterIDs {
		p.result.MeterStatuses[meterID] = domain.MeterOK
	}
	priorityCount := 0
	for _, anomaly := range p.result.Anomalies {
		status := anomaly.MeterStatus()
		if status.Rank() > p.result.MeterStatuses[anomaly.MeterID].Rank() {
			p.result.MeterStatuses[anomaly.MeterID] = status
		}
		if anomaly.IsPriority() {
			priorityCount++
		}
	}
	return fmt.Sprintf("%d anomalías detectadas · %d requieren atención prioritaria", len(p.result.Anomalies), priorityCount)
}

func lastDayConsumption(readings []domain.Reading) float64 {
	total := 0.0
	for i := len(readings) - 1; i >= 0 && i >= len(readings)-hoursPerDay; i-- {
		total += readings[i].ConsumptionKWh
	}
	return total
}
