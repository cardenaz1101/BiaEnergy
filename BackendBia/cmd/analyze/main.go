package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/biaenergy/backend/internal/engine"
	"github.com/biaenergy/backend/internal/storage"
)

type anomalyOutput struct {
	MeterID           string  `json:"meter_id"`
	Anomaly           bool    `json:"anomaly"`
	Type              string  `json:"type"`
	Severity          string  `json:"severity"`
	Confidence        float64 `json:"confidence"`
	Priority          int     `json:"priority"`
	Reason            string  `json:"reason"`
	RecommendedAction string  `json:"recommended_action"`
}

func main() {
	dataDir := flag.String("data", "data", "carpeta con readings.csv, events.csv y meters.csv")
	asJSON := flag.Bool("json", false, "imprimir el resultado como JSON")
	verbose := flag.Bool("v", false, "mostrar etapas y evidencia")
	flag.Parse()

	store, err := loadStore(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	report := func(stepKey, detail string) error {
		if *verbose && !*asJSON {
			fmt.Printf("✓ %-15s %s\n", stepKey, detail)
		}
		return nil
	}
	result, err := engine.New(engine.DefaultConfig()).Analyze(context.Background(), "cli", engine.Input{
		Meters:   store.ListMeters(),
		Readings: store.ReadingsByMeter(),
		Events:   store.ListEvents(""),
	}, report)
	if err != nil {
		log.Fatal(err)
	}

	if *asJSON {
		printJSON(result.Anomalies)
		return
	}
	printTable(result, store.ListMeters(), *verbose)
}

func loadStore(dataDir string) (*storage.Store, error) {
	readings, _, err := storage.LoadReadings(filepath.Join(dataDir, "readings.csv"))
	if err != nil {
		return nil, err
	}
	events, err := storage.LoadEvents(filepath.Join(dataDir, "events.csv"))
	if err != nil {
		return nil, err
	}
	catalog, err := storage.LoadMeterCatalog(filepath.Join(dataDir, "meters.csv"))
	if err != nil {
		return nil, err
	}
	return storage.NewStore(readings, events, catalog, "")
}

func printJSON(anomalies []domain.Anomaly) {
	output := make([]anomalyOutput, 0, len(anomalies))
	for _, anomaly := range anomalies {
		output = append(output, anomalyOutput{
			MeterID:           anomaly.MeterID,
			Anomaly:           anomaly.IsAnomaly,
			Type:              string(anomaly.Type),
			Severity:          string(anomaly.Severity),
			Confidence:        anomaly.Confidence,
			Priority:          anomaly.PriorityRank,
			Reason:            anomaly.Reason,
			RecommendedAction: anomaly.RecommendedAction,
		})
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(output)
}

func printTable(result engine.Result, meters []domain.Meter, verbose bool) {
	fmt.Printf("\n%-3s %-7s %-20s %-7s %-6s %-6s %s\n", "#", "Medidor", "Tipo", "Sev.", "Conf.", "Score", "Razón")
	for _, anomaly := range result.Anomalies {
		fmt.Printf("%-3d %-7s %-20s %-7s %-6.2f %-6.1f %s\n", anomaly.PriorityRank, anomaly.MeterID, anomaly.Type,
			anomaly.Severity, anomaly.Confidence, anomaly.PriorityScore, anomaly.Reason)
		if verbose {
			for _, evidence := range anomaly.Evidence {
				fmt.Printf("      [%s] %s\n", evidence.Stance, evidence.Description)
			}
			fmt.Printf("      → %s\n", anomaly.RecommendedAction)
		}
	}
	fmt.Println("\nEstado de medidores:")
	for _, meter := range meters {
		fmt.Printf("  %s %s\n", meter.MeterID, result.MeterStatuses[meter.MeterID])
	}
}
