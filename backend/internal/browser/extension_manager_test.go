package browser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ant-chrome/backend/internal/database"
)

func newExtensionTestDAO(t *testing.T) *SQLiteExtensionDAO {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "extensions.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewSQLiteExtensionDAO(db.GetConn())
}

func TestNormalizeExtensionID(t *testing.T) {
	validID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "raw id", in: validID, want: validID},
		{name: "upper id", in: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", want: validID},
		{name: "store url", in: "https://chromewebstore.google.com/detail/example/" + validID, want: validID},
		{name: "invalid", in: "not-an-extension", want: ""},
		{name: "wrong alphabet", in: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", want: ""},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeExtensionID(tt.in); got != tt.want {
				t.Fatalf("NormalizeExtensionID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProfileExtensionSettingsRoundTrip(t *testing.T) {
	dao := newExtensionTestDAO(t)
	ids := []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}

	settings, err := dao.SetProfileSettings("profile-1", ids, true)
	if err != nil {
		t.Fatalf("SetProfileSettings failed: %v", err)
	}
	wantIDs := []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	if !settings.Configured || !reflect.DeepEqual(settings.ExtensionIDs, wantIDs) {
		t.Fatalf("settings = %+v, want configured with %v", settings, wantIDs)
	}

	loaded, err := dao.GetProfileSettings("profile-1")
	if err != nil {
		t.Fatalf("GetProfileSettings failed: %v", err)
	}
	if !loaded.Configured || !reflect.DeepEqual(loaded.ExtensionIDs, wantIDs) {
		t.Fatalf("loaded = %+v, want configured with %v", loaded, wantIDs)
	}
}

func TestEnabledExtensionDirsForProfile(t *testing.T) {
	dao := newExtensionTestDAO(t)
	root := t.TempDir()
	firstDir := writeExtensionManifest(t, root, "one")
	secondDir := writeExtensionManifest(t, root, "two")
	disabledDir := writeExtensionManifest(t, root, "disabled")

	manager := NewManager(nil, root)
	manager.ExtensionDAO = dao

	items := []Extension{
		{ExtensionID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "one", Version: "1.0.0", InstallDir: firstDir, Enabled: true},
		{ExtensionID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Name: "two", Version: "1.0.0", InstallDir: secondDir, Enabled: true},
		{ExtensionID: "cccccccccccccccccccccccccccccccc", Name: "disabled", Version: "1.0.0", InstallDir: disabledDir, Enabled: false},
	}
	for _, item := range items {
		if err := dao.Upsert(item); err != nil {
			t.Fatalf("Upsert(%s) failed: %v", item.ExtensionID, err)
		}
	}

	if got := manager.EnabledExtensionDirsForProfile("profile-inherit"); !reflect.DeepEqual(got, []string{firstDir, secondDir}) {
		t.Fatalf("inherit dirs = %v", got)
	}

	_, err := dao.SetProfileSettings("profile-custom", []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"cccccccccccccccccccccccccccccccc",
	}, true)
	if err != nil {
		t.Fatalf("SetProfileSettings failed: %v", err)
	}
	if got := manager.EnabledExtensionDirsForProfile("profile-custom"); !reflect.DeepEqual(got, []string{firstDir}) {
		t.Fatalf("custom dirs = %v, want only enabled selected dir", got)
	}
}

func writeExtensionManifest(t *testing.T, root string, name string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("创建插件目录失败: %v", err)
	}
	manifest := []byte(`{"manifest_version":3,"name":"` + name + `","version":"1.0.0"}`)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatalf("写入 manifest 失败: %v", err)
	}
	return dir
}
