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
	"slices"
	"testing"
)

func genWords(n int, dupRate float64) [][]byte {
	rng := rand.New(rand.NewSource(42))
	pool := make([][]byte, 0, n)
	unique := int(float64(n) * (1 - dupRate))
	for range unique {
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
		name, expected := "NoInterning", 0
		if withInterning {
			name, expected = "Interning", 150_000
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				a := NewArena(expected)
				for _, w := range words {
					a.addBytes(w)
				}
			}
		})
	}
}

var boolVal bool

func BenchmarkEqual(b *testing.B) {
	shortA, shortB := NewArena(0), NewArena(0)
	gsShort1 := shortA.FromBytes([]byte("short"))
	gsShort2 := shortB.FromBytes([]byte("short"))

	longNoIntern := NewArena(0)
	gsLong1 := longNoIntern.FromBytes([]byte("a_really_long_string_with_words"))
	gsLong2 := longNoIntern.FromBytes([]byte("a_really_long_string_with_words"))

	longIntern := NewArena(10)
	gsLongI1 := longIntern.FromBytes([]byte("a_really_long_string_with_words"))
	gsLongI2 := longIntern.FromBytes([]byte("a_really_long_string_with_words"))

	b.Run("Short", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = gsShort1.Equal(shortA, gsShort2)
		}
	})
	b.Run("LongNoInterning", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = gsLong1.Equal(longNoIntern, gsLong2)
		}
	})
	b.Run("LongInterning", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = gsLongI1.Equal(longIntern, gsLongI2)
		}
	})
}

func BenchmarkCacheCompare(b *testing.B) {
	a := NewArena(0)
	us1 := a.FromBytes([]byte("a_really_long_string_with_words"))
	us2 := a.FromBytes([]byte("a_really_longother_string_with_words"))

	b.Run("BytesEqual", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = bytes.Equal(us1.cache(), us2.cache())
		}
	})
	b.Run("ArrayEqual", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = us1.cacheEqual(us2)
		}
	})
}

func BenchmarkCompare(b *testing.B) {
	a := NewArena(0)
	content := []byte("donaudampfschifffahrtsgesellschaft")
	us := a.FromBytes(content)

	b.Run("HasPrefix_Fast", func(b *testing.B) {
		needle := []byte("dona")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.HasPrefixBytes(a, needle)
		}
	})
	b.Run("HasPrefix_Arena", func(b *testing.B) {
		needle := []byte("donaudampfschifffahr")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.HasPrefixBytes(a, needle)
		}
	})
	b.Run("ContainsFast", func(b *testing.B) {
		needle := []byte("donau")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.ContainsBytes(a, needle)
		}
	})
	b.Run("ContainsMiddle", func(b *testing.B) {
		needle := []byte("schifffahrt")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.ContainsBytes(a, needle)
		}
	})
	b.Run("ContainsAtEnd", func(b *testing.B) {
		needle := []byte("gesellschaft")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.ContainsBytes(a, needle)
		}
	})
	b.Run("HasSuffix", func(b *testing.B) {
		needle := []byte("gesellschaft")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.HasSuffixBytes(a, needle)
		}
	})
	b.Run("EqualBytes", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = us.EqualBytes(a, content)
		}
	})
	b.Run("Equal", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = us.Equal(a, us)
		}
	})

	a = NewArena(16)
	us = a.FromBytes(content)
	b.Run("EqualIntern", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = us.Equal(a, us)
		}
	})
}

func BenchmarkShort(b *testing.B) {
	a := NewArena(0)
	content := []byte("0123456789ABCD")
	if len(content) != payloadLen {
		panic(string(content))
	}
	us := a.FromBytes(content)

	b.Run("HasPrefix", func(b *testing.B) {
		needle := []byte("0123")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.HasPrefixBytes(a, needle)
		}
	})
	b.Run("ContainsPrefix", func(b *testing.B) {
		needle := []byte("0123")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.ContainsBytes(a, needle)
		}
	})
	b.Run("ContainsMiddle", func(b *testing.B) {
		needle := []byte("5678")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.ContainsBytes(a, needle)
		}
	})
	b.Run("ContainsEnd", func(b *testing.B) {
		needle := []byte("ABCD")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.ContainsBytes(a, needle)
		}
	})
	b.Run("HasSuffix", func(b *testing.B) {
		needle := []byte("ABCD")
		b.ReportAllocs()
		for range b.N {
			boolVal = us.HasSuffixBytes(a, needle)
		}
	})
	b.Run("EqualBytes", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = us.EqualBytes(a, content)
		}
	})
	b.Run("EqualUmbra", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			boolVal = us.Equal(a, us)
		}
	})

	// The following benchmarks are to compare with the functions from
	// Go standard library.

	b.Run("BytesEqual", func(b *testing.B) {
		other := slices.Clone(content)
		b.ReportAllocs()
		for range b.N {
			boolVal = bytes.Equal(content, other)
		}
	})
	b.Run("BytesHasPrefix", func(b *testing.B) {
		needle := []byte("0123")
		b.ReportAllocs()
		for range b.N {
			boolVal = bytes.HasPrefix(content, needle)
		}
	})
	b.Run("BytesContainsEnd", func(b *testing.B) {
		needle := []byte("ABCD")
		b.ReportAllocs()
		for range b.N {
			boolVal = bytes.Contains(content, needle)
		}
	})
	b.Run("BytesHasSuffix", func(b *testing.B) {
		needle := []byte("ABCD")
		b.ReportAllocs()
		for range b.N {
			boolVal = bytes.HasSuffix(content, needle)
		}
	})
}
