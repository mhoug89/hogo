package rwguarded

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestMapSerialStoreThenLoad(t *testing.T) {
	t.Parallel()

	rwgMap := NewMap[string, string]()
	key := "k"
	val := "v"

	rwgMap.Store(key, val)
	var got string
	var ok bool
	got, ok = rwgMap.Load(key)
	if !ok {
		t.Fatalf("Load(%q) did not find entry, but should have", key)
	}
	if got != val {
		t.Fatalf("Load(%q) got %q, want %q", key, got, val)
	}
}

func TestMapDelete(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		keysToDelete []string
		wantKeysLeft []string
	}{
		{
			name:         "0_args",
			keysToDelete: []string{},
			wantKeysLeft: []string{"a", "b", "c"},
		},
		{
			name:         "1_arg",
			keysToDelete: []string{"a"},
			wantKeysLeft: []string{"b", "c"},
		},
		{
			name:         "multiple_args",
			keysToDelete: []string{"a", "b"},
			wantKeysLeft: []string{"c"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rwgMap := NewMap[string, string]()
			rwgMap.Store("a", "filler-value")
			rwgMap.Store("b", "filler-value")
			rwgMap.Store("c", "filler-value")

			rwgMap.Delete(tc.keysToDelete...)

			for _, k := range tc.keysToDelete {
				if _, ok := rwgMap.Load(k); ok {
					t.Errorf("Load(%q) found entry after Delete(), but should not have", k)
				}
			}
			for _, k := range tc.wantKeysLeft {
				if _, ok := rwgMap.Load(k); !ok {
					t.Errorf("Load(%q) did not find entry, but should have", k)
				}
			}
		})
	}
}

func TestMapLenViaStoreDeleteAndClear(t *testing.T) {
	t.Parallel()

	rwgMap := NewMap[string, string]()
	var got, want int
	got, want = rwgMap.Len(), 0
	if got != want {
		t.Fatalf("Len() got %d, want %d", got, want)
	}

	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"

	rwgMap.Store(k1, v1)
	got, want = rwgMap.Len(), 1
	if got != want {
		t.Fatalf("Len() got %d, want %d", got, want)
	}

	rwgMap.Store(k2, v2)
	rwgMap.Store(k3, v3)
	got, want = rwgMap.Len(), 3
	if got != want {
		t.Fatalf("After Store(), Len() got %d, want %d", got, want)
	}

	rwgMap.Delete(k1)
	got, want = rwgMap.Len(), 2
	if got != want {
		t.Fatalf("After Delete(), Len() got %d, want %d", got, want)
	}

	rwgMap.Clear()
	got, want = rwgMap.Len(), 0
	if got != want {
		t.Fatalf("After Clear(), Len() got %d, want %d", got, want)
	}
}

func TestMapConcurrentOpsNoPanic(t *testing.T) {
	t.Parallel()

	type MapItem struct {
		key   string
		value string
	}

	type ItemOpPair struct {
		item MapItem
		op   func(m *Map[string, string], item MapItem)
	}

	// One item for each operation
	itemOpPairs := []ItemOpPair{
		{
			item: MapItem{
				key:   "Clear",
				value: "Clear()",
			},
			op: func(m *Map[string, string], _ MapItem) {
				m.Clear()
			},
		},
		{
			item: MapItem{
				key:   "ClearWithCleanup",
				value: "ClearWithCleanup()",
			},
			op: func(m *Map[string, string], _ MapItem) {
				cleanupFn := func(_ string, _ string) error {
					return nil
				}
				_ = m.ClearWithCleanup(cleanupFn)
			},
		},
		{
			item: MapItem{
				key:   "Delete",
				value: "Delete()",
			},
			op: func(m *Map[string, string], item MapItem) {
				m.Delete(item.key)
			},
		},
		{
			item: MapItem{
				key:   "DeleteWithCleanup",
				value: "DeleteWithCleanup()",
			},
			op: func(m *Map[string, string], item MapItem) {
				cleanupFn := func(_ string, _ string) error {
					return nil
				}
				_ = m.DeleteWithCleanup(cleanupFn, item.key)
			},
		},
		{
			item: MapItem{
				key:   "Len",
				value: "Len()",
			},
			op: func(m *Map[string, string], _ MapItem) {
				_ = m.Len()
			},
		},
		{
			item: MapItem{
				key:   "Load",
				value: "Load()",
			},
			op: func(m *Map[string, string], item MapItem) {
				_, _ = m.Load(item.key)
			},
		},
		{
			item: MapItem{
				key:   "Store",
				value: "Store()",
			},
			op: func(m *Map[string, string], item MapItem) {
				m.Store(item.key, item.value)
			},
		},
		{
			item: MapItem{
				key:   "Update",
				value: "Update()",
			},
			op: func(m *Map[string, string], item MapItem) {
				updateFn := func(v string) (string, error) {
					return strings.ToUpper(item.value), nil
				}
				_ = m.Update(item.key, updateFn)
			},
		},
	}

	rwgMap := NewMap[string, string]()
	randGenSrc := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randGenSrc)
	// Check that doing concurrent operations, both on the same and different keys, doesn't cause a
	// panic.
	wg := sync.WaitGroup{}
	// Because each operation is simple and likely to finish quickly, we do each op several times
	// for each item to make it more likely that the goroutines will run concurrently.
	iterationsPerItem := 10000
	for range itemOpPairs {
		wg.Go(func() {
			for range iterationsPerItem {
				randomPair := ptrTo(itemOpPairs[randGen.Intn(len(itemOpPairs))])
				randomPair.op(rwgMap, randomPair.item)
			}
		})
	}
	wg.Wait()
}

