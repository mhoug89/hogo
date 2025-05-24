package rwguarded

import (
	"errors"
	"sync"
)

// ErrUpdateKeyNotFound is returned when the key is not found in the map during an update
// operation.
var ErrUpdateKeyNotFound = errors.New("key not found")

// Map is a thin wrapper around a map that uses a [sync.RWMutex] to synchronize operations. This
// struct should not be directly instantiated; callers should use the [NewMap] function instead.
type Map[K comparable, V any] struct {
	rwLock     *sync.RWMutex
	valueByKey map[K]V
}

// NewMap initializes and returns a [Map] of the provided types.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		rwLock:     &sync.RWMutex{},
		valueByKey: make(map[K]V),
	}
}

// Clear clears the underlying map by creating a new one.
func (m *Map[K, V]) Clear() {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	// Since the underlying map is not exported and thus nothing should be keeping a reference to
	// it, we can just make a new one and let the old one get garbage collected.
	m.valueByKey = make(map[K]V)
}

// Count returns the number of items in the underlying map.
func (m *Map[K, V]) Count() int {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	return len(m.valueByKey)
}

// Delete deletes the item(s) at the provided key(s) from the underlying map.
func (m *Map[K, V]) Delete(keys ...K) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	for _, k := range keys {
		delete(m.valueByKey, k)
	}
}

// Load returns the value associated with the provided key from the underlying map. If the key
// did not exist, the boolean return value will be false.
func (m *Map[K, V]) Load(key K) (V, bool) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	value, ok := m.valueByKey[key]
	return value, ok
}

// Store adds an item to the underlying map with the provided key and value.
func (m *Map[K, V]) Store(key K, value V) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	m.valueByKey[key] = value
}

// StoreIfAbsent checks if the given key exists in the map, and if not, executes the given function
// to obtain the value to store at that key. This method accepts a function that produces the
// desired value so that it can skip the potentially expensive operation of creating the value if
// the value should not be added to the map.
//
// Note that in scenarios where multiple routines are calling StoreIfAbsent in parallel for the same
// key, it's possible for valueCtor to be called by all the routines, but only the first routine
// that succeeds in obtaining the underlying writer lock will write its value to the map at the
// given key; the other constructed values will be discarded.
//
// For the boolean return value, this method returns true if the value was successfully constructed
// and added. Otherwise, it returns false, and the reason for not inserting the value can be
// determined by the returned error - if nil, the key was already present in the map; if non-nil,
// the key was not present, but the function to construct the new value returned an error.
func (m *Map[K, V]) StoreIfAbsent(key K, valueCtor func() (*V, error)) (bool, error) {
	// Try checking with only a reader lock first, as this is less expensive than obtaining a writer
	// lock when the key already exists.
	m.rwLock.RLock()
	_, found := m.valueByKey[key]
	m.rwLock.RUnlock()
	if found {
		return false, nil
	}

	// Since the key does not exist in the map, we should create the value to be stored at the key.
	// Note that we MUST do this before obtaining the writer lock, as we don't want to allow users
	// to deadlock by calling StoreIfAbsent with a value constructor that also calls StoreIfAbsent.
	// The potential downside is that we may end up creating the value but not using it, but this is
	// acceptable because it's more important to prevent the aforementioned possibility of deadlock.
	valPtr, err := valueCtor()
	if err != nil {
		return false, err
	}

	// Obtain the writer lock, check again if the key exists (because another process could have set
	// the value between when we released the reader lock and now), and if not, set the value.
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if _, found := m.valueByKey[key]; found {
		return false, nil
	}
	m.valueByKey[key] = *valPtr
	return true, nil
}

// Update fetches an existing item from the map, then calls the provided updater function and stores
// the new value at the provided key.
//
// If the provided key was not found, or the updater function fails, this method returns an error.
func (m *Map[K, V]) Update(key K, updater func(V) (V, error)) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	gotVal, ok := m.valueByKey[key]
	if !ok {
		return ErrUpdateKeyNotFound
	}

	gotVal, err := updater(gotVal)
	if err != nil {
		return err
	}
	m.valueByKey[key] = gotVal
	return nil
}

