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
	"github.com/prometheus/common/model"
)

type HealthAPIServer struct {
	prometheusClient v1.API
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

func NewHealthAPIServer(prometheusURL, port string) (*HealthAPIServer, error) {
	client, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}

	return &HealthAPIServer{
		prometheusClient: v1.NewAPI(client),
		port:             port,
	}, nil
}

func (s *HealthAPIServer) getHealthChecks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query for blackbox exporter probe success metrics
	query := `probe_success{job="blackbox"}`
	result, warnings, err := s.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		log.Printf("Error querying Prometheus: %v", err)
		http.Error(w, fmt.Sprintf("Error querying Prometheus: %v", err), http.StatusInternalServerError)
		return
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	summary := HealthSummary{
		Checks: []HealthCheckResult{},
	}

	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		for _, sample := range vector {
			status := "down"
			if sample.Value == 1 {
				status = "healthy"
				summary.Healthy++
			} else {
				summary.Down++
			}

			target := string(sample.Metric["instance"])
			probe := string(sample.Metric["probe"])

			check := HealthCheckResult{
				Target:      target,
				Status:      status,
				LastChecked: sample.Timestamp.Time(),
				Probe:       probe,
			}

			summary.Checks = append(summary.Checks, check)
			summary.Total++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *HealthAPIServer) getHealthCheckByTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target := vars["target"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf(`probe_success{job="blackbox",instance="%s"}`, target)
	result, warnings, err := s.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		log.Printf("Error querying Prometheus: %v", err)
		http.Error(w, fmt.Sprintf("Error querying Prometheus: %v", err), http.StatusInternalServerError)
		return
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		if len(vector) == 0 {
			http.Error(w, "Target not found", http.StatusNotFound)
			return
		}

		sample := vector[0]
		status := "down"
		if sample.Value == 1 {
			status = "healthy"
		}

		probe := string(sample.Metric["probe"])

		check := HealthCheckResult{
			Target:      target,
			Status:      status,
			LastChecked: sample.Timestamp.Time(),
			Probe:       probe,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(check)
		return
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
	log.Printf("  GET /healthz - Liveness probe")

	return http.ListenAndServe(":"+s.port, router)
}

func main() {
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://localhost:9090"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server, err := NewHealthAPIServer(prometheusURL, port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
