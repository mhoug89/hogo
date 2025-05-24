// Package runner provides a way to run multiple goroutines, optionally with a maximum number of
// simultaneous goroutines. When the max is reached, attempting to run a new goroutine will block
// until the number of running goroutines drops below the max. It also provides a way to aggregate
// the errors returned from functions executed in each goroutine.
package runner

import (
	"context"
	"errors"
	"sync"
)

// cancelOnFailure contains the cancellation function for a context to be canceled when a task
// fails. If the Cancel field is nil, the Runner should not cancel its context when a task fails.
type cancelOnFailure struct {
	cancel context.CancelCauseFunc
	once   sync.Once
}

// ShouldCancel returns true if the cancel function is non-nil.
func (c *cancelOnFailure) ShouldCancel() bool {
	return c.cancel != nil
}

// Cancel invokes the underlying cancel function once via a [sync.Once]. Subsequent calls to Cancel
// will be no-ops.
func (c *cancelOnFailure) Cancel(err error) {
	c.once.Do(func() {
		c.cancel(err)
	})
}

// Runner allows running multiple goroutines with built-in WaitGroup management and error
// accumulation.
//
// If the [WithLimit] option is provided, the maximum number of simultaneous goroutines is
// restricted to the provided limit. When the limit is reached, attempting to run a new goroutine
// will block until the number of running goroutines drops below the max.
//
// If the [WithContinueOnFailure] option is provided, a derived context is created and used to
// manage cancellation of the Runner's tasks. Upon receiving the first non-nil error from a task:
//   - The context is canceled, using the first encountered error as the cancellation reason.
//   - The Runner will avoid running tasks in subsequent calls to [Runner.Go].
type Runner struct {
	ctx          context.Context
	failCanceler cancelOnFailure
	wg           sync.WaitGroup
	errs         syncErrorSlice
	// Utilize a channel to act a semaphore.
	sem chan struct{}
}

// New returns a new Runner using the provided options.
func New(ctx context.Context, opts ...Option) *Runner {
	r := &Runner{ctx: ctx}

	ro := options{}
	for _, opt := range opts {
		opt(&ro)
	}
	if ro.Limit > 0 {
		r.sem = make(chan struct{}, ro.Limit)
	}
	if ro.CancelOnFailure {
		r.ctx, r.failCanceler.cancel = context.WithCancelCause(ctx)
	}

	return r
}

func (r *Runner) hasLimit() bool {
	return cap(r.sem) > 0
}

// maybeSemInc will add an item to the sem channel if a simultaneous goroutine limit was set,
// blocking if it is full. If no limit was set, this is a no-op.
func (r *Runner) maybeSemInc() {
	if r.hasLimit() {
		r.sem <- struct{}{}
	}
}

// maybeSemDec will remove an item from the sem channel if a simultaneous goroutine limit was set.
// If no limit was set, this is a no-op.
func (r *Runner) maybeSemDec() {
	if r.hasLimit() {
		<-r.sem
	}
}

// Go runs the given function in a goroutine when the number of running goroutines has not reached
// the limit. If the limit is reached, this method blocks until some goroutines finish.
//
// Go should not be used in a nested manner, i.e. nesting a Go call within another Go call.
func (r *Runner) Go(f func() error) {
	r.maybeSemInc()
	r.wg.Add(1)
	var result error
	go func() {
		defer func() {
			if result != nil {
				r.errs.Append(result)
				if r.failCanceler.ShouldCancel() {
					// If cancellation is desired, cancel using the first error we encounter as the
					// cause. Any goroutines that were already started before this cancellation will
					// still have their errors recorded, but will not be included in the
					// cancellation cause.
					r.failCanceler.Cancel(result)
				}
			}
			r.wg.Done()
			r.maybeSemDec()
		}()

		// If a context was provided and it's now done, don't run the function.
		if r.ctx.Err() != nil {
			result = causeForTaskSkip(r.ctx)
			return
		}
		result = f()
	}()
}

// Wait blocks until all function calls from the Go method have returned, then returns all the
// errors from all goroutines.
func (r *Runner) Wait() []error {
	r.wg.Wait()
	return r.errs.Clone()
}

func causeForTaskSkip(ctx context.Context) error {
	cause := context.Cause(ctx)
	if errors.Is(cause, context.Canceled) || errors.Is(cause, context.DeadlineExceeded) {
		return cause
	}
	// If the cancellation cause was an error from a previous task, mark subsequent tasks' errors as
	// being due to context.Canceled, rather than making it appear as if all the tasks were executed
	// and produced the same error.
	return context.Canceled
}

