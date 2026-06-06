package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ant-chrome/backend/internal/automation"
)

func TestAutomationScriptRefreshFromLocalDirectory(t *testing.T) {
	app := NewApp(t.TempDir())

	sourceDir := filepath.Join(t.TempDir(), "local-dir-script")
	if err := os.MkdirAll(filepath.Join(sourceDir, "scripts", "helpers"), 0o755); err != nil {
		t.Fatalf("create local dir source failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "automation.script.json"), []byte(`{
  "name": "本地目录脚本",
  "type": "playwright-cdp",
  "entryFile": "scripts/index.cjs"
}`), 0o644); err != nil {
		t.Fatalf("write local dir manifest failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "scripts", "index.cjs"), []byte("const helper = require('./helpers/helper.cjs')\nmodule.exports.run = async () => helper.run()"), 0o644); err != nil {
		t.Fatalf("write local dir entry failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "scripts", "helpers", "helper.cjs"), []byte("module.exports.run = async () => ({ ok: true, source: 'local-dir' })"), 0o644); err != nil {
		t.Fatalf("write local dir helper failed: %v", err)
	}

	saved, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "refresh-local-dir",
		Name:       "旧本地目录脚本",
		Type:       "launch-api",
		Status:     "ready",
		EntryFile:  "index.cjs",
		ScriptText: "module.exports.run = async () => ({ ok: false })",
		Source: automation.ScriptSource{
			Type: "local-dir",
			URI:  sourceDir,
		},
	})
	if err != nil {
		t.Fatalf("AutomationScriptSave returned error: %v", err)
	}

	refreshed, err := app.AutomationScriptRefresh(saved.ID)
	if err != nil {
		t.Fatalf("AutomationScriptRefresh returned error: %v", err)
	}
	if refreshed == nil {
		t.Fatalf("AutomationScriptRefresh returned nil result")
	}
	if refreshed.ID != saved.ID {
		t.Fatalf("expected same script id, got %q want %q", refreshed.ID, saved.ID)
	}
	if refreshed.Status != "ready" {
		t.Fatalf("expected status to be preserved, got %q", refreshed.Status)
	}
	if refreshed.EntryFile != "scripts/index.cjs" {
		t.Fatalf("expected nested entry file, got %q", refreshed.EntryFile)
	}
	if !strings.Contains(refreshed.ScriptText, "helper.run()") {
		t.Fatalf("expected refreshed script text from local directory, got %q", refreshed.ScriptText)
	}
}

func TestImportAutomationLocalLibraryImportsAndUpdatesExistingSource(t *testing.T) {
	app := NewApp(t.TempDir())
	libraryRoot := filepath.Join(t.TempDir(), "script-library")

	firstScriptDir := filepath.Join(libraryRoot, "first-script")
	writeAutomationScriptLibraryPackage(t, firstScriptDir, `{
  "name": "脚本一",
  "type": "playwright-cdp",
  "entryFile": "index.cjs"
}`, "module.exports.run = async () => ({ ok: true, source: 'first-script' })")

	secondScriptDir := filepath.Join(libraryRoot, "second-script")
	if err := os.MkdirAll(secondScriptDir, 0o755); err != nil {
		t.Fatalf("create second script dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondScriptDir, "index.cjs"), []byte("module.exports.run = async () => ({ ok: true, source: 'second-script' })"), 0o644); err != nil {
		t.Fatalf("write second script entry failed: %v", err)
	}

	existing, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "existing-local-library-script",
		Name:       "旧脚本一",
		Type:       "launch-api",
		Status:     "disabled",
		EntryFile:  "index.cjs",
		ScriptText: "module.exports.run = async () => ({ ok: false })",
		Source: automation.ScriptSource{
			Type: "local-dir",
			URI:  firstScriptDir,
		},
		PublicAPI: automation.ScriptPublicAPIConfig{
			Enabled: true,
			Path:    "library/existing-script",
		},
	})
	if err != nil {
		t.Fatalf("AutomationScriptSave returned error: %v", err)
	}

	result, err := app.importAutomationLocalLibrary(libraryRoot)
	if err != nil {
		t.Fatalf("importAutomationLocalLibrary returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("importAutomationLocalLibrary returned nil result")
	}
	if result.Scanned != 2 {
		t.Fatalf("expected scanned count 2, got %d", result.Scanned)
	}
	if len(result.Imported) != 2 {
		t.Fatalf("expected two imported scripts, got %d", len(result.Imported))
	}
	if len(result.Failed) != 0 {
		t.Fatalf("expected no failed imports, got %+v", result.Failed)
	}

	updatedFirst, err := app.AutomationScriptGet(existing.ID)
	if err != nil {
		t.Fatalf("AutomationScriptGet returned error: %v", err)
	}
	if updatedFirst.Name != "脚本一" {
		t.Fatalf("expected existing script to be refreshed from library, got %q", updatedFirst.Name)
	}
	if updatedFirst.Status != "disabled" {
		t.Fatalf("expected existing status to be preserved, got %q", updatedFirst.Status)
	}
	if updatedFirst.Source.Type != "local-dir" || updatedFirst.Source.URI != firstScriptDir {
		t.Fatalf("unexpected updated source: %+v", updatedFirst.Source)
	}
	if !strings.Contains(updatedFirst.ScriptText, "first-script") {
		t.Fatalf("expected refreshed first script body, got %q", updatedFirst.ScriptText)
	}
	if updatedFirst.PublicAPI.Path != "library/existing-script" || !updatedFirst.PublicAPI.Enabled {
		t.Fatalf("expected existing public api config to be preserved, got %+v", updatedFirst.PublicAPI)
	}

	allScripts, err := app.automationScriptStore().List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(allScripts) != 2 {
		t.Fatalf("expected two stored scripts after upsert, got %d", len(allScripts))
	}
}

func TestImportAutomationLocalLibraryContinuesOnSinglePackageFailure(t *testing.T) {
	app := NewApp(t.TempDir())
	libraryRoot := filepath.Join(t.TempDir(), "script-library")

	goodDir := filepath.Join(libraryRoot, "good-script")
	writeAutomationScriptLibraryPackage(t, goodDir, `{
  "name": "好脚本",
  "type": "playwright-cdp",
  "entryFile": "index.cjs"
}`, "module.exports.run = async () => ({ ok: true, source: 'good-script' })")

	badDir := filepath.Join(libraryRoot, "bad-script")
	writeAutomationScriptLibraryPackage(t, badDir, `{
  "name": "坏脚本",
  "type": "playwright-cdp",
  "entryFile": "missing.cjs"
}`, "")

	result, err := app.importAutomationLocalLibrary(libraryRoot)
	if err != nil {
		t.Fatalf("importAutomationLocalLibrary returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("importAutomationLocalLibrary returned nil result")
	}
	if result.Scanned != 2 {
		t.Fatalf("expected scanned count 2, got %d", result.Scanned)
	}
	if len(result.Imported) != 1 {
		t.Fatalf("expected one imported script, got %d", len(result.Imported))
	}
	if len(result.Failed) != 1 {
		t.Fatalf("expected one failed script, got %+v", result.Failed)
	}
	if result.Failed[0].Path != badDir {
		t.Fatalf("unexpected failed path: %+v", result.Failed[0])
	}
	if !strings.Contains(result.Failed[0].Message, "entry file missing.cjs not found") {
		t.Fatalf("unexpected failed message: %+v", result.Failed[0])
	}
}
