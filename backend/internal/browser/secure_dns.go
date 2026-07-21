package browser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const secureDNSTemplate = "https://chrome.cloudflare-dns.com/dns-query"

type SecureDNSOptions struct {
	Mode      string
	Templates []string
}

// EnsureSecureDNS configures Chromium's Local State and profile preferences to use Cloudflare secure DoH.
func EnsureSecureDNS(userDataDir string) error {
	return EnsureSecureDNSWithOptions(userDataDir, SecureDNSOptions{
		Mode:      "secure",
		Templates: []string{secureDNSTemplate},
	})
}

func EnsureSecureDNSWithOptions(userDataDir string, options SecureDNSOptions) error {
	if err := os.MkdirAll(userDataDir, 0o755); err != nil {
		return fmt.Errorf("create user data directory: %w", err)
	}

	options = normalizeSecureDNSOptions(options)
	if err := writeSecureDNSJSON(filepath.Join(userDataDir, "Local State"), options); err != nil {
		return err
	}
	preferencesDir := filepath.Join(userDataDir, "Default")
	if err := os.MkdirAll(preferencesDir, 0o755); err != nil {
		return fmt.Errorf("create Default profile directory: %w", err)
	}
	if err := writeSecureDNSJSON(filepath.Join(preferencesDir, "Preferences"), options); err != nil {
		return err
	}
	return nil
}

func normalizeSecureDNSOptions(options SecureDNSOptions) SecureDNSOptions {
	mode := strings.TrimSpace(options.Mode)
	if mode == "" {
		mode = "secure"
	}
	templates := make([]string, 0, len(options.Templates))
	for _, item := range options.Templates {
		value := strings.TrimSpace(item)
		if value != "" {
			templates = append(templates, value)
		}
	}
	if len(templates) == 0 {
		templates = []string{secureDNSTemplate}
	}
	return SecureDNSOptions{Mode: mode, Templates: templates}
}

func writeSecureDNSJSON(statePath string, options SecureDNSOptions) error {
	state := make(map[string]json.RawMessage)
	fileMode := os.FileMode(0o600)

	data, err := os.ReadFile(statePath)
	if err == nil {
		info, statErr := os.Stat(statePath)
		if statErr != nil {
			return fmt.Errorf("stat Local State: %w", statErr)
		}
		fileMode = info.Mode().Perm()
		if err := json.Unmarshal(data, &state); err != nil {
			return fmt.Errorf("parse Local State JSON: %w", err)
		}
		if state == nil {
			return errors.New("parse Local State JSON: root must be an object")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read Local State: %w", err)
	}

	doh := make(map[string]json.RawMessage)
	if raw, ok := state["dns_over_https"]; ok {
		if err := json.Unmarshal(raw, &doh); err != nil || doh == nil {
			doh = make(map[string]json.RawMessage)
		}
	}
	mode, err := json.Marshal(options.Mode)
	if err != nil {
		return fmt.Errorf("encode secure DNS mode: %w", err)
	}
	doh["mode"] = mode
	template, err := json.Marshal(strings.Join(options.Templates, " "))
	if err != nil {
		return fmt.Errorf("encode secure DNS template: %w", err)
	}
	doh["templates"] = template
	encodedDoH, err := json.Marshal(doh)
	if err != nil {
		return fmt.Errorf("encode dns_over_https: %w", err)
	}
	state["dns_over_https"] = encodedDoH

	updated, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Local State JSON: %w", err)
	}
	updated = append(updated, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(statePath), ".secure-dns-*.tmp")
	if err != nil {
		return fmt.Errorf("create Local State temporary file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(fileMode); err != nil {
		tmp.Close()
		return fmt.Errorf("set Local State temporary file permissions: %w", err)
	}
	if _, err := tmp.Write(updated); err != nil {
		tmp.Close()
		return fmt.Errorf("write Local State temporary file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync Local State temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close Local State temporary file: %w", err)
	}

	if err := replaceLocalState(tmpPath, statePath); err != nil {
		return err
	}
	return nil
}

func replaceLocalState(tmpPath, statePath string) error {
	if err := os.Rename(tmpPath, statePath); err == nil {
		return nil
	} else if _, statErr := os.Stat(statePath); statErr != nil {
		return fmt.Errorf("replace Local State: %w", err)
	}

	backup, err := os.CreateTemp(filepath.Dir(statePath), ".local-state-*.backup")
	if err != nil {
		return fmt.Errorf("create Local State rollback path: %w", err)
	}
	backupPath := backup.Name()
	if err := backup.Close(); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("close Local State rollback file: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("prepare Local State rollback path: %w", err)
	}

	if err := os.Rename(statePath, backupPath); err != nil {
		return fmt.Errorf("preserve existing Local State: %w", err)
	}
	if err := os.Rename(tmpPath, statePath); err != nil {
		if rollbackErr := os.Rename(backupPath, statePath); rollbackErr != nil {
			return fmt.Errorf("replace Local State: %v; restore original Local State: %w", err, rollbackErr)
		}
		return fmt.Errorf("replace Local State: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("Local State replaced but rollback cleanup failed: %w", err)
	}
	return nil
}
