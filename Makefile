.PHONY: help install-prereqs build-api push-api helm-deps helm-install helm-upgrade helm-uninstall test-api port-forward-api port-forward-prometheus port-forward-alloy clean kind-create kind-delete kind-load-image kind-setup kind-deploy kind-rebuild kind-verify check-kind install-prometheus-operator logs-api logs-prometheus logs-alloy status deploy redeploy

# Variables
REGISTRY ?= localhost:5001
API_IMAGE := $(REGISTRY)/health-api
API_TAG ?= latest
RELEASE_NAME ?= my-healthchecks
NAMESPACE ?= monitoring
CHART_PATH := helm-charts/charts/is-it-up-tho
KIND_CLUSTER_NAME ?= healthcheck-demo
KIND_CONFIG ?= kind-config.yaml

help:
	@echo "====================================================================="
	@echo "                 is-it-up-tho Makefile Targets"
	@echo "====================================================================="
	@echo ""
	@echo "Prerequisites:"
	@echo "  install-prereqs        - Install KIND, kubectl, helm (auto-detect OS)"
	@echo "  check-kind             - Check if all tools are installed"
	@echo ""
	@echo "KIND Cluster Management:"
	@echo "  kind-create            - Create a KIND cluster with local registry"
	@echo "  kind-delete            - Delete the KIND cluster"
	@echo "  kind-load-image        - Load Health API image into KIND"
	@echo "  kind-setup             - Complete KIND setup (create + operators)"
	@echo "  kind-deploy            - Full deployment to KIND cluster"
	@echo "  kind-rebuild           - Delete and rebuild cluster with deployment"
	@echo "  kind-verify            - Verify KIND deployment"
	@echo ""
	@echo "Docker Image Management:"
	@echo "  build-api              - Build the Health API Docker image"
	@echo "  push-api               - Push the Health API image to registry"
	@echo ""
	@echo "Helm Chart Management:"
	@echo "  helm-deps              - Update Helm chart dependencies"
	@echo "  helm-install           - Install the Helm chart"
	@echo "  helm-upgrade           - Upgrade the Helm release"
	@echo "  helm-uninstall         - Uninstall the Helm release"
	@echo "  install-prometheus-operator - Install Prometheus Operator"
	@echo ""
	@echo "Testing and Monitoring:"
	@echo "  test-api               - Test the Health API endpoints"
	@echo "  port-forward-api       - Port forward to Health API (8080)"
	@echo "  port-forward-prometheus- Port forward to Prometheus (9090)"
	@echo "  port-forward-alloy     - Port forward to Alloy (12345)"
	@echo "  status                 - Check status of all resources"
	@echo "  logs-api               - View Health API logs"
	@echo "  logs-prometheus        - View Prometheus logs"
	@echo "  logs-alloy             - View Alloy logs"
	@echo ""
	@echo "Utilities:"
	@echo "  deploy                 - Build, push, and install (full deployment)"
	@echo "  redeploy               - Build, push, and upgrade (update deployment)"
	@echo "  clean                  - Clean up generated files"
	@echo ""
	@echo "Variables:"
	@echo "  REGISTRY=$(REGISTRY)"
	@echo "  API_TAG=$(API_TAG)"
	@echo "  RELEASE_NAME=$(RELEASE_NAME)"
	@echo "  NAMESPACE=$(NAMESPACE)"
	@echo "  KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME)"
	@echo ""
	@echo "Quick Start with KIND:"
	@echo "  make install-prereqs   # Install required tools"
	@echo "  make kind-setup        # Create cluster and install operators"
	@echo "  make kind-deploy       # Deploy the full stack"
	@echo "  make status            # Check everything is running"
	@echo "====================================================================="

# Build the Health API Docker image
build-api:
	@echo "Building Health API image..."
	cd health-api && docker build -t $(API_IMAGE):$(API_TAG) .
	@echo "Image built: $(API_IMAGE):$(API_TAG)"

# Push the Health API image to registry
push-api: build-api
	@echo "Pushing Health API image..."
	docker push $(API_IMAGE):$(API_TAG)
	@echo "Image pushed: $(API_IMAGE):$(API_TAG)"

# Update Helm chart dependencies (download Alloy and Blackbox Exporter)
helm-deps:
	@echo "Updating Helm dependencies..."
	cd $(CHART_PATH) && helm dependency update
	@echo "Dependencies updated"

# Install the Helm chart
helm-install: helm-deps
	@echo "Installing Helm chart..."
	helm install $(RELEASE_NAME) $(CHART_PATH) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set healthApi.image.repository=$(API_IMAGE) \
		--set healthApi.image.tag=$(API_TAG)
	@echo "Chart installed as $(RELEASE_NAME) in namespace $(NAMESPACE)"
	@echo ""
	@echo "Run 'make port-forward-api' to access the Health API"
	@echo "Run 'make port-forward-prometheus' to access Prometheus"

