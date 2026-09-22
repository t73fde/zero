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
	"errors"
	"fmt"
	"io"
	"iter"
	"math/bits"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

// --- reference model and helpers

// model is the trivially correct reference implementation.
type model struct {
	ids   map[string]WordID
	words []string // words[id-1] is the word with that ID
}

func newModel() *model { return &model{ids: make(map[string]WordID)} }

// add mirrors Vocabulary.Add.
func (m *model) add(w string) WordID {
	if id, ok := m.ids[w]; ok {
		return id
	}
	m.words = append(m.words, w)
	id := WordID(len(m.words))
	m.ids[w] = id
	return id
}

// addAll adds all words to v and to m and checks that both agree on the IDs.
func addAll(t *testing.T, v *Vocabulary, m *model, words []string) {
	t.Helper()
	for _, w := range words {
		if got, want := v.AddBytes([]byte(w)), m.add(w); got != want {
			t.Fatalf("AddBytes(%q) = %d, want %d", w, got, want)
		}
	}
}

func assertPanics(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: no panic", name)
		}
	}()
	f()
}

// checkInvariants verifies the internal consistency of the hash table.
func checkInvariants(t *testing.T, v *Vocabulary) {
	t.Helper()
	if len(v.hashes) != len(v.hashedIDs) || uint64(len(v.hashes)) != v.mask+1 {
		t.Fatalf("inconsistent table: %d hashes, %d ids, mask %d",
			len(v.hashes), len(v.hashedIDs), v.mask)
	}
	occupied := 0
	seen := make(map[WordID]bool)
	for i, h := range v.hashes {
		if h == 0 {
			continue
		}
		occupied++
		id := v.hashedIDs[i]
		if id == 0 || int(id) >= len(v.ids) {
			t.Fatalf("position %d: invalid id %d", i, id)
		}
		if seen[id] {
			t.Fatalf("position %d: id %d stored twice", i, id)
		}
		seen[id] = true
		w := v.AppendBytes(nil, id)
		if got := v.hash(w); got != h {
			t.Fatalf("position %d: stored hash differs from hash of word %q", i, w)
		}
		if got := v.Lookup(w); got != id {
			t.Fatalf("Lookup(%q) = %d, want %d", w, got, id)
		}
	}
	if occupied != v.Len() {
		t.Fatalf("%d occupied positions, but Len() = %d", occupied, v.Len())
	}
	// Robin Hood: an entry is at most one step farther from its home position
	// than its predecessor.
	n := len(v.hashes)
	for i := range n {
		j := (i + 1) % n
		if v.hashes[i] == 0 || v.hashes[j] == 0 {
			continue
		}
		if v.dist(v.hashes[j], uint64(j)) > v.dist(v.hashes[i], uint64(i))+1 {
			t.Fatalf("Robin Hood invariant violated between positions %d and %d", i, j)
		}
	}
	if v.Len()*loadDen > len(v.hashes)*loadNum {
		t.Fatalf("load factor exceeded: %d words, %d positions", v.Len(), len(v.hashes))
	}
}

// verify compares v completely with the reference model m.
func verify(t *testing.T, v *Vocabulary, m *model) {
	t.Helper()
	if v.Len() != len(m.words) {
		t.Fatalf("Len() = %d, want %d", v.Len(), len(m.words))
	}
	for i, w := range m.words {
		id := WordID(i + 1)
		if got := v.Lookup([]byte(w)); got != id {
			t.Fatalf("Lookup(%q) = %d, want %d", w, got, id)
		}
		if got := string(v.AppendBytes(nil, id)); got != w {
			t.Fatalf("AppendBytes(%d) = %q, want %q", id, got, w)
		}
		if !v.EqualBytes(id, []byte(w)) {
			t.Fatalf("EqualBytes(%d, %q) = false", id, w)
		}
	}
	checkInvariants(t, v)
}

