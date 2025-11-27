// Package web provides a lightweight HTTP framework with middleware support.
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// HandlerFunc is the type for HTTP handlers in this framework.
type HandlerFunc func(ctx context.Context, r *http.Request) Encoder

// Middleware is a function that wraps a HandlerFunc.
type Middleware func(HandlerFunc) HandlerFunc

// Encoder defines behavior for encoding responses.
type Encoder interface {
	Encode() (data []byte, contentType string, error error)
}

// App is the entry point into our application.
type App struct {
	mux    *http.ServeMux
	mw     []Middleware
	tracer trace.Tracer
}

// NewApp creates an App with the specified middleware.
func NewApp(tracer trace.Tracer, mw ...Middleware) *App {
	return &App{
		mux:    http.NewServeMux(),
		mw:     mw,
		tracer: tracer,
	}
}

// ServeHTTP implements the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// HandlerFunc registers a handler function with middleware.
func (a *App) HandlerFunc(method, group, path string, hdl HandlerFunc, mw ...Middleware) {
	// Wrap with route-specific middleware first
	hdl = wrapMiddleware(mw, hdl)

	// Then wrap with app-level middleware
	hdl = wrapMiddleware(a.mw, hdl)

	// Convert to http.HandlerFunc
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := setWriter(r.Context(), w)

		// Add tracing span if tracer is available
		if a.tracer != nil {
			var span trace.Span
			ctx, span = a.tracer.Start(ctx, path)
			defer span.End()
		}

		// Generate trace ID if not present
		if getTraceID(ctx) == "" {
			ctx = setTraceID(ctx, generateTraceID())
		}

		// Call the handler
		resp := hdl(ctx, r)

		// Write the response
		if err := Respond(ctx, w, resp); err != nil {
			// Log error but don't fail - response may already be written
		}
	}

	// Register with the mux
	pattern := fmt.Sprintf("%s %s%s", method, group, path)
	a.mux.HandleFunc(pattern, handler)
}

// HandlerFuncNoMid registers a handler without app-level middleware.
func (a *App) HandlerFuncNoMid(method, group, path string, hdl HandlerFunc) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := setWriter(r.Context(), w)

		// Generate trace ID
		if getTraceID(ctx) == "" {
			ctx = setTraceID(ctx, generateTraceID())
		}

		resp := hdl(ctx, r)

		if err := Respond(ctx, w, resp); err != nil {
			// Error already logged by middleware
		}
	}

	pattern := fmt.Sprintf("%s %s%s", method, group, path)
	a.mux.HandleFunc(pattern, handler)
}

// wrapMiddleware chains middleware around a handler.
func wrapMiddleware(mw []Middleware, handler HandlerFunc) HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		mwFunc := mw[i]
		if mwFunc != nil {
			handler = mwFunc(handler)
		}
	}
	return handler
}

// Respond encodes and writes the response.
func Respond(ctx context.Context, w http.ResponseWriter, resp Encoder) error {
	// Handle nil responses as 204 No Content
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// Encode the response
	data, contentType, err := resp.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("encode: %w", err)
	}

	// Set content type
	w.Header().Set("Content-Type", contentType)

	// Get status code if available
	statusCode := http.StatusOK
	if v, ok := resp.(interface{ HTTPStatus() int }); ok {
		statusCode = v.HTTPStatus()
	}

	// Store status code in context values
	if v := GetValues(ctx); v != nil {
		v.StatusCode = statusCode
	}

	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// Decode decodes the request body into the provided value.
func Decode(r *http.Request, val any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(val); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	return nil
}

// Param extracts a path parameter from the request.
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

// =============================================================================

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

type ctxKey int

const (
	key ctxKey = iota
	writerKey
	traceKey
)

// SetValues stores the Values in the context.
func SetValues(ctx context.Context, v *Values) context.Context {
	return context.WithValue(ctx, key, v)
}

// GetValues retrieves the Values from the context.
func GetValues(ctx context.Context) *Values {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return &Values{
			TraceID: generateTraceID(),
			Now:     time.Now(),
		}
	}
	return v
}

func setWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, writerKey, w)
}

func GetWriter(ctx context.Context) http.ResponseWriter {
	w, ok := ctx.Value(writerKey).(http.ResponseWriter)
	if !ok {
		return nil
	}
	return w
}

func setTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey, traceID)
}

func getTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(traceKey).(string)
	if !ok {
		return ""
	}
	return traceID
}

// GetTraceID returns the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	return getTraceID(ctx)
}

func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// =============================================================================

// JSONResponse is a simple JSON response encoder.
type JSONResponse struct {
	Data any
}

func (r JSONResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r.Data)
	if err != nil {
		return nil, "", fmt.Errorf("marshal json: %w", err)
	}
	return data, "application/json", nil
}
