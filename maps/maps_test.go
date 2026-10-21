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

package maps_test

import (
	"maps"
	"testing"

	zeromaps "t73f.de/r/zero/maps"
)

func TestGrowShrink(t *testing.T) {
	m := map[int]string{}
	if got := zeromaps.Shrink(m); got != nil {
		t.Errorf("shrink of empty map must result in nil, but got: %v", got)
	}
	m = nil
	if got := zeromaps.Grow(m, 8); got == nil {
		t.Error("got nil map after growing an empty map")
	}
	m = map[int]string{0: "zero", 1: "one", 2: "two"}
	if got := zeromaps.Shrink(m); !maps.Equal(m, got) {
		t.Errorf("shrinked map %v differs in result %v", m, got)
	}
}

func TestGrowNegative(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Grow(-1) did not panic")
		}
	}()
	zeromaps.Grow(map[int]int{}, -1)
}