// checkSearches compares the ...Bytes predicates and the three searches for
// the term p with the reference computed by package bytes.
func checkSearches(t *testing.T, v *Vocabulary, m *model, p []byte) {
	t.Helper()
	var contains, prefix, suffix []WordID
	for i, w := range m.words {
		id, wb := WordID(i+1), []byte(w)
		eq, co := bytes.Equal(wb, p), bytes.Contains(wb, p)
		pre, suf := bytes.HasPrefix(wb, p), bytes.HasSuffix(wb, p)
		if got := v.EqualBytes(id, p); got != eq {
			t.Fatalf("EqualBytes(%q, %q) = %v, want %v", w, p, got, eq)
		}
		if got := v.ContainsBytes(id, p); got != co {
			t.Fatalf("ContainsBytes(%q, %q) = %v, want %v", w, p, got, co)
		}
		if got := v.HasPrefixBytes(id, p); got != pre {
			t.Fatalf("HasPrefixBytes(%q, %q) = %v, want %v", w, p, got, pre)
		}
		if got := v.HasSuffixBytes(id, p); got != suf {
			t.Fatalf("HasSuffixBytes(%q, %q) = %v, want %v", w, p, got, suf)
		}
		if co {
			contains = append(contains, id)
		}
		if pre {
			prefix = append(prefix, id)
		}
		if suf {
			suffix = append(suffix, id)
		}
	}
	if got := slices.Collect(v.WordsContaining(p)); !slices.Equal(got, contains) {
		t.Fatalf("WordsContaining(%q) = %v, want %v", p, got, contains)
	}
	if got := slices.Collect(v.WordsWithPrefix(p)); !slices.Equal(got, prefix) {
		t.Fatalf("WordsWithPrefix(%q) = %v, want %v", p, got, prefix)
	}
	if got := slices.Collect(v.WordsWithSuffix(p)); !slices.Equal(got, suffix) {
		t.Fatalf("WordsWithSuffix(%q) = %v, want %v", p, got, suffix)
	}
}

// testWords returns words around all layout boundaries: empty, inline
// (<= payloadLen), long with and without a full cache (cacheLen), multi-byte
// UTF-8, and words sharing prefixes and suffixes. It contains duplicates.
func testWords() []string {
	const base = "Internationalization-0123456789-abcdefghijklmnopqrstuvwxyz"
	words := []string{"", "House", "Houses", "Housing", "Treehouse", "Café", "Naïve"}
	lengths := []int{1, 2, cacheLen - 1, cacheLen, cacheLen + 1,
		payloadLen - 1, payloadLen, payloadLen + 1, payloadLen + 2,
		2 * payloadLen, len(base)}
	for _, n := range lengths {
		words = append(words, base[:n], base[len(base)-n:])
	}
	return words
}

// testProbes derives search terms from words: all prefixes and suffixes,
// substrings of boundary lengths, one non-matching variant per word, the empty
// term, and a term longer than any word.
func testProbes(words []string) [][]byte {
	probes := [][]byte{{}, bytes.Repeat([]byte{'z'}, 200)}
	for _, w := range words {
		for n := 0; n <= len(w); n++ {
			probes = append(probes, []byte(w[:n]), []byte(w[len(w)-n:]))
		}
		for _, n := range []int{1, cacheLen, cacheLen + 1, payloadLen, payloadLen + 1} {
			for i := 0; i+n <= len(w); i++ {
				probes = append(probes, []byte(w[i:i+n]))
			}
		}
		if len(w) > 0 {
			b := []byte(w)
			b[len(b)-1] ^= 0x01
			probes = append(probes, b)
		}
	}
	return probes
}

// --- tests

func TestBoundaryWords(t *testing.T) {
	words := testWords()
	v, m := New(len(words)), newModel()
	addAll(t, v, m, words)
	verify(t, v, m)
}

func TestSearches(t *testing.T) {
	words := testWords()
	v, m := New(len(words)), newModel()
	addAll(t, v, m, words)
	for _, p := range testProbes(m.words) {
		checkSearches(t, v, m, p)
	}
}

