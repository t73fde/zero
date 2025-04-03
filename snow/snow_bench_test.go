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
	"testing"

	"t73f.de/r/zero/snow"
)

func BenchmarkSnowflake(b *testing.B) {
	var generator snow.Generator
	for b.Loop() {
		generator.Create(0)
	}
}
func BenchmarkSnowflakeX(b *testing.B) {
	bits := 7
	generator := snow.NewGenerator(uint(bits))
	key := uint((1 << bits) - 1)
	for b.Loop() {
		generator.Create(key)
	}
}
