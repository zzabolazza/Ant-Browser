package proxy

import (
	"testing"

	"facade/backend/internal/config"
)

func TestParseNativeProxyURL(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "direct", input: "direct://", want: "direct://"},
		{name: "http", input: "http://1.2.3.4:8080", want: "http://1.2.3.4:8080"},
		{name: "https auth", input: "https://user:pass@host:8443", want: "https://user:pass@host:8443"},
		{name: "socks5", input: "socks5://127.0.0.1:1080", want: "socks5://127.0.0.1:1080"},
		{name: "vmess rejected", input: "vmess://abc", wantErr: true},
		{name: "missing port", input: "http://1.2.3.4", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseNativeProxyURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestValidateProxyConfig(t *testing.T) {
	proxies := []config.BrowserProxy{
		{ProxyId: "p1", ProxyName: "ok", ProxyConfig: "socks5://127.0.0.1:1080"},
	}
	ok, msg := ValidateProxyConfig("", proxies, "p1")
	if !ok || msg != "" {
		t.Fatalf("expected ok, got %v %q", ok, msg)
	}
	ok, msg = ValidateProxyConfig("vmess://x", nil, "")
	if ok {
		t.Fatalf("expected reject, got ok msg=%q", msg)
	}
}
