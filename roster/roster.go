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

// Package roster implements a set of non-negative integers as a sorted array.
//
// A Roster stores its values in strictly ascending order, without duplicates.
// Membership tests take O(log n), and set operations such as Delta take
// O(n+m) by merging the sorted arrays of both operands in a single pass.
//
// Roster is the counterpart to package bitset: a BitSet needs memory
// proportional to its largest value and is best for dense value ranges,
// while a Roster needs memory proportional to the number of values and is
// best for sparse sets or sets with large values.
//
// The zero value of Roster is an empty set and ready to use. Set operations
// do not modify their operands; results never share storage with them.
//
// Copying a Roster: Roster is a value type that holds a slice. Assigning a
// Roster or passing it by value copies only the slice header, so both copies
// share the same storage. Modifying one of them (Insert, Delete, DeleteAll,
// Pop, And, ...) is then not allowed to be relied upon for the other copy.
// An independent copy is obtained only by calling Clone.
package roster

import (
	"io"
	"iter"
	"slices"
	"strconv"
	"strings"
)

// Value is a type that can be stored in a Roster.
type Value interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Roster is a set of non-negative integer values, implemented as a sorted
// array, without duplicates.
//
// A Roster must not be copied by assignment if either copy is modified
// afterwards; use Clone to obtain an independent copy.
type Roster[V Value] struct {
	array []V
}

// ----- Constructors

// New returns a Roster containing all given values.
// Unsorted input and duplicate values are allowed.
func New[V Value](values ...V) Roster[V] {
	if len(values) == 0 {
		return Roster[V]{}
	}
	array := slices.Clone(values) // do not sort the caller's slice in place
	slices.Sort(array)
	return Roster[V]{array: slices.Compact(array)}
}

// Collect returns a Roster containing all values produced by seq.
func Collect[V Value](seq iter.Seq[V]) Roster[V] {
	array := slices.Collect(seq)
	if len(array) == 0 {
		return Roster[V]{}
	}
	slices.Sort(array)
	return Roster[V]{array: slices.Compact(array)}
}

// ----- Basic set operations

// Insert adds a non-negative integer to the set.
//
// Insert takes O(n), because subsequent values are shifted. To build a
// Roster from many values, use New or Collect instead of repeated calls.
func (r *Roster[V]) Insert(n V) {
	if i, found := slices.BinarySearch(r.array, n); !found {
		r.array = slices.Insert(r.array, i, n)
	}
}

// InsertSeq adds all values produced by seq to the set.
// Unsorted input and duplicate values are allowed.
//
// InsertSeq takes O(k log k + n + k) for k produced values, in contrast to
// O(k*n) for k calls of Insert.
func (r *Roster[V]) InsertSeq(seq iter.Seq[V]) {
	r.Or(Collect(seq))
}

// Delete removes a non-negative integer from the set.
//
// Delete takes O(n), because subsequent values are shifted.
func (r *Roster[V]) Delete(n V) {
	if i, found := slices.BinarySearch(r.array, n); found {
		r.array = slices.Delete(r.array, i, i+1)
	}
}

// DeleteAll removes all values from the set while retaining the allocated storage.
func (r *Roster[V]) DeleteAll() {
	r.array = r.array[:0]
}

// Pop removes and returns some containing value in the roster.
// It returns false if the roster is empty.
func (r *Roster[V]) Pop() (V, bool) {
	n := len(r.array)
	if n == 0 {
		var zero V
		return zero, false
	}

	n--
	val := r.array[n]
	r.array = r.array[:n]
	return val, true
}

// ----- Queries

// Contains reports whether a non-negative integer is in the set.
func (r Roster[V]) Contains(n V) bool {
	_, found := slices.BinarySearch(r.array, n)
	return found
}

// Count returns the number of values in the set.
func (r Roster[V]) Count() int {
	return len(r.array)
}

// IsEmpty reports whether the set contains no values.
func (r Roster[V]) IsEmpty() bool {
	return len(r.array) == 0
}

// Min returns the smallest value in the Roster.
// It reports false if the Roster is empty.
func (r Roster[V]) Min() (V, bool) {
	if len(r.array) > 0 {
		return r.array[0], true
	}
	return 0, false
}

// Max returns the largest value in the Roster.
// It reports false if the Roster is empty.
func (r Roster[V]) Max() (V, bool) {
	if l := len(r.array); l > 0 {
		return r.array[l-1], true
	}
	return 0, false
}

// Equal reports whether r and other contain the same values.
func (r Roster[V]) Equal(other Roster[V]) bool {
	return slices.Equal(r.array, other.array)
}

// IsSubset reports whether every value of r is also in other.
// The empty set is a subset of every set.
//
// IsSubset does not allocate memory and stops at the first value of r that is
// not in other.
func (r Roster[V]) IsSubset(other Roster[V]) bool {
	a, b := r.array, other.array
	if len(a) > len(b) {
		return false
	}
	j := 0
	for _, v := range a {
		for j < len(b) && b[j] < v {
			j++
		}
		if j == len(b) || b[j] != v {
			return false
		}
		j++
	}
	return true
}

// Intersects reports whether r and other have at least one value in common.
//
// Intersects does not allocate memory and stops at the first common value.
func (r Roster[V]) Intersects(other Roster[V]) bool {
	a, b := r.array, other.array
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	if a[len(a)-1] < b[0] || b[len(b)-1] < a[0] {
		return false
	}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			i++
		case a[i] > b[j]:
			j++
		default:
			return true
		}
	}
	return false
}

// ----- Iteration / conversion

