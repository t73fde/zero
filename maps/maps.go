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

// Package maps defines various functions useful with maps of any type, which
// are missing in std package maps.
package maps

import "maps"

// Grow ensures that n further values can be inserted into m without further
// allocation. It does not insert n.
//
// A Go map does not report its capacity. Therefore Grow always allocates a
// new map of size len(s)+n and copies all values; call it at most once before
// a bulk insertion.
//
// Grow panics if n is negative or too large to allocate the memory.
func Grow[M ~map[K]V, K comparable, V any](m M, n int) M {
	if n < 0 {
		panic("maps: negative Grow argument")
	}
	result := make(M, len(m)+n)
	maps.Copy(result, m)
	return result
}

// Shrink releases unused storage. A Go map never releases memory when values
// are deleted, so the values are copied to a new map of the needed size and
// the old map can be garbage collected.
//
// Shrink does not change the values of the set. Because it reallocates, it
// is relatively expensive. Use it for long-living sets that were built from
// larger intermediate results or that lost many values.
func Shrink[M ~map[K]V, K comparable, V any](m M) M {
	if len(m) == 0 {
		return nil
	}
	result := make(M, len(m))
	maps.Copy(result, m)
	return result
}
