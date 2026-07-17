package backend

import (
	"ant-chrome/backend/internal/proxy"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) BrowserProxyTestSpeed(proxyId string) ProxyTestResult {
	proxies := a.getLatestProxies()
	result := proxy.SpeedTest(proxyId, proxies, a.proxySpeedTestConfig())
	if a.browserMgr.ProxyDAO != nil {
		testedAt := time.Now().Format(time.RFC3339)
		_ = a.browserMgr.ProxyDAO.UpdateSpeedResult(proxyId, result.Ok, result.LatencyMs, testedAt)
	}
	return buildProxyTestResult(result)
}

const (
	defaultProxySpeedConcurrency = 5
	maxProxySpeedConcurrency     = 10
)

func (a *App) BrowserProxyBatchTestSpeed(proxyIds []string, concurrency int) []ProxyTestResult {
	if len(proxyIds) == 0 {
		return []ProxyTestResult{}
	}
	if concurrency <= 0 {
		concurrency = defaultProxySpeedConcurrency
	}
	if concurrency > maxProxySpeedConcurrency {
		concurrency = maxProxySpeedConcurrency
	}
	if concurrency > len(proxyIds) {
		concurrency = len(proxyIds)
	}

	proxies := a.getLatestProxies()
	results := make([]ProxyTestResult, len(proxyIds))
	type speedJob struct {
		Idx     int
		ProxyId string
	}
	jobs := make(chan speedJob, len(proxyIds))
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				result := proxy.SpeedTest(job.ProxyId, proxies, a.proxySpeedTestConfig())
				if a.browserMgr.ProxyDAO != nil {
					testedAt := time.Now().Format(time.RFC3339)
					_ = a.browserMgr.ProxyDAO.UpdateSpeedResult(job.ProxyId, result.Ok, result.LatencyMs, testedAt)
				}
				item := buildProxyTestResult(result)
				results[job.Idx] = item
				if a.ctx != nil {
					runtime.EventsEmit(a.ctx, "proxy:speed:result", item)
				}
			}
		}()
	}

	for i, proxyID := range proxyIds {
		jobs <- speedJob{Idx: i, ProxyId: proxyID}
	}
	close(jobs)
	wg.Wait()
	return results
}

func (a *App) testProxySpeed(proxyId string, proxies []BrowserProxy) proxy.TestResult {
	return proxy.SpeedTest(proxyId, proxies, a.proxySpeedTestConfig())
}

func (a *App) BrowserProxyCheckIPHealth(proxyId string) ProxyIPHealthResult {
	proxies := a.getLatestProxies()
	data, err := proxy.FetchIPHealthInfo(proxyId, proxies, a.proxyIPHealthConfig())
	result := buildProxyIPHealthResult(proxyId, data, err)
	a.persistProxyIPHealthResult(result)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "proxy:iphealth:result", result)
	}
	return result
}

func (a *App) BrowserProxyBatchCheckIPHealth(proxyIds []string, concurrency int) []ProxyIPHealthResult {
	if len(proxyIds) == 0 {
		return []ProxyIPHealthResult{}
	}
	if concurrency <= 0 {
		concurrency = 10
	}
	if concurrency > len(proxyIds) {
		concurrency = len(proxyIds)
	}

	proxies := a.getLatestProxies()
	results := make([]ProxyIPHealthResult, len(proxyIds))
	type healthJob struct {
		Idx     int
		ProxyId string
	}
	jobs := make(chan healthJob, len(proxyIds))
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				data, err := proxy.FetchIPHealthInfo(job.ProxyId, proxies, a.proxyIPHealthConfig())
				result := buildProxyIPHealthResult(job.ProxyId, data, err)
				a.persistProxyIPHealthResult(result)
				results[job.Idx] = result
				if a.ctx != nil {
					runtime.EventsEmit(a.ctx, "proxy:iphealth:result", result)
				}
			}
		}()
	}

	for i, proxyID := range proxyIds {
		jobs <- healthJob{Idx: i, ProxyId: proxyID}
	}
	close(jobs)
	wg.Wait()
	return results
}

func buildProxyIPHealthResult(proxyId string, data map[string]interface{}, err error) ProxyIPHealthResult {
	if data == nil {
		data = map[string]interface{}{}
	}

	if err != nil {
		data["error"] = err.Error()
		return ProxyIPHealthResult{
			ProxyId:   proxyId,
			Ok:        false,
			Source:    mapStringDefault(data, "_source", "ip_health"),
			Error:     err.Error(),
			RawData:   data,
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
	}

	country := mapString(data, "country")
	countryCode := strings.ToUpper(strings.TrimSpace(mapString(data, "countryCode")))
	if len(countryCode) != 2 {
		countryCode = normalizeCountryCode(country)
	}
	if len(countryCode) != 2 {
		countryCode = ""
	}

	return ProxyIPHealthResult{
		ProxyId:        proxyId,
		Ok:             true,
		Source:         mapStringDefault(data, "_source", "ip_health"),
		Error:          "",
		IP:             mapString(data, "ip"),
		FraudScore:     mapInt64(data, "fraudScore"),
		IsResidential:  mapBool(data, "isResidential"),
		IsBroadcast:    mapBool(data, "isBroadcast"),
		Country:        country,
		CountryCode:    countryCode,
		Region:         mapString(data, "region"),
		City:           mapString(data, "city"),
		AsOrganization: mapString(data, "asOrganization"),
		RawData:        data,
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}
}

func mapStringDefault(data map[string]interface{}, key string, fallback string) string {
	value := strings.TrimSpace(mapString(data, key))
	if value == "" {
		return fallback
	}
	return value
}

func (a *App) persistProxyIPHealthResult(result ProxyIPHealthResult) {
	if a.browserMgr.ProxyDAO == nil {
		return
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return
	}
	_ = a.browserMgr.ProxyDAO.UpdateIPHealthResult(result.ProxyId, string(payload))
}

func mapString(data map[string]interface{}, key string) string {
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}
	switch item := value.(type) {
	case string:
		return item
	default:
		return strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(stringifyAny(item)), "\n", " "))
	}
}

func stringifyAny(value interface{}) string {
	switch item := value.(type) {
	case string:
		return item
	case float64:
		if item == float64(int64(item)) {
			return strconv.FormatInt(int64(item), 10)
		}
		return strconv.FormatFloat(item, 'f', -1, 64)
	case bool:
		if item {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func mapInt64(data map[string]interface{}, key string) int64 {
	value, ok := data[key]
	if !ok || value == nil {
		return 0
	}
	switch item := value.(type) {
	case float64:
		return int64(item)
	case int64:
		return item
	case int:
		return int64(item)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(item), 10, 64)
		return n
	default:
		return 0
	}
}

func mapBool(data map[string]interface{}, key string) bool {
	value, ok := data[key]
	if !ok || value == nil {
		return false
	}
	switch item := value.(type) {
	case bool:
		return item
	case string:
		return strings.EqualFold(strings.TrimSpace(item), "true") || item == "1"
	case float64:
		return item != 0
	case int:
		return item != 0
	default:
		return false
	}
}
