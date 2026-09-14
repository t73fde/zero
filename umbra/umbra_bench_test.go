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
	"math/rand"
	"testing"
)

func genWords(n int, dupRate float64) [][]byte {
	rng := rand.New(rand.NewSource(42))
	pool := make([][]byte, 0, n)
	unique := int(float64(n) * (1 - dupRate))
	for i := 0; i < unique; i++ {
		length := 3 + rng.Intn(30)
		w := make([]byte, length)
		for j := range w {
			w[j] = byte('a' + rng.Intn(26))
		}
		pool = append(pool, w)
	}
	words := make([][]byte, n)
	for i := range words {
		words[i] = pool[rng.Intn(len(pool))]
	}
	return words
}

func BenchmarkAddBytes(b *testing.B) {
	words := genWords(200_000, 0.40)
	for _, withInterning := range []bool{false, true} {
		name := "NoInterning"
		if withInterning {
			name = "Interning"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				expected := 0
				if withInterning {
					expected = 150_000
				}
				a := NewArena(expected)
				for _, w := range words {
					a.addBytes(w)
				}
			}
		})
	}
}

func BenchmarkEqual(b *testing.B) {
	shortA, shortB := NewArena(0), NewArena(0)
	gsShort1 := shortA.FromBytes([]byte("short"))
	gsShort2 := shortB.FromBytes([]byte("short"))

	longNoIntern := NewArena(0)
	gsLong1 := longNoIntern.FromBytes([]byte("ein ziemlich langes wortfragment"))
	gsLong2 := longNoIntern.FromBytes([]byte("ein ziemlich langes wortfragment"))

	longIntern := NewArena(10)
	gsLongI1 := longIntern.FromBytes([]byte("ein ziemlich langes wortfragment"))
	gsLongI2 := longIntern.FromBytes([]byte("ein ziemlich langes wortfragment"))

	b.Run("Short", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = gsShort1.Equal(shortA, gsShort2)
		}
	})
	b.Run("LongNoInterning", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = gsLong1.Equal(longNoIntern, gsLong2)
		}
	})
	b.Run("LongInterning", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = gsLongI1.Equal(longIntern, gsLongI2)
		}
	})
}

func BenchmarkCacheCompare(b *testing.B) {
	a := NewArena(0)
	us1 := a.FromBytes([]byte("ein ziemlich langes wortfragment"))
	us2 := a.FromBytes([]byte("ein ziemlich anderes wortfragment"))

	b.Run("BytesEqual", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = bytes.Equal(us1.cache(), us2.cache())
		}
	})
	b.Run("Uint64Word", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = us1.cacheEqual(us2)
		}
	})
}

func BenchmarkHasPrefix(b *testing.B) {
	a := NewArena(0)
	g := a.FromBytes([]byte("donaudampfschifffahrtsgesellschaft"))

	b.Run("FastPath_4Byte", func(b *testing.B) {
		needle := []byte("dona")
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = g.HasPrefixBytes(a, needle)
		}
	})
	b.Run("ArenaFallback_20Byte", func(b *testing.B) {
		needle := []byte("donaudampfschifffahr")
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = g.HasPrefixBytes(a, needle)
		}
	})
}

func BenchmarkContainsHasSuffix(b *testing.B) {
	a := NewArena(0)
	g := a.FromBytes([]byte("donaudampfschifffahrtsgesellschaft"))
	needle := []byte("gesellschaft")

	b.Run("Contains", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = g.ContainsBytes(a, needle)
		}
	})
	b.Run("HasSuffix", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = g.HasSuffixBytes(a, needle)
		}
	})
}
