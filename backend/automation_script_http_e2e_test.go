package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"ant-chrome/backend/internal/automation"
	"ant-chrome/backend/internal/config"
)

type automationHTTPProfileCreateResponse struct {
	OK         bool   `json:"ok"`
	Created    bool   `json:"created"`
	Launched   bool   `json:"launched"`
	ProfileID  string `json:"profileId"`
	LaunchCode string `json:"launchCode"`
}

type automationHTTPScriptsResponse struct {
	OK   bool `json:"ok"`
	Data struct {
		Count int `json:"count"`
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	} `json:"data"`
}

type automationHTTPRunResponse struct {
	OK   bool `json:"ok"`
	Data struct {
		Run struct {
			Status     string `json:"status"`
			Summary    string `json:"summary"`
			Error      string `json:"error"`
			ResultText string `json:"resultText"`
		} `json:"run"`
	} `json:"data"`
}

type automationHTTPHookEnvelopeResponse struct {
	OK      bool                   `json:"ok"`
	Status  string                 `json:"status"`
	Summary string                 `json:"summary"`
	Result  map[string]interface{} `json:"result"`
}

type automationHTTPLaunchLogsResponse struct {
	OK    bool              `json:"ok"`
	Items []json.RawMessage `json:"items"`
}

func TestAutomationScriptRunHTTPReturnsSavedMailProbeScript(t *testing.T) {
	nodePath := lookupAutomationHTTPProbeNode(t)
	chromePath := lookupAutomationHTTPProbeChrome(t)
	repoRoot := automationHTTPRepoRoot(t)

	tempRoot := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Logging.FileEnabled = false
	cfg.LaunchServer.Port = automationHTTPFreePort(t)
	cfg.Automation.Enabled = true
	cfg.Automation.NodeSource = config.AutomationNodeSourceSystem
	cfg.Automation.SystemNodePath = nodePath
	cfg.Automation.HeadlessDefault = true
	if err := cfg.Save(filepath.Join(tempRoot, "config.yaml")); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	if err := prepareAutomationHTTPRuntime(tempRoot, repoRoot, cfg.Automation.RuntimeVersion); err != nil {
		t.Fatalf("prepare runtime failed: %v", err)
	}

	fixtureServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, automationHTTPMailFixtureHTML)
	}))
	defer fixtureServer.Close()

	app := NewApp(tempRoot)
	Start(app, nil)
	defer Stop(app, nil)

	if err := app.BrowserCoreSave(BrowserCoreInput{
		CoreId:    "system-chrome",
		CoreName:  "System Chrome",
		CorePath:  chromePath,
		IsDefault: true,
	}); err != nil {
		t.Fatalf("save core failed: %v", err)
	}
	if err := app.BrowserCoreSetDefault("system-chrome"); err != nil {
		t.Fatalf("set default core failed: %v", err)
	}

	baseURL, ok := app.GetLaunchServerInfo()["baseUrl"].(string)
	if !ok || strings.TrimSpace(baseURL) == "" {
		t.Fatalf("launch server baseUrl missing: %+v", app.GetLaunchServerInfo())
	}

	var createResp automationHTTPProfileCreateResponse
	if err := automationHTTPRequestJSON(http.MethodPost, baseURL+"/api/profiles", map[string]any{
		"profile": map[string]any{
			"profileName": "mail-probe",
			"launchArgs": []string{
				"--headless=new",
				"--disable-gpu",
				"--no-first-run",
				"--no-default-browser-check",
				"--window-size=1440,1024",
			},
		},
		"launchCode": "MAIL01",
	}, &createResp); err != nil {
		t.Fatalf("create profile via http failed: %v", err)
	}
	if !createResp.OK || !createResp.Created || createResp.LaunchCode != "MAIL01" {
		t.Fatalf("unexpected create response: %+v", createResp)
	}

	savedScript, err := app.AutomationScriptSave(automation.ScriptRecord{
		ID:         "mail-probe-script",
		Name:       "测试邮件探针",
		Type:       "playwright-cdp",
		Status:     "ready",
		EntryFile:  "index.cjs",
		ScriptText: automationHTTPMailProbeScriptText,
	})
	if err != nil {
		t.Fatalf("save mail probe script failed: %v", err)
	}
	if savedScript == nil {
		t.Fatalf("expected mail probe script to be saved")
	}
	savedScript.PublicAPI = automation.ScriptPublicAPIConfig{
		Enabled:      true,
		Method:       "POST",
		Path:         "mail/probe-message",
		RequestMode:  "params-only",
		ResponseMode: "envelope",
		TimeoutMs:    120000,
	}
	savedScript.SelectorText = fmt.Sprintf("{\n  \"code\": %q\n}", createResp.LaunchCode)
	if _, err := app.AutomationScriptSave(*savedScript); err != nil {
		t.Fatalf("save mail probe public api config failed: %v", err)
	}

	var scriptsResp automationHTTPScriptsResponse
	if err := automationHTTPRequestJSON(http.MethodGet, baseURL+"/api/automation/scripts", nil, &scriptsResp); err != nil {
		t.Fatalf("list scripts via http failed: %v", err)
	}
	if !scriptsResp.OK || scriptsResp.Data.Count == 0 {
		t.Fatalf("unexpected scripts response: %+v", scriptsResp)
	}
	if !automationHTTPHasScript(scriptsResp.Data.Items, "mail-probe-script") {
		t.Fatalf("saved mail probe script missing: %+v", scriptsResp)
	}

	var runResp automationHTTPRunResponse
	runErr := automationHTTPRequestJSON(http.MethodPost, baseURL+"/api/automation/scripts/run", map[string]any{
		"scriptId": "mail-probe-script",
		"selector": map[string]any{
			"code": createResp.LaunchCode,
		},
		"params": map[string]any{
			"inboxUrl":  fixtureServer.URL,
			"timeoutMs": 45000,
		},
		"timeoutMs": 120000,
	}, &runResp)
	if runErr != nil {
		t.Fatalf("run script via http failed: %v", runErr)
	}

	parsed := make(map[string]any)
	if text := strings.TrimSpace(runResp.Data.Run.ResultText); text != "" {
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			t.Fatalf("parse run result failed: %v; result=%s", err, text)
		}
		if nested, ok := parsed["result"].(map[string]any); ok && len(nested) > 0 {
			parsed = nested
		}
	}
	if runResp.Data.Run.Status != "success" {
		var logsResp automationHTTPLaunchLogsResponse
		_ = automationHTTPRequestJSON(http.MethodGet, baseURL+"/api/launch/logs?limit=10", nil, &logsResp)
		t.Fatalf("unexpected run response: status=%s summary=%s error=%s logs=%s",
			runResp.Data.Run.Status,
			runResp.Data.Run.Summary,
			runResp.Data.Run.Error,
			automationHTTPMarshal(t, logsResp),
		)
	}

	if got := automationHTTPStringValue(parsed, "mailboxName"); got != "ChatGPT" {
		t.Fatalf("unexpected mailboxName: %q parsed=%s", got, automationHTTPMarshal(t, parsed))
	}
	if got := automationHTTPStringValue(parsed, "senderEmail"); got != "noreply@tm.openai.com" {
		t.Fatalf("unexpected senderEmail: %q parsed=%s", got, automationHTTPMarshal(t, parsed))
	}
	if got := automationHTTPStringValue(parsed, "recipientEmail"); got != "target@example.com" {
		t.Fatalf("unexpected recipientEmail: %q parsed=%s", got, automationHTTPMarshal(t, parsed))
	}
	if got := automationHTTPStringValue(parsed, "verificationCode"); got != "429792" {
		t.Fatalf("unexpected verificationCode: %q parsed=%s", got, automationHTTPMarshal(t, parsed))
	}
	if got := parsed["permissionApplied"]; got != true {
		t.Fatalf("expected permissionApplied=true, got %#v parsed=%s", got, automationHTTPMarshal(t, parsed))
	}
	if got := automationHTTPStringValue(parsed, "permissionOrigin"); got != fixtureServer.URL {
		t.Fatalf("unexpected permissionOrigin: %q parsed=%s", got, automationHTTPMarshal(t, parsed))
	}
	signature := automationHTTPStringValue(parsed, "signature")
	if !strings.Contains(signature, "Best regards") || !strings.Contains(signature, "ChatGPT") {
		t.Fatalf("unexpected signature: %q parsed=%s", signature, automationHTTPMarshal(t, parsed))
	}

	var hookResp automationHTTPHookEnvelopeResponse
	hookErr := automationHTTPRequestJSON(http.MethodPost, baseURL+"/api/automation/hooks/mail/probe-message", map[string]any{
		"params": map[string]any{
			"inboxUrl": fixtureServer.URL,
		},
		"timeoutMs": 45000,
	}, &hookResp)
	if hookErr != nil {
		t.Fatalf("run public hook via http failed: %v", hookErr)
	}
	if !hookResp.OK || hookResp.Status != "success" {
		t.Fatalf("unexpected hook response: %+v", hookResp)
	}
	if got := automationHTTPStringValue(hookResp.Result, "verificationCode"); got != "429792" {
		t.Fatalf("unexpected hook verificationCode: %q resp=%s", got, automationHTTPMarshal(t, hookResp))
	}
	if got := automationHTTPStringValue(hookResp.Result, "senderEmail"); got != "noreply@tm.openai.com" {
		t.Fatalf("unexpected hook senderEmail: %q resp=%s", got, automationHTTPMarshal(t, hookResp))
	}

	t.Logf("automation http result: %s", automationHTTPMarshal(t, map[string]any{
		"profileId":        createResp.ProfileID,
		"launchCode":       createResp.LaunchCode,
		"runStatus":        runResp.Data.Run.Status,
		"runSummary":       runResp.Data.Run.Summary,
		"hookStatus":       hookResp.Status,
		"hookSummary":      hookResp.Summary,
		"mailboxName":      automationHTTPStringValue(parsed, "mailboxName"),
		"senderEmail":      automationHTTPStringValue(parsed, "senderEmail"),
		"recipientEmail":   automationHTTPStringValue(parsed, "recipientEmail"),
		"verificationCode": automationHTTPStringValue(parsed, "verificationCode"),
		"signature":        signature,
		"subject":          automationHTTPStringValue(parsed, "subject"),
	}))
}
