package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
	"ant-chrome/backend/internal/proxy"
	"strings"
	"sync"
	"time"
)

func (a *App) BrowserProxyList() []BrowserProxy {
	return browser.ListProxiesWithFallback(a.browserMgr.ProxyDAO, a.config.Browser.Proxies)
}

// BrowserProxyListGroups 获取所有代理分组名称
func (a *App) BrowserProxyListGroups() []string {
	return browser.ListProxyGroups(a.browserMgr.ProxyDAO)
}

// BrowserProxyListByGroup 按分组名称查询代理
func (a *App) BrowserProxyListByGroup(groupName string) []BrowserProxy {
	return browser.ListProxiesByGroupWithFallback(a.browserMgr.ProxyDAO, groupName, a.config.Browser.Proxies)
}

// ValidateProxyConfig 验证代理配置是否支持
func (a *App) ValidateProxyConfig(proxyConfig string, proxyId string) ProxyValidationResult {
	proxies := a.getLatestProxies()
	supported, errorMsg := proxy.ValidateProxyConfig(proxyConfig, proxies, proxyId)
	return ProxyValidationResult{
		Supported: supported,
		ErrorMsg:  errorMsg,
	}
}

// TestProxyConnectivity 测试代理连通性
func (a *App) TestProxyConnectivity(proxyId string, proxyConfig string) ProxyTestResult {
	proxies := a.getLatestProxies()
	result := proxy.TestConnectivity(proxyId, proxyConfig, proxies, nil)
	return ProxyTestResult{ProxyId: result.ProxyId, Ok: result.Ok, LatencyMs: result.LatencyMs, Error: result.Error}
}

// TestProxyRealConnectivity 通过真实 HTTP 请求测试代理连通性（Wails 绑定）
// 参考 Clash URLTest 策略：多 URL fallback + 复用桥接 + TCP ping 降级
func (a *App) TestProxyRealConnectivity(proxyId string) ProxyTestResult {
	proxies := a.getLatestProxies()
	result := proxy.TestRealConnectivityWithRuntimeConfig(proxyId, proxies, a.xrayMgr, a.singboxMgr, a.clashMgr, config.NormalizeBrowserConnectorType(a.config.Browser.DefaultConnectorType), nil)
	return ProxyTestResult{ProxyId: result.ProxyId, Ok: result.Ok, LatencyMs: result.LatencyMs, Error: result.Error}
}

// BrowserProxyWarmupBridge 只预热本地代理桥接，不执行外网测速。
func (a *App) BrowserProxyWarmupBridge(proxyId string) ProxyBridgeWarmupResult {
	proxies := a.getLatestProxies()
	return a.warmupProxyBridge(proxyId, "", proxies)
}

// BrowserProxyWarmupBridgeWithConfig 预热指定代理配置，proxyConfig 仅本次预热生效。
func (a *App) BrowserProxyWarmupBridgeWithConfig(proxyId string, proxyConfig string) ProxyBridgeWarmupResult {
	proxies := a.getLatestProxies()
	return a.warmupProxyBridge(proxyId, proxyConfig, proxies)
}

// BrowserProxyBatchWarmupBridge 批量预热代理桥接，concurrency 控制并发数（默认 5）。
func (a *App) BrowserProxyBatchWarmupBridge(proxyIds []string, concurrency int) []ProxyBridgeWarmupResult {
	if len(proxyIds) == 0 {
		return []ProxyBridgeWarmupResult{}
	}
	if concurrency <= 0 {
		concurrency = 5
	}
	if concurrency > len(proxyIds) {
		concurrency = len(proxyIds)
	}

	proxies := a.getLatestProxies()
	results := make([]ProxyBridgeWarmupResult, len(proxyIds))
	type warmupJob struct {
		idx     int
		proxyId string
	}
	jobs := make(chan warmupJob, len(proxyIds))
	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results[job.idx] = a.warmupProxyBridge(job.proxyId, "", proxies)
			}
		}()
	}
	for i, proxyID := range proxyIds {
		jobs <- warmupJob{idx: i, proxyId: proxyID}
	}
	close(jobs)
	wg.Wait()
	return results
}

func (a *App) warmupProxyBridge(proxyId string, proxyConfig string, proxies []BrowserProxy) ProxyBridgeWarmupResult {
	startedAt := time.Now()
	proxyId = strings.TrimSpace(proxyId)
	result := ProxyBridgeWarmupResult{ProxyId: proxyId}
	src := strings.TrimSpace(resolveProxyConfigForApp(proxyConfig, proxies, proxyId))
	if src == "" {
		result.Error = "代理配置为空"
		return result
	}
	if strings.EqualFold(src, "direct://") {
		result.Ok = true
		result.Engine = "direct"
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result
	}
	if !proxy.RequiresBridge(src, proxies, proxyId) && !proxy.RequiresLocalProxyBridgeForBrowser(src) && !proxy.IsSingBoxProtocol(src) {
		result.Ok = true
		result.Engine = "none"
		result.LatencyMs = time.Since(startedAt).Milliseconds()
		return result
	}

	var socksURL string
	var err error
	connectorType := config.NormalizeBrowserConnectorType(a.config.Browser.DefaultConnectorType)
	if connectorType == config.BrowserConnectorMihomo {
		result.Engine = "mihomo"
		if a.clashMgr == nil {
			result.Error = "mihomo 管理器不可用"
			return result
		}
		socksURL, err = a.clashMgr.EnsureNodeBridge(src, proxies, proxyId)
	} else if proxy.IsSingBoxProtocol(src) {
		result.Engine = "sing-box"
		if a.singboxMgr == nil {
			result.Error = "sing-box 管理器不可用"
			return result
		}
		socksURL, err = a.singboxMgr.EnsureBridge(src, proxies, proxyId)
	} else {
		result.Engine = "xray"
		if a.xrayMgr == nil {
			result.Error = "xray 管理器不可用"
			return result
		}
		socksURL, err = a.xrayMgr.EnsureBridge(src, proxies, proxyId)
	}
	result.LatencyMs = time.Since(startedAt).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Ok = true
	result.SocksURL = socksURL
	return result
}

func resolveProxyConfigForApp(proxyConfig string, proxies []BrowserProxy, proxyId string) string {
	proxyConfig = strings.TrimSpace(proxyConfig)
	proxyId = strings.TrimSpace(proxyId)
	if proxyId == "" {
		return proxyConfig
	}
	for _, item := range proxies {
		if strings.EqualFold(item.ProxyId, proxyId) {
			return strings.TrimSpace(item.ProxyConfig)
		}
	}
	return proxyConfig
}

// getLatestProxies 获取最新的代理列表，优先从数据库读取
func (a *App) getLatestProxies() []BrowserProxy {
	return browser.LatestProxiesWithFallback(a.browserMgr.ProxyDAO, a.config.Browser.Proxies)
}
