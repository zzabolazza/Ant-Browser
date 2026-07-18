package proxy

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"facade/backend/internal/config"
)

const (
	ProtocolDirect = "direct"
	ProtocolHTTP   = "http"
	ProtocolHTTPS  = "https"
	ProtocolSOCKS5 = "socks5"
)

// ValidateProxyConfig 验证代理配置是否为浏览器原生可用链接。
// 仅支持 direct://、http://、https://、socks5://。
func ValidateProxyConfig(proxyConfig string, proxies []config.BrowserProxy, proxyId string) (bool, string) {
	src := strings.TrimSpace(proxyConfig)
	if proxyId != "" {
		found := false
		for _, item := range proxies {
			if strings.EqualFold(item.ProxyId, proxyId) {
				src = strings.TrimSpace(item.ProxyConfig)
				found = true
				break
			}
		}
		if !found && src == "" {
			return false, fmt.Sprintf("代理链路不可用：代理池节点已不存在（proxyId=%s）。请重新选择代理后再启动。", proxyId)
		}
	}
	if src == "" {
		return true, ""
	}
	if _, err := ParseNativeProxyURL(src); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ParseNativeProxyURL 解析并校验原生代理链接，返回规范化 URL 字符串。
func ParseNativeProxyURL(proxyConfig string) (string, error) {
	src := strings.TrimSpace(proxyConfig)
	if src == "" {
		return "", fmt.Errorf("代理配置为空")
	}
	if strings.EqualFold(src, "direct://") {
		return "direct://", nil
	}

	parsed, err := url.Parse(src)
	if err != nil {
		return "", fmt.Errorf("代理地址解析失败: %w", err)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	switch scheme {
	case ProtocolHTTP, ProtocolHTTPS, ProtocolSOCKS5:
	default:
		return "", fmt.Errorf("不支持的代理协议 %q，仅支持 direct / http / https / socks5", scheme)
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return "", fmt.Errorf("代理地址缺少主机名")
	}
	portText := strings.TrimSpace(parsed.Port())
	if portText == "" {
		return "", fmt.Errorf("代理地址缺少端口")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("代理端口无效: %s", portText)
	}

	out := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		User:   parsed.User,
	}
	return out.String(), nil
}

// DetectProxyProtocol 返回原生协议名；未知协议返回 "unknown"。
func DetectProxyProtocol(proxyConfig string) string {
	src := strings.TrimSpace(proxyConfig)
	if src == "" || strings.EqualFold(src, "direct://") {
		return ProtocolDirect
	}
	parsed, err := url.Parse(src)
	if err != nil {
		return "unknown"
	}
	switch strings.ToLower(strings.TrimSpace(parsed.Scheme)) {
	case ProtocolHTTP:
		return ProtocolHTTP
	case ProtocolHTTPS:
		return ProtocolHTTPS
	case ProtocolSOCKS5:
		return ProtocolSOCKS5
	default:
		return "unknown"
	}
}

// IsNativeProxyConfig 判断是否为原生可用代理链接。
func IsNativeProxyConfig(proxyConfig string) bool {
	_, err := ParseNativeProxyURL(proxyConfig)
	return err == nil
}
