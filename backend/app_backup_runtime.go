package backend

import (
	"facade/backend/internal/browser"
	"os/exec"
	"time"
)

func (a *App) backupStopRuntimeForMaintenance() {
	if a.browserMgr != nil {
		a.browserMgr.Mutex.Lock()
		for _, cmd := range a.browserMgr.BrowserProcesses {
			if cmd != nil && cmd.Process != nil {
				_ = a.stopProcessCmd(cmd)
			}
		}
		a.browserMgr.BrowserProcesses = make(map[string]*exec.Cmd)
		a.browserMgr.Mutex.Unlock()
	}

	if a.speedScheduler != nil {
		a.speedScheduler.Stop()
		a.speedScheduler = nil
	}
}

func (a *App) backupReloadAfterMutation() error {
	if err := a.ReloadConfig(); err != nil {
		return err
	}

	if a.browserMgr != nil {
		a.browserMgr.Config = a.config
		a.browserMgr.Mutex.Lock()
		a.browserMgr.Profiles = make(map[string]*browser.Profile)
		a.browserMgr.BrowserProcesses = make(map[string]*exec.Cmd)
		a.browserMgr.Mutex.Unlock()
	}

	a.migrateToSQLite()
	if a.browserMgr != nil {
		a.browserMgr.InitData()
	}
	a.loadProxies()

	if a.launchCodeSvc != nil {
		_ = a.launchCodeSvc.LoadAll()
	}
	if a.browserMgr != nil {
		a.browserMgr.CodeProvider = a.launchCodeSvc
	}

	if a.browserMgr != nil && a.browserMgr.ProxyDAO != nil {
		a.speedScheduler = browser.NewProxySpeedScheduler(
			a.browserMgr.ProxyDAO,
			func(proxyID string) (bool, int64, string) {
				r := a.testProxySpeed(proxyID, a.getLatestProxies())
				return r.Ok, r.LatencyMs, r.Error
			},
			5*time.Minute,
			5,
		)
		a.speedScheduler.Start()
	}
	return nil
}
