package proxy

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestXrayLaunchLockSerializesSameKey(t *testing.T) {
	t.Parallel()

	manager := &XrayManager{}
	unlockFirst := manager.lockLaunchForKey("node-a")
	acquiredSecond := make(chan struct{})
	go func() {
		unlockSecond := manager.lockLaunchForKey("node-a")
		defer unlockSecond()
		close(acquiredSecond)
	}()

	select {
	case <-acquiredSecond:
		t.Fatalf("same-key launch lock was not serialized")
	case <-time.After(30 * time.Millisecond):
	}

	unlockFirst()
	select {
	case <-acquiredSecond:
	case <-time.After(time.Second):
		t.Fatalf("same-key launch lock did not release")
	}
}

func TestXrayLaunchLockAllowsDifferentKeys(t *testing.T) {
	t.Parallel()

	manager := &XrayManager{}
	unlockFirst := manager.lockLaunchForKey("node-a")
	defer unlockFirst()
	var acquired int32
	done := make(chan struct{})
	go func() {
		unlockSecond := manager.lockLaunchForKey("node-b")
		defer unlockSecond()
		atomic.StoreInt32(&acquired, 1)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("different-key launch lock was unexpectedly blocked")
	}
	if atomic.LoadInt32(&acquired) != 1 {
		t.Fatalf("different-key launch lock was not acquired")
	}
}
