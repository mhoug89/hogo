// Package erroror implements ErrorOr types, roughly equivalent to abseil's StatusOr<T> in C++.
//
// Since returning multiple values from a function, with the trailing value being an error, is a
// common Go idiom, that should be preferred over using ErrorOr in most code. However, ErrorOr is
// often useful for unit testing functions that use this idiom. E.g., for a function that returns
// (string, error), your TestCase struct could use a single field `Want ErrorOr[string]` to express
// the expected return values, rather than two separate fields `WantStr string` and `WantErr error`.
package erroror

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// ErrorOr holds a value of type T and an error. It is commonly used in unit tests to better
// encapsulate expected results from functions with 2 return values, where the trailing return value
// is an error.
type ErrorOr[T any] struct {
	val T
	err error
}

// Ok returns true if this ErrorOr's underlying error is nil.
func (e *ErrorOr[T]) Ok() bool {
	return e.err == nil
}

// Value returns this ErrorOr's underlying value. In most cases, this should only be called after
// verifying that [Ok] is true.
func (e *ErrorOr[T]) Value() T {
	return e.val
}

// Unpack returns the underlying value and error. It can be used to unpack an ErrorOr into the same
// values that it was created with, which is useful for tests where the ErrorOr is created in one
// scope and unpacked in another.
func (e *ErrorOr[T]) Unpack() (T, error) {
	return e.val, e.err
}

// Of returns an ErrorOr containing the given value and error.
func Of[T any](val T, err error) ErrorOr[T] {
	return ErrorOr[T]{
		val: val,
		err: err,
	}
}

// OfValue creates an ErrorOr from a value. Its underlying error is set to nil.
func OfValue[T any](val T) ErrorOr[T] {
	return ErrorOr[T]{
		val: val,
	}
}

// OfError creates an ErrorOr from an error. Its underlying value is set to T's zero value.
func OfError[T any](err error) ErrorOr[T] {
	return ErrorOr[T]{
		err: err,
	}
}

// CmpDiff is a convenience function for calling [cmp.Diff] with two [ErrorOr] structs. It uses the
// options returned from [cmpopts.EquateErrors] and [cmp.AllowUnexported], in addition to whatever
// options are provided.
func CmpDiff[T any](want ErrorOr[T], got ErrorOr[T], opts ...cmp.Option) string {
	defaultOpts := []cmp.Option{
		cmpopts.EquateErrors(),
		// Could also pass `want` or `&ErrorOr[T]` here; AllowUnexported only cares about the type
		// of the argument, not its values:
		cmp.AllowUnexported(got),
	}
	return cmp.Diff(want, got, append(defaultOpts, opts...)...)
}

// ErrorOr2 holds 2 values, of type T1 and T2, and an error. It is commonly used in unit tests to
// better encapsulate expected results from functions with 3 return values, where the trailing
// return value is an error.
type ErrorOr2[T1 any, T2 any] struct {
	val1 T1
	val2 T2
	err  error
}

// Ok returns true if this ErrorOr2's underlying error is nil.
func (e *ErrorOr2[T1, T2]) Ok() bool {
	return e.err == nil
}

// Value returns this ErrorOr2's underlying values. In most cases, this should only be called after
// verifying that [Ok] is true.
func (e *ErrorOr2[T1, T2]) Value() (T1, T2) {
	return e.val1, e.val2
}

// Unpack returns the underlying values and error. It can be used to unpack an ErrorOr2 into the
// same values that it was created with, which is useful for tests where the ErrorOr2 is created in
// one scope and unpacked in another.
func (e *ErrorOr2[T1, T2]) Unpack() (T1, T2, error) {
	return e.val1, e.val2, e.err
}

// Of2 returns an ErrorOr2 containing the given values and error.
func Of2[T1 any, T2 any](val1 T1, val2 T2, err error) ErrorOr2[T1, T2] {
	return ErrorOr2[T1, T2]{
		val1: val1,
		val2: val2,
		err:  err,
	}
}

// OfValue2 creates an ErrorOr2 from two values. Its underlying error is set to nil.
func OfValue2[T1 any, T2 any](val1 T1, val2 T2) ErrorOr2[T1, T2] {
	return ErrorOr2[T1, T2]{
		val1: val1,
		val2: val2,
	}
}

// OfError2 creates an ErrorOr2 from an error. Each of its underlying values is set to its
// respective type's zero value.
func OfError2[T1 any, T2 any](err error) ErrorOr2[T1, T2] {
	return ErrorOr2[T1, T2]{
		err: err,
	}
}

// CmpDiff2 is a convenience function for calling [cmp.Diff] with two [ErrorOr2] structs. It uses
// the options returned from [cmpopts.EquateErrors] and [cmp.AllowUnexported], in addition to
// whatever options are provided.
func CmpDiff2[T1 any, T2 any](want ErrorOr2[T1, T2], got ErrorOr2[T1, T2], opts ...cmp.Option) string {
	defaultOpts := []cmp.Option{
		cmpopts.EquateErrors(),
		// Could also pass `want` or `&ErrorOr[T]` here; AllowUnexported only cares about the type
		// of the argument, not its values:
		cmp.AllowUnexported(got),
	}
	return cmp.Diff(want, got, append(defaultOpts, opts...)...)
}
