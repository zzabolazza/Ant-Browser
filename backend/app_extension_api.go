package backend

import (
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
	"ant-chrome/backend/internal/logger"
	"ant-chrome/backend/internal/proxy"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type BrowserExtension = browser.Extension
type BrowserExtensionLookupResult = browser.ExtensionLookupResult
type BrowserProfileExtensionSettings = browser.ProfileExtensionSettings

type BrowserExtensionWebStoreRequest struct {
	Query          string `json:"query"`
	UseProxy       bool   `json:"useProxy"`
	ProxyConfig    string `json:"proxyConfig"`
	AllowOverwrite bool   `json:"allowOverwrite"`
}

func (a *App) BrowserExtensionList() ([]BrowserExtension, error) {
	if a.browserMgr == nil || a.browserMgr.ExtensionDAO == nil {
		return []BrowserExtension{}, nil
	}
	return a.browserMgr.ExtensionDAO.List()
}

func (a *App) BrowserExtensionLookup(query string) (BrowserExtensionLookupResult, error) {
	return a.BrowserExtensionLookupWithProxy(BrowserExtensionWebStoreRequest{Query: query})
}

func (a *App) BrowserExtensionLookupWithProxy(input BrowserExtensionWebStoreRequest) (BrowserExtensionLookupResult, error) {
	if a.browserMgr == nil {
		return BrowserExtensionLookupResult{}, fmt.Errorf("浏览器管理器未初始化")
	}
	client, err := a.extensionDownloadHTTPClient(input.UseProxy, input.ProxyConfig)
	if err != nil {
		return BrowserExtensionLookupResult{}, fmt.Errorf("下载代理配置错误: %w", err)
	}
	return a.browserMgr.LookupExtensionWithHTTPClient(input.Query, client)
}

func (a *App) BrowserExtensionInstall(query string) (BrowserExtension, error) {
	return a.BrowserExtensionInstallWithProxy(BrowserExtensionWebStoreRequest{Query: query})
}

func (a *App) BrowserExtensionInstallWithProxy(input BrowserExtensionWebStoreRequest) (BrowserExtension, error) {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()

	if a.ctx == nil {
		return BrowserExtension{}, fmt.Errorf("应用上下文未初始化")
	}
	if a.browserMgr == nil {
		return BrowserExtension{}, fmt.Errorf("浏览器管理器未初始化")
	}
	client, err := a.extensionDownloadHTTPClient(input.UseProxy, input.ProxyConfig)
	if err != nil {
		return BrowserExtension{}, fmt.Errorf("下载代理配置错误: %w", err)
	}
	return a.browserMgr.InstallExtensionFromWebStoreWithHTTPClient(a.ctx, input.Query, client, input.AllowOverwrite)
}

func (a *App) extensionDownloadHTTPClient(useProxy bool, proxyConfig string) (*http.Client, error) {
	proxyConfig = strings.TrimSpace(proxyConfig)
	log := logger.New("Extension")
	if useProxy && proxyConfig == "" {
		return nil, fmt.Errorf("已启用下载代理，但代理配置为空，请重新选择代理节点")
	}
	if proxyConfig == "" || strings.EqualFold(proxyConfig, "direct://") {
		log.Info("Chrome 插件下载使用直连")
		client, _, err := proxyCoreHTTPClient(browser.ExtensionDownloadTimeout(), "")
		return client, err
	}
	proxies := a.getLatestProxies()
	connectorType := config.BrowserConnectorXray
	if a != nil && a.config != nil {
		connectorType = a.config.Browser.DefaultConnectorType
	}
	log.Info("Chrome 插件下载使用代理", logger.F("connector", connectorType), logger.F("proxy_prefix", proxyConfigLogPrefix(proxyConfig)))
	return proxy.BuildProxyHTTPClient(proxyConfig, "", proxies, a.xrayMgr, a.singboxMgr, a.clashMgr, connectorType, browser.ExtensionDownloadTimeout())
}

func proxyConfigLogPrefix(proxyConfig string) string {
	proxyConfig = strings.TrimSpace(proxyConfig)
	if len(proxyConfig) <= 24 {
		return proxyConfig
	}
	return proxyConfig[:24]
}

func (a *App) BrowserExtensionInstallLocalFile() (BrowserExtension, error) {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()

	if a.ctx == nil {
		return BrowserExtension{}, fmt.Errorf("应用上下文未初始化")
	}
	if a.browserMgr == nil {
		return BrowserExtension{}, fmt.Errorf("浏览器管理器未初始化")
	}
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择 Chrome 插件包",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Chrome 插件包 (*.crx;*.zip)", Pattern: "*.crx;*.zip"},
		},
	})
	if err != nil {
		return BrowserExtension{}, fmt.Errorf("打开文件选择框失败: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return BrowserExtension{}, fmt.Errorf("已取消选择")
	}
	return a.browserMgr.InstallExtensionPackageFile(path)
}

