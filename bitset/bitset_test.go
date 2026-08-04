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

package bitset_test

import (
	"fmt"
	"slices"
	"testing"

	"t73f.de/r/zero/bitset"
)

func TestNewSparse(t *testing.T) {
	bs := bitset.New(uint(1), 1_000_000)
	if !bs.Contains(1) || !bs.Contains(1_000_000) {
		t.Fatal("New() did not contain inserted values")
	}
	if got := bs.Count(); got != 2 {
		t.Fatalf("Count() = %d, exp 2", got)
	}
}

func TestNewAndCollect(t *testing.T) {
	testcases := []struct {
		name string
		vals []uint
		exp  string
	}{
		{name: "nil", vals: nil, exp: "{}"},
		{name: "empty", vals: []uint{}, exp: "{}"},
		{name: "values", vals: []uint{1, 2, 63, 64, 1000}, exp: "{1,2,63,64,1000}"},
		{name: "duplicates", vals: []uint{1, 1, 2, 2, 1}, exp: "{1,2}"},
		{name: "sparse values", vals: []uint{1, 10000, 100000}, exp: "{1,10000,100000}"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("New", func(t *testing.T) {
				bs := bitset.New(tc.vals...)

				if got := bs.String(); got != tc.exp {
					t.Fatalf("New() = %v, exp %v", got, tc.exp)
				}
			})

			t.Run("Collect", func(t *testing.T) {
				bs := bitset.Collect(slices.Values(tc.vals))

				if got := bs.String(); got != tc.exp {
					t.Fatalf("Collect() = %v, exp %v", got, tc.exp)
				}
			})
		})
	}
}

func TestNewCollectByte(t *testing.T) {
	vals := []byte{0, 7, 8, 255}
	exp := "{0,7,8,255}"

	bs := bitset.New(vals...)
	if got := bs.String(); got != exp {
		t.Fatalf("Collect() = %v, exp %v", got, exp)
	}

	bs = bitset.Collect(slices.Values(vals))
	if got := bs.String(); got != exp {
		t.Fatalf("Collect() = %v, exp %v", got, exp)
	}
}

func TestInsertAndContains(t *testing.T) {
	var bs bitset.BitSet

	tests := []uint{
		0,
		1,
		7,
		8,
		63,
		64,
		100,
		1000,
	}

	for _, n := range tests {
		if bs.Contains(n) {
			t.Errorf("Contains(%d) = true on empty BitSet, exp false", n)
		}
	}
	for _, n := range tests {
		bs.Insert(n)
	}
	for _, n := range tests {
		if !bs.Contains(n) {
			t.Errorf("Contains(%d) = false, exp true", n)
		}
	}

	// Nicht gesetzte Bits prüfen
	notSet := []uint{
		2,
		9,
		62,
		65,
		999,
	}
	for _, n := range notSet {
		if bs.Contains(n) {
			t.Errorf("Contains(%d) = true, exp false", n)
		}
	}
}

func TestInsertDuplicate(t *testing.T) {
	var bs bitset.BitSet
	const val = 42

	bs.Insert(val)
	bs.Insert(val)
	if !bs.Contains(val) {
		t.Error("Contains(42) = false after duplicate Insert")
	}

	bs.Delete(val)
	if bs.Contains(val) {
		t.Errorf("Contains(%d) = true after Delete", val)
	}
}
func TestInsertGrowth(t *testing.T) {
	var bs bitset.BitSet

	for i := range uint(10000) {
		bs.Insert(i)
	}

	for i := range uint(10000) {
		if !bs.Contains(i) {
			t.Fatalf("missing bit %d after growth", i)
		}
	}
}

func TestDelete(t *testing.T) {
	var bs bitset.BitSet

	values := []uint{
		0,
		1,
		7,
		8,
		63,
		64,
		100,
	}

	for _, n := range values {
		bs.Insert(n)
	}

	for _, n := range values {
		if !bs.Contains(n) {
			t.Fatalf("Contains(%d) = false after Insert", n)
		}
	}

	for _, n := range values {
		bs.Delete(n)

		if bs.Contains(n) {
			t.Errorf("Contains(%d) = true after Delete", n)
		}
	}
}

