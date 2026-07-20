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
		// 清空内存实例，避免后续 ReloadConfig/绑定修复把恢复前的实例写回数据库。
		a.browserMgr.BrowserProcesses = make(map[string]*exec.Cmd)
		a.browserMgr.Profiles = make(map[string]*browser.Profile)
		a.browserMgr.Mutex.Unlock()
	}

	if a.speedScheduler != nil {
		a.speedScheduler.Stop()
		a.speedScheduler = nil
	}
}

func (a *App) backupReloadAfterMutation(migrateLegacy bool) error {
	// 必须先清空内存实例，再 ReloadConfig。否则 reconcileProfileProxyBindings
	// 会把恢复前残留在内存中的「默认实例」等 upsert 回已替换的数据库。
	if a.browserMgr != nil {
		a.browserMgr.Mutex.Lock()
		a.browserMgr.Profiles = make(map[string]*browser.Profile)
		a.browserMgr.BrowserProcesses = make(map[string]*exec.Cmd)
		a.browserMgr.Mutex.Unlock()
	}

	if err := a.ReloadConfig(); err != nil {
		return err
	}

	if a.browserMgr != nil {
		a.browserMgr.Config = a.config
	}

	if migrateLegacy {
		a.migrateToSQLite()
	}
	if a.browserMgr != nil {
		a.browserMgr.InitData()
	}
	a.loadProxies()
	a.reconcileProfileProxyBindings()

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
