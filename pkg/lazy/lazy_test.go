package lazy

import (
	"errors"
	"testing"
)

// mutator holds a value that changes after it's returned. This is used to ensure that a lazy-loader
// only calls the underlying load function once.
type mutator struct {
	value string
}

func (m *mutator) Value() string {
	ret := m.value
	m.value += "a"
	return ret
}

func TestLoad(t *testing.T) {
	t.Parallel()

	errLoadFailed := errors.New("failed to load")
	strToLoad := "some big expensive string"

	for _, tc := range []struct {
		name        string
		wantLoadVal string
		wantLoadErr error
	}{
		{
			name:        "load_ok",
			wantLoadVal: strToLoad,
			wantLoadErr: nil,
		},
		{
			name:        "load_fail",
			wantLoadVal: "",
			wantLoadErr: errLoadFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := mutator{value: tc.wantLoadVal}
			l := New(func() (string, error) {
				return m.Value(), tc.wantLoadErr
			})
			if l.value != nil {
				t.Fatalf("Lazy instance's initial value was %q, want nil ptr", *l.value)
			}

			gotVal, gotErr := l.Load()
			if tc.wantLoadVal != gotVal {
				t.Errorf("Load() value mismatch; got %q, want %q", gotVal, tc.wantLoadVal)
			}
			if tc.wantLoadErr != gotErr {
				t.Errorf("Load() error mismatch; got %v, want %v", gotErr, tc.wantLoadErr)
			}
			if t.Failed() {
				return
			}

			// Load value again to ensure the underlying loader doesn't get called twice.
			gotVal, gotErr = l.Load()
			if tc.wantLoadVal != gotVal {
				t.Errorf("Second Load() value mismatch; got %q, want %q", gotVal, tc.wantLoadVal)
			}
			if tc.wantLoadErr != gotErr {
				t.Errorf("Second Load() error mismatch; got %v, want %v", gotErr, tc.wantLoadErr)
			}
		})
	}
}

func TestMustLoad(t *testing.T) {
	t.Parallel()

	wantLoadVal := "some big expensive string"

	m := mutator{value: wantLoadVal}
	ml := NewMust(func() string {
		return m.Value()
	})
	if ml.value != nil {
		t.Fatalf("Lazy instance's initial value was %q, want nil ptr", *ml.value)
	}

	gotVal := ml.Load()
	if wantLoadVal != gotVal {
		t.Fatalf("Load() value mismatch; got %q, want %q", gotVal, wantLoadVal)
	}

	// Load value again to ensure the underlying loader doesn't get called twice.
	gotVal = ml.Load()
	if wantLoadVal != gotVal {
		t.Fatalf("Second Load() value mismatch; got %q, want %q", gotVal, wantLoadVal)
	}
}