func TestDeleteNonExisting(t *testing.T) {
	var bs bitset.BitSet

	bs.Delete(0)
	bs.Delete(1000)

	bs.Insert(42)
	bs.Delete(100)
	if !bs.Contains(42) {
		t.Error("Contains(42) = false after deleting unrelated bit")
	}
}

func TestCount(t *testing.T) {
	testcases := []struct {
		name string
		vals []uint
		exp  int
	}{
		{name: "empty", vals: nil, exp: 0},
		{name: "single word", vals: []uint{0, 1, 7, 63}, exp: 4},
		{name: "multiple words", vals: []uint{0, 63, 64, 1000}, exp: 4},
		{name: "duplicates", vals: []uint{1, 1, 2, 2, 2}, exp: 2},
		{name: "sparse", vals: []uint{1, 1_000_000}, exp: 2},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			bs := bitset.Collect(slices.Values(tc.vals))
			if got := bs.Count(); got != tc.exp {
				t.Fatalf("Count() = %d, exp %d", got, tc.exp)
			}
		})
	}
}

func TestCountAfterDelete(t *testing.T) {
	bs := bitset.New(uint(1), 2, 100)
	if got := bs.Count(); got != 3 {
		t.Fatalf("Count() = %d, exp 3", got)
	}

	bs.Delete(2)
	if got := bs.Count(); got != 2 {
		t.Fatalf("Count() after Delete = %d, exp 2", got)
	}

	bs.Delete(999) // non-existing
	if got := bs.Count(); got != 2 {
		t.Fatalf("Count() after deleting missing value = %d, exp 2", got)
	}
}

func TestIsEmpty(t *testing.T) {
	testcases := []struct {
		name string
		vals []uint
		exp  bool
	}{
		{name: "empty", vals: nil, exp: true},
		{name: "one value", vals: []uint{1}, exp: false},
		{name: "multiple values", vals: []uint{1, 64, 1000}, exp: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			bs := bitset.Collect(slices.Values(tc.vals))

			if got := bs.IsEmpty(); got != tc.exp {
				t.Fatalf("IsEmpty() = %v, exp %v", got, tc.exp)
			}
		})
	}
}

func TestIsEmptyWithAllocatedWords(t *testing.T) {
	bs := bitset.BitSet{}
	bs.EnsureBit(10000)

	if !bs.IsEmpty() {
		t.Fatal("IsEmpty() = false, exp true")
	}
}

