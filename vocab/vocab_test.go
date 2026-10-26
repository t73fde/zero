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

package vocab_test

import (
	"fmt"
	"maps"
	"strings"
	"testing"

	"t73f.de/r/zero/vocab"
)

func TestLookupString(t *testing.T) {
	v := vocab.New(0)

	helloID := v.AddString("hello")
	worldID := v.AddBytes([]byte("world"))
	umlautID := v.AddString("Übung")
	maxWord := strings.Repeat("a", vocab.MaxWordLen)
	maxID := v.AddString(maxWord)

	tests := []struct {
		name string
		word string
		want vocab.WordID
	}{
		{"added via AddString", "hello", helloID},
		{"added via AddBytes", "world", worldID},
		{"multibyte", "Übung", umlautID},
		{"max length", maxWord, maxID},
		{"absent", "absent", vocab.InvalidWordID},
		{"prefix of present word", "hell", vocab.InvalidWordID},
		{"extension of present word", "helloo", vocab.InvalidWordID},
		{"different case", "Hello", vocab.InvalidWordID},
		{"empty", "", vocab.InvalidWordID},
		{"too long", strings.Repeat("a", vocab.MaxWordLen+1), vocab.InvalidWordID},
		{"much too long", strings.Repeat("a", 10*vocab.MaxWordLen), vocab.InvalidWordID},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := v.LookupString(tc.word); got != tc.want {
				t.Errorf("LookupString(%q) = %v, want %v", tc.word, got, tc.want)
			}
		})
	}
}

func TestLookupStringMatchesLookupBytes(t *testing.T) {
	v := vocab.New(0)
	words := []string{"alpha", "beta", "Straße", "Übung", "x"}
	for _, w := range words {
		v.AddString(w)
	}
	for _, w := range append(words, "gamma", "alph", "") {
		if s, b := v.LookupString(w), v.LookupBytes([]byte(w)); s != b {
			t.Errorf("LookupString(%q) = %v, LookupBytes = %v", w, s, b)
		}
	}
}

func TestLookupStringEmptyVocabulary(t *testing.T) {
	v := vocab.New(0)
	if got := v.LookupString("anything"); got != vocab.InvalidWordID {
		t.Errorf("LookupString on empty vocabulary = %v, want InvalidWordID", got)
	}
}

func BenchmarkLookupString(b *testing.B) {
	v := vocab.New(0)
	v.AddString("hello")
	b.ReportAllocs()
	for b.Loop() {
		v.LookupString("hello")
	}
}

func TestGetString(t *testing.T) {
	v := vocab.New(0)

	words := []string{"hello", "world", "Straße", "Übung"}
	ids := make([]vocab.WordID, len(words))
	for i, w := range words {
		ids[i] = v.AddString(w)
	}

	for i, w := range words {
		if got := v.GetString(ids[i]); got != w {
			t.Errorf("GetString(%v) = %q, want %q", ids[i], got, w)
		}
	}

	// Word added via AddBytes must be retrievable via GetString.
	id := v.AddBytes([]byte("bytes"))
	if got := v.GetString(id); got != "bytes" {
		t.Errorf("GetString(%v) = %q, want %q", id, got, "bytes")
	}

	// GetString and AppendBytes must yield identical content.
	for i, w := range words {
		if got := string(v.AppendBytes(nil, ids[i])); got != w {
			t.Errorf("AppendBytes(nil, %v) = %q, want %q", ids[i], got, w)
		}
	}
}

// newVocabulary creates a vocabulary and adds all given words.
func newVocabulary(words ...string) *vocab.Vocabulary {
	v := vocab.New(len(words))
	for _, w := range words {
		v.AddString(w)
	}
	return v
}

func TestWordCount(t *testing.T) {
	t.Parallel()

	// Boundary between inline storage (<= 14 bytes) and store (> 14 bytes).
	inline14 := strings.Repeat("a", 14)
	store15 := strings.Repeat("b", 15)
	// Longest possible word.
	const maxLen = 65535
	longest := strings.Repeat("c", maxLen)

	testcases := []struct {
		name  string
		words []string
		want  map[uint16]int
	}{
		{"empty vocabulary", nil, map[uint16]int{}},
		{"single word", []string{"House"}, map[uint16]int{5: 1}},
		{
			"same length",
			[]string{"House", "Mouse", "Horse"},
			map[uint16]int{5: 3},
		},
		{
			"different lengths",
			[]string{"a", "ab", "abc", "Mouse", "House", "Horse"},
			map[uint16]int{1: 1, 2: 1, 3: 1, 5: 3},
		},
		{
			"duplicates are counted once",
			[]string{"House", "House", "Mouse", "House"},
			map[uint16]int{5: 2},
		},
		{
			"length is measured in bytes, not runes",
			[]string{"naïve", "naive"},
			map[uint16]int{6: 1, 5: 1},
		},
		{
			"inline and store boundary",
			[]string{inline14, store15},
			map[uint16]int{14: 1, 15: 1},
		},
		{
			"longest word",
			[]string{longest, "x"},
			map[uint16]int{maxLen: 1, 1: 1},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := newVocabulary(tc.words...).WordCount()
			if !maps.Equal(got, tc.want) {
				t.Errorf("WordCount() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWordCountOneWordPerLength(t *testing.T) {
	t.Parallel()

	v := vocab.New(0)
	want := make(map[uint16]int)
	for n := 1; n <= 40; n++ {
		v.AddString(strings.Repeat("x", n))
		want[uint16(n)] = 1
	}
	if got := v.WordCount(); !maps.Equal(got, want) {
		t.Errorf("WordCount() = %v, want %v", got, want)
	}
}

// TestWordCountGrowth adds more words than the size hint, so that internal
// tables must grow, and checks that the counts stay exact.
func TestWordCountGrowth(t *testing.T) {
	t.Parallel()

	const numWords = 5000
	v := vocab.New(0)
	for i := range numWords {
		v.AddString(fmt.Sprintf("word%05d", i))           // 9 bytes each
		v.AddString(fmt.Sprintf("a-longer-word-%05d", i)) // 19 bytes each
	}
	want := map[uint16]int{9: numWords, 19: numWords}
	if got := v.WordCount(); !maps.Equal(got, want) {
		t.Errorf("WordCount() = %v, want %v", got, want)
	}
}

func TestWordCountUpdatesAfterAdd(t *testing.T) {
	t.Parallel()

	v := newVocabulary("House")
	if got, want := v.WordCount(), (map[uint16]int{5: 1}); !maps.Equal(got, want) {
		t.Fatalf("WordCount() = %v, want %v", got, want)
	}
	v.AddString("Mouse")
	v.AddString("Elephant")
	want := map[uint16]int{5: 2, 8: 1}
	if got := v.WordCount(); !maps.Equal(got, want) {
		t.Errorf("WordCount() = %v, want %v", got, want)
	}
}

// TestWordCountResultIsIndependent verifies that modifying the returned map
// neither affects the vocabulary nor later results.
func TestWordCountResultIsIndependent(t *testing.T) {
	t.Parallel()

	v := newVocabulary("House", "Mouse", "Elephant")
	first := v.WordCount()
	first[5] = 100
	delete(first, 8)
	first[99] = 1

	want := map[uint16]int{5: 2, 8: 1}
	if got := v.WordCount(); !maps.Equal(got, want) {
		t.Errorf("WordCount() = %v, want %v", got, want)
	}
}
