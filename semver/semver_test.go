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
	"testing"

	"t73f.de/r/zero/semver"
)

var testcases = []struct {
	s string
	b bool
	v semver.SemVer
}{
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
	for _, tc := range testcases {
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

func TestString(t *testing.T) {
	for _, tc := range testcases {
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
}

func TestCompare(t *testing.T) {
	testcases := []struct {
		l, r string
		c    int
	}{
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
