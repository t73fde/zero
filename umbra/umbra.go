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

// Package umbra implements a compact string representation, e.g. for in-memory
// search indexes with many short words and word fragments or similar
// applications with matching contraints.
//
// Each string is encoded in a 16-byte struct: 2 bytes length, 14 bytes payload.
// Strings up to 14 bytes are stored entirely inline, with no allocation and no
// indirection. Longer strings store an 8-byte cache of the beginning inline
// plus an offset into a shared arena; most comparisons avoid touching the arena
// at all.
//
// The arena manages the storage for long strings in a growing buffer and
// optionally deduplicates identical fragments via an embedded hash table
// (interning).
//
// Strings are limited to 65535 bytes. A string must not be used with any arena
// other than the one it was created from.
package umbra

import (
	"bytes"
	"encoding/binary"
	"hash/maphash"
	"math"
)

// String is a compact, 16-byte string value: 2 bytes length, 14 bytes payload.
// Short strings (len<=14) hold their raw bytes directly in data. Long strings
// store an offset and a prefix cache in data instead.
type String struct {
	len  uint16
	data [payloadLen]byte
}

const payloadLen = 14

// Layout for long strings (len > payloadLen):
//
//	data[0:2]   reserved for future use, currently 0
//	data[2:6]   arena offset (uint32)
//	data[6:14]  cache of the first eight bytes
const (
	offsetOff = 2
	offsetLen = 4
	cacheOff  = 6
	cacheLen  = 8
)

func (us String) isShort() bool { return us.len <= payloadLen }

func (us String) offset() uint32 {
	return binary.NativeEndian.Uint32(us.data[offsetOff : offsetOff+offsetLen])
}
func (us *String) setOffset(off uint32) {
	binary.NativeEndian.PutUint32(us.data[offsetOff:offsetOff+offsetLen], off)
}
func (us String) cache() []byte { return us.data[cacheOff : cacheOff+cacheLen] }

// Arena stores the byte data of long strings (those exceeding the inline
// capacity of a String) in a single, growing buffer. Offsets into the buffer
// stay valid across internal reallocations, since they are relative positions
// rather than pointers.
//
// Interning is optional: when enabled via NewArena, identical long fragments
// are deduplicated through an embedded open-addressing hash table, so equal
// content always maps to the same offset. Entries are never removed; content
// that is no longer referenced simply stays unused in the buffer.
//
// An Arena is **NOT** safe for concurrent use.
type Arena struct {
	buf []byte

	useIntern bool
	seed      maphash.Seed
	entries   []internEntry
	mask      uint64
	count     int
}
type internEntry struct {
	hash   uint64
	offset uint32
	length uint16 // 0 = empty Slot
}

// NewArena creates an Arena object. sizeHint gives a hint about the expected
// number of strings to be interned. 0 deactives interning.
func NewArena(sizeHint int, useIntern bool) *Arena {
	const avgLongStringLen = 16

	a := &Arena{
		buf:       make([]byte, 0, max(0, sizeHint)*avgLongStringLen),
		useIntern: useIntern,
	}
	if useIntern {
		a.seed = maphash.MakeSeed()
		size := 16
		for sizeHint*10 >= size*7 { // 70% Load Factor
			size *= 2
		}
		a.entries = make([]internEntry, size)
		a.mask = uint64(size - 1)
	}
	return a
}

// FromBytes builds a String with the given content. The content is stored
// in the arena, if its length exceeds 14 bytes. Otherwise, the content is
// stored as a payload of the String.
func (a *Arena) FromBytes(b []byte) String {
	n := len(b)
	if n > math.MaxUint16 {
		panic("umbra.String: capacity 65535 bytes exceeded")
	}
	if n+len(a.buf) > math.MaxUint32 {
		panic("umbra.Arena: capacity 2**32 bytes exceeded")
	}
	var us String
	us.len = uint16(n)
	if n <= payloadLen {
		copy(us.data[:n], b)
		return us
	}
	copy(us.data[cacheOff:cacheOff+cacheLen], b)
	us.setOffset(a.addBytes(b))
	return us
}

