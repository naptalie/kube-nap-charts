// Package healthbus provides business logic for health checks.
package healthbus

import (
	"context"
	"time"

	"health-api/foundation/logger"
)

// Business manages health check operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// Storer defines the interface for health check data access.
type Storer interface {
	QueryHealthChecks(ctx context.Context) ([]HealthCheck, error)
	QueryHealthCheckByTarget(ctx context.Context, target string) (HealthCheck, error)
	QueryAlerts(ctx context.Context) (AlertSummary, error)
}

// NewBusiness creates a new health check business layer.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// QueryHealthChecks retrieves all health checks.
func (b *Business) QueryHealthChecks(ctx context.Context) (HealthSummary, error) {
	checks, err := b.storer.QueryHealthChecks(ctx)
	if err != nil {
		return HealthSummary{}, err
	}

	summary := HealthSummary{
		Checks: checks,
		Total:  len(checks),
	}

	// Count statuses
	for _, check := range checks {
		switch check.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusDown:
			summary.Down++
		case StatusUnknown:
			summary.Unknown++
		}
	}

	return summary, nil
}

// QueryHealthCheckByTarget retrieves a specific health check by target.
func (b *Business) QueryHealthCheckByTarget(ctx context.Context, target string) (HealthCheck, error) {
	return b.storer.QueryHealthCheckByTarget(ctx, target)
}

// QueryAlerts retrieves alert information.
func (b *Business) QueryAlerts(ctx context.Context) (AlertSummary, error) {
	return b.storer.QueryAlerts(ctx)
}

// =============================================================================

// Status represents the health status of a target.
type Status string

const (
	StatusHealthy Status = "healthy"
	StatusDown    Status = "down"
	StatusUnknown Status = "unknown"
)

// HealthCheck represents a single health check result.
type HealthCheck struct {
	Target      string    `json:"target"`
	Status      Status    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Probe       string    `json:"probe"`
	Instance    string    `json:"instance,omitempty"`
}

// HealthSummary represents a summary of all health checks.
type HealthSummary struct {
	Total   int           `json:"total"`
	Healthy int           `json:"healthy"`
	Down    int           `json:"down"`
	Unknown int           `json:"unknown"`
	Checks  []HealthCheck `json:"checks"`
}

// Alert represents a single alert.
type Alert struct {
	UID         string            `json:"uid"`
	Title       string            `json:"title"`
	State       string            `json:"state"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	ActiveAt    string            `json:"activeAt,omitempty"`
	Value       string            `json:"value,omitempty"`
}

// AlertSummary represents a summary of all alerts.
type AlertSummary struct {
	Total   int     `json:"total"`
	Firing  int     `json:"firing"`
	Pending int     `json:"pending"`
	Normal  int     `json:"normal"`
	Alerts  []Alert `json:"alerts"`
}
