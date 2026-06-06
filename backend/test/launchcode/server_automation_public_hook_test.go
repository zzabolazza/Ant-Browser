package launchcode_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ant-chrome/backend/internal/automation"
)

func TestAutomationPublicHookStandardModeReturnsEnvelope(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:   "proton-mail-first-message",
			Name: "Proton 邮件搜索并读取最新邮件",
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled:      true,
				Method:       "POST",
				Path:         "mail/proton-first-message",
				RequestMode:  "standard",
				ResponseMode: "envelope",
				TimeoutMs:    120000,
			},
		},
	}
	starter.runResult = &automation.ScriptRunRecord{
		ID:         "run-hook-1",
		ScriptID:   "proton-mail-first-message",
		ScriptName: "Proton 邮件搜索并读取最新邮件",
		Status:     "success",
		Summary:    "已返回最新命中邮件内容",
		ResultText: `{"ok":true,"result":{"verificationCode":"429792","recipientEmail":"target@example.com"}}`,
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/mail/proton-first-message", bytes.NewBufferString(`{
		"instance":{"type":"existing","selector":{"code":"BUYER_001"}},
		"params":{"recipientQuery":"target@example.com"}
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastRunRequest.ScriptID != "proton-mail-first-message" {
		t.Fatalf("scriptId 透传错误: %+v", starter.lastRunRequest)
	}
	if starter.lastRunRequest.UseScriptSelector || starter.lastRunRequest.UseScriptParams {
		t.Fatalf("公共 Hook 应使用请求里的 code/param: %+v", starter.lastRunRequest)
	}
	if starter.lastRunRequest.TargetMode != "existing" || starter.lastRunRequest.SelectorText != "" {
		t.Fatalf("instance 转换错误: %+v", starter.lastRunRequest)
	}
	if targetInput := automationRunTargetInputJSON(t, starter.lastRunRequest.TargetInput); !strings.Contains(targetInput, `"code":"BUYER_001"`) {
		t.Fatalf("targetInput 转换错误: %s", targetInput)
	}
	if starter.lastRunRequest.ParamsText != `{"recipientQuery":"target@example.com"}` {
		t.Fatalf("paramsText 转换错误: %s", starter.lastRunRequest.ParamsText)
	}

	var resp struct {
		OK     bool                   `json:"ok"`
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if !resp.OK || resp.Status != "success" {
		t.Fatalf("hook 响应错误: %+v", resp)
	}
	if resp.Data["verificationCode"] != "429792" {
		t.Fatalf("expected data payload, got %+v", resp.Data)
	}
}

func TestAutomationPublicHookSupportsInstanceModes(t *testing.T) {
	cases := []struct {
		name             string
		body             string
		expectedMode     string
		expectedSelector bool
		expectedInput    string
		useScriptTarget  bool
	}{
		{
			name: "rotate",
			body: `{
				"instance":{"type":"rotate","selector":{"groupId":"group-a","tags":["chatgpt"]}},
				"params":{"recipientQuery":"target@example.com"}
			}`,
			expectedMode:  "rotate",
			expectedInput: `"groupId":"group-a"`,
		},
		{
			name: "create",
			body: `{
				"instance":{"type":"create","templateSelector":{"code":"TEMPLATE_001"},"createNameTemplate":"ChatGPT-Image-${timestamp}"},
				"params":{"recipientQuery":"target@example.com"}
			}`,
			expectedMode:  "create",
			expectedInput: `"createNameTemplate":"ChatGPT-Image-${timestamp}"`,
		},
		{
			name: "script-default",
			body: `{
				"instance":{"type":"script-default"},
				"params":{"recipientQuery":"target@example.com"}
			}`,
			useScriptTarget: true,
		},
		{
			name: "legacy-code",
			body: `{
				"code":"BUYER_001",
				"params":{"recipientQuery":"target@example.com"}
			}`,
			expectedSelector: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newInMemoryService()
			starter := newMockAutomationStarter()
			starter.scripts = []automation.ScriptRecord{
				{
					ID:   "proton-mail-first-message",
					Name: "Proton 邮件搜索并读取最新邮件",
					PublicAPI: automation.ScriptPublicAPIConfig{
						Enabled:      true,
						Method:       "POST",
						Path:         "mail/proton-first-message",
						RequestMode:  "standard",
						ResponseMode: "envelope",
						TimeoutMs:    120000,
					},
				},
			}

			handler := buildTestHandlerWithManager(svc, starter, nil)
			req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/mail/proton-first-message", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
			}
			if starter.lastRunRequest.TargetMode != tc.expectedMode {
				t.Fatalf("targetMode 错误: %+v", starter.lastRunRequest)
			}
			if starter.lastRunRequest.UseScriptSelector != tc.useScriptTarget {
				t.Fatalf("useScriptSelector 错误: %+v", starter.lastRunRequest)
			}
			if tc.expectedInput != "" {
				targetInput := automationRunTargetInputJSON(t, starter.lastRunRequest.TargetInput)
				if !strings.Contains(targetInput, tc.expectedInput) {
					t.Fatalf("targetInput 转换错误: %s", targetInput)
				}
			}
			if tc.expectedSelector && starter.lastRunRequest.SelectorText != `{"code":"BUYER_001"}` {
				t.Fatalf("legacy selectorText 转换错误: %+v", starter.lastRunRequest)
			}
		})
	}
}

