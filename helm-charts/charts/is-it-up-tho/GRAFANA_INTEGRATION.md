# Grafana Operator Integration

This chart now includes Grafana Operator integration that automatically configures:

1. **Grafana Instance** - A Grafana instance deployed via the Grafana Operator
2. **Prometheus Datasource** - Automatically configured to connect to the Prometheus instance
3. **Alert Rules** - Automatically generated from probe targets configured in `values.yaml`

## Architecture

```
┌─────────────────┐
│ Probe Targets   │
│ (from CRD)      │
└────────┬────────┘
         │
         v
┌─────────────────┐      ┌──────────────────┐
│ Blackbox        │─────>│ Prometheus       │
│ Exporter        │      │                  │
└─────────────────┘      └────────┬─────────┘
                                  │
                         ┌────────v─────────┐
                         │ Grafana          │<──── Health API
                         │ - Datasource     │      queries here
                         │ - Alert Rules    │
                         └──────────────────┘
```

## Configuration

### Enabling Grafana

The Grafana operator is enabled by default. To disable it:

```yaml
grafanaOperator:
  enabled: false

grafana:
  enabled: false
```

### Grafana Settings

```yaml
grafana:
  enabled: true

  instance:
    name: is-it-up-tho
    adminUser: admin
    adminPassword: admin  # Change this in production!

  service:
    type: ClusterIP
    port: 3000

  datasource:
    name: prometheus
    type: prometheus
    access: proxy
    isDefault: true

  alerts:
    enabled: true
```

### Automatic Alert Generation

Alerts are automatically generated for each probe target. For example, if you have:

```yaml
probe:
  blackbox:
    enabled: true
    targets:
      - https://example.com
      - https://google.com
```

The chart will create Grafana alert rules that:
- Monitor `probe_success` metric for each target
- Trigger when a target is down for more than 5 minutes
- Label alerts with severity: critical and the target URL

## Health API Integration

The Health API is automatically configured to query Grafana's API when Grafana is enabled:

- **GRAFANA_URL**: Points to the Grafana service
- **GRAFANA_USER**: Admin username for API access
- **GRAFANA_PASSWORD**: Admin password for API access

When Grafana is disabled, it falls back to querying Prometheus directly.

## Deployment

1. **Update Helm dependencies:**

```bash
cd helm-charts/charts/is-it-up-tho
helm dependency update
```

2. **Install or upgrade the chart:**

```bash
helm upgrade --install is-it-up-tho . \
  --namespace monitoring \
  --create-namespace \
  --wait
```

3. **Access Grafana:**

```bash
# Port forward to access Grafana UI
kubectl port-forward -n monitoring svc/is-it-up-tho-grafana-service 3000:3000

# Open http://localhost:3000
# Login with admin/admin (or your configured credentials)
```

## Resources Created

The following Grafana Operator CRDs are created:

1. **Grafana** - The Grafana instance
   - File: `templates/grafana-instance.yaml`

2. **GrafanaDatasource** - Prometheus datasource configuration
   - File: `templates/grafana-datasource.yaml`

3. **GrafanaAlertRuleGroup** - Alert rules for probe targets
   - File: `templates/grafana-alerts.yaml`

## Customizing Alerts

Alert rules are generated with the following defaults:

- **Evaluation Interval**: Matches `prometheus.evaluationInterval`
- **Alert Duration**: 5 minutes
- **Severity**: critical
- **Condition**: `probe_success < 1`

To customize alert behavior, modify the template in `templates/grafana-alerts.yaml`.

## Troubleshooting

### Check Grafana Operator is running

```bash
kubectl get pods -n grafana-operator-system
```

### Check Grafana resources

```bash
kubectl get grafana,grafanadatasource,grafanaalertrulegroup -n monitoring
```

### View Grafana logs

```bash
kubectl logs -n monitoring deployment/is-it-up-tho-grafana -f
```

### Check datasource connection

```bash
# Access Grafana UI and go to:
# Configuration -> Data Sources -> prometheus
# Click "Test" to verify connectivity
```
