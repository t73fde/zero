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

package dense_test

import (
	"slices"
	"strconv"
	"testing"

	"t73f.de/r/zero/dense"
)

func TestZeroValue(t *testing.T) {
	var d dense.Dense[string, uint]
	if got := d.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
	if val, found := d.Find("a"); found || val != 0 {
		t.Errorf("Find(a) = (%d, %v), want (0, false)", val, found)
	}
	if key, found := d.Decode(0); found || key != "" {
		t.Errorf("Decode(0) = (%q, %v), want (\"\", false)", key, found)
	}
}

func TestEncodeSequential(t *testing.T) {
	var d dense.Dense[string, uint]
	for i, key := range []string{"alpha", "beta", "gamma"} {
		if got := d.Encode(key); got != uint(i) {
			t.Errorf("Encode(%q) = %d, want %d", key, got, i)
		}
	}
	if got := d.Count(); got != 3 {
		t.Errorf("Count() = %d, want 3", got)
	}
}

func TestEncodeIdempotent(t *testing.T) {
	var d dense.Dense[string, uint]
	first := d.Encode("alpha")
	d.Encode("beta")
	if got := d.Encode("alpha"); got != first {
		t.Errorf("second Encode(alpha) = %d, want %d", got, first)
	}
	if got := d.Count(); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
}

func TestFindDoesNotAssign(t *testing.T) {
	var d dense.Dense[string, uint]
	if _, found := d.Find("alpha"); found {
		t.Error("Find(alpha) found an unknown key")
	}
	if got := d.Count(); got != 0 {
		t.Errorf("Count() after Find = %d, want 0", got)
	}
	want := d.Encode("alpha")
	if got, found := d.Find("alpha"); !found || got != want {
		t.Errorf("Find(alpha) = (%d, %v), want (%d, true)", got, found, want)
	}
	if got := d.Encode("beta"); got != 1 {
		t.Errorf("Encode(beta) = %d, want 1 (Find must not consume codes)", got)
	}
}

func TestDecode(t *testing.T) {
	var d dense.Dense[string, uint]
	keys := []string{"alpha", "beta", "gamma"}
	for _, key := range keys {
		d.Encode(key)
	}
	for i, want := range keys {
		if got, found := d.Decode(uint(i)); !found || got != want {
			t.Errorf("Decode(%d) = (%q, %v), want (%q, true)", i, got, found, want)
		}
	}
	for _, val := range []uint{3, 4, 1000} {
		if got, found := d.Decode(val); found || got != "" {
			t.Errorf("Decode(%d) = (%q, %v), want (\"\", false)", val, got, found)
		}
	}
}

func TestBijection(t *testing.T) {
	var d dense.Dense[string, uint32]
	const n = 1000
	for i := range n {
		key := "key" + strconv.Itoa(i)
		val := d.Encode(key)
		if val != uint32(i) {
			t.Fatalf("Encode(%q) = %d, want %d", key, val, i)
		}
	}
	if got := d.Count(); got != n {
		t.Fatalf("Count() = %d, want %d", got, n)
	}
	seen := make(map[uint32]bool, n)
	for i := range n {
		key := "key" + strconv.Itoa(i)
		val, found := d.Find(key)
		if !found {
			t.Fatalf("Find(%q) not found", key)
		}
		if seen[val] {
			t.Fatalf("code %d assigned twice", val)
		}
		seen[val] = true
		if back, hasValue := d.Decode(val); !hasValue || back != key {
			t.Fatalf("Decode(Find(%q)) = (%q, %v)", key, back, hasValue)
		}
	}
}

func TestStructKey(t *testing.T) {
	type point struct{ x, y int }
	var d dense.Dense[point, uint]
	a := d.Encode(point{1, 2})
	b := d.Encode(point{2, 1})
	if a == b {
		t.Errorf("distinct keys share value %d", a)
	}
	if got := d.Encode(point{1, 2}); got != a {
		t.Errorf("Encode(point{1,2}) = %d, want %d", got, a)
	}
	if key, found := d.Decode(b); !found || key != (point{2, 1}) {
		t.Errorf("Decode(%d) = (%v, %v)", b, key, found)
	}
}

func TestDefinedCodeType(t *testing.T) {
	type val16 uint16
	var d dense.Dense[string, val16]
	if got := d.Encode("a"); got != val16(0) {
		t.Errorf("Encode(a) = %d, want 0", got)
	}
	if got := d.Encode("b"); got != val16(1) {
		t.Errorf("Encode(b) = %d, want 1", got)
	}
	if key, found := d.Decode(val16(1)); !found || key != "b" {
		t.Errorf("Decode(1) = (%q, %v)", key, found)
	}
}

