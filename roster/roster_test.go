//-----------------------------------------------------------------------------
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
//-----------------------------------------------------------------------------

package roster_test

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"t73f.de/r/zero/roster"
)

var newTestcases = []struct {
	name string
	in   []uint
	want []uint
}{
	{"empty", nil, nil},
	{"single", []uint{7}, []uint{7}},
	{"sorted", []uint{1, 2, 3}, []uint{1, 2, 3}},
	{"reverse", []uint{3, 2, 1}, []uint{1, 2, 3}},
	{"unsorted", []uint{5, 1, 4, 2, 3}, []uint{1, 2, 3, 4, 5}},
	{"duplicates at start", []uint{1, 1, 2, 3}, []uint{1, 2, 3}},
	{"duplicates in middle", []uint{1, 2, 2, 3}, []uint{1, 2, 3}},
	{"duplicates at end", []uint{1, 2, 3, 3}, []uint{1, 2, 3}},
	{"non-adjacent duplicates", []uint{3, 1, 2, 1, 3}, []uint{1, 2, 3}},
	{"all equal", []uint{4, 4, 4, 4}, []uint{4}},
	{"zero", []uint{0, 0}, []uint{0}},
}

func TestNew(t *testing.T) {
	t.Parallel()
	for _, tc := range newTestcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := slices.Collect(roster.New(tc.in...).Values())
			if !slices.Equal(got, tc.want) {
				t.Errorf("New(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestNewDoesNotShareInput(t *testing.T) {
	t.Parallel()
	in := []uint{3, 1, 2, 1}
	orig := slices.Clone(in)

	got := roster.New(in...)

	if !slices.Equal(in, orig) {
		t.Errorf("input modified: %v, want %v", in, orig)
	}
	in[0] = 99
	if gotArr, want := slices.Collect(got.Values()), []uint{1, 2, 3}; !slices.Equal(gotArr, want) {
		t.Errorf("result changed after input modification: %v, want %v", gotArr, want)
	}
}

func TestNewTypeBoundaries(t *testing.T) {
	t.Parallel()
	got := slices.Collect(roster.New[uint64](math.MaxUint64, 0, math.MaxUint64, 1).Values())
	if want := []uint64{0, 1, math.MaxUint64}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	type id uint16 // values with an underlying type are allowed (~uint16)
	gotID := slices.Collect(roster.New[id](300, 5, 300).Values())
	if want := []id{5, 300}; !slices.Equal(gotID, want) {
		t.Errorf("got %v, want %v", gotID, want)
	}
}

func TestCollect(t *testing.T) {
	t.Parallel()
	for _, tc := range newTestcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := roster.Collect(slices.Values(tc.in))
			if gotArr := slices.Collect(got.Values()); !slices.Equal(gotArr, tc.want) {
				t.Errorf("Collect(%v) = %v, want %v", tc.in, gotArr, tc.want)
			}
		})
	}
}

func TestCollectUnorderedSequence(t *testing.T) {
	t.Parallel()
	// Map iteration order is randomized: the sequence is deliberately unordered.
	m := map[uint]struct{}{10: {}, 2: {}, 7: {}, 300: {}, 0: {}}

	got := roster.Collect(maps.Keys(m))
	gotArr := slices.Collect(got.Values())
	if want := []uint{0, 2, 7, 10, 300}; !slices.Equal(gotArr, want) {
		t.Errorf("got %v, want %v", gotArr, want)
	}
}

func TestNewCollectProperties(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(1, 2))

	for range 500 {
		in := make([]uint, rng.IntN(40))
		ref := make(map[uint]struct{})
		for i := range in {
			in[i] = uint(rng.IntN(60)) // small range: many duplicates
			ref[in[i]] = struct{}{}
		}
		want := slices.Sorted(maps.Keys(ref))

		for name, got := range map[string]roster.Roster[uint]{
			"New":     roster.New(in...),
			"Collect": roster.Collect(slices.Values(in)),
		} {
			gotArr := slices.Collect(got.Values())
			if !slices.Equal(gotArr, want) {
				t.Fatalf("%s(%v) = %v, want %v", name, in, gotArr, want)
			}
			if !strictlyAscending(gotArr) {
				t.Fatalf("%s(%v) violates invariant: %v", name, in, gotArr)
			}
		}
	}
}

func strictlyAscending[V roster.Value](a []V) bool {
	for i := 1; i < len(a); i++ {
		if a[i-1] >= a[i] {
			return false
		}
	}
	return true
}

func TestInsertDelete(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(3, 4))

	var r roster.Roster[uint]
	ref := make(map[uint]struct{})

	for range 5000 {
		n := uint(rng.IntN(50))
		switch rng.IntN(10) {
		case 0:
			r.DeleteAll()
			clear(ref)
		case 1, 2, 3, 4:
			r.Delete(n)
			delete(ref, n)
		default:
			r.Insert(n)
			ref[n] = struct{}{}
		}

		got := slices.Collect(r.Values())
		if want := slices.Sorted(maps.Keys(ref)); !slices.Equal(got, want) {
			t.Fatalf("after op on %d: got %v, want %v", n, got, want)
		}
	}
}

func values(r roster.Roster[uint]) []uint { return slices.Collect(r.Values()) }