# Upgrade the Helm release
helm-upgrade: helm-deps
	@echo "Upgrading Helm release..."
	helm upgrade $(RELEASE_NAME) $(CHART_PATH) \
		--namespace $(NAMESPACE) \
		--set healthApi.image.repository=$(API_IMAGE) \
		--set healthApi.image.tag=$(API_TAG)
	@echo "Chart upgraded"

# Uninstall the Helm release
helm-uninstall:
	@echo "Uninstalling Helm release..."
	helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE)
	@echo "Chart uninstalled"

# Test the Health API endpoints
test-api:
	@echo "Testing Health API endpoints..."
	@echo ""
	@echo "1. Testing /healthz endpoint:"
	curl -s http://localhost:8080/healthz
	@echo ""
	@echo ""
	@echo "2. Testing /api/v1/health endpoint:"
	curl -s http://localhost:8080/api/v1/health | jq '.'
	@echo ""
	@echo "3. Testing metrics query:"
	curl -s http://localhost:8080/api/v1/metrics/probe_success | jq '.'

# Port forward to Health API
port-forward-api:
	@echo "Port forwarding to Health API on port 8080..."
	@echo "Access at: http://localhost:8080/api/v1/health"
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE_NAME)-is-it-up-tho-health-api 8080:8080

# Port forward to Prometheus
port-forward-prometheus:
	@echo "Port forwarding to Prometheus on port 9090..."
	@echo "Access at: http://localhost:9090"
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE_NAME)-is-it-up-tho-prometheus 9090:9090

# View logs from various components
logs-api:
	kubectl logs -n $(NAMESPACE) -l component=health-api --tail=100 -f

logs-prometheus:
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=prometheus --tail=100 -f

logs-alloy:
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=alloy --tail=100 -f

# Check status of resources
status:
	@echo "Checking resource status..."
	@echo ""
	@echo "Pods:"
	kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "Prometheus:"
	kubectl get prometheus -n $(NAMESPACE)
	@echo ""
	@echo "Probes:"
	kubectl get probe -n $(NAMESPACE)
	@echo ""
	@echo "Services:"
	kubectl get svc -n $(NAMESPACE)

# Clean up generated files
clean:
	@echo "Cleaning up..."
	cd $(CHART_PATH) && rm -rf charts/ Chart.lock
	@echo "Cleaned"

# Build and deploy everything
deploy: build-api push-api helm-install
	@echo "Deployment complete!"

# Update and redeploy
redeploy: build-api push-api helm-upgrade
	@echo "Redeployment complete!"

# Port forward to Alloy
port-forward-alloy:
	@echo "Port forwarding to Alloy on port 12345..."
	@echo "Access at: http://localhost:12345/metrics"
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE_NAME)-alloy 12345:12345

#################################################################
# Prerequisites Installation
#################################################################

# Install all prerequisites (KIND, kubectl, helm, docker)
install-prereqs:
	@echo "====================================================================="
	@echo "Installing Prerequisites"
	@echo "====================================================================="
	@echo ""
	@if [ -f ./install-prerequisites.sh ]; then \
		chmod +x ./install-prerequisites.sh; \
		./install-prerequisites.sh; \
	else \
		echo "ERROR: install-prerequisites.sh not found"; \
		echo "Please ensure install-prerequisites.sh is in the project root"; \
		exit 1; \
	fi

#################################################################
# KIND Cluster Management
#################################################################

# Check if KIND is installed
check-kind:
	@echo "Checking for required tools..."
	@echo ""
	@MISSING=0; \
	if ! which kind > /dev/null 2>&1; then \
		echo "✗ KIND is not installed"; \
		MISSING=1; \
	else \
		echo "✓ KIND is installed: $$(kind version)"; \
	fi; \
	if ! which kubectl > /dev/null 2>&1; then \
		echo "✗ kubectl is not installed"; \
		MISSING=1; \
	else \
		echo "✓ kubectl is installed: $$(kubectl version --client --short 2>/dev/null || kubectl version --client | head -1)"; \
	fi; \
	if ! which helm > /dev/null 2>&1; then \
		echo "✗ Helm is not installed"; \
		MISSING=1; \
	else \
		echo "✓ Helm is installed: $$(helm version --short)"; \
	fi; \
	if ! which docker > /dev/null 2>&1; then \
		echo "✗ Docker is not installed"; \
		MISSING=1; \
	else \
		if docker info > /dev/null 2>&1; then \
			echo "✓ Docker is installed and running: $$(docker version --format '{{.Client.Version}}')"; \
		else \
			echo "⚠ Docker is installed but not running"; \
			MISSING=1; \
		fi; \
	fi; \
	echo ""; \
	if [ $$MISSING -eq 1 ]; then \
		echo "=====================================================================";\
		echo "Some prerequisites are missing!"; \
		echo "=====================================================================" ;\
		echo ""; \
		echo "Run the following command to install missing tools:"; \
		echo "  make install-prereqs"; \
		echo ""; \
		echo "Or install manually:"; \
		echo "  macOS:   brew install kind kubectl helm docker"; \
		echo "  Linux:   See PREREQUISITES.md for installation instructions"; \
		echo "  Windows: choco install kind kubernetes-cli kubernetes-helm docker-desktop"; \
		echo ""; \
		exit 1; \
	else \
		echo "✓ All prerequisites are installed!"; \
	fi

