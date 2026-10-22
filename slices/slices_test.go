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

package slices_test

import (
	"testing"

	zeroslices "t73f.de/r/zero/slices"
)

func TestShrink(t *testing.T) {
	t.Parallel()

	if got := zeroslices.Shrink[int](nil); got != nil {
		t.Errorf("nil: got %v, want nil", got)
	}
	if got := zeroslices.Shrink([]int{}); got != nil {
		t.Errorf("empty: got %v, want nil", got)
	}

	exact := []int{1, 2, 3}
	if got := zeroslices.Shrink(exact); &got[0] != &exact[0] {
		t.Error("cap == len: expected same storage")
	}

	big := make([]int, 3, 100)
	copy(big, []int{4, 5, 6})
	got := zeroslices.Shrink(big)
	if len(got) != 3 || cap(got) != 3 {
		t.Fatalf("len/cap: got %d/%d, want 3/3", len(got), cap(got))
	}
	for i, v := range []int{4, 5, 6} {
		if got[i] != v {
			t.Errorf("got[%d] = %d, want %d", i, got[i], v)
		}
	}
	if &got[0] == &big[0] {
		t.Error("cap > len: expected new storage")
	}
}
