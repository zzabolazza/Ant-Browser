//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>
#import <math.h>

static BOOL antDockSyncInstalled = NO;
static BOOL antDockSyncInProgress = NO;

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

static BOOL antWindowLooksMaximized(NSWindow *window) {
    if (window == nil || !window.isVisible || window.isMiniaturized) {
        return NO;
    }
    if ([window isZoomed]) {
        return YES;
    }
    NSScreen *screen = [window screen];
    if (screen == nil) {
        return NO;
    }
    NSRect frame = [window frame];
    NSRect visible = [screen visibleFrame];
    const CGFloat epsilon = 8.0;
    return fabs(frame.size.width - visible.size.width) < epsilon
        && fabs(frame.size.height - visible.size.height) < epsilon
        && fabs(frame.origin.x - visible.origin.x) < epsilon
        && fabs(frame.origin.y - visible.origin.y) < epsilon;
}

static void antApplyFullscreenNone(NSWindow *window) {
    if (window == nil) {
        return;
    }
    NSWindowCollectionBehavior behavior = [window collectionBehavior];
    behavior &= ~(NSWindowCollectionBehaviorFullScreenPrimary | NSWindowCollectionBehaviorFullScreenAuxiliary);
    behavior |= NSWindowCollectionBehaviorFullScreenNone;
    [window setCollectionBehavior:behavior];
}

static void antSyncDockVisibility(void) {
    if (antDockSyncInProgress) {
        return;
    }
    NSWindow *window = antFindMainWindow();
    if (window == nil) {
        return;
    }
    antApplyFullscreenNone(window);

    BOOL maximized = antWindowLooksMaximized(window);
    BOOL appActive = [NSApp isActive];
    NSApplicationPresentationOptions current = [NSApp presentationOptions];
    BOOL dockHidden = (current & NSApplicationPresentationHideDock) != 0;

    antDockSyncInProgress = YES;
    if (maximized && appActive) {
        if (!dockHidden) {
            [NSApp setPresentationOptions:NSApplicationPresentationHideDock];
        }
        // Dock 隐藏后 visibleFrame 会变大，把窗口铺满新区域（仍保留标题栏红绿灯）。
        NSRect target = [[window screen] visibleFrame];
        NSRect frame = [window frame];
        const CGFloat epsilon = 2.0;
        if (fabs(frame.size.width - target.size.width) > epsilon
            || fabs(frame.size.height - target.size.height) > epsilon
            || fabs(frame.origin.x - target.origin.x) > epsilon
            || fabs(frame.origin.y - target.origin.y) > epsilon) {
            [window setFrame:target display:YES animate:NO];
        }
    } else if (dockHidden) {
        [NSApp setPresentationOptions:NSApplicationPresentationDefault];
    }
    antDockSyncInProgress = NO;
}

static void antInstallDockVisibilityObservers(void) {
    if (antDockSyncInstalled) {
        return;
    }
    antDockSyncInstalled = YES;

    NSNotificationCenter *center = [NSNotificationCenter defaultCenter];
    void (^sync)(NSNotification *) = ^(NSNotification *note) {
        (void)note;
        antSyncDockVisibility();
    };

    [center addObserverForName:NSWindowDidResizeNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:sync];
    [center addObserverForName:NSWindowDidEndLiveResizeNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:sync];
    [center addObserverForName:NSWindowDidBecomeKeyNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:sync];
    [center addObserverForName:NSApplicationDidBecomeActiveNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:sync];
    [center addObserverForName:NSApplicationDidResignActiveNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:sync];
    [center addObserverForName:NSApplicationWillTerminateNotification object:nil queue:[NSOperationQueue mainQueue] usingBlock:^(NSNotification *note) {
        (void)note;
        [NSApp setPresentationOptions:NSApplicationPresentationDefault];
    }];
}

void AntPreferWindowZoomOverFullscreen(void) {
    void (^apply)(void) = ^{
        antInstallDockVisibilityObservers();
        NSWindow *window = antFindMainWindow();
        antApplyFullscreenNone(window);
        antSyncDockVisibility();
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
