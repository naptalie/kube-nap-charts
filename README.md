# Kube-Nap Charts

> **Is It Up Tho?** - A comprehensive Kubernetes health monitoring solution

Lightweight, production-ready health check system combining Prometheus, Grafana Alloy, and a custom Go API.

## Features

- **Prometheus Operator Integration** - Declarative Prometheus instances via CRDs
- **Blackbox Exporter** - HTTP/HTTPS/TCP health checks
- **Grafana Alloy** - Metric enrichment with custom labels via remote write federation
- **Grafana Operator** - Automated Grafana instance with datasources and alerting
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
─────────────────         ──────────────────────────────────────────────
                         │
https://example.com ◄────┤  ┌─────────────────┐
https://google.com  ◄────┼──│ Blackbox        │
https://api.com     ◄────┤  │ Exporter        │
                         │  └────────┬────────┘
                         │           │
                         │           ▼ scrape
                         │  ┌──────────────────────────┐
                         │  │   Prometheus             │◄──┐
                         │  │   (Operator CRD)         │   │
                         │  │   + Remote Write Receiver│   │
                         │  └──────┬───────────────────┘   │
                         │         │                        │
                         │         │ remote_write           │
                         │         │ /api/v1/metrics/write  │
                         │         ▼                        │
                         │  ┌──────────────────────────┐   │
                         │  │  Grafana Alloy           │   │
                         │  │  Federation Layer        │   │
                         │  │  + Custom Labels:        │   │
                         │  │    • cluster             │   │
                         │  │    • environment         │   │
                         │  │    • source=alloy        │   │
                         │  └──────┬───────────────────┘   │
                         │         │                        │
                         │         └─ remote_write ─────────┘
                         │            /api/v1/write
                         │
                         │  ┌──────────────────────────┐
                         │  │   Grafana                │
                         │  │   (Operator CRD)         │
                         │  │   • Datasources          │
                         │  │   • Alert Rules          │
                         │  │   • Dashboards           │
                         │  └──────────────────────────┘
                         │
                         │  ┌──────────────────────────┐
Clients ─────────────────┼─►│  Health API (Go)         │
(REST API)              │  │  • /api/v1/health        │
                         │  │  • JSON responses        │
                         │  └──────────────────────────┘
                         │
                         ──────────────────────────────────────────────
```

## Components

### 1. Prometheus Instance
- Managed by Prometheus Operator
- 30-day retention (configurable)
- Remote write to Alloy via `/api/v1/metrics/write`
- Remote write receiver enabled for federation loop
- Service discovery via Probe CRDs

### 2. Blackbox Exporter
- Probes HTTP/HTTPS/TCP endpoints
- Supports multiple probe modules (http_2xx, tcp_connect)
- Automatic target discovery

### 3. Grafana Alloy (Federation Layer)
- **Receives metrics** from Prometheus via remote write
- **Enriches metrics** with custom labels:
  - `cluster` - Cluster identifier (e.g., "healthcheck-demo")
  - `environment` - Environment name (e.g., "production")
  - `source` - Set to "alloy" for tracking
  - Custom labels (team, region, etc.)
- **Writes back** enriched metrics to Prometheus
- Acts as a federation/transformation layer

### 4. Grafana Instance
- Managed by Grafana Operator
- Automatic Prometheus datasource configuration
- Pre-configured alert rules for probe failures
- Alert folder organization
- Admin credentials: admin/admin (change in production!)

### 5. Health API (Go)
RESTful API for querying health data from Grafana alerts:

#### Endpoints:
- `GET /api/v1/health` - Get all health checks (from Grafana alerts)
- `GET /api/v1/health/{target}` - Get specific target health status
- `GET /api/v1/alerts` - Get raw Grafana alert summary
- `GET /api/v1/metrics/{metric}` - Query Prometheus metrics
- `GET /healthz` - Liveness probe

#### How It Works:
The Health API queries Grafana's alerting API to determine the health status of targets:
- **Firing alerts** → `down` status
- **Pending alerts** → `unknown` status
- **Normal alerts** → `healthy` status

This provides a clean REST API that translates Grafana alert states into simple health check responses.

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
    module: http_2xx
    targets:
      - https://example.com
      - https://google.com

# Grafana Alloy - Federation and label enrichment
alloy:
  enabled: true
  remoteWritePort: 9009  # Port for receiving metrics from Prometheus
  customLabels:
    cluster: "healthcheck-demo"
    environment: "production"
    additional:
      team: "platform"
      region: "us-west-2"

# Grafana instance with operator
grafana:
  enabled: true
  instance:
    name: is-it-up-tho
    adminUser: admin
    adminPassword: admin  # Change in production!
  datasource:
    name: prometheus
    type: prometheus
  alerts:
    enabled: true  # Automatic alert rules for probe failures

# Health API
healthApi:
  enabled: true
  replicaCount: 2
  image:
    repository: your-registry/health-api
    tag: latest
```

## Grafana Alert Integration

