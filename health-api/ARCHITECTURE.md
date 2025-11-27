# Health API Architecture

Production-ready Go service following Ardanlabs service template patterns with layered architecture, comprehensive observability, and Kubernetes-native deployment.

## Architecture Overview

This service implements a **layered, domain-driven architecture** with clear separation of concerns:

```
┌──────────────────────────────────────────┐
│         HTTP Handlers (app/)             │
│   - Request/Response handling            │
│   - Route registration                   │
│   - Input validation                     │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│       Business Logic (business/)         │
│   - Domain rules                         │
│   - Business operations                  │
│   - Orchestration                        │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│      Data Access (stores/)               │
│   - Grafana API client                   │
│   - Prometheus queries                   │
│   - External service integration         │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│      Foundation (foundation/)            │
│   - Logger, Web framework                │
│   - OpenTelemetry                        │
│   - Utilities                            │
└──────────────────────────────────────────┘
```

## Directory Structure

```
health-api/
├── app/                              # Application layer
│   ├── domain/                       # Domain-specific handlers
│   │   └── healthapp/                # Health check handlers
│   │       ├── healthapp.go          # HTTP handlers
│   │       └── route.go              # Route registration
│   ├── sdk/                          # Application support libraries
│   │   ├── errs/                     # Structured error handling
│   │   │   └── errs.go               # Error codes & HTTP mapping
│   │   ├── metrics/                  # Metrics collection (expvar)
│   │   │   └── metrics.go            # Request/error/panic metrics
│   │   ├── mid/                      # HTTP middleware
│   │   │   ├── logger.go             # Request logging
│   │   │   ├── metrics.go            # Metrics collection
│   │   │   ├── errors.go             # Error handling
│   │   │   ├── panics.go             # Panic recovery
│   │   │   └── cors.go               # CORS headers
│   │   └── mux/                      # Server configuration
│   │       └── mux.go                # HTTP server setup
│   └── services/                     # Service entry points
│       └── health-api/               # Main service
│           └── main.go               # Application bootstrap
│
├── business/                         # Business logic layer
│   ├── domain/                       # Domain logic
│   │   └── healthbus/                # Health check business
│   │       ├── healthbus.go          # Core business logic
│   │       └── stores/               # Data access implementations
│   │           └── grafanastore/     # Grafana API client
│   │               └── grafanastore.go
│   └── sdk/                          # Business support utilities
│
├── foundation/                       # Foundation layer
│   ├── logger/                       # Structured logging
│   │   └── logger.go                 # slog wrapper with trace IDs
│   ├── web/                          # HTTP framework
│   │   └── web.go                    # Request/response handling
│   └── otel/                         # OpenTelemetry
│       └── otel.go                   # Tracing configuration
│
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
├── Dockerfile                        # Container image
└── ARCHITECTURE.md                   # This file
```

## Layered Architecture Principles

### 1. Foundation Layer (`foundation/`)

Provides reusable infrastructure components:

- **Logger**: Structured logging with slog
  - JSON output format
  - Trace ID injection
  - Context-aware logging
  - Multiple log levels (Debug, Info, Warn, Error)
  - Source location tracking for errors

- **Web Framework**: Lightweight HTTP framework
  - Middleware composition
  - Request/response encoding/decoding
  - Context value management
  - Trace ID propagation

- **OpenTelemetry**: Distributed tracing
  - OTLP gRPC exporter
  - Configurable sampling (5% default)
  - Span creation and propagation
  - Optional (gracefully degrades if not configured)

### 2. Business Layer (`business/`)

Contains pure business logic, isolated from HTTP concerns:

- **Domain Models**: Health checks, alerts, status types
- **Business Rules**: Status determination logic
- **Storer Interface**: Abstraction over data access
- **No HTTP Dependencies**: Can be tested independently

### 3. Data Access Layer (`business/domain/healthbus/stores/`)

Implements data access via external services:

- **Grafana Store**: Queries Grafana alert API for health status
- **Interface-based**: Easy to mock for testing
- **Error Handling**: Maps external errors to domain errors

### 4. Application Layer (`app/`)

HTTP-specific concerns:

- **Handlers**: Convert HTTP requests to business calls
- **Routes**: Register endpoints with middleware
- **Middleware**: Cross-cutting concerns (logging, metrics, errors)
- **Error Mapping**: Domain errors → HTTP status codes

## Middleware Stack

