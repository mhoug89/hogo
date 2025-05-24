package runner

import (
	"slices"
	"sync"
)

type syncErrorSlice struct {
	mutex sync.Mutex
	errs  []error
}

// Append adds the error to the underlying slice.
func (e *syncErrorSlice) Append(err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.errs = append(e.errs, err)
}

// Clone returns a clone of the accumulated slice of errors.
func (e *syncErrorSlice) Clone() []error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return slices.Clone(e.errs)
}

