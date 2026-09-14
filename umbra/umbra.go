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

// Package umbra implements a compact string representation for in-memory search
// indexes with many short words and word fragments.
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

func (us String) cacheEqualBytes(b []byte) bool {
	return [cacheLen]byte(us.cache()) == [cacheLen]byte(b)
}
func (us String) cacheEqual(o String) bool { return us.cacheEqualBytes(o.cache()) }

// Equal reports whether us and other have equal contents.
func (us String) Equal(a *Arena, other String) bool {
	if us.len != other.len {
		return false
	}
	if us.isShort() {
		return us.data == other.data
	}
	if a.interningActive() { // Interning active -> same content = same offset
		return us.offset() == other.offset()
	}
	if !us.cacheEqual(other) {
		return false
	}
	return bytes.Equal(a.safeBytes(us.offset(), us.len), a.safeBytes(other.offset(), other.len))
}

// EqualBytes reports whether us and b have equal contents.
func (us String) EqualBytes(a *Arena, b []byte) bool {
	if int(us.len) != len(b) {
		return false
	}
	if us.isShort() {
		return bytes.Equal(us.data[:us.len], b)
	}
	if !us.cacheEqualBytes(b[:cacheLen]) {
		// len(b) == us.len > payloadLen > cacheLen, b[:cacheLen] is safe
		return false
	}
	return bytes.Equal(a.safeBytes(us.offset(), us.len), b)
}

// HasPrefixBytes reports whether us starts with prefix.
func (us String) HasPrefixBytes(a *Arena, prefix []byte) bool {
	if len(prefix) > int(us.len) {
		return false
	}
	if us.isShort() {
		return bytes.Equal(us.data[:len(prefix)], prefix)
	}
	if len(prefix) <= cacheLen {
		return bytes.Equal(us.data[cacheOff:cacheOff+len(prefix)], prefix)
	}
	return bytes.HasPrefix(a.safeBytes(us.offset(), us.len), prefix)
}

// HasSuffixBytes reports whether us ends with suffix.
func (us String) HasSuffixBytes(a *Arena, suffix []byte) bool {
	if us.isShort() {
		return bytes.HasSuffix(us.data[:us.len], suffix)
	}
	return bytes.HasSuffix(a.safeBytes(us.offset(), us.len), suffix)
}

// ContainsBytes reports whether us contains sub.
func (us String) ContainsBytes(a *Arena, sub []byte) bool {
	if us.isShort() {
		return bytes.Contains(us.data[:us.len], sub)
	}
	return bytes.Contains(a.safeBytes(us.offset(), us.len), sub)
}

// Append appends the contents of us to dst and returns the resulting slice.
func (us String) Append(dst []byte, a *Arena) []byte {
	if us.isShort() {
		return append(dst, us.data[:us.len]...)
	}
	return append(dst, a.safeBytes(us.offset(), us.len)...)
}
