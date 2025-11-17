# KIND (Kubernetes IN Docker) Setup Guide

This guide will help you set up a local Kubernetes cluster using KIND to test the is-it-up-tho health monitoring solution.

## Prerequisites

### Required Tools

1. **Docker Desktop** or **Docker Engine**
   ```bash
   docker --version
   # Should show Docker version 20.10.0 or later
   ```

2. **KIND (Kubernetes IN Docker)**
   ```bash
   # macOS
   brew install kind

   # Linux
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind

   # Windows (PowerShell as Admin)
   choco install kind
   ```

3. **kubectl**
   ```bash
   # macOS
   brew install kubectl

   # Linux
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   chmod +x kubectl
   sudo mv kubectl /usr/local/bin/

   # Windows
   choco install kubernetes-cli
   ```

4. **Helm 3**
   ```bash
   # macOS
   brew install helm

   # Linux
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

   # Windows
   choco install kubernetes-helm
   ```

### Verify Installation

```bash
make check-kind
```

This will verify all required tools are installed.

## Quick Start

### Option 1: Automated Setup (Recommended)

```bash
# 1. Create KIND cluster and install Prometheus Operator
make kind-setup

# 2. Deploy the full health check stack
make kind-deploy

# 3. Verify deployment
make status
```

### Option 2: Manual Step-by-Step

```bash
# 1. Create KIND cluster
make kind-create

# 2. Install Prometheus Operator
make install-prometheus-operator

# 3. Build Health API image
make build-api

# 4. Load image into KIND
make kind-load-image

# 5. Install Helm chart
make kind-deploy

# 6. Check status
make kind-verify
```

## Accessing Services

### Health API

```bash
# Port forward
make port-forward-api

# In another terminal, test the API
make test-api

# Or manually
curl http://localhost:8080/api/v1/health | jq
```

### Prometheus

```bash
# Port forward
make port-forward-prometheus

# Open browser to http://localhost:9090
```

Query examples:
```promql
# Check probe status
probe_success

# Check with custom labels
probe_success{cluster="healthcheck-demo"}

# Failed probes
probe_success == 0
```

### Grafana Alloy

```bash
# Port forward
make port-forward-alloy

# View Alloy metrics at http://localhost:12345/metrics
```

## Configuration

### Customize KIND Cluster

Edit `kind-config.yaml` to change:
- Number of worker nodes
- Port mappings
- Network settings

```yaml
nodes:
  - role: control-plane
  - role: worker
  - role: worker  # Add/remove worker nodes
```

### Customize Health Check Targets

Edit `helm-charts/charts/is-it-up-tho/values.yaml`:

```yaml
probe:
  blackbox:
    enabled: true
    targets:
      - https://google.com
      - https://github.com
      - https://your-service.com
```

Then upgrade:
```bash
make helm-upgrade
```

### Customize Labels

```yaml
alloy:
  customLabels:
    cluster: "my-local-cluster"
    environment: "development"
    additional:
      team: "platform"
      owner: "your-name"
```

## Common Operations

### View Logs

```bash
# Health API logs
make logs-api

# Prometheus logs
make logs-prometheus

# Alloy logs
make logs-alloy

# All pods in monitoring namespace
kubectl logs -n monitoring --all-containers=true -f --max-log-requests=20
```

### Check Resources

```bash
# Quick status
make status

# Detailed verification
make kind-verify

# Get all resources
kubectl get all -n monitoring

# Check Prometheus CRDs
kubectl get prometheus,servicemonitor,probe -n monitoring
```

### Update Deployment

After making changes:

```bash
# Rebuild and reload image
make kind-load-image

# Upgrade Helm release
make helm-upgrade

# Or do both
make redeploy
```

### Reset Everything

```bash
# Delete and recreate cluster
make kind-rebuild

# Or just delete
make kind-delete
```

## Troubleshooting

### Cluster Creation Fails

```bash
# Check Docker is running
docker ps

# Check if cluster already exists
kind get clusters

# Delete existing cluster
make kind-delete

# Try creating again
make kind-create
```

### Image Not Found

```bash
# Rebuild and load image
make kind-load-image

# Verify image is in cluster
docker exec -it healthcheck-demo-control-plane crictl images | grep health-api

# Update deployment to use local image
kubectl set image deployment/my-healthchecks-is-it-up-tho-health-api \
  health-api=localhost:5001/health-api:latest -n monitoring
```

### Prometheus Operator Not Installing

```bash
# Check if CRDs exist
kubectl get crd | grep monitoring.coreos.com

# If exists, uninstall and reinstall
helm uninstall prometheus-operator -n monitoring
make install-prometheus-operator

# Check logs
kubectl logs -n monitoring -l app.kubernetes.io/name=kube-prometheus-stack-operator
```

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n monitoring

