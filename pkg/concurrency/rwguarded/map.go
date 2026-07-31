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

	m.clearWithoutLocking()
}

// ClearWithCleanup performs the provided cleanup function on each item in the map, then clears the
// underlying map by creating a new one. It accumulates the errors from each cleanup function call
// using [errors.Join] and returns the result.
func (m *Map[K, V]) ClearWithCleanup(cleanupFn func(k K, v V) error) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	var errs []error
	for k, v := range m.valueByKey {
		if err := cleanupFn(k, v); err != nil {
			errs = append(errs, err)
		}
	}
	m.clearWithoutLocking()

	return errors.Join(errs...)
}

// Len returns the number of items in the underlying map.
func (m *Map[K, V]) Len() int {
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

// DeleteIfMatchWithCleanup is like [Map.DeleteWithCleanup], but calls the provided matcher function
// on each key-value pair, and only performs cleanup and deletion if the matcher returns true. If no
// value is present at the provided key, this is treated as if the matcher returned false.
//
// This is useful for scenarios where many callers fetch the value at a provided key, discover it is
// in a bad state (e.g., a client's connection to its remote host may have been broken), and race to
// replace the value for a provided key. If the value should only be replaced once, the first
// caller's matcher (which might, for example, compare the address of the pointer stored at the
// provided key) would be the only one to return true, ensuring the value is only deleted and
// cleaned up once.
func (m *Map[K, V]) DeleteIfMatchWithCleanup(
	matchFn func(k K, v V) bool,
	cleanupFn func(k K, v V) error,
	keys ...K,
) (int, error) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	deletedCount := 0
	var errs []error
	for _, k := range keys {
		v, ok := m.valueByKey[k]
		if !ok || !matchFn(k, v) {
			continue
		}
		if err := cleanupFn(k, v); err != nil {
			errs = append(errs, err)
		}
		delete(m.valueByKey, k)
		deletedCount++
	}

	return deletedCount, errors.Join(errs...)
}

// DeleteWithCleanup calls the provided cleanup function on the item at the provided key, then
// deletes the item at the provided key from the underlying map; it repeats this for each provided
// key. It accumulates the errors from each cleanup function call using [errors.Join] and returns
// the result.
func (m *Map[K, V]) DeleteWithCleanup(cleanupFn func(k K, v V) error, keys ...K) error {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	var errs []error
	for _, k := range keys {
		if err := cleanupFn(k, m.valueByKey[k]); err != nil {
			errs = append(errs, err)
		}
		delete(m.valueByKey, k)
	}

	return errors.Join(errs...)
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

type storeIfAbsentOptions struct {
	UseWriteLockForValueCtor bool
}

// StoreIfAbsentOption is used to specify predefined options for [Map.StoreIfAbsent].
type StoreIfAbsentOption func(opt *storeIfAbsentOptions)

// UseWriteLockForValueCtor causes [Map.StoreIfAbsent] to obtain the underlying writer lock when calling
// the value constructor, rather than the default approach of calling the value constructor without
// any lock being held. It will still perform an initial read-locked check to avoid obtaining the
// writer lock unless absolutely necessary. If no value exists at the provided key, it will obtain
// the write lock, check for the value again, and if another thread has assigned a value at the
// provided key between the read-locked check and now, it will not call the value constructor.
//
// This ensures that the value constructor for a provided key is only called once. This is useful
// when the value being created requires cleanup, and it's unacceptable to call the constructor
// multiple times but only keep a reference to one of the created values.
//
// WARNING: When using this option, the provided value constructor function must NOT call any other
// methods of this Map, as they will fail to obtain the underlying lock and will deadlock.
func UseWriteLockForValueCtor() StoreIfAbsentOption {
	return func(opt *storeIfAbsentOptions) {
		opt.UseWriteLockForValueCtor = true
	}
}

// StoreIfAbsent checks if the provided key exists in the map, and if not, executes the provided
// function to obtain the value to store at that key. This method accepts a function that produces
// the desired value so that it can skip the potentially expensive operation of creating the value
// if the value should not be added to the map.
//
// Note that by default, in scenarios where multiple routines are calling StoreIfAbsent in parallel
// for the same key, it's possible for valueCtor to be called by all the routines, but only the
// first routine that succeeds in obtaining the underlying writer lock will write its value to the
// map at the provided key; the other constructed values will be discarded. To avoid this behavior
// and ensure the value constructor for the provided key is only called once, specify the
// [UseWriteLockForValueCtor] option.
//
// For the boolean return value, this method returns true if the value was successfully constructed
// and added. Otherwise, it returns false, and the reason for not inserting the value can be
// determined by the returned error - if nil, the key was already present in the map; if non-nil,
// the key was not present, but the function to construct the new value returned an error.
func (m *Map[K, V]) StoreIfAbsent(
	key K,
	valueCtor func() (*V, error),
	opts ...StoreIfAbsentOption,
) (bool, error) {
	options := storeIfAbsentOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	// Use the writer lock variant if specified.
	if options.UseWriteLockForValueCtor {
		return m.storeIfAbsentWithWriteLockForValueCtor(key, valueCtor)
	}

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

func (m *Map[K, V]) clearWithoutLocking() {
	// Since the underlying map is not exported and thus nothing should be keeping a reference to
	// it, we can just make a new one and let the old one get garbage collected.
	m.valueByKey = make(map[K]V)
}

// storeIfAbsentWithWriteLockForValueCtor is like [Map.StoreIfAbsent], but holds the writer lock
// when calling the value constructor. See [UseWriteLockForValueCtor].
func (m *Map[K, V]) storeIfAbsentWithWriteLockForValueCtor(
	key K,
	valueCtor func() (*V, error),
) (bool, error) {
	// Try checking with only a reader lock first, as this is less expensive than obtaining a writer
	// lock when the key already exists.
	m.rwLock.RLock()
	if _, found := m.valueByKey[key]; found {
		m.rwLock.RUnlock()
		return false, nil
	}
	m.rwLock.RUnlock()

	// If not found, obtain the writer lock, check again if the key exists (because another process
	// could have set the value between when we released the reader lock and now), and if not, set
	// the value.
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if _, found := m.valueByKey[key]; found {
		return false, nil
	}

	valPtr, err := valueCtor()
	if err != nil {
		return false, err
	}
	m.valueByKey[key] = *valPtr
	return true, nil
}

