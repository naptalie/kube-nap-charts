package mid

import (
	"context"
	"net/http"

	"health-api/app/sdk/metrics"
	"health-api/foundation/web"
)

// Metrics collects request metrics.
func Metrics() web.Middleware {
	m := func(handler web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			ctx = metrics.Set(ctx)

			resp := handler(ctx, r)

			n := metrics.AddRequests(ctx)

			// Sample goroutines every 1000 requests
			if n%1000 == 0 {
				metrics.AddGoroutines(ctx)
			}

			if checkIsError(resp) {
				metrics.AddErrors(ctx)
			}

			return resp
		}
		return h
	}
	return m
}

// checkIsError checks if the response is an error.
func checkIsError(e any) bool {
	if e == nil {
		return false
	}

	_, ok := e.(error)
	return ok
}
