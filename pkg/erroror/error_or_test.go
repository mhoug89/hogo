package erroror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var errSentinel = errors.New("mock sentinel error")

func TestErrorOr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name    string
		Got     ErrorOr[string]
		WantOk  bool
		WantVal string
		WantErr error
	}{
		{
			Name:    "CreatedWithOf_ValueAndNilError",
			Got:     Of("strvalue", nil),
			WantOk:  true,
			WantVal: "strvalue",
			WantErr: nil,
		},
		{
			Name:    "CreatedWithOf_ValueAndNonNilError",
			Got:     Of("strvalue", errSentinel),
			WantOk:  false,
			WantVal: "strvalue",
			WantErr: errSentinel,
		},
		{
			Name:    "CreatedWithOfValue",
			Got:     OfValue("strvalue"),
			WantOk:  true,
			WantVal: "strvalue",
			WantErr: nil,
		},
		{
			Name:    "CreatedWithOfError",
			Got:     OfError[string](errSentinel),
			WantOk:  false,
			WantVal: "",
			WantErr: errSentinel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			if gotOk := tc.Got.Ok(); gotOk != tc.WantOk {
				t.Errorf("Ok() = %v, want %v", gotOk, tc.WantOk)
			}

			if gotVal := tc.Got.Value(); gotVal != tc.WantVal {
				t.Errorf("Value() = %v, want %v", gotVal, tc.WantVal)
			}

			gotVal, gotErr := tc.Got.Unpack()
			if gotVal != tc.WantVal {
				t.Errorf("Unpack() val = %v, want %v", gotVal, tc.WantVal)
			}
			if !errors.Is(gotErr, tc.WantErr) {
				t.Errorf("Unpack() err = %v, want %v", gotErr, tc.WantErr)
			}
		})
	}
}

func TestErrorOr2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name     string
		Got      ErrorOr2[string, int]
		WantOk   bool
		WantVal1 string
		WantVal2 int
		WantErr  error
	}{
		{
			Name:     "CreatedWithOf2_ValueAndNilError",
			Got:      Of2("strvalue", 1, nil),
			WantOk:   true,
			WantVal1: "strvalue",
			WantVal2: 1,
			WantErr:  nil,
		},
		{
			Name:     "CreatedWithOf2_ValueAndNonNilError",
			Got:      Of2("strvalue", 1, errSentinel),
			WantOk:   false,
			WantVal1: "strvalue",
			WantVal2: 1,
			WantErr:  errSentinel,
		},
		{
			Name:     "CreatedWithOfValue2",
			Got:      OfValue2("strvalue", 1),
			WantOk:   true,
			WantVal1: "strvalue",
			WantVal2: 1,
			WantErr:  nil,
		},
		{
			Name:     "CreatedWithOfError2",
			Got:      OfError2[string, int](errSentinel),
			WantOk:   false,
			WantVal1: *(new(string)),
			WantVal2: *(new(int)),
			WantErr:  errSentinel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			if gotOk := tc.Got.Ok(); gotOk != tc.WantOk {
				t.Errorf("Ok() = %v, want %v", gotOk, tc.WantOk)
			}

			gotVal1, gotVal2 := tc.Got.Value()
			if gotVal1 != tc.WantVal1 || gotVal2 != tc.WantVal2 {
				t.Errorf("Value() = (%v, %v), want (%v, %v)", gotVal1, gotVal2, tc.WantVal1, tc.WantVal2)
			}

			gotVal1, gotVal2, gotErr := tc.Got.Unpack()
			if gotVal1 != tc.WantVal1 || gotVal2 != tc.WantVal2 {
				t.Errorf("Unpack() vals = (%v, %v), want (%v, %v)", gotVal1, gotVal2, tc.WantVal1, tc.WantVal2)
			}
			if !errors.Is(gotErr, tc.WantErr) {
				t.Errorf("Unpack() err = %v, want %v", gotErr, tc.WantErr)
			}
		})
	}
}

