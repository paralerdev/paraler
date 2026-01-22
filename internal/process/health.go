package process

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/paralerdev/paraler/internal/config"
)

// HealthStatus represents the health state of a service
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthUnhealthy
)

func (h HealthStatus) String() string {
	switch h {
	case HealthHealthy:
		return "healthy"
	case HealthUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthChecker performs health checks on services
type HealthChecker struct {
	client *http.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}
}

// CheckHealth performs a health check on a service
func (h *HealthChecker) CheckHealth(cfg config.Service) HealthStatus {
	if cfg.Health != "" {
		return h.checkHTTP(cfg.Health)
	}
	if cfg.Port > 0 {
		return h.checkPort(cfg.Port)
	}
	return HealthUnknown
}

// checkHTTP performs an HTTP health check
func (h *HealthChecker) checkHTTP(url string) HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return HealthUnhealthy
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return HealthUnhealthy
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return HealthHealthy
	}
	return HealthUnhealthy
}

// checkPort checks if a port is listening
func (h *HealthChecker) checkPort(port int) HealthStatus {
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return HealthUnhealthy
	}
	conn.Close()
	return HealthHealthy
}

// CheckPort checks if a specific port is available
func CheckPort(port int) bool {
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
