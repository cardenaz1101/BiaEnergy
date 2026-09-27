package engine

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/numeric"
)

const (
	highSeverityDeviation      = 0.5
	mediumSeverityDeviation    = 0.35
	highAffectedShare          = 0.2
	stableConsumptionMaxChange = 10.0
)

type finding struct {
	meterID           string
	baseline          Baseline
	deviation         *Deviation
	dataQuality       *DataQualityFinding
	variables         []domain.VariableChange
	electricalReasons []string
	events            []domain.Event
	lastDayKWh        float64
	anomaly           domain.Anomaly
}

func (f *finding) onset() time.Time {
	if f.deviation != nil {
		return f.deviation.Start
	}
	return f.dataQuality.First
}

func (f *finding) hasElectricalChange() bool {
	return len(f.electricalReasons) > 0
}

type anomalyBuilder struct {
	anomaly *domain.Anomaly
}

func (b anomalyBuilder) evidence(kind domain.EvidenceKind, stance domain.EvidenceStance, format string, args ...any) {
	b.anomaly.Evidence = append(b.anomaly.Evidence, domain.Evidence{Kind: kind, Stance: stance, Description: fmt.Sprintf(format, args...)})
}

func (b anomalyBuilder) factor(label string, contribution float64) {
	b.anomaly.ConfidenceBreakdown = append(b.anomaly.ConfidenceBreakdown,
		domain.ConfidenceFactor{Label: label, Contribution: numeric.Round(contribution, 2)})
}

func classify(f *finding, detectedAt time.Time, runID string) domain.Anomaly {
	anomaly := domain.Anomaly{
		RunID:            runID,
		MeterID:          f.meterID,
		DetectedAt:       detectedAt,
		Status:           domain.AnomalyOpen,
		RelatedEvents:    f.events,
		ChangedVariables: f.variables,
	}
	if anomaly.RelatedEvents == nil {
		anomaly.RelatedEvents = []domain.Event{}
	}
	anomaly.Metrics.BaselineDailyKWh = numeric.Round(f.baseline.DailyKWh, 1)
	anomaly.Metrics.CurrentDailyKWh = numeric.Round(f.lastDayKWh, 1)
	if f.baseline.DailyKWh > 0 {
		anomaly.Metrics.VariationPct = numeric.Round(numeric.PercentChange(f.lastDayKWh, f.baseline.DailyKWh), 1)
	}

	builder := anomalyBuilder{anomaly: &anomaly}
	if f.dataQuality != nil {
		classifyDataQuality(f, builder)
	} else {
		classifyDeviation(f, builder)
	}

	total := 0.0
	for _, factor := range anomaly.ConfidenceBreakdown {
		total += factor.Contribution
	}
	anomaly.Confidence = numeric.Round(clampConfidence(total), 2)
	anomaly.ID = fmt.Sprintf("ANM-%s-%s", strings.ReplaceAll(f.meterID, "-", ""), anomaly.Type.Code())
	return anomaly
}

func classifyDeviation(f *finding, b anomalyBuilder) {
	deviation := f.deviation
	anomaly := b.anomaly
	anomaly.Signal = deviation.Kind
	anomaly.StartedAt = deviation.Start
	anomaly.DurationHours = deviation.Hours
	if !deviation.Ongoing {
		end := deviation.End
		anomaly.EndedAt = &end
	}
	anomaly.Metrics.ExcessKWh = numeric.Round(deviation.ExcessKWh, 1)
	anomaly.Metrics.AffectedReadings = deviation.Hours

	addDeviationEvidence(f, b)

	outage := domain.FindEventByType(f.events, domain.EventScheduledOutage, domain.EventMaintenance)
	operationalChange := domain.FindEventByType(f.events, domain.EventOperationalChange)
	switch {
	case outage != nil && deviation.Kind == SignalTransientDrop:
		classifyScheduledOutage(f, b, outage)
	case operationalChange != nil:
		classifyOperationalChange(f, b, operationalChange)
	default:
		classifyUnexplained(f, b)
	}
}

