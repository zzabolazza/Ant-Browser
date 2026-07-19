package backend

import (
	"facade/backend/internal/browser"
	"facade/backend/internal/config"
	"facade/backend/internal/logger"
)

type BrowserProfile = browser.Profile
type BrowserProfileInput = browser.ProfileInput
type BrowserTab = browser.Tab
type BrowserSettings = browser.Settings
type BrowserProxy = browser.Proxy
type BrowserCore = browser.Core
type BrowserCoreInput = browser.CoreInput
type BrowserCoreValidateResult = browser.CoreValidateResult
type BrowserCoreExtendedInfo = browser.CoreExtendedInfo

type BrowserCorePickResult struct {
	CorePath      string `json:"corePath"`
	SuggestedName string `json:"suggestedName"`
}
type BrowserProfileCopyOptions = browser.ProfileCopyOptions

// BrowserProfileList 获取所有实例列表
func (a *App) BrowserProfileList() []BrowserProfile { return a.browserMgr.List() }

// BrowserProfileListByTag 按标签筛选实例列表
func (a *App) BrowserProfileListByTag(tag string) []BrowserProfile {
	return a.browserMgr.ListByTag(tag)
}

// BrowserGetAllTags 获取所有已使用的标签
func (a *App) BrowserGetAllTags() []string {
	return a.browserMgr.GetAllTags()
}

// BrowserProfileSetKeywords 设置实例关键字
func (a *App) BrowserProfileSetKeywords(profileId string, keywords []string) (*BrowserProfile, error) {
	return a.browserMgr.SetKeywords(profileId, keywords)
}

func (a *App) BrowserProfileCreate(input BrowserProfileInput) (*BrowserProfile, error) {
	return a.browserMgr.Create(input)
}

func (a *App) BrowserProfileUpdate(profileId string, input BrowserProfileInput) (*BrowserProfile, error) {
	return a.browserMgr.Update(profileId, input)
}

func (a *App) BrowserProfileDelete(profileId string) error { return a.browserMgr.Delete(profileId) }

// BrowserProfileCopy 复制实例配置（除指纹参数外全部复制）
func (a *App) BrowserProfileCopy(profileId string, newName string) (*BrowserProfile, error) {
	return a.browserMgr.Copy(profileId, newName)
}

// BrowserProfileCopyWithMode 按模式复制实例配置。
func (a *App) BrowserProfileCopyWithMode(profileId string, newName string, mode string) (*BrowserProfile, error) {
	return a.browserMgr.CopyWithMode(profileId, newName, mode)
}

// BrowserProfileCopyWithOptions 按结构化选项复制实例配置。
func (a *App) BrowserProfileCopyWithOptions(profileId string, newName string, options BrowserProfileCopyOptions) (*BrowserProfile, error) {
	return a.browserMgr.CopyWithOptions(profileId, newName, options)
}

// migrateToSQLite 一次性迁移：若 SQLite 表为空则从旧文件导入数据，或初始化默认数据
// 迁移顺序：cores → proxies → profiles → bookmarks
func (a *App) migrateToSQLite() {
	log := logger.New("Migration")

	if cores, err := a.browserMgr.CoreDAO.List(); err == nil && len(cores) == 0 {
		if len(a.config.Browser.Cores) > 0 {
			for _, c := range a.config.Browser.Cores {
				if err := a.browserMgr.CoreDAO.Upsert(c); err != nil {
					log.Error("内核迁移失败", logger.F("core_id", c.CoreId), logger.F("error", err))
				}
			}
			log.Info("内核数据已迁移", logger.F("count", len(a.config.Browser.Cores)))
		} else {
			log.Info("内核表为空，将通过自动检测初始化")
		}
	}

	if proxies, err := a.browserMgr.ProxyDAO.List(); err == nil && len(proxies) == 0 {
		var srcProxies []browser.Proxy
		if loaded, err := config.LoadProxies(a.resolveAppPath("proxies.yaml")); err == nil && len(loaded) > 0 {
			srcProxies = loaded
		} else if len(a.config.Browser.Proxies) > 0 {
			srcProxies = a.config.Browser.Proxies
		} else {
			srcProxies = []browser.Proxy{
				{ProxyId: "__direct__", ProxyName: "直连（不走代理）", ProxyConfig: "direct://"},
			}
			log.Info("代理表为空，初始化默认代理")
		}
		for _, p := range srcProxies {
			if err := a.browserMgr.ProxyDAO.Upsert(p); err != nil {
				log.Error("代理迁移失败", logger.F("proxy_id", p.ProxyId), logger.F("error", err))
			}
		}
		if len(srcProxies) > 0 {
			log.Info("代理数据已初始化", logger.F("count", len(srcProxies)))
		}
	}

	if bookmarks, err := a.browserMgr.BookmarkDAO.List(); err == nil && len(bookmarks) == 0 {
		src := a.config.Browser.DefaultBookmarks
		if len(src) > 0 {
			if err := a.browserMgr.BookmarkDAO.ReplaceAll(src); err != nil {
				log.Error("书签迁移失败", logger.F("error", err))
			} else {
				log.Info("书签数据已迁移", logger.F("count", len(src)))
			}
		}
	}
}