func TestMapNestedStoreIfAbsentCallsDoesNotDeadlock_SameKey(t *testing.T) {
	t.Parallel()

	rwgMap := NewMap[string, *string]()
	key := "key1"
	value := ptrTo("value1")
	thrownAwayValue := ptrTo("value2")

	var innerAdded, outerAdded bool

	outerAdded, _ = rwgMap.StoreIfAbsent(
		key,
		func() (**string, error) {
			// Since this inner StoreIfAbsent call finishes before the outer one, the value returned
			// from this inner function should be the one stored at our specified key.
			innerAdded, _ = rwgMap.StoreIfAbsent(
				key,
				func() (**string, error) {
					return &value, nil
				},
			)
			return &thrownAwayValue, nil
		},
	)

	if outerAdded {
		t.Errorf("Outer StoreIfAbsent() added item, but should not have")
	}
	if !innerAdded {
		t.Errorf("Inner StoreIfAbsent() did not add item, but should have")
	}
	gotLen := rwgMap.Len()
	wantLen := 1
	if gotLen != wantLen {
		t.Errorf("Len() got %d, want %d", gotLen, wantLen)
	}
	if got, ok := rwgMap.Load(key); !ok || got != value {
		t.Errorf("Load(%q) got %v, want %v", key, got, value)
	}
}

func TestMapNestedStoreIfAbsentCallsDoesNotDeadlock_DistinctKeys(t *testing.T) {
	t.Parallel()

	rwgMap := NewMap[string, *string]()
	innerKey := "innerKey"
	outerKey := "outerKey"
	innerValue := ptrTo("innerValue")
	outerValue := ptrTo("outerValue")

	var innerAdded, outerAdded bool

	outerAdded, _ = rwgMap.StoreIfAbsent(
		outerKey,
		func() (**string, error) {
			// Since this inner StoreIfAbsent call finishes before the outer one, the value returned
			// from this inner function should be the one stored at our specified key.
			innerAdded, _ = rwgMap.StoreIfAbsent(
				innerKey,
				func() (**string, error) {
					return &innerValue, nil
				},
			)
			return &outerValue, nil
		},
	)

	if !outerAdded {
		t.Errorf("Outer StoreIfAbsent() did not add item, but should have")
	}
	if !innerAdded {
		t.Errorf("Inner StoreIfAbsent() did not add item, but should have")
	}
	gotLen := rwgMap.Len()
	wantLen := 2
	if gotLen != wantLen {
		t.Errorf("Len() got %d, want %d", gotLen, wantLen)
	}
	if got, ok := rwgMap.Load(outerKey); !ok || got != outerValue {
		t.Errorf("Load(%q) got value %v, want %v", outerKey, got, outerValue)
	}
	if got, ok := rwgMap.Load(innerKey); !ok || got != innerValue {
		t.Errorf("Load(%q) got value %v, want %v", innerKey, got, innerValue)
	}
}

