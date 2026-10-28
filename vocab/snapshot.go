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
	"slices"
	"unsafe"
)

// Snapshot is a read-only view of the words that were registered in a
// Vocabulary when Vocabulary.Snapshot was called. Registered words are never
// modified, and the underlying arrays are only appended to, so a Snapshot
// stays valid and can be used without any lock, even if the Vocabulary
// grows afterwards.
type Snapshot struct {
	ids   []ustr
	store []byte
}

// Snapshot returns a read-only view of the current state of v.
//
// The caller must prevent concurrent modifications of v while Snapshot runs
// (e.g. by holding a read lock that also guards the Add methods). The result
// may be used without any lock afterwards.
func (v *Vocabulary) Snapshot() Snapshot {
	// Full slice expressions clip the capacity, so appending through a
	// Snapshot can never write into memory that belongs to v.
	return Snapshot{
		ids:   slices.Clip(v.ids),
		store: slices.Clip(v.store),
	}
}

// WordsWithPrefixString returns the IDs of all words that start with prefix.
func (s Snapshot) WordsWithPrefixString(prefix string) iter.Seq[WordID] {
	b := unsafe.Slice(unsafe.StringData(prefix), len(prefix))
	return s.WordsWithPrefixBytes(b)
}

// WordsWithPrefixBytes returns the IDs of all words that start with prefix.
func (s Snapshot) WordsWithPrefixBytes(prefix []byte) iter.Seq[WordID] {
	return func(yield func(WordID) bool) {
		ids, store := s.ids, s.store
		for i := 1; i < len(ids); i++ {
			u := &ids[i]
			if int(u.len) < len(prefix) {
				continue
			}
			var found bool
			switch {
			case u.isInline():
				found = bytes.HasPrefix(u.data[:u.len], prefix)
			case len(prefix) <= cacheLen:
				found = bytes.HasPrefix(u.cache(), prefix) // no store access
			default:
				found = bytes.Equal(u.cache(), prefix[:cacheLen]) &&
					bytes.HasPrefix(u.tail(store), prefix[cacheLen:])
			}
			if found && !yield(WordID(i)) {
				return
			}
		}
	}
}

// WordsWithSuffixString returns the IDs of all words that end with suffix.
func (s Snapshot) WordsWithSuffixString(suffix string) iter.Seq[WordID] {
	b := unsafe.Slice(unsafe.StringData(suffix), len(suffix))
	return s.WordsWithSuffixBytes(b)
}

// WordsWithSuffixBytes returns the IDs of all words that end with suffix.
func (s Snapshot) WordsWithSuffixBytes(suffix []byte) iter.Seq[WordID] {
	return func(yield func(WordID) bool) {
		ids, store := s.ids, s.store
		for i := 1; i < len(ids); i++ {
			u := &ids[i]
			if int(u.len) < len(suffix) {
				continue
			}
			if bytes.HasSuffix(u.view(store), suffix) && !yield(WordID(i)) {
				return
			}
		}
	}
}

// WordsContainingString returns the IDs of all words in the snapshot that
// contain needle, in ascending ID order. An empty needle matches all words.
func (s Snapshot) WordsContainingString(needle string) iter.Seq[WordID] {
	b := unsafe.Slice(unsafe.StringData(needle), len(needle))
	return s.WordsContainingBytes(b)
}

// WordsContainingBytes returns the IDs of all words in the snapshot that
// contain needle, in ascending ID order. An empty needle matches all words.
// needle must not be modified until the iteration has finished.
func (s Snapshot) WordsContainingBytes(needle []byte) iter.Seq[WordID] {
	return func(yield func(WordID) bool) {
		ids, store := s.ids, s.store
		for i := 1; i < len(ids); i++ {
			u := &ids[i]
			if int(u.len) < len(needle) {
				continue
			}
			var found bool
			if u.isInline() {
				found = bytes.Contains(u.data[:u.len], needle)
			} else {
				found = (len(needle) <= cacheLen && bytes.Contains(u.cache(), needle)) ||
					bytes.Contains(u.full(store), needle)
			}
			if found && !yield(WordID(i)) {
				return
			}
		}
	}
}

// GetString returns the word with the given WordID as a string.
func (s Snapshot) GetString(id WordID) string {
	return string(s.ids[id].view(s.store))
}

// AppendBytes appends the word with the given ID to dst and returns the
// extended slice.
func (s Snapshot) AppendBytes(dst []byte, id WordID) []byte {
	return append(dst, s.ids[id].view(s.store)...)
}
