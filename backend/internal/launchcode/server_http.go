package launchcode

import "net/http"

func (s *LaunchServer) buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/profiles", s.handleProfiles)
	mux.HandleFunc("/api/profiles/", s.handleProfileByID)
	mux.HandleFunc("/api/runtime/session", s.handleRuntimeSession)
	mux.HandleFunc("/api/runtime/status", s.handleRuntimeStatus)
	mux.HandleFunc("/api/runtime/stop", s.handleRuntimeStop)
	mux.HandleFunc("/api/launch", s.handleLaunchWithBody)
	mux.HandleFunc("/api/launch/logs", s.handleLaunchLogs)
	mux.HandleFunc("/api/launch/", s.handleLaunch)
	return mux
}

func (s *LaunchServer) buildHandler(includeLocalhost bool) http.Handler {
	var handler http.Handler = s.buildMux()
	handler = s.apiAuthMiddleware(handler)
	if includeLocalhost {
		handler = s.localhostMiddleware(handler)
	}
	return handler
}

// handleHealth GET /api/health
func (s *LaunchServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

// NewTestHandler 返回不含 localhost 限制的 handler，仅供测试使用
func NewTestHandler(s *LaunchServer) http.Handler {
	return s.buildHandler(false)
}