Middleware executes in order (outermost to innermost):

```
Request → Logger → Errors → Metrics → Panics → CORS → Handler
          ↓        ↓         ↓         ↓        ↓       ↓
       Log req   Catch    Count     Recover  Add     Execute
                 errors   requests  panics   headers business
                          ↓                          logic
Response ← Log res ← Map to HTTP ← Update metrics ← Return result
```

### Middleware Components

1. **Logger** ([mid/logger.go](app/sdk/mid/logger.go))
   - Logs request start/completion
   - Includes method, path, remote address, duration, status

2. **Errors** ([mid/errors.go](app/sdk/mid/errors.go))
   - Catches errors from handlers
   - Logs with source location
   - Maps error codes to HTTP status
   - Sanitizes internal errors

3. **Metrics** ([mid/metrics.go](app/sdk/mid/metrics.go))
   - Counts requests, errors
   - Samples goroutine count (every 1000 requests)
   - Exposes via `/debug/vars`

4. **Panics** ([mid/panics.go](app/sdk/mid/panics.go))
   - Recovers from panics
   - Converts to structured errors
   - Logs stack trace
   - Increments panic counter

5. **CORS** ([mid/cors.go](app/sdk/mid/cors.go))
   - Adds CORS headers
   - Handles preflight requests
   - Configurable origin (default: `*`)

## Error Handling

Structured errors with HTTP status mapping:

```go
type Error struct {
    Code     ErrCode `json:"code"`
    Message  string  `json:"message"`
    FuncName string  `json:"-"`      // For logging only
    FileName string  `json:"-"`      // For logging only
}
```

### Error Codes

- `InvalidArgument` → 400 Bad Request
- `Unauthenticated` → 401 Unauthorized
- `PermissionDenied` → 403 Forbidden
- `NotFound` → 404 Not Found
- `Internal` → 500 Internal Server Error
- `Unavailable` → 503 Service Unavailable

### Usage Pattern

```go
// In handler
if target == "" {
    return errs.Newf(errs.InvalidArgument, "target parameter required")
}

check, err := a.healthBus.QueryHealthCheckByTarget(ctx, target)
if err != nil {
    return errs.Newf(errs.NotFound, "health check not found: %s", err)
}

return web.JSONResponse{Data: check}
```

## Observability

### Structured Logging

All logs are JSON-formatted with consistent fields:

```json
{
  "timestamp": "2025-11-26T01:52:43Z",
  "level": "INFO",
  "msg": "request completed",
  "service": "HEALTH-API",
  "trace_id": "1732588363123456789",
  "method": "GET",
  "path": "/api/v1/health",
  "status": 200,
  "duration": "45.2ms"
}
```

### Metrics (expvar)

Exposed at `/debug/vars` on port 4000:

```json
{
  "requests": 1234,
  "errors": 5,
  "panics": 0,
  "goroutines": 12
}
```

### OpenTelemetry Tracing

Optional distributed tracing via OTLP gRPC:

```bash
# Enable tracing
export OTEL_REPORTER_URI=otel-collector:4317

# Traces include:
# - HTTP request spans
# - Error details
# - Request metadata
```

### Debug Endpoints

Available on port 4000:

- `/debug/pprof/` - CPU/memory profiling
- `/debug/pprof/heap` - Heap profile
- `/debug/pprof/goroutine` - Goroutine dump
- `/debug/vars` - Expvar metrics

## Configuration

All configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `API_HOST` | `:8080` | API server address |
| `DEBUG_HOST` | `:4000` | Debug server address |
| `CORS_ORIGIN` | `*` | CORS allowed origin |
| `GRAFANA_URL` | - | Grafana base URL |
| `GRAFANA_USER` | `admin` | Grafana username |
| `GRAFANA_PASSWORD` | `admin` | Grafana password |
| `OTEL_REPORTER_URI` | - | OpenTelemetry collector URI |

## API Endpoints

### Health Check Endpoints

```bash
# Get all health checks
GET /api/v1/health
Response: {
  "total": 3,
  "healthy": 2,
  "down": 1,
  "unknown": 0,
  "checks": [
    {
      "target": "https://example.com",
      "status": "healthy",
      "last_checked": "2025-11-26T01:00:00Z",
      "probe": "blackbox"
    }
  ]
}

# Get specific health check
GET /api/v1/health/{target}
Response: {
  "target": "https://example.com",
  "status": "healthy",
  "last_checked": "2025-11-26T01:00:00Z",
  "probe": "blackbox"
}

# Get Grafana alert summary
GET /api/v1/alerts
Response: {
  "total": 5,
  "firing": 1,
  "pending": 0,
  "normal": 4,
  "alerts": [...]
}
```

