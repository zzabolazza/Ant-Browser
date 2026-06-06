package launchcode_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/launchcode"
)

func buildTestHandlerWithManager(svc *launchcode.LaunchCodeService, starter launchcode.BrowserStarter, mgr *browser.Manager) http.Handler {
	srv := launchcode.NewLaunchServer(svc, starter, mgr, 0)
	return launchcode.NewTestHandler(srv)
}

func newSelectorTestManager(profiles ...*browser.Profile) *browser.Manager {
	items := make(map[string]*browser.Profile, len(profiles))
	for _, profile := range profiles {
		items[profile.ProfileId] = profile
	}
	return &browser.Manager{
		Profiles: items,
	}
}

func TestLaunchWithKeywordSelector(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profile := &browser.Profile{
		ProfileId:   "profile-keyword",
		ProfileName: "Amazon US",
		GroupId:     "group-sales",
		Tags:        []string{"电商", "北美"},
		Keywords:    []string{"amazon-us", "checkout", "buyer-account"},
		Pid:         9527,
		DebugPort:   9333,
	}
	starter.addProfile(profile)
	manager := newSelectorTestManager(profile)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	payload := bytes.NewBufferString(`{
		"selector": {
			"keyword": "checkout",
			"tags": ["电商"],
			"groupId": "group-sales"
		},
		"skipDefaultStartUrls": true
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/launch", payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profile.ProfileId {
		t.Fatalf("命中实例错误: got=%s want=%s", starter.lastProfile, profile.ProfileId)
	}

	var resp struct {
		OK         bool   `json:"ok"`
		ProfileID  string `json:"profileId"`
		LaunchCode string `json:"launchCode"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if !resp.OK || resp.ProfileID != profile.ProfileId {
		t.Fatalf("响应不正确: %+v", resp)
	}
	if strings.TrimSpace(resp.LaunchCode) == "" {
		t.Fatalf("期望返回 resolved launchCode，实际为空: %+v", resp)
	}
}

func TestLaunchWithTopLevelKeywordSelector(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profile := &browser.Profile{
		ProfileId:   "profile-top-level",
		ProfileName: "Billing Ops",
		Keywords:    []string{"billing", "invoice"},
		Pid:         1001,
		DebugPort:   9444,
	}
	starter.addProfile(profile)
	manager := newSelectorTestManager(profile)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"keyword":"billing"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profile.ProfileId {
		t.Fatalf("命中实例错误: got=%s want=%s", starter.lastProfile, profile.ProfileId)
	}
}

func TestLaunchWithTopLevelKeyAliasSelector(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profile := &browser.Profile{
		ProfileId:   "profile-top-level-key",
		ProfileName: "Buyer Account",
		Keywords:    []string{"buyer-001", "amazon"},
		Pid:         1002,
		DebugPort:   9445,
	}
	starter.addProfile(profile)
	manager := newSelectorTestManager(profile)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"key":"buyer-001"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profile.ProfileId {
		t.Fatalf("命中实例错误: got=%s want=%s", starter.lastProfile, profile.ProfileId)
	}
}

func TestLaunchWithTopLevelKeyPrefersExactKeywordMatch(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileFuzzy := &browser.Profile{
		ProfileId:   "profile-key-fuzzy",
		ProfileName: "Account A",
		Keywords:    []string{"buyer-001-old", "amazon"},
		Pid:         1004,
		DebugPort:   9447,
	}
	profileExact := &browser.Profile{
		ProfileId:   "profile-key-exact",
		ProfileName: "Z Account",
		Keywords:    []string{"buyer-001", "amazon"},
		Pid:         1005,
		DebugPort:   9448,
	}
	starter.addProfile(profileFuzzy)
	starter.addProfile(profileExact)
	manager := newSelectorTestManager(profileFuzzy, profileExact)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"key":"buyer-001"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profileExact.ProfileId {
		t.Fatalf("key 应优先命中精确关键字实例: got=%s want=%s", starter.lastProfile, profileExact.ProfileId)
	}
}

func TestLaunchWithNestedSelectorKeyPrefersExactKeywordMatch(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileFuzzy := &browser.Profile{
		ProfileId:   "profile-selector-key-fuzzy",
		ProfileName: "Account A",
		Keywords:    []string{"buyer-001-old", "amazon"},
		Pid:         1006,
		DebugPort:   9449,
	}
	profileExact := &browser.Profile{
		ProfileId:   "profile-selector-key-exact",
		ProfileName: "Z Account",
		Keywords:    []string{"buyer-001", "amazon"},
		Pid:         1007,
		DebugPort:   9450,
	}
	starter.addProfile(profileFuzzy)
	starter.addProfile(profileExact)
	manager := newSelectorTestManager(profileFuzzy, profileExact)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"selector":{"key":"buyer-001"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profileExact.ProfileId {
		t.Fatalf("selector.key 应优先命中精确关键字实例: got=%s want=%s", starter.lastProfile, profileExact.ProfileId)
	}
}

func TestLaunchWithTopLevelCodeFallbackToKeyword(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profile := &browser.Profile{
		ProfileId:   "profile-top-level-code-fallback",
		ProfileName: "Buyer Account 01",
		Keywords:    []string{"buyer-001", "amazon"},
		Pid:         1003,
		DebugPort:   9446,
	}
	starter.addProfile(profile)
	manager := newSelectorTestManager(profile)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"code":"buyer-001"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profile.ProfileId {
		t.Fatalf("code 关键字兜底命中实例错误: got=%s want=%s", starter.lastProfile, profile.ProfileId)
	}
}

func TestLaunchWithTopLevelCodeFallbackPrefersExactKeywordMatch(t *testing.T) {
	svc := newInMemoryService()
	starter := newMockStarterWithParams()

	profileFuzzy := &browser.Profile{
		ProfileId:   "profile-code-fuzzy",
		ProfileName: "Account A",
		Keywords:    []string{"buyer-001-old", "amazon"},
		Pid:         1008,
		DebugPort:   9451,
	}
	profileExact := &browser.Profile{
		ProfileId:   "profile-code-exact",
		ProfileName: "Z Account",
		Keywords:    []string{"buyer-001", "amazon"},
		Pid:         1009,
		DebugPort:   9452,
	}
	starter.addProfile(profileFuzzy)
	starter.addProfile(profileExact)
	manager := newSelectorTestManager(profileFuzzy, profileExact)

	handler := buildTestHandlerWithManager(svc, starter, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"code":"buyer-001"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profileExact.ProfileId {
		t.Fatalf("code 关键字兜底应优先命中精确关键字实例: got=%s want=%s", starter.lastProfile, profileExact.ProfileId)
	}
}

func TestLaunchWithAmbiguousKeywordSelectorReturnsFirstByDefault(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"selector":{"keyword":"shop"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profileA.ProfileId {
		t.Fatalf("关键字多命中时应默认取排序后的第一个实例: got=%s want=%s", starter.lastProfile, profileA.ProfileId)
	}
}

func TestLaunchWithTopLevelCodeFallbackReturnsFirstByDefault(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodPost, "/api/launch", bytes.NewBufferString(`{"code":"shop"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d，body=%s", w.Code, w.Body.String())
	}
	if starter.lastProfile != profileA.ProfileId {
		t.Fatalf("code 关键字兜底多命中时应默认取排序后的第一个实例: got=%s want=%s", starter.lastProfile, profileA.ProfileId)
	}
}
