package browser

import (
	"facade/backend/internal/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Delete 删除实例及其关联数据
func (m *Manager) Delete(profileId string) error {
	log := logger.New("Browser")
	m.InitData()
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.deleteProfilePermanentlyLocked(log, profileId)
}

// CleanupExpiredTrash 清理历史回收站遗留的软删除实例
func (m *Manager) CleanupExpiredTrash() error {
	log := logger.New("Browser")
	m.InitData()
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.cleanupLeftoverTrashLocked(log)
}

func (m *Manager) deleteProfilePermanentlyLocked(log *logger.Logger, profileId string) error {
	var profile *Profile
	if existing, ok := m.Profiles[profileId]; ok {
		profile = existing
	} else if m.ProfileDAO != nil {
		found, err := m.ProfileDAO.GetById(profileId)
		if err != nil {
			log.Error("浏览器配置不存在", logger.F("profile_id", profileId))
			return err
		}
		profile = found
	} else {
		log.Error("浏览器配置不存在", logger.F("profile_id", profileId))
		return fmt.Errorf("profile not found")
	}

	if err := m.deleteProfileRelatedDataLocked(log, profile); err != nil {
		return err
	}

	if m.ProfileDAO != nil {
		if err := m.ProfileDAO.Delete(profileId); err != nil {
			return err
		}
	} else {
		delete(m.Profiles, profileId)
		if err := m.SaveProfiles(); err != nil {
			return err
		}
	}

	delete(m.Profiles, profileId)
	log.Info("浏览器配置已删除", logger.F("profile_id", profileId))
	return nil
}

func (m *Manager) cleanupLeftoverTrashLocked(log *logger.Logger) error {
	if m.ProfileDAO == nil {
		return nil
	}
	profiles, err := m.ProfileDAO.ListDeleted()
	if err != nil {
		log.Error("清理历史回收站实例失败", logger.F("error", err))
		return err
	}
	cleaned := 0
	for _, profile := range profiles {
		if err := m.deleteProfileRelatedDataLocked(log, profile); err != nil {
			log.Error("清理历史回收站实例关联数据失败", logger.F("profile_id", profile.ProfileId), logger.F("error", err))
			continue
		}
		if err := m.ProfileDAO.Delete(profile.ProfileId); err != nil {
			log.Error("删除历史回收站实例记录失败", logger.F("profile_id", profile.ProfileId), logger.F("error", err))
			continue
		}
		cleaned++
	}
	if cleaned > 0 {
		log.Info("历史回收站实例已清理", logger.F("count", cleaned))
	}
	return nil
}

func (m *Manager) deleteProfileRelatedDataLocked(log *logger.Logger, profile *Profile) error {
	if profile == nil {
		return nil
	}
	var firstErr error
	if m.CodeProvider != nil {
		if err := m.CodeProvider.Remove(profile.ProfileId); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.ExtensionDAO != nil {
		if err := m.ExtensionDAO.DeleteProfileSettings(profile.ProfileId); err != nil {
			log.Error("删除实例插件配置失败", logger.F("profile_id", profile.ProfileId), logger.F("error", err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	userDataDir := m.ResolveUserDataDir(profile)
	if err := m.deleteProfileUserDataDir(userDataDir); err != nil {
		log.Error("删除实例数据目录失败", logger.F("profile_id", profile.ProfileId), logger.F("dir", userDataDir), logger.F("error", err))
		if firstErr == nil {
			firstErr = err
		}
	}
	if err := m.deleteProfileSnapshotDir(profile.ProfileId); err != nil {
		log.Error("删除实例快照目录失败", logger.F("profile_id", profile.ProfileId), logger.F("error", err))
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *Manager) deleteProfileSnapshotDir(profileId string) error {
	profileId = strings.TrimSpace(profileId)
	if profileId == "" {
		return nil
	}
	dataRoot, err := filepath.Abs(m.ResolveRelativePath("data"))
	if err != nil {
		return fmt.Errorf("解析数据根目录失败: %w", err)
	}
	snapshotRoot := filepath.Join(dataRoot, "snapshots")
	target, err := filepath.Abs(filepath.Join(snapshotRoot, profileId))
	if err != nil {
		return fmt.Errorf("解析快照目录失败: %w", err)
	}
	dataRoot = filepath.Clean(dataRoot)
	snapshotRoot = filepath.Clean(snapshotRoot)
	target = filepath.Clean(target)
	if samePath(target, snapshotRoot) || samePath(target, dataRoot) || !isPathInside(target, snapshotRoot) {
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("删除快照目录失败: %w", err)
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