func TestMinMax(t *testing.T) {
	testcases := []struct {
		name  string
		vals  []uint
		min   uint
		max   uint
		valid bool
	}{
		{name: "empty", valid: false},
		{name: "single", vals: []uint{42}, min: 42, max: 42, valid: true},
		{name: "multiple", vals: []uint{1000, 1, 64, 7}, min: 1, max: 1000, valid: true},
		{name: "duplicates", vals: []uint{5, 5, 5}, min: 5, max: 5, valid: true},
		{name: "word boundary", vals: []uint{63, 64}, min: 63, max: 64, valid: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			bs := bitset.New(tc.vals...)

			if got, ok := bs.Min(); ok != tc.valid || (ok && got != tc.min) {
				t.Fatalf("Min() = (%d, %v), want (%d, %v)", got, ok, tc.min, tc.valid)
			}

			if got, ok := bs.Max(); ok != tc.valid || (ok && got != tc.max) {
				t.Fatalf("Max() = (%d, %v), want (%d, %v)", got, ok, tc.max, tc.valid)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	testcases := []struct {
		name string
		a    []uint
		b    []uint
		exp  bool
	}{
		{name: "empty sets", a: nil, b: nil, exp: true},
		{name: "same values", a: []uint{1, 2, 64, 1000}, b: []uint{1, 2, 64, 1000}, exp: true},
		{name: "different values", a: []uint{1, 2, 64}, b: []uint{1, 2, 65}, exp: false},
		{name: "different order", a: []uint{1000, 1, 64}, b: []uint{64, 1000, 1}, exp: true},
		{name: "duplicates", a: []uint{1, 1, 2, 2}, b: []uint{1, 2}, exp: true},
		{name: "different values across words", a: []uint{1, 10001}, b: []uint{1, 10000}, exp: false},
		{name: "one set has additional high word", a: []uint{1}, b: []uint{1, 64}, exp: false},
		{name: "different highest values", a: []uint{1, 1000}, b: []uint{1, 10000}, exp: false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			a := bitset.Collect(slices.Values(tc.a))
			b := bitset.Collect(slices.Values(tc.b))

			if got := a.Equal(b); got != tc.exp {
				t.Fatalf("Equal() = %v, exp %v", got, tc.exp)
			}
			if got := b.Equal(a); got != tc.exp {
				t.Fatalf("Equal() not symmetric: %v, exp %v", got, tc.exp)
			}
		})
	}
}

func TestEqualIgnoresTrailingZeroWords(t *testing.T) {
	var a, b bitset.BitSet

	a.Insert(1)

	b.EnsureBit(10000)
	b.Insert(1)

	if !a.Equal(b) {
		t.Fatal("Equal() = false, exp true")
	}
	if !b.Equal(a) {
		t.Fatal("Equal() not symmetric")
	}
}

func TestAll(t *testing.T) {
	testcases := []struct {
		vals []uint
	}{
		{nil},
		{[]uint{0, 1, 63, 64, 1000}},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprint(tc.vals), func(t *testing.T) {
			var bs bitset.BitSet
			for _, n := range tc.vals {
				bs.Insert(n)
			}
			var got []uint
			for n := range bs.All() {
				got = append(got, n)
			}
			if !slices.Equal(got, tc.vals) {
				t.Fatalf("All() = %v, exp %v", got, tc.vals)
			}
		})
	}
}

func TestAllBreak(t *testing.T) {
	var bs bitset.BitSet
	for _, n := range []uint{1, 2, 3, 5, 7, 11} {
		bs.Insert(n)
	}

	var got []uint
	for n := range bs.All() {
		got = append(got, n)
		if n == 5 {
			break
		}
	}

	exp := []uint{1, 2, 3, 5}
	if !slices.Equal(got, exp) {
		t.Fatalf("All() = %v, exp %v", got, exp)
	}
}

func TestAllDelete(t *testing.T) {
	bs := bitset.New(uint(0), uint(1), uint(64), uint(1000))

	for n := range bs.All() {
		bs.Delete(n)
	}
	if !bs.IsEmpty() {
		t.Fatalf("Delete during All() left values: %v", bs)
	}
}

func TestString(t *testing.T) {
	testcases := []struct {
		name string
		vals []uint
		exp  string
	}{
		{"empty", nil, "{}"},
		{"single", []uint{5}, "{5}"},
		{"multiple", []uint{1, 3, 10}, "{1,3,10}"},
		{"word boundary", []uint{63, 64}, "{63,64}"},
		{"empty words", []uint{1000}, "{1000}"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var bs bitset.BitSet
			for _, n := range tc.vals {
				bs.Insert(n)
			}
			if got := bs.String(); got != tc.exp {
				t.Errorf("String() = %q, exp %q", got, tc.exp)
			}
		})
	}
}

func TestClone(t *testing.T) {
	original := bitset.New(uint(1), 64, 1000)
	clone := original.Clone()
	if !original.Equal(clone) {
		t.Fatalf("Clone() = %v, exp %v", clone, original)
	}
}

func TestCloneIndependent(t *testing.T) {
	original := bitset.New(uint(1), 64)
	clone := original.Clone()

	clone.Insert(1000)
	clone.Delete(1)

	if got, exp := clone.String(), "{64,1000}"; got != exp {
		t.Fatalf("clone = %v, exp %v", got, exp)
	}
	if got, exp := original.String(), "{1,64}"; got != exp {
		t.Fatalf("original = %v, exp %v", got, exp)
	}
}

func TestCloneEmpty(t *testing.T) {
	var bs bitset.BitSet

	clone := bs.Clone()
	if !clone.Equal(bs) {
		t.Fatal("Clone() of empty BitSet is not equal")
	}
}
