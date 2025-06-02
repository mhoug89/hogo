package set

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

var (
	allLetters = strings.Split("abcdefghijklmnopqrstuvwxyz", "")
)

func TestNew(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		items []string
	}{
		{
			name:  "no_args",
			items: []string{},
		},
		{
			name:  "1_item",
			items: []string{"a"},
		},
		{
			name:  "2_items",
			items: []string{"a", "b"},
		},
	} {
		s := New(tc.items...)

		gotLen := s.Len()
		wantLen := len(tc.items)
		if gotLen != wantLen {
			t.Errorf("Got len of %d after creating map with slice of len %d", gotLen, wantLen)
		}

		for _, item := range tc.items {
			if !s.Has(item) {
				t.Errorf("Did not find item %q after providing it in args to New", item)
			}
		}
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	itemToAdd := "a"
	for _, tc := range []struct {
		name    string
		items   []string
		wantLen int
	}{
		{
			name:    "add_new_item",
			items:   []string{},
			wantLen: 1,
		},
		{
			name:    "add_existing_item",
			items:   []string{itemToAdd},
			wantLen: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			s.Add(itemToAdd)
			gotLen := s.Len()

			if gotLen != tc.wantLen {
				t.Fatalf("Len() after Add(%q) was %d, want %d", itemToAdd, gotLen, tc.wantLen)
			}
		})
	}
}

func TestAddIfAbsent(t *testing.T) {
	t.Parallel()

	itemToAdd := "a"
	for _, tc := range []struct {
		name  string
		items []string
		want  bool
	}{
		{
			name:  "add_new_item",
			items: []string{},
			want:  true,
		},
		{
			name:  "add_existing_item",
			items: []string{itemToAdd},
			want:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			got := s.AddIfAbsent(itemToAdd)

			if got != tc.want {
				t.Fatalf("AddIfAbsent(%q) returned %v, want %v", itemToAdd, got, tc.want)
			}
		})
	}
}

func TestClear(t *testing.T) {
	t.Parallel()

	s := New(strings.Split("abcdefghijklmnopqrstuvwxyz", "")...)
	s.Clear()
	gotLen := s.Len()
	wantLen := 0
	if gotLen != wantLen {
		t.Fatalf("Got len of %d after Clear(), but wanted len of %d", gotLen, wantLen)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	itemToDelete := "a"
	for _, tc := range []struct {
		name    string
		items   []string
		wantLen int
	}{
		{
			name:    "delete_existing_item",
			items:   []string{itemToDelete},
			wantLen: 0,
		},
		{
			name:    "delete_nonexistent_item_is_noop",
			items:   []string{"some other string"},
			wantLen: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			s.Delete(itemToDelete)
			gotLen := s.Len()

			if s.Has(itemToDelete) {
				t.Errorf("Found %q in set, but Delete() should have removed it", itemToDelete)
			}
			if gotLen != tc.wantLen {
				t.Errorf("Got len of %d after Delete(), want %d", gotLen, tc.wantLen)
			}
		})
	}
}

