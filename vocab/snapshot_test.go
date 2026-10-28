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
	"iter"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"t73f.de/r/zero/vocab"
)

func TestSnapshotWordsContaining(t *testing.T) {
	v := vocab.New(0)
	words := []string{
		"hello", "shell", "world", "Übung", "a",
		"needle" + strings.Repeat("x", 40),                           // long, needle at start
		strings.Repeat("x", 20) + "needle" + strings.Repeat("y", 20), // long, needle in middle
		strings.Repeat("y", 40) + "needle",                           // long, needle at end
		strings.Repeat("z", 46),                                      // long, no needle
	}
	ids := make([]vocab.WordID, len(words))
	for i, w := range words {
		ids[i] = v.AddString(w)
	}
	s := v.Snapshot()

	needles := []string{
		"ell", "hello", "o", "Üb", "ung", "a", "needle", "xxy", "x", "zz",
		"hellx", "notpresent", "", strings.Repeat("z", 47),
	}
	for _, needle := range needles {
		var want []vocab.WordID
		for i, w := range words {
			if strings.Contains(w, needle) {
				want = append(want, ids[i])
			}
		}
		if got := slices.Collect(s.WordsContainingString(needle)); !slices.Equal(got, want) {
			t.Errorf("WordsContainingString(%q) = %v, want %v", needle, got, want)
		}
		if got := slices.Collect(s.WordsContainingBytes([]byte(needle))); !slices.Equal(got, want) {
			t.Errorf("WordsContainingBytes(%q) = %v, want %v", needle, got, want)
		}
	}
}

func TestSnapshotEmptyVocabulary(t *testing.T) {
	s := vocab.New(0).Snapshot()
	if got := slices.Collect(s.WordsContainingString("a")); len(got) != 0 {
		t.Errorf("WordsContainingString on empty snapshot = %v, want no IDs", got)
	}
}

func TestSnapshotGetStringAppendBytes(t *testing.T) {
	v := vocab.New(0)
	words := []string{"hello", "Straße", strings.Repeat("x", 40)}
	ids := make([]vocab.WordID, len(words))
	for i, w := range words {
		ids[i] = v.AddString(w)
	}
	s := v.Snapshot()
	for i, w := range words {
		if got := s.GetString(ids[i]); got != w {
			t.Errorf("GetString(%v) = %q, want %q", ids[i], got, w)
		}
		if got := string(s.AppendBytes([]byte("pre:"), ids[i])); got != "pre:"+w {
			t.Errorf("AppendBytes(%v) = %q, want %q", ids[i], got, "pre:"+w)
		}
	}
}

func TestSnapshotIsolation(t *testing.T) {
	v := vocab.New(0)
	idHello := v.AddString("hello")
	s := v.Snapshot()

	// Enough additions to force reallocation of the internal arrays.
	const extra = 10000
	for i := range extra {
		v.AddString("word" + strconv.Itoa(i) + "hello")
	}

	if got := slices.Collect(s.WordsContainingString("hello")); !slices.Equal(got, []vocab.WordID{idHello}) {
		t.Errorf("old snapshot sees %v, want only [%v]", got, idHello)
	}
	if got := s.GetString(idHello); got != "hello" {
		t.Errorf("GetString on old snapshot = %q, want %q", got, "hello")
	}
	if n := len(slices.Collect(v.Snapshot().WordsContainingString("hello"))); n != extra+1 {
		t.Errorf("new snapshot finds %d words, want %d", n, extra+1)
	}
}

func TestSnapshotWordsContainingBreak(t *testing.T) {
	v := vocab.New(0)
	for _, w := range []string{"ab", "cab", "dab"} {
		v.AddString(w)
	}
	n := 0
	for range v.Snapshot().WordsContainingString("ab") {
		n++
		break
	}
	if n != 1 {
		t.Errorf("iteration yielded %d IDs before break, want 1", n)
	}
}

// TestSnapshotConcurrentUse is meaningful only with -race.
func TestSnapshotConcurrentUse(t *testing.T) {
	v := vocab.New(0)
	for i := range 100 {
		v.AddString("base" + strconv.Itoa(i))
	}
	s := v.Snapshot()
	want := slices.Collect(s.WordsContainingString("base"))

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				got := slices.Collect(s.WordsContainingString("base"))
				if !slices.Equal(got, want) {
					t.Errorf("concurrent read: got %d IDs, want %d", len(got), len(want))
					return
				}
			}
		}()
	}
	// Only this goroutine touches v; the readers use the snapshot only.
	for i := range 10000 {
		v.AddString("later" + strconv.Itoa(i))
	}
	wg.Wait()
}

func TestIteratorEarlyStop(t *testing.T) {
	v := vocab.New(0)
	for i := range 10 {
		v.AddBytes(fmt.Appendf(nil, "abc%dabc", i))
	}
	sn := v.Snapshot()
	seqs := map[string]iter.Seq[vocab.WordID]{
		"WordsContainingBytes":  sn.WordsContainingBytes([]byte("bc")),
		"WordsWithPrefixBytes":  sn.WordsWithPrefixBytes([]byte("abc")),
		"WordsWithSuffixBytes":  sn.WordsWithSuffixBytes([]byte("abc")),
		"WordsContainingString": sn.WordsContainingString("bc"),
		"WordsWithPrefixString": sn.WordsWithPrefixString("abc"),
		"WordsWithSuffixString": sn.WordsWithSuffixString("abc"),
	}
	for name, seq := range seqs {
		calls := 0
		seq(func(vocab.WordID) bool {
			calls++
			return calls < 3
		})
		if calls != 3 {
			t.Errorf("%s: yield called %d times, want 3", name, calls)
		}
	}
}
