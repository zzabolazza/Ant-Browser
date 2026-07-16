package launchcode

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"ant-chrome/backend/internal/browser"
	"ant-chrome/backend/internal/logger"
)

// BrowserStarter 浏览器启动接口（由 App 层实现并注入）
type BrowserStarter interface {
	StartInstance(profileId string) (*browser.Profile, error)
}

// BrowserStatusProvider 可选接口：提供实例运行态查询。
type BrowserStatusProvider interface {
	StatusInstance(profileId string) (*browser.Profile, error)
}

// LaunchRequestParams 支持外部自动化透传的一次性启动参数
type LaunchRequestParams struct {
	LaunchArgs           []string `json:"launchArgs"`
	StartURLs            []string `json:"startUrls"`
	SkipDefaultStartURLs bool     `json:"skipDefaultStartUrls"`
	ProxyId              string   `json:"proxyId"`
	ProxyConfig          string   `json:"proxyConfig"`
}

// LaunchRequest POST /api/launch 的请求体
type LaunchRequest struct {
	Code        string          `json:"code"`
	Key         string          `json:"key"`
	ProfileID   string          `json:"profileId"`
	ProfileName string          `json:"profileName"`
	Keyword     string          `json:"keyword"`
	Keywords    []string        `json:"keywords"`
	Tag         string          `json:"tag"`
	Tags        []string        `json:"tags"`
	GroupID     string          `json:"groupId"`
	MatchMode   string          `json:"matchMode"`
	Selector    *LaunchSelector `json:"selector"`
	LaunchRequestParams
}

// BrowserStarterWithParams 可选接口：支持带参数启动实例
type BrowserStarterWithParams interface {
	StartInstanceWithParams(profileId string, params LaunchRequestParams) (*browser.Profile, error)
}

// BrowserStopper 可选接口：支持停止运行中的实例。
type BrowserStopper interface {
	StopInstance(profileId string) (*browser.Profile, error)
}

// BrowserDebugWaiter 可选接口：等待实例调试端口进入可接管状态。
type BrowserDebugWaiter interface {
	WaitInstanceDebugReady(profileId string, debugPort int, timeout time.Duration) (*browser.Profile, bool, error)
}

// LaunchCallRecord 接口调用记录
type LaunchCallRecord struct {
	Timestamp   string              `json:"timestamp"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	ClientIP    string              `json:"clientIp"`
	Code        string              `json:"code"`
	Selector    LaunchSelector      `json:"selector,omitempty"`
	ProfileID   string              `json:"profileId"`
	ProfileName string              `json:"profileName"`
	Params      LaunchRequestParams `json:"params"`
	OK          bool                `json:"ok"`
	Status      int                 `json:"status"`
	Error       string              `json:"error"`
	DurationMs  int64               `json:"durationMs"`
}

// LaunchServer 本地 HTTP 唤起服务
type LaunchServer struct {
	service    *LaunchCodeService
	starter    BrowserStarter
	browserMgr *browser.Manager
	port       int
	server     *http.Server
	mu         sync.Mutex
	authMu   sync.RWMutex
	logMu    sync.Mutex
	callLogs []LaunchCallRecord
	apiAuth  APIAuthConfig
}

// NewLaunchServer 创建 LaunchServer
func NewLaunchServer(service *LaunchCodeService, starter BrowserStarter, mgr *browser.Manager, port int) *LaunchServer {
	srv := &LaunchServer{
		service:    service,
		starter:    starter,
		browserMgr: mgr,
		port:       port,
	}
	srv.SetAPIAuthConfig(APIAuthConfig{})
	return srv
}

// Start 非阻塞启动 HTTP 服务。
// 规则：
//   - port <= 0：自动分配随机可用端口（仅内部测试/显式传 0 时）
//   - port > 0：绑定指定固定端口；若被占用则直接返回错误
func (s *LaunchServer) Start() error {
	handler := s.buildHandler(true)

	preferredPort := s.port
	ln, port, err := bindLaunchListener(preferredPort)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.port = port
	s.server = &http.Server{Handler: handler}
	s.mu.Unlock()

	log := logger.New("LaunchServer")
	if preferredPort <= 0 {
		log.Info("LaunchServer 使用随机端口", logger.F("port", port))
	} else {
		log.Info("LaunchServer 使用固定端口", logger.F("port", port))
	}
	auth := s.apiAuthConfig()
	if auth.Active() {
		log.Info("LaunchServer API 认证已启用", logger.F("header", auth.Header))
	} else if auth.Requested() && !auth.Configured() {
		log.Warn("LaunchServer API 认证配置未生效", logger.F("reason", "api_key is empty"), logger.F("header", auth.Header))
	}
	log.Info("LaunchServer 已启动", logger.F("port", port))

	go func() {
		if serveErr := s.server.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Error("LaunchServer 异常退出", logger.F("error", serveErr.Error()))
		}
	}()

	return nil
}

func bindLaunchListener(preferredPort int) (net.Listener, int, error) {
	if preferredPort <= 0 {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, 0, fmt.Errorf("自动分配端口失败: %w", err)
		}
		port, err := listenerPort(ln)
		if err != nil {
			_ = ln.Close()
			return nil, 0, err
		}
		return ln, port, nil
	}

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(preferredPort))
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		return ln, preferredPort, nil
	}
	return nil, 0, fmt.Errorf("端口 %d 不可用: %w", preferredPort, err)
}

func listenerPort(ln net.Listener) (int, error) {
	if ln == nil {
		return 0, fmt.Errorf("listener is nil")
	}
	if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
		return tcpAddr.Port, nil
	}

	_, rawPort, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return 0, fmt.Errorf("解析监听地址失败: %w", err)
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		return 0, fmt.Errorf("解析端口失败: %w", err)
	}
	return port, nil
}

// Stop 优雅关闭（5 秒超时）
func (s *LaunchServer) Stop() error {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	if srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

// Port 返回实际绑定的端口
func (s *LaunchServer) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}
