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

package semver_test

import (
	"strconv"
	"testing"

	"t73f.de/r/zero/semver"
)

var generalTestcases = []struct {
	s string
	b bool
	v semver.SemVer
}{
	{"", false, semver.SemVer{}},
	{"bad", false, semver.SemVer{}},
	{"1", false, semver.SemVer{}},
	{"1-pre", false, semver.SemVer{}},
	{"1+meta", false, semver.SemVer{}},
	{"1-pre+meta", false, semver.SemVer{}},
	{"1.2.3-", false, semver.SemVer{}},
	{"1.2.3-01", false, semver.SemVer{}},
	{"1.2.3-1.", false, semver.SemVer{}},
	{"1.2.3-alpha.01", false, semver.SemVer{}},
	{"1.2.3+", false, semver.SemVer{}},

	{"0.0.0", true, semver.SemVer{}},
	{"0.0.0", true, semver.SemVer{0, 0, 0, "", ""}},

	{"1.0.0-alpha", true, semver.SemVer{1, 0, 0, "alpha", ""}},
	{"1.0.0-alpha.1", true, semver.SemVer{1, 0, 0, "alpha.1", ""}},
	{"1.0.0-0.3.7", true, semver.SemVer{1, 0, 0, "0.3.7", ""}},
	{"1.0.0-x.7.z.92", true, semver.SemVer{1, 0, 0, "x.7.z.92", ""}},
	{"1.0.0-x-y-z.--", true, semver.SemVer{1, 0, 0, "x-y-z.--", ""}},

	{"1.0.0-alpha+001", true, semver.SemVer{1, 0, 0, "alpha", "001"}},
	{"1.0.0+20130313144700", true, semver.SemVer{1, 0, 0, "", "20130313144700"}},
	{"1.0.0-beta+exp.sha.5114f85", true, semver.SemVer{1, 0, 0, "beta", "exp.sha.5114f85"}},
	{"1.0.0+21AF26D3----117B344092BD", true, semver.SemVer{1, 0, 0, "", "21AF26D3----117B344092BD"}},
}

