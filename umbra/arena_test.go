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
	"fmt"
	"sync"
	"testing"
)

func TestInterningDedup(t *testing.T) {
	a := NewArena(2, true) // small enough to trigger resize of hash table
	seen := map[string]uint32{}
	words := []string{
		"donaudampfschifffahrtsgesellschaft",
		"dampfschifffahrtskapitaensmuetze",
		"donaudampfschifffahrtsgesellschaft",
		"kapitaensmuetzenvorschrift",
		"dampfschifffahrtskapitaensmuetze",
	}
	for _, w := range words {
		off := a.addBytes([]byte(w))
		if prev, ok := seen[w]; ok && prev != off {
			t.Fatalf("same content %q got different offsets: %d vs %d", w, prev, off)
		}
		seen[w] = off
		if !bytes.Equal(a.safeBytes(off, uint16(len(w))), []byte(w)) {
			t.Fatalf("Offset %d does not point to %q", off, w)
		}
	}
}

func TestConcurrentAdd(*testing.T) {
	a := NewArena(100, true)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a.addBytes(fmt.Appendf(nil, "word-number-%d", i%10)) // creates duplicates
		}(i)
	}
	wg.Wait()
}

func TestConcurrentReadWrite(*testing.T) {
	a := NewArena(100, true)
	var wg sync.WaitGroup

	initial := make([]String, 20)
	for i := range initial {
		initial[i] = a.FromBytes(fmt.Appendf(nil, "prefix-word-number-%d", i))
	}

	wg.Add(2)
	go func() { // write new strings
		defer wg.Done()
		for i := range 5000 {
			a.addBytes(fmt.Appendf(nil, "very-new-word-%d", i))
		}
	}()
	go func() { // parallel read
		defer wg.Done()
		for i := range 5000 {
			us := initial[i%len(initial)]
			_ = a.Append(nil, us)
			_ = a.HasPrefixBytes(us, []byte("current-word"))
			_ = a.Equal(us, us)
		}
	}()
	wg.Wait()
}
