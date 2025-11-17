# is-it-up-tho Helm Chart

A comprehensive Kubernetes health check and monitoring solution combining Prometheus Operator, Blackbox Exporter, Grafana Alloy, and a custom Go API for seamless health status monitoring.

## Overview

This Helm chart provides:
- **Prometheus Instance**: Managed by the Prometheus Operator for scraping and storing metrics
- **Blackbox Exporter**: Probes HTTP/HTTPS endpoints and other network targets
- **Grafana Alloy**: Processes and enriches metrics with custom labels before writing back to Prometheus
- **Health API**: A Go-based REST API for querying health check data from Prometheus
- **Probe CRDs**: Kubernetes-native health check definitions

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Blackbox        │────>│  Prometheus      │────>│  Grafana Alloy  │
│ Exporter        │     │  (via Operator)  │     │  (Add Labels)   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │ ^                         │
                               │ └─────────────────────────┘
                               │       (Remote Write)
                               │
                        ┌──────▼──────┐
                        │  Health API │
                        │  (Go)       │
                        └─────────────┘
```

### Data Flow

1. **Blackbox Exporter** probes configured targets (HTTP/TCP endpoints)
2. **Prometheus** scrapes metrics from Blackbox Exporter via Probe CRDs
3. **Prometheus** remote writes all metrics to **Grafana Alloy**
4. **Grafana Alloy** adds custom labels (cluster, environment, etc.) to metrics
5. **Grafana Alloy** remote writes enriched metrics back to **Prometheus**
6. **Health API** queries Prometheus for health check data and exposes a REST API

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Prometheus Operator installed in the cluster
- (Optional) Container registry for the Health API image

## Installation

### 1. Install Prometheus Operator

If not already installed:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus-operator prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
```

### 2. Add Helm Repository

```bash
helm repo add is-it-up-tho <your-helm-repo-url>
helm repo update
```

### 3. Install the Chart

```bash
helm install my-health-checks is-it-up-tho/is-it-up-tho \
  --namespace monitoring \
  --create-namespace \
  -f values.yaml
```

### 4. Build and Push Health API Image

```bash
cd health-api
docker build -t your-registry/health-api:latest .
docker push your-registry/health-api:latest
```

Update `values.yaml`:
```yaml
healthApi:
  image:
    repository: your-registry/health-api
    tag: latest
```

## Configuration

### Basic Configuration

```yaml
# values.yaml
nameOverride: ""
nsPrefix: ""

# Prometheus Configuration
prometheus:
  service:
    type: ClusterIP

config:
  prometheus:
    retention: "30d"
    scrapeInterval: 30s
    evaluationInterval: 30s

# Define health check targets
probe:
  blackbox:
    enabled: true
    module: http_2xx
    targets:
      - https://example.com
      - https://google.com
      - https://api.myservice.com/health
```

### Grafana Alloy Custom Labels

Add custom labels to all metrics:

```yaml
alloy:
  enabled: true
  customLabels:
    cluster: "production-us-west-2"
    environment: "production"
    additional:
      team: "platform"
      region: "us-west-2"
      cost_center: "engineering"
```

### Health API Configuration

```yaml
healthApi:
  enabled: true
  replicaCount: 2

  image:
    repository: your-registry/health-api
    tag: latest
    pullPolicy: IfNotPresent

  service:
    type: ClusterIP
    port: 8080

  resources:
    limits:
      cpu: 200m
      memory: 128Mi
    requests:
      cpu: 100m
      memory: 64Mi
```

### Blackbox Exporter Modules

Configure different probe types:

```yaml
blackboxExporter:
  enabled: true
  config:
    modules:
      http_2xx:
        prober: http
        timeout: 5s
        http:
          method: GET
          preferred_ip_protocol: "ip4"
          follow_redirects: true
          valid_status_codes: []

      http_post_2xx:
        prober: http
        timeout: 5s
        http:
          method: POST
          body: '{"health": "check"}'

      tcp_connect:
        prober: tcp
        timeout: 5s

      icmp_ping:
        prober: icmp
        timeout: 5s
```

## Usage

### Accessing the Health API

Port-forward to access the API locally:

```bash
kubectl port-forward -n monitoring svc/my-health-checks-is-it-up-tho-health-api 8080:8080
```

### API Endpoints

#### Get All Health Checks
```bash
curl http://localhost:8080/api/v1/health
```

Response:
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
    },
    {
      "target": "https://google.com",
      "status": "healthy",
      "last_checked": "2025-11-16T10:00:00Z",
      "probe": "blackbox"
    },
    {
      "target": "https://api.myservice.com/health",
      "status": "down",
      "last_checked": "2025-11-16T10:00:00Z",
      "probe": "blackbox"
    }
  ]
}
```

#### Get Health Check for Specific Target
```bash
curl http://localhost:8080/api/v1/health/https://example.com
```

#### Query Arbitrary Prometheus Metrics
```bash
curl http://localhost:8080/api/v1/metrics/probe_http_duration_seconds
```

### Accessing Prometheus

Port-forward to access Prometheus:

```bash
kubectl port-forward -n monitoring svc/my-health-checks-is-it-up-tho-prometheus 9090:9090
```

Visit http://localhost:9090 in your browser.

### Useful PromQL Queries

```promql
# Current status of all probes
probe_success