func TestRangeExhausted(t *testing.T) {
	var d dense.Dense[int, uint8]
	for i := range 256 { // values 0..255 fill the range of uint8 completely
		if got := d.Encode(i); got != uint8(i) {
			t.Fatalf("Encode(%d) = %d, want %d", i, got, i)
		}
	}
	// Known keys remain accessible after the range is full.
	if got := d.Encode(255); got != 255 {
		t.Errorf("Encode(255) = %d, want 255", got)
	}
	if got := d.Encode(0); got != 0 {
		t.Errorf("Encode(0) = %d, want 0", got)
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Error("Encode of a new key did not panic on exhausted range")
			}
		}()
		d.Encode(256)
	}()

	// The failed Encode must not have changed the state.
	if got := d.Count(); got != 256 {
		t.Errorf("Count() after panic = %d, want 256", got)
	}
	if _, found := d.Find(256); found {
		t.Error("key 256 was registered despite the panic")
	}
}

func TestRemoveKey(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("a")
	b := d.Encode("b")
	d.Encode("c")

	val, found := d.RemoveKey("b")
	if !found || val != b {
		t.Fatalf("Delete(b) = (%d, %v), want (%d, true)", val, found, b)
	}
	if _, found = d.Find("b"); found {
		t.Error("Find(b) found a deleted key")
	}
	if key, keyFound := d.Decode(b); keyFound || key != "" {
		t.Errorf("Decode(%d) = (%q, %v), want (\"\", false)", b, key, keyFound)
	}
	if got := d.Count(); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
	if got := d.Limit(); got != 3 {
		t.Errorf("Limit() = %d, want 3", got)
	}
	if val, found = d.RemoveKey("b"); found || val != 0 {
		t.Errorf("second Delete(b) = (%d, %v), want (0, false)", val, found)
	}
	if val, found = d.RemoveKey("unknown"); found || val != 0 {
		t.Errorf("Delete(unknown) = (%d, %v), want (0, false)", val, found)
	}
}

func TestRemoveKeyReuse(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("a")
	b := d.Encode("b")
	c := d.Encode("c")

	d.RemoveKey("b")
	d.RemoveKey("c")
	// Most recently freed code first.
	if got := d.Encode("x"); got != c {
		t.Errorf("Encode(x) = %d, want %d", got, c)
	}
	if got := d.Encode("y"); got != b {
		t.Errorf("Encode(y) = %d, want %d", got, b)
	}
	if got := d.Encode("z"); got != 3 {
		t.Errorf("Encode(z) = %d, want 3", got)
	}
	if got := d.Limit(); got != 4 {
		t.Errorf("Limit() = %d, want 4", got)
	}
	for key, want := range map[string]uint{"a": 0, "x": c, "y": b, "z": 3} {
		if got, found := d.Find(key); !found || got != want {
			t.Errorf("Find(%q) = (%d, %v), want (%d, true)", key, got, found, want)
		}
		if back, found := d.Decode(want); !found || back != key {
			t.Errorf("Decode(%d) = (%q, %v), want (%q, true)", want, back, found, key)
		}
	}
}

func TestRemoveKeyZeroKey(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("")
	d.Encode("a")
	d.RemoveKey("a")
	if key, found := d.Decode(1); found {
		t.Errorf("Decode(1) = (%q, true) for a freed value", key)
	}
	if key, found := d.Decode(0); !found || key != "" {
		t.Errorf("Decode(0) = (%q, %v), want (\"\", true)", key, found)
	}
}

func TestRemoveKeyAfterRangeExhausted(t *testing.T) {
	var d dense.Dense[int, uint8]
	for i := range 256 {
		d.Encode(i)
	}
	val, found := d.RemoveKey(100)
	if !found || val != 100 {
		t.Fatalf("Delete(100) = (%d, %v), want (100, true)", val, found)
	}
	// The freed value makes room for exactly one new key.
	if got := d.Encode(1000); got != 100 {
		t.Errorf("Encode(1000) = %d, want 100", got)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("Encode did not panic on exhausted range")
			}
		}()
		d.Encode(1001)
	}()
}

