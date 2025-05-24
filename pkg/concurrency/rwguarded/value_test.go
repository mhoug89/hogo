package rwguarded

import (
	"errors"
	"sync"
	"testing"
)

func TestValueSetAndGet(t *testing.T) {
	t.Parallel()

	initialStr := "initial-value"
	rwgVal := New[string](initialStr)

	// Initial Get should return the value passed to the New func:
	var got, want string
	got, want = rwgVal.Get(), initialStr
	if got != want {
		t.Fatalf("Get() got %q, want %q", got, want)
	}

	// Set a different value, then make sure we get the new value back:
	newStr := "new-value"
	rwgVal.Set(newStr)
	got, want = rwgVal.Get(), newStr
	if got != want {
		t.Fatalf("Get() got %q, want %q", got, want)
	}
}

func TestValueConcurrentOpsNoPanic(t *testing.T) {
	t.Parallel()

	rwgVal := New[string]("")

	type Op struct {
		name string
		do   func()
	}

	// One item for each operation
	items := []Op{
		{
			name: "Get",
			do: func() {
				_ = rwgVal.Get()
			},
		},
		{
			name: "Set",
			do: func() {
				rwgVal.Set("value")
			},
		},
		{
			name: "Update",
			do: func() {
				rwgVal.Update(func(val *string) error {
					*val = "updated-value"
					return nil
				})
			},
		},
	}

	routinesPerItem := len(items) * 24

	// Check that doing concurrent operations doesn't cause a panic.
	routines := make([]func(), 0, len(items)*routinesPerItem)
	wg := sync.WaitGroup{}
	// For a better chance of the routines running at the same time, define them
	// all first before running them.
	for itemNum := 0; itemNum < len(items); itemNum++ {
		for i := 0; i < routinesPerItem; i++ {
			wrapperFn := func() {
				items[itemNum].do()
				wg.Done()
			}
			wg.Add(1)
			routines = append(routines, wrapperFn)
		}
	}

	for i := 0; i < len(routines); i++ {
		go routines[i]()
	}
	wg.Wait()
}

func TestValueUpdate(t *testing.T) {
	t.Parallel()

	errUpdaterFailed := errors.New("updater failed")

	for _, tc := range []struct {
		name    string
		origVal int
		updater func(val *int) error
		wantVal int
		wantErr error
	}{
		{
			name:    "update_ok",
			origVal: 1,
			updater: func(val *int) error {
				*val++
				return nil
			},
			wantVal: 2,
			wantErr: nil,
		},
		{
			name:    "update_error",
			origVal: 1,
			updater: func(val *int) error {
				return errUpdaterFailed
			},
			wantVal: 1,
			wantErr: errUpdaterFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rwgVal := New[int](tc.origVal)
			if err := rwgVal.Update(tc.updater); err != tc.wantErr {
				t.Errorf("Update() got error %q, want error %q", err, tc.wantErr)
			}
			if gotVal := rwgVal.Get(); gotVal != tc.wantVal {
				t.Fatalf("Get() got %q, want %q", gotVal, tc.wantVal)
			}
		})
	}
}

