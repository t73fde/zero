//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package strings_test

import (
	"slices"
	"testing"

	"t73f.de/r/zero/strings"
)

func TestSlugify(t *testing.T) {
	t.Parallel()
	testcases := []struct{ in, exp string }{
		{"simple test", "simple-test"},
		{"I'm a go developer", "i-m-a-go-developer"},
		{"-!->simple   test<-!-", "simple-test"},
		{"äöüÄÖÜß", "aouaouß"},
		{"\"aèf", "aef"},
		{"a#b", "a-b"},
		{"*", ""},
	}
	for _, tc := range testcases {
		if got := strings.Slugify(tc.in); got != tc.exp {
			t.Errorf("%q: %q != %q", tc.in, got, tc.exp)
		}
	}
}

func TestNormalizeWord(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  string
		exp []string
	}{
		{"", nil},
		{" ", nil},
		{"ˋ", nil}, // No single diacritic char, such as U+02CB
		{"simple test", []string{"simple", "test"}},
		{"I'm a go developer", []string{"i", "m", "a", "go", "developer"}},
		{"-!->simple   test<-!-", []string{"simple", "test"}},
		{"äöüÄÖÜß", []string{"aouaouß"}},
		{"\"aèf", []string{"aef"}},
		{"a#b", []string{"a", "b"}},
		{"*", nil},
		{"123", []string{"123"}},
		{"1²3", []string{"123"}},
		{"Period.", []string{"period"}},
		{" WORD  NUMBER ", []string{"word", "number"}},
		{"^ABC$", []string{"abc"}},
	}
	for _, tc := range testcases {
		if got := strings.NormalizeWords(tc.in); !slices.Equal(got, tc.exp) {
			t.Errorf("%q: %q != %q", tc.in, got, tc.exp)
		}
		got := slices.Collect(strings.NormalizeWordsSeq(tc.in))
		if !slices.Equal(got, tc.exp) {
			t.Errorf("%q: %q != %q", tc.in, got, tc.exp)
		}
	}
}
