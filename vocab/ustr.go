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

import "encoding/binary"

const (
	// Number of bytes of a ustr that are available for word data.
	payloadLen = 14

	// Number of payload bytes used for a store offset.
	offsetLen = 4

	// Number of payload bytes caching the prefix of a long word.
	cacheLen = payloadLen - offsetLen
)

// ustr is the 16-byte descriptor of one word.
//
// Inline word (n <= payloadLen):
//
//	data[0:n]  the word itself, remaining bytes unused
//
// Long word (n > payloadLen):
//
//	data[0:4]  offset of the word in Vocabulary.store (little endian)
//	data[4:14] cache: the first cacheLen bytes of the word
//
// The cache allows prefix, equality and (partially) contains checks without
// touching the store.
type ustr struct {
	len  uint16
	data [payloadLen]byte
}

func (s *ustr) isInline() bool { return int(s.len) <= payloadLen }

// newUstr builds the descriptor for b and, for long words, appends b to the
// store. It panics, before modifying the store, if b does not fit anymore.
func (v *Vocabulary) newUstr(b []byte) ustr {
	s := ustr{len: uint16(len(b))}
	if len(b) <= payloadLen {
		copy(s.data[:], b)
		return s
	}
	if uint64(len(v.store))+uint64(len(b)) > uint64(v.maxStore) {
		panic("vocab: store full")
	}
	binary.NativeEndian.PutUint32(s.data[:offsetLen], uint32(len(v.store)))
	copy(s.data[offsetLen:], b[:cacheLen])

	v.store = append(v.store, b...)
	if uint64(cap(v.store)) > uint64(v.maxStore) {
		store := make([]byte, len(v.store), v.maxStore)
		copy(store, v.store)
		v.store = store
	}
	return s
}

// offset returns the store offset of a long word.
func (s *ustr) offset() int {
	return int(binary.NativeEndian.Uint32(s.data[:offsetLen]))
}

// cache returns the cached prefix of a long word.
func (s *ustr) cache() []byte { return s.data[offsetLen:] }

// view returns the bytes of a word. The result must not be modified.
func (s *ustr) view(store []byte) []byte {
	if s.isInline() {
		return s.data[:s.len]
	}
	return s.full(store)
}

// full returns the complete bytes of a long word from the store.
func (s *ustr) full(store []byte) []byte {
	off := s.offset()
	return store[off : off+int(s.len)]
}

// tail returns the bytes of a long word behind the cached prefix.
func (s *ustr) tail(store []byte) []byte {
	off := s.offset()
	return store[off+cacheLen : off+int(s.len)]
}
