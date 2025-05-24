package runner

type options struct {
	// CancelOnFailure indicates whether the Runner should cancel its context when a task fails.
	CancelOnFailure bool
	// Limit is the maximum number of goroutines that may run simultaneously.
	Limit uint
}

// Option allows specifying a configuration option when creating a new Runner.
type Option func(*options)

// WithCancelOnFailure is an option that will make the Runner cancel its context when a task fails,
// resulting in preventing subsequent attempts to run tasks via [Runner.Go].
func WithCancelOnFailure() Option {
	return func(o *options) {
		o.CancelOnFailure = true
	}
}

// WithLimit is an option that allows setting the maximum number of goroutines that may run
// simultaneously.
//
// Specifying a limit of 0 is equivalent to not specifying a limit.
func WithLimit(limit uint) Option {
	return func(o *options) {
		o.Limit = limit
	}
}

