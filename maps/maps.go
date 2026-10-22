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

// Shrink returns a copy of m that uses no more storage than needed for its
// entries. A Go map never releases memory when entries are deleted, so the
// old map can only be garbage collected after its entries were copied.
//
// Because it reallocates, Shrink is relatively expensive. Use it for
// long-living maps that were built from larger intermediate results or that
// lost many entries.
func Shrink[M ~map[K]V, K comparable, V any](m M) M {
	if m == nil {
		return nil
	}
	result := make(M, len(m))
	maps.Copy(result, m)
	return result
}
