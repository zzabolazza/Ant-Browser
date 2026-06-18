package proxy

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingBoxLaunchLockSerializesSameKey(t *testing.T) {
	manager := NewSingBoxManager(nil, "")
	defer manager.StopAll()

	const workers = 8
	var active int32
	var maxActive int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			unlock := manager.lockLaunchForKey("same-node")
			current := atomic.AddInt32(&active, 1)
			for {
				old := atomic.LoadInt32(&maxActive)
				if current <= old || atomic.CompareAndSwapInt32(&maxActive, old, current) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&active, -1)
			unlock()
		}()
	}
	close(start)
	wg.Wait()

	if maxActive != 1 {
		t.Fatalf("same-key launch lock allowed %d concurrent entries", maxActive)
	}
}
