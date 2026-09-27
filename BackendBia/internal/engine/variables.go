package engine

import (
	"fmt"
	"math"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/numeric"
)

const (
	VariableConsumption = "consumption_kwh"
	VariableCurrent     = "current_a"
	VariableVoltage     = "voltage_v"
	VariablePowerFactor = "power_factor"
	VariablePowerRatio  = "power_ratio"
)

const (
	loadChangeSignificantPct       = 25.0
	voltageSignificantSigmas       = 2.0
	powerFactorSignificantMinDelta = 0.05
	powerFactorSignificantSigmas   = 3.0
	powerRatioSignificantSigmas    = 3.0
)

type windowAverages struct {
	consumption, expectedConsumption, consumptionStdDev float64
	current, expectedCurrent                            float64
	voltage, powerFactor, powerRatio                    float64
}

func averageWindow(baseline Baseline, readings []domain.Reading) windowAverages {
	var consumption, expectedConsumption, current, expectedCurrent, voltage, powerFactor, powerRatio []float64
	stdDevSum := 0.0
	for _, reading := range readings {
		consumption = append(consumption, reading.ConsumptionKWh)
		expectedConsumption = append(expectedConsumption, baseline.ExpectedConsumption(reading.Timestamp))
		current = append(current, reading.CurrentA)
		expectedCurrent = append(expectedCurrent, baseline.ExpectedCurrent(reading.Timestamp))
		voltage = append(voltage, reading.VoltageV)
		powerFactor = append(powerFactor, reading.PowerFactor)
		if ratio := reading.PowerRatio(); !math.IsNaN(ratio) {
			powerRatio = append(powerRatio, ratio)
		}
		stdDevSum += baseline.ConsumptionStdDev(reading.Timestamp)
	}
	return windowAverages{
		consumption:         numeric.Mean(consumption),
		expectedConsumption: numeric.Mean(expectedConsumption),
		consumptionStdDev:   stdDevSum / float64(len(readings)),
		current:             numeric.Mean(current),
		expectedCurrent:     numeric.Mean(expectedCurrent),
		voltage:             numeric.Mean(voltage),
		powerFactor:         numeric.Mean(powerFactor),
		powerRatio:          numeric.Mean(powerRatio),
	}
}

func compareVariables(baseline Baseline, readings []domain.Reading) []domain.VariableChange {
	if len(readings) == 0 {
		return nil
	}
	avg := averageWindow(baseline, readings)
	consumptionChange := numeric.PercentChange(avg.consumption, avg.expectedConsumption)
	currentChange := numeric.PercentChange(avg.current, avg.expectedCurrent)

	return []domain.VariableChange{
		{
			Variable: VariableConsumption, Label: "Consumo", Unit: "kWh/h",
			Baseline: numeric.Round(avg.expectedConsumption, 2), Observed: numeric.Round(avg.consumption, 2),
			DeltaPct:    numeric.Round(consumptionChange, 1),
			ZScore:      numeric.Round((avg.consumption-avg.expectedConsumption)/avg.consumptionStdDev, 1),
			Significant: math.Abs(consumptionChange) >= loadChangeSignificantPct,
		},
		{
			Variable: VariableCurrent, Label: "Corriente", Unit: "A",
			Baseline: numeric.Round(avg.expectedCurrent, 1), Observed: numeric.Round(avg.current, 1),
			DeltaPct:    numeric.Round(currentChange, 1),
			ZScore:      numeric.Round((avg.current-avg.expectedCurrent)/baseline.Current.Std, 1),
			Significant: math.Abs(currentChange) >= loadChangeSignificantPct,
		},
		{
			Variable: VariableVoltage, Label: "Voltaje", Unit: "V",
			Baseline: numeric.Round(baseline.Voltage.Mean, 1), Observed: numeric.Round(avg.voltage, 1),
			DeltaPct:    numeric.Round(numeric.PercentChange(avg.voltage, baseline.Voltage.Mean), 1),
			ZScore:      numeric.Round((avg.voltage-baseline.Voltage.Mean)/baseline.Voltage.Std, 1),
			Significant: math.Abs(avg.voltage-baseline.Voltage.Mean) >= voltageSignificantSigmas*baseline.Voltage.Std,
		},
		{
			Variable: VariablePowerFactor, Label: "Factor de potencia", Unit: "",
			Baseline: numeric.Round(baseline.PowerFactor.Mean, 3), Observed: numeric.Round(avg.powerFactor, 3),
			DeltaPct: numeric.Round(numeric.PercentChange(avg.powerFactor, baseline.PowerFactor.Mean), 1),
			ZScore:   numeric.Round((avg.powerFactor-baseline.PowerFactor.Mean)/baseline.PowerFactor.Std, 1),
			Significant: math.Abs(avg.powerFactor-baseline.PowerFactor.Mean) >=
				math.Max(powerFactorSignificantMinDelta, powerFactorSignificantSigmas*baseline.PowerFactor.Std),
		},
		{
			Variable: VariablePowerRatio, Label: "Coherencia kWh vs V·I·FP", Unit: "",
			Baseline: numeric.Round(baseline.PowerRatio.Mean, 3), Observed: numeric.Round(avg.powerRatio, 3),
			DeltaPct:    numeric.Round(numeric.PercentChange(avg.powerRatio, baseline.PowerRatio.Mean), 1),
			ZScore:      numeric.Round((avg.powerRatio-baseline.PowerRatio.Mean)/baseline.PowerRatio.Std, 1),
			Significant: math.Abs(avg.powerRatio-baseline.PowerRatio.Mean) >= powerRatioSignificantSigmas*baseline.PowerRatio.Std,
		},
	}
}

func findVariable(changes []domain.VariableChange, name string) domain.VariableChange {
	for _, change := range changes {
		if change.Variable == name {
			return change
		}
	}
	return domain.VariableChange{}
}

func electricalChangeReasons(changes []domain.VariableChange) []string {
	var reasons []string
	if powerFactor := findVariable(changes, VariablePowerFactor); powerFactor.Significant {
		reasons = append(reasons, fmt.Sprintf("factor de potencia %s → %s",
			formatNumber(powerFactor.Baseline, 2), formatNumber(powerFactor.Observed, 2)))
	}
	if voltage := findVariable(changes, VariableVoltage); voltage.Significant {
		reasons = append(reasons, fmt.Sprintf("voltaje %s → %s V",
			formatNumber(voltage.Baseline, 1), formatNumber(voltage.Observed, 1)))
	}
	return reasons
}
