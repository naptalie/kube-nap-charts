// Package mux provides HTTP server initialization and configuration.
package mux

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	"health-api/app/sdk/mid"
	"health-api/foundation/logger"
	"health-api/foundation/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
)

// Config contains dependencies needed to construct the server.
type Config struct {
	Log    *logger.Logger
	Tracer trace.Tracer
}

// RouteAdder defines the interface for adding routes to the app.
type RouteAdder interface {
	Add(app *web.App, cfg Config)
}

// WebAPI constructs an HTTP server with the specified configuration.
func WebAPI(cfg Config, routeAdder RouteAdder, corsOrigin string) *web.App {
	// Create app with middleware stack
	app := web.NewApp(
		cfg.Tracer,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Prometheus(),
		mid.Metrics(),
		mid.Panics(),
		mid.Cors(corsOrigin),
	)

	// Add routes via route adder
	if routeAdder != nil {
		routeAdder.Add(app, cfg)
	}

	return app
}

// DebugMux registers debug and profiling routes.
func DebugMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Register expvar handler
	mux.Handle("/debug/vars", expvar.Handler())

	// Register Prometheus metrics handler
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
