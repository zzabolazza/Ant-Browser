package backend

import (
	"archive/zip"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func runGitForTest(t *testing.T, workdir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func buildAutomationZipBytesForTest(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	paths := make([]string, 0, len(files))
	for relativePath := range files {
		paths = append(paths, relativePath)
	}
	sort.Strings(paths)

	for _, relativePath := range paths {
		entry, err := writer.Create(relativePath)
		if err != nil {
			t.Fatalf("create zip entry failed: %v", err)
		}
		if _, err := entry.Write([]byte(files[relativePath])); err != nil {
			t.Fatalf("write zip entry failed: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer failed: %v", err)
	}
	return buf.Bytes()
}

func writeAutomationScriptLibraryPackage(t *testing.T, dir string, manifest string, entry string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create script library package dir failed: %v", err)
	}
	if strings.TrimSpace(manifest) != "" {
		if err := os.WriteFile(filepath.Join(dir, "automation.script.json"), []byte(manifest), 0o644); err != nil {
			t.Fatalf("write script library manifest failed: %v", err)
		}
	}
	if strings.TrimSpace(entry) != "" {
		if err := os.WriteFile(filepath.Join(dir, "index.cjs"), []byte(entry), 0o644); err != nil {
			t.Fatalf("write script library entry failed: %v", err)
		}
	}
}
