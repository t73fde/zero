//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package graph

import (
	"cmp"
	"slices"
)

// Edge is a pair of two vertices.
type Edge[T cmp.Ordered] struct {
	From, To T
}

// EdgeSlice is a slice of Edges
type EdgeSlice[T cmp.Ordered] []Edge[T]

// Equal returns true if both slices contain the same edges in the same order.
func (es EdgeSlice[T]) Equal(other EdgeSlice[T]) bool {
	return slices.Equal(es, other)
}

// Sort the slice in-place, ordered by From, then by To.
// The sorted slice is returned for chaining.
func (es EdgeSlice[T]) Sort() EdgeSlice[T] {
	slices.SortFunc(es, func(a, b Edge[T]) int {
		return cmp.Or(cmp.Compare(a.From, b.From), cmp.Compare(a.To, b.To))
	})
	return es
}
