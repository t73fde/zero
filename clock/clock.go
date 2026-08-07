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

// Package clock provides an abstraction over the current time,
// so callers can inject a fixed time in tests instead of relying
// on time.Now() directly.
package clock

import "time"

// Clock returns the current time. Production code uses UTCNow;
// tests can inject Fixed(t) for deterministic behavior.
type Clock func() time.Time

// UTCNow returns the current time in UTC. This is the standard
// Clock implementation used throughout the application.
func UTCNow() time.Time {
	return time.Now().UTC()
}

// Fixed returns a Clock that always returns t. Useful for tests.
func Fixed(t time.Time) Clock {
	return func() time.Time {
		return t
	}
}
