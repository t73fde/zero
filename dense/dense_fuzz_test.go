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

package dense_test

import (
	"math/rand/v2"
	"strconv"
	"testing"

	"t73f.de/r/zero/dense"
)

const keySpace = 32 // small, so that keys and values collide often

// model is a reference implementation of the specified behavior. It stores
// the bijection as two maps and the freed values as an explicit stack, so it
// is structurally independent of Dense.
type model struct {
	vals  map[int]uint8
	keys  map[uint8]int
	free  []uint8
	limit int
}

func newModel() *model {
	return &model{vals: map[int]uint8{}, keys: map[uint8]int{}}
}

func (m *model) encode(key int) uint8 {
	if val, found := m.vals[key]; found {
		return val
	}
	var val uint8
	if n := len(m.free); n > 0 { // most recently freed value first
		val = m.free[n-1]
		m.free = m.free[:n-1]
	} else {
		val = uint8(m.limit)
		m.limit++
	}
	m.vals[key] = val
	m.keys[val] = key
	return val
}

func (m *model) removeKey(key int) (uint8, bool) {
	val, found := m.vals[key]
	if !found {
		return 0, false
	}
	m.drop(key, val)
	return val, true
}

func (m *model) removeValue(val uint8) (int, bool) {
	key, found := m.keys[val]
	if !found {
		return 0, false
	}
	m.drop(key, val)
	return key, true
}

func (m *model) drop(key int, val uint8) {
	delete(m.vals, key)
	delete(m.keys, val)
	m.free = append(m.free, val)
}

// check compares the complete observable state of d with the model.
func (m *model) check(t *testing.T, d *dense.Dense[int, uint8], step int) {
	t.Helper()
	if got := d.Count(); got != len(m.vals) {
		t.Fatalf("step %d: Count() = %d, want %d", step, got, len(m.vals))
	}
	if got := d.Limit(); got != m.limit {
		t.Fatalf("step %d: Limit() = %d, want %d", step, got, m.limit)
	}
	for key, val := range m.vals {
		if got, found := d.Find(key); !found || got != val {
			t.Fatalf("step %d: Find(%d) = (%d, %v), want (%d, true)", step, key, got, found, val)
		}
		if got, found := d.Decode(val); !found || got != key {
			t.Fatalf("step %d: Decode(%d) = (%d, %v), want (%d, true)", step, val, got, found, key)
		}
	}
	for key := range keySpace {
		if _, inUse := m.vals[key]; inUse {
			continue
		}
		if val, found := d.Find(key); found {
			t.Fatalf("step %d: Find(%d) = (%d, true) for an absent key", step, key, val)
		}
	}
	for val := range m.limit {
		if _, inUse := m.keys[uint8(val)]; inUse {
			continue
		}
		if key, found := d.Decode(uint8(val)); found {
			t.Fatalf("step %d: Decode(%d) = (%d, true) for a free value", step, val, key)
		}
	}
	if m.limit < 256 {
		if key, found := d.Decode(uint8(m.limit)); found {
			t.Fatalf("step %d: Decode(Limit()) = (%d, true)", step, key)
		}
	}
}

// runModel interprets every byte of ops as one operation on a Dense and on
// the model, and compares the results and the state after each step.
func runModel(t *testing.T, ops []byte) {
	t.Helper()
	var d dense.Dense[int, uint8]
	m := newModel()
	for step, b := range ops {
		arg := int(b/3) % keySpace
		switch b % 3 {
		case 0:
			want := m.encode(arg)
			if got := d.Encode(arg); got != want {
				t.Fatalf("step %d: Encode(%d) = %d, want %d", step, arg, got, want)
			}
		case 1:
			wantVal, wantFound := m.removeKey(arg)
			gotVal, gotFound := d.RemoveKey(arg)
			if gotVal != wantVal || gotFound != wantFound {
				t.Fatalf("step %d: RemoveKey(%d) = (%d, %v), want (%d, %v)",
					step, arg, gotVal, gotFound, wantVal, wantFound)
			}
		case 2:
			wantKey, wantFound := m.removeValue(uint8(arg))
			gotKey, gotFound := d.RemoveValue(uint8(arg))
			if gotKey != wantKey || gotFound != wantFound {
				t.Fatalf("step %d: RemoveValue(%d) = (%d, %v), want (%d, %v)",
					step, arg, gotKey, gotFound, wantKey, wantFound)
			}
		}
		m.check(t, &d, step)
	}
}

func TestModelRandom(t *testing.T) {
	for seed := range uint64(200) {
		t.Run(strconv.FormatUint(seed, 10), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(seed, 0))
			ops := make([]byte, 500)
			for i := range ops {
				ops[i] = byte(rng.UintN(256))
			}
			runModel(t, ops)
		})
	}
}

// FuzzModel executes a fuzzing test.
//
// It is started via:
//
// go test -run='^$' -fuzz=FuzzModel ./dense
func FuzzModel(f *testing.F) {
	f.Add([]byte{0, 3, 6, 1, 2, 0, 5, 8})
	f.Add([]byte{0, 0, 3, 3, 2, 2, 1, 1, 0, 3})
	f.Fuzz(func(t *testing.T, ops []byte) { runModel(t, ops) })
}
