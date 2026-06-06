package launchcode_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ant-chrome/backend/internal/browser"
)

func TestLaunchWithAmbiguousKeywordSelectorAndExplicitUniqueReturnsConflict(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileA := &browser.Profile{
		ProfileId:   "profile-a",
		ProfileName: "Account A",
		Keywords:    []string{"shop", "checkout"},
		Pid:         1001,
		DebugPort:   9441,
	}
	profileB := &browser.Profile{
		ProfileId:   "profile-b",
		ProfileName: "Account B",
		Keywords:    []string{"shop", "refund"},
		Pid:         1002,
		DebugPort:   9442,
	}
	starter.addProfile(profileA)
	starter.addProfile(profileB)
	manager := newSelectorTestManager(profileA, profileB)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"selector":{"keyword":"shop","matchMode":"unique"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != "" {
		t.Fatalf("歧义场景不应启动实例: %s", starter.lastProfile)
	}
	if !strings.Contains(w.Body.String(), "matchMode=first") {
		t.Fatalf("错误信息未提示 matchMode=first: %s", w.Body.String())
	}
}

func TestGetLaunchByCodeDoesNotFallbackToKeyword(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profile := &browser.Profile{
		ProfileId:   "profile-get-code-only",
		ProfileName: "Buyer Account 02",
		Keywords:    []string{"buyer-002"},
		Pid:         1004,
		DebugPort:   9447,
	}
	starter.addProfile(profile)
	manager := newSelectorTestManager(profile)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodGet, "/api/launch/buyer-002", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /api/launch/{code} 应保持纯 code 语义，期望 404，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != "" {
		t.Fatalf("GET /api/launch/{code} 不应按关键字兜底启动实例: %s", starter.lastProfile)
	}
}

func TestLaunchWithMatchModeFirst(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileB := &browser.Profile{
		ProfileId:   "profile-b",
		ProfileName: "B Account",
		Keywords:    []string{"shop"},
		Pid:         2002,
		DebugPort:   9552,
	}
	profileA := &browser.Profile{
		ProfileId:   "profile-a",
		ProfileName: "A Account",
		Keywords:    []string{"shop"},
		Pid:         2001,
		DebugPort:   9551,
	}
	starter.addProfile(profileA)
	starter.addProfile(profileB)
	manager := newSelectorTestManager(profileB, profileA)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"selector":{"keyword":"shop","matchMode":"first"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profileA.ProfileId {
		t.Fatalf("matchMode=first 应命中排序后的第一个实例: got=%s want=%s", starter.lastProfile, profileA.ProfileId)
	}
}

func TestLaunchWithMatchModeAllStartsAllMatchedProfiles(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileA := &browser.Profile{
		ProfileId:   "profile-a",
		ProfileName: "A Account",
		Keywords:    []string{"shop"},
		Pid:         2001,
		DebugPort:   9551,
	}
	profileB := &browser.Profile{
		ProfileId:   "profile-b",
		ProfileName: "B Account",
		Keywords:    []string{"shop"},
		Pid:         2002,
		DebugPort:   9552,
	}
	starter.addProfile(profileA)
	starter.addProfile(profileB)
	manager := newSelectorTestManager(profileB, profileA)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"selector":{"keyword":"shop","matchMode":"all"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if len(starter.started) != 2 {
		t.Fatalf("matchMode=all 应启动 2 个实例: %+v", starter.started)
	}
	if starter.started[0] != profileA.ProfileId || starter.started[1] != profileB.ProfileId {
		t.Fatalf("matchMode=all 应按稳定排序依次启动: got=%+v", starter.started)
	}

	var resp struct {
		OK    bool `json:"ok"`
		Count int  `json:"count"`
		Items []struct {
			ProfileID string `json:"profileId"`
			IsActive  bool   `json:"isActive"`
		} `json:"items"`
		ActiveProfileID string `json:"activeProfileId"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if !resp.OK || resp.Count != 2 || len(resp.Items) != 2 {
		t.Fatalf("批量启动响应错误: %+v", resp)
	}
	if resp.ActiveProfileID != profileB.ProfileId {
		t.Fatalf("activeProfileId 错误: got=%s want=%s", resp.ActiveProfileID, profileB.ProfileId)
	}
	if resp.Items[0].ProfileID != profileA.ProfileId || resp.Items[1].ProfileID != profileB.ProfileId {
		t.Fatalf("items 顺序错误: %+v", resp.Items)
	}
	if resp.Items[0].IsActive || !resp.Items[1].IsActive {
		t.Fatalf("isActive 标记错误: %+v", resp.Items)
	}
}

func TestLaunchWithTopLevelCodeFallbackAndExplicitUniqueReturnsConflict(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileA := &browser.Profile{
		ProfileId:   "profile-a",
		ProfileName: "Account A",
		Keywords:    []string{"shop", "checkout"},
		Pid:         1001,
		DebugPort:   9441,
	}
	profileB := &browser.Profile{
		ProfileId:   "profile-b",
		ProfileName: "Account B",
		Keywords:    []string{"shop", "refund"},
		Pid:         1002,
		DebugPort:   9442,
	}
	starter.addProfile(profileA)
	starter.addProfile(profileB)
	manager := newSelectorTestManager(profileA, profileB)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"code":"shop","matchMode":"unique"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if len(starter.started) != 0 {
		t.Fatalf("显式 unique 不应启动任何实例: %+v", starter.started)
	}
}