func TestAutomationPublicHookParamsOnlyModeReturnsResultOnly(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:   "proton-mail-first-message",
			Name: "Proton 邮件搜索并读取最新邮件",
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled:      true,
				Method:       "POST",
				Path:         "mail/proton-result-only",
				RequestMode:  "params-only",
				ResponseMode: "result-only",
				TimeoutMs:    45000,
			},
		},
	}
	starter.runResult = &automation.ScriptRunRecord{
		ID:         "run-hook-2",
		ScriptID:   "proton-mail-first-message",
		ScriptName: "Proton 邮件搜索并读取最新邮件",
		Status:     "success",
		Summary:    "已返回最新命中邮件内容",
		ResultText: `{"ok":true,"result":{"verificationCode":"429792","mailboxName":"ChatGPT"}}`,
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/mail/proton-result-only?timeoutMs=60000", bytes.NewBufferString(`{
		"code":"BUYER_001",
		"params":{"recipientQuery":"target@example.com"}
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastRunRequest.UseScriptSelector || starter.lastRunRequest.UseScriptParams {
		t.Fatalf("公共 Hook 应透传 code/param: %+v", starter.lastRunRequest)
	}
	if starter.lastRunRequest.ParamsText != `{"recipientQuery":"target@example.com"}` {
		t.Fatalf("paramsText 转换错误: %s", starter.lastRunRequest.ParamsText)
	}
	if starter.lastRunRequest.TimeoutMs != 60000 {
		t.Fatalf("timeoutMs 透传错误: %+v", starter.lastRunRequest)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok || data["verificationCode"] != "429792" || data["mailboxName"] != "ChatGPT" {
		t.Fatalf("expected data payload, got %+v", resp)
	}
}

func TestAutomationPublicHookResultOnlyCompactsDownloadFields(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:   "grok-image-generate-download",
			Name: "Grok 生成图片并下载",
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled:      true,
				Method:       "POST",
				Path:         "image/grok-generate-download",
				RequestMode:  "params-only",
				ResponseMode: "result-only",
				TimeoutMs:    300000,
			},
		},
	}
	starter.runResult = &automation.ScriptRunRecord{
		ID:         "run-grok-image",
		ScriptID:   "grok-image-generate-download",
		ScriptName: "Grok 生成图片并下载",
		Status:     "success",
		Summary:    "Grok 图片已生成并下载",
		ResultText: `{"ok":true,"downloadAddress":"D:/tmp/grok.png","downloadPath":"D:/tmp/grok.png","sourceImageUrl":"https://example.com/image.png","steps":[{"step":"open"}],"startedAt":"2026-06-03T00:00:00Z"}`,
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/image/grok-generate-download", bytes.NewBufferString(`{"code":"BUYER_001","params":{"prompt":"ant"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok || data["downloadAddress"] != "D:/tmp/grok.png" || data["downloadPath"] != "D:/tmp/grok.png" || data["sourceImageUrl"] != "https://example.com/image.png" {
		t.Fatalf("下载字段缺失: %+v", resp)
	}
	if _, exists := data["steps"]; exists {
		t.Fatalf("不应返回冗余 steps: %+v", resp)
	}
	if _, exists := data["runId"]; exists {
		t.Fatalf("不应返回 runId: %+v", resp)
	}
}

