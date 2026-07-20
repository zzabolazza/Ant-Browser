package backend

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSaveAndGetUIPreferences(t *testing.T) {
	root := t.TempDir()
	app := NewApp(root)

	if err := app.SaveUIPreferences(map[string]interface{}{"theme": "ocean"}); err != nil {
		t.Fatalf("SaveUIPreferences: %v", err)
	}

	prefs, err := app.GetUIPreferences()
	if err != nil {
		t.Fatalf("GetUIPreferences: %v", err)
	}
	if prefs["theme"] != "ocean" {
		t.Fatalf("loaded theme = %v, want ocean", prefs["theme"])
	}

	path := app.uiPreferencesPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read preferences file %s: %v", path, err)
	}
	var stored map[string]string
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if stored["theme"] != "ocean" {
		t.Fatalf("stored theme = %q, want ocean", stored["theme"])
	}
}

func TestSaveUIPreferencesRejectsInvalidTheme(t *testing.T) {
	app := NewApp(t.TempDir())
	if err := app.SaveUIPreferences(map[string]interface{}{"theme": "neon"}); err == nil {
		t.Fatal("expected invalid theme to fail")
	}
}