func TestWordLengths(t *testing.T) {
	v := New(0)
	lengths := []int{0, 1, cacheLen, cacheLen + 1, payloadLen - 1, payloadLen,
		payloadLen + 1, 100, 1000, MaxWordLen}
	for i, n := range lengths {
		w := bytes.Repeat([]byte{byte('a' + i)}, n) // distinct per length
		before := v.StoreLen()
		id := v.AddBytes(w)
		if want := WordID(i + 1); id != want {
			t.Fatalf("length %d: AddBytes = %d, want %d", n, id, want)
		}
		// Only long words are placed in the store.
		wantStore := before
		if n > payloadLen {
			wantStore += n
		}
		if got := v.StoreLen(); got != wantStore {
			t.Errorf("length %d: StoreLen = %d, want %d", n, got, wantStore)
		}
		if got := v.AppendBytes(nil, id); !bytes.Equal(got, w) {
			t.Errorf("length %d: word not restored (got %d bytes)", n, len(got))
		}
		// A second Add returns the same ID and does not grow the store.
		if again := v.AddBytes(w); again != id || v.StoreLen() != wantStore {
			t.Errorf("length %d: second AddBytes = %d, StoreLen = %d", n, again, v.StoreLen())
		}
	}
	checkInvariants(t, v)
}

func TestAddTooLong(t *testing.T) {
	v := New(0)
	v.AddBytes([]byte("House"))
	tooLong := make([]byte, MaxWordLen+1)
	assertPanics(t, "AddBytes(too long)", func() { v.AddBytes(tooLong) })
	if v.Len() != 1 || v.StoreLen() != 0 {
		t.Errorf("state changed by panic: Len = %d, StoreLen = %d", v.Len(), v.StoreLen())
	}
	if id := v.Lookup(tooLong); id != 0 {
		t.Errorf("Lookup(too long) = %d, want 0", id)
	}
}

func TestAddString(t *testing.T) {
	v := New(0)

	id1 := v.AddString("hello")
	id2 := v.AddString("hello")
	if id1 != id2 {
		t.Errorf("AddString(%q) = %v, want %v (same as first call)", "hello", id2, id1)
	}

	id3 := v.AddString("world")
	if id3 == id1 {
		t.Errorf("AddString(%q) = %v, want different WordID than AddString(%q) = %v", "world", id3, "hello", id1)
	}

	idBytes := v.AddBytes([]byte("hello"))
	if idBytes != id1 {
		t.Errorf("AddBytes(%q) = %v, want %v (same as AddString(%q))", "hello", idBytes, id1, "hello")
	}
}

func TestTableSizeFor(t *testing.T) {
	for _, tc := range []struct{ words, want int }{
		{0, 16}, {14, 16}, {15, 32}, {28, 32}, {29, 64}, {1000, 2048},
	} {
		if got := tableSizeFor(tc.words); got != tc.want {
			t.Errorf("tableSizeFor(%d) = %d, want %d", tc.words, got, tc.want)
		}
	}
}

func TestSizeHintAvoidsReallocation(t *testing.T) {
	for _, hint := range []int{0, 1, 14, 15, 1000, 49_000, 57_344, 57_345, 195_000} {
		v := New(hint)
		tableLen, idsCap := len(v.hashes), cap(v.ids)
		for i := range hint {
			v.AddBytes(fmt.Appendf(nil, "word-%d", i))
		}
		if len(v.hashes) != tableLen || cap(v.ids) != idsCap {
			t.Errorf("hint %d: slices reallocated (table %d -> %d, ids cap %d -> %d)",
				hint, tableLen, len(v.hashes), idsCap, cap(v.ids))
		}
	}
}

func TestNormalizeHash(t *testing.T) {
	h := uint64(3)
	if got := normalizeHash(h); got != h {
		t.Errorf("hash %d wrong normalized to %d", h, got)
	}
	h = uint64(0)
	if got := normalizeHash(h); got != 1 {
		t.Errorf("hash %d wrong normalized to %d", h, got)
	}
}

func TestGrowth(t *testing.T) {
	v, m := New(0), newModel()
	for i := range 5000 {
		addAll(t, v, m, []string{fmt.Sprintf("word-%d", i)})
	}
	verify(t, v, m)
}

// TestGrowthThreshold checks that the table doubles exactly when the load
// factor would be exceeded, and not earlier or later.
func TestGrowthThreshold(t *testing.T) {
	v := New(0)
	for i := 1; i <= 100_000; i++ {
		before := len(v.hashes)
		v.AddBytes([]byte("w" + strconv.Itoa(i)))
		grown := len(v.hashes) != before
		if want := i*loadDen > before*loadNum; grown != want {
			t.Fatalf("word %d: table grown = %v, want %v", i, grown, want)
		}
	}
}

