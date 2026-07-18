package proxy

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"facade/backend/internal/config"
	"facade/backend/internal/logger"
)

const DefaultSpeedTestURL = "http://www.gstatic.com/generate_204"

type SpeedTestConfig struct {
	Timeout        time.Duration
	TCPTimeout     time.Duration
	URLs           []string
	ExpectedStatus []int
}

var DefaultSpeedTestConfig = SpeedTestConfig{
	Timeout:    3 * time.Second,
	TCPTimeout: 3 * time.Second,
}

// SpeedTest 对原生代理执行轻量 HTTP 延迟测试。
func SpeedTest(
	proxyId string,
	proxies []config.BrowserProxy,
	cfg *SpeedTestConfig,
) TestResult {
	return lightHTTPDelayTest(proxyId, proxies, cfg)
}

func lightHTTPDelayTest(
	proxyId string,
	proxies []config.BrowserProxy,
	cfg *SpeedTestConfig,
) TestResult {
	log := logger.New("SpeedTest")

	if cfg == nil {
		c := DefaultSpeedTestConfig
		cfg = &c
	}

	src := resolveProxyConfig("", proxies, proxyId)
	if src == "" {
		return TestResult{ProxyId: proxyId, Ok: false, Engine: "native", Error: "代理配置为空"}
	}
	if strings.EqualFold(src, "direct://") {
		return TestResult{ProxyId: proxyId, Ok: true, LatencyMs: 0, Engine: ProtocolDirect}
	}

	engine := DetectProxyProtocol(src)
	if engine == "unknown" {
		return TestResult{ProxyId: proxyId, Ok: false, Engine: "native", Error: "不支持的代理协议，仅支持 http / https / socks5"}
	}

	testURLs := speedTestTargetURLs(cfg)
	if len(testURLs) == 0 {
		return TestResult{ProxyId: proxyId, Ok: false, Engine: engine, Error: "测速目标 URL 为空"}
	}

	log.Info("开始代理测速",
		logger.F("proxy_id", proxyId),
		logger.F("engine", engine),
		logger.F("timeout_ms", cfg.Timeout.Milliseconds()),
		logger.F("targets", strings.Join(testURLs, ",")),
	)

	client, err := buildSpeedTestHTTPClient(src, proxyId, proxies, cfg)
	if err != nil {
		log.Warn("代理测速 HTTP 客户端创建失败",
			logger.F("proxy_id", proxyId),
			logger.F("error", err.Error()),
		)
		return TestResult{ProxyId: proxyId, Ok: false, Engine: engine, Error: err.Error()}
	}

	var lastErr error
	var lastLatency int64
	for _, testURL := range testURLs {
		latency, statusCode, err := doSpeedTestRequest(client, testURL)
		lastLatency = latency
		if err != nil {
			lastErr = err
			log.Warn("代理测速请求失败",
				logger.F("proxy_id", proxyId),
				logger.F("engine", engine),
				logger.F("url", testURL),
				logger.F("latency_ms", latency),
				logger.F("error", err.Error()),
			)
			continue
		}
		if speedTestStatusOK(statusCode, cfg) {
			log.Info("代理测速成功",
				logger.F("proxy_id", proxyId),
				logger.F("engine", engine),
				logger.F("url", testURL),
				logger.F("status", statusCode),
				logger.F("latency_ms", latency),
			)
			return TestResult{ProxyId: proxyId, Ok: true, LatencyMs: latency, Engine: engine}
		}
		lastErr = fmt.Errorf("HTTP %d", statusCode)
		log.Warn("代理测速状态码不符合预期",
			logger.F("proxy_id", proxyId),
			logger.F("engine", engine),
			logger.F("url", testURL),
			logger.F("status", statusCode),
			logger.F("latency_ms", latency),
		)
	}

	if lastErr != nil {
		return TestResult{ProxyId: proxyId, Ok: false, LatencyMs: lastLatency, Engine: engine, Error: lastErr.Error()}
	}
	return TestResult{ProxyId: proxyId, Ok: false, LatencyMs: lastLatency, Engine: engine, Error: "测速失败"}
}

func buildSpeedTestHTTPClient(
	src string,
	proxyId string,
	proxies []config.BrowserProxy,
	cfg *SpeedTestConfig,
) (*http.Client, error) {
	timeout := DefaultSpeedTestConfig.Timeout
	prepareTimeout := DefaultSpeedTestConfig.TCPTimeout
	if cfg != nil {
		if cfg.Timeout > 0 {
			timeout = cfg.Timeout
		}
		if cfg.TCPTimeout > 0 {
			prepareTimeout = cfg.TCPTimeout
		}
	}
	if prepareTimeout <= 0 {
		prepareTimeout = timeout
	}

	type clientResult struct {
		client *http.Client
		err    error
	}
	resultCh := make(chan clientResult, 1)
	go func() {
		client, err := buildProxyHTTPClient(src, proxyId, proxies, timeout)
		resultCh <- clientResult{client: client, err: err}
	}()

	timer := time.NewTimer(prepareTimeout)
	defer timer.Stop()
	select {
	case result := <-resultCh:
		return result.client, result.err
	case <-timer.C:
		return nil, fmt.Errorf("代理准备超时（%dms）", prepareTimeout.Milliseconds())
	}
}

func speedTestTargetURLs(cfg *SpeedTestConfig) []string {
	if cfg != nil {
		if urls := uniqueSpeedTestURLs(cfg.URLs); len(urls) > 0 {
			return urls
		}
	}
	return []string{DefaultSpeedTestURL}
}

func doSpeedTestRequest(client *http.Client, testURL string) (int64, int, error) {
	latency, statusCode, err := doSpeedTestRequestWithMethod(client, http.MethodHead, testURL)
	if err != nil || statusCode != http.StatusMethodNotAllowed {
		if err != nil {
			return latency, statusCode, err
		}
		return latency, statusCode, nil
	}
	return doSpeedTestRequestWithMethod(client, http.MethodGet, testURL)
}

func doSpeedTestRequestWithMethod(client *http.Client, method string, testURL string) (int64, int, error) {
	start := time.Now()
	req, err := http.NewRequest(method, testURL, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("测速请求创建失败: %w", err)
	}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return latency, 0, err
	}
	_ = resp.Body.Close()
	return latency, resp.StatusCode, nil
}

func speedTestStatusOK(statusCode int, cfg *SpeedTestConfig) bool {
	if cfg != nil && len(cfg.ExpectedStatus) > 0 {
		for _, expected := range cfg.ExpectedStatus {
			if statusCode == expected {
				return true
			}
		}
		return false
	}
	return isSpeedTestSuccessStatus(statusCode)
}
