package services

import (
	"context"
	"mashaghel/internal/repositories"
)

type Health string

const (
	Healthy    Health = "healthy"
	NotHealthy Health = "not_healthy"
)

type Ready string

const (
	IsReady  Ready = "ready"
	NotReady Ready = "not_ready"
)

type HealthStatus struct {
	Status  Health         `json:"status"`
	Error   string         `json:"error"`
	Details map[string]any `json:"details"`
}

type HealthCheckResult struct {
	Scylla HealthStatus `json:"scylla"`
}

type ReadyStatus struct {
	Status  Ready          `json:"status"`
	Error   string         `json:"error"`
	Details map[string]any `json:"details"`
}

type ReadyCheckResult struct {
	Scylla ReadyStatus `json:"scylla"`
}

type SystemService interface {
	HealthCheck(ctx context.Context) (HealthCheckResult, []error)
	ReadyCheck(ctx context.Context) (ReadyCheckResult, []error)
}

type systemService struct {
	systemRepository repositories.SystemRepository
}

func NewSystemService(systemRepository repositories.SystemRepository) SystemService {
	return &systemService{
		systemRepository: systemRepository,
	}
}

func (s *systemService) HealthCheck(ctx context.Context) (HealthCheckResult, []error) {
	statuses := HealthCheckResult{}
	var errors []error

	// Check ScyllaDB
	if err := s.systemRepository.ScyllaDBPing(ctx); err != nil {
		statuses.Scylla = HealthStatus{Status: NotHealthy, Error: err.Error(), Details: nil}
		errors = append(errors, err)
	} else {
		statuses.Scylla = HealthStatus{Status: Healthy, Error: "", Details: nil}
	}

	if len(errors) > 0 {
		return statuses, errors
	}

	return statuses, nil
}

func (s *systemService) ReadyCheck(ctx context.Context) (ReadyCheckResult, []error) {
	statuses := ReadyCheckResult{}
	var errors []error

	// Check ScyllaDB
	if err := s.systemRepository.ScyllaDBPing(ctx); err != nil {
		statuses.Scylla = ReadyStatus{Status: NotReady, Error: err.Error(), Details: nil}
		errors = append(errors, err)
	} else {
		statuses.Scylla = ReadyStatus{Status: IsReady, Error: "", Details: nil}
	}

	if len(errors) > 0 {
		return statuses, errors
	}

	return statuses, nil
}
