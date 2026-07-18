package backend

import (
	"facade/backend/internal/apppath"
	"facade/backend/internal/browser"
	"facade/backend/internal/config"
	"facade/backend/internal/logger"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// 工具函数
// ============================================================================

// resolveAppPath 将相对路径解析为绝对路径（基于 appRoot）。
// 如果传入的已经是绝对路径则直接返回。
func (a *App) resolveAppPath(p string) string {
	return apppath.Resolve(a.appRoot, p)
}

func generateUUID() string {
	return uuid.NewString()
}

func nextAvailablePort() (int, error) {
	// 二次验证策略：分配端口后立即再次绑定确认未被抢占，最多重试 10 次
	for i := 0; i < 10; i++ {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			continue
		}
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()
		// 短暂等待 OS 释放端口
		time.Sleep(5 * time.Millisecond)
		// 二次验证端口未被其他进程抢占
		v, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			continue
		}
		v.Close()
		return port, nil
	}
	return 0, fmt.Errorf("无法分配可用端口")
}

// ============================================================================
// 代理数据加载
// ============================================================================

// loadProxies 启动时加载代理数据。
// 优先从 ProxyDAO（SQLite）读取；若 DAO 未注入则降级到 proxies.yaml，最后降级到 config.yaml。
func (a *App) loadProxies() {
	log := logger.New("Browser")

	builtins := []browser.Proxy{
		{ProxyId: "__direct__", ProxyName: "直连（不走代理）", ProxyConfig: "direct://"},
	}

	ensureBuiltins := func(list []browser.Proxy) []browser.Proxy {
		for _, b := range builtins {
			found := false
			for _, p := range list {
				if p.ProxyId == b.ProxyId {
					found = true
					break
				}
			}
			if !found {
				list = append([]browser.Proxy{b}, list...)
			}
		}
		return list
	}

	// 优先从 SQLite 读取
	if a.browserMgr.ProxyDAO != nil {
		list, err := a.browserMgr.ProxyDAO.List()
		if err != nil {
			log.Error("从数据库读取代理失败", logger.F("error", err.Error()))
		} else if len(list) > 0 {
			a.config.Browser.Proxies = list
			log.Info("代理数据从数据库加载完成", logger.F("count", len(list)))
			return
		}
	}

	// 降级：从 proxies.yaml 加载
	loaded, err := config.LoadProxies(a.resolveAppPath("proxies.yaml"))
	if err != nil {
		log.Warn("读取 proxies.yaml 失败", logger.F("error", err.Error()))
	}
	if loaded != nil {
		proxies := ensureBuiltins(loaded)
		a.config.Browser.Proxies = proxies
		log.Info("代理数据从 proxies.yaml 加载完成", logger.F("count", len(proxies)))
		return
	}

	// 最终降级：使用 config.yaml 中的数据
	proxies := ensureBuiltins(a.config.Browser.Proxies)
	a.config.Browser.Proxies = proxies
	log.Info("代理数据使用 config.yaml 默认值", logger.F("count", len(proxies)))
}
