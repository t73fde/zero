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

// Equal return true if both slices are the same.
func (es EdgeSlice[T]) Equal(other EdgeSlice[T]) bool {
	return slices.Equal(es, other)
}

// Sort the slice.
func (es EdgeSlice[T]) Sort() EdgeSlice[T] {
	slices.SortFunc(es, func(e1, e2 Edge[T]) int {
		if e1.From < e2.From {
			return -1
		}
		if e1.From > e2.From {
			return 1
		}
		if e1.To < e2.To {
			return -1
		}
		if e1.To > e2.To {
			return 1
		}
		return 0
	})
	return es
}