func TestWriteTo(t *testing.T) {
	v := New(0)
	word := []byte("HelloWorld")
	v.AddBytes(word)
	id := v.Lookup(word)
	if id == 0 {
		panic("not found")
	}

	var buf bytes.Buffer
	n, err := v.WriteTo(&buf, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, int64(len(word)); got != want {
		t.Errorf("WriteTo() wrote %d bytes, want %d", got, want)
	}
	if got, want := buf.String(), string(word); got != want {
		t.Errorf("WriteTo() wrote %q, want %q", got, want)
	}

	var lw limitedWriter
	n, err = v.WriteTo(&lw, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, int64(len(word)); got != want {
		t.Errorf("WriteTo() wrote %d bytes, want %d", got, want)
	}
	if got, want := lw.data, word; !bytes.Equal(got, want) {
		t.Errorf("WriteTo() wrote %q, want %q", string(got), string(want))
	}

	wantErr := errors.New("write failed")
	fw := failingWriter{n: 2, err: wantErr}
	n, err = v.WriteTo(&fw, id)
	if !errors.Is(err, wantErr) {
		t.Errorf("WriteTo() error = %v, want %v", err, wantErr)
	}
	if got, want := n, int64(2); got != want {
		t.Errorf("WriteTo() wrote %d bytes, want %d", got, want)
	}

	sw := shortWriteWriter{n: 2}
	n, err = v.WriteTo(&sw, id)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("WriteTo() error = %v, want %v", err, io.ErrShortWrite)
	}
	if got, want := n, int64(2); got != want {
		t.Errorf("WriteTo() wrote %d bytes, want %d", got, want)
	}
}

type limitedWriter struct{ data []byte }
type failingWriter struct {
	n   int
	err error
}
type shortWriteWriter struct {
	n     int
	calls int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	n := min(len(p), 2)
	w.data = append(w.data, p[:n]...)
	return n, nil
}
func (w *failingWriter) Write(p []byte) (int, error) {
	n := min(len(p), w.n)
	return n, w.err
}
func (w *shortWriteWriter) Write(p []byte) (int, error) {
	if w.calls == 0 {
		w.calls++
		n := min(len(p), w.n)
		return n, nil
	}

	return 0, nil
}

func TestIteratorEarlyStop(t *testing.T) {
	v := New(0)
	for i := range 10 {
		v.AddBytes(fmt.Appendf(nil, "abc%dabc", i))
	}
	seqs := map[string]iter.Seq[WordID]{
		"WordsContaining": v.WordsContaining([]byte("bc")),
		"WordsWithPrefix": v.WordsWithPrefix([]byte("abc")),
		"WordsWithSuffix": v.WordsWithSuffix([]byte("abc")),
	}
	for name, seq := range seqs {
		calls := 0
		seq(func(WordID) bool {
			calls++
			return calls < 3
		})
		if calls != 3 {
			t.Errorf("%s: yield called %d times, want 3", name, calls)
		}
	}
}

func TestIteratorSnapshot(t *testing.T) {
	v := New(0)
	v.AddBytes([]byte("ab1"))
	v.AddBytes([]byte("ab2"))
	var got []WordID
	for id := range v.WordsWithPrefix([]byte("ab")) {
		got = append(got, id)
		// Adding words during the iteration must neither crash nor be
		// visited, even if the internal slices are reallocated.
		for i := range 50 {
			v.AddBytes(fmt.Appendf(nil, "ab-%d-%d", id, i))
		}
	}
	if !slices.Equal(got, []WordID{1, 2}) {
		t.Errorf("visited %v, want [1 2]", got)
	}
	if v.Len() != 102 {
		t.Errorf("Len() = %d, want 102", v.Len())
	}
	checkInvariants(t, v)
}

// --- randomized tests

var tokens = []string{"a", "b", "c", "ä"}

// randomWord builds a word of at most maxTokens tokens; every hundredth word
// is much longer.
func randomWord(rng *rand.Rand, maxTokens int) string {
	n := rng.IntN(maxTokens + 1)
	if rng.IntN(100) == 0 {
		n = rng.IntN(300)
	}
	var sb strings.Builder
	for range n {
		sb.WriteString(tokens[rng.IntN(len(tokens))])
	}
	return sb.String()
}

func TestRandomAgainstModel(t *testing.T) {
	for _, hint := range []int{0, 20000} {
		t.Run(fmt.Sprintf("hint=%d", hint), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(1, 2))
			v, m := New(hint), newModel()
			next := func() string {
				if len(m.words) > 0 && rng.IntN(3) == 0 {
					w := m.words[rng.IntN(len(m.words))]
					if rng.IntN(2) == 0 {
						return w // exact duplicate
					}
					return w[:rng.IntN(len(w)+1)] + randomWord(rng, 5) // shared prefix
				}
				return randomWord(rng, 40)
			}
			for range 20000 {
				addAll(t, v, m, []string{next()})
			}
			verify(t, v, m)

			// Unknown words are not found.
			for range 100 {
				w := randomWord(rng, 40)
				if _, known := m.ids[w]; !known && v.Lookup([]byte(w)) != 0 {
					t.Fatalf("Lookup(%q) found an unknown word", w)
				}
			}
			// Search terms: random parts of random known words.
			for range 100 {
				w := m.words[rng.IntN(len(m.words))]
				i := rng.IntN(len(w) + 1)
				j := i + rng.IntN(len(w)-i+1)
				checkSearches(t, v, m, []byte(w[i:j]))
			}
		})
	}
}

