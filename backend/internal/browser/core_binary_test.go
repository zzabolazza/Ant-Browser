package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindCoreExecutable_AppBundle(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("app bundle lookup is darwin-only")
	}

	root := t.TempDir()
	bundle := filepath.Join(root, "Chromium.app")
	exeDir := filepath.Join(bundle, "Contents", "MacOS")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(exeDir, "Chromium")
	if err := os.WriteFile(exePath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, candidate, ok := FindCoreExecutable(bundle)
	if !ok {
		t.Fatalf("expected to find executable inside %s", bundle)
	}
	if got != exePath {
		t.Fatalf("got path %q, want %q", got, exePath)
	}
	if candidate != "Chromium.app/Contents/MacOS/Chromium" {
		t.Fatalf("got candidate %q", candidate)
	}

	got, _, ok = FindCoreExecutable(root)
	if !ok || got != exePath {
		t.Fatalf("expected parent dir lookup to find %q, got %q ok=%v", exePath, got, ok)
	}
}
