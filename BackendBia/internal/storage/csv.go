package storage

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/domain"
)

type MeterInfo struct {
	Name     string
	Location string
}

const byteOrderMark = "\uFEFF"

var timeLayouts = []string{"2006-01-02 15:04:05", "2006-01-02 15:04", time.RFC3339, "2006-01-02T15:04:05"}

func parseTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range timeLayouts {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de fecha no soportado: %q", value)
}

type csvTable struct {
	reader  *csv.Reader
	columns map[string]int
	line    int
}

func openTable(path string, required ...string) (*csvTable, func() error, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("%s: leyendo encabezado: %w", path, err)
	}
	columns := map[string]int{}
	for i, name := range header {
		columns[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, byteOrderMark)))] = i
	}
	for _, name := range required {
		if _, ok := columns[name]; !ok {
			file.Close()
			return nil, nil, fmt.Errorf("%s: falta la columna %q", path, name)
		}
	}
	return &csvTable{reader: reader, columns: columns, line: 1}, file.Close, nil
}

func (t *csvTable) next() ([]string, error) {
	t.line++
	return t.reader.Read()
}

func (t *csvTable) value(record []string, column string) string {
	index, ok := t.columns[column]
	if !ok || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func LoadReadings(path string) ([]domain.Reading, []string, error) {
	table, closeFile, err := openTable(path, "meter_id", "timestamp", "consumption_kwh", "voltage_v", "current_a", "power_factor")
	if err != nil {
		return nil, nil, err
	}
	defer closeFile()

	var readings []domain.Reading
	var warnings []string
	for {
		record, err := table.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("línea %d: %v", table.line, err))
			continue
		}
		reading, err := parseReading(table, record)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("línea %d: %v", table.line, err))
			continue
		}
		reading.ID = int64(len(readings) + 1)
		readings = append(readings, reading)
	}
	sort.SliceStable(readings, func(i, j int) bool {
		if readings[i].MeterID != readings[j].MeterID {
			return readings[i].MeterID < readings[j].MeterID
		}
		return readings[i].Timestamp.Before(readings[j].Timestamp)
	})
	return readings, warnings, nil
}

func parseReading(table *csvTable, record []string) (domain.Reading, error) {
	timestamp, err := parseTimestamp(table.value(record, "timestamp"))
	if err != nil {
		return domain.Reading{}, err
	}
	numbers := map[string]float64{}
	for _, column := range []string{"consumption_kwh", "voltage_v", "current_a", "power_factor"} {
		value, err := strconv.ParseFloat(table.value(record, column), 64)
		if err != nil {
			return domain.Reading{}, fmt.Errorf("%s inválido", column)
		}
		numbers[column] = value
	}
	status := table.value(record, "status")
	if status == "" {
		status = "OK"
	}
	return domain.Reading{
		MeterID:        table.value(record, "meter_id"),
		Timestamp:      timestamp,
		ConsumptionKWh: numbers["consumption_kwh"],
		VoltageV:       numbers["voltage_v"],
		CurrentA:       numbers["current_a"],
		PowerFactor:    numbers["power_factor"],
		Status:         status,
	}, nil
}

func LoadEvents(path string) ([]domain.Event, error) {
	table, closeFile, err := openTable(path, "meter_id", "event_timestamp", "event_type", "description")
	if err != nil {
		return nil, err
	}
	defer closeFile()

	var events []domain.Event
	for {
		record, err := table.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("línea %d: %w", table.line, err)
		}
		timestamp, err := parseTimestamp(table.value(record, "event_timestamp"))
		if err != nil {
			return nil, fmt.Errorf("línea %d: %w", table.line, err)
		}
		events = append(events, domain.Event{
			ID:          fmt.Sprintf("EVT-%03d", len(events)+1),
			MeterID:     table.value(record, "meter_id"),
			Timestamp:   timestamp,
			Type:        strings.ToUpper(table.value(record, "event_type")),
			Description: table.value(record, "description"),
		})
	}
	return events, nil
}

func LoadMeterCatalog(path string) (map[string]MeterInfo, error) {
	catalog := map[string]MeterInfo{}
	table, closeFile, err := openTable(path, "meter_id", "name", "location")
	if errors.Is(err, os.ErrNotExist) {
		return catalog, nil
	}
	if err != nil {
		return nil, err
	}
	defer closeFile()

	for {
		record, err := table.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("línea %d: %w", table.line, err)
		}
		catalog[table.value(record, "meter_id")] = MeterInfo{Name: table.value(record, "name"), Location: table.value(record, "location")}
	}
	return catalog, nil
}
