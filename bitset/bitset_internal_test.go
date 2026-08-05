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
		t.Fatal("Count() = 0, want 3")
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

func TestDeleteAllEmpty(t *testing.T) {
	var bs BitSet

	bs.DeleteAll()
	if !bs.IsEmpty() {
		t.Fatal("DeleteAll() on empty set is not empty")
	}
}

func TestEnsureBit(t *testing.T) {
	var bs BitSet

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
	var bs BitSet

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
