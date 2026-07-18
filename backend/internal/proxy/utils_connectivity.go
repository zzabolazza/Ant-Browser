package proxy

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"facade/backend/internal/config"
)

// TestConnectivity 通过 TCP 握手测试代理服务器的可达性和延迟。
func TestConnectivity(proxyId string, proxyConfig string, proxies []config.BrowserProxy) TestResult {
	src := strings.TrimSpace(proxyConfig)
	if proxyId != "" {
		for _, item := range proxies {
			if strings.EqualFold(item.ProxyId, proxyId) {
				src = strings.TrimSpace(item.ProxyConfig)
				break
			}
		}
	}
	if src == "" {
		return TestResult{ProxyId: proxyId, Ok: false, Engine: "tcp", Error: "代理配置为空"}
	}
	if strings.EqualFold(src, "direct://") {
		return TestResult{ProxyId: proxyId, Ok: true, LatencyMs: 0, Engine: ProtocolDirect}
	}

	endpoint, err := proxyEndpoint(src)
	if err != nil {
		return TestResult{ProxyId: proxyId, Ok: false, Engine: "tcp", Error: fmt.Sprintf("地址解析失败: %v", err)}
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", endpoint, 10*time.Second)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return TestResult{ProxyId: proxyId, Ok: false, LatencyMs: latency, Engine: "tcp", Error: err.Error()}
	}
	_ = conn.Close()
	return TestResult{ProxyId: proxyId, Ok: true, LatencyMs: latency, Engine: "tcp"}
}

func normalizeSpeedTestURLs(urls []string) []string {
	result := make([]string, 0, len(urls))
	for _, item := range urls {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func uniqueSpeedTestURLs(urls []string) []string {
	result := make([]string, 0, len(urls))
	seen := map[string]struct{}{}
	for _, item := range normalizeSpeedTestURLs(urls) {
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func isSpeedTestSuccessStatus(statusCode int) bool {
	return statusCode == http.StatusNoContent || (statusCode >= 200 && statusCode < 300)
}
