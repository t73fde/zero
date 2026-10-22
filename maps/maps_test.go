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

type registry map[string]int

var (
	_ registry = zeromaps.Grow(registry{}, 1)
	_ registry = zeromaps.Shrink(registry{})
)

func TestGrow(t *testing.T) {
	var nilMap map[int]string
	if got := zeromaps.Grow(nilMap, 8); got == nil {
		t.Error("got nil map after growing a nil map")
	}

	m := map[int]string{0: "zero", 1: "one"}
	got := zeromaps.Grow(m, 8)
	if !maps.Equal(m, got) {
		t.Errorf("grown map %v differs from original %v", got, m)
	}
	got[2] = "two"
	if _, found := m[2]; found {
		t.Error("result of Grow shares storage with its argument")
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

func TestShrink(t *testing.T) {
	var nilMap map[int]string
	if got := zeromaps.Shrink(nilMap); got != nil {
		t.Errorf("shrunk nil map must stay nil, but got: %v", got)
	}

	empty := map[int]string{}
	if got := zeromaps.Shrink(empty); got == nil || len(got) != 0 {
		t.Errorf("shrunk empty map must be empty and non-nil, but got: %v", got)
	}

	m := map[int]string{0: "zero", 1: "one", 2: "two"}
	got := zeromaps.Shrink(m)
	if !maps.Equal(m, got) {
		t.Errorf("shrunk map %v differs from original %v", got, m)
	}
	got[3] = "three"
	if _, found := m[3]; found {
		t.Error("result of Shrink shares storage with its argument")
	}
}
