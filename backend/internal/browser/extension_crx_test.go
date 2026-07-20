package browser

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractCRXPublicKey_MetaMaskKeepsStoreID(t *testing.T) {
	data, err := os.ReadFile("/tmp/mm-keytest.crx")
	if err != nil {
		t.Skip("sample CRX not available:", err)
	}
	key, err := extractCRXPublicKey(data)
	if err != nil {
		t.Fatal(err)
	}
	gotID := extensionIDFromPublicKey(key)
	const wantID = "nkbihfbeogaeaoehlefnkodbefgpgknn"
	if gotID != wantID {
		t.Fatalf("extension id=%q want %q", gotID, wantID)
	}
}

func TestExtractCRXPublicKey_ReactDevToolsKeepsStoreID(t *testing.T) {
	data, err := os.ReadFile("/tmp/react-devtools.crx")
	if err != nil {
		t.Skip("sample CRX not available:", err)
	}
	key, err := extractCRXPublicKey(data)
	if err != nil {
		t.Fatal(err)
	}
	gotID := extensionIDFromPublicKey(key)
	const wantID = "fmkadmapgofadopljbjfkapdkoienihi"
	if gotID != wantID {
		t.Fatalf("extension id=%q want %q", gotID, wantID)
	}
}

func TestInjectManifestPublicKey(t *testing.T) {
	key := []byte{0x30, 0x82, 0x01, 0x0a}
	updated, err := injectManifestPublicKey([]byte(`{"name":"demo","version":"1.0.0"}`), key)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(updated, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["key"] == nil || raw["key"] == "" {
		t.Fatalf("key missing: %s", updated)
	}
	if raw["version"] != "1.0.0" {
		t.Fatalf("version changed: %v", raw["version"])
	}
}

func TestInstallExtensionPackageBytes_WritesManifestKey(t *testing.T) {
	data, err := os.ReadFile("/tmp/react-devtools.crx")
	if err != nil {
		t.Skip("sample CRX not available:", err)
	}

	root := t.TempDir()
	manager := &Manager{AppRoot: root}
	ext, err := manager.InstallExtensionPackageBytes("fmkadmapgofadopljbjfkapdkoienihi", "https://example.test", data, false)
	if err != nil {
		t.Fatal(err)
	}
	if ext.ExtensionID != "fmkadmapgofadopljbjfkapdkoienihi" {
		t.Fatalf("extension id=%q", ext.ExtensionID)
	}
	manifestPath := filepath.Join(ext.InstallDir, "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(manifestBytes, &raw); err != nil {
		t.Fatal(err)
	}
	key, _ := raw["key"].(string)
	if key == "" {
		t.Fatal("installed manifest missing key")
	}
	if extensionIDFromPublicKey(mustDecodePublicKey(t, key)) != "fmkadmapgofadopljbjfkapdkoienihi" {
		t.Fatal("installed key does not map to store id")
	}
}

func mustDecodePublicKey(t *testing.T, key string) []byte {
	t.Helper()
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}
