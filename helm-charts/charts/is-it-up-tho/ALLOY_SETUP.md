# Grafana Alloy Configuration Setup

This chart includes a custom Grafana Alloy configuration that enriches Prometheus metrics with custom labels.

## How It Works

The chart creates a ConfigMap (`<release-name>-is-it-up-tho-alloy-config`) containing the Alloy configuration that:

1. **Receives metrics** from Prometheus via remote write on port 9009
2. **Adds custom labels** to all metrics:
   - `cluster` - Your cluster identifier
   - `environment` - Environment name (production, staging, etc.)
   - `source` - Always set to "alloy"
   - Any additional custom labels you configure
3. **Writes enriched metrics** back to Prometheus

## Configuration Options

Edit `values.yaml` to customize the labels:

```yaml
alloy:
  enabled: true

  customLabels:
    cluster: "my-production-cluster"
    environment: "production"
    additional:
      team: "platform"
      region: "us-west-2"
      cost_center: "engineering"

  remoteWritePort: 9009
```

## Using with the Alloy Subchart

There are two ways to use the custom Alloy configuration:

### Option 1: Manual ConfigMap Mount (Recommended)

After installing the chart, manually patch the Alloy deployment to use the ConfigMap:

```bash
# 1. Install the chart
helm install my-healthchecks . -n monitoring

# 2. Get the Alloy deployment name
kubectl get deploy -n monitoring | grep alloy

# 3. Patch the deployment to mount the custom ConfigMap
kubectl patch deployment my-healthchecks-alloy -n monitoring --type='json' -p='[
  {
    "op": "add",
    "path": "/spec/template/spec/volumes/-",
    "value": {
      "name": "custom-config",
      "configMap": {
        "name": "my-healthchecks-is-it-up-tho-alloy-config"
      }
    }
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/volumeMounts/-",
    "value": {
      "name": "custom-config",
      "mountPath": "/etc/alloy-custom",
      "readOnly": true
    }
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/args/-",
    "value": "--config.file=/etc/alloy-custom/config.alloy"
  }
]'
```

### Option 2: Use values.yaml Override

Add this to your `values.yaml` before installation:

```yaml
alloy:
  enabled: true

  alloy:
    extraVolumes:
      - name: custom-alloy-config
        configMap:
          name: RELEASE_NAME-is-it-up-tho-alloy-config  # Replace RELEASE_NAME

    extraVolumeMounts:
      - name: custom-alloy-config
        mountPath: /etc/alloy-custom
        readOnly: true

    extraArgs:
      - --config.file=/etc/alloy-custom/config.alloy
      - --server.http.listen-addr=0.0.0.0:12345
      - --storage.path=/tmp/alloy

  customLabels:
    cluster: "default"
    environment: "production"

  remoteWritePort: 9009
```

**Note:** Replace `RELEASE_NAME` with your actual Helm release name.

### Option 3: Post-Install Hook Script

Create a post-install script:

```bash
#!/bin/bash
# post-install.sh

RELEASE_NAME="${1:-my-healthchecks}"
NAMESPACE="${2:-monitoring}"

echo "Configuring Alloy to use custom ConfigMap..."

# Wait for Alloy deployment to be ready
kubectl wait --for=condition=available --timeout=300s \
  deployment/${RELEASE_NAME}-alloy -n ${NAMESPACE}

# Get the ConfigMap name
CONFIGMAP_NAME="${RELEASE_NAME}-is-it-up-tho-alloy-config"

# Patch the deployment
kubectl patch deployment ${RELEASE_NAME}-alloy -n ${NAMESPACE} --type='json' -p="[
  {
    \"op\": \"add\",
    \"path\": \"/spec/template/spec/volumes/-\",
    \"value\": {
      \"name\": \"custom-config\",
      \"configMap\": {
        \"name\": \"${CONFIGMAP_NAME}\"
      }
    }
  },
  {
    \"op\": \"add\",
    \"path\": \"/spec/template/spec/containers/0/volumeMounts/-\",
    \"value\": {
      \"name\": \"custom-config\",
      \"mountPath\": \"/etc/alloy-custom\",
      \"readOnly\": true
    }
  },
  {
    \"op\": \"replace\",
    \"path\": \"/spec/template/spec/containers/0/args\",
    \"value\": [
      \"run\",
      \"--server.http.listen-addr=0.0.0.0:12345\",
      \"--storage.path=/tmp/alloy\",
      \"--config.file=/etc/alloy-custom/config.alloy\"
    ]
  }
]"

echo "Alloy configured successfully!"
echo "Waiting for pods to restart..."
kubectl rollout status deployment/${RELEASE_NAME}-alloy -n ${NAMESPACE}

echo "Configuration complete!"
```

Run after installation:

```bash
chmod +x post-install.sh
./post-install.sh my-healthchecks monitoring
```

## Verification

Check that Alloy is using the custom configuration:

```bash
# Check ConfigMap exists
kubectl get configmap my-healthchecks-is-it-up-tho-alloy-config -n monitoring

# View the configuration
kubectl get configmap my-healthchecks-is-it-up-tho-alloy-config -n monitoring -o yaml

# Check Alloy logs
kubectl logs -n monitoring -l app.kubernetes.io/name=alloy --tail=50

# Verify Alloy is receiving metrics
kubectl port-forward -n monitoring svc/my-healthchecks-alloy 12345:12345

# Visit http://localhost:12345/metrics to see Alloy's internal metrics
```

## Testing the Setup

1. **Check Prometheus is sending to Alloy:**

```bash
kubectl port-forward -n monitoring svc/my-healthchecks-is-it-up-tho-prometheus 9090:9090
```

Visit http://localhost:9090 and run:
```promql
prometheus_remote_storage_samples_total
```

2. **Check for enriched metrics in Prometheus:**

```promql
probe_success{source="alloy"}
probe_success{cluster="my-production-cluster"}
```

3. **View all labels:**

```promql
{__name__=~"probe_.*", source="alloy"}
```

## Troubleshooting

### Alloy not starting

Check the logs:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=alloy
```

Common issues:
- Config file path incorrect
- ConfigMap not mounted
- Syntax errors in config.alloy

### Metrics not enriched

1. Check Prometheus remote write config:
```bash
kubectl describe prometheus -n monitoring
```

2. Check Alloy is receiving data:
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=alloy | grep "receive_http"
```

3. Verify the remote write URL is correct:
```bash
# Should point to: http://<release>-alloy.<namespace>.svc.cluster.local:9009/api/v1/push
kubectl get prometheus -n monitoring -o yaml | grep remoteWrite -A 10
```

### ConfigMap changes not applied

Restart Alloy after modifying the ConfigMap:
```bash
kubectl rollout restart deployment my-healthchecks-alloy -n monitoring
```

## Configuration Reference

The ConfigMap structure:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: <release-name>-is-it-up-tho-alloy-config
data:
  config.alloy: |
    // Alloy configuration in Alloy syntax
    // See: https://grafana.com/docs/alloy/latest/
```

## Additional Resources

- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/latest/)
- [Alloy Configuration Syntax](https://grafana.com/docs/alloy/latest/reference/config-blocks/)
- [Prometheus Remote Write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
