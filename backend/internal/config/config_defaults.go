package config

import (
	"path/filepath"
	goruntime "runtime"
	"strings"
)

var defaultBrowserStartURLs = []string{}

const DefaultSecureDNSTemplate = "https://chrome.cloudflare-dns.com/dns-query"

func DefaultBrowserStartURLs() []string {
	return append([]string{}, defaultBrowserStartURLs...)
}

// normalizeConfig 对历史配置进行字段补齐，不覆盖用户已配置值。
func normalizeConfig(config *Config) {
	defaultConfig := DefaultConfig()

	if strings.TrimSpace(config.Database.Type) == "" {
		config.Database.Type = defaultConfig.Database.Type
	}
	if strings.TrimSpace(config.Database.SQLite.Path) == "" {
		config.Database.SQLite.Path = defaultConfig.Database.SQLite.Path
	}

	if strings.TrimSpace(config.App.Name) == "" {
		config.App.Name = defaultConfig.App.Name
	}
	if config.App.Window.Width <= 0 {
		config.App.Window.Width = defaultConfig.App.Window.Width
	}
	if config.App.Window.Height <= 0 {
		config.App.Window.Height = defaultConfig.App.Window.Height
	}
	if config.App.Window.MinWidth <= 0 {
		config.App.Window.MinWidth = defaultConfig.App.Window.MinWidth
	}
	if config.App.Window.MinHeight <= 0 {
		config.App.Window.MinHeight = defaultConfig.App.Window.MinHeight
	}
	if config.Runtime.MaxMemoryMB <= 0 {
		config.Runtime.MaxMemoryMB = defaultConfig.Runtime.MaxMemoryMB
	}
	if config.Runtime.GCPercent <= 0 {
		config.Runtime.GCPercent = defaultConfig.Runtime.GCPercent
	}
	config.Backup.WebDAV.URL = strings.TrimSpace(config.Backup.WebDAV.URL)
	config.Backup.WebDAV.Username = strings.TrimSpace(config.Backup.WebDAV.Username)
	config.Backup.WebDAV.RemoteDir = strings.Trim(strings.TrimSpace(config.Backup.WebDAV.RemoteDir), "/\\")

	if strings.TrimSpace(config.Logging.Level) == "" {
		config.Logging.Level = defaultConfig.Logging.Level
	}
	if isLegacyDefaultLogPath(config.Logging.FilePath) || strings.TrimSpace(config.Logging.FilePath) == "" {
		config.Logging.FilePath = defaultConfig.Logging.FilePath
	}
	if strings.TrimSpace(config.Logging.Format) == "" {
		config.Logging.Format = defaultConfig.Logging.Format
	}
	if config.Logging.BufferSize <= 0 {
		config.Logging.BufferSize = defaultConfig.Logging.BufferSize
	}
	if config.Logging.AsyncQueueSize <= 0 {
		config.Logging.AsyncQueueSize = defaultConfig.Logging.AsyncQueueSize
	}
	if config.Logging.FlushIntervalMs <= 0 {
		config.Logging.FlushIntervalMs = defaultConfig.Logging.FlushIntervalMs
	}
	if config.Logging.Rotation.MaxSizeMB <= 0 {
		config.Logging.Rotation.MaxSizeMB = defaultConfig.Logging.Rotation.MaxSizeMB
	}
	if config.Logging.Rotation.MaxAge <= 0 {
		config.Logging.Rotation.MaxAge = defaultConfig.Logging.Rotation.MaxAge
	}
	if config.Logging.Rotation.MaxBackups <= 0 {
		config.Logging.Rotation.MaxBackups = defaultConfig.Logging.Rotation.MaxBackups
	}
	if strings.TrimSpace(config.Logging.Rotation.TimeInterval) == "" {
		config.Logging.Rotation.TimeInterval = defaultConfig.Logging.Rotation.TimeInterval
	}

	interceptorAllZero := !config.Logging.Interceptor.Enabled &&
		!config.Logging.Interceptor.LogParameters &&
		!config.Logging.Interceptor.LogResults &&
		config.Logging.Interceptor.SensitiveFields == nil
	if interceptorAllZero {
		config.Logging.Interceptor = cloneInterceptorConfig(defaultConfig.Logging.Interceptor)
	} else if config.Logging.Interceptor.SensitiveFields == nil {
		config.Logging.Interceptor.SensitiveFields = append([]string{}, defaultConfig.Logging.Interceptor.SensitiveFields...)
	}

	if strings.TrimSpace(config.Browser.UserDataRoot) == "" {
		config.Browser.UserDataRoot = defaultConfig.Browser.UserDataRoot
	}
	if len(config.Browser.DefaultFingerprintArgs) == 0 {
		config.Browser.DefaultFingerprintArgs = append([]string{}, defaultConfig.Browser.DefaultFingerprintArgs...)
	}
	if len(config.Browser.DefaultLaunchArgs) == 0 {
		config.Browser.DefaultLaunchArgs = append([]string{}, defaultConfig.Browser.DefaultLaunchArgs...)
	}
	if config.Browser.DefaultStartURLs == nil {
		config.Browser.DefaultStartURLs = append([]string{}, defaultConfig.Browser.DefaultStartURLs...)
	} else if isLegacyVerificationStartURLs(config.Browser.DefaultStartURLs) {
		config.Browser.DefaultStartURLs = []string{}
	}
	normalizeBrowserSecureDNSConfig(&config.Browser.SecureDNS, defaultConfig.Browser.SecureDNS)
	normalizeBrowserPrivacyConfig(&config.Browser.Privacy, defaultConfig.Browser.Privacy)
	if config.Browser.LightStartEnabled == nil {
		config.Browser.LightStartEnabled = defaultConfig.Browser.LightStartEnabled
	}
	if config.Browser.StartReadyTimeoutMs <= 0 {
		config.Browser.StartReadyTimeoutMs = defaultConfig.Browser.StartReadyTimeoutMs
	}
	if config.Browser.StartStableWindowMs <= 0 {
		config.Browser.StartStableWindowMs = defaultConfig.Browser.StartStableWindowMs
	}
	if config.Browser.DefaultBookmarks == nil {
		config.Browser.DefaultBookmarks = []BrowserBookmark{}
	}
	if config.Browser.Cores == nil {
		config.Browser.Cores = []BrowserCore{}
	}
	if config.Browser.Proxies == nil {
		config.Browser.Proxies = []BrowserProxy{}
	}
	if config.Browser.Profiles == nil {
		config.Browser.Profiles = []BrowserProfileConfig{}
	}
	if config.ProxyCheck.PrepareTimeoutMs <= 0 {
		config.ProxyCheck.PrepareTimeoutMs = defaultConfig.ProxyCheck.PrepareTimeoutMs
	}
	if strings.TrimSpace(config.ProxyCheck.SpeedTargetID) == "" {
		config.ProxyCheck.SpeedTargetID = defaultConfig.ProxyCheck.SpeedTargetID
	}
	if strings.TrimSpace(config.ProxyCheck.IPHealthTargetID) == "" {
		config.ProxyCheck.IPHealthTargetID = defaultConfig.ProxyCheck.IPHealthTargetID
	}
	if len(config.ProxyCheck.Targets) == 0 {
		config.ProxyCheck.Targets = append([]ProxyCheckTarget{}, defaultConfig.ProxyCheck.Targets...)
	}

	if config.LaunchServer.Port <= 0 {
		config.LaunchServer.Port = defaultConfig.LaunchServer.Port
	}
	config.LaunchServer.Auth.APIKey = strings.TrimSpace(config.LaunchServer.Auth.APIKey)
	if strings.TrimSpace(config.LaunchServer.Auth.Header) == "" {
		config.LaunchServer.Auth.Header = defaultConfig.LaunchServer.Auth.Header
	}
}