func TestDeleteIfPresent(t *testing.T) {
	t.Parallel()

	itemToDelete := "a"
	for _, tc := range []struct {
		name    string
		items   []string
		want    bool
		wantLen int
	}{
		{
			name:    "delete_existing_item",
			items:   []string{itemToDelete},
			want:    true,
			wantLen: 0,
		},
		{
			name:    "delete_nonexistent_item_is_noop",
			items:   []string{"some other string"},
			want:    false,
			wantLen: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			got := s.DeleteIfPresent(itemToDelete)
			gotLen := s.Len()

			if got != tc.want {
				t.Errorf("Got %v from DeleteIfPresent(), want %v", got, tc.want)
			}
			if s.Has(itemToDelete) {
				t.Errorf("Found %q in set, but DeleteIfPresent() should have guaranteed its absence", itemToDelete)
			}
			if gotLen != tc.wantLen {
				t.Errorf("Got len of %d after DeleteIfPresent(), want %d", gotLen, tc.wantLen)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		items1 []string
		items2 []string
		want   bool
	}{
		{
			name:   "both_empty",
			items1: []string{},
			items2: []string{},
			want:   true,
		},
		{
			name:   "first_empty",
			items1: []string{},
			items2: allLetters,
			want:   false,
		},
		{
			name:   "other_empty",
			items1: allLetters,
			items2: []string{},
			want:   false,
		},
		{
			name:   "same_items",
			items1: allLetters,
			items2: allLetters,
			want:   true,
		},
		{
			name:   "same_length_but_different_items",
			items1: []string{"a", "b", "c", "d", "e"},
			items2: []string{"1", "2", "3", "4", "5"},
			want:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s1 := New(tc.items1...)
			s2 := New(tc.items2...)
			got := s1.Equal(s2)

			if got != tc.want {
				t.Fatalf("Equal() got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHas(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		items []string
		check string
		want  bool
	}{
		{
			name:  "empty",
			items: []string{},
			check: "a",
			want:  false,
		},
		{
			name:  "one_item_success",
			items: []string{"a"},
			check: "a",
			want:  true,
		},
		{
			name:  "one_item_absent",
			items: []string{"b"},
			check: "a",
			want:  false,
		},
		{
			name:  "all_letters_find_a",
			items: allLetters,
			check: "a",
			want:  true,
		},
		{
			name:  "all_letters_find_digit",
			items: allLetters,
			check: "1",
			want:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			got := s.Has(tc.check)

			if got != tc.want {
				t.Fatalf("Has(%q) got %v, want %v", tc.check, got, tc.want)
			}
		})
	}
}

func TestHasAll(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		items []string
		check []string
		want  bool
	}{
		{
			name:  "empty_find_0_args",
			items: []string{},
			check: []string{},
			want:  true,
		},
		{
			name:  "empty_find_1",
			items: []string{},
			check: []string{"1"},
			want:  false,
		},
		{
			name:  "one_item_success",
			items: []string{"a"},
			check: []string{"a"},
			want:  true,
		},
		{
			name:  "one_item_absent",
			items: []string{"b"},
			check: []string{"a"},
			want:  false,
		},
		{
			name:  "multi_item_none_found",
			items: allLetters,
			check: []string{"1", "2", "3"},
			want:  false,
		},
		{
			name:  "multi_item_some_found",
			items: allLetters,
			check: []string{"a", "b", "1"},
			want:  false,
		},
		{
			name:  "multi_item_find_subset",
			items: allLetters,
			check: []string{"a", "b", "c"},
			want:  true,
		},
		{
			name:  "multi_item_find_entire_collection",
			items: allLetters,
			check: allLetters,
			want:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			got := s.HasAll(tc.check...)

			if got != tc.want {
				t.Fatalf("HasAll() got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		ctor func() Set[string]
		want bool
	}{
		{
			name: "empty",
			ctor: func() Set[string] {
				return New[string]()
			},
			want: true,
		},
		{
			name: "one_item",
			ctor: func() Set[string] {
				return New[string]("a")
			},
			want: false,
		},
		{
			name: "multi_item",
			ctor: func() Set[string] {
				return New[string]("a", "b")
			},
			want: false,
		},
		{
			name: "empty_after_new_then_delete_item",
			ctor: func() Set[string] {
				s := New[string]("a")
				s.Delete("a")
				return s
			},
			want: true,
		},
		{
			name: "empty_after_add_then_delete_item",
			ctor: func() Set[string] {
				s := New[string]("a")
				s.Delete("a")
				return s
			},
			want: true,
		},
		{
			name: "empty_after_clear",
			ctor: func() Set[string] {
				s := New[string]("a")
				s.Add("b")
				s.Clear()
				return s
			},
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.ctor()
			got := s.IsEmpty()

			if got != tc.want {
				t.Fatalf("IsEmpty() got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLen(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		items []string
		want  int
	}{
		{
			name:  "empty",
			items: []string{},
			want:  0,
		},
		{
			name:  "one",
			items: []string{"a"},
			want:  1,
		},
		{
			name:  "multi",
			items: []string{"a", "b", "c"},
			want:  3,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			got := s.Len()

			if got != tc.want {
				t.Fatalf("Len() got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPop(t *testing.T) {
	t.Parallel()

	s := New[string]()
	val, didPop := s.Pop()
	if didPop || val != "" {
		t.Fatalf(`Pop() called on an empty set got (%q, %v), want ("", false)`, val, didPop)
	}

	s.Add("a")
	val, didPop = s.Pop()
	if !didPop || val != "a" {
		t.Fatalf(`Pop() got (%q, %v), want ("a", true)`, val, didPop)
	}

	val, didPop = s.Pop()
	if didPop || val != "" {
		t.Fatalf(`Pop() (after popping set until empty) called on an empty set got (%q, %v), want ("", false)`, val, didPop)
	}
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		items []string
	}{
		{
			name:  "empty",
			items: []string{},
		},
		{
			name:  "one",
			items: []string{"a"},
		},
		{
			name:  "multi",
			items: allLetters,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(tc.items...)
			got := s.ToSlice()
			slices.Sort(got)
			want := slices.Clone(tc.items)
			slices.Sort(want)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Did not get expected set of items from ToSlice; got %#v, want %#v", got, want)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		set       Set[string]
		toAdd     []Set[string]
		wantItems []string
	}{
		{
			name:      "empty_update_none",
			set:       New[string](),
			toAdd:     []Set[string]{},
			wantItems: []string{},
		},
		{
			name:      "empty_update_empty",
			set:       New[string](),
			toAdd:     []Set[string]{New[string]()},
			wantItems: []string{},
		},
		{
			name: "empty_update_one",
			set:  New[string](),
			toAdd: []Set[string]{
				New("a", "b", "c"),
			},
			wantItems: []string{"a", "b", "c"},
		},
		{
			name: "empty_update_multi",
			set:  New[string](),
			toAdd: []Set[string]{
				New(allLetters[0:13]...),
				New(allLetters[13:]...),
			},
			wantItems: allLetters,
		},
		{
			name: "empty_update_multi_with_duplicates",
			set:  New[string](),
			toAdd: []Set[string]{
				New(allLetters[0:13]...),
				New(allLetters[13:]...),
				New(allLetters...),
			},
			wantItems: allLetters,
		},
		{
			name:      "nonempty_update_none",
			set:       New("a"),
			toAdd:     []Set[string]{},
			wantItems: []string{"a"},
		},
		{
			name:      "nonempty_update_empty",
			set:       New("a"),
			toAdd:     []Set[string]{New[string]()},
			wantItems: []string{"a"},
		},
		{
			name: "nonempty_update_one",
			set:  New("a"),
			toAdd: []Set[string]{
				New("b", "c"),
			},
			wantItems: []string{"a", "b", "c"},
		},
		{
			name: "nonempty_update_multi",
			set:  New("a"),
			toAdd: []Set[string]{
				New(allLetters[1:13]...),
				New(allLetters[13:]...),
			},
			wantItems: allLetters,
		},
		{
			name: "nonempty_update_multi_with_duplicates",
			set:  New[string]("a"),
			toAdd: []Set[string]{
				New(allLetters[0:13]...),
				New(allLetters[13:]...),
				New(allLetters...),
			},
			wantItems: allLetters,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.set.Update(tc.toAdd...)
			gotItems := tc.set.ToSlice()
			slices.Sort(gotItems)

			if !reflect.DeepEqual(gotItems, tc.wantItems) {
				t.Fatalf("Did not get expected set of items after Update();\n- got:\n%#v\n- want:\n%#v\n", gotItems, tc.wantItems)
			}
		})
	}
}
