import { Suspense, useEffect, useState } from "react";
import { BrowserRouter as Router } from "react-router-dom";
import { Layout } from "./shared/layout";
import { ToastContainer, Modal, Button, Loading } from "./shared/components";
import { AlertCircle } from "lucide-react";
import { AppRoutes } from "./routes/AppRoutes";
import { lazyNamed } from "./routes/lazyNamed";
import { useNotificationStore } from "./store/notificationStore";
import { useBackupStore } from "./store/backupStore";
import { installWailsOperationLogger } from "./utils/wailsOperationLogger";
import {
  ForceQuit as ForceQuitApp,
  QuitAppOnly as QuitAppOnlyApp,
} from "./wailsjs/go/main/App";
import {
  Environment,
  Quit,
  WindowHide,
  WindowMinimise,
} from "./wailsjs/runtime/runtime";

const QuickLaunchModal = lazyNamed(
  () => import("./modules/browser/components/QuickLaunchModal"),
  "QuickLaunchModal",
);

function useWailsNotifications() {
  const addNotification = useNotificationStore((s) => s.addNotification);

  useEffect(() => {
    const runtime = (window as any).runtime;
    if (!runtime?.EventsOn) return;

    const offCrashed = runtime.EventsOn(
      "browser:instance:crashed",
      (data: { profileId: string; profileName: string; error: string }) => {
        addNotification({
          type: "error",
          title: "实例异常退出",
          message: `「${data.profileName || data.profileId}」意外崩溃：${data.error}`,
        });
      },
    );

    return () => {
      offCrashed?.();
    };
  }, [addNotification]);
}

function useGlobalErrorNotifications() {
  const addNotification = useNotificationStore((s) => s.addNotification);

  useEffect(() => {
    const toMessage = (value: unknown) => {
      if (value instanceof Error) return value.message || String(value);
      if (typeof value === "string") return value;
      try {
        return JSON.stringify(value);
      } catch {
        return String(value);
      }
    };

    const handleError = (event: ErrorEvent) => {
      addNotification({
        type: "error",
        title: "前端异常",
        message: event.message || toMessage(event.error) || "未知脚本错误",
      });
    };

    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      addNotification({
        type: "error",
        title: "未处理异步异常",
        message: toMessage(event.reason) || "未知 Promise 异常",
      });
    };

    window.addEventListener("error", handleError);
    window.addEventListener("unhandledrejection", handleUnhandledRejection);
    return () => {
      window.removeEventListener("error", handleError);
      window.removeEventListener("unhandledrejection", handleUnhandledRejection);
    };
  }, [addNotification]);
}

