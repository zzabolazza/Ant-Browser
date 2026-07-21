package browser

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cloudflareDoHTemplate = "https://chrome.cloudflare-dns.com/dns-query"

func TestEnsureSecureDNSCreatesLocalState(t *testing.T) {
	userDataDir := t.TempDir()

	if err := EnsureSecureDNS(userDataDir); err != nil {
		t.Fatalf("EnsureSecureDNS() error = %v", err)
	}

	state := readLocalState(t, userDataDir)
	doh := state["dns_over_https"].(map[string]any)
	if got := doh["mode"]; got != "secure" {
		t.Fatalf("mode = %v, want secure", got)
	}
	if got := doh["templates"]; got != cloudflareDoHTemplate {
		t.Fatalf("templates = %v, want %q", got, cloudflareDoHTemplate)
	}

	prefs := readProfilePreferences(t, userDataDir)
	prefsDoH := prefs["dns_over_https"].(map[string]any)
	if got := prefsDoH["mode"]; got != "secure" {
		t.Fatalf("Preferences mode = %v, want secure", got)
	}
	if got := prefsDoH["templates"]; got != cloudflareDoHTemplate {
		t.Fatalf("Preferences templates = %v, want %q", got, cloudflareDoHTemplate)
	}
}

func TestEnsureSecureDNSPreservesExistingFields(t *testing.T) {
	userDataDir := t.TempDir()
	writeLocalState(t, userDataDir, `{
		"browser": {"enabled_labs_experiments": ["example"]},
		"dns_over_https": {
			"mode": "automatic",
			"templates": "https://example.test/dns-query",
			"fallback": true
		}
	}`)

	if err := EnsureSecureDNS(userDataDir); err != nil {
		t.Fatalf("EnsureSecureDNS() error = %v", err)
	}

	state := readLocalState(t, userDataDir)
	browser := state["browser"].(map[string]any)
	experiments := browser["enabled_labs_experiments"].([]any)
	if len(experiments) != 1 || experiments[0] != "example" {
		t.Fatalf("browser field changed: %#v", browser)
	}
	doh := state["dns_over_https"].(map[string]any)
	if got := doh["fallback"]; got != true {
		t.Fatalf("dns_over_https fallback = %v, want true", got)
	}
	if got := doh["mode"]; got != "secure" {
		t.Fatalf("mode = %v, want secure", got)
	}
	if got := doh["templates"]; got != cloudflareDoHTemplate {
		t.Fatalf("templates = %v, want %q", got, cloudflareDoHTemplate)
	}
}

func TestEnsureSecureDNSCorrectsValuesAndIsIdempotent(t *testing.T) {
	userDataDir := t.TempDir()
	writeLocalState(t, userDataDir, `{"dns_over_https":{"mode":false,"templates":["wrong"]},"other":42}`)

	if err := EnsureSecureDNS(userDataDir); err != nil {
		t.Fatalf("first EnsureSecureDNS() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(userDataDir, "Local State"))
	if err != nil {
		t.Fatal(err)
	}

	if err := EnsureSecureDNS(userDataDir); err != nil {
		t.Fatalf("second EnsureSecureDNS() error = %v", err)
	}
	second, err := os.ReadFile(filepath.Join(userDataDir, "Local State"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("second run changed Local State:\nfirst:  %s\nsecond: %s", first, second)
	}
	state := readLocalState(t, userDataDir)
	if got := state["other"]; got != float64(42) {
		t.Fatalf("other = %v, want 42", got)
	}
}

func TestEnsureSecureDNSRejectsInvalidJSONWithoutChangingFile(t *testing.T) {
	userDataDir := t.TempDir()
	original := []byte(`{"dns_over_https":`)
	path := filepath.Join(userDataDir, "Local State")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	err := EnsureSecureDNS(userDataDir)
	if err == nil {
		t.Fatal("EnsureSecureDNS() error = nil, want invalid JSON error")
	}
	if !strings.Contains(err.Error(), "Local State") || !strings.Contains(err.Error(), "JSON") {
		t.Fatalf("error = %q, want clear Local State JSON error", err)
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(after, original) {
		t.Fatalf("invalid Local State changed: got %q, want %q", after, original)
	}
}

func TestEnsureSecureDNSWithOptionsUsesCustomTemplate(t *testing.T) {
	userDataDir := t.TempDir()

	err := EnsureSecureDNSWithOptions(userDataDir, SecureDNSOptions{
		Mode:      "automatic",
		Templates: []string{"https://dns.google/dns-query", "https://cloudflare-dns.com/dns-query"},
	})
	if err != nil {
		t.Fatalf("EnsureSecureDNSWithOptions() error = %v", err)
	}

	state := readLocalState(t, userDataDir)
	doh := state["dns_over_https"].(map[string]any)
	if got := doh["mode"]; got != "automatic" {
		t.Fatalf("mode = %v, want automatic", got)
	}
	if got := doh["templates"]; got != "https://dns.google/dns-query https://cloudflare-dns.com/dns-query" {
		t.Fatalf("templates = %v", got)
	}
}

func writeLocalState(t *testing.T, userDataDir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(userDataDir, "Local State"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readLocalState(t *testing.T, userDataDir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(userDataDir, "Local State"))
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("Local State is invalid JSON: %v", err)
	}
	return state
}

func readProfilePreferences(t *testing.T, userDataDir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(userDataDir, "Default", "Preferences"))
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("Preferences is invalid JSON: %v", err)
	}
	return state
}
