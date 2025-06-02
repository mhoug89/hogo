package hashset

import (
	"iter"
)

// Iterable is the interface for any type that implements [iter.Seq] via a method named Iter.
type Iterable[T comparable] interface {
	Iter() iter.Seq[T]
}

// A set is a collection of items with no duplicates, i.e. no two items compare equal to each other.
//
// TODO: If/when adding other types of sets, export this and make sure the other set types implement
// this interface.
type set[T comparable] interface {
	Add(items ...T)
	AddIfAbsent(item T) bool
	Clear()
	Delete(items ...T)
	DeleteIfPresent(item T) bool
	Equal(other set[T]) bool
	Has(item T) bool
	HasAll(item ...T) bool
	IsEmpty() bool
	Iter() iter.Seq[T]
	Len() int
	Pop() (T, bool)
	ToSlice() []T
	Update(others ...Iterable[T])

	// TODO: Consider adding these operations to the Set API:
	/*
		- Difference(s2 Set[T]) -> Set[T]
		- Intersection(s2 Set[T]) -> Set[T]
		- Union(s2 Set[T]) -> Set[T]
		- IsSubsetOf(s2 Set[T]) -> bool
	*/
}
