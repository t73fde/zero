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

package umbra

import (
	"bytes"
	"hash/maphash"
	"math"
	"sync"
)

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
// An Arena is safe for concurrent use.
type Arena struct {
	mu  sync.RWMutex
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

func (a *Arena) safeBytes(off uint32, n uint16) []byte {
	a.mu.RLock()
	b := a.rawBytes(off, n)
	a.mu.RUnlock()
	return b
}

func (a *Arena) safeEqual(off1, off2 uint32, n uint16) bool {
	a.mu.RLock()
	b := bytes.Equal(a.rawBytes(off1, n), a.rawBytes(off2, n))
	a.mu.RUnlock()
	return b
}

func (a *Arena) addBytes(s []byte) uint32 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.useIntern {
		return a.rawAppend(s)
	}

	h := maphash.Bytes(a.seed, s)
	for i := h & a.mask; ; i = (i + 1) & a.mask {
		e := &a.entries[i]
		if e.length == 0 {
			break // not found
		}
		if e.hash == h && int(e.length) == len(s) &&
			bytes.Equal(a.rawBytes(e.offset, e.length), s) {
			return e.offset
		}
	}

	off := a.rawAppend(s)
	a.insert(h, off, uint16(len(s)))
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