func addDeviationEvidence(f *finding, b anomalyBuilder) {
	deviation := f.deviation
	deviationPct := deviation.MeanDeviation * 100
	switch deviation.Kind {
	case SignalPersistentShift:
		b.evidence(domain.EvidenceBaseline, domain.StanceSupports,
			"Consumo %s frente al perfil horario esperado de forma sostenida desde %s (%d h consecutivas, sigue activo).",
			formatSignedPercent(deviationPct), formatDateTime(deviation.Start), deviation.Hours)
	case SignalTransientDrop:
		b.evidence(domain.EvidenceBaseline, domain.StanceSupports, "Caída de consumo del %s entre %s y %s (%d h).",
			formatUnsignedPercent(deviationPct), formatDateTime(deviation.Start), formatDateTime(deviation.End), deviation.Hours)
		b.evidence(domain.EvidenceBaseline, domain.StanceContext,
			"Después del %s el consumo volvió a su comportamiento esperado.", formatDateTime(deviation.End))
	case SignalTransientSpike:
		b.evidence(domain.EvidenceBaseline, domain.StanceSupports, "Pico de consumo de %s entre %s y %s (%d h).",
			formatSignedPercent(deviationPct), formatDateTime(deviation.Start), formatDateTime(deviation.End), deviation.Hours)
	}
	b.evidence(domain.EvidenceBaseline, domain.StanceContext, "Consumo último día %s kWh vs baseline %s kWh/día (%s).",
		formatNumber(b.anomaly.Metrics.CurrentDailyKWh, 0), formatNumber(f.baseline.DailyKWh, 0), formatSignedPercent(b.anomaly.Metrics.VariationPct))

	if current := findVariable(f.variables, VariableCurrent); current.Significant {
		b.evidence(domain.EvidenceElectrical, domain.StanceContext, "La corriente acompaña el cambio: %s → %s A (%s).",
			formatNumber(current.Baseline, 0), formatNumber(current.Observed, 0), formatSignedPercent(current.DeltaPct))
	}
	if f.hasElectricalChange() {
		b.evidence(domain.EvidenceElectrical, domain.StanceSupports,
			"Cambio en la calidad eléctrica, no solo en la carga: %s.", strings.Join(f.electricalReasons, "; "))
	} else if deviation.Kind != SignalTransientDrop {
		b.evidence(domain.EvidenceElectrical, domain.StanceContext,
			"Voltaje y factor de potencia se mantienen en su rango normal: el cambio es de carga, no eléctrico.")
	}
	if ratio := findVariable(f.variables, VariablePowerRatio); ratio.Significant {
		b.evidence(domain.EvidenceElectrical, domain.StanceSupports,
			"La relación entre energía registrada y potencia calculada (V·I·FP) se desplazó %s.", formatSignedPercent(ratio.DeltaPct))
	}
}

func classifyScheduledOutage(f *finding, b anomalyBuilder, outage *domain.Event) {
	deviation := f.deviation
	b.anomaly.Type, b.anomaly.IsAnomaly, b.anomaly.Severity = domain.FalsePositive, false, domain.SeverityLow
	b.evidence(domain.EvidenceEvent, domain.StanceAgainst, "Evento %s el %s: \"%s\".",
		outage.Type, formatDateTime(outage.Timestamp), outage.Description)
	b.factor("Evento de parada programada asociado", weightOutageEvent)
	b.factor("Inicio coincide con el evento", timingWeight(outage.Timestamp, deviation.Start))

	if declaredHours := parseDeclaredHours(outage.Description); declaredHours > 0 {
		if math.Abs(float64(declaredHours-deviation.Hours)) <= outageDurationToleranceHr {
			b.evidence(domain.EvidenceEvent, domain.StanceAgainst,
				"La duración observada (%d h) coincide con la duración declarada (%d h).", deviation.Hours, declaredHours)
			b.factor("Duración coincide con la declarada", weightDurationMatches)
		} else {
			b.evidence(domain.EvidenceEvent, domain.StanceSupports,
				"La duración observada (%d h) difiere de la declarada (%d h).", deviation.Hours, declaredHours)
		}
	}
	b.factor("Retorno al baseline tras el evento", weightReturnToBaseline)
}

