package backend

import (
	"context"
	"facade/backend/internal/config"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupUploadWebDAVCreatesDirectoryAndUploads(t *testing.T) {
	var uploaded []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "alice" || password != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.Method {
		case "MKCOL":
			w.WriteHeader(http.StatusCreated)
		case http.MethodPut:
			var err error
			uploaded, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	localPath := filepath.Join(root, "backup.facade")
	want := []byte("encrypted backup")
	if err := os.WriteFile(localPath, want, 0o600); err != nil {
		t.Fatal(err)
	}
	app := &App{config: &config.Config{Backup: config.BackupConfig{WebDAV: config.WebDAVConfig{
		URL: server.URL, Username: "alice", Password: "secret", RemoteDir: "Facade/Backups",
	}}}}
	remoteURL, err := app.backupUploadWebDAV(context.Background(), localPath, "test.facade")
	if err != nil {
		t.Fatal(err)
	}
	if remoteURL != server.URL+"/Facade/Backups/test.facade" {
		t.Fatalf("unexpected remote URL: %s", remoteURL)
	}
	if string(uploaded) != string(want) {
		t.Fatalf("unexpected upload: %q", uploaded)
	}
}
