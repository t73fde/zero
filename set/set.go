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

// Package set implements a set of values of a comparable type as a hash map.
//
// A Set stores its values without duplicates. Values that are equal under the
// == operator are the same element. For floating-point types this means that
// 0.0 and -0.0 are one element, but NaN is never equal to anything, including
// itself. Therefore a Set should not be used with floating-point values that
// may be NaN.
//
// Membership tests take O(1) expected time, and set operations such as Delta
// take O(n+m) expected time. The iteration order of Values and the output of
// String and WriteTo are unspecified.
//
// Set is a value type that holds a map. The zero value is an empty set that is
// ready to use. Methods that modify the set (Insert, Delete, Pop, And, ...)
// have pointer receivers; all other methods have value receivers.
//
// A modifying method may allocate the underlying map (the zero value has none).
// Therefore a Set that is stored in a map element or in a slice element must
// be modified via a variable and written back:
//
//	s := m[key]
//	s.Insert(n)
//	m[key] = s
//
// Alternatively, store *Set values.
//
// Assigning a Set or passing it by value copies only the map reference, so
// both copies share the same storage. After modifying one of them, the state
// of the other copy is unspecified. An independent copy is obtained only by
// calling Clone.
package set

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"strconv"
	"strings"
)

// Set is a set of comparable values, implemented as a hash map, without
// duplicates.
//
// A Set must not be copied by assignment if either copy is modified
// afterwards; use Clone to obtain an independent copy.
type Set[V comparable] struct {
	m map[V]struct{}
}

// ----- Constructors

// New returns a Set containing all given values.
// Duplicate values are allowed.
func New[V comparable](values ...V) Set[V] {
	if len(values) == 0 {
		return Set[V]{}
	}
	m := make(map[V]struct{}, len(values))
	for _, v := range values {
		m[v] = struct{}{}
	}
	return Set[V]{m: m}
}

// Collect returns a Set containing all values produced by seq.
func Collect[V comparable](seq iter.Seq[V]) Set[V] {
	var s Set[V]
	s.InsertSeq(seq)
	return s
}

// ----- Basic set operations

// Insert adds a value to the set.
func (s *Set[V]) Insert(v V) {
	if s.m == nil {
		s.m = make(map[V]struct{})
	}
	s.m[v] = struct{}{}
}

// InsertSeq adds all values produced by seq to the set.
// Duplicate values are allowed.
func (s *Set[V]) InsertSeq(seq iter.Seq[V]) {
	for v := range seq {
		s.Insert(v)
	}
}

// Delete removes a value from the set.
func (s *Set[V]) Delete(v V) { delete(s.m, v) }

// DeleteAll removes all values from the set while retaining the allocated storage.
func (s *Set[V]) DeleteAll() { clear(s.m) }

// Pop removes and returns some containing value in the set.
// It returns false if the set is empty.
//
// Which value is returned is unspecified.
func (s *Set[V]) Pop() (V, bool) {
	for v := range s.m {
		delete(s.m, v)
		return v, true
	}
	var zero V
	return zero, false
}

// ----- Queries

// Contains reports whether a value is in the set.
func (s Set[V]) Contains(v V) bool {
	_, found := s.m[v]
	return found
}

// Count returns the number of values in the set.
func (s Set[V]) Count() int { return len(s.m) }

// IsEmpty reports whether the set contains no values.
func (s Set[V]) IsEmpty() bool { return len(s.m) == 0 }

// Equal reports whether s and other contain the same values.
func (s Set[V]) Equal(other Set[V]) bool {
	if len(s.m) != len(other.m) {
		return false
	}
	for v := range s.m {
		if _, found := other.m[v]; !found {
			return false
		}
	}
	return true
}

// IsSubset reports whether every value of s is also in other.
// The empty set is a subset of every set.
//
// IsSubset does not allocate memory and stops at the first value of s that is
// not in other.
func (s Set[V]) IsSubset(other Set[V]) bool {
	if len(s.m) > len(other.m) {
		return false
	}
	for v := range s.m {
		if _, found := other.m[v]; !found {
			return false
		}
	}
	return true
}

// Intersects reports whether s and other have at least one value in common.
//
// Intersects does not allocate memory and stops at the first common value.
func (s Set[V]) Intersects(other Set[V]) bool {
	small, large := s.m, other.m
	if len(large) < len(small) {
		small, large = large, small
	}
	for v := range small {
		if _, found := large[v]; found {
			return true
		}
	}
	return false
}

// ----- Iteration / conversion

// Values returns an iterator over all values in the set in unspecified order.
//
// The iterator does not modify the Set. The Set must not be modified
// while the iteration is in progress.
func (s Set[V]) Values() iter.Seq[V] { return maps.Keys(s.m) }

// String returns the set in unspecified order as "1 2 7".
func (s Set[V]) String() string {
	var b strings.Builder
	_, _ = s.WriteTo(&b)
	return b.String()
}

var _ io.WriterTo = Set[uint]{}

