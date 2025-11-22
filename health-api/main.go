package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type HealthAPIServer struct {
	prometheusClient v1.API
	grafanaURL       string
	grafanaUser      string
	grafanaPassword  string
	port             string
}

type HealthCheckResult struct {
	Target      string    `json:"target"`
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Probe       string    `json:"probe"`
	Instance    string    `json:"instance,omitempty"`
}

type HealthSummary struct {
	Total   int                 `json:"total"`
	Healthy int                 `json:"healthy"`
	Down    int                 `json:"down"`
	Unknown int                 `json:"unknown"`
	Checks  []HealthCheckResult `json:"checks"`
}

type GrafanaAlert struct {
	UID         string            `json:"uid"`
	Title       string            `json:"title"`
	State       string            `json:"state"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	ActiveAt    string            `json:"activeAt,omitempty"`
	Value       string            `json:"value,omitempty"`
}

type GrafanaAlertSummary struct {
	Total   int            `json:"total"`
	Firing  int            `json:"firing"`
	Pending int            `json:"pending"`
	Normal  int            `json:"normal"`
	Alerts  []GrafanaAlert `json:"alerts"`
}

func NewHealthAPIServer(prometheusURL, grafanaURL, grafanaUser, grafanaPassword, port string) (*HealthAPIServer, error) {
	client, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}

	return &HealthAPIServer{
		prometheusClient: v1.NewAPI(client),
		grafanaURL:       grafanaURL,
		grafanaUser:      grafanaUser,
		grafanaPassword:  grafanaPassword,
		port:             port,
	}, nil
}

func (s *HealthAPIServer) getHealthChecks(w http.ResponseWriter, r *http.Request) {
	if s.grafanaURL == "" {
		http.Error(w, "Grafana not configured", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query for current alert state
	stateURL := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", s.grafanaURL)
	stateReq, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
	if err != nil {
		log.Printf("Error creating state request: %v", err)
		http.Error(w, fmt.Sprintf("Error creating state request: %v", err), http.StatusInternalServerError)
		return
	}

	if s.grafanaUser != "" && s.grafanaPassword != "" {
		stateReq.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	client := &http.Client{}
	stateResp, err := client.Do(stateReq)
	if err != nil {
		log.Printf("Error querying alert state: %v", err)
		http.Error(w, fmt.Sprintf("Error querying alert state: %v", err), http.StatusInternalServerError)
		return
	}
	defer stateResp.Body.Close()

	var stateData map[string]any
	if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
		log.Printf("Error decoding state response: %v", err)
		http.Error(w, fmt.Sprintf("Error decoding state response: %v", err), http.StatusInternalServerError)
		return
	}

	summary := HealthSummary{
		Checks: []HealthCheckResult{},
	}

	// Extract alert states and convert to health checks
	if data, ok := stateData["data"].(map[string]any); ok {
		if groups, ok := data["groups"].([]any); ok {
			for _, group := range groups {
				if g, ok := group.(map[string]any); ok {
					if rules, ok := g["rules"].([]any); ok {
						for _, rule := range rules {
							if r, ok := rule.(map[string]any); ok {
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

								status := "healthy"
								if state == "firing" {
									status = "down"
									summary.Down++
								} else if state == "pending" {
									status = "unknown"
									summary.Unknown++
								} else {
									summary.Healthy++
								}

								check := HealthCheckResult{
									Target:      target,
									Status:      status,
									LastChecked: lastChecked,
									Probe:       labels["probe"],
								}

								summary.Checks = append(summary.Checks, check)
								summary.Total++
							}
						}
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *HealthAPIServer) getHealthCheckByTarget(w http.ResponseWriter, r *http.Request) {
	if s.grafanaURL == "" {
		http.Error(w, "Grafana not configured", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	target := vars["target"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query for current alert state
	stateURL := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", s.grafanaURL)
	stateReq, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
	if err != nil {
		log.Printf("Error creating state request: %v", err)
		http.Error(w, fmt.Sprintf("Error creating state request: %v", err), http.StatusInternalServerError)
		return
	}

	if s.grafanaUser != "" && s.grafanaPassword != "" {
		stateReq.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	client := &http.Client{}
	stateResp, err := client.Do(stateReq)
	if err != nil {
		log.Printf("Error querying alert state: %v", err)
		http.Error(w, fmt.Sprintf("Error querying alert state: %v", err), http.StatusInternalServerError)
		return
	}
	defer stateResp.Body.Close()

	var stateData map[string]any
	if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
		log.Printf("Error decoding state response: %v", err)
		http.Error(w, fmt.Sprintf("Error decoding state response: %v", err), http.StatusInternalServerError)
		return
	}

	// Find the specific target in the alert rules
	if data, ok := stateData["data"].(map[string]any); ok {
		if groups, ok := data["groups"].([]any); ok {
			for _, group := range groups {
				if g, ok := group.(map[string]any); ok {
					if rules, ok := g["rules"].([]any); ok {
						for _, rule := range rules {
							if r, ok := rule.(map[string]any); ok {
								labels := getStringMap(r, "labels")
								if labels["target"] == target {
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

									status := "healthy"
									if state == "firing" {
										status = "down"
									} else if state == "pending" {
										status = "unknown"
									}

									check := HealthCheckResult{
										Target:      target,
										Status:      status,
										LastChecked: lastChecked,
										Probe:       labels["probe"],
									}

									w.Header().Set("Content-Type", "application/json")
									json.NewEncoder(w).Encode(check)
									return
								}
							}
						}
					}
				}
			}
		}
	}

	http.Error(w, "Target not found", http.StatusNotFound)
}

func (s *HealthAPIServer) getMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	metricName := vars["metric"]

	query := metricName
	if query == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	result, warnings, err := s.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		log.Printf("Error querying Prometheus: %v", err)
		http.Error(w, fmt.Sprintf("Error querying Prometheus: %v", err), http.StatusInternalServerError)
		return
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metric": metricName,
		"result": result,
	})
}

func (s *HealthAPIServer) getGrafanaAlerts(w http.ResponseWriter, r *http.Request) {
	if s.grafanaURL == "" {
		http.Error(w, "Grafana not configured", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build request to Grafana Alerting API
	url := fmt.Sprintf("%s/api/v1/provisioning/alert-rules", s.grafanaURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Add basic auth if credentials are provided
	if s.grafanaUser != "" && s.grafanaPassword != "" {
		req.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error querying Grafana: %v", err)
		http.Error(w, fmt.Sprintf("Error querying Grafana: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Grafana returned non-200 status: %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Grafana error: status %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// Parse alert rules response
	var alertRules []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&alertRules); err != nil {
		log.Printf("Error decoding Grafana response: %v", err)
		http.Error(w, fmt.Sprintf("Error decoding response: %v", err), http.StatusInternalServerError)
		return
	}

	// Query for current alert state
	stateURL := fmt.Sprintf("%s/api/prometheus/grafana/api/v1/rules", s.grafanaURL)
	stateReq, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
	if err != nil {
		log.Printf("Error creating state request: %v", err)
		http.Error(w, fmt.Sprintf("Error creating state request: %v", err), http.StatusInternalServerError)
		return
	}

	if s.grafanaUser != "" && s.grafanaPassword != "" {
		stateReq.SetBasicAuth(s.grafanaUser, s.grafanaPassword)
	}

	stateResp, err := client.Do(stateReq)
	if err != nil {
		log.Printf("Error querying alert state: %v", err)
		http.Error(w, fmt.Sprintf("Error querying alert state: %v", err), http.StatusInternalServerError)
		return
	}
	defer stateResp.Body.Close()

	var stateData map[string]interface{}
	if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
		log.Printf("Error decoding state response: %v", err)
		http.Error(w, fmt.Sprintf("Error decoding state response: %v", err), http.StatusInternalServerError)
		return
	}

	// Build summary
	summary := GrafanaAlertSummary{
		Alerts: []GrafanaAlert{},
	}

	// Extract alert states from the rules endpoint response
	if data, ok := stateData["data"].(map[string]interface{}); ok {
		if groups, ok := data["groups"].([]interface{}); ok {
			for _, group := range groups {
				if g, ok := group.(map[string]interface{}); ok {
					if rules, ok := g["rules"].([]interface{}); ok {
						for _, rule := range rules {
							if r, ok := rule.(map[string]interface{}); ok {
								alert := GrafanaAlert{
									Title:       getString(r, "name"),
									State:       getString(r, "state"),
									Labels:      getStringMap(r, "labels"),
									Annotations: getStringMap(r, "annotations"),
								}

								if alerts, ok := r["alerts"].([]interface{}); ok && len(alerts) > 0 {
									if a, ok := alerts[0].(map[string]interface{}); ok {
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
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringMap(m map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	if v, ok := m[key].(map[string]interface{}); ok {
		for k, val := range v {
			if s, ok := val.(string); ok {
				result[k] = s
			}
		}
	}
	return result
}

func (s *HealthAPIServer) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *HealthAPIServer) Start() error {
	router := mux.NewRouter()

	// Health check endpoints
	router.HandleFunc("/api/v1/health", s.getHealthChecks).Methods("GET")
	router.HandleFunc("/api/v1/health/{target}", s.getHealthCheckByTarget).Methods("GET")
	router.HandleFunc("/api/v1/metrics/{metric}", s.getMetrics).Methods("GET")
	router.HandleFunc("/api/v1/alerts", s.getGrafanaAlerts).Methods("GET")

	// Liveness probe
	router.HandleFunc("/healthz", s.healthz).Methods("GET")

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	log.Printf("Starting Health API server on port %s", s.port)
	log.Printf("Available endpoints:")
	log.Printf("  GET /api/v1/health - Get all health checks")
	log.Printf("  GET /api/v1/health/{target} - Get health check for specific target")
	log.Printf("  GET /api/v1/metrics/{metric} - Query arbitrary Prometheus metrics")
	log.Printf("  GET /api/v1/alerts - Get Grafana alert status")
	log.Printf("  GET /healthz - Liveness probe")

	return http.ListenAndServe(":"+s.port, router)
}

func main() {
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://localhost:9090"
	}

	grafanaURL := os.Getenv("GRAFANA_URL")
	if grafanaURL == "" {
		log.Printf("GRAFANA_URL not set, Grafana alerts endpoint will be unavailable")
	}

	grafanaUser := os.Getenv("GRAFANA_USER")
	if grafanaUser == "" {
		grafanaUser = "admin"
	}

	grafanaPassword := os.Getenv("GRAFANA_PASSWORD")
	if grafanaPassword == "" {
		grafanaPassword = "admin"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server, err := NewHealthAPIServer(prometheusURL, grafanaURL, grafanaUser, grafanaPassword, port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
