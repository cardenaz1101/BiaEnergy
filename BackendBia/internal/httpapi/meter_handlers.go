package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

const meterNotFound = "medidor no encontrado"

var queryTimeLayouts = []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"}

func parseMeterFilter(value string) service.MeterFilter {
	switch strings.ToLower(value) {
	case "normal", "ok":
		return service.FilterNormal
	case "alert", "alerts":
		return service.FilterAlert
	case "critical":
		return service.FilterCritical
	default:
		return service.FilterAll
	}
}

func parseQueryTime(value string) time.Time {
	for _, layout := range queryTimeLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func meterIDParam(r *http.Request) string {
	return strings.ToUpper(chi.URLParam(r, "meterID"))
}

func (s *Server) listMeters(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, s.Meters.List(service.MeterQuery{
		Filter:     parseMeterFilter(query.Get("status")),
		Search:     query.Get("q"),
		SortBy:     query.Get("sort"),
		Descending: query.Get("order") != "asc",
	}))
}

func (s *Server) getMeter(w http.ResponseWriter, r *http.Request) {
	detail, err := s.Meters.Get(meterIDParam(r))
	if err != nil {
		writeServiceError(w, err, meterNotFound)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

type readingsResponse struct {
	MeterID  string                 `json:"meter_id"`
	Count    int                    `json:"count"`
	Readings []service.ReadingPoint `json:"readings"`
}

func (s *Server) listReadings(w http.ResponseWriter, r *http.Request) {
	meterID := meterIDParam(r)
	query := r.URL.Query()
	points, err := s.Meters.Readings(meterID, service.ReadingQuery{
		From:  parseQueryTime(query.Get("from")),
		To:    parseQueryTime(query.Get("to")),
		Daily: query.Get("granularity") == "day",
	})
	if err != nil {
		writeServiceError(w, err, meterNotFound)
		return
	}
	writeJSON(w, http.StatusOK, readingsResponse{MeterID: meterID, Count: len(points), Readings: points})
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Meters.Events(strings.ToUpper(r.URL.Query().Get("meter_id"))))
}
