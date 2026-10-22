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

// Package slices provides functions for slices of any type that are missing
// in the standard package slices.
package slices

// Shrink releases unused storage. A Go slice keeps its backing array, even if
// values were deleted, so the values are copied to a new slice of exactly the
// needed size and the old backing array can be garbage collected, provided no
// other reference to it exists.
//
// Shrink is ShrinkFunc with no spare capacity. See ShrinkFunc for details.
func Shrink[E any](s []E) []E {
	return ShrinkFunc(s, nil)
}

// ShrinkFunc releases unused storage, but keeps some spare capacity.
//
// The function spare is called with len(s) and returns the number of unused
// elements that are allowed to remain. A negative result is treated as zero.
// If spare is nil, no spare capacity is kept, as if spare always returned
// zero.
//
// If the unused capacity of s, i.e. cap(s)-len(s), does not exceed that
// number, s itself is returned. Otherwise the values are copied to a new slice
// with exactly spare(len(s)) unused elements, which shares no storage with s.
// Calling ShrinkFunc again on the result with the same function therefore does
// not reallocate.
//
// If s is empty, nil is returned, even for a non-nil empty slice, and spare is
// not called.
//
// Example: keep one eighth of the length as spare capacity.
//
//	s = ShrinkFunc(s, func(n int) int { return n / 8 })
func ShrinkFunc[E any](s []E, spare func(length int) int) []E {
	n := len(s)
	if n == 0 {
		return nil
	}
	extra := 0
	if spare != nil {
		extra = max(spare(n), 0)
	}
	if extra >= cap(s)-n {
		return s
	}
	t := make([]E, n, n+extra)
	copy(t, s)
	return t
}
