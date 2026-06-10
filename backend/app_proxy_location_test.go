package backend

import "testing"

func TestResolveProxyLocationOptionUsesCityTimezone(t *testing.T) {
	option := resolveProxyLocationOption("US", "Los Angeles")
	if option.Timezone != "America/Los_Angeles" {
		t.Fatalf("timezone = %q, want America/Los_Angeles", option.Timezone)
	}
	if option.Lang != "en-US" {
		t.Fatalf("lang = %q, want en-US", option.Lang)
	}
}

func TestResolveProxyLocationOptionNormalizesCountryName(t *testing.T) {
	option := resolveProxyLocationOption("Japan", "Tokyo")
	if option.Timezone != "Asia/Tokyo" {
		t.Fatalf("timezone = %q, want Asia/Tokyo", option.Timezone)
	}
	if option.Lang != "ja-JP" {
		t.Fatalf("lang = %q, want ja-JP", option.Lang)
	}
}

func TestBuildProxyLocationResolveResultUnknownCountryFallsBackToManual(t *testing.T) {
	result := buildProxyLocationResolveResult("proxy-1", ProxyIPHealthResult{
		ProxyId: "proxy-1",
		Ok:      true,
		Country: "Unknownland",
		City:    "Nowhere",
	}, "cache", "2026-06-09T00:00:00Z")
	if result.Ok {
		t.Fatalf("expected unknown country to fail automatic resolution")
	}
	if len(result.Alternates) == 0 {
		t.Fatalf("expected manual alternates")
	}
}
