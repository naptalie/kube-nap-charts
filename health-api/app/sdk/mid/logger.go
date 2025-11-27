// Package mid provides HTTP middleware.
package mid

import (
	"context"
	"net/http"
	"time"

	"health-api/foundation/logger"
	"health-api/foundation/web"
)

// Logger logs each request.
func Logger(log *logger.Logger) web.Middleware {
	m := func(handler web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			v := &web.Values{
				TraceID: web.GetTraceID(ctx),
				Now:     time.Now(),
			}
			ctx = web.SetValues(ctx, v)

			log.Info(ctx, "request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)

			resp := handler(ctx, r)

			log.Info(ctx, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", v.StatusCode,
				"duration", time.Since(v.Now).String(),
			)

			return resp
		}
		return h
	}
	return m
}
