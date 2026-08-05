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
	"testing"

	"t73f.de/r/zero/bitset"
)

func FuzzBitSetOperations(f *testing.F) {
	f.Add([]byte{
		0, 0, 0, 0,
		31, 0, 0, 0,
		32, 0, 0, 0,
		63, 0, 0, 0,
		64, 0, 0, 0,
		0xff, 0xff, 0, 0,
	})
	f.Add(make([]byte, 256))
	f.Add([]byte{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	})

	f.Fuzz(func(t *testing.T, data []byte) {
		const maxValue = 1 << 18

		var bs bitset.BitSet
		ref := make(map[uint]struct{})

		for i := 0; i+3 < len(data); i += 4 {
			value := uint32(data[i]) |
				uint32(data[i+1])<<8 |
				uint32(data[i+2])<<16 |
				uint32(data[i+3])<<24

			op := value & 15
			n := uint((value >> 4) & (maxValue - 1))

			switch op {
			case 0, 1, 2, 3, 4, 5, 6, 7:
				bs.Insert(n)
				ref[n] = struct{}{}

			case 8, 9:
				bs.Delete(n)
				delete(ref, n)

			case 10, 11, 12, 13:
				if got, want := bs.Contains(n), containsRef(ref, n); got != want {
					t.Fatalf("Contains(%d) = %v, want %v", n, got, want)
				}

			case 14:
				bs.Clip()

			case 15:
				bs.DeleteAll()
				clear(ref)
			}

			checkBitSetInvariant(t, bs, ref)
		}
	})
}

func containsRef(ref map[uint]struct{}, n uint) bool {
	_, ok := ref[n]
	return ok
}

func checkBitSetInvariant(t *testing.T, bs bitset.BitSet, ref map[uint]struct{}) {
	t.Helper()

	// Cardinality must match.
	if got, want := bs.Count(), len(ref); got != want {
		t.Fatalf("Count() = %d, want %d", got, want)
	}

	// Every value in the reference set must exist in the BitSet.
	for n := range ref {
		if !bs.Contains(n) {
			t.Fatalf("Contains(%d) = false, want true", n)
		}
	}

	// Every value returned by the BitSet must exist in the reference set.
	for n := range bs.All() {
		if _, ok := ref[n]; !ok {
			t.Fatalf("All() returned %d, not in reference set", n)
		}
	}

	// Empty/non-empty invariant.
	if bs.IsEmpty() != (len(ref) == 0) {
		t.Fatalf("IsEmpty() = %v, want %v", bs.IsEmpty(), len(ref) == 0)
	}

	// Min/Max invariants.
	minValue, minOK := bs.Min()
	maxValue, maxOK := bs.Max()

	if len(ref) == 0 {
		if minOK || maxOK {
			t.Fatalf("empty BitSet has Min/Max values")
		}
		return
	}

	if !minOK || !maxOK {
		t.Fatalf("non-empty BitSet has no Min/Max")
	}

	if !bs.Contains(minValue) || !bs.Contains(maxValue) {
		t.Fatalf("Min/Max value not contained")
	}

	for n := range ref {
		if n < minValue {
			t.Fatalf("Min() = %d, but contains smaller value %d", minValue, n)
		}
		if n > maxValue {
			t.Fatalf("Max() = %d, but contains larger value %d", maxValue, n)
		}
	}
}
