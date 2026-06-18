package backend

import (
	"ant-chrome/backend/internal/config"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteArchiveFileRejectsParentDirectoryEntry(t *testing.T) {
	targetDir := t.TempDir()
	err := writeArchiveFile(targetDir, "..", 0o644, false, func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("bad")), nil
	})
	if err == nil {
		t.Fatalf("expected parent directory archive entry to be rejected")
	}
}

func TestSelectProxyCoreAssetPrefersMihomoCompatibleWindows(t *testing.T) {
	spec, err := normalizeProxyCoreSpec("mihomo")
	if err != nil {
		t.Fatalf("normalizeProxyCoreSpec returned error: %v", err)
	}
	asset, err := selectProxyCoreAsset(spec, []githubReleaseAsset{
		{Name: "mihomo-windows-amd64-v1-v1.19.27.zip"},
		{Name: "mihomo-windows-amd64-compatible-v1.19.27.zip"},
	}, "windows", "amd64")
	if err != nil {
		t.Fatalf("selectProxyCoreAsset returned error: %v", err)
	}
	if asset.Name != "mihomo-windows-amd64-compatible-v1.19.27.zip" {
		t.Fatalf("unexpected asset: %s", asset.Name)
	}
}

func TestSelectProxyCoreAssetSupportsMihomoLinuxGzip(t *testing.T) {
	spec, err := normalizeProxyCoreSpec("mihomo")
	if err != nil {
		t.Fatalf("normalizeProxyCoreSpec returned error: %v", err)
	}
	asset, err := selectProxyCoreAsset(spec, []githubReleaseAsset{
		{Name: "mihomo-linux-amd64-compatible-v1.19.27.gz"},
		{Name: "mihomo-linux-amd64-v1.19.27.gz"},
	}, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectProxyCoreAsset returned error: %v", err)
	}
	if asset.Name != "mihomo-linux-amd64-compatible-v1.19.27.gz" {
		t.Fatalf("unexpected asset: %s", asset.Name)
	}
}

func TestSelectProxyCoreAssetMatchesXrayLinux64(t *testing.T) {
	spec, err := normalizeProxyCoreSpec("xray")
	if err != nil {
		t.Fatalf("normalizeProxyCoreSpec returned error: %v", err)
	}
	asset, err := selectProxyCoreAsset(spec, []githubReleaseAsset{
		{Name: "Xray-linux-arm64-v8a.zip"},
		{Name: "Xray-linux-64.zip"},
		{Name: "Xray-linux-64.zip.dgst"},
	}, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectProxyCoreAsset returned error: %v", err)
	}
	if asset.Name != "Xray-linux-64.zip" {
		t.Fatalf("unexpected asset: %s", asset.Name)
	}
}