func TestRemoveValue(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("a")
	b := d.Encode("b")
	d.Encode("c")

	key, found := d.RemoveValue(b)
	if !found || key != "b" {
		t.Fatalf("RemoveValue(%d) = (%q, %v), want (\"b\", true)", b, key, found)
	}
	if _, found = d.Find("b"); found {
		t.Error("Find(b) found a removed key")
	}
	if key, found = d.Decode(b); found || key != "" {
		t.Errorf("Decode(%d) = (%q, %v), want (\"\", false)", b, key, found)
	}
	if got := d.Count(); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
	if got := d.Limit(); got != 3 {
		t.Errorf("Limit() = %d, want 3", got)
	}
	// The remaining keys are unaffected.
	for key, want := range map[string]uint{"a": 0, "c": 2} {
		if got, foundRemaining := d.Find(key); !foundRemaining || got != want {
			t.Errorf("Find(%q) = (%d, %v), want (%d, true)", key, got, foundRemaining, want)
		}
	}
}

func TestRemoveValueNotInUse(t *testing.T) {
	var d dense.Dense[string, uint]
	if key, found := d.RemoveValue(0); found || key != "" {
		t.Errorf("RemoveValue(0) on empty Dense = (%q, %v), want (\"\", false)", key, found)
	}

	d.Encode("a")
	d.Encode("b")
	d.RemoveValue(0)
	for _, val := range []uint{0, 2, 1000} { // removed, never assigned, far outside
		if key, found := d.RemoveValue(val); found || key != "" {
			t.Errorf("RemoveValue(%d) = (%q, %v), want (\"\", false)", val, key, found)
		}
	}
	if got := d.Count(); got != 1 {
		t.Errorf("Count() = %d, want 1", got)
	}
	if got := d.Limit(); got != 2 {
		t.Errorf("Limit() = %d, want 2", got)
	}
}

// A repeated RemoveValue must not put the same value twice into the free
// list, otherwise two different keys would receive the same value.
func TestRemoveValueTwice(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("a")
	b := d.Encode("b")
	d.Encode("c")

	d.RemoveValue(b)
	d.RemoveValue(b)

	x := d.Encode("x")
	y := d.Encode("y")
	if x == y {
		t.Fatalf("Encode(x) and Encode(y) both returned %d", x)
	}
	if x != b {
		t.Errorf("Encode(x) = %d, want freed value %d", x, b)
	}
	if y != 3 {
		t.Errorf("Encode(y) = %d, want 3", y)
	}
	if got := d.Count(); got != 4 {
		t.Errorf("Count() = %d, want 4", got)
	}
	if got := d.Limit(); got != 4 {
		t.Errorf("Limit() = %d, want 4", got)
	}
}

func TestRemoveValueReuse(t *testing.T) {
	var d dense.Dense[string, uint]
	a := d.Encode("a")
	b := d.Encode("b")
	d.Encode("c")

	// Most recently freed value first.
	d.RemoveValue(a)
	d.RemoveValue(b)
	if got := d.Encode("x"); got != b {
		t.Errorf("Encode(x) = %d, want %d", got, b)
	}
	if got := d.Encode("y"); got != a {
		t.Errorf("Encode(y) = %d, want %d", got, a)
	}
	if key, found := d.Decode(a); !found || key != "y" {
		t.Errorf("Decode(%d) = (%q, %v), want (\"y\", true)", a, key, found)
	}
}

func TestRemoveValueZeroKey(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("")
	d.Encode("a")

	if key, found := d.RemoveValue(1); !found || key != "a" {
		t.Fatalf("RemoveValue(1) = (%q, %v), want (\"a\", true)", key, found)
	}
	// The freed slot holds the zero key internally, but must not be reported.
	if key, found := d.Decode(1); found {
		t.Errorf("Decode(1) = (%q, true) for a freed value", key)
	}
	// The real key "" keeps its value.
	if val, found := d.Find(""); !found || val != 0 {
		t.Errorf("Find(\"\") = (%d, %v), want (0, true)", val, found)
	}
	if key, found := d.RemoveValue(0); !found || key != "" {
		t.Errorf("RemoveValue(0) = (%q, %v), want (\"\", true)", key, found)
	}
	if got := d.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
}

func TestRemoveValueAgreesWithRemoveKey(t *testing.T) {
	var viaKey, viaValue dense.Dense[int, uint8]
	for i := range 10 {
		viaKey.Encode(i)
		viaValue.Encode(i)
	}
	for _, i := range []int{3, 7, 0, 9} {
		val, found := viaKey.RemoveKey(i)
		if !found {
			t.Fatalf("RemoveKey(%d) not found", i)
		}
		key, found := viaValue.RemoveValue(val)
		if !found || key != i {
			t.Fatalf("RemoveValue(%d) = (%d, %v), want (%d, true)", val, key, found, i)
		}
	}
	// Both must continue identically.
	for i := 100; i < 106; i++ {
		if a, b := viaKey.Encode(i), viaValue.Encode(i); a != b {
			t.Errorf("Encode(%d) = %d (RemoveKey) vs. %d (RemoveValue)", i, a, b)
		}
	}
	if viaKey.Count() != viaValue.Count() || viaKey.Limit() != viaValue.Limit() {
		t.Errorf("Count/Limit differ: (%d, %d) vs. (%d, %d)",
			viaKey.Count(), viaKey.Limit(), viaValue.Count(), viaValue.Limit())
	}
}

