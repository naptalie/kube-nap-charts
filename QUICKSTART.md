# Quick Start Guide - is-it-up-tho

This guide will help you get started quickly with the is-it-up-tho health monitoring solution.

## Prerequisites

1. Kubernetes cluster (1.19+)
2. Helm 3.0+
3. Prometheus Operator installed
4. Docker (for building the Health API image)

## Step 1: Install Prometheus Operator

If you don't have Prometheus Operator installed:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus-operator prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

## Step 2: Build and Push the Health API Image

```bash
cd health-api

# Build the image
docker build -t <your-registry>/health-api:latest .

# Push to your registry
docker push <your-registry>/health-api:latest
```

## Step 3: Update Helm Dependencies

```bash
cd helm-charts/charts/is-it-up-tho

# Update dependencies to download Grafana Alloy and Blackbox Exporter
helm dependency update
```

## Step 4: Configure Your Values

Edit `values.yaml` to customize:

```yaml
# Set your health check targets
probe:
  blackbox:
    enabled: true
    targets:
      - https://your-app.com
      - https://your-api.com/health

# Set your Health API image
healthApi:
  enabled: true
  image:
    repository: <your-registry>/health-api
    tag: latest

# Configure custom labels for Alloy
alloy:
  enabled: true
  customLabels:
    cluster: "my-cluster"
    environment: "production"
```

## Step 5: Install the Chart

```bash
helm install my-healthchecks . \
  --namespace monitoring \
  --create-namespace
```

## Step 6: Verify Installation

Check that all components are running:

```bash
# Check pods
kubectl get pods -n monitoring

# Check Prometheus instance
kubectl get prometheus -n monitoring

# Check probes
kubectl get probe -n monitoring

# Check services
kubectl get svc -n monitoring
```

## Step 7: Access the Health API

Port-forward to access the API:

```bash
kubectl port-forward -n monitoring svc/my-healthchecks-is-it-up-tho-health-api 8080:8080
```

Test the API:

```bash
# Get all health checks
curl http://localhost:8080/api/v1/health | jq

# Get specific target
curl http://localhost:8080/api/v1/health/https://your-app.com | jq
```

## Step 8: Access Prometheus

Port-forward to Prometheus:

```bash
kubectl port-forward -n monitoring svc/my-healthchecks-is-it-up-tho-prometheus 9090:9090
```

Visit http://localhost:9090 and run queries:

```promql
# Check probe success
probe_success

# Check with custom labels from Alloy
probe_success{cluster="my-cluster"}
```

## Example API Response

```json
{
  "total": 2,
  "healthy": 2,
  "down": 0,
  "unknown": 0,
  "checks": [
    {
      "target": "https://your-app.com",
      "status": "healthy",
      "last_checked": "2025-11-16T10:00:00Z",
      "probe": "blackbox"
    },
    {
      "target": "https://your-api.com/health",
      "status": "healthy",
      "last_checked": "2025-11-16T10:00:00Z",
      "probe": "blackbox"
    }
  ]
}
```

## Troubleshooting

### No metrics showing up

1. Check if Prometheus is scraping targets:
```bash
kubectl port-forward -n monitoring svc/my-healthchecks-is-it-up-tho-prometheus 9090:9090
```
Visit http://localhost:9090/targets

2. Check probe CRD status:
```bash
kubectl describe probe -n monitoring
```

### Health API returns empty data

1. Verify Prometheus is accessible:
```bash
kubectl logs -n monitoring -l component=health-api
```

2. Check Prometheus service:
```bash
kubectl get svc -n monitoring | grep prometheus
```

### Alloy not adding labels

1. Check Alloy logs:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=alloy
```

2. Verify ConfigMap:
```bash
kubectl get configmap -n monitoring | grep alloy
kubectl describe configmap <configmap-name> -n monitoring
```

## Next Steps

- Add more health check targets in `values.yaml`
- Configure alerts based on `probe_success` metrics
- Set up Grafana dashboards for visualization
- Add ingress for external access to the Health API

For more details, see the full [README](helm-charts/charts/is-it-up-tho/README.md).
