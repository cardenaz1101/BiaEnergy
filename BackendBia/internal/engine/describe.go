package engine

import (
	"fmt"
	"strings"

	"github.com/biaenergy/backend/internal/domain"
)

func describe(f *finding) {
	switch f.anomaly.Type {
	case domain.RealAnomaly:
		describeRealAnomaly(f)
	case domain.ExplainableAnomaly:
		describeExplainableAnomaly(f)
	case domain.FalsePositive:
		describeFalsePositive(f)
	case domain.DataQuality:
		describeDataQuality(f)
	}
}

func describeRealAnomaly(f *finding) {
	anomaly := &f.anomaly
	deviationPct := formatSignedPercent(f.deviation.MeanDeviation * 100)
	magnitude := strings.TrimPrefix(deviationPct, "+")
	isDrop := f.deviation.IsDrop()

	switch {
	case isDrop:
		anomaly.Title = "Caída de consumo sin causa conocida"
	case anomaly.Severity == domain.SeverityHigh:
		anomaly.Title = "Aumento de consumo sin causa conocida"
	default:
		anomaly.Title = "Desviación de consumo no explicada"
	}

	cause := "sin evento operativo conocido"
	if domain.FindEventByType(f.events, domain.EventOperationalChange) != nil {
		cause = "con un evento operativo que no explica el deterioro eléctrico"
	}
	electrical := ""
	if f.hasElectricalChange() {
		electrical = ", con " + strings.Join(f.electricalReasons, " y ")
	}
	if isDrop {
		anomaly.Reason = fmt.Sprintf("Consumo %s frente al baseline %s%s.", deviationPct, cause, electrical)
	} else {
		anomaly.Reason = fmt.Sprintf("Consumo %s por encima del baseline %s%s.", magnitude, cause, electrical)
	}

	anomaly.Explanation = fmt.Sprintf(
		"Desde el %s el medidor %s consume en promedio %s más de lo esperado para cada hora del día y el cambio se ha sostenido %d horas. "+
			"El último día registró %s kWh frente a un baseline de %s kWh/día. ",
		formatDateTime(anomaly.StartedAt), anomaly.MeterID, magnitude, anomaly.DurationHours,
		formatNumber(anomaly.Metrics.CurrentDailyKWh, 0), formatNumber(anomaly.Metrics.BaselineDailyKWh, 0))
	if f.hasElectricalChange() {
		anomaly.Explanation += fmt.Sprintf("No es solo más carga: también cambió el comportamiento eléctrico (%s). "+
			"Una caída del factor de potencia junto con más corriente suele indicar carga inductiva anómala (motores o compresores forzados, "+
			"fallas de aislamiento) o un problema en la instalación. ", strings.Join(f.electricalReasons, "; "))
	}
	anomaly.Explanation += "No hay un evento operativo que lo justifique, por lo que se trata como anomalía real."

	anomaly.RecommendedAction = "Investigar medidor e instalación en sitio de forma prioritaria."
	anomaly.ActionSteps = []string{"Enviar técnico a inspeccionar la instalación y la carga conectada"}
	if f.hasElectricalChange() {
		anomaly.ActionSteps = append(anomaly.ActionSteps,
			"Revisar motores, compresores y bancos de condensadores (caída de FP)",
			"Verificar acometida y tablero por caída de voltaje o conexiones flojas")
	}
	anomaly.ActionSteps = append(anomaly.ActionSteps,
		"Confirmar con operación si hubo cambios no reportados en el proceso",
		fmt.Sprintf("Cuantificar el sobrecosto: %s kWh por encima de lo esperado hasta ahora", formatNumber(anomaly.Metrics.ExcessKWh, 0)))
}

