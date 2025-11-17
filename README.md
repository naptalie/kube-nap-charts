# Kube-Nap Charts

> **Is It Up Tho?** - A comprehensive Kubernetes health monitoring solution

Lightweight, production-ready health check system combining Prometheus, Grafana Alloy, and a custom Go API.

## Features

- **Prometheus Operator Integration** - Declarative Prometheus instances via CRDs
- **Blackbox Exporter** - HTTP/HTTPS/TCP health checks
- **Grafana Alloy** - Metric enrichment with custom labels
- **Go REST API** - Query health status via JSON endpoints
- **Kubernetes Native** - Uses CRDs for probe configuration
- **Production Ready** - HA configuration, resource limits, health checks

## Quick Start

### Local Development with KIND

The fastest way to try out is-it-up-tho locally:

```bash
# Install KIND and create cluster
make kind-setup

# Deploy everything
make kind-deploy

# Access the API
make port-forward-api

# Test it
make test-api
```

See [KIND_SETUP.md](KIND_SETUP.md) for detailed local development guide.

### Production Deployment

```bash
# 1. Build the Health API
cd health-api
docker build -t your-registry/health-api:latest .
docker push your-registry/health-api:latest

# 2. Update Helm dependencies
cd helm-charts/charts/is-it-up-tho
helm dependency update

# 3. Install
helm install my-healthchecks . -n monitoring --create-namespace

# 4. Access the API
kubectl port-forward -n monitoring svc/my-healthchecks-is-it-up-tho-health-api 8080:8080
curl http://localhost:8080/api/v1/health
```

Or use the Makefile:

```bash
# For production with your registry
export REGISTRY=your-registry
make deploy

# For local KIND testing
make kind-setup && make kind-deploy
```

## Architecture

```
External Targets          Kubernetes Cluster
─────────────────         ──────────────────────────────────────────
                         │
https://example.com ◄────┤  ┌─────────────────┐
https://google.com  ◄────┼──│ Blackbox        │
https://api.com     ◄────┤  │ Exporter        │
                         │  └────────┬────────┘
                         │           │
                         │           ▼ scrape
                         │  ┌─────────────────────┐
                         │  │   Prometheus        │◄──┐
                         │  │   (Operator CRD)    │   │
                         │  └──────┬──────────────┘   │
                         │         │                   │
                         │         │ remote_write      │
                         │         ▼                   │
                         │  ┌─────────────────────┐   │
                         │  │  Grafana Alloy      │   │
                         │  │  + Custom Labels:   │   │
                         │  │    • cluster        │   │
                         │  │    • environment    │   │
                         │  │    • source         │   │
                         │  └──────┬──────────────┘   │
                         │         │                   │
                         │         └─ remote_write ────┘
                         │
                         │  ┌──────────────────────┐
Clients ─────────────────┼─►│  Health API (Go)     │
(REST API)              │  │  • /api/v1/health    │
                         │  │  • JSON responses    │
                         │  └──────────────────────┘
                         │
                         ──────────────────────────────────────────
```

## Components

### 1. Prometheus Instance
- Managed by Prometheus Operator
- 30-day retention (configurable)
- Remote write to Alloy
- Service discovery via Probe CRDs

### 2. Blackbox Exporter
- Probes HTTP/HTTPS/TCP endpoints
- Supports multiple probe modules
- Automatic target discovery

### 3. Grafana Alloy
- Receives metrics from Prometheus
- Adds custom labels:
  - `cluster` - Cluster identifier
  - `environment` - Environment name
  - `source` - Set to "alloy"
  - Custom labels (team, region, etc.)
- Writes enriched metrics back to Prometheus

### 4. Health API (Go)
RESTful API for querying health data:

#### Endpoints:
- `GET /api/v1/health` - Get all health checks
- `GET /api/v1/health/{target}` - Get specific target
- `GET /api/v1/metrics/{metric}` - Query Prometheus metrics
- `GET /healthz` - Liveness probe

#### Example Response:
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

## Configuration

### values.yaml

```yaml
# Health check targets
probe:
  blackbox:
    enabled: true
    targets:
      - https://example.com
      - https://google.com

# Custom labels for metrics
alloy:
  enabled: true
  customLabels:
    cluster: "production"
    environment: "production"
    additional:
      team: "platform"
      region: "us-west-2"

# Health API
healthApi:
  enabled: true
  replicaCount: 2
  image:
    repository: your-registry/health-api
    tag: latest
```

## Documentation

- [Quick Start Guide](QUICKSTART.md) - Step-by-step installation
- [Implementation Summary](IMPLEMENTATION_SUMMARY.md) - Architecture and design decisions
- [Helm Chart README](helm-charts/charts/is-it-up-tho/README.md) - Full chart documentation
- [Health API README](health-api/README.md) - API documentation

## Makefile Commands

```bash
make help                    # Show all available commands
make build-api              # Build Health API Docker image
make push-api               # Push image to registry
make deploy                 # Build, push, and install
make helm-install           # Install Helm chart
make helm-upgrade           # Upgrade release
make port-forward-api       # Port forward to API
make port-forward-prometheus# Port forward to Prometheus
make status                 # Check resource status
make logs-api               # View API logs
make logs-prometheus        # View Prometheus logs
make logs-alloy             # View Alloy logs
```

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Prometheus Operator
- Docker (for building Health API)

## Installation

See [QUICKSTART.md](QUICKSTART.md) for detailed installation instructions.

## Examples

### Query Health Status
```bash
# All health checks
curl http://localhost:8080/api/v1/health | jq

# Specific target
curl http://localhost:8080/api/v1/health/https://example.com | jq
```

### Prometheus Queries
```promql
# Current probe status
probe_success

# Failed probes
probe_success == 0

# With custom labels
probe_success{cluster="production", environment="production"}

# HTTP duration
probe_http_duration_seconds
```

## Troubleshooting

### View Component Status
```bash
make status
```

### View Logs
```bash
make logs-api
make logs-prometheus
make logs-alloy
```

### Common Issues

**No metrics showing up:**
```bash
kubectl get probe -n monitoring
kubectl describe prometheus -n monitoring
```

**Health API connection issues:**
```bash
kubectl logs -n monitoring -l component=health-api
kubectl get svc -n monitoring
```

**Alloy not adding labels:**
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=alloy
kubectl get configmap -n monitoring
```

## Project Structure

```
.
├── health-api/                  # Go REST API
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   └── README.md
├── helm-charts/
│   └── charts/
│       └── is-it-up-tho/       # Helm chart
│           ├── Chart.yaml
│           ├── values.yaml
│           ├── README.md
│           └── templates/
│               ├── rbac.yaml
│               ├── prometheus-crd.yaml
│               ├── probe-crd.yaml
│               ├── alloy-config.yaml
│               ├── service.yaml
│               ├── health-api-deployment.yaml
│               ├── configmap.yaml
│               └── NOTES.txt
├── Makefile                    # Build and deployment automation
├── QUICKSTART.md               # Installation guide
├── IMPLEMENTATION_SUMMARY.md   # Architecture details
└── README.md                   # This file
```

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

[Your License Here]

## Support

- Documentation: See `/docs` directory
- Issues: [GitHub Issues](https://github.com/your-repo/kube-nap-charts/issues)

---

**Built with observability best practices**
