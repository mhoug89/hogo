// Package optional provides functionality for working with "optional" values, i.e. a wrapper for a
// value that may or may not be set.
//
// Semantically, using an [Optional] allows indicating whether a value is set/populated without
// having to compare the value to a sentinel or zero value. This eliminates the need for
// type-specific comparison logic and is useful in several cases, including when:
//   - The zero value of the desired type might also be a valid value, e.g. in the case of an int
//     where 0 should not be treated as the absence of a number.
//   - One might not want to use a pointer for this purpose (e.g. designating a nil pointer as
//     "unset") due to the additional overhead involved with logging, serialization, etc. of
//     pointers.
package optional

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// ErrNotSet is returned when calling [Optional.Get] on an empty [Optional].
var ErrNotSet = errors.New("optional value not set")

// Optional may hold a value of type T. To check whether an [Optional] has been populated, the
// [IsSet] and [IsEmpty] methods may be used.
//
// This type should not be directly instantiated; use [Of] or [Empty] instead.
type Optional[T any] struct {
	isSet bool
	value *T
}

// Empty creates a new, unset [Optional].
func Empty[T any]() Optional[T] {
	return Optional[T]{}
}

// Of creates a new [Optional] containing the specified value.
func Of[T any](value T) Optional[T] {
	return Optional[T]{
		isSet: true,
		value: &value,
	}
}

// IsSet returns true if the [Optional] is populated.
func (o *Optional[T]) IsSet() bool {
	return o.isSet
}

// IsEmpty returns true if the [Optional] is not populated.
func (o *Optional[T]) IsEmpty() bool {
	return !o.IsSet()
}

// Set populates the [Optional] with the given value.
func (o *Optional[T]) Set(value T) {
	o.isSet = true
	o.value = &value
}

// Get returns the value stored in the [Optional] if it is set.
//
// If the [Optional] is unset, the returned error will be non-nil.
func (o *Optional[T]) Get() (T, error) {
	if o.IsEmpty() {
		var tZeroVal T
		return tZeroVal, ErrNotSet
	}
	return *o.value, nil
}

// MustGet returns the value stored in the [Optional] if it is set.
//
// If the [Optional] is unset, this method panics.
func (o *Optional[T]) MustGet() T {
	if o.IsEmpty() {
		panic(ErrNotSet)
	}
	return *o.value
}

// OrElse returns the value stored in the [Optional] if it is set, otherwise it returns
// defaultValue.
func (o *Optional[T]) OrElse(defaultValue T) T {
	if o.IsSet() {
		return *o.value
	}
	return defaultValue
}

// OrElseLazy returns the value stored in the [Optional] if it is set, otherwise it returns the
// result of the given callback.
func (o *Optional[T]) OrElseLazy(callback func() (T, error)) (T, error) {
	if o.IsSet() {
		return *o.value, nil
	}
	return callback()
}

// OrElseMustLazy returns the value stored in the [Optional] if it is set, otherwise it returns the
// result of the given callback.
//
// If the required callback may return an error, use [Optional.OrElseLazy] instead.
func (o *Optional[T]) OrElseMustLazy(callback func() T) T {
	if o.IsSet() {
		return *o.value
	}
	return callback()
}

// Equal returns true if the given [Optional] is equal to another.
//
// The provided argument may be an [Optional] of the same type, or a pointer to one. If an argument
// of any other type is provided, the two objects are never considered equal.
//
// Two Optionals of the same type are considered to be equal if any of the following are true:
//   - They are both unset.
//   - They are both set and their underlying values are equal as determined by [reflect.DeepEqual].
func (o *Optional[T]) Equal(o2 any) bool {
	var other *Optional[T]
	switch o2.(type) {
	case *Optional[T]:
		other = o2.(*Optional[T])
	case Optional[T]:
		other = ptrTo(o2.(Optional[T]))
	default:
		return false
	}

	if o.IsSet() != other.IsSet() {
		return false
	}
	return o.IsEmpty() || reflect.DeepEqual(*o.value, *other.value)
}

// String returns a string representation of the [Optional].
//
// If the [Optional] in unset, the returned string will be "<empty>".
func (o *Optional[T]) String() string {
	if o.IsSet() {
		return fmt.Sprint(*o.value)
	}
	return "<empty>"
}

// MarshalJSON marshals the underlying value if set, otherwise returns "nil".
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if o.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(*o.value)
}

// UnmarshalJSON unmarshals the JSON-encoded data into the underlying value of an [Optional]. If the
// operation succeeds, the [Optional] is considered to be set.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	var dest T
	if err := json.Unmarshal(data, &dest); err != nil {
		return err
	}
	o.value = &dest
	o.isSet = true
	return nil
}

func ptrTo[T any](t T) *T {
	return &t
}

