package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"facade/backend/internal/config"
	"facade/backend/internal/logger"

	xproxy "golang.org/x/net/proxy"
)

// BuildProxyHTTPClient 根据原生代理链接构建 HTTP 客户端。
func BuildProxyHTTPClient(
	src string,
	proxyId string,
	proxies []config.BrowserProxy,
	timeout time.Duration,
) (*http.Client, error) {
	return buildProxyHTTPClient(src, proxyId, proxies, timeout)
}

func buildProxyHTTPClient(
	src string,
	proxyId string,
	proxies []config.BrowserProxy,
	timeout time.Duration,
) (*http.Client, error) {
	src = strings.TrimSpace(resolveProxyConfig(src, proxies, proxyId))
	log := logger.New("ProxyHTTPClient")
	if src == "" || strings.EqualFold(src, "direct://") {
		log.Info("使用直连 HTTP 客户端", logger.F("proxy_id", proxyId))
		return &http.Client{Timeout: timeout}, nil
	}

	normalized, err := ParseNativeProxyURL(src)
	if err != nil {
		log.Warn("代理配置无效", logger.F("proxy_id", proxyId), logger.F("error", err.Error()))
		return nil, err
	}
	protocol := DetectProxyProtocol(normalized)
	log.Info("代理 HTTP 客户端",
		logger.F("proxy_id", proxyId),
		logger.F("protocol", protocol),
	)

	parsed, err := url.Parse(normalized)
	if err != nil {
		return nil, fmt.Errorf("代理地址解析失败: %w", err)
	}

	switch protocol {
	case ProtocolSOCKS5:
		var auth *xproxy.Auth
		if parsed.User != nil {
			pass, _ := parsed.User.Password()
			auth = &xproxy.Auth{User: parsed.User.Username(), Password: pass}
		}
		dialer, err := xproxy.SOCKS5("tcp", parsed.Host, auth, xproxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("SOCKS5 dialer 创建失败: %w", err)
		}
		contextDialer, ok := dialer.(xproxy.ContextDialer)
		if !ok {
			return nil, fmt.Errorf("SOCKS5 dialer 不支持 ContextDialer")
		}
		return &http.Client{
			Transport: &http.Transport{DialContext: contextDialer.DialContext},
			Timeout:   timeout,
		}, nil
	case ProtocolHTTP, ProtocolHTTPS:
		return &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(parsed)},
			Timeout:   timeout,
		}, nil
	default:
		return nil, fmt.Errorf("不支持的代理协议: %s", protocol)
	}
}