func TestMapUpdateOK(t *testing.T) {
	t.Parallel()

	rwgMap := NewMap[string, string]()
	key := "key1"
	value := "value1"
	rwgMap.Store(key, value)

	newValue := "newValue"
	err := rwgMap.Update(key, func(v string) (string, error) {
		return newValue, nil
	})

	if err != nil {
		t.Fatalf("Update() failed with error %v", err)
	}
	if got, ok := rwgMap.Load(key); !ok || got != newValue {
		t.Errorf("Load(%q) got value %q, want %q", key, got, newValue)
	}
}

func TestMapUpdateError(t *testing.T) {
	t.Parallel()

	errMutatingValue := errors.New("could not load new value")
	origValue := "originalValue"
	newValue := "newValue"

	for _, tc := range []struct {
		name        string
		existingKey string
		keyToUpdate string
		updaterErr  error
		want        error
	}{
		{
			name:        "missing_key",
			existingKey: "key1",
			keyToUpdate: "key404",
			updaterErr:  nil,
			want:        ErrUpdateKeyNotFound,
		},
		{
			name:        "update_fn_err",
			existingKey: "key1",
			keyToUpdate: "key1",
			updaterErr:  errMutatingValue,
			want:        errMutatingValue,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rwgMap := NewMap[string, string]()
			rwgMap.Store(tc.existingKey, origValue)

			err := rwgMap.Update(tc.keyToUpdate, func(v string) (string, error) {
				if tc.updaterErr == nil {
					return newValue, nil
				}
				return "", tc.updaterErr
			})

			if err == nil {
				t.Fatalf("Update() did not fail, but should have")
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("Update() failed with wrong error; got %q, want %q", err, tc.want)
			}
			if tc.existingKey == tc.keyToUpdate {
				if gotValue, ok := rwgMap.Load(tc.existingKey); !ok || gotValue != origValue {
					t.Errorf("Update() should not have changed existing value if error was returned; got %v, want %v", gotValue, origValue)
				}
			}
		})
	}
}

func TestMapClearWithCleanup(t *testing.T) {
	t.Parallel()

	type Item struct {
		UID       string
		CleanedUp bool
	}

	rwgMap := NewMap[string, *Item]()
	items := []*Item{
		ptrTo(Item{UID: "item1"}),
		ptrTo(Item{UID: "item2"}),
		ptrTo(Item{UID: "item3"}),
	}
	for _, item := range items {
		rwgMap.Store(item.UID, item)
	}

	err := rwgMap.ClearWithCleanup(func(k string, v *Item) error {
		v.CleanedUp = true
		return fmt.Errorf("err from uid %s", v.UID)
	})

	// Check that the map is now empty.
	if rwgMap.Len() != 0 {
		t.Errorf("Len() = %d, want 0", rwgMap.Len())
	}

	// Check that a non-nil error was returned.
	if err == nil {
		t.Fatalf("ClearWithCleanup() = nil; failed to return cleanup errors")
	}

	for _, item := range items {
		// Check that each item's cleanup function was called.
		if !item.CleanedUp {
			t.Errorf("item %q was not cleaned up", item.UID)
		}
		// Check that each item's cleanup error message was present in the wrapped error.
		wantMsg := fmt.Sprintf("err from uid %s", item.UID)
		if !strings.Contains(err.Error(), wantMsg) {
			t.Errorf("ClearWithCleanup did not include error from item %q:\nGot: %s\nWant: %s", item.UID, err.Error(), wantMsg)
		}
	}
}

func TestMapDeleteWithCleanup(t *testing.T) {
	t.Parallel()

	type Item struct {
		UID       string
		CleanedUp bool
	}

	rwgMap := NewMap[string, *Item]()
	item := &Item{UID: "item1"}
	rwgMap.Store(item.UID, item)

	cleanupFn := func(k string, v *Item) error {
		v.CleanedUp = true
		return fmt.Errorf("err from uid %s", v.UID)
	}
	err := rwgMap.DeleteWithCleanup(cleanupFn, item.UID)

	// Check that a non-nil error was returned.
	if err == nil {
		t.Fatalf("DeleteWithCleanup() = nil; failed to return cleanup error")
	}

	// Check that the item's cleanup function was called.
	if !item.CleanedUp {
		t.Errorf("item %q was not cleaned up", item.UID)
	}

	// Check that the item's cleanup error message was present in the wrapped error.
	wantMsg := fmt.Sprintf("err from uid %s", item.UID)
	if !strings.Contains(err.Error(), wantMsg) {
		t.Errorf("DeleteWithCleanup did not include error from item %q:\nGot: %s\nWant: %s", item.UID, err.Error(), wantMsg)
	}
}

