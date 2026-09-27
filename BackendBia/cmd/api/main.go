package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/biaenergy/backend/internal/auth"
	"github.com/biaenergy/backend/internal/config"
	"github.com/biaenergy/backend/internal/engine"
	"github.com/biaenergy/backend/internal/httpapi"
	"github.com/biaenergy/backend/internal/service"
	"github.com/biaenergy/backend/internal/storage"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("el servidor se detuvo con error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	store, err := loadStore(cfg, logger)
	if err != nil {
		return err
	}
	analyzer := engine.New(engine.DefaultConfig())
	meters := service.NewMeterService(store, analyzer)
	server := httpapi.NewServer(httpapi.Dependencies{
		Auth: auth.NewService(cfg.JWTSecret, auth.Credentials{
			Email:    cfg.DemoEmail,
			Password: cfg.DemoPassword,
			Name:     cfg.DemoName,
		}, cfg.TokenTTL),
		Meters:         meters,
		Anomalies:      service.NewAnomalyService(store),
		Analyses:       service.NewAnalysisService(store, analyzer, cfg.AnalysisStepDelay, logger),
		Dashboard:      service.NewDashboardService(meters, store, store),
		AllowedOrigins: cfg.AllowedOrigins,
	})

	httpServer := &http.Server{Addr: cfg.Address, Handler: server.Routes(), ReadHeaderTimeout: readHeaderTimeout}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API escuchando", "address", cfg.Address)
		serverErrors <- httpServer.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			return err
		}
		logger.Info("servidor detenido")
	}
	return nil
}

func loadStore(cfg config.Config, logger *slog.Logger) (*storage.Store, error) {
	readings, warnings, err := storage.LoadReadings(cfg.ReadingsPath())
	if err != nil {
		return nil, err
	}
	for _, warning := range warnings {
		logger.Warn("lectura descartada", "detail", warning)
	}
	events, err := storage.LoadEvents(cfg.EventsPath())
	if err != nil {
		return nil, err
	}
	catalog, err := storage.LoadMeterCatalog(cfg.MetersPath())
	if err != nil {
		return nil, err
	}
	store, err := storage.NewStore(readings, events, catalog, cfg.StatePath)
	if err != nil {
		return nil, err
	}
	logger.Info("datos cargados", "readings", len(readings), "meters", len(store.ListMeters()), "events", len(events))
	return store, nil
}
