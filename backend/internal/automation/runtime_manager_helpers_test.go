package automation

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func buildTestNodeZip(t *testing.T) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	header := &zip.FileHeader{
		Name:   "node-v22.15.1-win-x64/node.exe",
		Method: zip.Deflate,
	}
	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatalf("create node zip header failed: %v", err)
	}
	if _, err := fileWriter.Write([]byte("fake-node-runtime")); err != nil {
		t.Fatalf("write node zip failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close node zip failed: %v", err)
	}

	hash := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(hash[:])
}

func buildTestPlaywrightTGZ(t *testing.T) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	payload := []byte(`{"name":"playwright-core","version":"1.59.0"}`)
	header := &tar.Header{
		Name: "package/package.json",
		Mode: 0o644,
		Size: int64(len(payload)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("write playwright header failed: %v", err)
	}
	if _, err := tarWriter.Write(payload); err != nil {
		t.Fatalf("write playwright payload failed: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close playwright tar failed: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("close playwright gzip failed: %v", err)
	}

	hash := sha1.Sum(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(hash[:])
}

func buildTestPlayablePlaywrightTGZ(t *testing.T, version string) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	files := map[string][]byte{
		"package/package.json": []byte(`{"name":"playwright-core","version":"` + version + `","main":"index.js"}`),
		"package/index.js":     []byte("exports.chromium = {};"),
	}

	for name, payload := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(payload)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write playable playwright header failed: %v", err)
		}
		if _, err := tarWriter.Write(payload); err != nil {
			t.Fatalf("write playable playwright payload failed: %v", err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close playable playwright tar failed: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("close playable playwright gzip failed: %v", err)
	}

	hash := sha1.Sum(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(hash[:])
}

func writeBrokenPlaywrightModule(runtimeDir, version string) error {
	moduleDir := filepath.Join(runtimeDir, "node_modules", "playwright-core")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		return err
	}

	packageJSON := []byte(`{"name":"playwright-core","version":"` + version + `","main":"index.js"}`)
	if err := os.WriteFile(filepath.Join(moduleDir, "package.json"), packageJSON, 0o644); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(moduleDir, "index.js"), []byte("module.exports = {};"), 0o644)
}
