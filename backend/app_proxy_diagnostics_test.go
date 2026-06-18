package backend

import (
	"testing"
	"time"
)

func TestBuildProxyBrowserProbeConfigLimitsConcurrency(t *testing.T) {
	cfg := buildProxyBrowserProbeConfig(ProxyBrowserProbeRequest{
		URLs:        []string{"https://example.com/a", "  ", "https://example.com/b"},
		Concurrency: 99,
		TimeoutMs:   1234,
	})
	if cfg.Concurrency != 16 {
		t.Fatalf("concurrency = %d, want 16", cfg.Concurrency)
	}
	if cfg.Timeout != 1234*time.Millisecond {
		t.Fatalf("timeout = %v, want 1234ms", cfg.Timeout)
	}
	if len(cfg.URLs) != 2 || cfg.URLs[0] != "https://example.com/a" || cfg.URLs[1] != "https://example.com/b" {
		t.Fatalf("urls = %#v, want trimmed request urls", cfg.URLs)
	}
}

func TestBuildProxyBrowserProbeConfigUsesDefaults(t *testing.T) {
	cfg := buildProxyBrowserProbeConfig(ProxyBrowserProbeRequest{})
	if cfg.Concurrency <= 0 {
		t.Fatalf("expected default concurrency, got %d", cfg.Concurrency)
	}
	if cfg.Timeout <= 0 {
		t.Fatalf("expected default timeout, got %v", cfg.Timeout)
	}
	if len(cfg.URLs) == 0 {
		t.Fatalf("expected default urls")
	}
}
