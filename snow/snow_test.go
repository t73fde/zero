// -----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
// -----------------------------------------------------------------------------

package snow_test

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"t73f.de/r/zero/snow"
)

func TestKeyString(t *testing.T) {
	var testcases = []struct {
		key snow.Key
		exp string
	}{
		{0, "0"},
		{1, "1"},
		{0xffffffffffffffff, "FZZZZZZZZZZZZ"},
	}
	for _, tc := range testcases {
		t.Run(strconv.FormatUint(uint64(tc.key), 10), func(t *testing.T) {
			got := tc.key.String()
			if got != tc.exp {
				t.Errorf("%q expected, but got %q", tc.exp, got)
			}
			key, err := snow.Parse(got)
			if err != nil {
				panic(err)
			}
			if key != tc.key {
				t.Errorf("key %d was printed as %q, but parsed as %d/%q", tc.key, got, key, key)
			}
		})
	}
}

func TestGenerator(t *testing.T) {
	var generator snow.Generator
	var lastKey snow.Key

	for range 1000000 {
		key := generator.Create(0)
		if key <= lastKey {
			t.Errorf("key does not increase: %v -> %v", lastKey, key)
			return
		}
		lastKey = key
		checkParse(t, key)
	}

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("should panic, but did not")
			}
		}()
		_ = generator.Create(1)
	})
}

func TestNewGenerator(t *testing.T) {
	_ = snow.New(0)

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("should panic, but did not")
			}
		}()
		_ = snow.New(32)
	})
}

func checkParse(t *testing.T, key snow.Key) {
	s := key.String()
	parsedKey, err := snow.Parse(s)
	if err != nil {
		panic(err)
	}
	if parsedKey != key {
		t.Errorf("key %d/%q was parsed, but got %d/%v", key, s, parsedKey, parsedKey)
	}
}

func TestKeyID(t *testing.T) {
	for intBits := uint(0); intBits <= snow.MaxAppBits; intBits++ {
		maxID := int32(1 << intBits)
		generator := snow.New(intBits)
		if got := generator.MaxAppID(); got+1 != uint(maxID) {
			t.Errorf("MaxAppID should be %d, but is %d", maxID-1, got)
		}
		for range 512 {
			exp := uint(rand.Int31n(maxID))
			key := generator.Create(exp)
			got := generator.AppID(key)
			if got != exp {
				t.Errorf("id of %v should be %d, but got %d", key, exp, got)
			}

			checkParse(t, key)
		}
	}
}

func TestKeyID2(t *testing.T) {
	var key snow.Key
	if key.IsValid() || !key.IsInvalid() {
		t.Errorf("key %v/%d is not invalid, but should be", key, key)
	}
}

func TestParseKey(t *testing.T) {
	var testcases = []struct {
		s   string
		r   int
		exp snow.Key
	}{
		{"0000000000000", 0, 0},
		{"00-000-000-00-000", 0, 0},
		{"000-000-000-00-00", 0, 0},
		{"0-00-0-0-0-0-0-0-0-0-0-0", 0, 0},
		{"0000000000001", 0, 1},
		{"0E34NNFRTCQ15", 0, 507945423712181285},
		{"0DXZBE2D7TB04", 0, 502128752335858692},
		{"-0000000000000", 1, 0},
		{"0000000000000-", 1, 0},
		{"0DXZBE2D7<>04", 1, 0},
		{"1DXZBE2D7TB040", 2, 0},
		{"FZZZZZZZZZZZZ", 0, math.MaxUint64},
		{"F-zz-ZZZZZZZZ-zz", 0, math.MaxUint64},
	}

	for _, tc := range testcases {
		t.Run(tc.s, func(t *testing.T) {
			got, err := snow.Parse(tc.s)
			if err != nil {
				switch tc.r {
				case 0:
					t.Errorf("error %v returned, but none expected", err)
				case 1:
					if !strings.HasPrefix(err.Error(), "non base-32 character ") {
						t.Errorf("error 'non base-32 character' expected, but got: %v", err)
					}
				case 2:
					if !strings.HasPrefix(err.Error(), "does not fit in uint64: \"") {
						t.Errorf("error 'string does not fit' expected, but got: %v", err)
					}
				default:
					t.Errorf("unknown result code %d in test case", tc.r)
				}
				return
			}
			if tc.r != 0 {
				t.Error("error expected, but got value:", got)
				return
			}
			if got != tc.exp {
				t.Errorf("external key %v/%d expected, but got %v/%d", tc.exp, tc.exp, got, got)
				return
			}
			checkParse(t, got)
		})
	}
}

func TestMustParse(t *testing.T) {
	_ = snow.MustParse("0000000000000")

	t.Run("panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("should panic, but did not")
			}
		}()
		_ = snow.MustParse("-1")
	})
}

func TestKeySeq(t *testing.T) {
	generator := snow.New(0)
	key := generator.Create(0)
	lastTime := key.Time()
	lastSeqno := generator.KeySeq(key)
	for range 10000000 {
		key = generator.Create(0)
		if key.Time() != lastTime {
			lastTime = key.Time()
			lastSeqno = 0
		}
		seqno := generator.KeySeq(key)
		if lastSeqno > 0 && seqno <= lastSeqno {
			t.Error("sequence number is not increasing:", seqno, ", must be greater than:", lastSeqno)
			break
		}
		lastSeqno = seqno
	}
}
