package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/biaenergy/backend/internal/auth"
	"github.com/biaenergy/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const (
	requestTimeout = 30 * time.Second
	corsMaxAge     = 300
)

type Dependencies struct {
	Auth           *auth.Service
	Meters         *service.MeterService
	Anomalies      *service.AnomalyService
	Analyses       *service.AnalysisService
	Dashboard      *service.DashboardService
	AllowedOrigins []string
}

type Server struct {
	Dependencies
}

func NewServer(deps Dependencies) *Server {
	return &Server{Dependencies: deps}
}

func (s *Server) Routes() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(requestTimeout))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: trimmed(s.AllowedOrigins),
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodOptions},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         corsMaxAge,
	}))

	router.Route("/api", func(api chi.Router) {
		api.Get("/health", s.health)
		api.Post("/auth/login", s.login)

		api.Group(func(protected chi.Router) {
			protected.Use(s.requireAuth)
			protected.Get("/auth/me", s.currentUser)
			protected.Get("/dashboard/summary", s.dashboardSummary)
			protected.Get("/events", s.listEvents)

			protected.Route("/meters", func(meters chi.Router) {
				meters.Get("/", s.listMeters)
				meters.Get("/{meterID}", s.getMeter)
				meters.Get("/{meterID}/readings", s.listReadings)
			})

			protected.Route("/anomalies", func(anomalies chi.Router) {
				anomalies.Get("/", s.listAnomalies)
				anomalies.Get("/{anomalyID}", s.getAnomaly)
				anomalies.Patch("/{anomalyID}", s.updateAnomaly)
			})

			protected.Route("/ai", func(ai chi.Router) {
				ai.Post("/analyze", s.startAnalysis)
				ai.Get("/analysis", s.listAnalyses)
				ai.Get("/analysis/latest", s.latestAnalysis)
				ai.Get("/analysis/{runID}", s.getAnalysis)
			})
		})
	})
	return router
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "meters": len(s.Meters.Summaries())})
}

func trimmed(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSuffix(strings.TrimSpace(value), "/"); value != "" {
			result = append(result, value)
		}
	}
	return result
}
