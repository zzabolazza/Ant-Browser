package browser

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	extensionDownloadTimeout = 90 * time.Second
	extensionMaxPackageBytes = 128 << 20
	extensionsRootDir        = "extensions"
)

func ExtensionDownloadTimeout() time.Duration {
	return extensionDownloadTimeout
}

var extensionIDPattern = regexp.MustCompile(`^[a-p]{32}$`)

type extensionManifest struct {
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Description   string            `json:"description"`
	ShortName     string            `json:"short_name"`
	DefaultLocale string            `json:"default_locale"`
	Icons         map[string]string `json:"icons"`
	Action        map[string]any    `json:"action"`
	BrowserAction map[string]any    `json:"browser_action"`
}

func NormalizeExtensionID(value string) string {
	trimmed := strings.TrimSpace(value)
	if parsed := extractExtensionIDFromURL(trimmed); parsed != "" {
		trimmed = parsed
	}
	trimmed = strings.ToLower(strings.Trim(trimmed, "/#?& "))
	if extensionIDPattern.MatchString(trimmed) {
		return trimmed
	}
	return ""
}

func BuildChromeWebStoreURL(extensionID string) string {
	normalizedID := NormalizeExtensionID(extensionID)
	if normalizedID == "" {
		return ""
	}
	return "https://chromewebstore.google.com/detail/" + normalizedID
}

func BuildChromeExtensionDownloadURL(extensionID, prodVersion string) string {
	normalizedID := NormalizeExtensionID(extensionID)
	prodVersion = NormalizeChromeProdVersion(prodVersion)
	if normalizedID == "" || prodVersion == "" {
		return ""
	}
	return "https://clients2.google.com/service/update2/crx?response=redirect&prodversion=" +
		prodVersion +
		"&acceptformat=crx2,crx3&x=id%3D" + normalizedID + "%26installsource%3Dondemand%26uc"
}

// NormalizeChromeProdVersion 从内核版本字符串中提取可供商店下载使用的 Chrome 版本号。
func NormalizeChromeProdVersion(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return ""
	}
	match := chromeProdVersionPattern.FindString(trimmed)
	return match
}

var chromeProdVersionPattern = regexp.MustCompile(`\d+(?:\.\d+){1,3}`)

// ResolveExtensionDownloadProdVersion 使用已配置内核路径（优先默认内核）解析出的
// Chromium/Chrome 版本，作为商店下载时的 prodversion。
func (m *Manager) ResolveExtensionDownloadProdVersion() (string, error) {
	if m == nil {
		return "", fmt.Errorf("浏览器管理器未初始化")
	}
	if core, ok := m.GetDefaultCore(); ok {
		if version := NormalizeChromeProdVersion(m.GetChromeVersion(core.CorePath)); version != "" {
			return version, nil
		}
	}
	for _, core := range m.ListCores() {
		if version := NormalizeChromeProdVersion(m.GetChromeVersion(core.CorePath)); version != "" {
			return version, nil
		}
	}
	return "", fmt.Errorf("无法从已配置的 Chromium/Chrome 内核路径读取版本，请确认内核管理中的路径有效")
}

func (m *Manager) LookupExtension(query string) (ExtensionLookupResult, error) {
	return m.LookupExtensionWithHTTPClient(query, nil)
}

func (m *Manager) LookupExtensionWithHTTPClient(query string, client *http.Client) (ExtensionLookupResult, error) {
	extensionID := NormalizeExtensionID(query)
	if extensionID == "" {
		return ExtensionLookupResult{}, fmt.Errorf("请输入 Chrome 插件 ID 或 Chrome Web Store 链接")
	}
	result := ExtensionLookupResult{
		ExtensionID: extensionID,
		Name:        extensionID,
		StoreURL:    BuildChromeWebStoreURL(extensionID),
		Installable: true,
		Message:     "已识别插件 ID，可下载安装",
	}
	prodVersion, err := m.ResolveExtensionDownloadProdVersion()
	if err != nil {
		result.Message = "已识别插件 ID，但暂时无法读取商店元信息: " + err.Error()
		return result, nil
	}
	data, err := downloadChromeExtensionCRX(context.Background(), extensionID, prodVersion, client)
	if err != nil {
		result.Message = "已识别插件 ID，但暂时无法读取商店元信息: " + err.Error()
		return result, nil
	}
	zipData, err := normalizeExtensionArchiveData(data)
	if err != nil {
		result.Message = "已识别插件 ID，但插件包格式无法解析: " + err.Error()
		return result, nil
	}
	manifestData, err := readExtensionManifestFromZip(zipData)
	if err != nil {
		result.Message = "已识别插件 ID，但 manifest 无法解析: " + err.Error()
		return result, nil
	}
	manifest, err := parseExtensionManifest(manifestData)
	if err != nil {
		result.Message = "已识别插件 ID，但 manifest 无法解析: " + err.Error()
		return result, nil
	}
	localeMessages := readExtensionLocaleMessagesFromZip(zipData, manifest)
	result.Name = resolveExtensionName(manifest, extensionID)
	result.Name = resolveExtensionMessage(result.Name, localeMessages)
	result.Version = strings.TrimSpace(manifest.Version)
	result.Description = resolveExtensionDescription(manifest, localeMessages)
	result.Message = "已读取插件信息，可下载安装"
	return ExtensionLookupResult{
		ExtensionID: result.ExtensionID,
		Name:        result.Name,
		Version:     result.Version,
		Description: result.Description,
		StoreURL:    result.StoreURL,
		Installable: result.Installable,
		Message:     result.Message,
	}, nil
}