func (a *App) BrowserExtensionSetEnabled(extensionID string, enabled bool) (BrowserExtension, error) {
	if a.browserMgr == nil || a.browserMgr.ExtensionDAO == nil {
		return BrowserExtension{}, fmt.Errorf("插件管理器未初始化")
	}
	extensionID = strings.TrimSpace(extensionID)
	if extensionID == "" {
		return BrowserExtension{}, fmt.Errorf("插件 ID 不能为空")
	}
	if err := a.browserMgr.ExtensionDAO.SetEnabled(extensionID, enabled); err != nil {
		return BrowserExtension{}, err
	}
	return a.browserMgr.ExtensionDAO.Get(extensionID)
}

func (a *App) BrowserExtensionDelete(extensionID string) error {
	a.maintenanceMu.Lock()
	defer a.maintenanceMu.Unlock()

	if a.browserMgr == nil || a.browserMgr.ExtensionDAO == nil {
		return fmt.Errorf("插件管理器未初始化")
	}
	extensionID = strings.TrimSpace(extensionID)
	if extensionID == "" {
		return fmt.Errorf("插件 ID 不能为空")
	}
	extension, err := a.browserMgr.ExtensionDAO.Get(extensionID)
	if err != nil {
		return err
	}
	target, err := a.resolveBrowserExtensionInstallDir(extension.InstallDir)
	if err != nil {
		return err
	}
	if target != "" {
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("删除插件目录失败: %w", err)
		}
	}
	if err := a.browserMgr.ExtensionDAO.Delete(extensionID); err != nil {
		return err
	}
	return nil
}

func canonicalPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return filepath.Clean(abs), nil
}

func (a *App) resolveBrowserExtensionInstallDir(installDir string) (string, error) {
	installDir = strings.TrimSpace(installDir)
	if installDir == "" {
		return "", nil
	}
	target, err := canonicalPath(installDir)
	if err != nil {
		return "", fmt.Errorf("解析插件目录失败: %w", err)
	}
	root, err := canonicalPath(a.resolveAppPath(filepath.ToSlash(filepath.Join("data", "extensions"))))
	if err != nil {
		return "", fmt.Errorf("解析插件根目录失败: %w", err)
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || rel == "" || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("拒绝删除插件根目录外的路径: %s", installDir)
	}
	return target, nil
}

func (a *App) BrowserProfileExtensionGet(profileID string) (BrowserProfileExtensionSettings, error) {
	if a.browserMgr == nil || a.browserMgr.ExtensionDAO == nil {
		return BrowserProfileExtensionSettings{}, fmt.Errorf("插件管理器未初始化")
	}
	return a.browserMgr.ExtensionDAO.GetProfileSettings(profileID)
}

func (a *App) BrowserProfileExtensionSave(profileID string, extensionIDs []string, configured bool) (BrowserProfileExtensionSettings, error) {
	if a.browserMgr == nil || a.browserMgr.ExtensionDAO == nil {
		return BrowserProfileExtensionSettings{}, fmt.Errorf("插件管理器未初始化")
	}
	return a.browserMgr.ExtensionDAO.SetProfileSettings(profileID, extensionIDs, configured)
}