# Failed probes
probe_success == 0

# Probe duration
probe_http_duration_seconds

# Metrics with custom labels from Alloy
probe_success{cluster="production", environment="production"}
```

## Advanced Configuration

### Using NodePort Services

Expose services externally:

```yaml
prometheus:
  service:
    type: NodePort
    nodePort: 30090

healthApi:
  service:
    type: NodePort
    nodePort: 30080
```

### Adding Ingress

```yaml
# Create ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: health-api-ingress
  namespace: monitoring
spec:
  rules:
    - host: health-api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-health-checks-is-it-up-tho-health-api
                port:
                  number: 8080
```

### Resource Management

```yaml
healthApi:
  resources:
    limits:
      cpu: 500m
      memory: 256Mi
    requests:
      cpu: 200m
      memory: 128Mi

  nodeSelector:
    node-role: monitoring

  tolerations:
    - key: "monitoring"
      operator: "Equal"
      value: "true"
      effect: "NoSchedule"

  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
                - key: component
                  operator: In
                  values:
                    - health-api
            topologyKey: kubernetes.io/hostname
```

## Monitoring and Troubleshooting

### Check Prometheus Status

```bash
kubectl get prometheus -n monitoring
kubectl describe prometheus my-health-checks-is-it-up-tho -n monitoring
```

### Check Probe Status

```bash
kubectl get probes -n monitoring
kubectl describe probe my-health-checks-is-it-up-tho-blackbox-probe -n monitoring
```

### View Logs

```bash
# Prometheus logs
kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus

# Health API logs
kubectl logs -n monitoring -l component=health-api

# Alloy logs
kubectl logs -n monitoring -l app.kubernetes.io/name=alloy
```

### Common Issues

#### Prometheus Not Scraping Targets

Check ServiceMonitor and Probe CRD labels:
```bash
kubectl get servicemonitor -n monitoring
kubectl get probe -n monitoring
```

Ensure the Prometheus instance's `serviceMonitorSelector` and `probeSelector` match:
```yaml
spec:
  serviceMonitorSelector:
    matchLabels:
      release: my-health-checks
  probeSelector:
    matchLabels:
      release: my-health-checks
```

#### Health API Can't Connect to Prometheus

Check the service name and port:
```bash
kubectl get svc -n monitoring | grep prometheus
```

Verify the environment variable in the Health API deployment:
```bash
kubectl get deployment my-health-checks-is-it-up-tho-health-api -n monitoring -o yaml | grep PROMETHEUS_URL
```

#### Alloy Not Receiving Metrics

Check remote write configuration:
```bash
kubectl get prometheus my-health-checks-is-it-up-tho -n monitoring -o yaml | grep -A 10 remoteWrite
```

Check Alloy configuration:
```bash
kubectl get configmap my-health-checks-is-it-up-tho-alloy-config -n monitoring -o yaml
```

## Upgrading

```bash
helm upgrade my-health-checks is-it-up-tho/is-it-up-tho \
  --namespace monitoring \
  -f values.yaml
```

## Uninstallation

```bash
helm uninstall my-health-checks --namespace monitoring
```

Clean up CRDs (if needed):
```bash
kubectl delete prometheus my-health-checks-is-it-up-tho -n monitoring
kubectl delete probe my-health-checks-is-it-up-tho-blackbox-probe -n monitoring
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nameOverride` | Override chart name | `""` |
| `nsPrefix` | Namespace prefix | `""` |
| `prometheus.service.type` | Prometheus service type | `ClusterIP` |
| `config.prometheus.retention` | Prometheus data retention period | `30d` |
| `config.prometheus.scrapeInterval` | Metrics scrape interval | `30s` |
| `probe.blackbox.enabled` | Enable blackbox probes | `true` |
| `probe.blackbox.targets` | List of targets to probe | `["https://example.com"]` |
| `alloy.enabled` | Enable Grafana Alloy | `true` |
| `alloy.customLabels.cluster` | Cluster label | `default` |
| `alloy.customLabels.environment` | Environment label | `production` |
| `healthApi.enabled` | Enable Health API | `true` |
| `healthApi.replicaCount` | Number of Health API replicas | `2` |
| `healthApi.image.repository` | Health API image repository | `your-registry/health-api` |
| `healthApi.service.port` | Health API service port | `8080` |

## Contributing

Contributions are welcome! Please submit issues and pull requests.

## License

[Your License Here]
