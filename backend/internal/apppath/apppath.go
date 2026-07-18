package apppath

import (
	"facade/backend/internal/fsutil"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
)

const appStateDirName = "facade"

type roots struct {
	installRoot string
	stateRoot   string
	detached    bool
}

var rootsCache sync.Map

// InstallRoot 返回应用安装根目录的绝对路径。
func InstallRoot(appRoot string) string {
	return detect(appRoot).installRoot
}

// StateRoot 返回应用可写状态目录的绝对路径。
func StateRoot(appRoot string) string {
	return detect(appRoot).stateRoot
}

// IsDetached 返回当前是否将可写状态放到独立于安装/项目目录的用户状态根。
func IsDetached(appRoot string) bool {
	return detect(appRoot).detached
}

// Resolve 将相对路径解析到安装目录或用户状态目录。
// 在 detached 模式下，除 bin/ 外的相对路径都会落到用户可写状态目录。
func Resolve(appRoot, p string) string {
	return resolveForOS(appRoot, p, goruntime.GOOS)
}

func resolveForOS(appRoot, p, goos string) string {
	p = fsutil.NormalizePathInput(p)
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}

	root := detectForOS(appRoot, goos)
	base := root.installRoot
	if root.detached && useStateRoot(p) {
		base = root.stateRoot
	}
	return filepath.Join(base, p)
}

// EnsureWritableLayout 为需要 detached 状态目录的已安装应用准备首启所需的可写目录，
// 并把随包默认配置迁移到用户目录。
func EnsureWritableLayout(appRoot string) error {
	return ensureWritableLayoutForOS(appRoot, goruntime.GOOS)
}

func ensureWritableLayoutForOS(appRoot, goos string) error {
	root := detectForOS(appRoot, goos)
	if !root.detached {
		return nil
	}

	if err := os.MkdirAll(root.stateRoot, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root.stateRoot, "data"), 0755); err != nil {
		return err
	}

	if err := copyFileIfMissing(
		filepath.Join(root.installRoot, "config.yaml"),
		filepath.Join(root.stateRoot, "config.yaml"),
	); err != nil {
		return err
	}
	if err := copyFileIfMissing(
		filepath.Join(root.installRoot, "proxies.yaml"),
		filepath.Join(root.stateRoot, "proxies.yaml"),
	); err != nil {
		return err
	}

	return nil
}

func detect(appRoot string) roots {
	return detectForOS(appRoot, goruntime.GOOS)
}

func detectForOS(appRoot, goos string) roots {
	normalized := normalizeRoot(appRoot)
	cacheKey := buildCacheKey(goos, normalized)
	if cached, ok := rootsCache.Load(cacheKey); ok {
		return cached.(roots)
	}

	root := roots{
		installRoot: normalized,
		stateRoot:   normalized,
	}
	if shouldDetachStateRoot(goos, normalized) {
		root.stateRoot = userStateRootForOS(goos, normalized)
		root.detached = root.stateRoot != "" && root.stateRoot != normalized
	}

	actual, _ := rootsCache.LoadOrStore(cacheKey, root)
	return actual.(roots)
}

func buildCacheKey(goos, root string) string {
	return normalizeGOOS(goos) + "\x00" + root
}

func normalizeGOOS(goos string) string {
	return strings.ToLower(strings.TrimSpace(goos))
}

func shouldDetachStateRoot(goos, installRoot string) bool {
	_ = installRoot
	switch normalizeGOOS(goos) {
	case "linux", "darwin", "windows":
		return true
	default:
		return false
	}
}

func normalizeRoot(appRoot string) string {
	root := strings.TrimSpace(appRoot)
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		}
	}
	if root == "" {
		root = "."
	}
	if abs, err := filepath.Abs(root); err == nil {
		return abs
	}
	return filepath.Clean(root)
}

func userStateRootForOS(goos, fallback string) string {
	switch normalizeGOOS(goos) {
	case "linux":
		if base := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); base != "" {
			return filepath.Join(base, appStateDirName)
		}
		if home := configuredHomeDir(); home != "" {
			return filepath.Join(home, ".local", "share", appStateDirName)
		}
	case "darwin":
		if home := configuredHomeDir(); home != "" {
			return filepath.Join(home, "Library", "Application Support", appStateDirName)
		}
	case "windows":
		if base := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); base != "" {
			return filepath.Join(base, appStateDirName)
		}
		if home := configuredHomeDir(); home != "" {
			return filepath.Join(home, "AppData", "Local", appStateDirName)
		}
	}
	if tmp := strings.TrimSpace(os.TempDir()); tmp != "" {
		return filepath.Join(tmp, appStateDirName)
	}
	return fallback
}

func configuredHomeDir() string {
	if home := strings.TrimSpace(os.Getenv("HOME")); home != "" {
		return home
	}
	if home, err := os.UserHomeDir(); err == nil {
		return strings.TrimSpace(home)
	}
	return ""
}

func useStateRoot(p string) bool {
	clean := filepath.ToSlash(fsutil.NormalizePathInput(p))
	if clean == "" || clean == "." {
		return false
	}
	return clean != "bin" && !strings.HasPrefix(clean, "bin/")
}

func copyFileIfMissing(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
