//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package strings

import (
	"iter"
	"slices"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NormalizeWords produces a word list that is normalized for better searching.
func NormalizeWords(s string) (result []string) {
	return slices.Collect(NormalizeWordsSeq(s))
}

// NormalizeWordsSeq produces an iterator over normalized words.
func NormalizeWordsSeq(s string) iter.Seq[string] {
	return func(yield func(string) bool) {
		word := make([]rune, 0, 64)

		for _, r := range norm.NFKD.String(s) {
			if unicode.Is(unicode.Diacritic, r) {
				continue
			}

			if unicode.In(r, unicode.Letter, unicode.Number) {
				word = append(word, unicode.ToLower(r))
			} else if !unicode.In(r, unicode.Mark, unicode.Sk, unicode.Lm) && len(word) > 0 {
				if !yield(string(word)) {
					return
				}
				word = word[:0]
			}
		}

		if len(word) > 0 {
			yield(string(word))
		}
	}
}

// Slugify returns a string that can be used as part of an URL
func Slugify(s string) string {
	return JoinSeq(NormalizeWordsSeq(s), "-")
}