func TestInsert(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		in   []uint
		n    uint
		want []uint
	}{
		{"into empty", nil, 5, []uint{5}},
		{"at front", []uint{3, 5, 7}, 1, []uint{1, 3, 5, 7}},
		{"in middle", []uint{3, 5, 7}, 4, []uint{3, 4, 5, 7}},
		{"at end", []uint{3, 5, 7}, 9, []uint{3, 5, 7, 9}},
		{"zero at front", []uint{3, 5}, 0, []uint{0, 3, 5}},
		{"duplicate at front", []uint{3, 5, 7}, 3, []uint{3, 5, 7}},
		{"duplicate in middle", []uint{3, 5, 7}, 5, []uint{3, 5, 7}},
		{"duplicate at end", []uint{3, 5, 7}, 7, []uint{3, 5, 7}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roster.New(tc.in...)

			r.Insert(tc.n)

			if got := values(r); !slices.Equal(got, tc.want) {
				t.Errorf("Insert(%d) on %v = %v, want %v", tc.n, tc.in, got, tc.want)
			}
		})
	}
}

func TestInsertIntoZeroValue(t *testing.T) {
	t.Parallel()
	var r roster.Roster[uint] // zero value must be usable

	r.Insert(2)
	r.Insert(1)

	if got, want := values(r), []uint{1, 2}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		in   []uint
		n    uint
		want []uint
	}{
		{"from empty", nil, 5, nil},
		{"only element", []uint{5}, 5, nil},
		{"first", []uint{3, 5, 7}, 3, []uint{5, 7}},
		{"middle", []uint{3, 5, 7}, 5, []uint{3, 7}},
		{"last", []uint{3, 5, 7}, 7, []uint{3, 5}},
		{"missing below", []uint{3, 5, 7}, 1, []uint{3, 5, 7}},
		{"missing between", []uint{3, 5, 7}, 4, []uint{3, 5, 7}},
		{"missing above", []uint{3, 5, 7}, 9, []uint{3, 5, 7}},
		{"zero", []uint{0, 3}, 0, []uint{3}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roster.New(tc.in...)

			r.Delete(tc.n)

			if got := values(r); !slices.Equal(got, tc.want) {
				t.Errorf("Delete(%d) on %v = %v, want %v", tc.n, tc.in, got, tc.want)
			}
		})
	}
}

