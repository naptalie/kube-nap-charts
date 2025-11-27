// Package healthapp provides HTTP handlers for health check endpoints.
package healthapp

import (
	"context"
	"net/http"

	"health-api/app/sdk/errs"
	"health-api/business/domain/healthbus"
	"health-api/foundation/logger"
	"health-api/foundation/web"
)

// App handles health check HTTP requests.
type App struct {
	log       *logger.Logger
	healthBus *healthbus.Business
}

// NewApp constructs a new health app.
func NewApp(log *logger.Logger, healthBus *healthbus.Business) *App {
	return &App{
		log:       log,
		healthBus: healthBus,
	}
}

// QueryHealthChecks handles GET /api/v1/health requests.
func (a *App) QueryHealthChecks(ctx context.Context, r *http.Request) web.Encoder {
	summary, err := a.healthBus.QueryHealthChecks(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query health checks: %s", err)
	}

	return web.JSONResponse{Data: summary}
}

// QueryHealthCheckByTarget handles GET /api/v1/health/{target} requests.
func (a *App) QueryHealthCheckByTarget(ctx context.Context, r *http.Request) web.Encoder {
	target := web.Param(r, "target")
	if target == "" {
		return errs.Newf(errs.InvalidArgument, "target parameter required")
	}

	check, err := a.healthBus.QueryHealthCheckByTarget(ctx, target)
	if err != nil {
		return errs.Newf(errs.NotFound, "health check not found: %s", err)
	}

	return web.JSONResponse{Data: check}
}

// QueryAlerts handles GET /api/v1/alerts requests.
func (a *App) QueryAlerts(ctx context.Context, r *http.Request) web.Encoder {
	summary, err := a.healthBus.QueryAlerts(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query alerts: %s", err)
	}

	return web.JSONResponse{Data: summary}
}

// Readiness handles GET /readiness requests.
func (a *App) Readiness(ctx context.Context, r *http.Request) web.Encoder {
	data := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	return web.JSONResponse{Data: data}
}

// Liveness handles GET /liveness requests.
func (a *App) Liveness(ctx context.Context, r *http.Request) web.Encoder {
	data := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	return web.JSONResponse{Data: data}
}
