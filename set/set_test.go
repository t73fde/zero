//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package set_test

import (
	"slices"
	"strings"
	"testing"

	"t73f.de/r/zero/set"
)

func TestNewHas(t *testing.T) {
	s := set.New(1, 2, 3, 2)
	if s.Contains(0) {
		t.Error("0")
	}
	if !s.Contains(1) {
		t.Error("1")
	}
	if !s.Contains(2) {
		t.Error("2")
	}
	if !s.Contains(3) {
		t.Error("3")
	}
	vals := slices.Collect(s.Values())
	if len(vals) != 3 {
		t.Error(vals)
	}
}

func TestSetString(t *testing.T) {
	var s *set.Set[int]
	if got := s.String(); got != "{}" {
		t.Error("nil string got:", got)
	}
	s = set.New[int]()
	if got := s.String(); got != "{}" {
		t.Error("empty string got:", got)
	}
	s = set.New(3)
	if got := s.String(); got != "{3}" {
		t.Error("{3} string got:", got)
	}
	s.Add(5)
	got := s.String()
	if !strings.ContainsRune(got, ',') {
		t.Error(s, "got not comma:", got)
	}
}

func TestSetLength(t *testing.T) {
	testdata := []struct {
		name string
		s    *set.Set[int]
		exp  int
	}{
		{"empty", nil, 0},
		{"new", set.New[int](), 0},
		{"one", set.New(3), 1},
		{"dup-one", set.New(3, 3), 1},
		{"two", set.New(3, 5), 2},
	}
	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.Length(); tc.exp != got {
				t.Errorf("set %v length exp: %d, got %d", tc.s, tc.exp, got)
			}
		})
	}
}
