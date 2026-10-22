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
	"math"
	"slices"
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

func mkSlice(length, capacity int) []int {
	s := make([]int, length, capacity)
	for i := range s {
		s[i] = i + 1
	}
	return s
}

func TestShrinkFunc(t *testing.T) {
	t.Parallel()

	eighth := func(n int) int { return n / 8 }

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		called := false
		spare := func(int) int { called = true; return 1 }
		if got := zeroslices.ShrinkFunc[int](nil, spare); got != nil {
			t.Errorf("nil: got %v, want nil", got)
		}
		if got := zeroslices.ShrinkFunc([]int{}, spare); got != nil {
			t.Errorf("empty: got %v, want nil", got)
		}
		if called {
			t.Error("spare must not be called for empty slice")
		}
	})

	t.Run("argument is length", func(t *testing.T) {
		t.Parallel()
		var arg int
		zeroslices.ShrinkFunc(mkSlice(5, 50), func(n int) int { arg = n; return 0 })
		if arg != 5 {
			t.Errorf("spare called with %d, want 5", arg)
		}
	})

	t.Run("within allowance", func(t *testing.T) {
		t.Parallel()
		s := mkSlice(8, 9) // unused 1 <= eighth(8) == 1
		got := zeroslices.ShrinkFunc(s, eighth)
		if &got[0] != &s[0] || cap(got) != 9 {
			t.Errorf("expected s itself, got cap %d", cap(got))
		}
	})

	t.Run("exceeds allowance", func(t *testing.T) {
		t.Parallel()
		s := mkSlice(8, 100)
		got := zeroslices.ShrinkFunc(s, eighth)
		if len(got) != 8 || cap(got) != 9 {
			t.Fatalf("len/cap: got %d/%d, want 8/9", len(got), cap(got))
		}
		if !slices.Equal(got, []int{1, 2, 3, 4, 5, 6, 7, 8}) {
			t.Errorf("values: got %v", got)
		}
		if &got[0] == &s[0] {
			t.Error("expected new storage")
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		t.Parallel()
		first := zeroslices.ShrinkFunc(mkSlice(64, 1000), eighth)
		second := zeroslices.ShrinkFunc(first, eighth)
		if &second[0] != &first[0] {
			t.Error("second call reallocated")
		}
		if cap(first) != 72 {
			t.Errorf("cap: got %d, want 72", cap(first))
		}
	})

	t.Run("negative spare", func(t *testing.T) {
		t.Parallel()
		got := zeroslices.ShrinkFunc(mkSlice(3, 100), func(int) int { return -5 })
		if len(got) != 3 || cap(got) != 3 {
			t.Errorf("len/cap: got %d/%d, want 3/3", len(got), cap(got))
		}
	})

	t.Run("huge spare", func(t *testing.T) {
		t.Parallel()
		s := mkSlice(3, 100)
		got := zeroslices.ShrinkFunc(s, func(int) int { return math.MaxInt })
		if &got[0] != &s[0] || cap(got) != 100 {
			t.Error("expected s itself")
		}
	})

	t.Run("nil spare", func(t *testing.T) {
		t.Parallel()
		got := zeroslices.ShrinkFunc(mkSlice(3, 100), nil)
		if len(got) != 3 || cap(got) != 3 {
			t.Errorf("len/cap: got %d/%d, want 3/3", len(got), cap(got))
		}
	})
}
