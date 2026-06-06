package browser

import (
	"ant-chrome/backend/internal/config"
	"testing"
)

func TestBuildDashboardStats(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Browser.Proxies = []config.BrowserProxy{{ProxyId: "p1"}, {ProxyId: "p2"}}
	cfg.Browser.Cores = []config.BrowserCore{{CoreId: "c1"}}
	cfg.App.MaxProfileLimit = 7

	stats := BuildDashboardStats([]Profile{
		{ProfileId: "a", Running: true},
		{ProfileId: "b", Running: false},
		{ProfileId: "c", Running: true},
	}, cfg)

	if stats.TotalInstances != 3 || stats.RunningInstances != 2 {
		t.Fatalf("instance stats = %#v", stats)
	}
	if stats.ProxyCount != 2 || stats.CoreCount != 1 || stats.MaxProfileLimit != 7 {
		t.Fatalf("config stats = %#v", stats)
	}
}

func TestBuildDashboardStatsDefaultsWithoutConfig(t *testing.T) {
	stats := BuildDashboardStats(nil, nil)
	if stats.MaxProfileLimit != DefaultMaxProfileLimit {
		t.Fatalf("max profile limit = %d", stats.MaxProfileLimit)
	}
}

func TestRunningProfiles(t *testing.T) {
	profiles := RunningProfiles([]Profile{
		{ProfileId: "a", Running: true},
		{ProfileId: "b"},
	})
	if len(profiles) != 1 || profiles[0].ProfileId != "a" {
		t.Fatalf("profiles = %#v", profiles)
	}
}
