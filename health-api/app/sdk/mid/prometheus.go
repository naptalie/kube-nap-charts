package mid

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"health-api/app/sdk/errs"
	"health-api/app/sdk/metrics"
	"health-api/foundation/web"
)

// Prometheus middleware for collecting HTTP metrics.
func Prometheus() web.Middleware {
	m := func(handler web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			start := time.Now()

			// Call handler
			resp := handler(ctx, r)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Get path and method
			path := r.URL.Path
			method := r.Method

			// Record duration
			metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

			// Get status code
			statusCode := 200
			if resp != nil {
				if v, ok := resp.(interface{ HTTPStatus() int }); ok {
					statusCode = v.HTTPStatus()
				}
			}

			// Record request count
			metrics.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()

			// Record errors if status >= 400
			if statusCode >= 400 {
				errCode := "unknown"
				if e, ok := resp.(*errs.Error); ok {
					errCode = e.Code.String()
				}
				metrics.HTTPErrorsTotal.WithLabelValues(method, path, errCode).Inc()
			}

			return resp
		}
		return h
	}
	return m
}

// PrometheusPanic increments the panic counter.
func PrometheusPanic() {
	metrics.PanicsTotal.Inc()
}
