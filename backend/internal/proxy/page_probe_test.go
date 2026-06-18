package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunBrowserPageProbeCollectsConcurrentMetrics(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	result := runBrowserPageProbe("node-a", server.Client(), BrowserPageProbeConfig{
		URLs:        []string{server.URL},
		Timeout:     time.Second,
		Concurrency: 4,
	})
	if !result.Ok {
		t.Fatalf("probe Ok = false, error=%q", result.Error)
	}
	if result.Completed != 4 || result.Failed != 0 || result.Concurrency != 4 {
		t.Fatalf("unexpected counts: completed=%d failed=%d concurrency=%d", result.Completed, result.Failed, result.Concurrency)
	}
	if result.Bytes != 8 {
		t.Fatalf("bytes = %d, want 8", result.Bytes)
	}
	if result.P95Ms < 0 || result.AverageMs < 0 || result.TotalMs < 0 {
		t.Fatalf("unexpected latency metrics: %#v", result)
	}
}

func TestRunBrowserPageProbeReportsFailures(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadGateway)
	}))
	defer server.Close()

	result := runBrowserPageProbe("node-a", server.Client(), BrowserPageProbeConfig{
		URLs:        []string{server.URL},
		Timeout:     time.Second,
		Concurrency: 3,
	})
	if result.Ok {
		t.Fatalf("probe Ok = true, want false")
	}
	if result.Completed != 0 || result.Failed != 3 {
		t.Fatalf("unexpected counts: completed=%d failed=%d", result.Completed, result.Failed)
	}
	if result.Error != fmt.Sprintf("%d %s", http.StatusBadGateway, http.StatusText(http.StatusBadGateway)) {
		t.Fatalf("error = %q", result.Error)
	}
}

func TestPercentileLatency(t *testing.T) {
	t.Parallel()

	latencies := []int64{10, 20, 30, 40, 50}
	if got := percentileLatency(latencies, 0.95); got != 50 {
		t.Fatalf("p95 = %d, want 50", got)
	}
	if got := percentileLatency(latencies, 0.50); got != 30 {
		t.Fatalf("p50 = %d, want 30", got)
	}
}
