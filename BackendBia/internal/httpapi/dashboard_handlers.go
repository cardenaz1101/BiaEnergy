package httpapi

import "net/http"

func (s *Server) dashboardSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Dashboard.Summary())
}
