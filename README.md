# Kube-Nap Charts

> **Is It Up Tho?** - A comprehensive Kubernetes health monitoring solution

Lightweight, production-ready health check system combining Prometheus, Grafana Alloy, and a custom Go API.

## Features

- **Prometheus Operator Integration** - Declarative Prometheus instances via CRDs
- **Blackbox Exporter** - HTTP/HTTPS/TCP health checks
- **Grafana Alloy** - Metric enrichment with custom labels and intelligent routing
- **Grafana Mimir** - Time series database with out-of-order sample support
- **Grafana Operator** - Automated Grafana instance with datasources and alerting
- **Go REST API** - Query health status via JSON endpoints
- **Kubernetes Native** - Uses CRDs for probe configuration
- **Production Ready** - HA configuration, resource limits, health checks
- **Flexible Architecture** - Toggle between Mimir (default) or Prometheus federation

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

### Production (With Mimir - Handles Out-of-Order Samples)

```
External Targets          Kubernetes Cluster
─────────────────         ──────────────────────────────────────────────────────
                         │
https://example.com ◄────┤  ┌─────────────────┐
https://google.com  ◄────┼──│ Blackbox        │
https://api.com     ◄────┤  │ Exporter        │
                         │  └────────┬────────┘
                         │           │
                         │           ▼ scrape
                         │  ┌──────────────────────────┐
                         │  │   Prometheus             │
                         │  │   (Operator CRD)         │
                         │  └──────┬───────────────────┘
                         │         │
                         │         │ remote_write
                         │         │ /api/v1/metrics/write
                         │         ▼
                         │  ┌──────────────────────────┐
                         │  │  Grafana Alloy           │
                         │  │  Enrichment Layer        │
                         │  │  + Custom Labels:        │
                         │  │    • cluster             │
                         │  │    • environment         │
                         │  │    • source=alloy        │
                         │  └──────┬───────────────────┘
                         │         │
                         │         │ remote_write + X-Scope-OrgID: 1
                         │         │ /api/v1/push
                         │         ▼
                         │  ┌──────────────────────────┐
                         │  │   Grafana Mimir          │
                         │  │   TSDB + Query Engine    │
                         │  │   • Out-of-order samples │
                         │  │   • Long-term storage    │
                         │  │   • S3 backend (MinIO)   │
                         │  └──────┬───────────────────┘
                         │         │
                         │         │ PromQL queries + X-Scope-OrgID: 1
                         │         │ /prometheus/api/v1/query
                         │         ▼
                         │  ┌──────────────────────────┐
                         │  │   Grafana                │
                         │  │   (Operator CRD)         │
                         │  │   • Mimir Datasource     │
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
                         ──────────────────────────────────────────────────────
```

### Local Development (Without Mimir - Lightweight)

**Optimized for KIND/local clusters with minimal resource footprint**

```
External Targets          Kubernetes Cluster
─────────────────         ──────────────────────────────────────────────
                         │
https://example.com ◄────┤  ┌─────────────────┐
https://google.com  ◄────┼──│ Blackbox        │
                         │  │ Exporter        │
                         │  └────────┬────────┘
                         │           │
                         │           ▼ scrape
                         │  ┌──────────────────────────┐
                         │  │   Prometheus             │◄───────┐
                         │  │   (Standalone)           │        │
                         │  │                          │        │
                         │  │ - Local storage only     │        │
                         │  │ - 7d retention           │        │
                         │  │ - RemoteWrite: OFF       │        │
                         │  │ - RemoteWrite RX: ON     │        │
                         │  └──────────────────────────┘        │
                         │                                       │
                         │  ┌──────────────────────────┐        │
Health API ──────────────┼─►│  Grafana Alloy           │        │
(Custom Telemetry)      │  │  • Port 9009             │        │
                         │  │  • Adds labels:          │────────┘
                         │  │    - cluster             │ remote_write
                         │  │    - environment         │ /api/v1/write
                         │  │    - source              │
                         │  └──────────────────────────┘
                         │           ▲
                         │           │ query datasource
                         │  ┌────────┴─────────────────┐
                         │  │   Grafana                │
                         │  │   (Operator CRD)         │
                         │  │   • Prometheus Datasource│
                         │  │   • Alert Rules          │
                         │  └──────────────────────────┘
                         │
                         │  ┌──────────────────────────┐
Clients ─────────────────┼─►│  Health API (Go)         │
(REST API)              │  │  • /api/v1/health        │
                         │  │  • /api/v1/metrics       │
                         │  │  • Sends telemetry →     │
                         │  │    Alloy :9009           │
                         │  └──────────────────────────┘
                         │
                         ──────────────────────────────────────────────

**Resource Footprint:**
- 8 pods total (vs 16+ with Mimir)
- ~1GB RAM (vs ~3GB with Mimir)
- 0 PVCs (vs 4-6 with Mimir)
- ~400m CPU requests (vs ~1000m with Mimir)
```

## Components

### 1. Prometheus Instance
- Managed by Prometheus Operator
- 7-30 day retention (configurable per environment)
- Remote write to Alloy via `/api/v1/metrics/write`
- Remote write receiver enabled for federation loop (when Mimir disabled)
- Service discovery via Probe CRDs

