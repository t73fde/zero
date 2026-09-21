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

// ----- Base predicates: equal, prefix, suffix, contains

import "bytes"

// EqualBytes reports whether the word id is equal to b.
func (v *Vocabulary) EqualBytes(id WordID, b []byte) bool {
	s := &v.ids[id]
	if int(s.len) != len(b) {
		return false
	}
	if s.isInline() {
		return bytes.Equal(s.data[:s.len], b)
	}
	// Long word: a cache mismatch rejects without touching the store.
	return bytes.Equal(s.cache(), b[:cacheLen]) &&
		bytes.Equal(s.tail(v.store), b[cacheLen:])
}

// HasPrefixBytes reports whether the word id starts with prefix.
func (v *Vocabulary) HasPrefixBytes(id WordID, prefix []byte) bool {
	s := &v.ids[id]
	if len(prefix) > int(s.len) {
		return false
	}
	if s.isInline() {
		return bytes.HasPrefix(s.data[:s.len], prefix)
	}
	if len(prefix) <= cacheLen {
		return bytes.HasPrefix(s.cache(), prefix) // no store access
	}
	return bytes.Equal(s.cache(), prefix[:cacheLen]) &&
		bytes.HasPrefix(s.tail(v.store), prefix[cacheLen:])
}

// HasSuffixBytes reports whether the word id ends with suffix.
func (v *Vocabulary) HasSuffixBytes(id WordID, suffix []byte) bool {
	s := &v.ids[id]
	if len(suffix) > int(s.len) {
		return false
	}
	return bytes.HasSuffix(s.view(v.store), suffix)
}

// ContainsBytes reports whether the word id contains needle.
func (v *Vocabulary) ContainsBytes(id WordID, needle []byte) bool {
	s := &v.ids[id]
	if len(needle) > int(s.len) {
		return false
	}
	if s.isInline() {
		return bytes.Contains(s.data[:s.len], needle)
	}
	// Fast accept: a match inside the cached prefix needs no store access.
	if len(needle) <= cacheLen && bytes.Contains(s.cache(), needle) {
		return true
	}
	return bytes.Contains(s.full(v.store), needle)
}