// Equal reports whether us and other have equal contents.
func (a *Arena) Equal(us String, other String) bool {
	if us.len != other.len {
		return false
	}
	if us.isShort() {
		return us.data == other.data
	}
	if a.useIntern { // Interning active -> same content = same offset
		return us.offset() == other.offset()
	}
	if !bytes.Equal(us.cache(), other.cache()) {
		return false
	}
	n := us.len - cacheLen
	return bytes.Equal(a.rawBytes(us.offset()+cacheLen, n), a.rawBytes(other.offset()+cacheLen, n))
}

// EqualBytes reports whether us and b have equal contents.
func (a *Arena) EqualBytes(us String, b []byte) bool {
	if int(us.len) != len(b) {
		return false
	}
	if us.isShort() {
		return bytes.Equal(us.data[:us.len], b)
	}
	if !bytes.Equal(us.cache(), b[:cacheLen]) {
		// len(b) == us.len > payloadLen > cacheLen, b[:cacheLen] is safe
		return false
	}
	return bytes.Equal(a.rawBytes(us.offset()+cacheLen, us.len-cacheLen), b[cacheLen:])
}

// HasPrefixBytes reports whether us starts with prefix.
func (a *Arena) HasPrefixBytes(us String, prefix []byte) bool {
	if len(prefix) > int(us.len) {
		return false
	}
	if us.isShort() {
		return bytes.Equal(us.data[:len(prefix)], prefix)
	}
	if len(prefix) <= cacheLen {
		return bytes.Equal(us.data[cacheOff:cacheOff+len(prefix)], prefix)
	}
	return bytes.HasPrefix(a.rawBytes(us.offset(), us.len), prefix)
}

// HasSuffixBytes reports whether us ends with suffix.
func (a *Arena) HasSuffixBytes(us String, suffix []byte) bool {
	if us.isShort() {
		return bytes.HasSuffix(us.data[:us.len], suffix)
	}
	return bytes.HasSuffix(a.rawBytes(us.offset(), us.len), suffix)
}

// ContainsBytes reports whether us contains sub.
func (a *Arena) ContainsBytes(us String, sub []byte) bool {
	if us.isShort() {
		return bytes.Contains(us.data[:us.len], sub)
	}
	return bytes.Contains(a.rawBytes(us.offset(), us.len), sub)
}

// Append appends the contents of us to dst and returns the resulting slice.
func (a *Arena) Append(dst []byte, us String) []byte {
	if us.isShort() {
		return append(dst, us.data[:us.len]...)
	}
	return append(dst, a.rawBytes(us.offset(), us.len)...)
}

func (a *Arena) addBytes(b []byte) uint32 {
	if !a.useIntern {
		return a.rawAppend(b)
	}

	h := maphash.Bytes(a.seed, b)
	for i := h & a.mask; ; i = (i + 1) & a.mask {
		e := &a.entries[i]
		if e.length == 0 {
			break // not found
		}
		if e.hash == h && int(e.length) == len(b) &&
			bytes.Equal(a.rawBytes(e.offset, e.length), b) {
			return e.offset
		}
	}

	off := a.rawAppend(b)
	a.insert(h, off, uint16(len(b)))
	return off
}

func (a *Arena) rawBytes(off uint32, n uint16) []byte {
	end := off + uint32(n)
	return a.buf[off:end:end]
}

func (a *Arena) rawAppend(s []byte) uint32 {
	off := uint32(len(a.buf))
	a.buf = append(a.buf, s...)
	return off
}

func (a *Arena) insert(h uint64, offset uint32, length uint16) {
	if (a.count+1)*10 >= len(a.entries)*7 {
		a.growTable()
	}
	for i := h & a.mask; ; i = (i + 1) & a.mask {
		if a.entries[i].length == 0 {
			a.entries[i] = internEntry{hash: h, offset: offset, length: length}
			a.count++
			return
		}
	}
}

func (a *Arena) growTable() {
	old := a.entries
	a.entries = make([]internEntry, len(old)*2)
	a.mask = uint64(len(a.entries) - 1)
	for _, e := range old {
		if e.length == 0 {
			continue
		}
		for i := e.hash & a.mask; ; i = (i + 1) & a.mask {
			if a.entries[i].length == 0 {
				a.entries[i] = e
				break
			}
		}
	}
}
