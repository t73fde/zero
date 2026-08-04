//-----------------------------------------------------------------------------
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
//-----------------------------------------------------------------------------

package bitset_test

import (
	"testing"

	"t73f.de/r/zero/bitset"
)

var benchBool bool
var benchInt int
var benchBitSet bitset.BitSet

func BenchmarkBitSetHierarchy(b *testing.B) {
	base := bitset.New(uint('a'), 'b', 'c')
	b.ResetTimer()
	for b.Loop() {
		level1 := base.Clone()
		level1.Insert(uint('{'))

		level2 := level1.Clone()
		level2.Insert(uint('e'))
		level2.Insert(uint('}'))

		if level2.Contains('z') {
			benchInt++
		}
		if level2.Contains('c') {
			benchInt++
		}
	}
}

func BenchmarkBitSetTextScan(b *testing.B) {
	base := bitset.New(uint('a'), 'e', 'i', 'o', 'u')

	text := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

	b.ResetTimer()
	for b.Loop() {
		bs := base.Clone()

		// Hierarchische Erweiterung simulieren.
		bs.Insert(uint('x'))
		bs.Insert(uint('y'))

		found := 0
		for _, c := range text {
			if bs.Contains(uint(c)) {
				found++
			}
		}
		if found == 0 {
			b.Fatal("expected match")
		}
	}
}

func BenchmarkBitSetTextScanOnly(b *testing.B) {
	bs := bitset.New(uint('a'), 'e', 'i', 'o', 'u')

	text := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	b.ResetTimer()
	for b.Loop() {
		found := 0
		for _, c := range text {
			if bs.Contains(uint(c)) {
				found++
			}
		}
		benchInt = found
	}
}

func BenchmarkBitSetContainsASCII(b *testing.B) {
	bs := bitset.New(uint('a'), 'e', 'i', 'o', 'u')

	text := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	b.ResetTimer()
	for b.Loop() {
		for _, c := range text {
			benchBool = bs.Contains(uint(c))
		}
	}
}

func BenchmarkArrayContainsASCII(b *testing.B) {
	var table [256]bool
	for _, c := range "aeiou" {
		table[c] = true
	}

	text := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	b.ResetTimer()
	for b.Loop() {
		for _, c := range text {
			benchBool = table[c]
		}
	}
}

func BenchmarkBitSetCloneAndExtend(b *testing.B) {
	base := bitset.New(uint('a'), 'b', 'c')
	for b.Loop() {
		bs := base.Clone()
		for c := byte('d'); c <= 'z'; c++ {
			bs.Insert(uint(c))
		}
	}
}

func BenchmarkBitSetSparseClone(b *testing.B) {
	base := bitset.New(uint(1_000_000))

	b.ResetTimer()
	for b.Loop() {
		benchBitSet = base.Clone()
	}
}

func BenchmarkBitSetClip(b *testing.B) {
	for b.Loop() {
		var bs bitset.BitSet
		bs.EnsureBit(1_000_000)
		bs.Insert(1)
		bs.Clip()
	}
}

func BenchmarkBitSetASCIIClone(b *testing.B) {
	var base bitset.BitSet
	for i := uint(32); i < 128; i++ {
		base.Insert(i)
	}

	b.ResetTimer()
	for b.Loop() {
		benchBitSet = base.Clone()
	}
}

func BenchmarkBitSetCloneLevels(b *testing.B) {
	base := bitset.New(uint('a'), 'b', 'c')

	b.ResetTimer()
	for b.Loop() {
		bs := base.Clone()
		for i := range 10 {
			bs.Insert(uint('a' + i))
			bs = bs.Clone()
		}
		benchBitSet = bs
	}
}
