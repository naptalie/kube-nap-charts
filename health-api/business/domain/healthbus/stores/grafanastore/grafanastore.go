// Package grafanastore implements the health check store using Grafana APIs.
package grafanastore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"health-api/business/domain/healthbus"
	"health-api/foundation/logger"
)

// Store implements healthbus.Storer using Grafana.
type Store struct {
	log             *logger.Logger
	grafanaURL      string
	grafanaUser     string
	grafanaPassword string
	httpClient      *http.Client
}

// NewStore creates a new Grafana-backed health check store.
func NewStore(log *logger.Logger, grafanaURL, grafanaUser, grafanaPassword string) *Store {
	return &Store{
		log:             log,
		grafanaURL:      grafanaURL,
		grafanaUser:     grafanaUser,
		grafanaPassword: grafanaPassword,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QueryHealthChecks retrieves all health checks from Grafana alerts.
func (s *Store) QueryHealthChecks(ctx context.Context) ([]healthbus.HealthCheck, error) {
	if s.grafanaURL == "" {
		return nil, fmt.Errorf("grafana not configured")
	}

	// Query for current alert state
	stateURL := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", s.grafanaURL)
	stateReq, err := http.NewRequestWithContext(ctx, http.MethodGet, stateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating state request: %w", err)
	}

	if s.grafanaUser != "" && s.grafanaPassword != "" {
		stateReq.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	stateResp, err := s.httpClient.Do(stateReq)
	if err != nil {
		return nil, fmt.Errorf("querying alert state: %w", err)
	}
	defer stateResp.Body.Close()

	if stateResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grafana returned status %d", stateResp.StatusCode)
	}

	var stateData map[string]any
	if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
		return nil, fmt.Errorf("decoding state response: %w", err)
	}

	// Extract health checks from alert rules
	var checks []healthbus.HealthCheck

	data, ok := stateData["data"].(map[string]any)
	if !ok {
		return checks, nil
	}

	groups, ok := data["groups"].([]any)
	if !ok {
		return checks, nil
	}

	for _, group := range groups {
		g, ok := group.(map[string]any)
		if !ok {
			continue
		}

		rules, ok := g["rules"].([]any)
		if !ok {
			continue
		}

		for _, rule := range rules {
			r, ok := rule.(map[string]any)
			if !ok {
				continue
			}

			state := getString(r, "state")
			labels := getStringMap(r, "labels")
			target := labels["target"]

			if target == "" {
				continue
			}

			var lastChecked time.Time
			if alerts, ok := r["alerts"].([]any); ok && len(alerts) > 0 {
				if a, ok := alerts[0].(map[string]any); ok {
					activeAt := getString(a, "activeAt")
					if t, err := time.Parse(time.RFC3339, activeAt); err == nil {
						lastChecked = t
					}
				}
			}

			if lastChecked.IsZero() {
				lastChecked = time.Now()
			}

			status := healthbus.StatusHealthy
			if state == "firing" {
				status = healthbus.StatusDown
			} else if state == "pending" {
				status = healthbus.StatusUnknown
			}

			check := healthbus.HealthCheck{
				Target:      target,
				Status:      status,
				LastChecked: lastChecked,
				Probe:       labels["probe"],
			}

			checks = append(checks, check)
		}
	}

	return checks, nil
}

// QueryHealthCheckByTarget retrieves a specific health check by target.
func (s *Store) QueryHealthCheckByTarget(ctx context.Context, target string) (healthbus.HealthCheck, error) {
	if s.grafanaURL == "" {
		return healthbus.HealthCheck{}, fmt.Errorf("grafana not configured")
	}

	// Query for current alert state
	stateURL := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", s.grafanaURL)
	stateReq, err := http.NewRequestWithContext(ctx, http.MethodGet, stateURL, nil)
	if err != nil {
		return healthbus.HealthCheck{}, fmt.Errorf("creating state request: %w", err)
	}

	if s.grafanaUser != "" && s.grafanaPassword != "" {
		stateReq.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	stateResp, err := s.httpClient.Do(stateReq)
	if err != nil {
		return healthbus.HealthCheck{}, fmt.Errorf("querying alert state: %w", err)
	}
	defer stateResp.Body.Close()

	if stateResp.StatusCode != http.StatusOK {
		return healthbus.HealthCheck{}, fmt.Errorf("grafana returned status %d", stateResp.StatusCode)
	}

	var stateData map[string]any
	if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
		return healthbus.HealthCheck{}, fmt.Errorf("decoding state response: %w", err)
	}

	// Find the specific target in the alert rules
	data, ok := stateData["data"].(map[string]any)
	if !ok {
		return healthbus.HealthCheck{}, fmt.Errorf("target not found: %s", target)
	}

	groups, ok := data["groups"].([]any)
	if !ok {
		return healthbus.HealthCheck{}, fmt.Errorf("target not found: %s", target)
	}

	for _, group := range groups {
		g, ok := group.(map[string]any)
		if !ok {
			continue
		}

		rules, ok := g["rules"].([]any)
		if !ok {
			continue
		}

		for _, rule := range rules {
			r, ok := rule.(map[string]any)
			if !ok {
				continue
			}

			labels := getStringMap(r, "labels")
			if labels["target"] != target {
				continue
			}

			state := getString(r, "state")

			var lastChecked time.Time
			if alerts, ok := r["alerts"].([]any); ok && len(alerts) > 0 {
				if a, ok := alerts[0].(map[string]any); ok {
					activeAt := getString(a, "activeAt")
					if t, err := time.Parse(time.RFC3339, activeAt); err == nil {
						lastChecked = t
					}
				}
			}

			if lastChecked.IsZero() {
				lastChecked = time.Now()
			}

			status := healthbus.StatusHealthy
			if state == "firing" {
				status = healthbus.StatusDown
			} else if state == "pending" {
				status = healthbus.StatusUnknown
			}

			return healthbus.HealthCheck{
				Target:      target,
				Status:      status,
				LastChecked: lastChecked,
				Probe:       labels["probe"],
			}, nil
		}
	}

	return healthbus.HealthCheck{}, fmt.Errorf("target not found: %s", target)
}

// QueryAlerts retrieves alert summary from Grafana.
func (s *Store) QueryAlerts(ctx context.Context) (healthbus.AlertSummary, error) {
	if s.grafanaURL == "" {
		return healthbus.AlertSummary{}, fmt.Errorf("grafana not configured")
	}

	// Query for current alert state
	stateURL := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", s.grafanaURL)
	stateReq, err := http.NewRequestWithContext(ctx, http.MethodGet, stateURL, nil)
	if err != nil {
		return healthbus.AlertSummary{}, fmt.Errorf("creating state request: %w", err)
	}

	if s.grafanaUser != "" && s.grafanaPassword != "" {
		stateReq.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	stateResp, err := s.httpClient.Do(stateReq)
	if err != nil {
		return healthbus.AlertSummary{}, fmt.Errorf("querying alert state: %w", err)
	}
	defer stateResp.Body.Close()

	if stateResp.StatusCode != http.StatusOK {
		return healthbus.AlertSummary{}, fmt.Errorf("grafana returned status %d", stateResp.StatusCode)
	}

	var stateData map[string]any
	if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
		return healthbus.AlertSummary{}, fmt.Errorf("decoding state response: %w", err)
	}

	summary := healthbus.AlertSummary{
		Alerts: []healthbus.Alert{},
	}

	// Extract alert states from the rules endpoint response
	data, ok := stateData["data"].(map[string]any)
	if !ok {
		return summary, nil
	}

	groups, ok := data["groups"].([]any)
	if !ok {
		return summary, nil
	}

	for _, group := range groups {
		g, ok := group.(map[string]any)
		if !ok {
			continue
		}

		rules, ok := g["rules"].([]any)
		if !ok {
			continue
		}

		for _, rule := range rules {
			r, ok := rule.(map[string]any)
			if !ok {
				continue
			}

			alert := healthbus.Alert{
				Title:       getString(r, "name"),
				State:       getString(r, "state"),
				Labels:      getStringMap(r, "labels"),
				Annotations: getStringMap(r, "annotations"),
			}

			if alerts, ok := r["alerts"].([]any); ok && len(alerts) > 0 {
				if a, ok := alerts[0].(map[string]any); ok {
					alert.ActiveAt = getString(a, "activeAt")
					alert.Value = getString(a, "value")
				}
			}

			summary.Alerts = append(summary.Alerts, alert)
			summary.Total++

			switch alert.State {
			case "firing":
				summary.Firing++
			case "pending":
				summary.Pending++
			case "normal":
				summary.Normal++
			}
		}
	}

	return summary, nil
}

// Helper functions

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringMap(m map[string]any, key string) map[string]string {
	result := make(map[string]string)
	if v, ok := m[key].(map[string]any); ok {
		for k, val := range v {
			if s, ok := val.(string); ok {
				result[k] = s
			}
		}
	}
	return result
}
