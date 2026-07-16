//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>
#import <objc/runtime.h>
#import <math.h>

static BOOL antDockSyncInstalled = NO;
static BOOL antDockSyncInProgress = NO;
static BOOL antIgnoreDockSyncEcho = NO;
static BOOL antCustomMaximized = NO;
static BOOL antHasRestoreFrame = NO;
static NSRect antRestoreFrame;
static dispatch_block_t antPendingDockSync = nil;
static void (*antOriginalZoom)(id, SEL, id) = NULL;

static NSWindow *antFindMainWindow(void) {
    NSWindow *window = [NSApp mainWindow];
    if (window != nil) {
        return window;
    }
    window = [NSApp keyWindow];
    if (window != nil) {
        return window;
    }
    for (NSWindow *candidate in [NSApp windows]) {
        if (candidate.isVisible && (candidate.styleMask & NSWindowStyleMaskTitled)) {
            return candidate;
        }
    }
    return nil;
}

static BOOL antFrameFillsRect(NSRect frame, NSRect target, CGFloat epsilon) {
    return fabs(frame.size.width - target.size.width) < epsilon
        && fabs(frame.size.height - target.size.height) < epsilon
        && fabs(frame.origin.x - target.origin.x) < epsilon
        && fabs(frame.origin.y - target.origin.y) < epsilon;
}

static BOOL antWindowLooksMaximized(NSWindow *window) {
    if (window == nil || !window.isVisible || window.isMiniaturized) {
        return NO;
    }
    if (antCustomMaximized) {
        return YES;
    }
    NSScreen *screen = [window screen];
    if (screen == nil) {
        return NO;
    }
    return antFrameFillsRect([window frame], [screen visibleFrame], 12.0);
}

static void antApplyFullscreenNone(NSWindow *window) {
    if (window == nil) {
        return;
    }
    NSWindowCollectionBehavior behavior = [window collectionBehavior];
    behavior &= ~(NSWindowCollectionBehaviorFullScreenPrimary | NSWindowCollectionBehaviorFullScreenAuxiliary);
    behavior |= NSWindowCollectionBehaviorFullScreenNone;
    [window setCollectionBehavior:behavior];
    // 关闭窗口尺寸动画（含系统 zoom:）。
    [window setAnimationBehavior:NSWindowAnimationBehaviorNone];
}

static void antSetFrameInstant(NSWindow *window, NSRect frame) {
    antIgnoreDockSyncEcho = YES;
    [window setFrame:frame display:YES animate:NO];
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.05 * NSEC_PER_SEC)),
                   dispatch_get_main_queue(), ^{
        antIgnoreDockSyncEcho = NO;
    });
}

static void antMaximizeInstant(NSWindow *window) {
    if (window == nil) {
        return;
    }
    antApplyFullscreenNone(window);
    if (!antCustomMaximized) {
        antRestoreFrame = [window frame];
        antHasRestoreFrame = YES;
    }
    antCustomMaximized = YES;
    if ([NSApp isActive]) {
        [NSApp setPresentationOptions:NSApplicationPresentationHideDock];
    }
    NSScreen *screen = [window screen];
    if (screen == nil) {
        return;
    }
    antSetFrameInstant(window, [screen visibleFrame]);
}

static void antRestoreInstant(NSWindow *window) {
    if (window == nil) {
        return;
    }
    antCustomMaximized = NO;
    [NSApp setPresentationOptions:NSApplicationPresentationDefault];
    if (antHasRestoreFrame) {
        antSetFrameInstant(window, antRestoreFrame);
    }
}

static void antToggleMaximizeInstant(NSWindow *window) {
    if (antWindowLooksMaximized(window)) {
        antRestoreInstant(window);
        return;
    }
    antMaximizeInstant(window);
}

static void antZoomOverride(id self, SEL _cmd, id sender) {
    NSWindow *window = (NSWindow *)self;
    NSWindow *mainWindow = antFindMainWindow();
    // 只接管主窗口；其它窗口仍走系统 zoom。
    if (mainWindow != nil && window != mainWindow) {
        if (antOriginalZoom != NULL) {
            antOriginalZoom(self, _cmd, sender);
        }
        return;
    }
    if (!(window.styleMask & NSWindowStyleMaskTitled)) {
        if (antOriginalZoom != NULL) {
            antOriginalZoom(self, _cmd, sender);
        }
        return;
    }
    antToggleMaximizeInstant(window);
}