func TestStoreFull(t *testing.T) {
	v := New(0, WithMaxStoreLen(60))
	long1 := bytes.Repeat([]byte("a"), 60)
	long2 := bytes.Repeat([]byte("b"), payloadLen+1)
	id1 := v.AddBytes(long1) // fits exactly
	if v.StoreLen() != 60 {
		t.Fatalf("StoreLen = %d, want 60", v.StoreLen())
	}
	assertPanics(t, "AddBytes(store full)", func() { v.AddBytes(long2) })
	if v.Len() != 1 || v.StoreLen() != 60 {
		t.Errorf("state changed by panic: Len = %d, StoreLen = %d", v.Len(), v.StoreLen())
	}
	// Known long words and inline words are still accepted.
	if got := v.AddBytes(long1); got != id1 {
		t.Errorf("AddBytes(known long word) = %d, want %d", got, id1)
	}
	if got := v.AddBytes([]byte("Haus")); got == 0 {
		t.Error("AddBytes(inline word) failed")
	}
	checkInvariants(t, v)
}

func TestStoreLenZero(t *testing.T) {
	v := New(0, WithMaxStoreLen(0))
	if v.AddBytes([]byte("Haus")) == 0 { // inline words need no store
		t.Fatal("AddBytes(inline word) failed")
	}
	assertPanics(t, "AddBytes(long word)", func() { v.AddBytes(bytes.Repeat([]byte("a"), payloadLen+1)) })
}

func TestWordsFull(t *testing.T) {
	v := New(0, WithMaxWords(3))
	ids := []WordID{v.AddBytes([]byte("a")), v.AddBytes([]byte("b")), v.AddBytes([]byte("c"))}
	if !slices.Equal(ids, []WordID{1, 2, 3}) {
		t.Fatalf("ids = %v, want [1 2 3]", ids)
	}
	// A long new word must be rejected before it reaches the store.
	assertPanics(t, "AddBytes(too many words)", func() { v.AddBytes(bytes.Repeat([]byte("x"), 100)) })
	if v.Len() != 3 || v.StoreLen() != 0 {
		t.Errorf("state changed by panic: Len = %d, StoreLen = %d", v.Len(), v.StoreLen())
	}
	// Known words are still accepted.
	if got := v.AddBytes([]byte("b")); got != 2 {
		t.Errorf("AddBytes(known word) = %d, want 2", got)
	}
	checkInvariants(t, v)
}

func TestWordsZero(t *testing.T) {
	v := New(0, WithMaxWords(0))
	assertPanics(t, "Add", func() { v.AddBytes([]byte("a")) })
	if v.Lookup([]byte("a")) != 0 || v.Len() != 0 {
		t.Error("empty Vocabulary changed")
	}
}

func TestStoreCapacityBeforeFull(t *testing.T) {
	const maxStore = uint32(payloadLen + 1)
	v := New(0)
	v.maxStore = maxStore
	v.store = nil

	v.newUstr(bytes.Repeat([]byte{'x'}, payloadLen+1))

	if got := cap(v.store); got != int(maxStore) {
		t.Errorf("cap(store) = %d, want %d", maxStore, got)
	}
}

