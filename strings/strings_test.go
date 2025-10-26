//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2021-present Detlef Stern
//-----------------------------------------------------------------------------

package strings_test

import (
	"slices"
	"testing"

	"t73f.de/r/zero/strings"
)

func TestLength(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  string
		exp int
	}{
		{"", 0},
		{"äbc", 3},
	}
	for i, tc := range testcases {
		got := strings.Length(tc.in)
		if got != tc.exp {
			t.Errorf("%d/%q: expected %v, got %v", i, tc.in, tc.exp, got)
		}
	}
}

func TestJustifyLeft(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  string
		ml  int
		exp string
	}{
		{"", 0, ""},
		{"äbc", 0, ""},
		{"äbc", 1, "\u2025"},
		{"äbc", 2, "ä\u2025"},
		{"äbc", 3, "äbc"},
		{"äbc", 4, "äbc:"},
	}
	for i, tc := range testcases {
		got := strings.JustifyLeft(tc.in, tc.ml, ':')
		if got != tc.exp {
			t.Errorf("%d/%q/%d: expected %q, got %q", i, tc.in, tc.ml, tc.exp, got)
		}
	}
}

func TestSplitLinesAndSeq(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  string
		exp []string
	}{
		{"", nil},
		{"\n", nil},
		{"a", []string{"a"}},
		{"a\n", []string{"a"}},
		{"a\n\n", []string{"a"}},
		{"a\n\nb", []string{"a", "b"}},
	}
	for i, tc := range testcases {
		if got := strings.SplitLines(tc.in); !slices.Equal(tc.exp, got) {
			t.Errorf("%d/%q: expected %q, got %q", i, tc.in, tc.exp, got)
		}
		if got := slices.Collect(strings.SplitLineSeq(tc.in)); !slices.Equal(tc.exp, got) {
			t.Errorf("%d/%q: expected %q, got %q", i, tc.in, tc.exp, got)
		}
	}
}

func TestMakeWordsAndSeq(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  string
		exp []string
	}{
		{"", nil},
		{"\n", nil},
		{"a", []string{"a"}},
		{"a\n", []string{"a"}},
		{"a\n\n", []string{"a"}},
		{"a\n\nb", []string{"a", "b"}},
		{" ", nil},
		{"a\t", []string{"a"}},
		{"a \r", []string{"a"}},
		{"a  b", []string{"a", "b"}},
	}
	for i, tc := range testcases {
		if got := strings.SplitWords(tc.in); !slices.Equal(tc.exp, got) {
			t.Errorf("%d/%q: expected %q, got %q", i, tc.in, tc.exp, got)
		}
		if got := slices.Collect(strings.SplitWordSeq(tc.in)); !slices.Equal(tc.exp, got) {
			t.Errorf("%d/%q: expected %q, got %q", i, tc.in, tc.exp, got)
		}
	}
}
