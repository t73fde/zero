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

package umbra

import (
	"bytes"
	"testing"
)

func TestInterningDedup(t *testing.T) {
	a := NewArena(2, true) // small enough to trigger resize of hash table
	seen := map[string]uint32{}
	words := []string{
		"donaudampfschifffahrtsgesellschaft",
		"dampfschifffahrtskapitaensmuetze",
		"donaudampfschifffahrtsgesellschaft",
		"kapitaensmuetzenvorschrift",
		"dampfschifffahrtskapitaensmuetze",
	}
	for _, w := range words {
		off := a.addBytes([]byte(w))
		if prev, ok := seen[w]; ok && prev != off {
			t.Fatalf("same content %q got different offsets: %d vs %d", w, prev, off)
		}
		seen[w] = off
		if !bytes.Equal(a.rawBytes(off, uint16(len(w))), []byte(w)) {
			t.Fatalf("Offset %d does not point to %q", off, w)
		}
	}
}
