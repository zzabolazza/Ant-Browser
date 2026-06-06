package fsutil

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
)

func TestNormalizePathInputConvertsWindowsSeparators(t *testing.T) {
	t.Parallel()

	got := NormalizePathInput(`chrome\Chrom-144\chrome.exe`)
	want := filepath.Join("chrome", "Chrom-144", "chrome.exe")
	if got != want {
		t.Fatalf("NormalizePathInput() = %q, want %q", got, want)
	}
}

func TestResolveUserDataDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	got, err := ResolveUserDataDir(func(path string) string {
		return filepath.Join(root, path)
	}, "profiles", "profile-a")
	if err != nil {
		t.Fatalf("ResolveUserDataDir() 返回错误: %v", err)
	}
	want := filepath.Join(root, "profiles", "profile-a")
	if got != want {
		t.Fatalf("ResolveUserDataDir() = %q, want %q", got, want)
	}
}

func TestResolveUserDataDirUsesDefaultRoot(t *testing.T) {
	t.Parallel()

	got, err := ResolveUserDataDir(func(path string) string { return filepath.Join("app", path) }, "", "profile-a")
	if err != nil {
		t.Fatalf("ResolveUserDataDir() 返回错误: %v", err)
	}
	want := filepath.Join("app", "data", "profile-a")
	if got != want {
		t.Fatalf("ResolveUserDataDir() = %q, want %q", got, want)
	}
}

func TestResolveExistingPathUsesResolverForRelativePath(t *testing.T) {
	t.Parallel()

	got, err := ResolveExistingPath(func(path string) string { return filepath.Join("app", path) }, "chrome/core", "不能为空")
	if err != nil {
		t.Fatalf("ResolveExistingPath() 返回错误: %v", err)
	}
	want := filepath.Join("app", "chrome/core")
	if got != want {
		t.Fatalf("ResolveExistingPath() = %q, want %q", got, want)
	}
}

func TestEnsureExecutableRepairsMissingExecBitsOnUnix(t *testing.T) {
	t.Parallel()

	if goruntime.GOOS == "windows" {
		t.Skip("Windows does not use POSIX execute bits")
	}

	path := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(path, []byte("stub"), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	if err := EnsureExecutable(path); err != nil {
		t.Fatalf("EnsureExecutable() 返回错误: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("读取测试文件状态失败: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("EnsureExecutable() 未补充执行权限: mode=%#o", info.Mode().Perm())
	}
}
