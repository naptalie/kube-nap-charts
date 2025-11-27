package healthapp

import (
	"net/http"

	"health-api/business/domain/healthbus"
	"health-api/foundation/logger"
	"health-api/foundation/web"
)

// Config contains dependencies needed to construct handlers.
type Config struct {
	Log       *logger.Logger
	HealthBus *healthbus.Business
}

// Routes registers all health check routes.
func Routes(app *web.App, cfg Config) {
	const version = "/api/v1"

	api := NewApp(cfg.Log, cfg.HealthBus)

	// Health check endpoints (with full middleware)
	app.HandlerFunc(http.MethodGet, version, "/health", api.QueryHealthChecks)
	app.HandlerFunc(http.MethodGet, version, "/health/{target}", api.QueryHealthCheckByTarget)
	app.HandlerFunc(http.MethodGet, version, "/alerts", api.QueryAlerts)

	// Liveness and readiness probes (no middleware except CORS)
	app.HandlerFuncNoMid(http.MethodGet, "", "/liveness", api.Liveness)
	app.HandlerFuncNoMid(http.MethodGet, "", "/readiness", api.Readiness)
	app.HandlerFuncNoMid(http.MethodGet, "", "/healthz", api.Liveness) // Legacy endpoint
}