func TestOptionValues(t *testing.T) {
	assertPanics(t, "WithMaxWords(-1)", func() { New(0, WithMaxWords(-1)) })
	assertPanics(t, "WithMaxStoreLen(-1)", func() { New(0, WithMaxStoreLen(-1)) })
	if bits.UintSize == 64 { // above the limit only representable in a 64-bit int
		tooBig := uint64(MaxStoreLen) + 1
		assertPanics(t, "WithMaxWords(too big)", func() { New(0, WithMaxWords(int(tooBig))) })
		assertPanics(t, "WithMaxStoreLen(too big)", func() { New(0, WithMaxStoreLen(int(tooBig))) })
	}
	// The maximum values themselves are valid.
	New(0, WithMaxWords(MaxWords&(1<<(bits.UintSize-1)-1)), WithMaxStoreLen(0))
}

func TestSizeHintClamped(t *testing.T) {
	v := New(1000, WithMaxWords(10))
	if cap(v.ids) != 11 || len(v.hashes) != tableSizeFor(10) {
		t.Errorf("cap(ids) = %d, len(hashes) = %d", cap(v.ids), len(v.hashes))
	}
}

// TestUstrSize guards the layout of the word descriptor.
func TestUstrSize(t *testing.T) {
	if got := unsafe.Sizeof(ustr{}); got != 16 {
		t.Errorf("sizeof(ustr) = %d, want 16", got)
	}
}

// TestTableSizeBoundaries checks the doubling thresholds of the hash table,
// including the exact boundary values.
func TestTableSizeBoundaries(t *testing.T) {
	for _, tc := range []struct{ words, want int }{
		{49_000, 1 << 16}, {50_000, 1 << 16}, {57_344, 1 << 16}, {57_345, 1 << 17},
		{195_000, 1 << 18}, {200_000, 1 << 18}, {229_376, 1 << 18}, {229_377, 1 << 19}} {
		if got := tableSizeFor(tc.words); got != tc.want {
			t.Errorf("tableSizeFor(%d) = %d, want %d", tc.words, got, tc.want)
		}
	}
}

// TestSpaceBudget guards the memory consumption. It computes the bytes reserved
// by the internal slices per expected word; the numbers are deterministic
// (independent of garbage collection and allocator rounding).
func TestSpaceBudget(t *testing.T) {
	for _, tc := range []struct {
		hint   int
		budget float64 // bytes per word
	}{
		{49_000, 35.5}, {50_000, 35.1}, {57_344, 33.0}, {57_345, 46.7},
		{195_000, 35.5}, {200_000, 35.1},
	} {
		v := New(tc.hint)
		reserved := uintptr(cap(v.ids))*unsafe.Sizeof(ustr{}) +
			uintptr(cap(v.hashes))*unsafe.Sizeof(uint64(0)) +
			uintptr(cap(v.hashedIDs))*unsafe.Sizeof(WordID(0)) +
			uintptr(cap(v.store))
		if got := float64(reserved) / float64(tc.hint); got > tc.budget {
			t.Errorf("hint %d: %.2f bytes per word, budget %.2f", tc.hint, got, tc.budget)
		}
	}
}

func TestProbeLengths(t *testing.T) {
	const tableSize = 1 << 16
	for _, fill := range []float64{0.4, 0.55, 0.65, 0.75, float64(loadNum) / loadDen} {
		count := int(fill * tableSize)
		v := New(tableSize * loadNum / loadDen) // largest hint for this table size
		if len(v.hashes) != tableSize {
			t.Fatalf("table size %d, want %d", len(v.hashes), tableSize)
		}
		for i := range count {
			v.AddBytes([]byte("w" + strconv.Itoa(i)))
		}
		var sum, maxDist uint64
		for i, h := range v.hashes {
			if h != 0 {
				d := v.dist(h, uint64(i))
				sum += d
				maxDist = max(maxDist, d)
			}
		}
		t.Logf("fill %.3f: mean displacement %.2f, max %d", fill, float64(sum)/float64(count), maxDist)
	}
}
