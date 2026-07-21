package backend

import (
	"facade/backend/internal/config"
	"slices"
	"testing"
)

func TestBuildBrowserLaunchArgsAddsAcceptLanguageFromLang(t *testing.T) {
	profile := &BrowserProfile{
		ProfileId:       "profile-1",
		FingerprintArgs: []string{"--lang=en-SG"},
	}

	args := buildBrowserLaunchArgs(profile, "/tmp/profile-1", 9222, "direct://", nil, nil, nil, nil)

	if !slices.Contains(args, "--accept-lang=en-SG,en") {
		t.Fatalf("expected derived accept language argument, got %v", args)
	}
}

func TestAppendDerivedAcceptLanguageArgPreservesExplicitValue(t *testing.T) {
	args := []string{"--lang=en-SG", "--accept-lang=fr-FR,fr"}

	got := appendDerivedAcceptLanguageArg(args)

	if !slices.Equal(got, args) {
		t.Fatalf("expected explicit accept language to be preserved, got %v", got)
	}
}

func TestAppendDerivedAcceptLanguageArgIgnoresMalformedLang(t *testing.T) {
	args := []string{"--lang=---"}

	got := appendDerivedAcceptLanguageArg(args)

	if !slices.Equal(got, args) {
		t.Fatalf("expected malformed language to be ignored, got %v", got)
	}
}

func TestBuildBrowserLaunchArgsDropsUnsupportedFingerprintArgs(t *testing.T) {
	profile := &BrowserProfile{
		ProfileId: "profile-1",
		FingerprintArgs: []string{
			"--fingerprint-unsupported-renderer=NVIDIA",
			"--fingerprint-unsupported-device=NVIDIA GeForce RTX 3080",
		},
	}

	args := buildBrowserLaunchArgsWithPrivacy(profile, "/tmp/profile-1", 9222, "direct://", nil, nil, nil, nil, config.DefaultConfig())

	if slices.Contains(args, "--fingerprint-unsupported-renderer=NVIDIA") || slices.Contains(args, "--fingerprint-unsupported-device=NVIDIA GeForce RTX 3080") {
		t.Fatalf("unsupported fingerprint args should be dropped, got %v", args)
	}
}

func TestAppendPrivacyLaunchArgsMergesDisableFeatures(t *testing.T) {
	args := appendPrivacyLaunchArgs([]string{
		"--disable-features=Translate,WebGPU",
		"--webrtc-ip-handling-policy=default_public_interface_only",
	}, config.DefaultConfig())

	if !slices.Contains(args, "--disable-features=Translate,WebGPU,WebGPUService") {
		t.Fatalf("expected merged disable-features, got %v", args)
	}
	if !slices.Contains(args, "--webrtc-ip-handling-policy=default_public_interface_only") {
		t.Fatalf("expected explicit WebRTC policy to be preserved, got %v", args)
	}
	if slices.Contains(args, "--webrtc-ip-handling-policy=disable_non_proxied_udp") {
		t.Fatalf("explicit WebRTC policy should not be duplicated, got %v", args)
	}
	if !slices.Contains(args, "--disable-non-proxied-udp") || !slices.Contains(args, "--disable-quic") || !slices.Contains(args, "--dns-prefetch-disable") {
		t.Fatalf("expected hardened networking args, got %v", args)
	}
}
