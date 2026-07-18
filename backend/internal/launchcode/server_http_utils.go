package launchcode

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"facade/backend/internal/browser"
)

// localhostMiddleware 只允许 127.0.0.1 访问
func (s *LaunchServer) localhostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || host != "127.0.0.1" {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"ok":    false,
				"error": "forbidden: only localhost is allowed",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func normalizeStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		v := strings.TrimSpace(item)
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeLaunchRequestParams(params LaunchRequestParams) LaunchRequestParams {
	params.LaunchArgs = normalizeStringSlice(params.LaunchArgs)
	params.StartURLs = normalizeStringSlice(params.StartURLs)
	params.ProxyId = strings.TrimSpace(params.ProxyId)
	params.ProxyConfig = strings.TrimSpace(params.ProxyConfig)
	return params
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func profileDirectCDP(profile *browser.Profile) (int, string) {
	if profile == nil || !profile.DebugReady || profile.DebugPort <= 0 {
		return 0, ""
	}
	return profile.DebugPort, fmt.Sprintf("http://127.0.0.1:%d", profile.DebugPort)
}
