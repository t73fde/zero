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

// Package vocab stores a growing set of words (or word fragments) in a
// compact form and maps each distinct word to a stable WordID.
//
// A Vocabulary is built for the following workload:
//
//   - words are only added, never removed (a full rebuild replaces the whole
//     Vocabulary);
//   - most words are short: about 80-90% of all words fit into 14 bytes;
//   - searches are mostly "contains", then "equals", rarely prefix or suffix.
//
// Storage layout:
//
//   - Every distinct word gets a 16-byte descriptor. Words of up to 14 bytes
//     live completely inside the descriptor.
//   - Longer words are appended once to a shared byte store. The descriptor
//     then holds the store offset and a cache of the first 10 bytes.
//   - A word is found by content through a hash table.
//
// Limits: a word is at most `MaxWordLen` bytes long, and the store of long
// words holds at most `MaxStoreLen` bytes. `Add` panics if one of these
// limits would be exceeded; `StoreLen` allows callers to check beforehand.
//
// A Vocabulary is NOT safe for concurrent use.
package vocab

import (
	"hash/maphash"
	"io"
	"math"
	"unsafe"
)

// WordID identifies a distinct word within one Vocabulary. Valid IDs start at
// 1; the zero value means "no word".
type WordID uint32

const (
	// MaxWordLen is the maximum length of a word in bytes.
	MaxWordLen = math.MaxUint16

	// MaxWords is the maximum number of distinct words.
	MaxWords = math.MaxUint32

	// MaxStoreLen is the maximum number of bytes of the store for long words.
	MaxStoreLen = math.MaxUint32

	// Smallest size of the hash table; must be a power of two.
	minTableSize = 16
)

// Vocabulary is a set of distinct words, each identified by a WordID.
type Vocabulary struct {
	// ids[id] is the canonical descriptor of the word with that ID.
	// ids[0] is unused, so that WordID 0 stays invalid.
	ids []ustr

	// store holds the bytes of all long words. Bytes are appended once and
	// never moved or removed.
	store []byte

	// Configured limits, see WithMaxWords and WithMaxStoreLen.
	maxWords uint32
	maxStore uint32

	// Open hash table, Robin Hood probing, linear steps, power-of-two size.
	// Both slices have the same length and are indexed by table position.
	//
	// hashes[i] is the hash of the word at position i; 0 marks a free
	// position (real hashes are never 0, see hash).
	// hashedIDs[i] is the WordID of that word; only valid if hashes[i] != 0.
	hashes    []uint64
	hashedIDs []WordID
	mask      uint64 // len(hashes) - 1
	seed      maphash.Seed
}

// New returns an empty Vocabulary. `sizeHint` is the expected number of
// distinct words. It is reduced to the maximum number of words if it is
// larger. The word list and the hash table are dimensioned so that adding up
// to `sizeHint` words needs no reallocation. The store for long words is
// presized by an estimate, which never exceeds the maximum store length. The
// store grows on demand if the estimate is too low. A sizeHint of 0 is valid.
func New(sizeHint int, opts ...Option) *Vocabulary {
	v := &Vocabulary{
		maxWords: MaxWords,
		maxStore: MaxStoreLen,
		seed:     maphash.MakeSeed(),
	}
	for _, opt := range opts {
		opt(v)
	}
	sizeHint = int(min(uint64(max(sizeHint, 0)), uint64(v.maxWords)))
	n := tableSizeFor(sizeHint)
	v.ids = make([]ustr, 1, sizeHint+1)
	v.hashes = make([]uint64, n)
	v.hashedIDs = make([]WordID, n)
	v.mask = uint64(n - 1)

	// storeCap is an estimate how many bytes are needed for the store of long
	// words. Roughly 20% of all words are placed into the store. The average
	// size of a long word is estimated as 16 byte. This results, as a very
	// simple estimation, that every word need 3.2 bytes, on average.
	// However, the capacity of the store may not exceed `v.maxStore` (at most:
	// `MaxStoreLen`) bytes.

	const estLongWordPercent = 20
	const estLongWordBytes = 16

	// Assert: sizeHint <= v.maxWords <= math.MaxUint32
	storeCap := min(
		uint64(sizeHint)*estLongWordPercent*estLongWordBytes/100,
		uint64(v.maxStore))
	v.store = make([]byte, 0, storeCap)
	return v
}

// Maximum load factor of the hash table, loadNum/loadDen. Variables instead of
// constants only so that tests and benchmarks can vary them.
const loadNum, loadDen = 7, 8

// tableSizeFor returns the smallest table size (a power of two, at least
// minTableSize) that holds the given number of words at a load factor of at
// most loadNum/loadDen.
func tableSizeFor(words int) int {
	n := minTableSize
	for n*loadNum < words*loadDen {
		n <<= 1
	}
	return n
}

// Option configures a Vocabulary created by New.
type Option func(*Vocabulary)

// WithMaxWords limits the number of distinct words to n, where
// 0 <= n <= MaxWords. Add panics when a further new word would exceed the
// limit. The default is MaxWords.
func WithMaxWords(n int) Option {
	return func(v *Vocabulary) {
		if n < 0 || uint64(n) > MaxWords {
			panic("vocab: invalid maximum number of words")
		}
		v.maxWords = uint32(n)
	}
}

// WithMaxStoreLen limits the store for long words to n bytes, where
// 0 <= n <= MaxStoreLen. Add panics when a new long word would not fit anymore.
// Words of up to 14 bytes are not stored there and are not affected. The
// default is MaxStoreLen.
func WithMaxStoreLen(n int) Option {
	return func(v *Vocabulary) {
		if n < 0 || uint64(n) > MaxStoreLen {
			panic("vocab: invalid maximum store length")
		}
		v.maxStore = uint32(n)
	}
}

