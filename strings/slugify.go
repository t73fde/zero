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
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// NormalizeWordsSeq produces an iterator over normalized words.
func NormalizeWordsSeq(s string) iter.Seq[string] {
	return func(yield func(string) bool) {
		word := make([]byte, 0, 64)

		for _, r := range norm.NFKD.String(s) {
			if r < utf8.RuneSelf {
				switch {
				case 'A' <= r && r <= 'Z':
					word = append(word, byte(r)-'A'+'a')
				case 'a' <= r && r <= 'z', '0' <= r && r <= '9':
					word = append(word, byte(r))
				default:
					if len(word) > 0 {
						if !yield(string(word)) {
							return
						}
						word = word[:0]
					}
				}
				continue
			}

			if unicode.In(r, unicode.Mark, unicode.Diacritic) {
				continue
			}

			if unicode.In(r, unicode.Letter, unicode.Number) {
				word = utf8.AppendRune(word, unicode.ToLower(r))
			} else if len(word) > 0 {
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
