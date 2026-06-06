package backend

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"
)

func mustListenLoopback(t *testing.T) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听测试端口失败: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	return ln
}

func listenerPort(t *testing.T, ln net.Listener) int {
	t.Helper()

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("解析监听地址失败: %T", ln.Addr())
	}
	return tcpAddr.Port
}

func shortLivedCommand() *exec.Cmd {
	if goruntime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "exit", "0")
	}
	return exec.Command("sh", "-c", "exit 0")
}

func longLivedCommand(duration time.Duration) *exec.Cmd {
	if goruntime.GOOS == "windows" {
		seconds := int(duration / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		return exec.Command("cmd", "/c", fmt.Sprintf("ping -n %d 127.0.0.1 >nul", seconds+1))
	}
	return exec.Command("sh", "-c", fmt.Sprintf("sleep %.1f", duration.Seconds()))
}

func stderrFailingCommand(message string) *exec.Cmd {
	if goruntime.GOOS == "windows" {
		return exec.Command("cmd", "/c", fmt.Sprintf("echo %s 1>&2 & exit 5", message))
	}
	return exec.Command("sh", "-c", fmt.Sprintf("echo '%s' 1>&2; exit 5", message))
}

func stderrPortCommand(port int, holdFor time.Duration) *exec.Cmd {
	if goruntime.GOOS == "windows" {
		seconds := int(holdFor / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		// ping -n N waits roughly N-1 seconds on Windows.
		return exec.Command("cmd", "/c", fmt.Sprintf("echo DevTools listening on ws://127.0.0.1:%d/devtools/browser/test 1>&2 & ping -n %d 127.0.0.1 >nul", port, seconds+1))
	}
	return exec.Command("sh", "-c", fmt.Sprintf("echo 'DevTools listening on ws://127.0.0.1:%d/devtools/browser/test' 1>&2; sleep %.1f", port, holdFor.Seconds()))
}

func waitForCondition(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("等待条件成立超时")
}

func freeLoopbackPort(t *testing.T) int {
	t.Helper()

	ln := mustListenLoopback(t)
	port := listenerPort(t, ln)
	_ = ln.Close()
	return port
}

type devToolsTestServer struct {
	port   int
	server *http.Server
	done   chan struct{}
}

func startDevToolsServer(t *testing.T, handler http.Handler) *devToolsTestServer {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("启动 DevTools 测试服务失败: %v", err)
	}

	srv := &http.Server{Handler: handler}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = srv.Serve(ln)
	}()

	return &devToolsTestServer{
		port:   listenerPort(t, ln),
		server: srv,
		done:   done,
	}
}

func startDevToolsServerOnPort(t *testing.T, port int, handler http.Handler) *devToolsTestServer {
	t.Helper()

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("在指定端口启动 DevTools 测试服务失败: %v", err)
	}

	srv := &http.Server{Handler: handler}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = srv.Serve(ln)
	}()

	return &devToolsTestServer{
		port:   port,
		server: srv,
		done:   done,
	}
}

func (s *devToolsTestServer) Close() error {
	if s == nil || s.server == nil {
		return nil
	}
	err := s.server.Close()
	<-s.done
	return err
}

func writeDevToolsActivePortFile(t *testing.T, userDataDir string, port int) {
	t.Helper()

	content := fmt.Sprintf("%d\n/devtools/browser/test\n", port)
	if err := os.WriteFile(filepath.Join(userDataDir, "DevToolsActivePort"), []byte(content), 0644); err != nil {
		t.Fatalf("写入 DevToolsActivePort 失败: %v", err)
	}
}
