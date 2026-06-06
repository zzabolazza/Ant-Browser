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

func TestAutomationScriptRunsEndpointPassesLimit(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	starter.runs = []automation.ScriptRunRecord{
		{ID: "run-1", ScriptID: "script-a", Status: "success"},
		{ID: "run-2", ScriptID: "script-b", Status: "failed"},
	}

	handler := buildTestHandlerWithManager(svc, starter, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/scripts/runs?limit=1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastRunListLimit != 1 {
		t.Fatalf("limit 透传错误: %d", starter.lastRunListLimit)
	}

	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			Count int                          `json:"count"`
			Items []automation.ScriptRunRecord `json:"items"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if !resp.OK || resp.Data.Count != 1 || len(resp.Data.Items) != 1 {
		t.Fatalf("runs 响应错误: %+v", resp)
	}
}

func TestAutomationScriptAPIUnavailableReturnsServiceUnavailable(t *testing.T) {
	handler := buildTestHandlerWithManager(newInMemoryService(), newMockStarterWithParams(), nil)
	req := httptest.NewRequest(http.MethodGet, "/api/automation/scripts", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("期望 503，实际 %d，body=%s", w.Code, w.Body.String())
	}
}

func TestAutomationScriptRunEndpointRejectsInvalidBody(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockAutomationStarter()
	handler := buildTestHandlerWithManager(svc, starter, nil)

	t.Run("invalid-json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/automation/scripts/run", bytes.NewBufferString("{bad json}"))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("期望 400，实际 %d，body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("selector-must-be-object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/automation/scripts/run", bytes.NewBufferString(`{
			"scriptId":"news-query-txt",
			"selector":"BUYER_001"
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("期望 400，实际 %d，body=%s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "selector must be a JSON object") {
			t.Fatalf("错误信息不正确: %s", w.Body.String())
		}
	})

	t.Run("timeout-must-be-in-range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/automation/scripts/run", bytes.NewBufferString(`{
			"scriptId":"news-query-txt",
			"timeoutMs":1800001
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("期望 400，实际 %d，body=%s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "timeoutMs must be between 1000 and 1800000") {
			t.Fatalf("错误信息不正确: %s", w.Body.String())
		}
	})
}
