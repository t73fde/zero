//-----------------------------------------------------------------------------
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
//-----------------------------------------------------------------------------

package clock_test

import (
	"testing"
	"time"

	"t73f.de/r/zero/clock"
)

func TestUTCNow(t *testing.T) {
	before := time.Now().UTC()
	got := clock.UTCNow()
	after := time.Now().UTC()

	if got.Before(before) || got.After(after) {
		t.Errorf("UTCNow() = %v, want between %v and %v", got, before, after)
	}

	if got.Location() != time.UTC {
		t.Errorf("UTCNow() location = %v, want UTC", got.Location())
	}
}

func TestFixed(t *testing.T) {
	want := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	c := clock.Fixed(want)

	if got := c(); !got.Equal(want) {
		t.Errorf("Fixed(%v)() = %v, want %v", want, got, want)
	}

	// Aufruf mehrfach möglich, liefert immer denselben Wert
	if got := c(); !got.Equal(want) {
		t.Errorf("second call: Fixed(%v)() = %v, want %v", want, got, want)
	}
}
