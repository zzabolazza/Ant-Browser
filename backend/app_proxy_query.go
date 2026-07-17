package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/proxy"
)

func (a *App) BrowserProxyList() []BrowserProxy {
	return browser.ListProxiesWithFallback(a.browserMgr.ProxyDAO, a.config.Browser.Proxies)
}

func (a *App) BrowserProxyListGroups() []string {
	return browser.ListProxyGroups(a.browserMgr.ProxyDAO)
}

func (a *App) ValidateProxyConfig(proxyConfig string, proxyId string) ProxyValidationResult {
	proxies := a.getLatestProxies()
	supported, errorMsg := proxy.ValidateProxyConfig(proxyConfig, proxies, proxyId)
	return ProxyValidationResult{
		Supported: supported,
		ErrorMsg:  errorMsg,
	}
}

func (a *App) TestProxyConnectivity(proxyId string, proxyConfig string) ProxyTestResult {
	proxies := a.getLatestProxies()
	result := proxy.TestConnectivity(proxyId, proxyConfig, proxies)
	return buildProxyTestResult(result)
}

func (a *App) getLatestProxies() []BrowserProxy {
	return browser.LatestProxiesWithFallback(a.browserMgr.ProxyDAO, a.config.Browser.Proxies)
}
