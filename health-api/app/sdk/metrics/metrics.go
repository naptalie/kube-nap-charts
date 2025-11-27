// Package metrics provides application metrics collection using expvar.
package metrics

import (
	"context"
	"expvar"
	"runtime"
)

// metrics holds the application metrics.
type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

// m is the global metrics instance.
var m *metrics

func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

// ctxKey is the type for metrics context key.
type ctxKey int

const key ctxKey = 1

// Set adds the metrics to the context.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// AddRequests increments the requests counter.
func AddRequests(ctx context.Context) int64 {
	v, ok := ctx.Value(key).(*metrics)
	if !ok {
		return 0
	}

	v.requests.Add(1)
	return v.requests.Value()
}

// AddErrors increments the errors counter.
func AddErrors(ctx context.Context) int64 {
	v, ok := ctx.Value(key).(*metrics)
	if !ok {
		return 0
	}

	v.errors.Add(1)
	return v.errors.Value()
}

// AddPanics increments the panics counter.
func AddPanics(ctx context.Context) int64 {
	v, ok := ctx.Value(key).(*metrics)
	if !ok {
		return 0
	}

	v.panics.Add(1)
	return v.panics.Value()
}

// AddGoroutines sets the current number of goroutines.
func AddGoroutines(ctx context.Context) int64 {
	v, ok := ctx.Value(key).(*metrics)
	if !ok {
		return 0
	}

	v.goroutines.Set(int64(runtime.NumGoroutine()))
	return v.goroutines.Value()
}