func TestCmpDiff(t *testing.T) {
	t.Parallel()

	equalStrLen := cmp.Comparer(func(a, b string) bool {
		return len(a) == len(b)
	})

	tests := []struct {
		Name     string
		Want     ErrorOr[string]
		Got      ErrorOr[string]
		Opts     []cmp.Option
		WantDiff bool
	}{
		{
			Name:     "EqualFields_EmptyDiff",
			Want:     OfValue("strvalue"),
			Got:      OfValue("strvalue"),
			WantDiff: false,
		},
		{
			Name:     "EqualValues_EquatedErrors_EmptyDiff",
			Want:     Of("strvalue", errSentinel),
			Got:      Of("strvalue", fmt.Errorf("wrapped: %w", errSentinel)),
			WantDiff: false,
		},
		{
			Name:     "DifferentValues_NilErrors_NonEmptyDiff",
			Want:     OfValue("strvalue1"),
			Got:      OfValue("strvalue2"),
			WantDiff: true,
		},
		{
			Name:     "ValuesDifferentButConsideredEqualDueToCustomOpt_EmptyDiff",
			Want:     OfValue("strvalue1"),
			Got:      OfValue("strvalue2"),
			Opts:     []cmp.Option{equalStrLen},
			WantDiff: false,
		},
		{
			Name:     "EqualValues_NilAndNonNilError_NonEmptyDiff",
			Want:     OfValue("strvalue"),
			Got:      Of("strvalue", errSentinel),
			WantDiff: true,
		},
		{
			Name:     "EqualValues_DifferentErrors_NonEmptyDiff",
			Want:     Of("strvalue", errSentinel),
			Got:      Of("strvalue", errors.New("different error")),
			WantDiff: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			gotDiff := CmpDiff(tc.Want, tc.Got, tc.Opts...)
			gotDiffIsEmpty := gotDiff == ""
			if tc.WantDiff && gotDiffIsEmpty {
				t.Errorf("CmpDiff() was empty, wanted non-empty diff")
			} else if !tc.WantDiff && !gotDiffIsEmpty {
				t.Errorf("Wanted empty diff, but CmpDiff() was non-empty; got:\n%s", gotDiff)
			}
		})
	}
}

func TestCmpDiff2(t *testing.T) {
	t.Parallel()

	equalStrLen := cmp.Comparer(func(a, b string) bool {
		return len(a) == len(b)
	})

	tests := []struct {
		Name     string
		Want     ErrorOr2[string, int]
		Got      ErrorOr2[string, int]
		Opts     []cmp.Option
		WantDiff bool
	}{
		{
			Name:     "EqualFields_EmptyDiff",
			Want:     OfValue2("strvalue", 1),
			Got:      OfValue2("strvalue", 1),
			WantDiff: false,
		},
		{
			Name:     "EqualValues_EquatedErrors_EmptyDiff",
			Want:     Of2("strvalue", 1, errSentinel),
			Got:      Of2("strvalue", 1, fmt.Errorf("wrapped: %w", errSentinel)),
			WantDiff: false,
		},
		{
			Name:     "DifferentValues_NilErrors_NonEmptyDiff",
			Want:     OfValue2("strvalue1", 1),
			Got:      OfValue2("strvalue2", 1),
			WantDiff: true,
		},
		{
			Name:     "DifferentSecondValue_NilErrors_NonEmptyDiff",
			Want:     OfValue2("strvalue", 1),
			Got:      OfValue2("strvalue", 2),
			WantDiff: true,
		},
		{
			Name:     "ValuesDifferentButConsideredEqualDueToCustomOpt_EmptyDiff",
			Want:     OfValue2("strvalue1", 1),
			Got:      OfValue2("strvalue2", 1),
			Opts:     []cmp.Option{equalStrLen},
			WantDiff: false,
		},
		{
			Name:     "EqualValues_NilAndNonNilError_NonEmptyDiff",
			Want:     OfValue2("strvalue", 1),
			Got:      Of2("strvalue", 1, errSentinel),
			WantDiff: true,
		},
		{
			Name:     "EqualValues_DifferentErrors_NonEmptyDiff",
			Want:     Of2("strvalue", 1, errSentinel),
			Got:      Of2("strvalue", 1, errors.New("different error")),
			WantDiff: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			gotDiff := CmpDiff2(tc.Want, tc.Got, tc.Opts...)
			gotDiffIsEmpty := gotDiff == ""
			if tc.WantDiff && gotDiffIsEmpty {
				t.Errorf("CmpDiff2() was empty, wanted non-empty diff")
			} else if !tc.WantDiff && !gotDiffIsEmpty {
				t.Errorf("Wanted empty diff, but CmpDiff2() was non-empty; got:\n%s", gotDiff)
			}
		})
	}
}
