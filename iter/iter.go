//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

// Package iter supports working with iterators.
package iter

import (
	"iter"
	"math"
)

// EmptySeq returns an empty iterator.
func EmptySeq[V any]() iter.Seq[V] { return func(func(V) bool) {} }

// OneSeq returns an iterator that returns just one given element.
func OneSeq[V any](elem V) iter.Seq[V] {
	return func(yield func(V) bool) {
		yield(elem)
	}

}

// CatSeq returns an iterator that is the concatenation of all given iterators.
func CatSeq[V any](seqs ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, seq := range seqs {
			for elem := range seq {
				if !yield(elem) {
					return
				}
			}
		}
	}
}

// MapSeq applies a function to each element of an iterator, producing an
// iterator of mapped elements.
func MapSeq[V, W any](seq iter.Seq[V], fn func(V) W) iter.Seq[W] {
	return func(yield func(W) bool) {
		for elem := range seq {
			if !yield(fn(elem)) {
				return
			}
		}
	}
}

// FilterSeq produces an iterator of all elements of the originating
// iterator that satisfy a predicate.
func FilterSeq[V any](seq iter.Seq[V], pred func(V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		for elem := range seq {
			if pred(elem) && !yield(elem) {
				return
			}
		}
	}
}

// MapFilterSeq applies a function for each element of an iterator. If the
// function returns a false value, it is ignored, otherwise the mapped value
// is added to the resulting iterator.
func MapFilterSeq[V, W any](seq iter.Seq[V], predfn func(V) (W, bool)) iter.Seq[W] {
	return func(yield func(W) bool) {
		for elem := range seq {
			if e, ok := predfn(elem); ok && !yield(e) {
				return
			}
		}
	}
}

// ReduceSeq reduces an iterator by applying its elements to an operator.
func ReduceSeq[V, W any](seq iter.Seq[V], init W, op func(W, V) W) W {
	cur := init
	for elem := range seq {
		cur = op(cur, elem)
	}
	return cur
}

// DeduplicateSeq returns an iterator with all duplicate values from
// the original iterator removed.
func DeduplicateSeq[V comparable](seq iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		seen := make(map[V]struct{})
		for elem := range seq {
			if _, found := seen[elem]; !found {
				seen[elem] = struct{}{}
				if !yield(elem) {
					return
				}
			}
		}
	}
}

// CountSeq returns an iterator that counts, starting with 0, up to and
// including math.MaxInt.
func CountSeq() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; ; i++ {
			if !yield(i) || i == math.MaxInt {
				return
			}
		}
	}
}

// TakeSeq returns an iterator that has at most num elements. It does not
// request more elements from seq than needed.
func TakeSeq[V any](num int, seq iter.Seq[V]) iter.Seq[V] {
	if num <= 0 {
		return EmptySeq[V]()
	}
	return func(yield func(V) bool) {
		cur := 0
		for elem := range seq {
			if !yield(elem) {
				return
			}
			cur++
			if cur >= num {
				return
			}
		}
	}
}

// ZipSeq returns an iterator of pairs, built from the elements of the two
// given iterators. It ends with the shorter of both.
func ZipSeq[K, V any](kseq iter.Seq[K], vseq iter.Seq[V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		vnext, vstop := iter.Pull(vseq)
		defer vstop()
		for k := range kseq {
			v, ok := vnext()
			if !ok || !yield(k, v) {
				return
			}
		}
	}
}

// KeySeq produces an iterator only of the first / key value of the given iterator.
func KeySeq[K, V any](seq iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range seq {
			if !yield(k) {
				return
			}
		}
	}
}

// ValSeq produces an iterator only of the second / val value of the given iterator.
func ValSeq[K, V any](seq iter.Seq2[K, V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range seq {
			if !yield(v) {
				return
			}
		}
	}
}

// EnumerateSeq retruns an iterator that adds a number to each value of the
// given iterator.
func EnumerateSeq[V any](seq iter.Seq[V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0
		for v := range seq {
			if !yield(i, v) || i == math.MaxInt {
				return
			}
			i++
		}
	}
}
