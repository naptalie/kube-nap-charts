# Makefile Command Reference

Complete guide to all available Makefile targets for is-it-up-tho.

## Quick Reference

```bash
make help          # Show all available commands
make kind-setup    # Create local KIND cluster (fastest way to start)
make kind-deploy   # Deploy to KIND
make status        # Check deployment status
```

## KIND Cluster Commands

### Setup and Teardown

| Command | Description |
|---------|-------------|
| `make check-kind` | Verify KIND, kubectl, helm, and docker are installed |
| `make kind-create` | Create a new KIND cluster with 3 nodes |
| `make kind-delete` | Delete the KIND cluster |
| `make kind-setup` | Complete setup: create cluster + install Prometheus Operator |
| `make kind-rebuild` | Delete and recreate cluster with full deployment |

### Deployment

| Command | Description |
|---------|-------------|
| `make kind-load-image` | Build and load Health API image into KIND |
| `make kind-deploy` | Deploy full stack to KIND cluster |
| `make kind-verify` | Verify all resources are deployed correctly |

**Example workflow:**
```bash
# First time setup
make kind-setup

# Deploy the application
make kind-deploy

# Verify everything is working
make kind-verify
```

## Docker Image Commands

| Command | Description |
|---------|-------------|
| `make build-api` | Build the Health API Docker image |
| `make push-api` | Build and push image to registry |

**Variables:**
- `REGISTRY` - Docker registry (default: `localhost:5001` for KIND)
- `API_TAG` - Image tag (default: `latest`)

**Examples:**
```bash
# Build for KIND (local)
make build-api

# Build for production
REGISTRY=gcr.io/my-project make push-api

# Build specific version
API_TAG=v1.2.3 make build-api
```

## Helm Chart Commands

| Command | Description |
|---------|-------------|
| `make helm-deps` | Download Helm chart dependencies (Alloy, Blackbox Exporter) |
| `make helm-install` | Install the Helm chart |
| `make helm-upgrade` | Upgrade existing Helm release |
| `make helm-uninstall` | Uninstall the Helm release |
| `make install-prometheus-operator` | Install Prometheus Operator only |

**Variables:**
- `RELEASE_NAME` - Helm release name (default: `my-healthchecks`)
- `NAMESPACE` - Kubernetes namespace (default: `monitoring`)

**Examples:**
```bash
# Install with custom release name
RELEASE_NAME=prod-health make helm-install

# Install to different namespace
NAMESPACE=observability make helm-install

# Upgrade with new configuration
make helm-upgrade
```

## Monitoring and Debugging

### Port Forwarding

| Command | Description |
|---------|-------------|
| `make port-forward-api` | Port forward to Health API (localhost:8080) |
| `make port-forward-prometheus` | Port forward to Prometheus (localhost:9090) |
| `make port-forward-alloy` | Port forward to Grafana Alloy (localhost:12345) |

**Usage:**
```bash
# Start port forward (runs in foreground)
make port-forward-api

# In another terminal
curl http://localhost:8080/api/v1/health

# Or use background mode
make port-forward-api &
```

### Logs

| Command | Description |
|---------|-------------|
| `make logs-api` | Stream Health API logs (follows) |
| `make logs-prometheus` | Stream Prometheus logs |
| `make logs-alloy` | Stream Grafana Alloy logs |

**Usage:**
```bash
# Follow logs in real-time
make logs-api

# Or use kubectl directly for more control
kubectl logs -n monitoring -l component=health-api --tail=50
```

### Status Checks

| Command | Description |
|---------|-------------|
| `make status` | Show status of all resources (pods, services, CRDs) |
| `make kind-verify` | Comprehensive verification of KIND deployment |
| `make test-api` | Test Health API endpoints |

**Usage:**
```bash
# Quick status check
make status

# Full verification (KIND only)
make kind-verify

# Test the API (requires port-forward to be running)
make port-forward-api &
sleep 5
make test-api
```

## Deployment Workflows

### Complete Deployments

| Command | Description |
|---------|-------------|
| `make deploy` | Build, push, and install (production) |
| `make redeploy` | Build, push, and upgrade (production) |
| `make kind-deploy` | Full deployment to KIND |

### Production Workflow

```bash
# Set your registry
export REGISTRY=gcr.io/my-project
export API_TAG=v1.0.0

# First deployment
make deploy

# Later updates
make redeploy
```

### Local Development Workflow

```bash
# One-time setup
make kind-setup

# Deploy
make kind-deploy

# Make changes to Go code
vim health-api/main.go

# Rebuild and redeploy
make kind-load-image
kubectl rollout restart deployment/my-healthchecks-is-it-up-tho-health-api -n monitoring

# Or update Helm values
vim helm-charts/charts/is-it-up-tho/values.yaml
make helm-upgrade
```

## Utilities

| Command | Description |
|---------|-------------|
| `make clean` | Remove generated Helm charts and dependencies |
| `make help` | Display all available commands with descriptions |

## Environment Variables