function CloseConfirmModal() {
  const [open, setOpen] = useState(false);
  const [platform, setPlatform] = useState("windows");
  const [quittingAction, setQuittingAction] = useState<
    "app-only" | "app-and-browser" | null
  >(null);
  const importInProgress = useBackupStore((s) => s.importInProgress);
  const importProgress = useBackupStore((s) => s.importProgress);
  const importMessage = useBackupStore((s) => s.importMessage);
  const supportsTray = platform === "windows";
  const quitting = quittingAction !== null;

  useEffect(() => {
    const runtime = (window as any).runtime;
    if (!runtime?.EventsOn) return;

    const off = runtime.EventsOn("app:request-close", () => {
      setQuittingAction(null);
      setOpen(true);
    });
    return () => {
      if (typeof off === "function") off();
    };
  }, []);

  useEffect(() => {
    let cancelled = false;

    Environment()
      .then((info) => {
        if (!cancelled && info?.platform) {
          setPlatform(info.platform);
        }
      })
      .catch(() => {});

    return () => {
      cancelled = true;
    };
  }, []);

  const closeModal = () => {
    if (quitting) return;
    setOpen(false);
  };

  const handleMinimize = () => {
    if (quitting) return;
    setOpen(false);
    if (supportsTray) {
      WindowHide();
      return;
    }
    WindowMinimise();
  };

  const handleQuitAppOnly = async () => {
    setQuittingAction("app-only");
    try {
      await QuitAppOnlyApp();
    } catch (error) {
      console.error("QuitAppOnly failed", error);
      setQuittingAction(null);
    }
  };

  const handleQuitAppAndBrowsers = async () => {
    setQuittingAction("app-and-browser");
    try {
      await Promise.race([
        ForceQuitApp(),
        new Promise((resolve) => setTimeout(resolve, 1200)),
      ]);
    } catch (error) {
      console.error("ForceQuit failed, falling back to runtime.Quit()", error);
    }
    Quit();
  };

  return (
    <Modal
      open={open}
      onClose={closeModal}
      title={importInProgress ? "关闭应用确认" : "退出确认"}
      width="400px"
      closable={!quitting}
    >
      {importInProgress ? (
        <div className="flex flex-col items-center py-2">
          <div className="w-12 h-12 rounded-full bg-amber-50 text-amber-500 flex items-center justify-center mb-4">
            <AlertCircle className="w-6 h-6" />
          </div>
          <p className="text-sm text-[var(--color-text-secondary)] text-center mb-6">
            当前正在加载配置
            {importProgress > 0 ? `（${importProgress}%）` : ""}，
            <br />
            {importMessage || "强制关闭会中断本次加载，是否仍要关闭应用？"}
          </p>
          <div className="flex gap-3 w-full">
            <Button
              variant="secondary"
              className="flex-1"
              onClick={closeModal}
              disabled={quitting}
            >
              继续加载
            </Button>
            <Button
              variant="danger"
              className="flex-1"
              onClick={handleQuitAppAndBrowsers}
              loading={quittingAction === "app-and-browser"}
            >
              仍要关闭
            </Button>
          </div>
        </div>
      ) : (
        <div className="flex flex-col gap-2 -mt-1">
          <p className="text-sm text-[var(--color-text-secondary)] mb-2">
            选择退出方式：
          </p>

          {supportsTray && (
            <button
              onClick={handleMinimize}
              disabled={quitting}
              className="flex items-center gap-3 w-full px-4 py-3 rounded-lg border border-[var(--color-border)] hover:bg-[var(--color-bg-muted)] transition-colors text-left disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <div className="w-8 h-8 rounded-md bg-blue-50 text-blue-500 flex items-center justify-center shrink-0">
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="2" y="3" width="20" height="14" rx="2" />
                  <line x1="8" y1="21" x2="16" y2="21" />
                  <line x1="12" y1="17" x2="12" y2="21" />
                </svg>
              </div>
              <div className="min-w-0">
                <div className="text-sm font-medium text-[var(--color-text-primary)]">最小化至托盘</div>
                <div className="text-xs text-[var(--color-text-muted)] mt-0.5">隐藏窗口，应用在后台继续运行</div>
              </div>
            </button>
          )}

          <button
            onClick={handleQuitAppOnly}
            disabled={quitting}
            className="flex items-center gap-3 w-full px-4 py-3 rounded-lg border border-[var(--color-border)] hover:bg-[var(--color-bg-muted)] transition-colors text-left disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <div className="w-8 h-8 rounded-md bg-amber-50 text-amber-500 flex items-center justify-center shrink-0">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
            </div>
            <div className="min-w-0">
              <div className="text-sm font-medium text-[var(--color-text-primary)]">仅退出客户端</div>
              <div className="text-xs text-[var(--color-text-muted)] mt-0.5">关闭客户端，保留已打开的浏览器</div>
            </div>
            {quittingAction === "app-only" && (
              <svg className="w-4 h-4 text-[var(--color-accent)] animate-spin ml-auto shrink-0" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
            )}
          </button>

          <button
            onClick={handleQuitAppAndBrowsers}
            disabled={quitting}
            className="flex items-center gap-3 w-full px-4 py-3 rounded-lg border border-red-200 bg-red-50/50 hover:bg-red-50 transition-colors text-left disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <div className="w-8 h-8 rounded-md bg-red-100 text-red-500 flex items-center justify-center shrink-0">
              <AlertCircle className="w-4 h-4" />
            </div>
            <div className="min-w-0">
              <div className="text-sm font-medium text-red-600">退出全部进程</div>
              <div className="text-xs text-[var(--color-text-muted)] mt-0.5">关闭客户端，终止所有浏览器实例</div>
            </div>
            {quittingAction === "app-and-browser" && (
              <svg className="w-4 h-4 text-red-500 animate-spin ml-auto shrink-0" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
            )}
          </button>
        </div>
      )}
    </Modal>
  );
}

function App() {
  useEffect(() => {
    installWailsOperationLogger();
  }, []);
  useWailsNotifications();
  useGlobalErrorNotifications();
  const [quickLaunchOpen, setQuickLaunchOpen] = useState(false);
  const routeFallback = (
    <div className="flex min-h-[240px] items-center justify-center py-10">
      <Loading text="页面加载中..." />
    </div>
  );

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.isComposing) return;
      if (!(event.ctrlKey || event.metaKey)) return;
      if (event.key.toLowerCase() !== "k") return;
      event.preventDefault();
      setQuickLaunchOpen((prev) => !prev);
    };

    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
    };
  }, []);

  return (
    <Router>
      <Layout>
        <Suspense fallback={routeFallback}>
          <AppRoutes />
        </Suspense>
      </Layout>
      <ToastContainer />
      <CloseConfirmModal />
      <Suspense fallback={null}>
        {quickLaunchOpen ? (
          <QuickLaunchModal
            open={quickLaunchOpen}
            onClose={() => setQuickLaunchOpen(false)}
          />
        ) : null}
      </Suspense>
    </Router>
  );
}

export default App;
