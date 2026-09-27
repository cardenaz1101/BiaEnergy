package httpapi

import (
	"errors"
	"net/http"

	"github.com/biaenergy/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

const analysisNotFound = "análisis no encontrado"

func (s *Server) startAnalysis(w http.ResponseWriter, _ *http.Request) {
	run, err := s.Analyses.Start()
	switch {
	case errors.Is(err, domain.ErrAnalysisInProgress):
		writeJSON(w, http.StatusConflict, run)
	case err != nil:
		writeServiceError(w, err, analysisNotFound)
	default:
		writeJSON(w, http.StatusAccepted, run)
	}
}

func (s *Server) getAnalysis(w http.ResponseWriter, r *http.Request) {
	run, err := s.Analyses.Get(chi.URLParam(r, "runID"))
	if err != nil {
		writeServiceError(w, err, analysisNotFound)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) latestAnalysis(w http.ResponseWriter, _ *http.Request) {
	run, ok := s.Analyses.Latest()
	if !ok {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) listAnalyses(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Analyses.List())
}
