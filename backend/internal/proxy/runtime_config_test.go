package proxy

import (
	"encoding/json"
	"os"
	"testing"

	"ant-chrome/backend/internal/config"
)

func TestXrayRuntimeConfigUsesWarningLogLevel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Browser.UserDataRoot = t.TempDir()
	manager := &XrayManager{Config: cfg, AppRoot: t.TempDir()}

	cfgPath, err := manager.buildRuntimeConfigWithRoute(
		"log-level-test",
		[]interface{}{map[string]interface{}{"protocol": "freedom", "tag": "proxy-out"}},
		[]interface{}{},
		19092,
		"",
	)
	if err != nil {
		t.Fatalf("buildRuntimeConfigWithRoute returned error: %v", err)
	}

	runtimeConfig := readRuntimeConfigMap(t, cfgPath)
	logConfig := runtimeConfig["log"].(map[string]interface{})
	if got := logConfig["loglevel"]; got != "warning" {
		t.Fatalf("xray loglevel = %v, want warning", got)
	}
}

func TestSingBoxRuntimeConfigUsesWarnLogLevel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Browser.UserDataRoot = t.TempDir()
	manager := &SingBoxManager{Config: cfg, AppRoot: t.TempDir()}

	cfgPath, err := manager.buildConfig("log-level-test", map[string]interface{}{"type": "direct", "tag": "proxy-out"}, 19093)
	if err != nil {
		t.Fatalf("buildConfig returned error: %v", err)
	}

	runtimeConfig := readRuntimeConfigMap(t, cfgPath)
	logConfig := runtimeConfig["log"].(map[string]interface{})
	if got := logConfig["level"]; got != "warn" {
		t.Fatalf("sing-box log level = %v, want warn", got)
	}
}

func readRuntimeConfigMap(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime config failed: %v", err)
	}
	var runtimeConfig map[string]interface{}
	if err := json.Unmarshal(data, &runtimeConfig); err != nil {
		t.Fatalf("unmarshal runtime config failed: %v", err)
	}
	return runtimeConfig
}