// WriteTo writes the set's values to w in unspecified order, separated by
// spaces. It returns the number of bytes written and any error encountered.
func (s Set[V]) WriteTo(w io.Writer) (n int64, err error) {
	buf := make([]byte, 0, 32)
	first := true
	for val := range s.m {
		b := buf[:0]
		if !first {
			b = append(b, ' ')
		}
		first = false
		b = appendValue(b, val)
		buf = b // retain a possibly grown buffer
		m, e := w.Write(b)
		n += int64(m)
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

func appendValue[V comparable](buf []byte, val V) []byte {
	switch v := any(val).(type) {
	case string:
		return append(buf, v...)
	case int:
		return strconv.AppendInt(buf, int64(v), 10)
	case uint:
		return strconv.AppendUint(buf, uint64(v), 10)
	default:
		return fmt.Append(buf, val)
	}
}

// ----- Set operations (non-mutating)

// Union returns the union of s and other.
func (s Set[V]) Union(other Set[V]) Set[V] {
	large, small := s, other
	if len(small.m) > len(large.m) {
		large, small = small, large
	}
	result := large.Clone()
	result.Or(small)
	return result
}

// Intersection returns the intersection of s and other.
func (s Set[V]) Intersection(other Set[V]) Set[V] {
	small, large := s.m, other.m
	if len(large) < len(small) {
		small, large = large, small
	}
	var result Set[V]
	for v := range small {
		if _, found := large[v]; found {
			result.Insert(v)
		}
	}
	return result
}

// Difference returns the values of s that are not in other.
func (s Set[V]) Difference(other Set[V]) Set[V] {
	var result Set[V]
	for v := range s.m {
		if _, found := other.m[v]; !found {
			result.Insert(v)
		}
	}
	return result
}

// SymmetricDifference returns the values that are in exactly one of s and other.
func (s Set[V]) SymmetricDifference(other Set[V]) Set[V] {
	onlyS, onlyOther := s.Delta(other)
	onlyS.Or(onlyOther)
	return onlyS
}

// Delta returns the values that are only in s and the values that are only
// in other. Neither s nor other is modified; the results do not share
// storage with them.
//
// Example: removed, added := old.Delta(new)
func (s Set[V]) Delta(other Set[V]) (onlyS, onlyOther Set[V]) {
	for v := range s.m {
		if _, found := other.m[v]; !found {
			onlyS.Insert(v)
		}
	}
	for v := range other.m {
		if _, found := s.m[v]; !found {
			onlyOther.Insert(v)
		}
	}
	return onlyS, onlyOther
}

// ----- Set operations (mutating)

// Or sets s to the union of s and other (s |= other).
//
// Or inserts the values of other into the existing map of s; it allocates
// only if the map has to grow.
func (s *Set[V]) Or(other Set[V]) {
	for v := range other.m {
		s.Insert(v)
	}
}

// And sets s to the intersection of s and other (s &= other).
//
// And works in place and does not allocate.
func (s *Set[V]) And(other Set[V]) {
	for v := range s.m {
		if _, found := other.m[v]; !found {
			delete(s.m, v)
		}
	}
}

// AndNot removes all values of other from s (s &^= other).
//
// AndNot works in place and does not allocate.
func (s *Set[V]) AndNot(other Set[V]) {
	if len(other.m) < len(s.m) {
		for v := range other.m {
			delete(s.m, v)
		}
		return
	}
	for v := range s.m {
		if _, found := other.m[v]; found {
			delete(s.m, v)
		}
	}
}

// Xor sets s to the symmetric difference of s and other (s ^= other).
//
// Xor works in place; it allocates only if the map of s has to grow.
func (s *Set[V]) Xor(other Set[V]) {
	for v := range other.m {
		if _, found := s.m[v]; found {
			delete(s.m, v)
		} else {
			s.Insert(v)
		}
	}
}

// ----- Memory management

// Clone returns a copy of the Set.
func (s Set[V]) Clone() Set[V] {
	return Set[V]{m: maps.Clone(s.m)}
}

// Grow ensures that n further values can be inserted without further
// allocation. It does not insert n.
//
// A Go map does not report its capacity. Therefore Grow always allocates a
// new map of size len(s)+n and copies all values; call it at most once before
// a bulk insertion.
//
// Grow panics if n is negative or too large to allocate the memory.
func (s *Set[V]) Grow(n int) {
	if n < 0 {
		panic("set: negative Grow argument")
	}
	m := make(map[V]struct{}, len(s.m)+n)
	maps.Copy(m, s.m)
	s.m = m
}

// Shrink releases unused storage. A Go map never releases memory when values
// are deleted, so the values are copied to a new map of the needed size and
// the old map can be garbage collected.
//
// Shrink does not change the values of the set. Because it reallocates, it
// is relatively expensive. Use it for long-living sets that were built from
// larger intermediate results or that lost many values.
func (s *Set[V]) Shrink() {
	if len(s.m) == 0 {
		s.m = nil
		return
	}
	m := make(map[V]struct{}, len(s.m))
	maps.Copy(m, s.m)
	s.m = m
}
