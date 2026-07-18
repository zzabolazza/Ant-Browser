package proxy

import (
	"facade/backend/internal/config"
	"strings"
)

func NormalizeBrowserProxies(proxies []config.BrowserProxy, generateID func() string) []config.BrowserProxy {
	normalized := make([]config.BrowserProxy, 0, len(proxies)+1)
	for i, item := range proxies {
		proxyName := strings.TrimSpace(item.ProxyName)
		proxyConfig := strings.TrimSpace(item.ProxyConfig)
		if proxyName == "" || proxyConfig == "" {
			continue
		}

		proxyID := strings.TrimSpace(item.ProxyId)
		if proxyID == "" && generateID != nil {
			proxyID = generateID()
		}

		normalized = append(normalized, config.BrowserProxy{
			ProxyId:     proxyID,
			ProxyName:   proxyName,
			ProxyConfig: proxyConfig,
			GroupName:   strings.TrimSpace(item.GroupName),
			SortOrder:   i,
		})
	}

	return ensureBuiltinDirectProxy(normalized)
}

func ensureBuiltinDirectProxy(proxies []config.BrowserProxy) []config.BrowserProxy {
	const directProxyID = "__direct__"
	for _, item := range proxies {
		if item.ProxyId == directProxyID {
			return proxies
		}
	}

	builtin := config.BrowserProxy{
		ProxyId:     directProxyID,
		ProxyName:   "直连（不走代理）",
		ProxyConfig: "direct://",
	}
	return append([]config.BrowserProxy{builtin}, proxies...)
}
