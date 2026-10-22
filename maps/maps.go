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

// Package maps provides functions for maps of any type that are missing in
// the standard library package maps.
package maps

import "maps"

// Grow returns a copy of m that is sized for n further entries. The result
// does not share storage with m, and m is not modified.
//
// A Go map does not report its capacity. Therefore Grow always allocates a
// new map and copies all entries; call it at most once before a bulk
// insertion.
//
// Grow panics if n is negative.
func Grow[M ~map[K]V, K comparable, V any](m M, n int) M {
	if n < 0 {
		panic("maps: negative Grow argument")
	}
	result := make(M, len(m)+n)
	maps.Copy(result, m)
	return result
}

// Shrink releases unused storage. A Go map never releases memory when entries
// are deleted, so the entries are copied to a new map of exactly the needed
// size and the old map can be garbage collected, provided no other reference
// to it exists.
//
// Shrink is ShrinkFunc with no spare capacity. See ShrinkFunc for details.
func Shrink[M ~map[K]V, K comparable, V any](m M) M {
	return ShrinkFunc(m, nil)
}

// ShrinkFunc releases unused storage, but keeps some spare capacity.
//
// The function spare is called with len(m) and returns the number of further
// entries the new map is sized for. A negative result is treated as zero.
// If spare is nil, no spare capacity is kept, as if spare always returned
// zero.
//
// A Go map does not report its capacity. Therefore ShrinkFunc cannot detect
// whether m is oversized and always returns a new map, which shares no
// storage with m. Because of that it is relatively expensive; use it for
// long-living maps that were built from larger intermediate results or that
// lost many entries.
//
// If m is nil, nil is returned and spare is not called.
//
// Example: keep one eighth of the length as spare capacity.
//
//	m = ShrinkFunc(m, func(n int) int { return n / 8 })
func ShrinkFunc[M ~map[K]V, K comparable, V any](m M, spare func(length int) int) M {
	if m == nil {
		return nil
	}
	n := len(m)
	extra := 0
	if spare != nil {
		extra = max(spare(n), 0)
	}
	result := make(M, n+extra)
	maps.Copy(result, m)
	return result
}
