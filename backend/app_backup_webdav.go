package backend

import (
	"context"
	"facade/backend/internal/config"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type BackupWebDAVSettings struct {
	URL         string `json:"url"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"`
	RemoteDir   string `json:"remoteDir"`
	HasPassword bool   `json:"hasPassword"`
}

func (a *App) BackupGetWebDAVSettings() BackupWebDAVSettings {
	if a.config == nil {
		return BackupWebDAVSettings{}
	}
	item := a.config.Backup.WebDAV
	return BackupWebDAVSettings{
		URL:         item.URL,
		Username:    item.Username,
		RemoteDir:   item.RemoteDir,
		HasPassword: strings.TrimSpace(item.Password) != "",
	}
}

func (a *App) BackupSaveWebDAVSettings(settings BackupWebDAVSettings) error {
	if a.config == nil {
		return fmt.Errorf("配置未初始化")
	}
	normalized, err := normalizeWebDAVSettings(settings, a.config.Backup.WebDAV.Password)
	if err != nil {
		return err
	}
	a.config.Backup.WebDAV = normalized
	if err := a.config.Save(a.resolveAppPath("config.yaml")); err != nil {
		return fmt.Errorf("保存 WebDAV 设置失败: %w", err)
	}
	return nil
}

func normalizeWebDAVSettings(settings BackupWebDAVSettings, existingPassword string) (config.WebDAVConfig, error) {
	rawURL := strings.TrimSpace(settings.URL)
	if rawURL == "" {
		return config.WebDAVConfig{}, fmt.Errorf("WebDAV 地址不能为空")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return config.WebDAVConfig{}, fmt.Errorf("WebDAV 地址必须是有效的 HTTP 或 HTTPS 地址")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	password := settings.Password
	if password == "" {
		password = existingPassword
	}
	remoteDir := strings.Trim(strings.TrimSpace(settings.RemoteDir), "/\\")
	if remoteDir != "" {
		normalizedDir := strings.ReplaceAll(remoteDir, "\\", "/")
		for _, segment := range strings.Split(normalizedDir, "/") {
			if segment == ".." {
				return config.WebDAVConfig{}, fmt.Errorf("WebDAV 远程目录不合法")
			}
		}
		cleaned := path.Clean("/" + normalizedDir)
		if cleaned == "/" {
			return config.WebDAVConfig{}, fmt.Errorf("WebDAV 远程目录不合法")
		}
		remoteDir = strings.TrimPrefix(cleaned, "/")
	}
	return config.WebDAVConfig{
		URL:       strings.TrimRight(parsed.String(), "/"),
		Username:  strings.TrimSpace(settings.Username),
		Password:  password,
		RemoteDir: remoteDir,
	}, nil
}

func (a *App) backupUploadWebDAV(ctx context.Context, localPath, fileName string) (string, error) {
	if a.config == nil {
		return "", fmt.Errorf("配置未初始化")
	}
	settings := a.config.Backup.WebDAV
	if strings.TrimSpace(settings.URL) == "" {
		return "", fmt.Errorf("请先配置 WebDAV")
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	baseURL := strings.TrimRight(settings.URL, "/")
	currentURL := baseURL
	if settings.RemoteDir != "" {
		for _, segment := range strings.Split(strings.ReplaceAll(settings.RemoteDir, "\\", "/"), "/") {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}
			currentURL += "/" + url.PathEscape(segment)
			request, err := http.NewRequestWithContext(ctx, "MKCOL", currentURL, nil)
			if err != nil {
				return "", err
			}
			setWebDAVAuth(request, settings)
			response, err := client.Do(request)
			if err != nil {
				return "", fmt.Errorf("创建 WebDAV 目录失败: %w", err)
			}
			_ = response.Body.Close()
			if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusMethodNotAllowed {
				return "", fmt.Errorf("创建 WebDAV 目录失败: HTTP %s", response.Status)
			}
		}
	}
	remoteURL := currentURL + "/" + url.PathEscape(fileName)
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, remoteURL, file)
	if err != nil {
		return "", err
	}
	request.ContentLength = info.Size()
	request.Header.Set("Content-Type", "application/octet-stream")
	setWebDAVAuth(request, settings)
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("上传 WebDAV 失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("上传 WebDAV 失败: HTTP %s", response.Status)
	}
	return remoteURL, nil
}

func setWebDAVAuth(request *http.Request, settings config.WebDAVConfig) {
	if settings.Username != "" || settings.Password != "" {
		request.SetBasicAuth(settings.Username, settings.Password)
	}
}
