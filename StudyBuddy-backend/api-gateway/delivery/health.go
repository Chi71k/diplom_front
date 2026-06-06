package delivery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"studybuddy/backend/pkg/httputil"
)

// ServiceHealthNames lists downstream services in aggregation response order.
var ServiceHealthNames = []string{
	"auth", "users", "courses", "availability", "matching", "groups", "reviews", "points",
}

// ServiceHealthEntry is one downstream health probe result.
type ServiceHealthEntry struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latencyMs"`
}

// AggregatedHealthResponse is the gateway GET /health body.
type AggregatedHealthResponse struct {
	Status   string                        `json:"status"`
	Services map[string]ServiceHealthEntry `json:"services"`
}

// HealthChecker probes downstream /health endpoints in parallel.
type HealthChecker struct {
	URLs    map[string]string
	Timeout time.Duration
	Client  *http.Client
}

// NewHealthChecker builds a checker from service name to base URL (no trailing slash).
func NewHealthChecker(urls map[string]string, timeout time.Duration) *HealthChecker {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &HealthChecker{
		URLs:    urls,
		Timeout: timeout,
		Client:  &http.Client{Timeout: timeout},
	}
}

type probeResult struct {
	name    string
	entry   ServiceHealthEntry
	healthy bool
}

// Check runs parallel health probes and returns the aggregated response and HTTP status.
func (h *HealthChecker) Check(ctx context.Context) (AggregatedHealthResponse, int) {
	names := ServiceHealthNames
	results := make([]probeResult, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Add(1)
		go func(idx int, svc string) {
			defer wg.Done()
			base := h.URLs[svc]
			results[idx] = probeOne(ctx, h.Client, svc, base)
		}(i, name)
	}
	wg.Wait()

	services := make(map[string]ServiceHealthEntry, len(names))
	okCount := 0
	for _, pr := range results {
		services[pr.name] = pr.entry
		if pr.healthy {
			okCount++
		}
	}

	resp := AggregatedHealthResponse{Services: services}
	switch {
	case okCount == len(names):
		resp.Status = "ok"
		return resp, http.StatusOK
	case okCount == 0:
		resp.Status = "down"
		return resp, http.StatusServiceUnavailable
	default:
		resp.Status = "degraded"
		return resp, http.StatusMultiStatus
	}
}

func probeOne(ctx context.Context, client *http.Client, name, base string) probeResult {
	start := time.Now()
	entry := ServiceHealthEntry{Status: "error", LatencyMs: 0}
	if base == "" {
		entry.LatencyMs = 0
		return probeResult{name: name, entry: entry, healthy: false}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/health", nil)
	if err != nil {
		entry.LatencyMs = time.Since(start).Milliseconds()
		return probeResult{name: name, entry: entry, healthy: false}
	}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	entry.LatencyMs = latency
	if err != nil {
		return probeResult{name: name, entry: entry, healthy: false}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return probeResult{name: name, entry: entry, healthy: false}
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return probeResult{name: name, entry: entry, healthy: false}
	}
	if body.Status == "ok" {
		entry.Status = "ok"
		return probeResult{name: name, entry: entry, healthy: true}
	}
	return probeResult{name: name, entry: entry, healthy: false}
}

// HandleHealth serves GET /health with aggregated downstream status.
func (h *HealthChecker) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.Timeout)
		defer cancel()
	}
	agg, status := h.Check(ctx)
	httputil.JSON(w, status, agg)
}