func (m *Manager) InstallExtensionFromWebStore(ctx context.Context, query string) (Extension, error) {
	return m.InstallExtensionFromWebStoreWithHTTPClient(ctx, query, nil, false)
}

func (m *Manager) InstallExtensionFromWebStoreWithHTTPClient(ctx context.Context, query string, client *http.Client, allowOverwrite bool) (Extension, error) {
	extensionID := NormalizeExtensionID(query)
	if extensionID == "" {
		return Extension{}, fmt.Errorf("请输入 Chrome 插件 ID 或 Chrome Web Store 链接")
	}
	prodVersion, err := m.ResolveExtensionDownloadProdVersion()
	if err != nil {
		return Extension{}, err
	}
	data, err := downloadChromeExtensionCRX(ctx, extensionID, prodVersion, client)
	if err != nil {
		return Extension{}, err
	}
	return m.InstallExtensionPackageBytes(extensionID, BuildChromeWebStoreURL(extensionID), data, allowOverwrite)
}

func (m *Manager) InstallExtensionPackageBytes(extensionID string, sourceURL string, data []byte, allowOverwrite bool) (Extension, error) {
	if len(data) == 0 {
		return Extension{}, fmt.Errorf("插件包为空")
	}
	if len(data) > extensionMaxPackageBytes {
		return Extension{}, fmt.Errorf("插件包超过限制")
	}
	publicKey, err := extractCRXPublicKey(data)
	if err != nil {
		return Extension{}, err
	}
	keyDerivedID := extensionIDFromPublicKey(publicKey)
	if keyDerivedID == "" {
		return Extension{}, fmt.Errorf("无法从 CRX 公钥计算扩展 ID")
	}

	zipData, err := normalizeExtensionArchiveData(data)
	if err != nil {
		return Extension{}, err
	}
	manifestData, err := readExtensionManifestFromZip(zipData)
	if err != nil {
		return Extension{}, err
	}
	manifest, err := parseExtensionManifest(manifestData)
	if err != nil {
		return Extension{}, err
	}

	resolvedID := keyDerivedID
	if requestedID := NormalizeExtensionID(extensionID); requestedID != "" && requestedID != resolvedID {
		return Extension{}, fmt.Errorf("插件包公钥对应的扩展 ID 为 %s，与请求的 %s 不一致", resolvedID, requestedID)
	}

	installed, err := m.extensionInstalled(resolvedID)
	if err != nil {
		return Extension{}, err
	}
	if installed && !allowOverwrite {
		return Extension{}, fmt.Errorf("插件已安装")
	}

	installDir := filepath.Join(m.ResolveRelativePath(filepath.Join("data", extensionsRootDir)), resolvedID)
	if err := replaceExtensionDirFromZip(zipData, installDir); err != nil {
		return Extension{}, err
	}
	manifestJSONBytes, err := writeManifestPublicKey(installDir, publicKey)
	if err != nil {
		_ = os.RemoveAll(installDir)
		return Extension{}, err
	}

	localeMessages := readExtensionLocaleMessagesFromZip(zipData, manifest)
	extension := Extension{
		ExtensionID:  resolvedID,
		Name:         resolveExtensionMessage(resolveExtensionName(manifest, resolvedID), localeMessages),
		Version:      strings.TrimSpace(manifest.Version),
		Description:  resolveExtensionDescription(manifest, localeMessages),
		IconDataURL:  readExtensionIconDataURLFromZip(zipData, manifest),
		ManifestJSON: string(manifestJSONBytes),
		SourceURL:    strings.TrimSpace(sourceURL),
		InstallDir:   installDir,
		Enabled:      true,
	}
	if m.ExtensionDAO != nil {
		if err := m.ExtensionDAO.Upsert(extension); err != nil {
			return Extension{}, err
		}
		stored, err := m.ExtensionDAO.Get(resolvedID)
		if err == nil {
			return stored, nil
		}
	}
	return extension, nil
}

func (m *Manager) InstallExtensionPackageFile(path string) (Extension, error) {
	normalizedPath := strings.TrimSpace(path)
	if normalizedPath == "" {
		return Extension{}, fmt.Errorf("插件文件路径不能为空")
	}
	data, err := os.ReadFile(normalizedPath)
	if err != nil {
		return Extension{}, fmt.Errorf("读取插件文件失败: %w", err)
	}
	return m.InstallExtensionPackageBytes("", normalizedPath, data, false)
}

func (m *Manager) extensionInstalled(extensionID string) (bool, error) {
	if m == nil || m.ExtensionDAO == nil {
		return false, nil
	}
	extensionID = strings.TrimSpace(extensionID)
	if extensionID == "" {
		return false, nil
	}
	_, err := m.ExtensionDAO.Get(extensionID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (m *Manager) EnabledExtensionDirs() []string {
	if m == nil || m.ExtensionDAO == nil {
		return nil
	}
	items, err := m.ExtensionDAO.ListEnabled()
	if err != nil {
		return nil
	}
	dirs := make([]string, 0, len(items))
	for _, item := range items {
		dir := strings.TrimSpace(item.InstallDir)
		if dir == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err == nil {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func (m *Manager) EnabledExtensionDirsForProfile(profileID string) []string {
	if m == nil || m.ExtensionDAO == nil {
		return nil
	}
	settings, err := m.ExtensionDAO.GetProfileSettings(profileID)
	if err != nil || !settings.Configured {
		return m.EnabledExtensionDirs()
	}
	items, err := m.ExtensionDAO.ListByIDs(settings.ExtensionIDs)
	if err != nil {
		return nil
	}
	dirs := make([]string, 0, len(items))
	for _, item := range items {
		dir := strings.TrimSpace(item.InstallDir)
		if dir != "" {
			if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err == nil {
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}
