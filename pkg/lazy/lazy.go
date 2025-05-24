// Package lazy provides functionality for lazy-loading.
//
// See https://en.wikipedia.org/wiki/Lazy_loading for an explanation of this technique.
package lazy

import (
	"sync"
)

type lazy[T any, TLoader func() T | func() (T, error)] struct {
	once  sync.Once
	value *T
	err   error
	load  TLoader
}

// Lazy allows retrieving a lazy-loaded value, allowing for cases where the loading logic may
// produce an error.
//
// This type should not be directly instantiated; use [New] instead.
type Lazy[T any] lazy[T, func() (T, error)]

// Load returns a 2-tuple of the lazy-loaded value and an error.
//
// If the error was non-nil, the value will be a zero-value of the specified type.
func (l *Lazy[T]) Load() (T, error) {
	l.once.Do(func() {
		result, err := l.load()
		if err != nil {
			l.value = new(T)
			l.err = err
			return
		}
		l.value = &result
	})

	return *l.value, l.err
}

// New returns a new [Lazy].
//
// The provided loader function must not be nil.
func New[T any](loader func() (T, error)) *Lazy[T] {
	return &Lazy[T]{load: loader}
}

// MustLazy allows retrieving a lazy-loaded value.
//
// The provided loader function may not return an error. In cases where the loading logic may return
// an error, use [Lazy] instead.
//
// This type should not be directly instantiated; use [NewMust] instead.
type MustLazy[T any] lazy[T, func() T]

// Load returns the lazy-loaded value.
func (lm *MustLazy[T]) Load() T {
	lm.once.Do(func() {
		result := lm.load()
		lm.value = &result
	})

	return *lm.value
}

// NewMust returns a new [MustLazy].
//
// The provided loader function must not be nil.
func NewMust[T any](loader func() T) *MustLazy[T] {
	return &MustLazy[T]{load: loader}
}

