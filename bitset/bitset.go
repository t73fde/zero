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

// Package bitset implements a compact set of non-negative integers.
//
// BitSet is optimized for dense value ranges and fast membership tests.
// Typical use cases include character classes, encoder escape tables,
// parser states, and bitmap containers.
package bitset

import (
	"iter"
	"math/bits"
	"slices"
	"strconv"
)

// Value is a type that can be stored in a BitSet.
type Value interface {
	~uint | ~uint8 | ~uint16 | ~uint32
}

// BitSet is a set of non-negative integer values.
// Every integer maps to a single bit.
type BitSet[V Value] struct {
	words []word
}

type word = uint

const (
	wordSizeBits = bits.UintSize
	wordMask     = wordSizeBits - 1
)

// ----- Constructors

// New returns a BitSet containing all given values.
func New[V Value](values ...V) BitSet[V] {
	var bs BitSet[V]
	if len(values) == 0 {
		return bs
	}
	bs.EnsureBit(slices.Max(values))
	for _, n := range values {
		index := uint(n) / wordSizeBits
		bs.words[index] |= word(1) << (uint(n) & wordMask)
	}
	return bs
}

// Collect returns a BitSet containing all values produced by seq.
func Collect[V Value](seq iter.Seq[V]) BitSet[V] {
	var bs BitSet[V]
	for n := range seq {
		bs.Insert(n)
	}
	return bs
}

// ----- Basic set operations

// Insert a non-negative integer to the set.
func (bs *BitSet[V]) Insert(n V) {
	bs.words[bs.ensureWord(n)] |= word(1) << (n & wordMask)
}

// Delete a non-negative integer from the set.
func (bs *BitSet[V]) Delete(n V) {
	index := n / wordSizeBits
	if len(bs.words) <= int(index) {
		return
	}
	bs.words[index] &^= 1 << (n & wordMask)
}

// DeleteAll removes all values from the set while retaining the allocated storage.
func (bs *BitSet[V]) DeleteAll() {
	clear(bs.words)
}

// ----- Queries

// Contains reports whether a non-negative integer is in the set.
func (bs BitSet[V]) Contains(n V) bool {
	index := n / wordSizeBits
	if len(bs.words) <= int(index) {
		return false
	}
	return bs.words[index]&(1<<(n&wordMask)) != 0
}

// Count returns the number of values in the set.
func (bs BitSet[V]) Count() int {
	count := 0
	for _, w := range bs.words {
		count += bits.OnesCount(w)
	}
	return count
}

// IsEmpty reports whether the set contains no values.
func (bs BitSet[V]) IsEmpty() bool {
	for _, w := range bs.words {
		if w != 0 {
			return false
		}
	}
	return true
}

// Min returns the smallest value in the BitSet.
// It reports false if the BitSet is empty.
func (bs BitSet[V]) Min() (V, bool) {
	for i, w := range bs.words {
		if w != 0 {
			return V(i)*wordSizeBits + V(bits.TrailingZeros(uint(w))), true
		}
	}
	return 0, false
}

// Max returns the largest value in the BitSet.
// It reports false if the BitSet is empty.
func (bs BitSet[V]) Max() (V, bool) {
	for i := len(bs.words) - 1; i >= 0; i-- {
		if w := bs.words[i]; w != 0 {
			return V(i)*wordSizeBits +
				(wordSizeBits - 1 - V(bits.LeadingZeros(uint(w)))), true
		}
	}
	return 0, false
}

// Equal reports whether bs and other contain the same values.
func (bs BitSet[V]) Equal(other BitSet[V]) bool {
	i := len(bs.words) - 1
	j := len(other.words) - 1

	for i >= 0 && bs.words[i] == 0 {
		i--
	}
	for j >= 0 && other.words[j] == 0 {
		j--
	}

	if i != j {
		return false
	}

	for i >= 0 {
		if bs.words[i] != other.words[i] {
			return false
		}
		i--
	}
	return true
}

