package proxy

import (
	"ant-chrome/backend/internal/config"
	"testing"
	"time"
)

func TestNormalizeCheckSettingsDefaultsAndSelectsTargets(t *testing.T) {
	settings := NormalizeCheckSettings(config.ProxyCheckConfig{
		Targets: []config.ProxyCheckTarget{
			{ID: " speed-main ", URL: " https://speed.example.com ", Type: " speed ", TimeoutMs: 0},
			{ID: "health-main", URL: "https://health.example.com", Type: "ip_health", Parser: " ipqualityscore ", TimeoutMs: 1500},
		},
	})

	if settings.BridgeStartTimeoutMs != defaultBridgeStartTimeoutMs {
		t.Fatalf("bridge timeout = %d", settings.BridgeStartTimeoutMs)
	}
	if settings.SpeedTargetID != "speed-main" {
		t.Fatalf("speed target id = %q", settings.SpeedTargetID)
	}
	if settings.IPHealthTargetID != "health-main" {
		t.Fatalf("ip health target id = %q", settings.IPHealthTargetID)
	}
	if settings.Targets[0].TimeoutMs != defaultTargetTimeoutMs {
		t.Fatalf("target timeout = %d", settings.Targets[0].TimeoutMs)
	}
}

func TestNormalizeCheckTargetsDropsInvalidAndDuplicateTargets(t *testing.T) {
	targets := NormalizeCheckTargets([]config.ProxyCheckTarget{
		{ID: "main", URL: "https://example.com"},
		{ID: " MAIN ", URL: "https://duplicate.example.com"},
		{ID: "missing-url"},
	})

	if len(targets) != 1 {
		t.Fatalf("len = %d", len(targets))
	}
	if targets[0].Name != "main" || targets[0].Type != "speed" {
		t.Fatalf("target defaults were not applied: %#v", targets[0])
	}
}

func TestBuildProxyCheckConfigs(t *testing.T) {
	settings := config.ProxyCheckConfig{
		SpeedTargetID:    "speed-main",
		IPHealthTargetID: "health-main",
		Targets: []config.ProxyCheckTarget{
			{ID: "speed-main", Type: "speed", URL: "https://speed.example.com", TimeoutMs: 1200},
			{ID: "health-main", Type: "ip_health", URL: "https://health.example.com", Parser: "ipqualityscore", TimeoutMs: 2300},
		},
	}

	speed := BuildSpeedTestConfig(settings)
	if len(speed.URLs) != 1 || speed.URLs[0] != "https://speed.example.com" || speed.Timeout != 1200*time.Millisecond {
		t.Fatalf("speed config = %#v", speed)
	}

	health := BuildIPHealthConfig(settings)
	if health.URL != "https://health.example.com" || health.Source != "health-main" || health.Parser != "ipqualityscore" || health.Timeout != 2300*time.Millisecond {
		t.Fatalf("health config = %#v", health)
	}
}