func TestDeleteTwice(t *testing.T) {
	t.Parallel()
	r := roster.New[uint](1, 2, 3)

	r.Delete(2)
	r.Delete(2) // must have no further effect

	if got, want := values(r), []uint{1, 3}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDeleteAll(t *testing.T) {
	t.Parallel()

	t.Run("on empty", func(t *testing.T) {
		t.Parallel()
		var r roster.Roster[uint]
		r.DeleteAll()
		if got := values(r); len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("on non-empty", func(t *testing.T) {
		t.Parallel()
		r := roster.New[uint](1, 2, 3)
		r.DeleteAll()
		if got := values(r); len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("insert after DeleteAll", func(t *testing.T) {
		t.Parallel()
		r := roster.New[uint](10, 20, 30, 40)
		r.DeleteAll()

		r.Insert(25)
		r.Insert(5)

		// No value from before DeleteAll may reappear from reused storage.
		if got, want := values(r), []uint{5, 25}; !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("delete after DeleteAll", func(t *testing.T) {
		t.Parallel()
		r := roster.New[uint](10, 20)
		r.DeleteAll()

		r.Delete(10) // must not find stale data

		if got := values(r); len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})
}

func TestDeleteAllEmpty(t *testing.T) {
	var bs roster.Roster[uint]

	bs.DeleteAll()
	if !bs.IsEmpty() {
		t.Fatal("DeleteAll() on empty set is not empty")
	}
}

func testPop[V roster.Value](t *testing.T, values ...V) {
	t.Helper()

	r := roster.New(values...)

	want := make(map[V]bool, len(values))
	for _, v := range values {
		want[v] = true
	}

	for len(want) > 0 {
		v, ok := r.Pop()
		if !ok {
			t.Fatal("Pop returned false for non-empty roster")
		}

		if !want[v] {
			t.Fatalf("Pop returned unexpected or duplicate value %v", v)
		}
		delete(want, v)

		if r.Contains(v) {
			t.Fatalf("roster still contains popped value %v", v)
		}
	}

	if !r.IsEmpty() {
		t.Fatal("roster is not empty after popping all values")
	}

	if _, ok := r.Pop(); ok {
		t.Fatal("Pop returned true for empty roster")
	}
}

func TestRosterPop(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		testPop(t, uint(0), 1, 63, 64, 127, 1000)
	})

	t.Run("uint8", func(t *testing.T) {
		testPop(t, uint8(0), 1, 127, 255)
	})

	t.Run("uint16", func(t *testing.T) {
		testPop(t, uint16(0), 1, 255, 256, 65535)
	})

	t.Run("uint32", func(t *testing.T) {
		testPop(t, uint32(0), 1, 1<<16, 1<<31, ^uint32(0))
	})
}

func TestContains(t *testing.T) {
	t.Parallel()
	r := roster.New[uint](3, 5, 7)

	testcases := []struct {
		n    uint
		want bool
	}{
		{0, false},
		{2, false}, // below the smallest value
		{3, true},  // first
		{4, false}, // between values
		{5, true},  // middle
		{6, false},
		{7, true},  // last
		{8, false}, // above the largest value
	}
	for _, tc := range testcases {
		if got := r.Contains(tc.n); got != tc.want {
			t.Errorf("Contains(%d) = %v, want %v", tc.n, got, tc.want)
		}
	}
}

func TestContainsSpecialSets(t *testing.T) {
	t.Parallel()

	var zero roster.Roster[uint]
	if zero.Contains(0) {
		t.Error("zero value contains 0")
	}
	if roster.New[uint]().Contains(0) {
		t.Error("empty set contains 0")
	}

	withZero := roster.New[uint](0, 9)
	if !withZero.Contains(0) || !withZero.Contains(9) || withZero.Contains(1) {
		t.Errorf("Contains wrong on %v", values(withZero))
	}

	big := roster.New[uint64](0, math.MaxUint64)
	if !big.Contains(math.MaxUint64) || big.Contains(math.MaxUint64-1) {
		t.Errorf("Contains wrong at type boundary on %v", slices.Collect(big.Values()))
	}
}

func TestContainsAfterModification(t *testing.T) {
	t.Parallel()
	r := roster.New[uint](1, 2, 3)

	r.Delete(2)
	if r.Contains(2) || !r.Contains(1) || !r.Contains(3) {
		t.Errorf("after Delete: %v", values(r))
	}
	r.Insert(2)
	if !r.Contains(2) {
		t.Errorf("after Insert: %v", values(r))
	}
	r.DeleteAll()
	if r.Contains(1) || r.Contains(2) || r.Contains(3) {
		t.Errorf("after DeleteAll: %v", values(r))
	}
}

func TestCountAndIsEmpty(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name      string
		in        []uint
		wantCount int
	}{
		{"empty", nil, 0},
		{"single", []uint{4}, 1},
		{"zero only", []uint{0}, 1},
		{"distinct", []uint{1, 2, 3}, 3},
		{"duplicates", []uint{1, 1, 2, 2, 2}, 2},
		{"all equal", []uint{9, 9, 9}, 1},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roster.New(tc.in...)

			if got := r.Count(); got != tc.wantCount {
				t.Errorf("Count() = %d, want %d", got, tc.wantCount)
			}
			if got, want := r.IsEmpty(), tc.wantCount == 0; got != want {
				t.Errorf("IsEmpty() = %v, want %v", got, want)
			}
		})
	}
}

func TestCountAndIsEmptyAfterModification(t *testing.T) {
	t.Parallel()
	var r roster.Roster[uint]
	if r.Count() != 0 || !r.IsEmpty() {
		t.Fatalf("zero value: Count=%d IsEmpty=%v", r.Count(), r.IsEmpty())
	}

	r.Insert(5)
	r.Insert(5) // duplicate must not change the count
	if r.Count() != 1 || r.IsEmpty() {
		t.Errorf("after Insert: Count=%d IsEmpty=%v", r.Count(), r.IsEmpty())
	}

	r.Delete(6) // missing value must not change the count
	if r.Count() != 1 {
		t.Errorf("after Delete of missing value: Count=%d", r.Count())
	}

	r.Delete(5)
	if r.Count() != 0 || !r.IsEmpty() {
		t.Errorf("after Delete: Count=%d IsEmpty=%v", r.Count(), r.IsEmpty())
	}

	r.Insert(1)
	r.Insert(2)
	r.DeleteAll()
	if r.Count() != 0 || !r.IsEmpty() {
		t.Errorf("after DeleteAll: Count=%d IsEmpty=%v", r.Count(), r.IsEmpty())
	}
}

func TestQueriesAgainstReference(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(5, 6))

	for range 300 {
		in := make([]uint, rng.IntN(30))
		ref := make(map[uint]struct{})
		for i := range in {
			in[i] = uint(rng.IntN(40))
			ref[in[i]] = struct{}{}
		}
		r := roster.New(in...)

		if r.Count() != len(ref) {
			t.Fatalf("%v: Count() = %d, want %d", in, r.Count(), len(ref))
		}
		if r.IsEmpty() != (len(ref) == 0) {
			t.Fatalf("%v: IsEmpty() = %v", in, r.IsEmpty())
		}
		for n := range uint(45) { // includes values outside the input range
			_, want := ref[n]
			if got := r.Contains(n); got != want {
				t.Fatalf("%v: Contains(%d) = %v, want %v", in, n, got, want)
			}
		}
	}
}

func TestMinMax(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name             string
		in               []uint
		wantMin, wantMax uint
		wantOK           bool
	}{
		{"empty", nil, 0, 0, false},
		{"single", []uint{7}, 7, 7, true},
		{"zero only", []uint{0}, 0, 0, true},
		{"zero and other", []uint{0, 9}, 0, 9, true},
		{"sorted", []uint{1, 2, 3}, 1, 3, true},
		{"unsorted", []uint{5, 1, 4}, 1, 5, true},
		{"duplicates", []uint{2, 2, 8, 8}, 2, 8, true},
		{"large gap", []uint{3, 100000}, 3, 100000, true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := roster.New(tc.in...)

			gotMin, okMin := r.Min()
			gotMax, okMax := r.Max()

			if okMin != tc.wantOK || okMax != tc.wantOK {
				t.Fatalf("ok = (%v, %v), want %v", okMin, okMax, tc.wantOK)
			}
			if !tc.wantOK {
				return // value is unspecified if the set is empty
			}
			if gotMin != tc.wantMin {
				t.Errorf("Min() = %d, want %d", gotMin, tc.wantMin)
			}
			if gotMax != tc.wantMax {
				t.Errorf("Max() = %d, want %d", gotMax, tc.wantMax)
			}
		})
	}
}

