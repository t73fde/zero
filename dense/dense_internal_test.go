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

package dense

import (
	"slices"
	"testing"
)

// populated returns a Dense with the keys 0..99, from which every third key
// was removed again, so that keys, used and free are not empty.
func populated() *Dense[int, uint16] {
	d := new(Dense[int, uint16])
	for i := range 100 {
		d.Encode(i)
	}
	for i := 0; i < 100; i += 3 {
		d.RemoveKey(i)
	}
	return d
}

// withSurplus returns a populated Dense whose slices have unused capacity.
func withSurplus(t *testing.T) *Dense[int, uint16] {
	t.Helper()
	d := populated()
	d.Grow(500)
	d.free = slices.Grow(d.free, 100)
	if cap(d.keys)-len(d.keys) < 500 || cap(d.used)-len(d.used) < 500 ||
		cap(d.free)-len(d.free) < 100 {
		t.Fatal("precondition: slices have no surplus capacity")
	}
	return d
}

func TestGrowReservesCapacity(t *testing.T) {
	var d Dense[int, uint16]
	for i := range 10 {
		d.Encode(i)
	}
	d.Grow(200)
	if got := cap(d.keys) - len(d.keys); got < 200 {
		t.Fatalf("keys: free capacity = %d, want >= 200", got)
	}
	if got := cap(d.used) - len(d.used); got < 200 {
		t.Fatalf("used: free capacity = %d, want >= 200", got)
	}

	// Inserting 200 further keys must not reallocate keys or used.
	keysCap, usedCap := cap(d.keys), cap(d.used)
	for i := 10; i < 210; i++ {
		d.Encode(i)
	}
	if cap(d.keys) != keysCap {
		t.Errorf("cap(keys) changed from %d to %d", keysCap, cap(d.keys))
	}
	if cap(d.used) != usedCap {
		t.Errorf("cap(used) changed from %d to %d", usedCap, cap(d.used))
	}
}

func TestGrowZeroValue(t *testing.T) {
	var d Dense[int, uint16]
	d.Grow(200)
	if cap(d.keys) < 200 || cap(d.used) < 200 {
		t.Errorf("cap(keys) = %d, cap(used) = %d, want >= 200", cap(d.keys), cap(d.used))
	}
	if len(d.keys) != 0 || len(d.used) != 0 || len(d.free) != 0 {
		t.Errorf("Grow changed lengths: keys %d, used %d, free %d",
			len(d.keys), len(d.used), len(d.free))
	}
}

func TestGrowLeavesLengthsAndFreeUntouched(t *testing.T) {
	d := populated()
	keysLen, usedLen, freeLen, freeCap := len(d.keys), len(d.used), len(d.free), cap(d.free)
	count, limit := d.Count(), d.Limit()

	d.Grow(1000)

	if len(d.keys) != keysLen || len(d.used) != usedLen || len(d.free) != freeLen {
		t.Errorf("lengths changed: keys %d/%d, used %d/%d, free %d/%d",
			len(d.keys), keysLen, len(d.used), usedLen, len(d.free), freeLen)
	}
	if cap(d.free) != freeCap {
		t.Errorf("cap(free) changed from %d to %d", freeCap, cap(d.free))
	}
	if d.Count() != count || d.Limit() != limit {
		t.Errorf("Count()/Limit() = %d/%d, want %d/%d", d.Count(), d.Limit(), count, limit)
	}
}

func TestShrinkReleasesSurplus(t *testing.T) {
	d := withSurplus(t)
	keysLen, usedLen, freeLen := len(d.keys), len(d.used), len(d.free)

	d.Shrink()

	if len(d.keys) != keysLen || len(d.used) != usedLen || len(d.free) != freeLen {
		t.Errorf("lengths changed: keys %d/%d, used %d/%d, free %d/%d",
			len(d.keys), keysLen, len(d.used), usedLen, len(d.free), freeLen)
	}
	if cap(d.keys) != len(d.keys) {
		t.Errorf("keys: cap = %d, len = %d", cap(d.keys), len(d.keys))
	}
	if cap(d.used) != len(d.used) {
		t.Errorf("used: cap = %d, len = %d", cap(d.used), len(d.used))
	}
	if cap(d.free) != len(d.free) {
		t.Errorf("free: cap = %d, len = %d", cap(d.free), len(d.free))
	}
}

func TestShrinkFuncSpare(t *testing.T) {
	tests := []struct {
		name    string
		spare   func(int) int
		allowed int // maximum unused capacity after ShrinkFunc
	}{
		{"nil", nil, 0},
		{"negative", func(int) int { return -7 }, 0},
		{"zero", func(int) int { return 0 }, 0},
		{"constant", func(int) int { return 5 }, 5},
		{"larger than surplus", func(int) int { return 1 << 20 }, 1 << 20},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := withSurplus(t)
			keysCap, usedCap, freeCap := cap(d.keys), cap(d.used), cap(d.free)
			keysLen, usedLen, freeLen := len(d.keys), len(d.used), len(d.free)

			d.ShrinkFunc(tc.spare)

			if len(d.keys) != keysLen || len(d.used) != usedLen || len(d.free) != freeLen {
				t.Errorf("lengths changed: keys %d/%d, used %d/%d, free %d/%d",
					len(d.keys), keysLen, len(d.used), usedLen, len(d.free), freeLen)
			}
			check := func(name string, length, capacity, oldCap int) {
				t.Helper()
				if capacity < length {
					t.Errorf("%s: cap = %d < len = %d", name, capacity, length)
				}
				if capacity > length+tc.allowed {
					t.Errorf("%s: cap = %d, want <= len+%d = %d",
						name, capacity, tc.allowed, length+tc.allowed)
				}
				if capacity > oldCap {
					t.Errorf("%s: cap grew from %d to %d", name, oldCap, capacity)
				}
			}
			check("keys", len(d.keys), cap(d.keys), keysCap)
			check("used", len(d.used), cap(d.used), usedCap)
			check("free", len(d.free), cap(d.free), freeCap)
		})
	}
}