// Values returns an iterator over all values in the set in ascending order.
//
// The iterator does not modify the Roster. The Roster must not be modified
// while the iteration is in progress.
func (r Roster[V]) Values() iter.Seq[V] {
	return slices.Values(r.array)
}

// String returns the set in ascending order as "1 2 7".
func (r Roster[V]) String() string {
	var b strings.Builder
	_, _ = r.WriteTo(&b)
	return b.String()
}

var _ io.WriterTo = Roster[uint]{}

// WriteTo writes the roster's values to w, separated by spaces.
// It returns the number of bytes written and any error encountered.
func (r Roster[V]) WriteTo(w io.Writer) (n int64, err error) {
	var buf [21]byte // 20 bytes for digits, one for separator
	for i, val := range r.array {
		b := buf[:0]
		if i > 0 {
			b = append(b, ' ')
		}
		b = strconv.AppendUint(b, uint64(val), 10)
		m, e := w.Write(b)
		n += int64(m)
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

// ----- Set operations (non-mutating)

// Union returns the union of r and other.
func (r Roster[V]) Union(other Roster[V]) Roster[V] {
	return Roster[V]{array: union(r.array, other.array)}
}

// Intersection returns the intersection of r and other.
func (r Roster[V]) Intersection(other Roster[V]) Roster[V] {
	if len(other.array) < len(r.array) {
		result := other.Clone()
		result.And(r)
		return result
	}
	result := r.Clone()
	result.And(other)
	return result
}

// Difference returns the values of r that are not in other.
func (r Roster[V]) Difference(other Roster[V]) Roster[V] {
	result := r.Clone()
	result.AndNot(other)
	return result
}

// SymmetricDifference returns the values that are in exactly one of r and other.
func (r Roster[V]) SymmetricDifference(other Roster[V]) Roster[V] {
	a, b := r.array, other.array
	var result []V
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			result = append(result, a[i])
			i++
		case a[i] > b[j]:
			result = append(result, b[j])
			j++
		default: // in both sets: in neither part of the result
			i++
			j++
		}
	}
	result = append(result, a[i:]...)
	result = append(result, b[j:]...)
	return Roster[V]{array: result}
}

// Delta returns the values that are only in r and the values that are only
// in other. Neither r nor other is modified; the results do not share
// storage with them.
//
// Example: removed, added := old.Delta(new)
func (r Roster[V]) Delta(other Roster[V]) (onlyR, onlyOther Roster[V]) {
	a, b := r.array, other.array
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			onlyR.array = append(onlyR.array, a[i])
			i++
		case a[i] > b[j]:
			onlyOther.array = append(onlyOther.array, b[j])
			j++
		default:
			i++
			j++
		}
	}
	// At most one of the operands has a remaining tail.
	onlyR.array = append(onlyR.array, a[i:]...)
	onlyOther.array = append(onlyOther.array, b[j:]...)
	return onlyR, onlyOther
}

// ----- Set operations (mutating)

// Or sets r to the union of r and other (r |= other).
//
// Or allocates a new array of size O(len(r)+len(other)); the storage of r
// is not reused. To combine many sets, prefer New or Collect over repeated
// calls of Or.
func (r *Roster[V]) Or(other Roster[V]) {
	if len(other.array) == 0 {
		return
	}
	r.array = union(r.array, other.array)
}

// And sets r to the intersection of r and other (r &= other).
//
// And works in place and does not allocate.
func (r *Roster[V]) And(other Roster[V]) {
	a, b := r.array, other.array
	w, i, j := 0, 0, 0 // w: next write position in a, always w <= i
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			i++
		case a[i] > b[j]:
			j++
		default:
			a[w] = a[i]
			w++
			i++
			j++
		}
	}
	r.array = a[:w]
}

// AndNot removes all values of other from r (r &^= other).
//
// AndNot works in place and does not allocate.
func (r *Roster[V]) AndNot(other Roster[V]) {
	a, b := r.array, other.array
	w, j := 0, 0 // w: next write position in a, always w <= i
	for i := range a {
		for j < len(b) && b[j] < a[i] {
			j++
		}
		if j < len(b) && b[j] == a[i] {
			continue // a[i] is in other: drop it
		}
		a[w] = a[i]
		w++
	}
	r.array = a[:w]
}

// Xor sets r to the symmetric difference of r and other (r ^= other).
//
// Xor allocates a new array; the storage of r is not reused.
func (r *Roster[V]) Xor(other Roster[V]) {
	if len(other.array) == 0 {
		return
	}
	*r = r.SymmetricDifference(other)
}

// union returns the sorted union of the sorted, duplicate-free arrays a and b
// as a new array.
func union[V Value](a, b []V) []V {
	if len(a)+len(b) == 0 {
		return nil
	}
	result := make([]V, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			result = append(result, a[i])
			i++
		case a[i] > b[j]:
			result = append(result, b[j])
			j++
		default:
			result = append(result, a[i])
			i++
			j++
		}
	}
	result = append(result, a[i:]...)
	return append(result, b[j:]...)
}

// ----- Memory management

// Clone returns a copy of the Roster.
func (r Roster[V]) Clone() Roster[V] {
	return Roster[V]{array: slices.Clone(r.array)}
}

// Grow ensures that n values can be inserted without further allocation.
// It does not insert n.
//
// Grow panics if n is negative or too large to allocate the memory
func (r *Roster[V]) Grow(n int) {
	r.array = slices.Grow(r.array, n)
}

// Clip reduces the Roster storage to the minimum size needed for its values.
func (r *Roster[V]) Clip() {
	if len(r.array) == 0 {
		r.array = nil
	} else {
		r.array = slices.Clip(r.array)
	}
}