func TestMinMaxZeroValue(t *testing.T) {
	t.Parallel()
	var r roster.Roster[uint]

	if _, ok := r.Min(); ok {
		t.Error("Min() on zero value reports ok")
	}
	if _, ok := r.Max(); ok {
		t.Error("Max() on zero value reports ok")
	}
}

func TestMinMaxTypeBoundaries(t *testing.T) {
	t.Parallel()
	r := roster.New[uint64](math.MaxUint64, 0, 1<<63)

	if got, ok := r.Min(); !ok || got != 0 {
		t.Errorf("Min() = %d, %v, want 0, true", got, ok)
	}
	if got, ok := r.Max(); !ok || got != math.MaxUint64 {
		t.Errorf("Max() = %d, %v, want %d, true", got, ok, uint64(math.MaxUint64))
	}

	only := roster.New[uint8](255)
	if lo, _ := only.Min(); lo != 255 {
		t.Errorf("Min() = %d, want 255", lo)
	}
	if hi, _ := only.Max(); hi != 255 {
		t.Errorf("Max() = %d, want 255", hi)
	}
}

func TestMinMaxAfterModification(t *testing.T) {
	t.Parallel()
	r := roster.New[uint](3, 5, 7)

	check := func(step string, wantMin, wantMax uint, wantOK bool) {
		t.Helper()
		gotMin, okMin := r.Min()
		gotMax, okMax := r.Max()
		if okMin != wantOK || okMax != wantOK {
			t.Fatalf("%s: ok = (%v, %v), want %v", step, okMin, okMax, wantOK)
		}
		if wantOK && (gotMin != wantMin || gotMax != wantMax) {
			t.Errorf("%s: (min, max) = (%d, %d), want (%d, %d)",
				step, gotMin, gotMax, wantMin, wantMax)
		}
	}

	check("initial", 3, 7, true)
	r.Delete(3)
	check("delete min", 5, 7, true)
	r.Delete(7)
	check("delete max", 5, 5, true)
	r.Insert(1)
	check("insert below min", 1, 5, true)
	r.Insert(9)
	check("insert above max", 1, 9, true)
	r.Insert(4)
	check("insert in between", 1, 9, true)
	r.DeleteAll()
	check("delete all", 0, 0, false)
	r.Insert(2)
	check("insert after delete all", 2, 2, true)
}

func TestMinMaxAgainstReference(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(7, 8))

	for range 300 {
		in := make([]uint, 1+rng.IntN(30))
		for i := range in {
			in[i] = uint(rng.IntN(1000))
		}
		r := roster.New(in...)

		if got, ok := r.Min(); !ok || got != slices.Min(in) {
			t.Fatalf("%v: Min() = %d, %v, want %d", in, got, ok, slices.Min(in))
		}
		if got, ok := r.Max(); !ok || got != slices.Max(in) {
			t.Fatalf("%v: Max() = %d, %v, want %d", in, got, ok, slices.Max(in))
		}
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		a, b []uint
		want bool
	}{
		{"both empty", nil, nil, true},
		{"single equal", []uint{4}, []uint{4}, true},
		{"single different", []uint{4}, []uint{5}, false},
		{"zero vs empty", []uint{0}, nil, false},
		{"empty vs non-empty", nil, []uint{1, 2}, false},
		{"non-empty vs empty", []uint{1, 2}, nil, false},
		{"same values", []uint{1, 2, 3}, []uint{1, 2, 3}, true},
		{"same values, other order and duplicates", []uint{3, 1, 2, 2}, []uint{2, 3, 1}, true},
		{"prefix", []uint{1, 2}, []uint{1, 2, 3}, false},
		{"extension", []uint{1, 2, 3}, []uint{1, 2}, false},
		{"differ at first", []uint{1, 5, 9}, []uint{2, 5, 9}, false},
		{"differ in middle", []uint{1, 5, 9}, []uint{1, 6, 9}, false},
		{"differ at last", []uint{1, 5, 9}, []uint{1, 5, 8}, false},
		{"same size, disjoint", []uint{1, 3, 5}, []uint{2, 4, 6}, false},
		{"large values", []uint{0, 1 << 20}, []uint{0, 1 << 20}, true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, b := roster.New(tc.a...), roster.New(tc.b...)

			if got := a.Equal(b); got != tc.want {
				t.Errorf("%v.Equal(%v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
			// Equality is symmetric.
			if got := b.Equal(a); got != tc.want {
				t.Errorf("%v.Equal(%v) = %v, want %v", tc.b, tc.a, got, tc.want)
			}
			// Equality is reflexive.
			if !a.Equal(a) || !b.Equal(b) {
				t.Error("Equal is not reflexive")
			}
		})
	}
}

