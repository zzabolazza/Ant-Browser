package browser

import "ant-chrome/backend/internal/config"

const DefaultMaxProfileLimit = 20

type DashboardStats struct {
	TotalInstances   int
	RunningInstances int
	ProxyCount       int
	CoreCount        int
	MaxProfileLimit  int
}

func BuildDashboardStats(profiles []Profile, cfg *config.Config) DashboardStats {
	stats := DashboardStats{
		TotalInstances:  len(profiles),
		MaxProfileLimit: DefaultMaxProfileLimit,
	}
	for _, profile := range profiles {
		if profile.Running {
			stats.RunningInstances++
		}
	}
	if cfg != nil {
		stats.ProxyCount = len(cfg.Browser.Proxies)
		stats.CoreCount = len(cfg.Browser.Cores)
		if cfg.App.MaxProfileLimit > 0 {
			stats.MaxProfileLimit = cfg.App.MaxProfileLimit
		}
	}
	return stats
}

func RunningProfiles(profiles []Profile) []Profile {
	result := make([]Profile, 0)
	for _, profile := range profiles {
		if profile.Running {
			result = append(result, profile)
		}
	}
	return result
}
