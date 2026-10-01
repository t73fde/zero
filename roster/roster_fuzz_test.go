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

package roster_test

import (
	"testing"

	"t73f.de/r/zero/roster"
)

func FuzzPredicates(f *testing.F) {
	f.Add([]byte{1, 2, 3}, []byte{2, 3, 4})
	f.Add([]byte{}, []byte{1})
	f.Add([]byte{7}, []byte{})
	f.Fuzz(func(t *testing.T, ab, bb []byte) {
		av, bv := toUints(ab), toUints(bb)
		a, b := roster.New(av...), roster.New(bv...)
		ma, mb := refSet(av), refSet(bv)

		if got, want := a.IsSubset(b), refSubset(ma, mb); got != want {
			t.Fatalf("IsSubset(%v, %v) = %v, want %v", a, b, got, want)
		}
		if got, want := a.Intersects(b), refIntersects(ma, mb); got != want {
			t.Fatalf("Intersects(%v, %v) = %v, want %v", a, b, got, want)
		}
	})
}

func toUints(bs []byte) []uint {
	vals := make([]uint, len(bs))
	for i, b := range bs {
		vals[i] = uint(b)
	}
	return vals
}
