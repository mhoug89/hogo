// Package hashset provides a HashSet type and associated functionality.
package hashset

import (
	"iter"
	"maps"
)

type emptyStruct = struct{}

// New returns a new [HashSet].
func New[T comparable](items ...T) HashSet[T] {
	s := make(HashSet[T], len(items))
	s.Add(items...)
	return s
}

// HashSet is a set backed by Go's native [map].
type HashSet[T comparable] map[T]emptyStruct

// Verify interface compliance:
var _ set[string] = (HashSet[string])(nil)

// Add adds the provided items to the set.
func (s HashSet[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = emptyStruct{}
	}
}

// AddIfAbsent adds the provided item to the set if it doesn't already exist. Returns false if the
// item already exists in the set.
func (s HashSet[T]) AddIfAbsent(item T) bool {
	if s.Has(item) {
		return false
	}
	s[item] = emptyStruct{}
	return true
}

// Clear removes all items from the set.
func (s HashSet[T]) Clear() {
	clear(s)
}

// Delete removes the provided items from the set.
func (s HashSet[T]) Delete(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

// DeleteIfPresent removes the provided item from the set if it exists. Returns false if the item
// does not exist in the set.
func (s HashSet[T]) DeleteIfPresent(item T) bool {
	if !s.Has(item) {
		return false
	}
	delete(s, item)
	return true
}

// Equal returns whether two sets contain the same items. This is true iff the sets are the same
// length and every item in one set is found via Has in the other set.
func (s HashSet[T]) Equal(other set[T]) bool {
	if s.Len() != other.Len() {
		return false
	}
	for item := range s {
		if !other.Has(item) {
			return false
		}
	}
	return true
}

// Has returns whether the provided item is in the set.
func (s HashSet[T]) Has(item T) bool {
	_, found := s[item]
	return found
}

// HasAll returns whether all of the provided items are in the set.
func (s HashSet[T]) HasAll(items ...T) bool {
	for _, item := range items {
		if _, found := s[item]; !found {
			return false
		}
	}
	return true
}

// IsEmpty returns whether the set contains 0 items.
func (s HashSet[T]) IsEmpty() bool {
	return len(s) == 0
}

// Len returns the size of the set.
func (s HashSet[T]) Len() int {
	return len(s)
}

// Pop removes an element from the set, if the set is not empty, and returns it. If the set is
// empty, the boolean return value will be false, and the first return value will be the zero value
// of the type stored in the set.
func (s HashSet[T]) Pop() (T, bool) {
	for item := range s {
		delete(s, item)
		return item, true
	}
	var tZero T
	return tZero, false
}

// Iter returns an iterator over the items in the set.
func (s HashSet[T]) Iter() iter.Seq[T] {
	return maps.Keys(s)
}

// ToSlice returns a slice containing all the items in the set.
func (s HashSet[T]) ToSlice() []T {
	items := make([]T, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

// Update adds to the set all items from all the provided sets.
func (s HashSet[T]) Update(others ...Iterable[T]) {
	for _, other := range others {
		for item := range other.Iter() {
			s.Add(item)
		}
	}
}

// TODO: Maybe utilize a builder pattern for additional options when creating a new set?
// The New function satisfies most general cases, but callers might also want to be able to specify
// an initial capacity, supply items differently (inline via vardiadic args, referencing an
// existing slice without expanding it to variadic args, or from an existing Set), etc. E.g.:
//
//     mySlice := []string{"x", "y", "z"}
//     mySet := set.New("1", "2", "3")
//     myNewSet := set.NewBuilder().
//         WithInitialCapacity(256).
//         WithItems("a", "b", "c").
//         WithItemsFromSet(mySet).
//         WithItemsFromSlice(mySlice).
//         Build()
