package backend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ant-chrome/backend/internal/automation"
	"ant-chrome/backend/internal/config"
)

func TestAutomationScriptRefreshFromRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "manifest": {
    "name": "远程刷新脚本",
    "description": "来自远程",
    "type": "playwright-cdp",
    "entryFile": "index.cjs"
  },
  "script": "module.exports.run = async () => ({ ok: true, source: 'remote' })"
}`))
	}))
	defer server.Close()

	app := NewApp(t.TempDir())
	saved, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "refresh-remote",
		Name:       "旧远程脚本",
		Type:       "launch-api",
		Status:     "ready",
		EntryFile:  "index.cjs",
		ScriptText: "module.exports.run = async () => ({ ok: false })",
		Source: automation.ScriptSource{
			Type: "remote-url",
			URI:  server.URL + "/script.json",
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
	if refreshed.Name != "远程刷新脚本" {
		t.Fatalf("expected remote manifest name, got %q", refreshed.Name)
	}
	if refreshed.Status != "ready" {
		t.Fatalf("expected status to be preserved, got %q", refreshed.Status)
	}
	if !strings.Contains(refreshed.ScriptText, "source: 'remote'") {
		t.Fatalf("expected refreshed remote script text, got %q", refreshed.ScriptText)
	}
	if refreshed.Source.Type != "remote-url" || refreshed.Source.URI != server.URL+"/script.json" {
		t.Fatalf("unexpected refreshed source: %+v", refreshed.Source)
	}
}

func TestLoadAutomationRemoteBundleSupportsZip(t *testing.T) {
	app := NewApp(t.TempDir())

	zipData := buildAutomationZipBytesForTest(t, map[string]string{
		"automation.script.json": `{
  "name": "远程 ZIP",
  "type": "playwright-cdp",
  "entryFile": "scripts/index.cjs"
}`,
		"scripts/index.cjs": "module.exports.run = async () => ({ ok: true, source: 'remote-zip' })",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(zipData)
	}))
	defer server.Close()

	bundle, err := app.loadAutomationRemoteBundle(server.URL + "/demo.zip")
	if err != nil {
		t.Fatalf("loadAutomationRemoteBundle returned error: %v", err)
	}

	if bundle.Record.Name != "远程 ZIP" {
		t.Fatalf("unexpected bundle name: %s", bundle.Record.Name)
	}
	if bundle.Record.Source.Type != "remote-url" || bundle.Record.Source.URI != server.URL+"/demo.zip" {
		t.Fatalf("unexpected bundle source: %+v", bundle.Record.Source)
	}
	if !strings.Contains(bundle.Record.ScriptText, "remote-zip") {
		t.Fatalf("unexpected script text: %s", bundle.Record.ScriptText)
	}
}

func TestLoadAutomationRemoteBundleBuildsTypeScriptWhenEnabled(t *testing.T) {
	app := NewApp(t.TempDir())
	app.config = config.DefaultConfig()
	app.config.Automation.AllowTypeScriptBuild = true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(`export async function run() {
  return { ok: true, source: 'remote-ts' }
}`))
	}))
	defer server.Close()

	bundle, err := app.loadAutomationRemoteBundle(server.URL + "/demo-script.ts")
	if err != nil {
		t.Fatalf("loadAutomationRemoteBundle returned error: %v", err)
	}

	if bundle.Record.EntryFile != "demo-script.cjs" {
		t.Fatalf("unexpected compiled entry file: %s", bundle.Record.EntryFile)
	}
	if !strings.Contains(bundle.Record.ScriptText, "remote-ts") {
		t.Fatalf("unexpected compiled script text: %s", bundle.Record.ScriptText)
	}
	if bundle.Record.Source.Type != "remote-url" || bundle.Record.Source.URI != server.URL+"/demo-script.ts" {
		t.Fatalf("unexpected bundle source: %+v", bundle.Record.Source)
	}
}

func TestAutomationScriptRefreshFromRemoteTypeScriptWhenEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`export async function run() {
  return { ok: true, source: 'remote-ts-refresh' }
}`))
	}))
	defer server.Close()

	app := NewApp(t.TempDir())
	app.config = config.DefaultConfig()
	app.config.Automation.AllowTypeScriptBuild = true

	saved, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "refresh-remote-ts",
		Name:       "旧远程 TS 脚本",
		Type:       "launch-api",
		Status:     "ready",
		EntryFile:  "index.cjs",
		ScriptText: "module.exports.run = async () => ({ ok: false })",
		Source: automation.ScriptSource{
			Type: "remote-url",
			URI:  server.URL + "/refresh-script.ts",
		},
	})
	if err != nil {
		t.Fatalf("AutomationScriptSave returned error: %v", err)
	}

	refreshed, err := app.AutomationScriptRefresh(saved.ID)
	if err != nil {
		t.Fatalf("AutomationScriptRefresh returned error: %v", err)
	}
	if refreshed.EntryFile != "refresh-script.cjs" {
		t.Fatalf("unexpected refreshed entry file: %s", refreshed.EntryFile)
	}
	if !strings.Contains(refreshed.ScriptText, "remote-ts-refresh") {
		t.Fatalf("unexpected refreshed script text: %s", refreshed.ScriptText)
	}
	if refreshed.Source.Type != "remote-url" || refreshed.Source.URI != server.URL+"/refresh-script.ts" {
		t.Fatalf("unexpected refreshed source: %+v", refreshed.Source)
	}
}

func TestLoadAutomationGitBundleBuildsTypeScriptWhenEnabled(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := filepath.Join(t.TempDir(), "automation-ts-repo")
	if err := os.MkdirAll(filepath.Join(repoDir, "scripts", "demo", "helpers"), 0o755); err != nil {
		t.Fatalf("create repo dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "demo", "automation.script.json"), []byte(`{
  "name": "Git TS 导入",
  "type": "playwright-cdp",
  "entryFile": "index.ts"
}`), 0o644); err != nil {
		t.Fatalf("write git manifest failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "demo", "index.ts"), []byte(`import { flag } from './helpers/flag'

