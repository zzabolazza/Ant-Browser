package backend

import (
	"facade/backend/internal/proxy"
	"fmt"
	"strings"
	"time"
)

func (a *App) browserExitConsistencyWarning(effectiveProxy string) string {
	mode := browserExitConsistencyCheckMode(a.config)
	if mode == "off" || mode == "disabled" || mode == "none" || strings.EqualFold(strings.TrimSpace(effectiveProxy), "direct://") {
		return ""
	}

	targets := []*proxy.IPHealthConfig{a.proxyIPHealthConfig(), {
		URL:     "https://www.cloudflare.com/cdn-cgi/trace",
		Source:  "cloudflare_trace",
		Parser:  "cloudflare_trace",
		Timeout: 5 * time.Second,
	}}
	proxies := a.getLatestProxies()
	codes := []string{}
	ips := []string{}
	for _, target := range targets {
		if target == nil {
			continue
		}
		if target.Timeout <= 0 || target.Timeout > 5*time.Second {
			copyTarget := *target
			copyTarget.Timeout = 5 * time.Second
			target = &copyTarget
		}
		data, err := proxy.FetchIPHealthInfoWithConfig("__exit_check__", effectiveProxy, proxies, target)
		if err != nil {
			continue
		}
		code := countryCodeFromIPHealthData(data)
		if code == "" {
			continue
		}
		codes = append(codes, code)
		if ip := mapString(data, "ip"); ip != "" {
			ips = append(ips, ip)
		}
	}
	if len(codes) < 2 {
		return ""
	}
	first := codes[0]
	for _, code := range codes[1:] {
		if !strings.EqualFold(first, code) {
			return fmt.Sprintf("出口一致性预检发现多个检测目标国家不一致（%s，IP=%s）。请检查 VPN/代理客户端是否为全局或 TUN 模式。", strings.Join(codes, " / "), strings.Join(ips, " / "))
		}
	}
	return ""
}

func countryCodeFromIPHealthData(data map[string]interface{}) string {
	code := strings.ToUpper(strings.TrimSpace(mapString(data, "countryCode")))
	if len(code) == 2 {
		return code
	}
	code = strings.ToUpper(strings.TrimSpace(mapString(data, "loc")))
	if len(code) == 2 {
		return code
	}
	return normalizeCountryCode(mapString(data, "country"))
}
