//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of App.
//
// App is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package set_test

import (
	"slices"
	"testing"

	"t73f.de/r/app/set"
)

func TestNewHas(t *testing.T) {
	s := set.New(1, 2, 3, 2)
	if s.Contains(0) {
		panic("1")
	}
	if !s.Contains(1) {
		panic("1")
	}
	if !s.Contains(2) {
		panic("1")
	}
	if !s.Contains(3) {
		panic("1")
	}
	vals := slices.Collect(s.Values())
	if len(vals) != 3 {
		panic(vals)
	}
}
