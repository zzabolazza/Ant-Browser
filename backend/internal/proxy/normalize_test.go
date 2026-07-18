package proxy

import (
	"facade/backend/internal/config"
	"testing"
)

func TestNormalizeBrowserProxiesTrimsAndAddsBuiltin(t *testing.T) {
	proxies := NormalizeBrowserProxies([]config.BrowserProxy{
		{ProxyName: "  main  ", ProxyConfig: " http://127.0.0.1:8080 ", GroupName: " group-a "},
		{ProxyName: "missing config"},
	}, func() string { return "generated-id" })

	if len(proxies) != 2 {
		t.Fatalf("len = %d, want 2", len(proxies))
	}
	if proxies[0].ProxyId != "__direct__" {
		t.Fatalf("first proxy id = %q, want builtin direct", proxies[0].ProxyId)
	}
	if proxies[1].ProxyId != "generated-id" {
		t.Fatalf("generated proxy id = %q", proxies[1].ProxyId)
	}
	if proxies[1].ProxyName != "main" || proxies[1].ProxyConfig != "http://127.0.0.1:8080" {
		t.Fatalf("proxy was not trimmed: %#v", proxies[1])
	}
	if proxies[1].GroupName != "group-a" {
		t.Fatalf("metadata was not trimmed: %#v", proxies[1])
	}
}

func TestNormalizeBrowserProxiesKeepsExistingBuiltin(t *testing.T) {
	proxies := NormalizeBrowserProxies([]config.BrowserProxy{
		{ProxyId: "__direct__", ProxyName: "direct", ProxyConfig: "direct://"},
	}, nil)

	if len(proxies) != 1 {
		t.Fatalf("len = %d, want 1", len(proxies))
	}
}