func TestNormalizeProxyCoreTargetAliases(t *testing.T) {
	target, err := normalizeProxyCoreTarget("macos", "x64")
	if err != nil {
		t.Fatalf("normalizeProxyCoreTarget returned error: %v", err)
	}
	if target.GOOS != "darwin" || target.GOARCH != "amd64" {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestNormalizeProxyCoreSpecPinsStableVersions(t *testing.T) {
	cases := map[string]string{
		"xray":     "v26.3.27",
		"mihomo":   "v1.19.27",
		"sing-box": "v1.13.13",
	}
	for core, want := range cases {
		spec, err := normalizeProxyCoreSpec(core)
		if err != nil {
			t.Fatalf("normalizeProxyCoreSpec(%q) returned error: %v", core, err)
		}
		if spec.Version != want {
			t.Fatalf("%s version = %q, want %q", core, spec.Version, want)
		}
	}
}

func TestProxyCoreReleaseURLUsesPinnedTag(t *testing.T) {
	got := proxyCoreReleaseURL("MetaCubeX/mihomo", "v1.19.27")
	want := "https://github.com/MetaCubeX/mihomo/releases/tag/v1.19.27"
	if got != want {
		t.Fatalf("release URL = %q, want %q", got, want)
	}
}

func TestSelectProxyCoreAssetMatchesDarwinAMD64(t *testing.T) {
	spec, err := normalizeProxyCoreSpec("sing-box")
	if err != nil {
		t.Fatalf("normalizeProxyCoreSpec returned error: %v", err)
	}
	asset, err := selectProxyCoreAsset(spec, []githubReleaseAsset{
		{Name: "sing-box-1.13.13-darwin-arm64.tar.gz"},
		{Name: "sing-box-1.13.13-darwin-amd64.tar.gz"},
	}, "darwin", "amd64")
	if err != nil {
		t.Fatalf("selectProxyCoreAsset returned error: %v", err)
	}
	if asset.Name != "sing-box-1.13.13-darwin-amd64.tar.gz" {
		t.Fatalf("unexpected asset: %s", asset.Name)
	}
}

func TestSelectProxyCoreAssetPrefersGenericSingBoxLinux(t *testing.T) {
	spec, err := normalizeProxyCoreSpec("sing-box")
	if err != nil {
		t.Fatalf("normalizeProxyCoreSpec returned error: %v", err)
	}
	asset, err := selectProxyCoreAsset(spec, []githubReleaseAsset{
		{Name: "sing-box-1.13.13-linux-amd64-glibc.tar.gz"},
		{Name: "sing-box-1.13.13-linux-amd64-musl.tar.gz"},
		{Name: "sing-box-1.13.13-linux-amd64.tar.gz"},
	}, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectProxyCoreAsset returned error: %v", err)
	}
	if asset.Name != "sing-box-1.13.13-linux-amd64.tar.gz" {
		t.Fatalf("unexpected asset: %s", asset.Name)
	}
}

func TestProxyCoreStatusFindsRuntimePlatformBin(t *testing.T) {
	root := t.TempDir()
	spec, err := normalizeProxyCoreSpec("xray")
	if err != nil {
		t.Fatalf("normalizeProxyCoreSpec returned error: %v", err)
	}
	target := proxyCoreTarget{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	binaryName := "xray"
	if runtime.GOOS == "windows" {
		binaryName = "xray.exe"
	}
	binaryPath := filepath.Join(root, "bin", runtime.GOOS+"-"+runtime.GOARCH, binaryName)
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("test"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	app := &App{
		appRoot: root,
		config:  &config.Config{},
	}
	app.config.Browser.DefaultConnectorType = "xray"

	status := app.proxyCoreStatus(spec, target)
	if !status.Installed || !status.Active {
		t.Fatalf("status = %+v, want installed active", status)
	}
	if status.Message != "已启用" {
		t.Fatalf("message = %q, want 已启用", status.Message)
	}
	if status.BinaryPath == "" {
		t.Fatalf("expected binary path, got empty")
	}
}

func TestFindProxyCoreBinaryMatchesVersionedMihomoWindowsExe(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "mihomo-windows-amd64-compatible.exe")
	if err := os.WriteFile(binaryPath, []byte("test"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := findProxyCoreBinary(dir, "mihomo", "windows")
	if err != nil {
		t.Fatalf("findProxyCoreBinary returned error: %v", err)
	}
	if got != binaryPath {
		t.Fatalf("binary path = %q, want %q", got, binaryPath)
	}
}

func TestNormalizeInstalledProxyCoreBinaryRenamesMihomoExe(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "mihomo-windows-amd64-compatible.exe")
	if err := os.WriteFile(binaryPath, []byte("test"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := normalizeInstalledProxyCoreBinary(binaryPath, dir, "mihomo", "windows")
	if err != nil {
		t.Fatalf("normalizeInstalledProxyCoreBinary returned error: %v", err)
	}
	want := filepath.Join(dir, "mihomo.exe")
	if got != want {
		t.Fatalf("normalized path = %q, want %q", got, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected normalized binary to exist: %v", err)
	}
}
