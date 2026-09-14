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
	"math"
	"testing"
)

func FuzzString(f *testing.F) {
	f.Add([]byte("short"), []byte("ku"))
	f.Add([]byte("a_really_long_word_more_than_14_bytes"), []byte("one"))
	f.Fuzz(func(t *testing.T, b, needle []byte) {
		if len(b) > math.MaxUint16 {
			t.Skip()
		}
		a := NewArena(0)
		us := a.FromBytes(b)

		var buf []byte
		if got := us.Append(buf, a); !bytes.Equal(got, b) {
			t.Fatalf("Full() = %q, want %q", got, b)
		}
		if got, want := us.HasPrefixBytes(a, needle), bytes.HasPrefix(b, needle); got != want {
			t.Fatalf("HasPrefix(%q, %q) = %v, want %v", b, needle, got, want)
		}
		if got, want := us.HasSuffixBytes(a, needle), bytes.HasSuffix(b, needle); got != want {
			t.Fatalf("HasSuffix(%q, %q) = %v, want %v", b, needle, got, want)
		}
		if got, want := us.ContainsBytes(a, needle), bytes.Contains(b, needle); got != want {
			t.Fatalf("Contains(%q, %q) = %v, want %v", b, needle, got, want)
		}
		if !us.EqualBytes(a, b) {
			t.Fatalf("EqualBytes with itself must be true")
		}
	})
}