### All Configurable Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REGISTRY` | `localhost:5001` | Docker registry URL |
| `API_IMAGE` | `$(REGISTRY)/health-api` | Full image name |
| `API_TAG` | `latest` | Image tag |
| `RELEASE_NAME` | `my-healthchecks` | Helm release name |
| `NAMESPACE` | `monitoring` | Kubernetes namespace |
| `CHART_PATH` | `helm-charts/charts/is-it-up-tho` | Path to Helm chart |
| `KIND_CLUSTER_NAME` | `healthcheck-demo` | KIND cluster name |
| `KIND_CONFIG` | `kind-config.yaml` | KIND configuration file |

### Setting Variables

```bash
# Inline
REGISTRY=gcr.io/my-project make build-api

# Export for session
export REGISTRY=gcr.io/my-project
export NAMESPACE=observability
make deploy

# .env file (create at project root)
echo 'REGISTRY=gcr.io/my-project' > .env
echo 'API_TAG=v1.0.0' >> .env
source .env
make deploy
```

## Common Scenarios

### Scenario 1: First Time Local Setup

```bash
# Check prerequisites
make check-kind

# Create cluster and deploy
make kind-setup
make kind-deploy

# Access services
make port-forward-api &
make port-forward-prometheus &

# Test
make test-api
```

### Scenario 2: Deploy to Production Kubernetes

```bash
# Ensure Prometheus Operator is installed
make install-prometheus-operator

# Build and push image
export REGISTRY=your-registry.io/your-project
make push-api

# Deploy
make helm-install

# Check status
make status
```

### Scenario 3: Update Health Check Targets

```bash
# Edit values.yaml
vim helm-charts/charts/is-it-up-tho/values.yaml

# Add targets under probe.blackbox.targets
# - https://newservice.com

# Upgrade release
make helm-upgrade

# Verify new probes
kubectl get probe -n monitoring
make logs-prometheus
```

### Scenario 4: Debug Issues

```bash
# Check overall status
make status

# Check specific component logs
make logs-api
make logs-prometheus
make logs-alloy

# Verify configuration
kubectl get configmap -n monitoring
kubectl describe prometheus -n monitoring

# Check events
kubectl get events -n monitoring --sort-by='.lastTimestamp'
```

### Scenario 5: Test Changes Locally

```bash
# Make code changes
vim health-api/main.go

# Rebuild and load to KIND
make kind-load-image

# Restart deployment
kubectl rollout restart deployment/my-healthchecks-is-it-up-tho-health-api -n monitoring

# Watch rollout
kubectl rollout status deployment/my-healthchecks-is-it-up-tho-health-api -n monitoring

# Test
make test-api
```

### Scenario 6: Clean Slate

```bash
# KIND cluster
make kind-delete
make kind-setup
make kind-deploy

# Helm release (keeps cluster)
make helm-uninstall
make clean
make helm-deps
make helm-install
```

## Troubleshooting

### Command Fails: "command not found"

```bash
# Run prerequisite check
make check-kind

# Install missing tools (macOS)
brew install kind kubectl helm docker

# Verify
which kind kubectl helm docker
```

### Image Pull Errors in KIND

```bash
# Ensure image is loaded
make kind-load-image

# Verify image in cluster
docker exec -it healthcheck-demo-control-plane crictl images | grep health-api

# Check pod events
kubectl describe pod -n monitoring -l component=health-api
```

### Helm Dependencies Not Found

```bash
# Clean and reinstall
make clean
make helm-deps

# Verify charts downloaded
ls -la helm-charts/charts/is-it-up-tho/charts/
```

### Port Forward Connection Refused

```bash
# Check pod is running
kubectl get pods -n monitoring

# Check service exists
kubectl get svc -n monitoring

# Try with pod directly
kubectl port-forward -n monitoring pod/<pod-name> 8080:8080
```

## Tips and Tricks

### Parallel Port Forwards

```bash
# Run multiple port forwards in background
make port-forward-api &
make port-forward-prometheus &
make port-forward-alloy &

# Stop all background jobs
jobs
kill %1 %2 %3
```

### Watch Mode

```bash
# Watch pod status
watch -n 2 'make status'

# Watch logs from multiple components
kubectl logs -n monitoring -l app=is-it-up-tho --all-containers=true -f
```

### Quick Iteration

```bash
# After code changes
make build-api && make kind-load-image && \
kubectl rollout restart deployment/my-healthchecks-is-it-up-tho-health-api -n monitoring

# Watch rollout
kubectl rollout status deployment/my-healthchecks-is-it-up-tho-health-api -n monitoring --watch
```

### Custom Values

```bash
# Use custom values file
helm upgrade my-healthchecks helm-charts/charts/is-it-up-tho \
  -n monitoring \
  -f my-custom-values.yaml

# Or use --set flags
helm upgrade my-healthchecks helm-charts/charts/is-it-up-tho \
  -n monitoring \
  --set probe.blackbox.targets={https://example.com,https://google.com}
```

## Related Documentation

- [README.md](README.md) - Project overview
- [KIND_SETUP.md](KIND_SETUP.md) - Detailed KIND setup guide
- [QUICKSTART.md](QUICKSTART.md) - Step-by-step installation
- [helm-charts/charts/is-it-up-tho/README.md](helm-charts/charts/is-it-up-tho/README.md) - Helm chart documentation
- [ALLOY_SETUP.md](helm-charts/charts/is-it-up-tho/ALLOY_SETUP.md) - Grafana Alloy configuration