// Len returns the number of distinct words.
func (v *Vocabulary) Len() int { return len(v.ids) - 1 }

// StoreLen returns the current length of the store for long words in bytes.
func (v *Vocabulary) StoreLen() int { return len(v.store) }

// AddBytes registers the word b and returns its ID. If the word is already
// known, the existing ID is returned. The content of b is copied.
//
// AddBytes panics if b is longer than `MaxWordLen`, or if b is a new long
// word and would not fit into the store anymore (see `MaxStoreLen`). A panic
// happens before the Vocabulary is modified.
func (v *Vocabulary) AddBytes(b []byte) WordID { return v.add(b) }

// AddString registers s in the vocabulary and returns its WordID.
// If s is already present, the existing WordID is returned; otherwise
// s is stored and a new WordID is assigned.
//
// Similar to AddBytes, it may panic under certain circumstances.
func (v *Vocabulary) AddString(s string) WordID {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	return v.add(b)
}

// Lookup returns the ID of the word b, or 0 if b was not added. Words longer
// than MaxWordLen cannot be present; Lookup does not panic for them.
func (v *Vocabulary) Lookup(b []byte) WordID {
	if len(b) > MaxWordLen {
		return 0
	}
	return v.probe(b, v.hash(b))
}

// add registers b in the vocabulary and returns its WordID. If b is
// already present, the existing WordID is returned; otherwise b is
// stored and a new WordID is assigned. add is the shared implementation
// for AddBytes and AddString.
func (v *Vocabulary) add(b []byte) WordID {
	if len(b) > MaxWordLen {
		panic("vocab: word too long")
	}
	h := v.hash(b)
	if id := v.probe(b, h); id != 0 {
		return id
	}
	// len(v.ids) is the number of words including the one to be added.
	if uint64(len(v.ids)) > uint64(v.maxWords) {
		panic("vocab: too many words")
	}
	s := v.newUstr(b) // may panic; nothing has been modified so far
	// Keep the load factor at or below loadNum/loadDen, counting the word to be added.
	if len(v.ids)*loadDen > len(v.hashes)*loadNum {
		v.grow()
	}
	id := WordID(len(v.ids))
	v.ids = append(v.ids, s)
	v.insert(h, id)
	return id
}

// hash returns the hash of b. The value 0 is reserved as free-position marker
// and mapped to 1. The resulting collision is harmless, because equal hashes
// are always confirmed by a byte comparison.
func (v *Vocabulary) hash(b []byte) uint64 {
	return normalizeHash(maphash.Bytes(v.seed, b))
}
func normalizeHash(h uint64) uint64 {
	if h == 0 {
		return 1
	}
	return h
}

// dist returns how far position i is from the home position of hash h.
func (v *Vocabulary) dist(h, i uint64) uint64 {
	return (i - (h & v.mask)) & v.mask
}

// probe returns the ID of the word b with hash h, or 0 if it is not present.
func (v *Vocabulary) probe(b []byte, h uint64) WordID {
	i := h & v.mask
	for d := uint64(0); ; d++ {
		hi := v.hashes[i]
		if hi == 0 {
			return 0 // free position: b is not present
		}
		if v.dist(hi, i) < d {
			// Robin Hood invariant: had b been inserted, it would have
			// displaced this closer-to-home entry. So b is not present.
			return 0
		}
		if hi == h {
			if id := v.hashedIDs[i]; v.EqualBytes(id, b) {
				return id
			}
		}
		i = (i + 1) & v.mask
	}
}

// insert places (h, id) into the table. The word must not be present yet and
// the table must have a free position.
func (v *Vocabulary) insert(h uint64, id WordID) {
	i := h & v.mask
	d := uint64(0)
	for {
		hi := v.hashes[i]
		if hi == 0 {
			v.hashes[i], v.hashedIDs[i] = h, id
			return
		}
		if di := v.dist(hi, i); di < d {
			// Take the position from the entry that is closer to its home
			// and continue inserting that displaced entry.
			v.hashes[i], h = h, hi
			v.hashedIDs[i], id = id, v.hashedIDs[i]
			d = di
		}
		i = (i + 1) & v.mask
		d++
	}
}

// grow doubles the table size and reinserts all entries. Hashes are stored,
// so no word has to be rehashed.
func (v *Vocabulary) grow() {
	oldHashes, oldIDs := v.hashes, v.hashedIDs
	n := len(oldHashes) * 2
	v.hashes = make([]uint64, n)
	v.hashedIDs = make([]WordID, n)
	v.mask = uint64(n - 1)
	for i, h := range oldHashes {
		if h != 0 {
			v.insert(h, oldIDs[i])
		}
	}
}

// AppendBytes appends the word with the given ID to dst and returns the
// extended slice.
func (v *Vocabulary) AppendBytes(dst []byte, id WordID) []byte {
	return append(dst, v.ids[id].view(v.store)...)
}

// WriteTo writes the word with the given ID to w and returns the number of
// bytes written.
func (v *Vocabulary) WriteTo(w io.Writer, id WordID) (int64, error) {
	b := v.ids[id].view(v.store)

	var total int64
	for len(b) > 0 {
		n, err := w.Write(b)
		total += int64(n)

		if err != nil {
			return total, err
		}
		if n == 0 {
			return total, io.ErrShortWrite
		}
		b = b[n:]
	}
	return total, nil
}
