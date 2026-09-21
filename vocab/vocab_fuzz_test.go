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

package vocab

import (
	"bytes"
	"testing"
)

func FuzzAgainstModel(f *testing.F) {
	f.Add([]byte("House\xffHouses\xffHouse\xff\xffInternationalization"))
	f.Fuzz(func(t *testing.T, data []byte) {
		v, m := New(0), newModel()
		for w := range bytes.SplitSeq(data, []byte{0xff}) {
			if got, want := v.Add(w), m.add(string(w)); got != want {
				t.Fatalf("Add(%q) = %d, want %d", w, got, want)
			}
		}
		verify(t, v, m)
		for _, w := range m.words[:min(len(m.words), 20)] {
			checkSearches(t, v, m, []byte(w))
			if len(w) > 1 {
				checkSearches(t, v, m, []byte(w)[1:])
			}
		}
	})
}
