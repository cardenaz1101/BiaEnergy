package httpapi

import (
	"net/http"

	"github.com/biaenergy/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

const anomalyNotFound = "anomalía no encontrada"

type updateAnomalyRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

func (s *Server) listAnomalies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	writeJSON(w, http.StatusOK, s.Anomalies.List(service.AnomalyQuery{
		Type:     query.Get("type"),
		Severity: query.Get("severity"),
		Status:   query.Get("status"),
		MeterID:  query.Get("meter_id"),
	}))
}

func (s *Server) getAnomaly(w http.ResponseWriter, r *http.Request) {
	anomaly, err := s.Anomalies.Get(chi.URLParam(r, "anomalyID"))
	if err != nil {
		writeServiceError(w, err, anomalyNotFound)
		return
	}
	writeJSON(w, http.StatusOK, anomaly)
}

func (s *Server) updateAnomaly(w http.ResponseWriter, r *http.Request) {
	var request updateAnomalyRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	anomaly, err := s.Anomalies.Update(chi.URLParam(r, "anomalyID"), service.AnomalyUpdate{
		Status: request.Status,
		Note:   request.Note,
		Author: userFromContext(r.Context()).Name,
	})
	if err != nil {
		writeServiceError(w, err, anomalyNotFound)
		return
	}
	writeJSON(w, http.StatusOK, anomaly)
}
