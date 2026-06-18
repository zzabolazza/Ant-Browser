package browser

import (
	"ant-chrome/backend/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteRemovesProfileUserDataDir(t *testing.T) {
	appRoot := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Browser.UserDataRoot = "data"
	mgr := NewManager(cfg, appRoot)
	profile := &Profile{ProfileId: "profile-1", UserDataDir: "profile-1"}
	mgr.Profiles[profile.ProfileId] = profile

	profileDir := filepath.Join(appRoot, "data", "profile-1")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if err := mgr.Delete(profile.ProfileId); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Fatalf("expected profile dir removed, stat err=%v", err)
	}
}

func TestDeleteKeepsUserDataRootWhenProfileDirIsRoot(t *testing.T) {
	appRoot := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Browser.UserDataRoot = "data"
	mgr := NewManager(cfg, appRoot)
	profile := &Profile{ProfileId: "profile-root", UserDataDir: ""}
	mgr.Profiles[profile.ProfileId] = profile

	rootDir := filepath.Join(appRoot, "data")
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	profile.UserDataDir = rootDir

	if err := mgr.Delete(profile.ProfileId); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(rootDir); err != nil {
		t.Fatalf("expected data root kept, stat err=%v", err)
	}
}