export async function run() {
  return { ok: flag, source: 'git-ts' }
}`), 0o644); err != nil {
		t.Fatalf("write git entry file failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "demo", "helpers", "flag.ts"), []byte(`export const flag = true`), 0o644); err != nil {
		t.Fatalf("write git helper file failed: %v", err)
	}

	runGitForTest(t, repoDir, "init")
	runGitForTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitForTest(t, repoDir, "config", "user.name", "Test User")
	runGitForTest(t, repoDir, "add", ".")
	runGitForTest(t, repoDir, "commit", "-m", "init")

	app := NewApp(t.TempDir())
	app.config = config.DefaultConfig()
	app.config.Automation.AllowTypeScriptBuild = true

	bundle, err := app.loadAutomationGitBundle(repoDir, "", "scripts/demo")
	if err != nil {
		t.Fatalf("loadAutomationGitBundle returned error: %v", err)
	}

	if bundle.Record.Name != "Git TS 导入" {
		t.Fatalf("unexpected bundle name: %s", bundle.Record.Name)
	}
	if bundle.Record.EntryFile != "index.cjs" {
		t.Fatalf("unexpected compiled entry file: %s", bundle.Record.EntryFile)
	}
	if !strings.Contains(bundle.Record.ScriptText, "git-ts") {
		t.Fatalf("unexpected compiled script text: %s", bundle.Record.ScriptText)
	}
	if bundle.Record.Source.Type != "git" || bundle.Record.Source.URI != repoDir || bundle.Record.Source.Path != "scripts/demo" {
		t.Fatalf("unexpected bundle source: %+v", bundle.Record.Source)
	}
}

func TestAutomationScriptRefreshFromGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := filepath.Join(t.TempDir(), "automation-repo")
	if err := os.MkdirAll(filepath.Join(repoDir, "scripts", "demo"), 0o755); err != nil {
		t.Fatalf("create repo dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "demo", "automation.script.json"), []byte(`{
  "name": "Git 刷新脚本",
  "type": "playwright-cdp",
  "entryFile": "index.cjs"
}`), 0o644); err != nil {
		t.Fatalf("write git manifest failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "scripts", "demo", "index.cjs"), []byte("module.exports.run = async () => ({ ok: true, source: 'git' })"), 0o644); err != nil {
		t.Fatalf("write git entry file failed: %v", err)
	}

	runGitForTest(t, repoDir, "init")
	runGitForTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitForTest(t, repoDir, "config", "user.name", "Test User")
	runGitForTest(t, repoDir, "add", ".")
	runGitForTest(t, repoDir, "commit", "-m", "init")

	app := NewApp(t.TempDir())
	saved, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "refresh-git",
		Name:       "旧 Git 脚本",
		Type:       "launch-api",
		Status:     "ready",
		EntryFile:  "index.cjs",
		ScriptText: "module.exports.run = async () => ({ ok: false })",
		Source: automation.ScriptSource{
			Type: "git",
			URI:  repoDir,
			Path: "scripts/demo",
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
	if refreshed.Name != "Git 刷新脚本" {
		t.Fatalf("expected git manifest name, got %q", refreshed.Name)
	}
	if refreshed.Status != "ready" {
		t.Fatalf("expected status to be preserved, got %q", refreshed.Status)
	}
	if !strings.Contains(refreshed.ScriptText, "source: 'git'") {
		t.Fatalf("expected refreshed git script text, got %q", refreshed.ScriptText)
	}
	if refreshed.Source.Type != "git" || refreshed.Source.URI != repoDir || refreshed.Source.Path != "scripts/demo" {
		t.Fatalf("unexpected refreshed source: %+v", refreshed.Source)
	}
}

func TestAutomationScriptRefreshRejectsUnsupportedSource(t *testing.T) {
	app := NewApp(t.TempDir())
	saved, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "refresh-manual",
		Name:       "手动脚本",
		Type:       "playwright-cdp",
		Status:     "ready",
		EntryFile:  "index.cjs",
		ScriptText: "module.exports.run = async () => ({ ok: true })",
		Source: automation.ScriptSource{
			Type: "manual",
		},
	})
	if err != nil {
		t.Fatalf("AutomationScriptSave returned error: %v", err)
	}

	if _, err := app.AutomationScriptRefresh(saved.ID); err == nil {
		t.Fatalf("expected unsupported source refresh to fail")
	}
}
