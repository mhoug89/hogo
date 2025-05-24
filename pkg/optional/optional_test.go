package optional

import (
	"errors"
	"testing"
)

func TestOfViaGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wantValue int
	}{
		{
			name:      "of_zero_value",
			wantValue: 0,
		},
		{
			name:      "of_nonzero_value",
			wantValue: 10,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := Of(tc.wantValue)
			got, err := o.Get()
			if err != nil {
				t.Errorf("Get() returned an error: %v", err)
			}
			if got != tc.wantValue {
				t.Errorf("Get() got %v, want: %v", got, tc.wantValue)
			}
		})
	}
}

func TestSetViaGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opt       Optional[int]
		setter    func(*Optional[int])
		wantValue int
	}{
		{
			name: "set_on_empty",
			opt:  Empty[int](),
			setter: func(o *Optional[int]) {
				o.Set(1)
			},
			wantValue: 1,
		},
		{
			name: "set_on_nonempty_same_value",
			opt:  Of[int](1),
			setter: func(o *Optional[int]) {
				o.Set(1)
			},
			wantValue: 1,
		},
		{
			name: "set_on_nonempty_different_value",
			opt:  Of[int](9001),
			setter: func(o *Optional[int]) {
				o.Set(1)
			},
			wantValue: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.setter(&tc.opt)
			got, err := tc.opt.Get()
			if err != nil {
				t.Errorf("Get() returned an error: %v", err)
			}
			if got != tc.wantValue {
				t.Errorf("Get() got %v, want: %v", got, tc.wantValue)
			}
		})
	}
}

func TestEmptinessFuncs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opt         func() *Optional[int]
		wantIsSet   bool
		wantIsEmpty bool
	}{
		{
			name: "empty",
			opt: func() *Optional[int] {
				o := Empty[int]()
				return &o
			},
			wantIsEmpty: true,
		},
		{
			name: "of_zero_value",
			opt: func() *Optional[int] {
				o := Of[int](0)
				return &o
			},
			wantIsSet: true,
		},
		{
			name: "empty_then_set_with_zero_value",
			opt: func() *Optional[int] {
				o := Empty[int]()
				o.Set(0)
				return &o
			},
			wantIsSet: true,
		},
		{
			name: "of_nonzero_value",
			opt: func() *Optional[int] {
				o := Of[int](10)
				return &o
			},
			wantIsSet: true,
		},
		{
			name: "empty_then_set_with_nonzero_value",
			opt: func() *Optional[int] {
				o := Empty[int]()
				o.Set(0)
				return &o
			},
			wantIsSet: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opt := tc.opt()
			if got := opt.IsSet(); got != tc.wantIsSet {
				t.Errorf("IsSet() got %v, want: %v", got, tc.wantIsSet)
			}
			if got := opt.IsEmpty(); got != tc.wantIsEmpty {
				t.Errorf("IsEmpty() got %v, want: %v", got, tc.wantIsEmpty)
			}
		})
	}
}

func TestMustGet_OK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		opt        Optional[int]
		wantValue  int
		shouldFail bool
	}{
		{
			name:      "of_zero_value",
			opt:       Of[int](0),
			wantValue: 0,
		},
		{
			name:      "of_nonzero_value",
			opt:       Of[int](10),
			wantValue: 10,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.opt.MustGet(); got != tc.wantValue {
				t.Fatalf("MustGet() got %v, want: %v", got, tc.wantValue)
			}
		})
	}
}

func TestMustGet_Panic(t *testing.T) {
	t.Parallel()

	opt := Empty[int]()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustGet() did not panic, but should have")
		}
	}()

	_ = opt.MustGet()
}

func TestOrElse(t *testing.T) {
	t.Parallel()

	defaultValue := 3

	tests := []struct {
		name      string
		opt       Optional[int]
		wantValue int
	}{
		{
			name:      "empty",
			opt:       Empty[int](),
			wantValue: defaultValue,
		},
		{
			name:      "populated",
			opt:       Of[int](100),
			wantValue: 100,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.opt.OrElse(defaultValue); got != tc.wantValue {
				t.Fatalf("MustGet() got %v, want: %v", got, tc.wantValue)
			}
		})
	}
}

func TestOrElseLazy(t *testing.T) {
	t.Parallel()

	defaultValue := 3
	defaultValueCallback := func() (int, error) {
		return defaultValue, nil
	}
	errFailedCallback := errors.New("oops, callback failed")

	tests := []struct {
		name      string
		opt       Optional[int]
		callback  func() (int, error)
		wantValue int
		wantErr   error
	}{
		{
			name:      "empty_and_callback_ok",
			opt:       Empty[int](),
			callback:  defaultValueCallback,
			wantValue: defaultValue,
		},
		{
			name: "empty_and_callback_fail",
			opt:  Empty[int](),
			callback: func() (int, error) {
				return 0, errFailedCallback
			},
			wantErr: errFailedCallback,
		},
		{
			name:      "populated",
			opt:       Of[int](100),
			callback:  defaultValueCallback,
			wantValue: 100,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotVal, gotErr := tc.opt.OrElseLazy(tc.callback)
			if tc.wantErr != nil {
				if tc.wantErr != gotErr {
					t.Fatalf("OrElseLazy() error mismatch; got %v, want %v", gotErr, tc.wantErr)
				}
			} else {
				if gotVal != tc.wantValue {
					t.Fatalf("OrElseLazy() got %v, want: %v", gotVal, tc.wantValue)
				}
			}
		})
	}
}

