// -----------------------------------------------------------------------------
// Copyright (c) 2026-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2026-present Detlef Stern
// -----------------------------------------------------------------------------

package roster

import (
	"slices"
	"testing"
)

func contents(r Roster[uint]) []uint { return slices.Collect(r.Values()) }

func TestGrow(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		in   []uint
		n    int
	}{
		{"zero value, zero", nil, 0},
		{"zero value", nil, 10},
		{"non-empty, zero", []uint{1, 2, 3}, 0},
		{"non-empty, one", []uint{1, 2, 3}, 1},
		{"non-empty", []uint{1, 2, 3}, 100},
		{"large", []uint{1, 2, 3}, 100_000},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := New(tc.in...)
			want := contents(r)

			r.Grow(tc.n)

			if free := cap(r.array) - len(r.array); free < tc.n {
				t.Errorf("free capacity = %d, want at least %d", free, tc.n)
			}
			// Grow must neither add values nor change existing ones.
			if got := contents(r); !slices.Equal(got, want) {
				t.Errorf("contents = %v, want %v", got, want)
			}
			if got := r.Count(); got != len(want) {
				t.Errorf("Count() = %d, want %d", got, len(want))
			}
		})
	}
}

func TestGrowAvoidsReallocation(t *testing.T) {
	t.Parallel()
	r := New[uint](10, 20, 30)
	r.Grow(5)
	capAfterGrow, first := cap(r.array), &r.array[0]

	// Five insertions: front, middle, end, front again, end again.
	for _, n := range []uint{5, 25, 40, 1, 50} {
		r.Insert(n)
	}

	if cap(r.array) != capAfterGrow || &r.array[0] != first {
		t.Error("Insert reallocated the storage although Grow reserved enough space")
	}
	if got, want := contents(r), []uint{1, 5, 10, 20, 25, 30, 40, 50}; !slices.Equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
	}
}

func TestGrowIntoZeroValue(t *testing.T) {
	t.Parallel()
	var r Roster[uint]
	r.Grow(3)
	capAfterGrow := cap(r.array)

	for _, n := range []uint{3, 1, 2} {
		r.Insert(n)
	}

	if cap(r.array) != capAfterGrow {
		t.Errorf("cap changed from %d to %d", capAfterGrow, cap(r.array))
	}
	if got, want := contents(r), []uint{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
	}
}

func TestGrowKeepsSufficientStorage(t *testing.T) {
	t.Parallel()
	r := New[uint](1, 2, 3)
	r.Grow(10)
	first, capBefore := &r.array[0], cap(r.array)

	r.Grow(10) // already enough
	r.Grow(5)
	r.Grow(0)

	if &r.array[0] != first || cap(r.array) != capBefore {
		t.Error("Grow reallocated although the free capacity was sufficient")
	}
}

func TestGrowNegativePanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Error("Grow(-1) did not panic")
		}
	}()
	r := New[uint](1, 2, 3)
	r.Grow(-1)
}

func TestClip(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name  string
		setup func() Roster[uint]
		want  []uint
	}{
		{"zero value", func() Roster[uint] { return Roster[uint]{} }, nil},
		{"tight", func() Roster[uint] {
			r := New[uint](1, 2, 3)
			r.Clip()
			return r
		}, []uint{1, 2, 3}},
		{"spare capacity after Grow", func() Roster[uint] {
			r := New[uint](1, 2, 3)
			r.Grow(100)
			return r
		}, []uint{1, 2, 3}},
		{"spare capacity after Delete", func() Roster[uint] {
			r := New[uint](1, 2, 3, 4, 5)
			r.Delete(2)
			r.Delete(5)
			return r
		}, []uint{1, 3, 4}},
		{"shrunk by And", func() Roster[uint] {
			r := New[uint](1, 2, 3, 4, 5)
			r.And(New[uint](2, 4, 6))
			return r
		}, []uint{2, 4}},
		{"shrunk by AndNot", func() Roster[uint] {
			r := New[uint](1, 2, 3, 4, 5)
			r.AndNot(New[uint](1, 2, 3))
			return r
		}, []uint{4, 5}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := tc.setup()

			r.Clip()

			if cap(r.array) != len(r.array) {
				t.Errorf("cap = %d, len = %d, want equal", cap(r.array), len(r.array))
			}
			if got := contents(r); !slices.Equal(got, tc.want) {
				t.Errorf("contents = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClipEmptyReleasesStorage(t *testing.T) {
	t.Parallel()
	for name, empty := range map[string]func() Roster[uint]{
		"DeleteAll": func() Roster[uint] {
			r := New[uint](1, 2, 3)
			r.DeleteAll()
			return r
		},
		"Delete": func() Roster[uint] {
			r := New[uint](1)
			r.Delete(1)
			return r
		},
		"And": func() Roster[uint] {
			r := New[uint](1, 2, 3)
			r.And(New[uint](7))
			return r
		},
		"Xor": func() Roster[uint] {
			r := New[uint](1, 2, 3)
			r.Xor(New[uint](1, 2, 3))
			return r
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := empty()
			if cap(r.array) == 0 {
				t.Skip("setup left no storage to release")
			}

			r.Clip()

			if r.array != nil {
				t.Errorf("array = %v with cap %d, want nil", r.array, cap(r.array))
			}
			if !r.IsEmpty() {
				t.Errorf("not empty: %v", contents(r))
			}
		})
	}
}

func TestClipTwiceKeepsStorage(t *testing.T) {
	t.Parallel()
	r := New[uint](1, 2, 3)
	r.Grow(10)
	r.Clip()
	first, capBefore := &r.array[0], cap(r.array)

	r.Clip() // already tight: must not copy

	if &r.array[0] != first || cap(r.array) != capBefore {
		t.Error("second Clip reallocated")
	}
}

func TestClipThenReuse(t *testing.T) {
	t.Parallel()
	r := New[uint](1, 2, 3, 4, 5)
	r.And(New[uint](2, 4, 6))
	r.Clip()

	// Values dropped by And must not reappear, and growth after Clip works.
	r.Insert(3)
	r.Insert(9)
	r.Insert(0)

	if got, want := contents(r), []uint{0, 2, 3, 4, 9}; !slices.Equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
	}
}

func TestClipDoesNotAffectCopy(t *testing.T) {
	t.Parallel()
	r1 := New[uint](1, 2, 3)
	r1.Grow(10)
	r2 := r1 // shares the storage

	r1.Clip()

	if got, want := contents(r2), []uint{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
	if got, want := contents(r1), []uint{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("clipped = %v, want %v", got, want)
	}
}

func TestGrowThenClipRoundTrip(t *testing.T) {
	t.Parallel()
	r := New[uint](4, 8)

	r.Grow(1000)
	r.Clip()

	if cap(r.array) != 2 {
		t.Errorf("cap = %d, want 2", cap(r.array))
	}
	if got, want := contents(r), []uint{4, 8}; !slices.Equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
	}
}
