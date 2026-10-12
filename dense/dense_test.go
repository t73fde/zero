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
	"strconv"
	"testing"

	"t73f.de/r/zero/dense"
)

func TestZeroValue(t *testing.T) {
	var d dense.Dense[string, uint]
	if got := d.Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
	if code, found := d.Find("a"); found || code != 0 {
		t.Errorf("Find(a) = (%d, %v), want (0, false)", code, found)
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
	for _, code := range []uint{3, 4, 1000} {
		if got, found := d.Decode(code); found || got != "" {
			t.Errorf("Decode(%d) = (%q, %v), want (\"\", false)", code, got, found)
		}
	}
}

func TestBijection(t *testing.T) {
	var d dense.Dense[string, uint32]
	const n = 1000
	for i := range n {
		key := "key" + strconv.Itoa(i)
		code := d.Encode(key)
		if code != uint32(i) {
			t.Fatalf("Encode(%q) = %d, want %d", key, code, i)
		}
	}
	if got := d.Count(); got != n {
		t.Fatalf("Count() = %d, want %d", got, n)
	}
	seen := make(map[uint32]bool, n)
	for i := range n {
		key := "key" + strconv.Itoa(i)
		code, found := d.Find(key)
		if !found {
			t.Fatalf("Find(%q) not found", key)
		}
		if seen[code] {
			t.Fatalf("code %d assigned twice", code)
		}
		seen[code] = true
		if back, hasCode := d.Decode(code); !hasCode || back != key {
			t.Fatalf("Decode(Find(%q)) = (%q, %v)", key, back, hasCode)
		}
	}
}

func TestStructKey(t *testing.T) {
	type point struct{ x, y int }
	var d dense.Dense[point, uint]
	a := d.Encode(point{1, 2})
	b := d.Encode(point{2, 1})
	if a == b {
		t.Errorf("distinct keys share code %d", a)
	}
	if got := d.Encode(point{1, 2}); got != a {
		t.Errorf("Encode(point{1,2}) = %d, want %d", got, a)
	}
	if key, found := d.Decode(b); !found || key != (point{2, 1}) {
		t.Errorf("Decode(%d) = (%v, %v)", b, key, found)
	}
}

func TestDefinedCodeType(t *testing.T) {
	type code16 uint16
	var d dense.Dense[string, code16]
	if got := d.Encode("a"); got != code16(0) {
		t.Errorf("Encode(a) = %d, want 0", got)
	}
	if got := d.Encode("b"); got != code16(1) {
		t.Errorf("Encode(b) = %d, want 1", got)
	}
	if key, found := d.Decode(code16(1)); !found || key != "b" {
		t.Errorf("Decode(1) = (%q, %v)", key, found)
	}
}

func TestRangeExhausted(t *testing.T) {
	var d dense.Dense[int, uint8]
	for i := range 256 { // codes 0..255 fill the range of uint8 completely
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

func TestRemove(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("a")
	b := d.Encode("b")
	d.Encode("c")

	code, found := d.Remove("b")
	if !found || code != b {
		t.Fatalf("Delete(b) = (%d, %v), want (%d, true)", code, found, b)
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
	if code, found = d.Remove("b"); found || code != 0 {
		t.Errorf("second Delete(b) = (%d, %v), want (0, false)", code, found)
	}
	if code, found = d.Remove("unknown"); found || code != 0 {
		t.Errorf("Delete(unknown) = (%d, %v), want (0, false)", code, found)
	}
}

func TestRemoveReuse(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("a")
	b := d.Encode("b")
	c := d.Encode("c")

	d.Remove("b")
	d.Remove("c")
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

func TestRemoveZeroKey(t *testing.T) {
	var d dense.Dense[string, uint]
	d.Encode("")
	d.Encode("a")
	d.Remove("a")
	if key, found := d.Decode(1); found {
		t.Errorf("Decode(1) = (%q, true) for a freed code", key)
	}
	if key, found := d.Decode(0); !found || key != "" {
		t.Errorf("Decode(0) = (%q, %v), want (\"\", true)", key, found)
	}
}

func TestRemoveAfterRangeExhausted(t *testing.T) {
	var d dense.Dense[int, uint8]
	for i := range 256 {
		d.Encode(i)
	}
	code, found := d.Remove(100)
	if !found || code != 100 {
		t.Fatalf("Delete(100) = (%d, %v), want (100, true)", code, found)
	}
	// The freed code makes room for exactly one new key.
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
