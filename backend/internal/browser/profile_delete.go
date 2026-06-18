package browser

import (
	"ant-chrome/backend/internal/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Delete 删除配置
func (m *Manager) Delete(profileId string) error {
	log := logger.New("Browser")
	m.InitData()
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	profile, exists := m.Profiles[profileId]
	if !exists {
		log.Error("浏览器配置不存在", logger.F("profile_id", profileId))
		return fmt.Errorf("profile not found")
	}
	userDataDir := m.ResolveUserDataDir(profile)
	delete(m.Profiles, profileId)
	log.Info("浏览器配置删除", logger.F("profile_id", profileId))

	if m.ProfileDAO != nil {
		if err := m.ProfileDAO.Delete(profileId); err != nil {
			log.Error("数据库删除实例失败", logger.F("profile_id", profileId), logger.F("error", err))
			return err
		}
	} else {
		if err := m.SaveProfiles(); err != nil {
			return err
		}
	}

	if m.CodeProvider != nil {
		_ = m.CodeProvider.Remove(profileId)
	}
	if err := m.deleteProfileUserDataDir(userDataDir); err != nil {
		log.Error("删除实例数据目录失败", logger.F("profile_id", profileId), logger.F("dir", userDataDir), logger.F("error", err))
		return err
	}
	return nil
}

func (m *Manager) deleteProfileUserDataDir(userDataDir string) error {
	userDataDir = strings.TrimSpace(userDataDir)
	if userDataDir == "" {
		return nil
	}
	target, err := filepath.Abs(userDataDir)
	if err != nil {
		return fmt.Errorf("解析实例数据目录失败: %w", err)
	}
	root := strings.TrimSpace(m.Config.Browser.UserDataRoot)
	if root == "" {
		root = "data"
	}
	rootAbs, err := filepath.Abs(m.ResolveRelativePath(root))
	if err != nil {
		return fmt.Errorf("解析用户数据根目录失败: %w", err)
	}
	target = filepath.Clean(target)
	rootAbs = filepath.Clean(rootAbs)
	if samePath(target, rootAbs) || !isPathInside(target, rootAbs) {
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("删除实例数据目录失败: %w", err)
	}
	return nil
}

func samePath(a string, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func isPathInside(path string, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil || rel == "." || rel == "" {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