func TestParse(t *testing.T) {
	for _, tc := range generalTestcases {
		t.Run(tc.s, func(t *testing.T) {
			v, b := semver.Parse(tc.s)
			if b != tc.b {
				if b {
					t.Errorf("should fail, but does not: %q", tc.s)
				} else {
					t.Errorf("should parse, but does not: %q", tc.s)
				}
				return
			}
			if b && v != tc.v {
				t.Errorf("expected %v, but got %v", tc.v, v)
			}
			if pv := semver.MayParse(tc.s); b != (pv != nil) {
				t.Errorf("semver.MayParse does not match semver.Parse: %v != %v", b, pv)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but function did not panic")
		}
	}()

	_ = semver.MustParse("bad")
}

func TestIsValid(t *testing.T) {
	testcases := []struct {
		v semver.SemVer
		b bool
	}{
		{semver.SemVer{}, true},
		{semver.SemVer{1, 1, 1, "", ""}, true},
		{semver.SemVer{-1, 1, 1, "", ""}, false},
		{semver.SemVer{1, -1, 1, "", ""}, false},
		{semver.SemVer{1, 1, -1, "", ""}, false},
		{semver.SemVer{0, 0, 0, "bla", ""}, true},
		{semver.SemVer{0, 0, 0, "", "fasel"}, true},
		{semver.SemVer{0, 0, 0, "bla", "fasel"}, true},
		{semver.SemVer{0, 0, 0, " ", ""}, false},
		{semver.SemVer{0, 0, 0, "b!a", ""}, false},
		{semver.SemVer{0, 0, 0, "bla.", ""}, false},
		{semver.SemVer{0, 0, 0, "", " "}, false},
		{semver.SemVer{0, 0, 0, "", "fa?el"}, false},
		{semver.SemVer{0, 0, 0, "", "fasel."}, false},
	}
	for _, tc := range testcases {
		t.Run(tc.v.String(), func(t *testing.T) {
			if got := tc.v.IsValid(); got != tc.b {
				t.Errorf("should be %v, but got %v", tc.b, got)
			}
		})
	}
}

func TestString(t *testing.T) {
	for _, tc := range generalTestcases {
		if !tc.b {
			continue
		}
		t.Run(tc.s, func(t *testing.T) {
			v := semver.MustParse(tc.s)
			if got := v.String(); got != tc.s {
				t.Errorf("expected %q, but got %q", tc.s, got)
			}
		})
	}

	testcases := []struct {
		in  semver.SemVer
		exp string
	}{
		{semver.SemVer{}, "0.0.0"},
		{semver.SemVer{PreRelease: "abc"}, "0.0.0-abc"},
		{semver.SemVer{Build: "cafebabe.dirty"}, "0.0.0+cafebabe.dirty"},
		{semver.SemVer{PreRelease: "dev", Build: "c3f113c"}, "0.0.0-dev+c3f113c"},
	}
	for i, tc := range testcases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := tc.in.String(); tc.exp != got {
				t.Errorf("expected: %q, but got: %q", tc.exp, got)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	testcases := []struct {
		l, r string
		c    int
	}{
		{"0.0.0", "0.0.0", 0},
		{"0.0.0", "1.0.0", -1},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "2.1.0", -1},
		{"2.1.0", "2.1.1", -1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-beta.2", -1},
		{"1.0.0-beta.2", "1.0.0-beta.11", -1},
		{"1.0.0-beta.11", "1.0.0-rc.1", -1},
		{"1.0.0-rc.1", "1.0.0", -1},

		{"1.0.0-alpha+001", "1.0.0-alpha+20130313144700", 0},
		{"1.0.0-beta+20130313144700", "1.0.0-beta+exp.sha.5114f85", 0},
		{"1.0.0-beta+exp.sha.5114f85", "1.0.0-beta+21AF26D3----117B344092BD", 0},
	}
	for _, tc := range testcases {
		t.Run(tc.l+"/"+tc.r, func(t *testing.T) {
			v := semver.MustParse(tc.l)
			o := semver.MustParse(tc.r)
			c := v.Compare(o)
			if c != tc.c {
				t.Errorf("%v vs %v should result in %d, but got %d", tc.l, tc.r, tc.c, c)
			}
			c = o.Compare(v)
			if -c != tc.c {
				t.Errorf("%v vs %v should result in %d, but got %d", tc.r, tc.l, tc.c, c)
			}
		})
	}
}

func TestCompatible(t *testing.T) {
	testcases := []struct {
		l, r string
		c    bool
	}{
		{"0.0.0", "0.0.0", true},
		{"0.0.0", "0.1.0", true},
		{"0.0.0", "1.0.0", false},
		{"2.0.0", "1.0.0", false},
		{"2.0.0", "2.1.0", true},
		{"2.1.0", "2.0.0", false},
		{"2.1.0", "2.1.1", true},
		{"2.1.1", "2.1.0", true},
		{"1.0.0-alpha", "1.0.0", true},
		{"1.0.0-alpha", "1.0.0-alpha.1", true},
	}
	for _, tc := range testcases {
		t.Run(tc.l+"/"+tc.r, func(t *testing.T) {
			v := semver.MustParse(tc.l)
			o := semver.MustParse(tc.r)
			c := v.Compatible(o)
			if c != tc.c {
				t.Errorf("%v vs %v should result in %v, but got %v", tc.l, tc.r, tc.c, c)
			}
		})
	}
}

func TestInc(t *testing.T) {
	s := "1.4.16-dev+sha"

	v := semver.MustParse(s)
	v.IncPatch()
	exp := "1.4.17"
	if got := v.String(); exp != got {
		t.Errorf("IncPatch %q: expected %q, but got %q", s, exp, got)
	}

	v = semver.MustParse(s)
	v.IncMinor()
	exp = "1.5.0"
	if got := v.String(); exp != got {
		t.Errorf("IncPatch %q: expected %q, but got %q", s, exp, got)
	}

	v = semver.MustParse(s)
	v.IncMajor()
	exp = "2.0.0"
	if got := v.String(); exp != got {
		t.Errorf("IncPatch %q: expected %q, but got %q", s, exp, got)
	}
}