func TestEqualZeroValue(t *testing.T) {
	t.Parallel()
	var zero roster.Roster[uint]

	if !zero.Equal(roster.New[uint]()) || !roster.New[uint]().Equal(zero) {
		t.Error("zero value differs from empty set")
	}
	if !zero.Equal(zero) {
		t.Error("zero value differs from itself")
	}
}

func TestEqualIndependentOfHistory(t *testing.T) {
	t.Parallel()
	want := roster.New[uint](2, 4, 6)

	// Same content, different construction history and storage capacity.
	var viaInsert roster.Roster[uint]
	for _, n := range []uint{6, 2, 8, 4, 2} {
		viaInsert.Insert(n)
	}
	viaInsert.Delete(8)

	viaDeleteAll := roster.New[uint](1, 2, 3, 4, 5, 6, 7, 8, 9)
	viaDeleteAll.DeleteAll()
	for _, n := range []uint{2, 4, 6} {
		viaDeleteAll.Insert(n)
	}

	for name, got := range map[string]roster.Roster[uint]{
		"insert/delete": viaInsert,
		"delete all":    viaDeleteAll,
	} {
		if !got.Equal(want) || !want.Equal(got) {
			t.Errorf("%s: %v != %v", name, values(got), values(want))
		}
	}
}

func TestEqualDoesNotModifyOperands(t *testing.T) {
	t.Parallel()
	a, b := roster.New[uint](1, 2, 3), roster.New[uint](1, 2, 4)

	_ = a.Equal(b)

	if got, want := values(a), []uint{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("a = %v, want %v", got, want)
	}
	if got, want := values(b), []uint{1, 2, 4}; !slices.Equal(got, want) {
		t.Errorf("b = %v, want %v", got, want)
	}
}

func TestEqualAgainstReference(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(9, 10))

	randomInput := func() ([]uint, map[uint]struct{}) {
		in := make([]uint, rng.IntN(6)) // small: equal sets occur frequently
		ref := make(map[uint]struct{})
		for i := range in {
			in[i] = uint(rng.IntN(4))
			ref[in[i]] = struct{}{}
		}
		return in, ref
	}

	equalCount := 0
	for range 1000 {
		inA, refA := randomInput()
		inB, refB := randomInput()
		want := maps.Equal(refA, refB)
		if want {
			equalCount++
		}

		a, b := roster.New(inA...), roster.New(inB...)

		if got := a.Equal(b); got != want {
			t.Fatalf("%v.Equal(%v) = %v, want %v", inA, inB, got, want)
		}
	}
	if equalCount == 0 {
		t.Error("test setup: no pair of equal sets generated")
	}
}

func TestValues(t *testing.T) {
	t.Parallel()
	r := roster.New[uint](5, 1, 3, 1)

	if got, want := slices.Collect(r.Values()), []uint{1, 3, 5}; !slices.Equal(got, want) {
		t.Errorf("Values = %v, want %v", got, want)
	}

	// Early termination must be honored.
	var seen []uint
	for v := range r.Values() {
		seen = append(seen, v)
		break
	}
	if want := []uint{1}; !slices.Equal(seen, want) {
		t.Errorf("after break: %v, want %v", seen, want)
	}

	// Empty set yields nothing.
	if got := slices.Collect(roster.New[uint]().Values()); len(got) != 0 {
		t.Errorf("empty Values = %v, want nothing", got)
	}
}

func TestString(t *testing.T) {
	testcases := []struct {
		name string
		vals []uint
		exp  string
	}{
		{"empty", nil, ""},
		{"single", []uint{5}, "5"},
		{"multiple", []uint{1, 3, 10}, "1 3 10"},
		{"word boundary", []uint{63, 64}, "63 64"},
		{"empty words", []uint{1000}, "1000"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var bs roster.Roster[uint]
			for _, n := range tc.vals {
				bs.Insert(n)
			}
			if got := bs.String(); got != tc.exp {
				t.Errorf("String() = %q, exp %q", got, tc.exp)
			}
		})
	}
}

type failOnWrite struct {
	writes int
	failAt int
	err    error
}

func (w *failOnWrite) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		return 0, w.err
	}
	return len(p), nil
}

func TestRosterWriteToWriteError(t *testing.T) {
	wantErr := errors.New("write failed")

	tests := []struct {
		name   string
		failAt int
		wantN  int64
	}{
		{"value", 1, 0},
		{"separator", 2, 1},
		{"value after separator", 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := roster.New[uint](1, 2, 3)

			w := &failOnWrite{
				failAt: tt.failAt,
				err:    wantErr,
			}

			n, err := r.WriteTo(w)

			if !errors.Is(err, wantErr) {
				t.Fatalf("err = %v, want %v", err, wantErr)
			}
			if n != tt.wantN {
				t.Fatalf("n = %d, want %d", n, tt.wantN)
			}
		})
	}
}

