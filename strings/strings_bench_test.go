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

var dummyS string

func BenchmarkSplitLines(b *testing.B) {
	for b.Loop() {
		for _, s := range strings.SplitLines(benchLinesText) {
			dummyS = s
		}
	}
}
func BenchmarkSplitLinesSeq(b *testing.B) {
	for b.Loop() {
		for s := range strings.SplitLineSeq(benchLinesText) {
			dummyS = s
		}
	}
}

const benchWordsText = "This, is a text; with some\nstrange runes$$änd ünicüdes"

func BenchmarkSplitWords(b *testing.B) {
	for b.Loop() {
		for _, s := range strings.SplitWords(benchWordsText) {
			dummyS = s
		}
	}
}
func BenchmarkSplitWordSeq(b *testing.B) {
	for b.Loop() {
		for s := range strings.SplitWordSeq(benchWordsText) {
			dummyS = s
		}
	}
}
