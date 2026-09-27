package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Address           string
	DataDir           string
	StatePath         string
	AllowedOrigins    []string
	JWTSecret         string
	TokenTTL          time.Duration
	DemoEmail         string
	DemoPassword      string
	DemoName          string
	AnalysisStepDelay time.Duration
}

func Load() Config {
	return Config{
		Address:           env("ADDR", ":8080"),
		DataDir:           env("DATA_DIR", "data"),
		StatePath:         env("STATE_PATH", filepath.Join("data", "state.json")),
		AllowedOrigins:    strings.Split(env("CORS_ORIGINS", "*"), ","),
		JWTSecret:         env("JWT_SECRET", "bia-demo-secret-change-me"),
		TokenTTL:          durationEnv("TOKEN_TTL", 12*time.Hour),
		DemoEmail:         env("DEMO_EMAIL", "demo@bia.energy"),
		DemoPassword:      env("DEMO_PASSWORD", "demo123"),
		DemoName:          env("DEMO_NAME", "Laura Gómez"),
		AnalysisStepDelay: durationEnv("ANALYSIS_STEP_DELAY", 550*time.Millisecond),
	}
}

func (c Config) ReadingsPath() string { return filepath.Join(c.DataDir, "readings.csv") }

func (c Config) EventsPath() string { return filepath.Join(c.DataDir, "events.csv") }

func (c Config) MetersPath() string { return filepath.Join(c.DataDir, "meters.csv") }

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if parsed, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return parsed
	}
	return fallback
}