func TestDelta(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name         string
		a, b         []uint
		onlyA, onlyB []uint
	}{
		{"basic", []uint{1, 2, 3, 100}, []uint{3, 4, 200}, []uint{1, 2, 100}, []uint{4, 200}},
		{"a longer", []uint{1, 200}, []uint{1}, []uint{200}, nil},
		{"b longer", []uint{1}, []uint{1, 200}, nil, []uint{200}},
		{"both empty", nil, nil, nil, nil},
		{"a empty", nil, []uint{5, 70}, nil, []uint{5, 70}},
		{"b empty", []uint{5, 70}, nil, []uint{5, 70}, nil},
		{"identical", []uint{0, 63, 64, 130}, []uint{0, 63, 64, 130}, nil, nil},
		{"disjoint", []uint{0, 2, 4}, []uint{1, 3, 5}, []uint{0, 2, 4}, []uint{1, 3, 5}},
		{"word boundaries", []uint{31, 32, 63, 64}, []uint{32, 64, 65}, []uint{31, 63}, []uint{65}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, b := roster.New(tc.a...), roster.New(tc.b...)

			onlyA, onlyB := a.Delta(b)

			if want := roster.New(tc.onlyA...); !onlyA.Equal(want) {
				t.Errorf("onlyA = %v, want %v", onlyA, want)
			}
			if want := roster.New(tc.onlyB...); !onlyB.Equal(want) {
				t.Errorf("onlyB = %v, want %v", onlyB, want)
			}
		})
	}
}

type (
	uis        = []uint
	uintRoster = roster.Roster[uint]
)

// setCase describes two sets and the expected results of all set operations.
type setCase struct {
	name string
	a, b uis

	or      uis // a ∪ b
	and     uis // a ∩ b
	andNot  uis // a \ b (also the first result of a.Delta(b))
	bAndNot uis // b \ a (also the second result of a.Delta(b))
	xor     uis // a Δ b
}

// Columns: name, a, b, or, and, andNot, bAndNot, xor.
var setCases = []setCase{
	{"both empty", nil, nil, nil, nil, nil, nil, nil},
	{"a empty", nil, uis{5, 70}, uis{5, 70}, nil, nil, uis{5, 70}, uis{5, 70}},
	{"b empty", uis{5, 70}, nil, uis{5, 70}, nil, uis{5, 70}, nil, uis{5, 70}},
	{"zero vs empty", uis{0}, nil, uis{0}, nil, uis{0}, nil, uis{0}},
	{"identical", uis{0, 63, 64, 130}, uis{0, 63, 64, 130}, uis{0, 63, 64, 130}, uis{0, 63, 64, 130}, nil, nil, nil},
	{"different singles", uis{3}, uis{4}, uis{3, 4}, nil, uis{3}, uis{4}, uis{3, 4}},
	{"disjoint interleaved", uis{0, 2, 4}, uis{1, 3, 5}, uis{0, 1, 2, 3, 4, 5}, nil, uis{0, 2, 4}, uis{1, 3, 5}, uis{0, 1, 2, 3, 4, 5}},
	{"disjoint, a below b", uis{1, 2}, uis{200, 300}, uis{1, 2, 200, 300}, nil, uis{1, 2}, uis{200, 300}, uis{1, 2, 200, 300}},
	{"disjoint, b below a", uis{200, 300}, uis{1, 2}, uis{1, 2, 200, 300}, nil, uis{200, 300}, uis{1, 2}, uis{1, 2, 200, 300}},
	{"overlap", uis{1, 2, 3, 100}, uis{3, 4, 200}, uis{1, 2, 3, 4, 100, 200}, uis{3}, uis{1, 2, 100}, uis{4, 200}, uis{1, 2, 4, 100, 200}},
	{"a subset of b", uis{2, 4}, uis{1, 2, 3, 4}, uis{1, 2, 3, 4}, uis{2, 4}, nil, uis{1, 3}, uis{1, 3}},
	{"b subset of a", uis{1, 2, 3, 4}, uis{2, 4}, uis{1, 2, 3, 4}, uis{2, 4}, uis{1, 3}, nil, uis{1, 3}},
	{"common only first", uis{1, 5, 9}, uis{1, 6, 10}, uis{1, 5, 6, 9, 10}, uis{1}, uis{5, 9}, uis{6, 10}, uis{5, 6, 9, 10}},
	{"common only middle", uis{1, 5, 9}, uis{2, 5, 10}, uis{1, 2, 5, 9, 10}, uis{5}, uis{1, 9}, uis{2, 10}, uis{1, 2, 9, 10}},
	{"common only last", uis{1, 5, 9}, uis{2, 6, 9}, uis{1, 2, 5, 6, 9}, uis{9}, uis{1, 5}, uis{2, 6}, uis{1, 2, 5, 6}},
	{"b has tail", uis{1, 2}, uis{2, 3, 4}, uis{1, 2, 3, 4}, uis{2}, uis{1}, uis{3, 4}, uis{1, 3, 4}},
	{"a has tail", uis{2, 3, 4}, uis{1, 2}, uis{1, 2, 3, 4}, uis{2}, uis{3, 4}, uis{1}, uis{1, 3, 4}},
	{"zero in both", uis{0, 5}, uis{0, 6}, uis{0, 5, 6}, uis{0}, uis{5}, uis{6}, uis{5, 6}},
	{"several drops in place", uis{1, 2, 3, 4, 5}, uis{2, 4, 6}, uis{1, 2, 3, 4, 5, 6}, uis{2, 4}, uis{1, 3, 5}, uis{6}, uis{1, 3, 5, 6}},
	{"alternating", uis{1, 3, 5, 7, 9}, uis{3, 4, 5, 6, 7}, uis{1, 3, 4, 5, 6, 7, 9}, uis{3, 5, 7}, uis{1, 9}, uis{4, 6}, uis{1, 4, 6, 9}},
	{"large values", uis{0, 1 << 20}, uis{1 << 20, 1 << 21}, uis{0, 1 << 20, 1 << 21}, uis{1 << 20}, uis{0}, uis{1 << 21}, uis{0, 1 << 21}},
}