func TestOrElseMustLazy(t *testing.T) {
	t.Parallel()

	defaultValue := 3
	defaultValueCallback := func() int {
		return defaultValue
	}

	tests := []struct {
		name      string
		opt       Optional[int]
		callback  func() int
		wantValue int
	}{
		{
			name:      "empty",
			opt:       Empty[int](),
			callback:  defaultValueCallback,
			wantValue: defaultValue,
		},
		{
			name:      "populated",
			opt:       Of[int](100),
			callback:  defaultValueCallback,
			wantValue: 100,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.opt.OrElseMustLazy(tc.callback)
			if got != tc.wantValue {
				t.Fatalf("OrElseLazy() got %v, want: %v", got, tc.wantValue)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		o1   Optional[int]
		o2   any
		want bool
	}{
		{
			name: "compare_to_arg_not_an_optional",
			o1:   Of[int](1),
			o2:   map[string]string{},
			want: false,
		},
		{
			name: "compare_to_optional_of_different_type",
			o1:   Of[int](1),
			o2:   Of[string]("abc"),
			want: false,
		},
		{
			name: "both_empty",
			o1:   Empty[int](),
			o2:   Empty[int](),
			want: true,
		},
		{
			name: "o1_empty",
			o1:   Empty[int](),
			o2:   Of[int](1),
			want: false,
		},
		{
			name: "o2_empty",
			o1:   Of[int](1),
			o2:   Empty[int](),
			want: false,
		},
		{
			name: "both_set_different_values",
			o1:   Of[int](1),
			o2:   Of[int](2),
			want: false,
		},
		{
			name: "both_set_same_value",
			o1:   Of[int](2),
			o2:   Of[int](2),
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.o1.Equal(tc.o2)
			if got != tc.want {
				t.Fatalf("o1.Equal(o2) got %v, want: %v", got, tc.want)
			}
		})
	}
}

type stringer interface {
	String() string
}

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opt  stringer
		want string
	}{
		{
			name: "empty",
			opt:  ptrTo(Empty[string]()),
			want: "<empty>",
		},
		{
			name: "of_int",
			opt:  ptrTo(Of[string]("abc")),
			want: "abc",
		},
		{
			name: "of_string",
			opt:  ptrTo(Of[string]("abc")),
			want: "abc",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.opt.String()
			if got != tc.want {
				t.Fatalf("String() got %q, want: %q", got, tc.want)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opt     Optional[any]
		wantVal string
		wantErr error
	}{
		{
			name:    "empty",
			opt:     Empty[any](),
			wantVal: `null`,
		},
		{
			name:    "of_int",
			opt:     Of[any](1),
			wantVal: `1`,
		},
		{
			name:    "of_string",
			opt:     Of[any]("abc"),
			wantVal: `"abc"`,
		},
		{
			name:    "of_string_ptr",
			opt:     Of[any](ptrTo("abc")),
			wantVal: `"abc"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotBytes, gotErr := tc.opt.MarshalJSON()
			if tc.wantErr != nil {
				if gotErr != tc.wantErr {
					t.Fatalf("MarshalJSON() error mismatch; got %v, want %v", gotErr, tc.wantErr)
				}
			} else {
				gotVal := string(gotBytes)
				if gotVal != tc.wantVal {
					t.Fatalf("MarshalJSON() got %q, want: %q", gotVal, tc.wantVal)
				}
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jsonBytes []byte
		wantErr   bool
		wantVal   string
	}{
		{
			name:      "json_for_correct_type",
			jsonBytes: []byte(`"abc"`),
			wantVal:   "abc",
		},
		{
			name:      "json_for_mismatched_type",
			jsonBytes: []byte(`{"a": 1, "b": 2}`),
			wantErr:   true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opt := Empty[string]()
			err := opt.UnmarshalJSON(tc.jsonBytes)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalJSON() got nil error, want error")
				}
			} else {
				if err != nil {
					t.Fatalf("UnmarshalJSON() got error: %v; want nil", err)
				}
				got, getErr := opt.Get()
				if getErr != nil {
					t.Fatalf("After successful UnmarshalJSON(), Get() returned error: %v", getErr)
				}
				if got != tc.wantVal {
					t.Fatalf("After successful UnmarshalJSON(), Get() got %v, want: %v", got, tc.wantVal)
				}
			}
		})
	}
}

