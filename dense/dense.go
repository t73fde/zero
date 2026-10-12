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
// unsigned integer codes.
//
// Codes are assigned in order of first use, starting at 0. When a key is
// deleted, its code is freed and reused by the next new key (most recently
// freed first). Callers must remove a freed code from every structure
// derived from it, otherwise that structure refers to the next key that
// receives the code. The order of the codes reflects neither the order of
// the keys nor strictly the order of insertion.
//
// A Dense is not safe for concurrent use. The zero value is an empty Dense
// ready to use. A Dense must not be copied after first use.
package dense

// Value is the set of types that are usable as codes.
type Value interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Dense is a bijection between keys of type K and codes of type V.
// All codes in use are less than Limit().
type Dense[K comparable, V Value] struct {
	codes map[K]V
	keys  []K
	used  []bool
	free  []V
}

// Encode returns the code of key. If key is not yet known, a code is
// assigned: a freed code if available, otherwise the next unused one.
// It panics if the value range of V is exhausted.
func (d *Dense[K, V]) Encode(key K) V {
	if code, found := d.codes[key]; found {
		return code
	}
	var code V
	if n := len(d.free); n > 0 {
		code = d.free[n-1]
		d.free = d.free[:n-1]
		d.keys[code] = key
		d.used[code] = true
	} else {
		code = V(len(d.keys))
		if int(code) != len(d.keys) {
			panic("dense: code range exhausted")
		}
		d.keys = append(d.keys, key)
		d.used = append(d.used, true)
	}
	if d.codes == nil {
		d.codes = make(map[K]V)
	}
	d.codes[key] = code
	return code
}

// Find returns the code of key without assigning a new one.
func (d *Dense[K, V]) Find(key K) (V, bool) {
	code, found := d.codes[key]
	return code, found
}

// Decode returns the key that belongs to code.
func (d *Dense[K, V]) Decode(code V) (K, bool) {
	if uint64(code) < uint64(len(d.used)) && d.used[code] {
		return d.keys[code], true
	}
	var zero K
	return zero, false
}

// Remove deletes key and frees its code for reuse. It returns the freed
// code, or false if key is unknown.
func (d *Dense[K, V]) Remove(key K) (V, bool) {
	code, found := d.codes[key]
	if !found {
		return 0, false
	}
	delete(d.codes, key)
	var zero K
	d.keys[code] = zero // release the key for garbage collection
	d.used[code] = false
	d.free = append(d.free, code)
	return code, true
}

// Count returns the number of keys, i.e. the number of codes in use.
func (d *Dense[K, V]) Count() int { return len(d.codes) }

// Limit returns an upper bound for all codes in use: every code in use is
// less than Limit(). It never decreases.
func (d *Dense[K, V]) Limit() int { return len(d.keys) }
