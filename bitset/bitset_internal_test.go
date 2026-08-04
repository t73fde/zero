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
