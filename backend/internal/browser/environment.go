package browser

import (
	"path/filepath"
	"strings"
)

// GetProxyConfigById 根据代理 ID 获取代理配置
func (m *Manager) GetProxyConfigById(proxyId string) (string, bool) {
	if proxy, ok := m.GetProxyByID(proxyId); ok {
		return strings.TrimSpace(proxy.ProxyConfig), true
	}
	return "", false
}

// ResolveUserDataDir 解析用户数据目录
func (m *Manager) ResolveUserDataDir(profile *Profile) string {
	userDataDir := strings.TrimSpace(profile.UserDataDir)
	if userDataDir == "" {
		userDataDir = profile.ProfileId
	}
	if filepath.IsAbs(userDataDir) {
		return userDataDir
	}
	root := strings.TrimSpace(m.Config.Browser.UserDataRoot)
	if root == "" {
		root = "data"
	}
	root = m.ResolveRelativePath(root)
	return filepath.Join(root, userDataDir)
}