# Create KIND cluster with local registry
kind-create: check-kind
	@echo "Creating KIND cluster: $(KIND_CLUSTER_NAME)"
	@if kind get clusters | grep -q "^$(KIND_CLUSTER_NAME)$$"; then \
		echo "Cluster $(KIND_CLUSTER_NAME) already exists"; \
	else \
		kind create cluster --name $(KIND_CLUSTER_NAME) --config $(KIND_CONFIG) || \
		kind create cluster --name $(KIND_CLUSTER_NAME); \
		echo "✓ KIND cluster created"; \
	fi
	@kubectl cluster-info --context kind-$(KIND_CLUSTER_NAME)
	@echo ""
	@echo "✓ Cluster is ready!"

# Delete KIND cluster
kind-delete:
	@echo "Deleting KIND cluster: $(KIND_CLUSTER_NAME)"
	kind delete cluster --name $(KIND_CLUSTER_NAME)
	@echo "✓ Cluster deleted"

# Load Health API image into KIND
kind-load-image: build-api
	@echo "Loading Health API image into KIND cluster..."
	kind load docker-image $(API_IMAGE):$(API_TAG) --name $(KIND_CLUSTER_NAME)
	@echo "✓ Image loaded into KIND"

# Install Prometheus Operator
install-prometheus-operator:
	@echo "Installing Prometheus Operator..."
	@if kubectl get namespace monitoring > /dev/null 2>&1; then \
		echo "Namespace 'monitoring' already exists"; \
	else \
		kubectl create namespace monitoring; \
	fi
	@if helm list -n monitoring | grep -q prometheus-operator; then \
		echo "Prometheus Operator already installed"; \
	else \
		helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true; \
		helm repo update; \
		helm install prometheus-operator prometheus-community/kube-prometheus-stack \
			--namespace monitoring \
			--set prometheus.enabled=false \
			--set alertmanager.enabled=false \
			--set grafana.enabled=false \
			--set prometheusOperator.enabled=true \
			--wait --timeout=5m; \
		echo "✓ Prometheus Operator installed"; \
	fi
	@echo ""
	@echo "Waiting for Prometheus Operator to be ready..."
	@kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=kube-prometheus-stack-operator -n monitoring --timeout=300s
	@echo "✓ Prometheus Operator is ready"

# Complete KIND setup (create cluster + install operators)
kind-setup: kind-create install-prometheus-operator
	@echo ""
	@echo "====================================================================="
	@echo "✓ KIND cluster setup complete!"
	@echo "====================================================================="
	@echo ""
	@echo "Cluster: $(KIND_CLUSTER_NAME)"
	@echo "Context: kind-$(KIND_CLUSTER_NAME)"
	@echo ""
	@echo "Next steps:"
	@echo "  make kind-deploy       # Deploy the full health check stack"
	@echo "  make status            # Check deployment status"
	@echo "====================================================================="

# Full deployment to KIND cluster
kind-deploy: kind-load-image helm-deps
	@echo "Deploying to KIND cluster..."
	@echo ""
	helm install $(RELEASE_NAME) $(CHART_PATH) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set healthApi.image.repository=$(API_IMAGE) \
		--set healthApi.image.tag=$(API_TAG) \
		--set healthApi.image.pullPolicy=Never \
		--set alloy.customLabels.cluster=$(KIND_CLUSTER_NAME) \
		--set alloy.customLabels.environment=development \
		--wait --timeout=10m
	@echo ""
	@echo "====================================================================="
	@echo "✓ Deployment complete!"
	@echo "====================================================================="
	@echo ""
	@echo "Resources deployed:"
	@kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "Access services:"
	@echo "  make port-forward-api         # Health API at http://localhost:8080"
	@echo "  make port-forward-prometheus  # Prometheus at http://localhost:9090"
	@echo "  make port-forward-alloy       # Alloy at http://localhost:12345"
	@echo ""
	@echo "Test the setup:"
	@echo "  make test-api"
	@echo "====================================================================="

# KIND quick teardown and rebuild
kind-rebuild: kind-delete kind-setup kind-deploy
	@echo "✓ KIND cluster rebuilt and redeployed"

# Verify KIND deployment
kind-verify:
	@echo "Verifying KIND deployment..."
	@echo ""
	@echo "1. Checking cluster:"
	@kubectl cluster-info --context kind-$(KIND_CLUSTER_NAME)
	@echo ""
	@echo "2. Checking namespaces:"
	@kubectl get namespaces
	@echo ""
	@echo "3. Checking monitoring namespace:"
	@kubectl get all -n $(NAMESPACE)
	@echo ""
	@echo "4. Checking Prometheus CRDs:"
	@kubectl get prometheus,servicemonitor,probe -n $(NAMESPACE)
	@echo ""
	@echo "5. Checking ConfigMaps:"
	@kubectl get configmap -n $(NAMESPACE)
	@echo ""
	@echo "✓ Verification complete"
