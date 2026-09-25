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

package bitset

import "testing"

func TestDeleteAll(t *testing.T) {
	bs := New[uint](1, 3, 7, 42, 100)
	capBefore := cap(bs.words)
	if bs.Count() == 0 {
		t.Fatal("Count() = 0, > 0")
	}

	bs.DeleteAll()

	if !bs.IsEmpty() {
		t.Fatal("DeleteAll(): set is not empty")
	}
	if bs.Count() != 0 {
		t.Fatalf("DeleteAll(): Count() = %d, want 0", bs.Count())
	}
	if got, ok := bs.Min(); ok {
		t.Fatalf("DeleteAll(): Min() = (%d, true), want (_, false)", got)
	}
	if got, ok := bs.Max(); ok {
		t.Fatalf("DeleteAll(): Max() = (%d, true), want (_, false)", got)
	}
	if got := bs.String(); got != "{}" {
		t.Fatalf("DeleteAll(): String() = %q, want %q", got, "{}")
	}
	if cap(bs.words) != capBefore {
		t.Fatalf("DeleteAll(): capacity changed from %d to %d", capBefore, cap(bs.words))
	}
}

func TestEnsureBit(t *testing.T) {
	var bs BitSet[uint16]

	bs.EnsureBit(1000)
	index := int(1000 / wordSizeBits)
	if len(bs.words) <= index {
		t.Fatalf("len(words) = %d, want > %d", len(bs.words), index)
	}

	if got := bs.Count(); got != 0 {
		t.Fatalf("Count() = %d, want 0", got)
	}

	bs.Insert(1000)
	if !bs.Contains(1000) {
		t.Fatal("Insert after EnsureBit failed")
	}
	if got := bs.Count(); got != 1 {
		t.Fatalf("Count() = %d, want 1", got)
	}
}

func TestClipRemovesTrailingWords(t *testing.T) {
	var bs BitSet[uint16]

	bs.EnsureBit(10000)
	bs.Clip()

	if bs.words != nil {
		t.Fatalf("words = %v, want nil", bs.words)
	}

	bs.EnsureBit(10000)
	bs.Insert(0)
	bs.Clip()

	if got, want := cap(bs.words), len(bs.words); got != want {
		t.Fatalf("cap(words) = %d, len(words) = %d", got, want)
	}
	if got := len(bs.words); got != 1 {
		t.Fatalf("len(words) = %d, want 1", got)
	}
}

func TestGrowWordsReusesCapacity(t *testing.T) {
	var bs BitSet[uint16]

	bs.EnsureBit(100) // forces first allocation
	wordsBefore := bs.words
	capBefore := cap(bs.words)

	bs.growWords(len(bs.words) - 1) // smaller value, must not change anything
	if cap(bs.words) != capBefore {
		t.Fatalf("growWords(smaller value): cap = %d, want %d", cap(bs.words), capBefore)
	}

	// reset words to force the within-cap path
	bs.words = bs.words[:1]
	bs.growWords(capBefore)
	if cap(bs.words) != capBefore {
		t.Fatalf("growWords(within cap): cap = %d, want %d (no reallocation expected)", cap(bs.words), capBefore)
	}
	if &bs.words[0] != &wordsBefore[0] {
		t.Fatal("growWords(within cap): backing array was reallocated")
	}
}

func TestGrowWordsReallocatesBeyondCapacity(t *testing.T) {
	var bs BitSet[uint16]

	bs.EnsureBit(10) // small initial allocation
	capBefore := cap(bs.words)

	bs.growWords(capBefore + 10)
	if cap(bs.words) <= capBefore {
		t.Fatalf("growWords(beyond cap): cap = %d, want > %d", cap(bs.words), capBefore)
	}
	if len(bs.words) != capBefore+10 {
		t.Fatalf("len(words) = %d, want %d", len(bs.words), capBefore+10)
	}
	// prior content must survive reallocation
	bs.Insert(5)
	if !bs.Contains(5) {
		t.Fatal("growWords: content not preserved after reallocation")
	}
}

func TestGrowWordsNoOpOnSameLength(t *testing.T) {
	var bs BitSet[uint16]
	bs.EnsureBit(100)

	before := bs.words
	bs.growWords(len(bs.words))

	if len(bs.words) != len(before) || cap(bs.words) != cap(before) {
		t.Fatal("growWords(same length): unexpected change")
	}
}