### Kubernetes Probes

```bash
# Liveness probe
GET /liveness
Response: {"status":"ok"}

# Readiness probe
GET /readiness
Response: {"status":"ok"}

# Legacy health endpoint
GET /healthz
Response: {"status":"ok"}
```

### Debug Endpoints

```bash
# Metrics
GET /debug/vars
Response: {
  "requests": 1234,
  "errors": 5,
  "panics": 0,
  "goroutines": 12
}

# CPU Profile (30 second sample)
GET /debug/pprof/profile?seconds=30

# Heap Profile
GET /debug/pprof/heap

# Goroutine Dump
GET /debug/pprof/goroutine
```

## Building and Running

### Local Development

```bash
# Build
go build -o health-api ./app/services/health-api

# Run
export GRAFANA_URL=http://grafana.monitoring.svc.cluster.local:3000
export GRAFANA_USER=admin
export GRAFANA_PASSWORD=admin
./health-api

# Test
curl http://localhost:8080/liveness
curl http://localhost:8080/api/v1/health
curl http://localhost:4000/debug/vars
```

### Docker Build

```bash
# Build image
docker build -t health-api:latest .

# Run container
docker run -p 8080:8080 -p 4000:4000 \
  -e GRAFANA_URL=http://grafana:3000 \
  -e GRAFANA_USER=admin \
  -e GRAFANA_PASSWORD=admin \
  health-api:latest
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: health-api
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: health-api
        image: health-api:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 4000
          name: debug
        env:
        - name: GRAFANA_URL
          value: "http://grafana:3000"
        - name: GRAFANA_USER
          valueFrom:
            secretKeyRef:
              name: grafana-creds
              key: username
        - name: GRAFANA_PASSWORD
          valueFrom:
            secretKeyRef:
              name: grafana-creds
              key: password
        - name: OTEL_REPORTER_URI
          value: "otel-collector:4317"
        livenessProbe:
          httpGet:
            path: /liveness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readiness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 128Mi
```

## Graceful Shutdown

The service handles shutdown gracefully:

1. Receives SIGINT/SIGTERM signal
2. Stops accepting new requests
3. Drains existing requests (20s timeout)
4. Shuts down OpenTelemetry
5. Exits cleanly

```bash
# Shutdown logs:
{"level":"INFO","msg":"shutdown","status":"shutdown started","signal":"interrupt"}
# ... draining requests ...
{"level":"INFO","msg":"shutdown","status":"shutdown complete","signal":"interrupt"}
```

## Testing

### Unit Tests

```bash
# Test business logic
go test ./business/domain/healthbus/...

# Test handlers
go test ./app/domain/healthapp/...

# Test middleware
go test ./app/sdk/mid/...
```

### Integration Tests

```bash
# Start test dependencies
docker run -d -p 3000:3000 grafana/grafana

# Run integration tests
go test -tags=integration ./...
```

## Performance Considerations

- **Connection Pooling**: HTTP client reuses connections to Grafana
- **Timeouts**: 30s timeout on external requests
- **Goroutine Monitoring**: Tracks goroutine count via metrics
- **Memory Efficient**: Structured logging avoids string concatenation
- **Graceful Degradation**: Works without OpenTelemetry if not configured

## Security

- **No Secrets in Logs**: Credentials only in environment variables
- **CORS Configuration**: Configurable allowed origins
- **Error Sanitization**: Internal errors not exposed to clients
- **Resource Limits**: Kubernetes resource constraints
- **Read/Write Timeouts**: Prevents slowloris attacks

## Future Enhancements

- **Prometheus Metrics**: Native Prometheus `/metrics` endpoint
- **Health Checks**: Database connectivity checks for readiness
- **Rate Limiting**: Per-client request throttling
- **Caching**: Cache health check results (short TTL)
- **Authentication**: JWT/API key support
- **Audit Logging**: Track who accessed which endpoints

## References

- [Ardanlabs Service Template](https://github.com/ardanlabs/service)
- [Go Standard Project Layout](https://github.com/golang-standards/project-layout)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Structured Logging with slog](https://pkg.go.dev/log/slog)
