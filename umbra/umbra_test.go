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

package umbra

import (
	"bytes"
	"fmt"
	"math"
	"testing"
)

func TestLengthBoundaries(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"empty", 0},
		{"just inline", 14},
		{"just too long", 15},
		{"maximum", math.MaxUint16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := bytes.Repeat([]byte("x"), tt.n)
			a := NewArena(0)
			us := a.FromBytes(s)
			if got := us.Append(nil, a); !bytes.Equal(got, s) {
				t.Fatal("Roundtrip fehlgeschlagen")
			}
		})
	}
}

func TestTooLongPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic: len>65535 byte")
		}
	}()
	NewArena(0).FromBytes(make([]byte, math.MaxUint16+1))
}

func TestEqual(t *testing.T) {
	const data = "01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := range 10 {
		t.Run(fmt.Sprintf("NewArena(%d)", i), func(t *testing.T) {
			testEqualArena(t, data, NewArena(i))
		})
	}
}

func testEqualArena(t *testing.T, data string, a *Arena) {
	t.Helper()
	prev := a.FromBytes(nil)
	for i := range data {
		content := []byte(data[:i])
		t.Run(data[:i], func(t *testing.T) {
			usA := a.FromBytes(content)
			usB := a.FromBytes(content)

			switch {
			case !usA.Equal(a, usB):
				t.Errorf("%q!=%q", usToString(usA, a), usToString(usB, a))
			case !usB.Equal(a, usA):
				t.Errorf("%q!=%q, but a==b", usToString(usB, a), usToString(usA, a))
			case i > 0:
				content[0] = '?'
				usC := a.FromBytes(content)
				if usA.Equal(a, usC) {
					t.Errorf("%q==%q (usC)", usToString(usA, a), usToString(usC, a))
				} else if usC.Equal(a, usA) {
					t.Errorf("%q==%q, but a!=c", usToString(usC, a), usToString(usA, a))
				}
				if prev.Equal(a, usA) {
					t.Errorf("%q==%q (prev)", usToString(prev, a), usToString(usA, a))
				} else if usA.Equal(a, prev) {
					t.Errorf("%q==%q, but prev!=a", usToString(usA, a), usToString(prev, a))
				}
			}
			prev = usA
		})
	}
}
func usToString(us String, a *Arena) string { return string(us.Append(nil, a)) }

func TestEqualBytes(t *testing.T) {
	const data = "01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	a := NewArena(100)
	var prev []byte
	for i := range data {
		content := []byte(data[:i])
		us := a.FromBytes(content)
		switch {
		case !us.EqualBytes(a, content):
			t.Errorf("%q!=%q", usToString(us, a), string(content))
		case i == 0:
			if !us.EqualBytes(a, prev) {
				t.Errorf("%q!=%q (nil)", usToString(us, a), string(prev))
			}
		case us.EqualBytes(a, prev):
			t.Errorf("%q==%q", usToString(us, a), prev)
		default:
			content[0] = '!'
			if us.EqualBytes(a, content) {
				t.Errorf("%q==%q (!)", usToString(us, a), content)
			}
		}
	}
}

func TestHasPrefixBytes(t *testing.T) {
	a := NewArena(0)
	usShort := a.FromBytes([]byte("01234567890"))
	prefix := []byte("01234567890ABC")
	if usShort.HasPrefixBytes(a, prefix) {
		t.Errorf("%q must not have prefix %q", usToString(usShort, a), string(prefix))
	}
}
