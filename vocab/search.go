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

package vocab

import (
	"bytes"
	"iter"
)

// The searches below are deliberately written out separately instead of
// sharing a generic filter loop. Each one can be optimized on its own (for
// example by a different use of the prefix cache) without a function call per
// word. Their logic mirrors the ...Bytes predicates above; the tests should
// check that both variants agree.
//
// All iterators work on a snapshot taken when the iteration starts: words
// added meanwhile are not visited.

// WordsContaining returns the IDs of all words that contain needle.
func (v *Vocabulary) WordsContaining(needle []byte) iter.Seq[WordID] {
	return func(yield func(WordID) bool) {
		ids, store := v.ids, v.store
		for i := 1; i < len(ids); i++ {
			s := &ids[i]
			if int(s.len) < len(needle) {
				continue
			}
			var found bool
			if s.isInline() {
				found = bytes.Contains(s.data[:s.len], needle)
			} else {
				found = (len(needle) <= cacheLen && bytes.Contains(s.cache(), needle)) ||
					bytes.Contains(s.full(store), needle)
			}
			if found && !yield(WordID(i)) {
				return
			}
		}
	}
}

// WordsWithPrefix returns the IDs of all words that start with prefix.
func (v *Vocabulary) WordsWithPrefix(prefix []byte) iter.Seq[WordID] {
	return func(yield func(WordID) bool) {
		ids, store := v.ids, v.store
		for i := 1; i < len(ids); i++ {
			s := &ids[i]
			if int(s.len) < len(prefix) {
				continue
			}
			var found bool
			switch {
			case s.isInline():
				found = bytes.HasPrefix(s.data[:s.len], prefix)
			case len(prefix) <= cacheLen:
				found = bytes.HasPrefix(s.cache(), prefix) // no store access
			default:
				found = bytes.Equal(s.cache(), prefix[:cacheLen]) &&
					bytes.HasPrefix(s.tail(store), prefix[cacheLen:])
			}
			if found && !yield(WordID(i)) {
				return
			}
		}
	}
}

// WordsWithSuffix returns the IDs of all words that end with suffix.
func (v *Vocabulary) WordsWithSuffix(suffix []byte) iter.Seq[WordID] {
	return func(yield func(WordID) bool) {
		ids, store := v.ids, v.store
		for i := 1; i < len(ids); i++ {
			s := &ids[i]
			if int(s.len) < len(suffix) {
				continue
			}
			if bytes.HasSuffix(s.view(store), suffix) && !yield(WordID(i)) {
				return
			}
		}
	}
}