func check(t *testing.T, what string, got uintRoster, want uis) {
	t.Helper()
	if !got.Equal(roster.New(want...)) {
		t.Errorf("%s = %v, want %v", what, values(got), want)
	}
}

// checkSetOperations applies every set operation to fresh copies of tc.a and
// tc.b and compares the results with the expectations of tc.
func checkSetOperations(t *testing.T, tc setCase) {
	t.Helper()
	a, b := roster.New(tc.a...), roster.New(tc.b...)

	// Non-mutating operations, in both operand orders.
	check(t, "Union(a, b)", a.Union(b), tc.or)
	check(t, "Union(b, a)", b.Union(a), tc.or)
	check(t, "Intersection(a, b)", a.Intersection(b), tc.and)
	check(t, "Intersection(b, a)", b.Intersection(a), tc.and)
	check(t, "Difference(a, b)", a.Difference(b), tc.andNot)
	check(t, "Difference(b, a)", b.Difference(a), tc.bAndNot)
	check(t, "SymmetricDifference(a, b)", a.SymmetricDifference(b), tc.xor)
	check(t, "SymmetricDifference(b, a)", b.SymmetricDifference(a), tc.xor)

	first, second := a.Delta(b)
	check(t, "Delta(a, b), first result", first, tc.andNot)
	check(t, "Delta(a, b), second result", second, tc.bAndNot)
	first, second = b.Delta(a)
	check(t, "Delta(b, a), first result", first, tc.bAndNot)
	check(t, "Delta(b, a), second result", second, tc.andNot)

	check(t, "a after non-mutating operations", a, tc.a)
	check(t, "b after non-mutating operations", b, tc.b)

	// Relations between the actual results (independent of the table).
	and, andNot, bAndNot := a.Intersection(b), a.Difference(b), b.Difference(a)
	or, xor := a.Union(b), a.SymmetricDifference(b)
	check(t, "a == (a∩b) ∪ (a\\b)", and.Union(andNot), values(a))
	check(t, "b == (a∩b) ∪ (b\\a)", and.Union(bAndNot), values(b))
	check(t, "a∪b == (a∩b) ∪ (aΔb)", and.Union(xor), values(or))
	check(t, "aΔb == (a\\b) ∪ (b\\a)", andNot.Union(bAndNot), values(xor))
	check(t, "(a∩b) ∩ (aΔb)", and.Intersection(xor), nil)

	// Mutating operations, in both operand orders.
	ops := []struct {
		name           string
		apply          func(*uintRoster, uintRoster)
		wantAB, wantBA uis // r=a, other=b and r=b, other=a
	}{
		{"Or", (*uintRoster).Or, tc.or, tc.or},
		{"And", (*uintRoster).And, tc.and, tc.and},
		{"AndNot", (*uintRoster).AndNot, tc.andNot, tc.bAndNot},
		{"Xor", (*uintRoster).Xor, tc.xor, tc.xor},
	}
	for _, op := range ops {
		for _, dir := range []struct {
			name              string
			from, other, want uis
		}{
			{"a op b", tc.a, tc.b, op.wantAB},
			{"b op a", tc.b, tc.a, op.wantBA},
		} {
			r, other := roster.New(dir.from...), roster.New(dir.other...)

			op.apply(&r, other)

			check(t, op.name+" ("+dir.name+")", r, dir.want)
			check(t, op.name+" ("+dir.name+"), operand other", other, dir.other)
		}
	}
}

func TestSetOperations(t *testing.T) {
	t.Parallel()
	for _, tc := range setCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkSetOperations(t, tc)
		})
	}
}

func TestSetOperationsRandom(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(13, 14))

	for i := range 1000 {
		densityA, densityB := 1+rng.IntN(5), 1+rng.IntN(5) // 1 = all, 5 = about 20%
		tc := setCase{name: fmt.Sprintf("random %d", i)}
		for n := range uint(60) {
			inA, inB := rng.IntN(densityA) == 0, rng.IntN(densityB) == 0
			if inA {
				tc.a = append(tc.a, n)
			}
			if inB {
				tc.b = append(tc.b, n)
			}
			if inA || inB {
				tc.or = append(tc.or, n)
			}
			if inA && inB {
				tc.and = append(tc.and, n)
			}
			if inA && !inB {
				tc.andNot = append(tc.andNot, n)
				tc.xor = append(tc.xor, n)
			}
			if !inA && inB {
				tc.bAndNot = append(tc.bAndNot, n)
				tc.xor = append(tc.xor, n)
			}
		}

		checkSetOperations(t, tc)

		if t.Failed() {
			t.Logf("failing input: a=%v b=%v", tc.a, tc.b)
			return
		}
	}
}

