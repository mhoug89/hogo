// Package rwguarded provides a thread-safe wrapper around underlying values using [sync.RWMutex] to
// synchronize the provided operations.
package rwguarded

import (
	"sync"
)

// RWGuarded is a thin wrapper around the provided value that uses a [sync.RWMutex] to synchronize
// operations. This struct should not be directly instantiated; callers should use the [New]
// function instead.
type RWGuarded[V any] struct {
	rwLock *sync.RWMutex
	value  V
}

// New initializes and returns a [RWGuarded] of the provided type.
func New[V any](val V) *RWGuarded[V] {
	return &RWGuarded[V]{
		rwLock: &sync.RWMutex{},
		value:  val,
	}
}

// Get returns the underlying value.
func (g *RWGuarded[V]) Get() V {
	g.rwLock.RLock()
	defer g.rwLock.RUnlock()

	return g.value
}

// Set sets the underlying value.
func (g *RWGuarded[V]) Set(val V) {
	g.rwLock.Lock()
	defer g.rwLock.Unlock()

	g.value = val
}

// Update allows performing a read-modify-write transaction on the underlying value while holding
// the writer lock. The updater function is passed a pointer to the underlying value, which it may
// change in place. The error value returned from the updater is returned from this method.
//
// The updater should not call any other method of this [RWGuarded], as this will result in a
// deadlock.
func (g *RWGuarded[V]) Update(updater func(*V) error) error {
	g.rwLock.Lock()
	defer g.rwLock.Unlock()

	return updater(&g.value)
}