func classifyOperationalChange(f *finding, b anomalyBuilder, event *domain.Event) {
	deviation := f.deviation
	b.evidence(domain.EvidenceEvent, domain.StanceAgainst, "Evento %s el %s: \"%s\".",
		event.Type, formatDateTime(event.Timestamp), event.Description)
	b.anomaly.Severity = domain.SeverityMedium
	b.anomaly.IsAnomaly = true

	if f.hasElectricalChange() {
		b.anomaly.Type = domain.RealAnomaly
		b.factor("Desviación vs baseline", weightDeviationWithEvent+magnitudeBonus(deviation.MeanDeviation))
		b.factor("Deterioro eléctrico no explicado por el evento", weightUnexplainedElectrical)
		b.factor("Evento operativo explica parte del aumento", weightPartialEventExplanation)
		return
	}
	b.anomaly.Type = domain.ExplainableAnomaly
	b.factor("Evento operativo asociado", weightOperationalEvent)
	b.factor("Inicio coincide con el evento", timingWeight(event.Timestamp, deviation.Start))
	if deviation.Kind == SignalPersistentShift {
		b.factor("Nuevo nivel estable y sostenido", weightStableNewLevel)
	}
	b.factor("Variables eléctricas coherentes con más carga", weightCoherentElectrical)
}

func classifyUnexplained(f *finding, b anomalyBuilder) {
	deviation := f.deviation
	b.anomaly.Type, b.anomaly.IsAnomaly = domain.RealAnomaly, true

	if unknown := domain.FindEventByType(f.events, domain.EventUnknown); unknown != nil {
		b.evidence(domain.EvidenceEvent, domain.StanceSupports,
			"El único evento registrado es %s (\"%s\"): no hay causa operativa conocida.", unknown.Type, unknown.Description)
	} else {
		b.evidence(domain.EvidenceEvent, domain.StanceSupports, "No hay eventos operativos registrados que expliquen el cambio.")
	}

	b.factor("Desviación vs baseline", weightDeviation+magnitudeBonus(deviation.MeanDeviation))
	if deviation.Kind == SignalPersistentShift {
		b.factor(fmt.Sprintf("Persistencia (%d h sostenidas)", deviation.Hours), weightPersistence)
	} else {
		b.factor("Tramo transitorio", weightTransient)
	}
	if f.hasElectricalChange() {
		b.factor("Cambio en variables eléctricas", weightElectricalChange)
	}
	b.factor("Sin evento operativo que lo explique", weightNoExplainingEvent)
	b.factor("Lecturas consistentes (sin fallas de calidad)", weightConsistentReadings)

	magnitude := math.Abs(deviation.MeanDeviation)
	switch {
	case magnitude >= highSeverityDeviation || (f.hasElectricalChange() && deviation.Kind == SignalPersistentShift):
		b.anomaly.Severity = domain.SeverityHigh
	case deviation.Kind == SignalPersistentShift || magnitude >= mediumSeverityDeviation:
		b.anomaly.Severity = domain.SeverityMedium
	default:
		b.anomaly.Severity = domain.SeverityLow
	}
}

