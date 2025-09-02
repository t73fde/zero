//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package runes_test

import (
	"os"
	"testing"

	"t73f.de/r/zero/runes"
)

func BenchmarkSubset(b *testing.B) {
	data, err := os.ReadFile("sample.html")
	if err != nil {
		b.Error(err)
		return
	}
	content := string(data)
	b.SetBytes(int64(len(content)))
	b.ReportAllocs()

	for b.Loop() {
		for _, r := range content {
			_ = runes.IsScalar(r)
			_ = runes.IsXmlChar(r)
			_ = runes.IsAssignable(r)
			_ = runes.IsAttributeName(r)
		}
	}
}
