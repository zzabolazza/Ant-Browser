package backend

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"ant-chrome/backend/internal/automation"
	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/config"
	"ant-chrome/backend/internal/launchcode"
)

func lookupAutomationTestNode(t *testing.T) string {
	t.Helper()

	nodeExecPath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	return nodeExecPath
}

func prepareAutomationTestRuntime(t *testing.T, manager *automation.Manager, playwrightVersion string) {
	t.Helper()

	prepareAutomationTestRuntimeWithPlaywrightModule(
		t,
		manager,
		playwrightVersion,
		"module.exports = { chromium: {} }\n",
	)
}

func prepareAutomationTestRuntimeWithPlaywrightModule(t *testing.T, manager *automation.Manager, playwrightVersion string, playwrightModuleSource string) {
	t.Helper()

	state := manager.CurrentState()

	playwrightCoreDir := filepath.Join(state.RuntimeDir, "node_modules", "playwright-core")
	if err := os.MkdirAll(playwrightCoreDir, 0o755); err != nil {
		t.Fatalf("create playwright-core dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(playwrightCoreDir, "package.json"), []byte("{\"name\":\"playwright-core\",\"version\":\""+playwrightVersion+"\"}\n"), 0o644); err != nil {
		t.Fatalf("write playwright-core package failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(playwrightCoreDir, "index.js"), []byte(playwrightModuleSource), 0o644); err != nil {
		t.Fatalf("write playwright-core stub failed: %v", err)
	}
	if err := os.WriteFile(state.RunnerPath, []byte(automationTestRunnerScript), 0o755); err != nil {
		t.Fatalf("write runner script failed: %v", err)
	}
}

const automationTestConnectProbePlaywrightModule = `const http = require('http')

module.exports = {
  chromium: {
    connectOverCDP: async (endpoint) => {
      const target = new URL('/json/version', endpoint)
      await new Promise((resolve, reject) => {
        const req = http.get(target, (res) => {
          res.resume()
          res.on('end', () => {
            const status = res.statusCode || 0
            if (status >= 200 && status < 300) {
              resolve()
              return
            }
            reject(new Error('cdp connect probe failed with http ' + String(status)))
          })
        })
        req.on('error', reject)
      })

      return {
        contexts: () => [{
          pages: () => [],
          newPage: async () => ({})
        }],
        close: async () => {}
      }
    }
  }
}
`

const automationTestRunnerScript = `const fs = require('fs')
const path = require('path')

async function main() {
  const payloadPath = process.argv[2]
  const payload = JSON.parse(fs.readFileSync(payloadPath, 'utf8'))
  const script = require(payload.ScriptPath)
  const startedAt = new Date().toISOString()
  const result = await script.run({
    selector: payload.Selector || {},
    params: payload.Params || {},
    artifact: (name) => {
      const dir = payload.ArtifactDir || path.dirname(payload.ScriptPath)
      fs.mkdirSync(dir, { recursive: true })
      return path.join(dir, name)
    },
    log: () => {},
    launch: async () => ({ ok: true }),
    connect: async () => ({
      browser: { contexts: () => [] },
      context: {
        pages: () => [],
        newPage: async () => ({})
      },
      page: null
    })
  })

  console.log(JSON.stringify({
    ok: result && result.ok !== false,
    summary: result && result.summary ? String(result.summary) : '',
    error: result && result.error ? String(result.error) : '',
    startedAt,
    finishedAt: new Date().toISOString(),
    ...result
  }))
}

main().catch((error) => {
  console.error(error && error.stack ? error.stack : String(error))
  process.exit(1)
})
`

type automationConcurrentRunResult struct {
	run *automation.ScriptRunRecord
	err error
}

func newAutomationPlaywrightRunTestApp(t *testing.T, playwrightModuleSource string) (*App, func()) {
	t.Helper()

	nodeExecPath := lookupAutomationTestNode(t)

	app := NewApp(t.TempDir())
	app.config = config.DefaultConfig()
	app.config.Automation.Enabled = true
	app.config.Automation.NodeSource = config.AutomationNodeSourceSystem
	app.config.Automation.SystemNodePath = nodeExecPath
	app.config.Automation.NodeVersion = "test-node"
	app.config.Automation.PlaywrightCoreVersion = "1.59.0"
	app.config.Automation.RuntimeVersion = "test-runtime"
	app.browserMgr = browser.NewManager(app.config, app.appRoot)
	app.launchCodeSvc = launchcode.NewLaunchCodeService(launchcode.NewMemoryLaunchCodeDAO())
	app.browserMgr.CodeProvider = app.launchCodeSvc
	app.automationMgr = automation.NewManager(app.appRoot, app.config, nil, automation.Options{})

	prepareAutomationTestRuntimeWithPlaywrightModule(
		t,
		app.automationMgr,
		app.config.Automation.PlaywrightCoreVersion,
		playwrightModuleSource,
	)

	app.launchServer = launchcode.NewLaunchServer(
		app.launchCodeSvc,
		app,
		app.browserMgr,
		0,
	)
	if err := app.launchServer.Start(); err != nil {
		t.Fatalf("start launch server failed: %v", err)
	}

	return app, func() {
		_ = app.launchServer.Stop()
	}
}

func automationTestServerPort(t *testing.T, rawURL string) int {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse server url failed: %v", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse server port failed: %v", err)
	}
	return port
}

func createAutomationRunningProfileWithCode(t *testing.T, app *App, name string, code string, debugPort int) *browser.Profile {
	t.Helper()

	profile, err := app.browserMgr.Create(browser.ProfileInput{
		ProfileName: name,
	})
	if err != nil {
		t.Fatalf("create profile failed: %v", err)
	}
	if profile == nil {
		t.Fatal("create profile returned nil")
	}
	if strings.TrimSpace(code) != "" {
		if _, err := app.launchCodeSvc.SetCode(profile.ProfileId, code); err != nil {
			t.Fatalf("set code failed: %v", err)
		}
	}

	app.browserMgr.Profiles[profile.ProfileId].Running = true
	app.browserMgr.Profiles[profile.ProfileId].DebugReady = true
	app.browserMgr.Profiles[profile.ProfileId].DebugPort = debugPort
	app.browserMgr.Profiles[profile.ProfileId].Pid = 12345

	return profile
}

func runAutomationScriptsConcurrently(t *testing.T, count int, runner func(index int) (*automation.ScriptRunRecord, error)) []automationConcurrentRunResult {
	t.Helper()

	results := make([]automationConcurrentRunResult, count)
	start := make(chan struct{})
	var wg sync.WaitGroup

	for index := 0; index < count; index++ {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			run, err := runner(index)
			results[index] = automationConcurrentRunResult{
				run: run,
				err: err,
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	close(start)
	wg.Wait()

	return results
}
