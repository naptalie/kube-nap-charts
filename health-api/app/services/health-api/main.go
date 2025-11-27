// Package main is the entry point for the health API service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"health-api/app/domain/healthapp"
	"health-api/app/sdk/mux"
	"health-api/business/domain/healthbus"
	"health-api/business/domain/healthbus/stores/grafanastore"
	"health-api/foundation/logger"
	"health-api/foundation/otel"
	"health-api/foundation/web"
)

var build = "develop"

func main() {
	// Initialize logger
	log := logger.New(os.Stdout, logger.LevelInfo, "HEALTH-API", traceIDFunc)

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "build", build)

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		Web struct {
			ReadTimeout     time.Duration
			WriteTimeout    time.Duration
			IdleTimeout     time.Duration
			ShutdownTimeout time.Duration
			APIHost         string
			DebugHost       string
			CORSOrigin      string
		}
		Grafana struct {
			URL      string
			User     string
			Password string
		}
		Otel struct {
			ReporterURI string
			Probability float64
		}
	}{
		Web: struct {
			ReadTimeout     time.Duration
			WriteTimeout    time.Duration
			IdleTimeout     time.Duration
			ShutdownTimeout time.Duration
			APIHost         string
			DebugHost       string
			CORSOrigin      string
		}{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 20 * time.Second,
			APIHost:         getEnv("API_HOST", ":8080"),
			DebugHost:       getEnv("DEBUG_HOST", ":4000"),
			CORSOrigin:      getEnv("CORS_ORIGIN", "*"),
		},
		Grafana: struct {
			URL      string
			User     string
			Password string
		}{
			URL:      getEnv("GRAFANA_URL", ""),
			User:     getEnv("GRAFANA_USER", "admin"),
			Password: getEnv("GRAFANA_PASSWORD", "admin"),
		},
		Otel: struct {
			ReporterURI string
			Probability float64
		}{
			ReporterURI: getEnv("OTEL_REPORTER_URI", ""),
			Probability: 0.05, // 5% sampling
		},
	}

	log.Info(ctx, "startup", "config",
		"api_host", cfg.Web.APIHost,
		"debug_host", cfg.Web.DebugHost,
		"grafana_configured", cfg.Grafana.URL != "",
		"otel_configured", cfg.Otel.ReporterURI != "",
	)

	// -------------------------------------------------------------------------
	// Initialize OpenTelemetry

	tracer, shutdown, err := otel.InitTracing(otel.Config{
		ServiceName: "health-api",
		ReporterURI: cfg.Otel.ReporterURI,
		Probability: cfg.Otel.Probability,
	})
	if err != nil {
		return fmt.Errorf("initializing tracing: %w", err)
	}
	defer shutdown(ctx)

	// -------------------------------------------------------------------------
	// Start Debug Service

	log.Info(ctx, "startup", "status", "debug service started", "host", cfg.Web.DebugHost)

	debugMux := mux.DebugMux()
	debugServer := http.Server{
		Addr:           cfg.Web.DebugHost,
		Handler:        debugMux,
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		IdleTimeout:    cfg.Web.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := debugServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(ctx, "debug server error", "error", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Initialize Business Layer

	grafanaStore := grafanastore.NewStore(log, cfg.Grafana.URL, cfg.Grafana.User, cfg.Grafana.Password)
	healthBus := healthbus.NewBusiness(log, grafanaStore)

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing API", "host", cfg.Web.APIHost)

	// Create route adder
	routeAdder := Routes{
		HealthBus: healthBus,
	}

	// Create API app
	apiApp := mux.WebAPI(mux.Config{
		Log:    log,
		Tracer: tracer,
	}, routeAdder, cfg.Web.CORSOrigin)

	apiServer := http.Server{
		Addr:           cfg.Web.APIHost,
		Handler:        apiApp,
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		IdleTimeout:    cfg.Web.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", apiServer.Addr)
		serverErrors <- apiServer.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdownChan:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := apiServer.Shutdown(ctx); err != nil {
			apiServer.Close()
			return fmt.Errorf("could not stop api server gracefully: %w", err)
		}

		if err := debugServer.Shutdown(ctx); err != nil {
			debugServer.Close()
			return fmt.Errorf("could not stop debug server gracefully: %w", err)
		}
	}

	return nil
}

// Routes implements mux.RouteAdder.
type Routes struct {
	HealthBus *healthbus.Business
}

// Add registers all routes for the service.
func (r Routes) Add(app *web.App, cfg mux.Config) {
	healthapp.Routes(app, healthapp.Config{
		Log:       cfg.Log,
		HealthBus: r.HealthBus,
	})
}

// traceIDFunc extracts the trace ID from the context.
func traceIDFunc(ctx context.Context) string {
	return web.GetTraceID(ctx)
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
