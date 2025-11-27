package mid

import (
	"context"
	"net/http"

	"health-api/foundation/web"
)

// Cors adds CORS headers to responses.
func Cors(origin string) web.Middleware {
	m := func(handler web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			// Get the response writer
			w := web.GetWriter(ctx)
			if w != nil {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				// Handle preflight requests
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return nil
				}
			}

			return handler(ctx, r)
		}
		return h
	}
	return m
}