// ----- Iteration / conversion

// All returns an iterator over all values in the set in ascending order.
//
// The iterator does not modify the BitSet.
func (bs BitSet[V]) All() iter.Seq[V] {
	return func(yield func(V) bool) {
		base := V(0)
		for _, w := range bs.words {
			for w != 0 {
				pos := V(bits.TrailingZeros(w))
				if !yield(base + pos) {
					return
				}
				w &= w - 1 // clear lowest set bit
			}
			base += wordSizeBits
		}
	}
}

// String returns the set in ascending order as "{1,2,7}".
func (bs BitSet[V]) String() string {
	buf := make([]byte, 0, 32)
	buf = append(buf, '{')

	first := true
	for num := range bs.All() {
		if first {
			first = false
		} else {
			buf = append(buf, ',')
		}
		buf = strconv.AppendUint(buf, uint64(num), 10)
	}
	buf = append(buf, '}')
	return string(buf)
}

// ----- Set operations (non-mutating)

// Union returns the union of bs and other.
func (bs BitSet[V]) Union(other BitSet[V]) BitSet[V] {
	result := bs.Clone()
	result.Or(other)
	return result
}

// Intersection returns the intersection of bs and other.
func (bs BitSet[V]) Intersection(other BitSet[V]) BitSet[V] {
	result := bs.Clone()
	result.And(other)
	return result
}

// Difference returns the difference of bs and other.
func (bs BitSet[V]) Difference(other BitSet[V]) BitSet[V] {
	result := bs.Clone()
	result.AndNot(other)
	return result
}

// ----- Set operations (mutating)

// Or sets bs to the union of bs and other (bs |= other).
func (bs *BitSet[V]) Or(other BitSet[V]) {
	bs.growWords(len(other.words))
	for i, w := range other.words {
		bs.words[i] |= w
	}
}

// And sets bs to the intersection of bs and other (bs &= other).
func (bs *BitSet[V]) And(other BitSet[V]) {
	n := min(len(bs.words), len(other.words))
	for i := range n {
		bs.words[i] &= other.words[i]
	}
	for i := n; i < len(bs.words); i++ {
		bs.words[i] = 0
	}
}

// AndNot removes all values of other from bs (bs &^= other).
func (bs *BitSet[V]) AndNot(other BitSet[V]) {
	n := min(len(bs.words), len(other.words))
	for i := range n {
		bs.words[i] &^= other.words[i]
	}
}

// Xor sets bs to the symmetric difference of bs and other (bs ^= other).
func (bs *BitSet[V]) Xor(other BitSet[V]) {
	bs.growWords(len(other.words))
	for i, w := range other.words {
		bs.words[i] ^= w
	}
}

// ----- Memory management

// Clone returns a copy of the bitset.
func (bs BitSet[V]) Clone() BitSet[V] {
	words := make([]word, len(bs.words))
	copy(words, bs.words)
	return BitSet[V]{words: words}
}

// EnsureBit ensures that n can be inserted without further allocation.
// It does not insert n.
func (bs *BitSet[V]) EnsureBit(n V) {
	_ = bs.ensureWord(n)
}

// Clip reduces the BitSet storage to the minimum size needed for its values.
func (bs *BitSet[V]) Clip() {
	i := len(bs.words)
	for i > 0 && bs.words[i-1] == 0 {
		i--
	}
	if i == 0 {
		bs.words = nil
	} else {
		bs.words = bs.words[:i:i]
	}
}

// ----- internal helpers

func (bs *BitSet[V]) ensureWord(n V) int {
	index := int(n / wordSizeBits)
	bs.growWords(index + 1)
	return index
}

func (bs *BitSet[V]) growWords(n int) {
	if n <= len(bs.words) {
		return
	}
	if n <= cap(bs.words) {
		bs.words = bs.words[:n]
		return
	}
	newWords := make([]word, n, max(n, 2*cap(bs.words)))
	copy(newWords, bs.words)
	bs.words = newWords
}
