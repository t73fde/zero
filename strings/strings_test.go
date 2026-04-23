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
	"fmt"
	"slices"
	"testing"
	"time"

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

func TestAnyToString(t *testing.T) {
	t.Parallel()

	type S struct{}
	type MyString string

	var nilTypedPointer *S

	testcases := []struct {
		name string
		val  any
		exp  string
	}{
		{"empty string", "", ""},
		{"int", 1, "1"},
		{"bool", true, "true"},
		{"nil interface", nil, "<nil>"},
		{"typed nil interface", nilTypedPointer, "<nil>"},
		{"time", time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC), "2020-01-02 03:04:05 +0000 UTC"},
		{"alias string", MyString("hello"), "hello"},
		{"go stringer only", OnlyGoString{V: 42}, "OnlyGoString{V:42}"},
		{"stringer only", StringerOnly{}, "stringer"},
		{"go stringer only interface", GoStringerOnly{}, "gostringer"},
		{"both stringer priority", Both{}, "string"},
		{"pointer stringer", &PointerStringer{}, "pointer"},
		{"pointer go stringer", &PointerGoStringer{}, "pointergo"},
		{"struct fallback", S{}, fmt.Sprint(S{})},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := strings.AnyToString(tc.val)
			if got != tc.exp {
				t.Errorf("%s: expected %q, got %q", tc.name, tc.exp, got)
			}
		})
	}
}

type OnlyGoString struct{ V int }

func (o OnlyGoString) GoString() string { return fmt.Sprintf("OnlyGoString{V:%d}", o.V) }

type Both struct{}

func (Both) String() string   { return "string" }
func (Both) GoString() string { return "go" }

type StringerOnly struct{}

func (StringerOnly) String() string { return "stringer" }

type GoStringerOnly struct{}

func (GoStringerOnly) GoString() string { return "gostringer" }

type PointerStringer struct{}

func (*PointerStringer) String() string { return "pointer" }

type PointerGoStringer struct{}

func (*PointerGoStringer) GoString() string { return "pointergo" }

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