func TestRemoveValueFreesRange(t *testing.T) {
	var d dense.Dense[int, uint8]
	for i := range 256 {
		d.Encode(i)
	}
	if key, found := d.RemoveValue(200); !found || key != 200 {
		t.Fatalf("RemoveValue(200) = (%d, %v), want (200, true)", key, found)
	}
	// The freed value makes room for exactly one new key.
	if got := d.Encode(1000); got != 200 {
		t.Errorf("Encode(1000) = %d, want 200", got)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("Encode did not panic on exhausted range")
			}
		}()
		d.Encode(1001)
	}()
	if got := d.Count(); got != 256 {
		t.Errorf("Count() after panic = %d, want 256", got)
	}
}

type pair struct {
	key string
	val uint
}

func collectAll(d *dense.Dense[string, uint]) []pair {
	var result []pair
	for key, val := range d.All() {
		result = append(result, pair{key, val})
	}
	return result
}

// newABCD returns a Dense in which the freed value 0 of "a" was reused for
// "d", so that the order of the values differs from the order of insertion.
func newABCD() *dense.Dense[string, uint] {
	d := new(dense.Dense[string, uint])
	d.Encode("a")
	d.Encode("b")
	d.Encode("c")
	d.RemoveKey("a")
	d.Encode("d")
	return d
}

func TestAll(t *testing.T) {
	var empty dense.Dense[string, uint]
	if got := collectAll(&empty); len(got) != 0 {
		t.Errorf("All() of zero value = %v, want nothing", got)
	}

	d := newABCD()
	want := []pair{{"d", 0}, {"b", 1}, {"c", 2}}
	if got := collectAll(d); !slices.Equal(got, want) {
		t.Errorf("All() = %v, want %v", got, want)
	}

	d.RemoveKey("b")
	want = []pair{{"d", 0}, {"c", 2}}
	if got := collectAll(d); !slices.Equal(got, want) {
		t.Errorf("All() after RemoveKey(b) = %v, want %v", got, want)
	}
}

func TestAllConsistent(t *testing.T) {
	d := populate(100)
	n := 0
	prev := -1
	for key, val := range d.All() {
		n++
		if int(val) <= prev {
			t.Fatalf("values not ascending: %d after %d", val, prev)
		}
		prev = int(val)
		if got, found := d.Find(key); !found || got != val {
			t.Fatalf("Find(%d) = (%d, %v), want (%d, true)", key, got, found, val)
		}
		if got, found := d.Decode(val); !found || got != key {
			t.Fatalf("Decode(%d) = (%d, %v), want (%d, true)", val, got, found, key)
		}
	}
	if n != d.Count() {
		t.Errorf("All() produced %d pairs, Count() = %d", n, d.Count())
	}
}

func TestKeysValues(t *testing.T) {
	d := newABCD()
	if got, want := slices.Collect(d.Keys()), []string{"d", "b", "c"}; !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
	if got, want := slices.Collect(d.Values()), []uint{0, 1, 2}; !slices.Equal(got, want) {
		t.Errorf("Values() = %v, want %v", got, want)
	}

	var empty dense.Dense[string, uint]
	if got := slices.Collect(empty.Keys()); len(got) != 0 {
		t.Errorf("Keys() of zero value = %v, want nothing", got)
	}
	if got := slices.Collect(empty.Values()); len(got) != 0 {
		t.Errorf("Values() of zero value = %v, want nothing", got)
	}
}

func TestIteratorsStopEarly(t *testing.T) {
	var d dense.Dense[int, uint]
	for i := range 10 {
		d.Encode(i)
	}
	stopAtThree := func(t *testing.T, name string, count func() int) {
		t.Helper()
		if got := count(); got != 3 {
			t.Errorf("%s: %d iterations, want 3", name, got)
		}
	}
	stopAtThree(t, "All", func() int {
		n := 0
		for range d.All() {
			if n++; n == 3 {
				break
			}
		}
		return n
	})
	stopAtThree(t, "Keys", func() int {
		n := 0
		for range d.Keys() {
			if n++; n == 3 {
				break
			}
		}
		return n
	})
	stopAtThree(t, "Values", func() int {
		n := 0
		for range d.Values() {
			if n++; n == 3 {
				break
			}
		}
		return n
	})
}