func cloneInterceptorConfig(src InterceptorConfig) InterceptorConfig {
	dst := src
	dst.SensitiveFields = append([]string{}, src.SensitiveFields...)
	return dst
}

func isLegacyDefaultLogPath(path string) bool {
	return strings.EqualFold(filepath.ToSlash(strings.TrimSpace(path)), "logs/app.log")
}

func isLegacyVerificationStartURLs(urls []string) bool {
	legacy := []string{"https://ippure.com/", "https://iplark.com/", "https://ping0.cc/"}
	if len(urls) != len(legacy) {
		return false
	}
	for i, url := range urls {
		if !strings.EqualFold(strings.TrimSpace(url), legacy[i]) {
			return false
		}
	}
	return true
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Type: "sqlite",
			SQLite: SQLiteConfig{
				Path: "data/app.db",
			},
		},
		App: AppConfig{
			Name: "Facade",
			Window: WindowConfig{
				Width:     1750,
				Height:    1000,
				MinWidth:  1200,
				MinHeight: 700,
			},
		},
		Runtime: RuntimeConfig{
			MaxMemoryMB: 0,
			GCPercent:   100,
		},
		Backup: BackupConfig{
			WebDAV: WebDAVConfig{},
		},
		Browser: BrowserConfig{
			UserDataRoot:           "data",
			DefaultFingerprintArgs: defaultFingerprintArgsForOS(goruntime.GOOS),
			DefaultLaunchArgs:      []string{"--disable-sync", "--no-first-run"},
			DefaultStartURLs:       DefaultBrowserStartURLs(),
			SecureDNS: BrowserSecureDNSConfig{
				Enabled:   boolPtr(true),
				Mode:      "secure",
				Templates: []string{DefaultSecureDNSTemplate},
			},
			Privacy: BrowserPrivacyConfig{
				HardenedLaunchArgsEnabled: boolPtr(true),
				SpoofSpeechVoices:         boolPtr(true),
				DisableWebGPU:             boolPtr(true),
				ExitConsistencyCheck:      "warn",
			},
			LightStartEnabled:   boolPtr(true),
			RestoreLastSession:  false,
			StartReadyTimeoutMs: 3000,
			StartStableWindowMs: 1200,
		},
		ProxyCheck: ProxyCheckConfig{
			PrepareTimeoutMs: 15000,
			SpeedTargetID:    "gstatic_generate_204",
			IPHealthTargetID: "ippure_info",
			Targets: []ProxyCheckTarget{
				{
					ID:             "gstatic_generate_204",
					Name:           "Google generate_204",
					Type:           "speed",
					URL:            "http://www.gstatic.com/generate_204",
					TimeoutMs:      3000,
					ExpectedStatus: []int{204},
				},
				{
					ID:        "ippure_info",
					Name:      "IPPure 出口信息",
					Type:      "ip_health",
					URL:       "https://my.ippure.com/v1/info",
					Parser:    "json",
					TimeoutMs: 10000,
				},
			},
		},
		Logging: LoggingConfig{
			Level:           "info",
			FileEnabled:     false,
			FilePath:        "data/logs/app.log",
			Format:          "text",
			BufferSize:      4,
			AsyncQueueSize:  1000,
			FlushIntervalMs: 1000,
			Rotation: RotationConfig{
				Enabled:      false,
				MaxSizeMB:    100,
				MaxAge:       7,
				MaxBackups:   5,
				TimeInterval: "daily",
			},
			Interceptor: InterceptorConfig{
				Enabled:         true,
				LogParameters:   true,
				LogResults:      true,
				SensitiveFields: []string{"password", "token", "secret"},
			},
		},
		LaunchServer: LaunchServerConfig{
			Port: DefaultLaunchServerPort,
			Auth: LaunchServerAuthConfig{
				Enabled: false,
				APIKey:  "",
				Header:  DefaultLaunchServerAPIKeyHeader,
			},
		},
	}
}

