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

// LockedDo allows performing an operation using the underlying value while holding the writer lock.
// The provided function is passed a pointer to the underlying value. Some use cases for using
// LockedDo might include:
//   - Calling a method on the underlying object
//   - Modifying the underlying object in-place
//
// This function returns the error returned from the provided function.
//
// The provided function should NOT call any other method of this RWGuarded, as this will result in
// a deadlock.
func (g *RWGuarded[V]) LockedDo(fn func(*V) error) error {
	g.rwLock.Lock()
	defer g.rwLock.Unlock()

	return fn(&g.value)
}

