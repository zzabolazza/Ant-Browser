package proxy

import "sync"

type xrayLaunchLock struct {
	mu   sync.Mutex
	refs int
}

func (m *XrayManager) lockLaunchForKey(key string) func() {
	if m == nil {
		return func() {}
	}
	m.mu.Lock()
	if m.launchLocks == nil {
		m.launchLocks = make(map[string]*xrayLaunchLock)
	}
	lock := m.launchLocks[key]
	if lock == nil {
		lock = &xrayLaunchLock{}
		m.launchLocks[key] = lock
	}
	lock.refs++
	m.mu.Unlock()

	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		m.mu.Lock()
		lock.refs--
		if lock.refs <= 0 && m.launchLocks[key] == lock {
			delete(m.launchLocks, key)
		}
		m.mu.Unlock()
	}
}