static void antInstallZoomOverride(void) {
    static BOOL installed = NO;
    if (installed) {
        return;
    }
    installed = YES;
    Method method = class_getInstanceMethod([NSWindow class], @selector(zoom:));
    if (method == NULL) {
        return;
    }
    antOriginalZoom = (void (*)(id, SEL, id))method_getImplementation(method);
    method_setImplementation(method, (IMP)antZoomOverride);
}

static void antSyncDockVisibility(void);
static void antScheduleDockSync(NSTimeInterval delaySeconds);

// 用户拖拽改尺寸后：若已离开最大化区域，恢复 Dock；若仍铺满则保持最大化。
static void antSyncDockVisibility(void) {
    if (antDockSyncInProgress || antIgnoreDockSyncEcho) {
        return;
    }
    NSWindow *window = antFindMainWindow();
    if (window == nil) {
        return;
    }
    if ([window inLiveResize]) {
        antScheduleDockSync(0.05);
        return;
    }
    antApplyFullscreenNone(window);

    BOOL fills = NO;
    NSScreen *screen = [window screen];
    if (screen != nil) {
        fills = antFrameFillsRect([window frame], [screen visibleFrame], 12.0);
    }
    BOOL appActive = [NSApp isActive];
    NSApplicationPresentationOptions current = [NSApp presentationOptions];
    BOOL dockHidden = (current & NSApplicationPresentationHideDock) != 0;

    antDockSyncInProgress = YES;
    if (antCustomMaximized || fills) {
        if (!fills) {
            // 用户已拖离最大化区域。
            antCustomMaximized = NO;
            if (dockHidden) {
                [NSApp setPresentationOptions:NSApplicationPresentationDefault];
            }
        } else if (appActive) {
            antCustomMaximized = YES;
            if (!dockHidden) {
                [NSApp setPresentationOptions:NSApplicationPresentationHideDock];
            }
            // Dock 刚隐藏时 visibleFrame 变大，无动画补满。
            if (screen != nil && !antFrameFillsRect([window frame], [screen visibleFrame], 2.0)) {
                antSetFrameInstant(window, [screen visibleFrame]);
            }
        } else if (dockHidden) {
            [NSApp setPresentationOptions:NSApplicationPresentationDefault];
        }
    } else if (dockHidden) {
        antCustomMaximized = NO;
        [NSApp setPresentationOptions:NSApplicationPresentationDefault];
    }
    antDockSyncInProgress = NO;
}

static void antScheduleDockSync(NSTimeInterval delaySeconds) {
    if (antPendingDockSync != nil) {
        dispatch_block_cancel(antPendingDockSync);
        antPendingDockSync = nil;
    }
    dispatch_block_t block = dispatch_block_create(0, ^{
        antPendingDockSync = nil;
        antSyncDockVisibility();
    });
    antPendingDockSync = block;
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(delaySeconds * NSEC_PER_SEC)),
                   dispatch_get_main_queue(),
                   block);
}

static void antInstallDockVisibilityObservers(void) {
    if (antDockSyncInstalled) {
        return;
    }
    antDockSyncInstalled = YES;

    NSNotificationCenter *center = [NSNotificationCenter defaultCenter];
    void (^syncSoon)(NSNotification *) = ^(NSNotification *note) {
        (void)note;
        antScheduleDockSync(0.02);
    };

    // 不再依赖 DidResize 防抖等动画结束：zoom 已无动画；仅同步拖拽改尺寸与激活状态。
    [center addObserverForName:NSWindowDidEndLiveResizeNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:syncSoon];
    [center addObserverForName:NSWindowDidBecomeKeyNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:syncSoon];
    [center addObserverForName:NSApplicationDidBecomeActiveNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:syncSoon];
    [center addObserverForName:NSApplicationDidResignActiveNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:syncSoon];
    [center addObserverForName:NSApplicationWillTerminateNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:^(NSNotification *note) {
        (void)note;
        if (antPendingDockSync != nil) {
            dispatch_block_cancel(antPendingDockSync);
            antPendingDockSync = nil;
        }
        [NSApp setPresentationOptions:NSApplicationPresentationDefault];
    }];
}

void AntPreferWindowZoomOverFullscreen(void) {
    void (^apply)(void) = ^{
        antInstallZoomOverride();
        antInstallDockVisibilityObservers();
        NSWindow *window = antFindMainWindow();
        antApplyFullscreenNone(window);
        antScheduleDockSync(0);
    };

    if ([NSThread isMainThread]) {
        apply();
        return;
    }
    dispatch_async(dispatch_get_main_queue(), apply);
}
*/
import "C"

func preferWindowZoomOverFullscreen() {
	C.AntPreferWindowZoomOverFullscreen()
}
