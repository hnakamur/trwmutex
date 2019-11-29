package trwmutex

import "sync"

// TRWMutex is extended RWMutex which have TryLock() and TryRLock().
type TRWMutex struct {
	mu   sync.Mutex
	rwmu sync.RWMutex
	w    int
	r    int
}

// Lock locks m and wait until all other Lock or RLock is unlocked.
func (m *TRWMutex) Lock() {
	m.mu.Lock()
	m.w++
	if m.r > 0 || m.w > 1 {
		// other one already acquired lock. wait outside of m.mu lock
		m.mu.Unlock()
		m.rwmu.Lock()
	} else {
		// m.rwmu.Lock() never blocks
		m.rwmu.Lock()
		m.mu.Unlock()
	}
}

// TryLock try to lock m. returns false if fails.
func (m *TRWMutex) TryLock() bool {
	m.mu.Lock()
	if m.r > 0 || m.w > 0 {
		// other one already acquired lock.
		m.mu.Unlock()
		return false
	}
	m.w++
	// m.rwmu.Lock() never blocks
	m.rwmu.Lock()
	m.mu.Unlock()
	return true
}

// Unlock unlocks m.
func (m *TRWMutex) Unlock() {
	m.mu.Lock()
	m.w--
	m.rwmu.Unlock()
	m.mu.Unlock()
}

// RLock locks m shared and until other Lock is unlocked.
func (m *TRWMutex) RLock() {
	m.mu.Lock()
	m.r++
	if m.w > 0 {
		// other one already acquired lock. wait outside of m.mu lock
		m.mu.Unlock()
		m.rwmu.RLock()
	} else {
		// m.rwmu.RLock() never blocks
		m.rwmu.RLock()
		m.mu.Unlock()
	}
}

// TryRLock try to lock m shared. returns false if fails.
func (m *TRWMutex) TryRLock() bool {
	m.mu.Lock()
	if m.w > 0 {
		// other one already acquired lock.
		m.mu.Unlock()
		return false
	}
	m.r++
	// m.rwmu.RLock() never blocks
	m.rwmu.RLock()
	m.mu.Unlock()
	return true
}

// RUnlock unlocks m.
func (m *TRWMutex) RUnlock() {
	m.mu.Lock()
	m.r--
	m.rwmu.RUnlock()
	m.mu.Unlock()
}

// Upgrade converts reader lock to writer lock and returns success (true) or dead-lock (false).
// If Upgrade by multi reader locker at same time then dead-lock.
// Upgrade is given priority to Lock.
func (m *TRWMutex) Upgrade() {
	m.mu.Lock()
	m.w++
	if m.r > 1 || m.w > 1 {
		// other one already acquired lock.
		// release the read lock and wait for write lock outside of m.mu lock.
		m.rwmu.RUnlock()
		m.r--
		m.mu.Unlock()
		m.rwmu.Lock()
	} else {
		// release the read lock and acquire write lock.
		m.rwmu.RUnlock()
		m.r--
		// m.rwmu.Lock() never blocks
		m.rwmu.Lock()
		m.mu.Unlock()
	}
}

// TryUpgrade try to convert reader lock to writer lock and returns success (true) or dead-lock (false).
// If Upgrade by multi reader locker at same time then dead-lock.
// Upgrade is given priority to Lock.
func (m *TRWMutex) TryUpgrade() bool {
	m.mu.Lock()
	if m.r > 1 || m.w > 1 {
		// other one already acquired lock.
		m.mu.Unlock()
		return false
	}
	m.w++
	m.rwmu.RUnlock()
	m.r--
	// m.rwmu.Lock() never blocks
	m.rwmu.Lock()
	m.mu.Unlock()
	return true
}
