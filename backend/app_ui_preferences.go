package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const uiPreferencesRelativePath = "data/ui-preferences.json"

var allowedUIThemes = map[string]struct{}{
	"dark":  {},
	"light": {},
	"cream": {},
	"mint":  {},
	"ocean": {},
}

func (a *App) uiPreferencesPath() string {
	return a.resolveAppPath(uiPreferencesRelativePath)
}

// GetUIPreferences 读取本机 UI 偏好（主题等），文件位于 data/ 下，会随系统备份一起导出。
func (a *App) GetUIPreferences() (map[string]interface{}, error) {
	path := a.uiPreferencesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("读取 UI 偏好失败: %w", err)
	}
	var prefs map[string]interface{}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("解析 UI 偏好失败: %w", err)
	}
	if prefs == nil {
		prefs = map[string]interface{}{}
	}
	if theme, ok := prefs["theme"].(string); ok {
		theme = strings.TrimSpace(theme)
		if _, allowed := allowedUIThemes[theme]; !allowed {
			delete(prefs, "theme")
		} else {
			prefs["theme"] = theme
		}
	}
	return prefs, nil
}

// SaveUIPreferences 保存本机 UI 偏好。theme 为空时删除该字段。
func (a *App) SaveUIPreferences(prefs map[string]interface{}) error {
	if prefs == nil {
		prefs = map[string]interface{}{}
	}
	normalized := map[string]interface{}{}
	if raw, ok := prefs["theme"]; ok {
		theme := strings.TrimSpace(fmt.Sprint(raw))
		if theme != "" {
			if _, allowed := allowedUIThemes[theme]; !allowed {
				return fmt.Errorf("不支持的主题: %s", theme)
			}
			normalized["theme"] = theme
		}
	}

	path := a.uiPreferencesPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建 UI 偏好目录失败: %w", err)
	}
	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 UI 偏好失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入 UI 偏好失败: %w", err)
	}
	return nil
}
