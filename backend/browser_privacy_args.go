package backend

import (
	"facade/backend/internal/browser"
	"facade/backend/internal/config"
	"strings"
)

var webGPUDisableFeatures = []string{"WebGPU", "WebGPUService"}

func browserSecureDNSEnabled(cfg *config.Config) bool {
	if cfg == nil || cfg.Browser.SecureDNS.Enabled == nil {
		return true
	}
	return *cfg.Browser.SecureDNS.Enabled
}

func browserSecureDNSOptions(cfg *config.Config) browser.SecureDNSOptions {
	if cfg == nil {
		return browser.SecureDNSOptions{Mode: "secure", Templates: []string{config.DefaultSecureDNSTemplate}}
	}
	mode := strings.TrimSpace(cfg.Browser.SecureDNS.Mode)
	if mode == "" {
		mode = "secure"
	}
	templates := normalizeNonEmptyStrings(cfg.Browser.SecureDNS.Templates)
	if len(templates) == 0 {
		templates = []string{config.DefaultSecureDNSTemplate}
	}
	return browser.SecureDNSOptions{Mode: mode, Templates: templates}
}

func browserPrivacyHardenedLaunchArgsEnabled(cfg *config.Config) bool {
	if cfg == nil || cfg.Browser.Privacy.HardenedLaunchArgsEnabled == nil {
		return true
	}
	return *cfg.Browser.Privacy.HardenedLaunchArgsEnabled
}

func browserPrivacySpoofSpeechVoicesEnabled(cfg *config.Config) bool {
	if cfg == nil || cfg.Browser.Privacy.SpoofSpeechVoices == nil {
		return true
	}
	return *cfg.Browser.Privacy.SpoofSpeechVoices
}

func browserPrivacyDisableWebGPUEnabled(cfg *config.Config) bool {
	if cfg == nil || cfg.Browser.Privacy.DisableWebGPU == nil {
		return true
	}
	return *cfg.Browser.Privacy.DisableWebGPU
}

func browserExitConsistencyCheckMode(cfg *config.Config) string {
	if cfg == nil {
		return "warn"
	}
	mode := strings.ToLower(strings.TrimSpace(cfg.Browser.Privacy.ExitConsistencyCheck))
	if mode == "" {
		return "warn"
	}
	return mode
}

func normalizeFingerprintArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, 0, len(args))
	for _, item := range normalizeNonEmptyStrings(args) {
		key, _, found := strings.Cut(item, "=")
		if !found {
			out = append(out, item)
			continue
		}
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if !isAllowedFingerprintArgKey(normalizedKey) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func isAllowedFingerprintArgKey(key string) bool {
	if !strings.HasPrefix(key, "--fingerprint") {
		return true
	}
	switch key {
	case "--fingerprint",
		"--fingerprint-brand",
		"--fingerprint-platform",
		"--fingerprint-color-depth",
		"--fingerprint-hardware-concurrency",
		"--fingerprint-device-memory",
		"--fingerprint-canvas-noise",
		"--fingerprint-audio-noise",
		"--fingerprint-fonts",
		"--fingerprint-do-not-track",
		"--fingerprint-media-devices",
		"--fingerprint-touch-points":
		return true
	default:
		return false
	}
}

func appendPrivacyLaunchArgs(args []string, cfg *config.Config) []string {
	if browserPrivacyHardenedLaunchArgsEnabled(cfg) {
		args = appendSwitchIfMissing(args, "--disable-non-proxied-udp")
		args = appendSwitchValueIfMissing(args, "--webrtc-ip-handling-policy", "disable_non_proxied_udp")
		args = appendSwitchValueIfMissing(args, "--force-webrtc-ip-handling-policy", "disable_non_proxied_udp")
		args = appendSwitchIfMissing(args, "--disable-quic")
		args = appendSwitchIfMissing(args, "--dns-prefetch-disable")
	}
	if browserPrivacyDisableWebGPUEnabled(cfg) {
		args = mergeDisableFeatures(args, webGPUDisableFeatures)
	}
	if browserSecureDNSEnabled(cfg) {
		options := browserSecureDNSOptions(cfg)
		args = appendSwitchValueIfMissing(args, "--dns-over-https-mode", options.Mode)
		args = appendSwitchValueIfMissing(args, "--dns-over-https-templates", strings.Join(options.Templates, " "))
	}
	return args
}

func appendSwitchIfMissing(args []string, key string) []string {
	if hasSwitch(args, key) {
		return args
	}
	return append(args, key)
}

func appendSwitchValueIfMissing(args []string, key string, value string) []string {
	if strings.TrimSpace(value) == "" || hasSwitch(args, key) {
		return args
	}
	return append(args, key+"="+value)
}

func hasSwitch(args []string, key string) bool {
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.EqualFold(arg, key) || strings.HasPrefix(strings.ToLower(arg), strings.ToLower(key)+"=") {
			return true
		}
	}
	return false
}

func mergeDisableFeatures(args []string, required []string) []string {
	if len(required) == 0 {
		return args
	}
	requiredSet := make([]string, 0, len(required))
	for _, item := range required {
		value := strings.TrimSpace(item)
		if value != "" {
			requiredSet = append(requiredSet, value)
		}
	}
	if len(requiredSet) == 0 {
		return args
	}
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		key, value, found := strings.Cut(arg, "=")
		if found && strings.EqualFold(strings.TrimSpace(key), "--disable-features") {
			args[i] = "--disable-features=" + mergeFeatureList(value, requiredSet)
			return args
		}
		if strings.EqualFold(arg, "--disable-features") && i+1 < len(args) {
			args[i+1] = mergeFeatureList(args[i+1], requiredSet)
			return args
		}
	}
	return append(args, "--disable-features="+strings.Join(requiredSet, ","))
}

func mergeFeatureList(existing string, required []string) string {
	features := make([]string, 0, len(required)+4)
	seen := map[string]struct{}{}
	for _, item := range strings.Split(existing, ",") {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		features = append(features, value)
	}
	for _, item := range required {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		features = append(features, value)
	}
	return strings.Join(features, ",")
}

func effectiveLaunchLanguage(args []string) string {
	for i := len(args) - 1; i >= 0; i-- {
		key, value, found := strings.Cut(strings.TrimSpace(args[i]), "=")
		if found && strings.EqualFold(strings.TrimSpace(key), "--lang") {
			return strings.TrimSpace(value)
		}
	}
	return "en-US"
}
