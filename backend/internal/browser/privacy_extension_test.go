package browser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureManagedPrivacyExtensionWritesSGVoices(t *testing.T) {
	userDataDir := t.TempDir()

	dir, err := EnsureManagedPrivacyExtension(userDataDir, "en-SG")
	if err != nil {
		t.Fatalf("EnsureManagedPrivacyExtension() error = %v", err)
	}
	if filepath.Dir(dir) != userDataDir {
		t.Fatalf("extension dir = %q, want under %q", dir, userDataDir)
	}
	manifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), `"world": "MAIN"`) || !strings.Contains(string(manifest), `"run_at": "document_start"`) {
		t.Fatalf("manifest missing MAIN document_start content script: %s", manifest)
	}
	script, err := os.ReadFile(filepath.Join(dir, "voices.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(script), `"lang":"en-SG"`) {
		t.Fatalf("voices script missing en-SG voice: %s", script)
	}
	if strings.Contains(string(script), "zh-CN") {
		t.Fatalf("voices script should not leak zh-CN for en-SG: %s", script)
	}
}