# Describe problem pod
kubectl describe pod <pod-name> -n monitoring

# Check events
kubectl get events -n monitoring --sort-by='.lastTimestamp'

# Check resources
kubectl top nodes
kubectl top pods -n monitoring
```

### Port Forward Connection Refused

```bash
# Check service exists
kubectl get svc -n monitoring

# Check pod is running
kubectl get pods -n monitoring

# Try with pod directly
kubectl port-forward -n monitoring pod/<pod-name> 8080:8080
```

### Health API Returns Empty Data

```bash
# Check Prometheus is running
kubectl get prometheus -n monitoring

# Check Prometheus service
kubectl get svc -n monitoring | grep prometheus

# Verify probe is configured
kubectl get probe -n monitoring

# Check blackbox exporter
kubectl get pods -n monitoring | grep blackbox
kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus-blackbox-exporter
```

## Advanced Configuration

### Enable NodePort Access

Modify `values.yaml`:

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

Access without port-forward:
- Prometheus: http://localhost:30090
- Health API: http://localhost:30080/api/v1/health

### Add More Probe Modules

```yaml
blackboxExporter:
  config:
    modules:
      http_2xx:
        prober: http
        timeout: 5s

      http_post_json:
        prober: http
        timeout: 5s
        http:
          method: POST
          headers:
            Content-Type: application/json
          body: '{"status":"check"}'

      tcp_connect:
        prober: tcp
        timeout: 5s

      dns_check:
        prober: dns
        timeout: 5s
        dns:
          query_name: "example.com"
```

### Resource Limits for Testing

Reduce resources for local development:

```yaml
healthApi:
  resources:
    limits:
      cpu: 100m
      memory: 64Mi
    requests:
      cpu: 50m
      memory: 32Mi
```

## Performance Tuning

### For Low-Resource Machines

1. Use single worker node:
   ```yaml
   # kind-config.yaml
   nodes:
     - role: control-plane
     - role: worker  # Only one worker
   ```

2. Reduce scrape frequency:
   ```yaml
   prometheus:
     scrapeInterval: 60s  # Instead of 30s
   ```

3. Lower retention:
   ```yaml
   prometheus:
     retention: "7d"  # Instead of 30d
   ```

### For Testing at Scale

1. Add more worker nodes in `kind-config.yaml`
2. Increase probe targets in `values.yaml`
3. Deploy multiple Prometheus instances

## Integration Testing

### Test the Full Flow

```bash
# 1. Deploy everything
make kind-deploy

# 2. Wait for all pods
kubectl wait --for=condition=ready pod --all -n monitoring --timeout=300s

# 3. Port forward API
make port-forward-api &

# 4. Test endpoints
sleep 5
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/api/v1/health | jq '.total'

# 5. Check enriched metrics in Prometheus
make port-forward-prometheus &
sleep 5
curl -s 'http://localhost:9090/api/v1/query?query=probe_success{source="alloy"}' | jq
```

### Automated Test Script

```bash
#!/bin/bash
# test-kind-deployment.sh

set -e

echo "Starting KIND deployment test..."

# Setup
make kind-setup
make kind-deploy

# Wait for readiness
echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod --all -n monitoring --timeout=600s

# Test Health API
echo "Testing Health API..."
kubectl port-forward -n monitoring svc/my-healthchecks-is-it-up-tho-health-api 8080:8080 &
PF_PID=$!
sleep 5

if curl -s http://localhost:8080/healthz | grep -q "OK"; then
    echo "✓ Health API is responding"
else
    echo "✗ Health API is not responding"
    exit 1
fi

kill $PF_PID

echo "✓ All tests passed!"
```

## Cleanup

```bash
# Uninstall Helm release
make helm-uninstall

# Delete KIND cluster
make kind-delete

# Remove Docker images
docker rmi localhost:5001/health-api:latest
```

## Next Steps

After successfully deploying to KIND:

1. **Customize targets** - Add your own services to monitor
2. **Configure alerts** - Set up alerting rules in Prometheus
3. **Add dashboards** - Create Grafana dashboards for visualization
4. **Test failover** - Kill pods and watch recovery
5. **Scale up** - Increase replicas and test load

## Resources

- [KIND Documentation](https://kind.sigs.k8s.io/)
- [Prometheus Operator](https://prometheus-operator.dev/)
- [Grafana Alloy](https://grafana.com/docs/alloy/latest/)
- [Blackbox Exporter](https://github.com/prometheus/blackbox_exporter)
