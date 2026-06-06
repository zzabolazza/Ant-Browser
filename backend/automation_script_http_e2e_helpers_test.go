package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strings"
	"testing"
	"time"
)

func lookupAutomationHTTPProbeNode(t *testing.T) string {
	t.Helper()

	const preferred = `D:\code\plugin\nodejs\node.exe`
	if _, err := os.Stat(preferred); err == nil {
		return preferred
	}
	return lookupAutomationTestNode(t)
}

func lookupAutomationHTTPProbeChrome(t *testing.T) string {
	t.Helper()

	const preferred = `C:\Program Files\Google\Chrome\Application\chrome.exe`
	if _, err := os.Stat(preferred); err == nil {
		return preferred
	}
	t.Skip("system chrome is not installed")
	return ""
}

func automationHTTPRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		t.Fatal("resolve repo root failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

func automationHTTPFreePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate port failed: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func prepareAutomationHTTPRuntime(appRoot string, repoRoot string, runtimeVersion string) error {
	repoRuntimeDir := filepath.Join(repoRoot, "data", "runtime", "automation", strings.TrimSpace(runtimeVersion))
	tempRuntimeDir := filepath.Join(appRoot, "data", "runtime", "automation", strings.TrimSpace(runtimeVersion))
	if _, err := os.Stat(repoRuntimeDir); err != nil {
		return fmt.Errorf("repo runtime not found: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(tempRuntimeDir, "node_modules"), 0o755); err != nil {
		return err
	}
	if err := automationHTTPCopyFile(
		filepath.Join(repoRuntimeDir, "runner.cjs"),
		filepath.Join(tempRuntimeDir, "runner.cjs"),
	); err != nil {
		return err
	}
	return automationHTTPCopyDir(
		filepath.Join(repoRuntimeDir, "node_modules", "playwright-core"),
		filepath.Join(tempRuntimeDir, "node_modules", "playwright-core"),
	)
}

func automationHTTPCopyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func automationHTTPCopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relativePath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func automationHTTPRequestJSON(method string, url string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned %d: %s", method, url, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode %s %s failed: %w; body=%s", method, url, err, string(raw))
	}
	return nil
}

func automationHTTPHasScript(items []struct {
	ID string `json:"id"`
}, scriptID string) bool {
	for _, item := range items {
		if strings.TrimSpace(item.ID) == scriptID {
			return true
		}
	}
	return false
}

func automationHTTPStringValue(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, _ := payload[key].(string)
	return strings.TrimSpace(value)
}

func automationHTTPMarshal(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal debug payload failed: %v", err)
	}
	return string(data)
}

const automationHTTPMailFixtureHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Mail Fixture</title>
  <script>
    window.__notificationProbe = {
      supported: typeof Notification !== 'undefined',
      requested: false,
      result: '',
      error: '',
    };
    document.addEventListener('DOMContentLoaded', () => {
      if (typeof Notification === 'undefined' || typeof Notification.requestPermission !== 'function') {
        document.documentElement.setAttribute('data-notification-probe', 'unsupported');
        return;
      }
      window.__notificationProbe.requested = true;
      Notification.requestPermission()
        .then((result) => {
          window.__notificationProbe.result = String(result || '');
          document.documentElement.setAttribute('data-notification-probe', window.__notificationProbe.result || 'empty');
        })
        .catch((error) => {
          window.__notificationProbe.error = String(error && error.message ? error.message : error);
          document.documentElement.setAttribute('data-notification-probe', 'error');
        });
    });
  </script>
  <style>
    body { margin: 0; font-family: Arial, sans-serif; background: #f6f7fb; }
    main { display: flex; gap: 20px; padding: 24px; min-height: 100vh; box-sizing: border-box; }
    .sidebar { width: 32%; min-width: 320px; background: #fff; border: 1px solid #d9dce6; border-radius: 12px; padding: 20px; box-sizing: border-box; }
    .viewer { width: 60%; min-height: 420px; background: #fff; border: 1px solid #d9dce6; border-radius: 12px; padding: 24px; box-sizing: border-box; }
    input { width: 100%; height: 42px; padding: 0 12px; font-size: 16px; box-sizing: border-box; }
    [role="row"] { margin-top: 16px; min-height: 56px; border: 1px solid #c8cfdd; border-radius: 10px; padding: 16px; cursor: pointer; background: #fafbff; }
    p { margin: 0 0 12px; line-height: 1.55; }
    h1 { margin: 0 0 16px; font-size: 28px; }
  </style>
</head>
<body>
  <main>
    <section class="sidebar">
      <div role="dialog" tabindex="-1" data-focus-root="1" class="overlay no-outline" data-testid="overlay-button" id="advanced-search-overlay-14">
        <input
          type="search"
          readonly
          title="关键词"
          placeholder="搜索邮件"
          value=""
          aria-label="Search messages"
          data-testid="search-keyword"
          class="input-element w-full cursor-text"
        />
      </div>
      <div role="row">target@example.com ChatGPT verification code 429792</div>
    </section>
    <article role="article" class="viewer">
      <h1>Your ChatGPT verification code</h1>
      <p>From: ChatGPT &lt;noreply@tm.openai.com&gt;</p>
      <p>To: target@example.com</p>
      <p>Hello,</p>
      <p>Your verification code is 429792.</p>
      <p>Please use this code to continue signing in.</p>
      <p>Best regards</p>
      <p>ChatGPT</p>
    </article>
  </main>
</body>
</html>`

const automationHTTPMailProbeScriptTextRaw = `module.exports.run = async ({ launch, connect, openPage, selector, params = {} }) => {
  const normalizeText = (value) => String(value == null ? '' : value).trim()
  const timeoutMs = Number.isFinite(Number(params.timeoutMs))
    ? Math.max(5000, Math.round(Number(params.timeoutMs)))
    : 45000
  const inboxUrl = normalizeText(params.inboxUrl)

  if (!inboxUrl) {
    throw new Error('inboxUrl is required')
  }

  const session = await launch({
    selector,
    skipDefaultStartUrls: true,
    startUrls: [inboxUrl],
  })
  const connection = await connect(session, { timeoutMs })
  const browser = connection.browser
  if (!browser) {
    throw new Error('browser connection is unavailable')
  }

  const context =
    connection.context ||
    browser.contexts()[0] ||
    (typeof browser.newContext === 'function' ? await browser.newContext() : null)
  if (!context) {
    throw new Error('browser context is unavailable')
  }

  const opened = await openPage(connection, {
    url: inboxUrl,
    timeoutMs,
    permissions: ['notifications'],
  })
  const page = opened.page
  await page.waitForLoadState('networkidle', {
    timeout: Math.min(timeoutMs, 2500),
  }).catch(() => {})

  const result = await page.evaluate(() => {
    const normalizeText = (value) => String(value == null ? '' : value).replace(/\s+/g, ' ').trim()
    const article = document.querySelector('article')
    const lines = Array.from(document.querySelectorAll('article p'))
      .map((node) => normalizeText(node.textContent))
      .filter(Boolean)
    const subject = normalizeText(document.querySelector('article h1')?.textContent)
    const fromLine = lines.find((line) => line.startsWith('From:')) || ''
    const toLine = lines.find((line) => line.startsWith('To:')) || ''
    const articleText = normalizeText(article?.textContent)
    const mailboxMatch = fromLine.match(/^From:\s*([^<]+?)\s*</)
    const senderEmailMatch = fromLine.match(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i)
    const recipientEmailMatch = toLine.match(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/i)
    const verificationCodeMatch = articleText.match(/\b\d{6}\b/)
    const signature = lines.slice(-2).join('\n')

    return {
      notificationPermission: typeof Notification !== 'undefined' ? Notification.permission : '',
      notificationProbe: document.documentElement.getAttribute('data-notification-probe') || '',
      mailboxName: mailboxMatch ? normalizeText(mailboxMatch[1]) : '',
      senderEmail: senderEmailMatch ? senderEmailMatch[0] : '',
      recipientEmail: recipientEmailMatch ? recipientEmailMatch[0] : '',
      subject,
      verificationCode: verificationCodeMatch ? verificationCodeMatch[0] : '',
      signature,
    }
  })

  return {
    ok: true,
    permissionApplied: opened.permissionResult && opened.permissionResult.applied === true,
    permissionOrigin: opened.permissionResult && opened.permissionResult.origin ? opened.permissionResult.origin : '',
    summary: '已提取测试邮件内容',
    ...result,
  }
}`

var automationHTTPMailProbeScriptSummaryLine = regexp.MustCompile(`summary:[^\n]+`)

var automationHTTPMailProbeScriptText = automationHTTPMailProbeScriptSummaryLine.ReplaceAllString(
	automationHTTPMailProbeScriptTextRaw,
	"summary: 'mail probe extracted message',",
)
