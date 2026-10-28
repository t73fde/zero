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

// Package dense provides a bijection between comparable keys and dense
// unsigned integer values.
//
// Values are assigned in order of first use, starting at 0. When a key is
// removed, its value is freed and reused by the next new key (most recently
// freed first). Callers must remove a freed value from every structure
// derived from it, otherwise that structure refers to the next key that
// receives the value. Therefore, the order of the values reflects neither
// the order of the keys nor, after a removal, the order of insertion.
//
// Keys must be equal to themselves, that is, key == key must hold. This
// excludes floating-point NaN as well as arrays, structs, and interface
// values that contain a NaN. Such keys can never be found again, so every
// Encode of one creates a new entry that neither Find nor RemoveKey can
// reach, and Count becomes unreliable. As with Go maps, an interface-typed
// key whose dynamic type is not comparable causes a run-time panic.
//
// A Dense is not safe for concurrent use. The zero value is an empty Dense
// ready to use. A Dense must not be copied after first use.
package dense

import (
	"iter"
	"slices"

	zeromaps "t73f.de/r/zero/maps"
	zeroslices "t73f.de/r/zero/slices"
)

// Value is the set of types that are usable as values of a Dense.
type Value interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Dense is a bijection between keys of type K and values of type V.
// All values in use are less than Limit().
type Dense[K comparable, V Value] struct {
	values map[K]V
	keys   []K
	used   []bool
	free   []V
}

// Encode returns the val of key. If key is not yet known, a value is
// assigned: a freed val if available, otherwise the next unused one.
// It panics if the value range of V is exhausted.
// The key must satisfy key == key (see package documentation).
func (d *Dense[K, V]) Encode(key K) V {
	if val, found := d.values[key]; found {
		return val
	}
	var val V
	if n := len(d.free); n > 0 {
		val = d.free[n-1]
		d.free = d.free[:n-1]
		d.keys[val] = key
		d.used[val] = true
	} else {
		val = V(len(d.keys))
		if int(val) != len(d.keys) {
			panic("dense: value range exhausted")
		}
		d.keys = append(d.keys, key)
		d.used = append(d.used, true)
	}
	if d.values == nil {
		d.values = make(map[K]V)
	}
	d.values[key] = val
	return val
}

// Find returns the value of key without assigning a new one.
func (d *Dense[K, V]) Find(key K) (V, bool) {
	val, found := d.values[key]
	return val, found
}

// Decode returns the key that belongs to val.
func (d *Dense[K, V]) Decode(val V) (K, bool) {
	if uint64(val) < uint64(len(d.used)) && d.used[val] {
		return d.keys[val], true
	}
	var zero K
	return zero, false
}

// RemoveKey removes key and frees its value for reuse. It returns the
// freed value, or false if key is unknown.
func (d *Dense[K, V]) RemoveKey(key K) (V, bool) {
	val, found := d.values[key]
	if !found {
		return 0, false
	}
	d.release(key, val)
	return val, true
}

// RemoveValue removes val together with its key and frees val for reuse.
// It returns the removed key, or false if val is not in use.
//
// It is equivalent to, but more efficient than:
//
//	if key, ok := d.Decode(val); ok {
//		d.RemoveKey(key)
//		return key, true
//	}
//	var zero K
//	return zero, false
func (d *Dense[K, V]) RemoveValue(val V) (K, bool) {
	if uint64(val) >= uint64(len(d.used)) || !d.used[val] {
		var zero K
		return zero, false
	}
	key := d.keys[val]
	d.release(key, val)
	return key, true
}

// release frees val, which must be in use and belong to key.
func (d *Dense[K, V]) release(key K, val V) {
	delete(d.values, key)
	var zero K
	d.keys[val] = zero // allow the key to be garbage collected
	d.used[val] = false
	d.free = append(d.free, val)
}

// Count returns the number of keys, i.e. the number of values in use.
func (d *Dense[K, V]) Count() int { return len(d.values) }

// Limit returns an upper bound for all values in use: every value in use is
// less than Limit(). It never decreases.
func (d *Dense[K, V]) Limit() int { return len(d.keys) }

// All returns an iterator over all key-value pairs of d, in ascending order
// of the values.
//
// Removing keys during the iteration is allowed: a removed key that was not
// yet reached is not produced. Whether keys added during the iteration are
// produced is unspecified.
func (d *Dense[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// d.keys and d.used are read anew in every step, because yield may
		// modify d.
		for i := 0; i < len(d.keys); i++ {
			if d.used[i] && !yield(d.keys[i], V(i)) {
				return
			}
		}
	}
}

// Keys returns an iterator over all keys of d, in ascending order of their
// values. See All() for modifications during the iteration.
func (d *Dense[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for key := range d.All() {
			if !yield(key) {
				return
			}
		}
	}
}

// Values returns an iterator over all values in use, in ascending order.
// See All() for modifications during the iteration.
func (d *Dense[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, val := range d.All() {
			if !yield(val) {
				return
			}
		}
	}
}

// Grow ensures that n further values can be inserted into d without further
// allocation.
func (d *Dense[K, V]) Grow(n int) {
	d.values = zeromaps.Grow(d.values, n)
	d.keys = slices.Grow(d.keys, n)
	d.used = slices.Grow(d.used, n)

	// do not touch d.free since it is only used for remove operations.
}

// Shrink releases unused storage.
func (d *Dense[K, V]) Shrink() {
	d.values = zeromaps.Shrink(d.values)
	d.keys = zeroslices.Shrink(d.keys)
	d.used = zeroslices.Shrink(d.used)
	d.free = zeroslices.Shrink(d.free)
}

// ShrinkFunc releases unused storage, but keeps some spare capacity.
//
// The function spare is called with r.Count() and returns the number of unused
// elements that are allowed to remain. A negative result is treated as zero.
// If spare is nil, no spare capacity is kept, as if spare always returned
// zero.
//
//	r.ShrinkFunc(s, func(n int) int { return n / 8 })
func (d *Dense[K, V]) ShrinkFunc(spare func(int) int) {
	d.values = zeromaps.ShrinkFunc(d.values, spare)
	d.keys = zeroslices.ShrinkFunc(d.keys, spare)
	d.used = zeroslices.ShrinkFunc(d.used, spare)
	d.free = zeroslices.ShrinkFunc(d.free, spare)
}
