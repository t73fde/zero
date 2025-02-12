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

// CatSeq returns an iterator that is the concatenation of all given iterators.
func CatSeq[V any](seqs ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, seq := range seqs {
			seq(yield)
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

// ReduceSeq reduces an iterator by applying its elements to an operator.
func ReduceSeq[V, W any](seq iter.Seq[V], init W, op func(W, V) W) W {
	cur := init
	for elem := range seq {
		cur = op(cur, elem)
	}
	return cur
}

// CountSeq returns an iterator that counts, starting with 0.
func CountSeq() iter.Seq[int] {
	const maxMinusOne = math.MaxInt - 1
	return func(yield func(int) bool) {
		for i := 0; i < maxMinusOne; i++ {
			if !yield(i) {
				return
			}
		}
		yield(math.MaxInt)
	}
}

// TakeSeq returns an iterator that only has a maximum number of elements.
func TakeSeq[V any](num int, seq iter.Seq[V]) iter.Seq[V] {
	if num <= 0 {
		return func(func(V) bool) {}
	}
	return func(yield func(V) bool) {
		cur := 0
		for elem := range seq {
			if cur >= num || !yield(elem) {
				return
			}
			cur++
		}
	}
}

// ZipSeq returns an iterator that is an K/V iterator of the given two iterators.
// I.e. is produces pairs.
func ZipSeq[K, V any](kseq iter.Seq[K], vseq iter.Seq[V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		knext, kdone := iter.Pull(kseq)
		vnext, vdone := iter.Pull(vseq)
		defer kdone()
		defer vdone()
		for {
			k, ok := knext()
			if !ok {
				return
			}
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
