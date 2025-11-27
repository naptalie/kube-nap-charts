package mid

import (
	"context"
	"net/http"
	"runtime/debug"

	"health-api/app/sdk/errs"
	"health-api/app/sdk/metrics"
	"health-api/foundation/web"
)

// Panics recovers from panics and converts them to errors.
func Panics() web.Middleware {
	m := func(handler web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) (resp web.Encoder) {
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					resp = errs.Newf(errs.Internal, "panic: %v\n%s", rec, string(trace))
					metrics.AddPanics(ctx)
					PrometheusPanic()
				}
			}()

			return handler(ctx, r)
		}
		return h
	}
	return m
}