func TestIteratorRemoveDuringIteration(t *testing.T) {
	var d dense.Dense[int, uint]
	for i := range 10 {
		d.Encode(i)
	}
	var seen []int
	for key := range d.Keys() {
		seen = append(seen, key)
		d.RemoveKey(key) // the current key
		if key == 0 {
			d.RemoveKey(9) // a key that was not yet reached
		}
	}
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	if !slices.Equal(seen, want) {
		t.Errorf("seen %v, want %v", seen, want)
	}
	if got := d.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
}

// populate returns a Dense with the keys 0..n-1, from which every third key
// was removed again, so that the free list is not empty.
func populate(n int) *dense.Dense[int, uint16] {
	d := new(dense.Dense[int, uint16])
	for i := range n {
		d.Encode(i)
	}
	for i := 0; i < n; i += 3 {
		d.RemoveKey(i)
	}
	return d
}

func assertSame(t *testing.T, got, want *dense.Dense[int, uint16], maxKey int) {
	t.Helper()
	if g, w := got.Count(), want.Count(); g != w {
		t.Fatalf("Count() = %d, want %d", g, w)
	}
	if g, w := got.Limit(), want.Limit(); g != w {
		t.Fatalf("Limit() = %d, want %d", g, w)
	}
	for key := range maxKey {
		gv, gf := got.Find(key)
		wv, wf := want.Find(key)
		if gv != wv || gf != wf {
			t.Fatalf("Find(%d) = (%d, %v), want (%d, %v)", key, gv, gf, wv, wf)
		}
	}
	for val := range want.Limit() + 1 { // including one value beyond the limit
		gk, gf := got.Decode(uint16(val))
		wk, wf := want.Decode(uint16(val))
		if gk != wk || gf != wf {
			t.Fatalf("Decode(%d) = (%d, %v), want (%d, %v)", val, gk, gf, wk, wf)
		}
	}
}

func TestCapacityOperationsKeepState(t *testing.T) {
	const n = 100
	tests := []struct {
		name string
		op   func(*dense.Dense[int, uint16])
	}{
		{"Grow", func(d *dense.Dense[int, uint16]) { d.Grow(500) }},
		{"Grow 0", func(d *dense.Dense[int, uint16]) { d.Grow(0) }},
		{"Shrink", func(d *dense.Dense[int, uint16]) { d.Shrink() }},
		{"ShrinkFunc nil", func(d *dense.Dense[int, uint16]) { d.ShrinkFunc(nil) }},
		{"ShrinkFunc negative", func(d *dense.Dense[int, uint16]) {
			d.ShrinkFunc(func(int) int { return -1 })
		}},
		{"ShrinkFunc n/8", func(d *dense.Dense[int, uint16]) {
			d.ShrinkFunc(func(n int) int { return n / 8 })
		}},
		{"ShrinkFunc constant", func(d *dense.Dense[int, uint16]) {
			d.ShrinkFunc(func(int) int { return 10 })
		}},
		{"Grow, Shrink", func(d *dense.Dense[int, uint16]) {
			d.Grow(500)
			d.Shrink()
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, want := populate(n), populate(n)
			tc.op(got)
			assertSame(t, got, want, n+50)

			// Both continue identically: first the freed values (most
			// recently freed first), then fresh ones.
			for i := n; i < n+50; i++ {
				if g, w := got.Encode(i), want.Encode(i); g != w {
					t.Fatalf("Encode(%d) = %d, want %d", i, g, w)
				}
			}
			assertSame(t, got, want, n+50)
		})
	}
}

func TestCapacityOperationsZeroValue(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Grow(0)
	d.Shrink()
	d.ShrinkFunc(nil)
	d.ShrinkFunc(func(n int) int { return n })
	if d.Count() != 0 || d.Limit() != 0 {
		t.Errorf("Count() = %d, Limit() = %d, want 0, 0", d.Count(), d.Limit())
	}
	if got := d.Encode("a"); got != 0 {
		t.Errorf("Encode(a) = %d, want 0", got)
	}

	var e dense.Dense[string, uint]
	e.Grow(10)
	if e.Count() != 0 || e.Limit() != 0 {
		t.Errorf("after Grow: Count() = %d, Limit() = %d, want 0, 0", e.Count(), e.Limit())
	}
	if got := e.Encode("a"); got != 0 {
		t.Errorf("Encode(a) after Grow = %d, want 0", got)
	}
}

func TestGrowNegative(t *testing.T) {
	var d dense.Dense[int, uint]
	defer func() {
		if recover() == nil {
			t.Error("Grow(-1) did not panic")
		}
	}()
	d.Grow(-1)
}