### 2. Blackbox Exporter
- Probes HTTP/HTTPS/TCP endpoints
- Supports multiple probe modules (http_2xx, tcp_connect)
- Automatic target discovery

### 3. Grafana Alloy (Enrichment Layer)

**Production Mode** (when `mimir.enabled=true`):
- **Receives metrics** from Prometheus via remote write
- **Enriches metrics** with custom labels:
  - `cluster` - Cluster identifier (e.g., "default")
  - `environment` - Environment name (e.g., "production")
  - `source` - Set to "alloy" for tracking
  - Custom labels (team, region, etc.)
- **Routes to Mimir** - Handles out-of-order samples
- Adds `X-Scope-OrgID` header for Mimir multi-tenancy

**Local Development Mode** (when `mimir.enabled=false`):
- **Receives telemetry** from Health API on port 9009
- **Enriches metrics** with custom labels
- **Routes to Prometheus** via remote write receiver at `/api/v1/write`
- **Prometheus does NOT write to Alloy** - prevents circular dependency
- Lightweight: ~100m CPU, ~128Mi RAM

### 4. Grafana Mimir (Default TSDB)
- **Multi-tenant time series database** optimized for Prometheus
- **Out-of-order sample support** - 10 minute window
- **Long-term storage** - S3-compatible backend (MinIO)
- **Horizontal scaling** - Microservices architecture:
  - Distributor - Accepts remote write from Alloy
  - Ingester - Stores recent samples
  - Querier - Executes queries
  - Query Frontend - Query caching and splitting
  - Store Gateway - Queries long-term storage
  - Compactor - Compacts blocks
- **Replication factor** - Configurable (1 for local, 3 for production)
- **Authentication** - X-Scope-OrgID header (org "1")

### 5. Grafana Instance
- Managed by Grafana Operator
- Automatic datasource configuration:
  - Mimir datasource (when `mimir.enabled=true`)
  - Prometheus datasource (when `mimir.enabled=false`)
- Pre-configured alert rules for probe failures
- Alert folder organization
- Admin credentials: admin/admin (change in production!)

### 6. Health API (Go)
RESTful API for querying health data from Grafana alerts and sending custom telemetry:

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

#### Telemetry Integration:
The Health API can send custom application metrics to Alloy for enrichment:
- **Endpoint**: `http://my-healthchecks-alloy.monitoring.svc.cluster.local:9009/api/v1/metrics/write`
- **Protocol**: Prometheus Remote Write
- **Labels Added by Alloy**: cluster, environment, source
- **Example**: API latency metrics, request counts, custom business metrics

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

### Production Configuration (values.yaml)

```yaml
# Health check targets
probe:
  blackbox:
    enabled: true
    module: http_2xx
    targets:
      - https://example.com
      - https://google.com

# Grafana Alloy - Enrichment and routing
alloy:
  enabled: true
  remoteWritePort: 9009  # Port for receiving metrics from Prometheus
  customLabels:
    cluster: "default"
    environment: "production"
    additional:
      team: "platform"
      region: "us-west-2"

# Grafana Mimir - TSDB with out-of-order support
mimir:
  enabled: true  # Set to false for local dev to save resources

mimir-distributed:
  mimir:
    structuredConfig:
      # Out-of-order sample handling
      limits:
        out_of_order_time_window: 10m  # Accept samples up to 10 min old

      # Production replica configuration
      ingester:
        ring:
          replication_factor: 3  # Use 1 for local dev

      store_gateway:
        sharding_ring:
          replication_factor: 3  # Use 1 for local dev

  # MinIO for S3-compatible storage
  minio:
    enabled: true
    resources:
      limits:
        memory: 512Mi

# Grafana instance with operator
grafana:
  enabled: true
  instance:
    name: is-it-up-tho
    adminUser: admin
    adminPassword: admin  # Change in production!
  alerts:
    enabled: true  # Automatic alert rules for probe failures
  # Datasource automatically configured based on mimir.enabled

# Health API
healthApi:
  enabled: true
  replicaCount: 2
  image:
    repository: your-registry/health-api
    tag: latest

# Prometheus configuration
prometheus:
  retention: "30d"
  scrapeInterval: 30s
  evaluationInterval: 30s
```

### Local Development Configuration (values-local.yaml)

**Optimized for KIND clusters with minimal resources:**

```yaml
# Disable resource-intensive components
mimir:
  enabled: false  # Saves ~2GB RAM and 4 PVCs

# Enable Alloy for Health API telemetry
alloy:
  enabled: true
  customLabels:
    cluster: "healthcheck-demo"
    environment: "development"
  alloy:
    configMap:
      name: "<release-name>-is-it-up-tho-alloy-config"
    resources:
      limits:
        cpu: 100m
        memory: 128Mi
      requests:
        cpu: 25m
        memory: 64Mi

# Minimal Prometheus configuration
prometheus:
  retention: "7d"
  scrapeInterval: 30s
  evaluationInterval: 30s
  resources:
    limits:
      cpu: 200m
      memory: 512Mi
    requests:
      cpu: 50m
      memory: 256Mi
  # Disable Prometheus → Alloy remote write
  # (Alloy receives from Health API only)
  remoteWrite:
    enabled: false

# Minimal Health API resources
healthApi:
  replicaCount: 1
  image:
    repository: health-api
    tag: latest
    pullPolicy: Never  # Use local image for KIND
  resources:
    limits:
      cpu: 100m
      memory: 64Mi
    requests:
      cpu: 25m
      memory: 32Mi
```

