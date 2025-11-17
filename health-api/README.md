# Health API

A lightweight Go API for reading health check data from Prometheus.

## Features

- Query all health checks from blackbox exporter
- Get health status for specific targets
- Query arbitrary Prometheus metrics
- RESTful API with JSON responses
- CORS enabled for browser access

## API Endpoints

### Get All Health Checks
```
GET /api/v1/health
```

Returns a summary of all health checks including total, healthy, and down counts.

**Response:**
```json
{
  "total": 3,
  "healthy": 2,
  "down": 1,
  "unknown": 0,
  "checks": [
    {
      "target": "https://example.com",
      "status": "healthy",
      "last_checked": "2025-11-16T10:00:00Z",
      "probe": "blackbox"
    }
  ]
}
```

### Get Health Check for Specific Target
```
GET /api/v1/health/{target}
```

Returns health check information for a specific target.

**Response:**
```json
{
  "target": "https://example.com",
  "status": "healthy",
  "last_checked": "2025-11-16T10:00:00Z",
  "probe": "blackbox"
}
```

### Query Prometheus Metrics
```
GET /api/v1/metrics/{metric}
```

Query arbitrary Prometheus metrics. The metric parameter should be a valid PromQL query.

**Example:**
```
GET /api/v1/metrics/probe_http_duration_seconds
```

### Liveness Probe
```
GET /healthz
```

Returns HTTP 200 OK if the service is running.

## Environment Variables

- `PROMETHEUS_URL`: URL of the Prometheus server (default: `http://localhost:9090`)
- `PORT`: Port to listen on (default: `8080`)

## Building

```bash
go build -o health-api .
```

## Running Locally

```bash
export PROMETHEUS_URL=http://localhost:9090
export PORT=8080
./health-api
```

## Docker

Build the Docker image:
```bash
docker build -t health-api:latest .
```

Run the container:
```bash
docker run -p 8080:8080 \
  -e PROMETHEUS_URL=http://prometheus:9090 \
  health-api:latest
```

## Kubernetes Deployment

This API is designed to be deployed alongside the `is-it-up-tho` Helm chart, which provides:
- Prometheus instance with blackbox exporter
- Grafana Alloy for metrics processing
- Automatic service discovery

Enable the health API in your values.yaml:
```yaml
healthApi:
  enabled: true
  service:
    type: ClusterIP
    port: 8080
```
