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

package strings_test

import (
	"testing"

	"t73f.de/r/zero/strings"
)

const benchLinesText = `This is test data to benchmark string.SplitLines and
strings.SplitLinesSeq.


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

package strings_test

import "testing"

const benchText = "not a quine, also no quiné";
`

func BenchmarkSplitLines(b *testing.B) {
	for b.Loop() {
		for range strings.SplitLines(benchLinesText) {
		}
	}
}
func BenchmarkSplitLinesSeq(b *testing.B) {
	for b.Loop() {
		for range strings.SplitLineSeq(benchLinesText) {
		}
	}
}

const benchWordsText = "This, is a text; with some\nstrange runes$$änd ünicüdes"

func BenchmarkSplitWords(b *testing.B) {
	for b.Loop() {
		for range strings.SplitWords(benchWordsText) {
		}
	}
}
func BenchmarkSplitWordSeq(b *testing.B) {
	for b.Loop() {
		for range strings.SplitWordSeq(benchWordsText) {
		}
	}
}