### Key Configuration Differences

| Setting | Production | Local Development |
|---------|-----------|------------------|
| **Mimir** | ✅ Enabled | ❌ Disabled |
| **Alloy** | Enriches all metrics | Receives Health API telemetry only |
| **Prometheus** | Remote writes to Alloy | Standalone, no remote write |
| **Prometheus RX** | Disabled (uses Mimir) | Enabled (receives from Alloy) |
| **RAM Usage** | ~3GB | ~1GB |
| **PVCs** | 4-6 (Mimir storage) | 0 |
| **Pods** | 16+ | 8 |
| **Retention** | 30 days | 7 days |
| **API Replicas** | 2 (HA) | 1 |
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
- Check Prometheus has `enableRemoteWriteReceiver: true` if using federation loop (Mimir disabled)
- Verify Alloy service is accessible: `kubectl get svc -n monitoring | grep alloy`

**Mimir issues:**
```bash
# Check Mimir pods
kubectl get pods -n monitoring | grep mimir

# "too many unhealthy instances in the ring" errors:
# - Ensure replication_factor matches number of replicas
# - For single replica: set replication_factor: 1
# - Check values-local.yaml for proper configuration

# "no org id" / 401 Unauthorized errors:
# - Verify X-Scope-OrgID header in Alloy config (alloy-config.yaml)
# - Verify X-Scope-OrgID header in Grafana datasource (grafana-datasource.yaml)

# MinIO OOMKilled errors:
# - Increase MinIO memory limits in values-local.yaml
# - Default: 512Mi for local, 128Mi may be too low
```

**"Out of order sample" warnings (when Mimir disabled):**
- This is expected when using Prometheus federation loop
- Samples with older timestamps than existing data are rejected
- Not critical - the remote write connection is working
- **Solution**: Enable Mimir (`mimir.enabled: true`) which handles out-of-order samples natively

**Grafana alerts firing incorrectly:**
```bash
# Check alert status
kubectl get grafanaalertrulegroup -n monitoring
kubectl get grafanafolder -n monitoring

# View Grafana logs for errors
kubectl logs -n monitoring -l app=my-healthchecks-is-it-up-tho-grafana --tail=50

# Common causes:
# - execErrState: Alerting causes alerts to fire on query errors
#   Solution: Changed to execErrState: OK in grafana-alerts.yaml
# - Alert interval not divisible by scheduler interval (10s)
#   Solution: Use 30s, 60s, etc. (not 15s)
# - Datasource not synced yet
#   Solution: Wait 1-2 minutes after deployment
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

## Technical Decisions

1. **Prometheus Operator CRD**: Chose CRD over standalone Prometheus for better Kubernetes integration and declarative configuration
2. **Grafana Mimir for TSDB (Production)**: Handles out-of-order samples natively, enabling reliable metric enrichment without data loss or warnings
3. **Standalone Prometheus (Local Dev)**: Lightweight alternative to Mimir for local development, saves ~2GB RAM
4. **Grafana Operator**: Enables declarative Grafana configuration with automatic datasource and alert provisioning
5. **Grafana Alert-Driven Health API**: Health status derived from alert states provides a single source of truth for system health
6. **Go for API**: Lightweight, fast, with excellent Prometheus and HTTP client libraries
7. **Alloy for Enrichment**: Grafana Alloy provides reliable metric transformation and routing with custom label injection
8. **Alloy Telemetry Endpoint**: Health API can send custom metrics to Alloy for enrichment and routing to Prometheus
9. **No Prometheus → Alloy Remote Write (Local)**: Prevents circular dependency and "out of order" errors in local development
10. **MinIO for Object Storage**: S3-compatible storage for Mimir blocks, simple to deploy in Kubernetes
11. **Helm Dependencies**: Using official charts (Alloy, Mimir, Blackbox, Grafana Operator) ensures compatibility and maintainability
12. **Multi-tenancy Ready**: X-Scope-OrgID header support enables future multi-tenant deployments
13. **Environment-Specific Values**: Separate values files (local, dev, production) optimize resource usage per environment
14. **Alert execErrState: OK**: Prevents false alerts from query execution errors during datasource initialization

## Performance Considerations

- Health API configured with resource limits (200m CPU, 128Mi memory)
- Prometheus retention set to 30 days (configurable)
- 2 replicas of Health API for high availability
- Efficient querying via Prometheus and Grafana client libraries
- Remote write queue configuration optimized for throughput
- Alert evaluation interval of 30s balances responsiveness with resource usage

## Support

- Documentation: See `/docs` directory
- Issues: [GitHub Issues](https://github.com/your-repo/kube-nap-charts/issues)

---

**Built with observability best practices**
