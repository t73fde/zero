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
// Shrink does not change the values of the slice. If s is empty, nil is
// returned, even for a non-nil empty slice. If cap(s) == len(s), s itself is
// returned and shares its storage with the result. Otherwise the result is a
// copy that shares no storage with s.
//
// Because it reallocates, Shrink is relatively expensive. Use it for
// long-living slices that were built from larger intermediate results or that
// lost many values.
func Shrink[E any](s []E) []E {
	if len(s) == 0 {
		return nil
	}
	if cap(s) == len(s) {
		return s
	}
	t := make([]E, len(s))
	copy(t, s)
	return t
}
