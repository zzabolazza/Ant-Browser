package proxy

import (
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestParseDnsConfigFromClashYAML(t *testing.T) {
	t.Parallel()

	raw := `
dns:
  enable: true
  nameserver:
    - 8.8.8.8
    - tls://1.1.1.1
  fallback:
    - https://dns.google/dns-query
`

	got := parseDnsConfig(raw)
	want := map[string]interface{}{
		"servers": []interface{}{"8.8.8.8", "https://dns.google/dns-query"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDnsConfig() = %#v, want %#v", got, want)
	}
}

func TestBuildDnsDiagnosticSummaryFromClashYAML(t *testing.T) {
	t.Parallel()

	raw := `
dns:
  enable: true
  enhanced-mode: fake-ip
  nameserver:
    - 8.8.8.8
    - tls://1.1.1.1
  fallback:
    - https://dns.google/dns-query
`

	got := buildDnsDiagnosticSummary(raw)
	if !got.HasConfig || got.SourceFormat != "clash" || got.EnhancedMode != "fake-ip" {
		t.Fatalf("unexpected summary identity: %+v", got)
	}
	if got.NameserverCount != 2 || got.FallbackCount != 1 || got.XrayServerCount != 2 {
		t.Fatalf("unexpected summary counts: %+v", got)
	}
	if len(got.Unsupported) != 1 || got.Unsupported[0] != "tls://1.1.1.1" {
		t.Fatalf("unsupported = %#v, want tls://1.1.1.1", got.Unsupported)
	}
}

func TestBuildDnsDiagnosticSummaryFromCommaList(t *testing.T) {
	t.Parallel()

	got := buildDnsDiagnosticSummary("8.8.8.8, tls://1.1.1.1, https://dns.google/dns-query")
	if !got.HasConfig || got.SourceFormat != "list" {
		t.Fatalf("unexpected summary identity: %+v", got)
	}
	if got.NameserverCount != 3 || got.XrayServerCount != 2 || len(got.Unsupported) != 1 {
		t.Fatalf("unexpected summary counts: %+v", got)
	}
}

func TestParseDnsConfigFromCommaList(t *testing.T) {
	t.Parallel()

	got := parseDnsConfig("8.8.8.8, tls://1.1.1.1, 127.0.0.1:53")
	want := map[string]interface{}{
		"servers": []interface{}{"8.8.8.8", "127.0.0.1:53"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDnsConfig() = %#v, want %#v", got, want)
	}
}

func TestNormalizeNodeScheme(t *testing.T) {
	t.Parallel()

	if got := normalizeNodeScheme("hysteria://example"); got != "hysteria2://example" {
		t.Fatalf("normalizeNodeScheme() = %q", got)
	}
	if got := normalizeNodeScheme("vmess://example"); got != "vmess://example" {
		t.Fatalf("normalizeNodeScheme() unexpectedly changed vmess: %q", got)
	}
}

func TestWaitPortReadyIncludesLastDialError(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	err = waitPortReady("127.0.0.1", port, 50*time.Millisecond)
	if err == nil {
		t.Fatalf("expected waitPortReady error")
	}
	if got := err.Error(); got == "" || got == "端口 0 不可用" {
		t.Fatalf("expected detailed port error, got %q", got)
	}
}

func TestWaitSocks5ReadyRequiresHandshake(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 3)
		_, _ = io.ReadFull(conn, buf)
		_, _ = conn.Write([]byte{0x05, 0x00})
	}()

	if err := waitSocks5Ready("127.0.0.1", port, time.Second); err != nil {
		t.Fatalf("waitSocks5Ready() error = %v", err)
	}
	<-done
}