func TestSetOperationsWithSelf(t *testing.T) {
	t.Parallel()
	in := uis{1, 2, 3, 4, 5}

	mutating := []struct {
		name  string
		apply func(*uintRoster, uintRoster)
		want  uis
	}{
		{"Or", (*uintRoster).Or, in},
		{"And", (*uintRoster).And, in},
		{"AndNot", (*uintRoster).AndNot, nil},
		{"Xor", (*uintRoster).Xor, nil},
	}
	for _, op := range mutating {
		r := roster.New(in...)
		op.apply(&r, r) // the operand shares its storage with the receiver
		check(t, op.name+"(self)", r, op.want)
	}

	r := roster.New(in...)
	check(t, "Union(self)", r.Union(r), in)
	check(t, "Intersection(self)", r.Intersection(r), in)
	check(t, "Difference(self)", r.Difference(r), nil)
	check(t, "SymmetricDifference(self)", r.SymmetricDifference(r), nil)
	onlyR, onlyOther := r.Delta(r)
	check(t, "Delta(self), first result", onlyR, nil)
	check(t, "Delta(self), second result", onlyOther, nil)
	check(t, "operand after non-mutating operations", r, in)
}

func TestSetOperationsNoSharedStorage(t *testing.T) {
	t.Parallel()
	producers := []struct {
		name string
		f    func(a, b uintRoster) []uintRoster
	}{
		{"Union", func(a, b uintRoster) []uintRoster { return []uintRoster{a.Union(b)} }},
		{"Intersection", func(a, b uintRoster) []uintRoster { return []uintRoster{a.Intersection(b)} }},
		{"Difference", func(a, b uintRoster) []uintRoster { return []uintRoster{a.Difference(b)} }},
		{"SymmetricDifference", func(a, b uintRoster) []uintRoster { return []uintRoster{a.SymmetricDifference(b)} }},
		{"Delta", func(a, b uintRoster) []uintRoster {
			x, y := a.Delta(b)
			return []uintRoster{x, y}
		}},
	}
	inputs := [][2]uis{
		{{1, 2, 3}, {2, 4}},
		{{2, 4}, {1, 2, 3}},
		{{1, 2}, {1, 2}},
		{nil, {5}},
		{{5}, nil}, // a result equal to an operand must still be a copy
	}

	for _, p := range producers {
		for _, in := range inputs {
			what := fmt.Sprintf("%s(%v, %v)", p.name, in[0], in[1])
			a, b := roster.New(in[0]...), roster.New(in[1]...)

			// Modifying the results must not affect the operands.
			for _, r := range p.f(a, b) {
				r.Delete(1)
				r.Delete(5)
				r.Insert(1000)
			}
			check(t, what+": a after modifying results", a, in[0])
			check(t, what+": b after modifying results", b, in[1])

			// Modifying the operands must not affect the results.
			results := p.f(a, b)
			want := make([]uis, len(results))
			for i, r := range results {
				want[i] = values(r)
			}
			for _, n := range in[0] {
				a.Delete(n)
			}
			for _, n := range in[1] {
				b.Delete(n)
			}
			a.Insert(900)
			b.Insert(901)
			for i, r := range results {
				check(t, fmt.Sprintf("%s: result %d after modifying operands", what, i), r, want[i])
			}
		}
	}
}

func TestInPlaceOperationsThenReuse(t *testing.T) {
	t.Parallel()
	r := roster.New[uint](1, 2, 3, 4, 5)

	r.And(roster.New[uint](2, 4, 6))
	check(t, "after And", r, uis{2, 4})

	// Values dropped by And must not reappear from the retained storage.
	r.Insert(9)
	r.Insert(3)
	check(t, "after Insert", r, uis{2, 3, 4, 9})

	r.AndNot(roster.New[uint](2, 9))
	check(t, "after AndNot", r, uis{3, 4})

	r.Insert(1)
	r.Insert(100)
	check(t, "after second Insert", r, uis{1, 3, 4, 100})
}

// ----- Memory management

func TestClone(t *testing.T) {
	original := roster.New(uint(1), 64, 1000)
	clone := original.Clone()
	if !original.Equal(clone) {
		t.Fatalf("Clone() = %v, exp %v", clone, original)
	}
}

func TestCloneIndependent(t *testing.T) {
	original := roster.New(uint(1), 64)
	clone := original.Clone()

	clone.Insert(1000)
	clone.Delete(1)

	if got, exp := clone.String(), "64 1000"; got != exp {
		t.Fatalf("clone = %v, exp %v", got, exp)
	}
	if got, exp := original.String(), "1 64"; got != exp {
		t.Fatalf("original = %v, exp %v", got, exp)
	}
}

func TestCloneEmpty(t *testing.T) {
	var bs roster.Roster[uint32]

	clone := bs.Clone()
	if !clone.Equal(bs) {
		t.Fatal("Clone() of empty BitSet is not equal")
	}
}