func classifyDataQuality(f *finding, b anomalyBuilder) {
	quality := f.dataQuality
	anomaly := b.anomaly
	anomaly.Type, anomaly.IsAnomaly = domain.DataQuality, true
	anomaly.Signal = SignalDataQuality
	anomaly.StartedAt = quality.First
	anomaly.DurationHours = int(quality.Last.Sub(quality.First).Hours()) + 1
	anomaly.Metrics.AffectedReadings = len(quality.Flagged)

	b.evidence(domain.EvidenceDataQuality, domain.StanceSupports,
		"%d de %d lecturas desde %s (%s%%) son inconsistentes, intercaladas con lecturas normales (%d alternancias).",
		len(quality.Flagged), quality.SpanReadings, formatDateTime(quality.First), formatNumber(quality.Fraction*100, 0), quality.Transitions)
	for _, reason := range quality.Reasons {
		b.evidence(domain.EvidenceDataQuality, domain.StanceSupports, "%s en %d lecturas.", capitalize(reason.Reason), reason.Count)
	}
	if quality.VoltageStdDev > 0 {
		b.evidence(domain.EvidenceElectrical, domain.StanceSupports, "Voltaje en lecturas marcadas entre %s y %s V (normal %s ± %s V).",
			formatNumber(quality.MinVoltage, 0), formatNumber(quality.MaxVoltage, 0),
			formatNumber(f.baseline.Voltage.Mean, 1), formatNumber(f.baseline.Voltage.Std, 1))
	}
	if len(quality.RepeatedPowerFactors) > 0 {
		values := make([]string, len(quality.RepeatedPowerFactors))
		for i, value := range quality.RepeatedPowerFactors {
			values[i] = formatNumber(value, 2)
		}
		b.evidence(domain.EvidenceDataQuality, domain.StanceSupports,
			"El factor de potencia repite valores exactos (%s): patrón típico de telemetría corrupta.", strings.Join(values, ", "))
	}
	if quality.VoltageJumps > 0 {
		b.evidence(domain.EvidenceDataQuality, domain.StanceSupports,
			"%d saltos bruscos de voltaje (>%s V) entre lecturas consecutivas.", quality.VoltageJumps, formatNumber(voltageJumpVolts, 0))
	}
	if quality.InvalidReadings > 0 {
		b.evidence(domain.EvidenceDataQuality, domain.StanceSupports, "%d lecturas con valores físicamente imposibles.", quality.InvalidReadings)
	}
	if quality.MissingHours > 0 {
		b.evidence(domain.EvidenceDataQuality, domain.StanceSupports, "%d horas sin lectura.", quality.MissingHours)
	}

	consumption := findVariable(f.variables, VariableConsumption)
	consumptionStable := math.Abs(consumption.DeltaPct) < stableConsumptionMaxChange
	if consumptionStable {
		b.evidence(domain.EvidenceBaseline, domain.StanceSupports,
			"El consumo se mantiene estable (%s vs baseline): el problema está en la medición, no en la carga.",
			formatSignedPercent(consumption.DeltaPct))
	}
	qualityEvent := domain.FindEventByType(f.events, domain.EventDataQuality)
	if qualityEvent != nil {
		b.evidence(domain.EvidenceEvent, domain.StanceSupports, "Evento %s el %s: \"%s\".",
			qualityEvent.Type, formatDateTime(qualityEvent.Timestamp), qualityEvent.Description)
	}

	b.factor("Lecturas inconsistentes intermitentes", weightIntermittentReadings)
	if quality.Fraction >= highAffectedShare {
		b.factor("Proporción alta de lecturas afectadas", weightHighAffectedShare)
	}
	if consumptionStable {
		b.factor("Consumo estable mientras V/FP son erráticos", weightStableConsumption)
	}
	if len(quality.RepeatedPowerFactors) > 0 || quality.VoltageJumps > 0 {
		b.factor("Patrón de telemetría corrupta", weightCorruptTelemetry)
	}
	if qualityEvent != nil {
		b.factor("Evento de calidad de datos reportado", weightDataQualityEvent)
	}

	if quality.Fraction >= highAffectedShare || quality.InvalidReadings > 0 {
		anomaly.Severity = domain.SeverityHigh
	} else {
		anomaly.Severity = domain.SeverityMedium
	}
}

func parseDeclaredHours(description string) int {
	words := strings.Fields(strings.ToLower(description))
	for i := 0; i+1 < len(words); i++ {
		var hours int
		if _, err := fmt.Sscanf(words[i], "%d", &hours); err != nil || hours <= 0 {
			continue
		}
		unit := words[i+1]
		if strings.HasPrefix(unit, "hour") || strings.HasPrefix(unit, "hora") || unit == "h" {
			return hours
		}
	}
	return 0
}