func normalizeBrowserSecureDNSConfig(cfg *BrowserSecureDNSConfig, defaults BrowserSecureDNSConfig) {
	if cfg == nil {
		return
	}
	if cfg.Enabled == nil {
		cfg.Enabled = boolPtr(browserSecureDNSEnabled(defaults))
	}
	cfg.Mode = strings.TrimSpace(cfg.Mode)
	if cfg.Mode == "" {
		cfg.Mode = strings.TrimSpace(defaults.Mode)
	}
	if cfg.Mode == "" {
		cfg.Mode = "secure"
	}
	cfg.Templates = normalizeNonEmptyConfigStrings(cfg.Templates)
	if len(cfg.Templates) == 0 {
		cfg.Templates = append([]string{}, defaults.Templates...)
	}
	if len(cfg.Templates) == 0 {
		cfg.Templates = []string{DefaultSecureDNSTemplate}
	}
}

func normalizeBrowserPrivacyConfig(cfg *BrowserPrivacyConfig, defaults BrowserPrivacyConfig) {
	if cfg == nil {
		return
	}
	if cfg.HardenedLaunchArgsEnabled == nil {
		cfg.HardenedLaunchArgsEnabled = boolPtr(browserPrivacyBool(defaults.HardenedLaunchArgsEnabled, true))
	}
	if cfg.SpoofSpeechVoices == nil {
		cfg.SpoofSpeechVoices = boolPtr(browserPrivacyBool(defaults.SpoofSpeechVoices, true))
	}
	if cfg.DisableWebGPU == nil {
		cfg.DisableWebGPU = boolPtr(browserPrivacyBool(defaults.DisableWebGPU, true))
	}
	cfg.ExitConsistencyCheck = strings.TrimSpace(cfg.ExitConsistencyCheck)
	if cfg.ExitConsistencyCheck == "" {
		cfg.ExitConsistencyCheck = strings.TrimSpace(defaults.ExitConsistencyCheck)
	}
	if cfg.ExitConsistencyCheck == "" {
		cfg.ExitConsistencyCheck = "warn"
	}
}

func normalizeNonEmptyConfigStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func browserSecureDNSEnabled(cfg BrowserSecureDNSConfig) bool {
	return browserPrivacyBool(cfg.Enabled, true)
}

func browserPrivacyBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultFingerprintArgsForOS(goos string) []string {
	platform := "windows"
	switch strings.ToLower(strings.TrimSpace(goos)) {
	case "darwin":
		platform = "macos"
	case "linux":
		platform = "linux"
	}
	return []string{"--fingerprint-brand=Chrome", "--fingerprint-platform=" + platform}
}
func boolPtr(value bool) *bool {
	v := value
	return &v
}