The system automatically generates Grafana alert rules for each probe target and uses these alerts to drive the Health API.

### How It Works

1. **Automatic Alert Generation**: For each target in `values.yaml`, a Grafana alert rule is created
2. **Alert Evaluation**: Grafana monitors `probe_success` metrics and fires alerts when targets are down >5 minutes
3. **Health API Integration**: The Health API queries Grafana's alert state to report health status

### Alert Lifecycle

```
Target Down → Prometheus detects probe_success=0 → Grafana alert fires → Health API returns "down" status
```

### Configuring Alerts

Alerts are automatically created from probe targets:

```yaml
probe:
  blackbox:
    targets:
      - https://example.com  # → Creates "https-example-com-down" alert
      - https://google.com   # → Creates "https-google-com-down" alert
```

Each alert:
- Monitors the `probe_success` metric for that specific target
- Fires when the target is down for >5 minutes
- Includes labels: `severity: critical`, `probe: blackbox`, `target: <url>`
- Is queryable via the Health API

### Accessing Grafana

```bash
make port-forward-grafana
# Open http://localhost:3000
# Login: admin / admin
```

## Documentation

- [Quick Start Guide](QUICKSTART.md) - Step-by-step installation
- [Implementation Summary](IMPLEMENTATION_SUMMARY.md) - Architecture and design decisions
- [Helm Chart README](helm-charts/charts/is-it-up-tho/README.md) - Full chart documentation
- [Grafana Integration Guide](helm-charts/charts/is-it-up-tho/GRAFANA_INTEGRATION.md) - Grafana setup and alerts
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
make port-forward-grafana   # Port forward to Grafana UI
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
kubectl logs -n monitoring -l app.kubernetes.io/name=grafana-operator
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

**Prometheus remote write 404 errors:**
- Ensure Alloy is using the correct endpoint: `/api/v1/metrics/write`
- Check Prometheus has `enableRemoteWriteReceiver: true` if using federation loop
- Verify Alloy service is accessible: `kubectl get svc -n monitoring | grep alloy`

**"Out of order sample" warnings:**
- This is expected when Prometheus scrapes directly AND receives via remote write
- Samples with older timestamps than existing data are rejected
- Not critical - the remote write connection is working
- To eliminate: disable direct scraping and only use remote write from Alloy

**Grafana alerts not working:**
```bash
kubectl get grafanaalertrulegroup -n monitoring
kubectl get grafanafolder -n monitoring
kubectl logs -n monitoring -l app.kubernetes.io/name=grafana-operator
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
│           ├── GRAFANA_INTEGRATION.md
│           └── templates/
│               ├── rbac.yaml
│               ├── prometheus-crd.yaml
│               ├── probe-crd.yaml
│               ├── alloy-config.yaml
│               ├── grafana-instance.yaml
│               ├── grafana-datasource.yaml
│               ├── grafana-alerts.yaml
│               ├── grafana-folder.yaml
│               ├── service.yaml
│               ├── health-api-deployment.yaml
│               ├── configmap.yaml
│               └── NOTES.txt
├── Makefile                    # Build and deployment automation
├── QUICKSTART.md               # Installation guide
├── IMPLEMENTATION_SUMMARY.md   # Architecture details
├── GRAFANA_SETUP_SUMMARY.md    # Grafana integration summary
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

## Remote Write Federation Details

### How It Works

1. **Prometheus → Alloy**: Prometheus sends all scraped metrics to Alloy via remote write at `/api/v1/metrics/write`
2. **Alloy Processing**: Alloy adds custom labels (cluster, environment, source) to all metrics
3. **Alloy → Prometheus**: Alloy writes the enriched metrics back to Prometheus at `/api/v1/write`

### Key Configuration

**Prometheus CRD** ([prometheus-crd.yaml:35](helm-charts/charts/is-it-up-tho/templates/prometheus-crd.yaml#L35)):
```yaml
spec:
  enableRemoteWriteReceiver: true  # Required for receiving from Alloy
  remoteWrite:
    - url: http://my-release-alloy.monitoring.svc.cluster.local:9009/api/v1/metrics/write
```

**Alloy Config** ([alloy-config.yaml](helm-charts/charts/is-it-up-tho/templates/alloy-config.yaml)):
```alloy
prometheus.receive_http "from_prometheus" {
  http {
    listen_port = 9009
  }
  forward_to = [prometheus.relabel.add_custom_labels.receiver]
}

prometheus.remote_write "back_to_prometheus" {
  endpoint {
    url = "http://prometheus.monitoring.svc.cluster.local:9090/api/v1/write"
  }
}
```

### Verifying Federation

```bash
# Check metrics with custom labels
kubectl exec -n monitoring prometheus-is-it-up-tho-0 -c prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/query?query=probe_success{source="alloy"}' | jq

# Should show metrics with cluster, environment, and source labels
```

## Support

- Documentation: See `/docs` directory
- Issues: [GitHub Issues](https://github.com/your-repo/kube-nap-charts/issues)

---

**Built with observability best practices**
