package web

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type CheckResult struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Uptime    string                 `json:"uptime,omitempty"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

type RootResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var startTime = time.Now()

func RootHandler(w http.ResponseWriter, r *http.Request) {
	response := RootResponse{
		Status:    "ok",
		Message:   "course API is running",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func ReadinessHandler(checkers ...HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]CheckResult)
		var mu sync.Mutex
		var wg sync.WaitGroup

		for _, checker := range checkers {
			wg.Add(1)
			go func(c HealthChecker) {
				defer wg.Done()

				start := time.Now()
				err := c.Check(ctx)
				latency := time.Since(start)

				result := CheckResult{
					Status:  "healthy",
					Latency: latency.String(),
				}

				if err != nil {
					result.Status = "unhealthy"
					result.Error = err.Error()
				}

				mu.Lock()
				checks[c.Name()] = result
				mu.Unlock()
			}(checker)
		}

		wg.Wait()

		status := HealthStatusHealthy
		statusCode := http.StatusOK
		unhealthyCount := 0

		for _, check := range checks {
			if check.Status == "unhealthy" {
				unhealthyCount++
			}
		}

		if unhealthyCount > 0 {
			if unhealthyCount == len(checks) {
				status = HealthStatusUnhealthy
				statusCode = http.StatusServiceUnavailable
			} else {
				status = HealthStatusDegraded
				statusCode = http.StatusOK
			}
		}

		response := HealthResponse{
			Status:    status,
			Timestamp: time.Now(),
			Uptime:    time.Since(startTime).String(),
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}
}