func TestAutomationPublicHookAppliesRequestBodyVariables(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:         "image-generate",
			Name:       "图片生成",
			ParamsText: `{"prompt":"默认提示词","selectors":{"promptInput":"#prompt-textarea","generatedImage":"img.generated"},"timeoutMs":300000}`,
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled:         true,
				Method:          "POST",
				Path:            "image/generate",
				RequestMode:     "standard",
				ResponseMode:    "envelope",
				TimeoutMs:       300000,
				RequestBodyText: `{"instance":{"type":"existing","selector":{"code":"{{code}}"}},"params":{"prompt":"{{prompt}}","outputFileName":"{{outputFileName}}"}}`,
				Variables: []automation.ScriptPublicAPIVariable{
					{Name: "prompt", DefaultValue: "默认提示词", Required: true},
					{Name: "outputFileName", DefaultValue: "generated-image.png"},
				},
			},
		},
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/image/generate", bytes.NewBufferString(`{
		"code":"BUYER_001",
		"params":{"prompt":"海边的机器人","outputFileName":"robot.png"}
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastRunRequest.UseScriptParams {
		t.Fatalf("变量模板应生成 params: %+v", starter.lastRunRequest)
	}
	if starter.lastRunRequest.TargetMode != "existing" || starter.lastRunRequest.SelectorText != "" {
		t.Fatalf("变量模板应生成 instance target: %+v", starter.lastRunRequest)
	}
	if targetInput := automationRunTargetInputJSON(t, starter.lastRunRequest.TargetInput); !strings.Contains(targetInput, `"code":"BUYER_001"`) {
		t.Fatalf("变量模板 targetInput 错误: %s", targetInput)
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(starter.lastRunRequest.ParamsText), &params); err != nil {
		t.Fatalf("解析 paramsText 失败: %v", err)
	}
	if params["prompt"] != "海边的机器人" || params["outputFileName"] != "robot.png" {
		t.Fatalf("变量未映射到 params: %+v", params)
	}
	selectors, ok := params["selectors"].(map[string]interface{})
	if !ok || selectors["promptInput"] != "#prompt-textarea" || selectors["generatedImage"] != "img.generated" {
		t.Fatalf("默认 selectors 不应被变量模板覆盖丢失: %+v", params)
	}
	if params["timeoutMs"] != float64(300000) {
		t.Fatalf("默认 timeoutMs 不应被变量模板覆盖丢失: %+v", params)
	}
	if starter.lastRunRequest.TimeoutMs != 300000 {
		t.Fatalf("timeoutMs 错误: %+v", starter.lastRunRequest)
	}
}

func TestAutomationPublicHookReturnsNotFoundWhenDisabled(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:   "disabled-hook",
			Name: "Disabled Hook",
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled: false,
				Path:    "mail/disabled-hook",
			},
		},
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/mail/disabled-hook", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestAutomationPublicHookRejectsLegacyParamAndTopLevelVariables(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:   "strict-hook",
			Name: "Strict Hook",
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled:   true,
				Method:    "POST",
				Path:      "mail/strict-hook",
				TimeoutMs: 120000,
			},
		},
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	cases := []struct {
		name string
		body string
	}{
		{name: "legacy-param", body: `{"code":"BUYER_001","param":{"recipientQuery":"target@example.com"}}`},
		{name: "top-level-variable", body: `{"code":"BUYER_001","recipientQuery":"target@example.com"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/mail/strict-hook", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("期望 400，实际 %d，body=%s", w.Code, w.Body.String())
			}

			var resp struct {
				OK    bool `json:"ok"`
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}
			if resp.OK || resp.Error.Code != "invalid_request" {
				t.Fatalf("错误 envelope 不正确: %+v", resp)
			}
		})
	}
}

func TestAutomationPublicHookRejectsInvalidTimeout(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.scripts = []automation.ScriptRecord{
		{
			ID:   "strict-hook-timeout",
			Name: "Strict Hook Timeout",
			PublicAPI: automation.ScriptPublicAPIConfig{
				Enabled:   true,
				Method:    "POST",
				Path:      "mail/strict-hook-timeout",
				TimeoutMs: 120000,
			},
		},
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/hooks/mail/strict-hook-timeout", bytes.NewBufferString(`{"code":"BUYER_001","params":{},"timeoutMs":999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "timeoutMs must be between 1000 and 1800000") {
		t.Fatalf("错误信息不正确: %s", w.Body.String())
	}
}
