// Package set provides a Set type and associated functionality.
package set

type emptyStruct struct{}

// Set is a collection of items with no duplicates, i.e. no two items compare equal to each other.
type Set[T comparable] interface {
	// Add adds the provided items to the set.
	Add(items ...T)

	// AddIfAbsent adds the provided item to the set if it doesn't already exist. Returns false
	// if the item already exists in the set.
	AddIfAbsent(item T) bool

	// Clear removes all items from the set.
	Clear()

	// Delete removes the provided items from the set.
	Delete(items ...T)

	// DeleteIfPresent removes the provided item from the set if it exists. Returns false if the
	// item did not exist in the set.
	DeleteIfPresent(item T) bool

	// Equal returns whether two sets contain the same items. This is true iff the sets are the
	// same length and every item in one set is found via Has in the other set.
	Equal(other Set[T]) bool

	// Has returns whether the provided item is in the set.
	Has(item T) bool

	// HasAll returns whether all of the provided items are in the set.
	HasAll(item ...T) bool

	// IsEmpty returns whether the set contains 0 items.
	IsEmpty() bool

	// Len returns the size of the set.
	Len() int

	// Pop removes an element from the set, if the set is not empty, and returns it. If the set is
	// empty, the boolean return value will be false, and the first return value will be the zero
	// value of the type stored in the set.
	Pop() (T, bool)

	// ToSlice returns a slice containing all the items in the Set.
	ToSlice() []T

	// TODO: Consider adding these operations to the Set API:
	/*
		- Difference(s2 Set[T]) -> Set[T]
		- Intersection(s2 Set[T]) -> Set[T]
		- Union(s2 Set[T]) -> Set[T]
		- Update(s2 Set[T]) -> Set[T]
		- IsSubsetOf(s2 Set[T]) -> bool
	*/
}

type mapSet[T comparable] map[T]emptyStruct

// Verify interface compliance:
var _ Set[int] = (mapSet[int])(nil)

func (s mapSet[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = emptyStruct{}
	}
}

func (s mapSet[T]) AddIfAbsent(item T) bool {
	if s.Has(item) {
		return false
	}
	s[item] = emptyStruct{}
	return true
}

func (s mapSet[T]) Clear() {
	clear(s)
}

func (s mapSet[T]) Delete(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s mapSet[T]) DeleteIfPresent(item T) bool {
	if !s.Has(item) {
		return false
	}
	delete(s, item)
	return true
}

func (s mapSet[T]) Equal(other Set[T]) bool {
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

func (s mapSet[T]) Has(item T) bool {
	_, found := s[item]
	return found
}

func (s mapSet[T]) HasAll(items ...T) bool {
	for _, item := range items {
		if _, found := s[item]; !found {
			return false
		}
	}
	return true
}

func (s mapSet[T]) IsEmpty() bool {
	return len(s) == 0
}

func (s mapSet[T]) Len() int {
	return len(s)
}

func (s mapSet[T]) Pop() (T, bool) {
	for item := range s {
		delete(s, item)
		return item, true
	}
	var tZero T
	return tZero, false
}

func (s mapSet[T]) ToSlice() []T {
	items := make([]T, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

// New returns a new Set backed by Go's native [map].
func New[T comparable](items ...T) Set[T] {
	s := make(mapSet[T], len(items))
	s.Add(items...)
	return s
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