func describeExplainableAnomaly(f *finding) {
	anomaly := &f.anomaly
	deviationPct := formatSignedPercent(f.deviation.MeanDeviation * 100)
	eventText := ""
	if event := domain.FindEventByType(f.events, domain.EventOperationalChange); event != nil {
		eventText = fmt.Sprintf(" coincide con el evento %s del %s (\"%s\")", event.Type, formatDateTime(event.Timestamp), event.Description)
	}

	anomaly.Title = "Aumento explicado por cambio operativo"
	anomaly.Reason = fmt.Sprintf("Consumo %s sobre el baseline desde el %s;%s. Variables eléctricas estables.",
		deviationPct, formatDay(anomaly.StartedAt), eventText)
	anomaly.Explanation = fmt.Sprintf(
		"El consumo subió %s y se estabilizó en un nuevo nivel. La corriente aumentó en la misma proporción y el voltaje y el factor de potencia "+
			"se mantienen normales, lo que corresponde a más carga conectada y no a una falla. El cambio%s, así que es una anomalía explicable: "+
			"requiere validación, no una intervención técnica.", deviationPct, eventText)
	anomaly.RecommendedAction = "Validar con operación y actualizar el baseline del medidor."
	anomaly.ActionSteps = []string{
		"Confirmar con el jefe de planta que la nueva línea está operando",
		"Validar que el consumo adicional está dentro de lo presupuestado para la línea",
		"Actualizar el baseline para que el nuevo nivel no siga generando alertas",
	}
}

func describeFalsePositive(f *finding) {
	anomaly := &f.anomaly
	drop := "del " + formatUnsignedPercent(f.deviation.MeanDeviation*100)
	eventText := "un evento programado"
	if event := domain.FindEventByType(f.events, domain.EventScheduledOutage, domain.EventMaintenance); event != nil {
		eventText = fmt.Sprintf("el evento %s (\"%s\")", event.Type, event.Description)
	}

	anomaly.Title = "Caída explicada por parada programada"
	anomaly.Reason = fmt.Sprintf("Caída %s durante %d h el %s, coincide con %s; el consumo volvió al baseline.",
		drop, anomaly.DurationHours, formatDay(anomaly.StartedAt), eventText)
	anomaly.Explanation = fmt.Sprintf(
		"Estadísticamente hubo una desviación fuerte (caída %s durante %d h), pero coincide en inicio y duración con %s y después el medidor volvió a su perfil normal. "+
			"No hay cambio eléctrico ni impacto persistente: es un falso positivo y no debe escalarse.", drop, anomaly.DurationHours, eventText)
	anomaly.RecommendedAction = "No escalar. Cerrar como falso positivo."
	anomaly.ActionSteps = []string{
		"Cerrar la alerta como falso positivo",
		"Mantener el registro de paradas programadas actualizado para suprimir estas alertas automáticamente",
	}
}

func describeDataQuality(f *finding) {
	anomaly := &f.anomaly
	quality := f.dataQuality

	anomaly.Title = "Lecturas eléctricas inconsistentes"
	anomaly.Reason = fmt.Sprintf("%s%% de las lecturas desde el %s tienen voltaje/FP fuera de rango e incoherentes con el consumo, que permanece estable.",
		formatNumber(quality.Fraction*100, 0), formatDay(quality.First))
	anomaly.Explanation = fmt.Sprintf(
		"El consumo de %s se mantiene en su nivel normal, pero desde el %s una de cada %d lecturas trae valores eléctricos imposibles para la carga: "+
			"voltajes que saltan entre %s y %s V (normal %s V), factores de potencia repetidos y energía que no cuadra con V·I·FP. "+
			"Ese patrón intermitente apunta al medidor o a la telemetría, no a la instalación. Mientras no se corrija, cualquier indicador calculado con estos datos es poco confiable.",
		anomaly.MeterID, formatDateTime(quality.First), quality.ReadingsPerFlag(),
		formatNumber(quality.MinVoltage, 0), formatNumber(quality.MaxVoltage, 0), formatNumber(f.baseline.Voltage.Mean, 0))
	anomaly.RecommendedAction = "Validar medidor y telemetría antes de usar estos datos."
	anomaly.ActionSteps = []string{
		"Revisar transformadores de corriente/potencial (TC/TP) y cableado del medidor",
		"Verificar el canal de comunicación y el firmware del medidor",
		"Marcar las lecturas afectadas como no confiables para facturación y reportes",
		"Si persiste, programar reemplazo o recalibración del medidor",
	}
}
