//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package set_test

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"t73f.de/r/zero/set"
)

// ----- Helpers

// checkSet verifies that s contains exactly the values in exp
// (any order, duplicates allowed).
func checkSet[V cmp.Ordered](t *testing.T, name string, s set.Set[V], exp ...V) {
	t.Helper()
	want := slices.Compact(slices.Sorted(slices.Values(exp)))
	if got := slices.Sorted(s.Values()); !slices.Equal(got, want) {
		t.Errorf("%s: values exp %v, got %v", name, want, got)
	}
	if got := s.Count(); got != len(want) {
		t.Errorf("%s: count exp %d, got %d", name, len(want), got)
	}
	if got := s.IsEmpty(); got != (len(want) == 0) {
		t.Errorf("%s: IsEmpty exp %v, got %v", name, len(want) == 0, got)
	}
	for _, v := range want {
		if !s.Contains(v) {
			t.Errorf("%s: does not contain %v", name, v)
		}
	}
}

func ints(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = i
	}
	return r
}

// ----- Constructors and basic operations

func TestNewHas(t *testing.T) {
	s := set.New(1, 2, 3, 2)
	checkSet(t, "New", s, 1, 2, 3)
	if s.Contains(0) {
		t.Error("contains 0")
	}
	checkSet(t, "New empty", set.New[int]())
	checkSet(t, "New strings", set.New("b", "a", "b"), "a", "b")
}

func TestSetCount(t *testing.T) {
	testdata := []struct {
		name string
		s    set.Set[int]
		exp  int
	}{
		{"empty", set.Set[int]{}, 0},
		{"new", set.New[int](), 0},
		{"one", set.New(3), 1},
		{"dup-one", set.New(3, 3), 1},
		{"two", set.New(3, 5), 2},
	}
	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.Count(); tc.exp != got {
				t.Errorf("set %v length exp: %d, got %d", tc.s, tc.exp, got)
			}
		})
	}
}

func TestZeroValue(t *testing.T) {
	var s set.Set[int]
	checkSet(t, "zero", s)
	if s.Contains(1) {
		t.Error("zero set contains 1")
	}
	if _, ok := s.Pop(); ok {
		t.Error("Pop on zero set reports a value")
	}
	s.Delete(1)
	s.DeleteAll()
	s.Shrink()
	checkSet(t, "zero after no-ops", s)
	checkSet(t, "zero clone", s.Clone())
	if !s.Equal(set.New[int]()) || !s.IsSubset(set.Set[int]{}) || s.Intersects(s) {
		t.Error("zero set relations")
	}
	s.Insert(4)
	checkSet(t, "zero after Insert", s, 4)
}

func TestInsertDelete(t *testing.T) {
	var s set.Set[int]
	s.Insert(3)
	s.Insert(1)
	s.Insert(3)
	checkSet(t, "Insert", s, 1, 3)
	s.Delete(2)
	checkSet(t, "Delete missing", s, 1, 3)
	s.Delete(1)
	checkSet(t, "Delete", s, 3)
	s.DeleteAll()
	checkSet(t, "DeleteAll", s)
	s.Insert(5)
	checkSet(t, "Insert after DeleteAll", s, 5)
}

func TestCollectInsertSeq(t *testing.T) {
	s := set.Collect(slices.Values([]int{3, 1, 3}))
	checkSet(t, "Collect", s, 1, 3)
	s.InsertSeq(slices.Values([]int{2, 3, 9}))
	checkSet(t, "InsertSeq", s, 1, 2, 3, 9)

	var z set.Set[int]
	z.InsertSeq(slices.Values([]int(nil)))
	checkSet(t, "InsertSeq empty", z)
	checkSet(t, "Collect empty", set.Collect(slices.Values([]int(nil))))
}

func TestPop(t *testing.T) {
	s := set.New(1, 2, 3)
	var got []int
	for {
		v, ok := s.Pop()
		if !ok {
			break
		}
		got = append(got, v)
	}
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("popped values exp [1 2 3], got %v", got)
	}
	checkSet(t, "after Pop", s)
	if _, ok := s.Pop(); ok {
		t.Error("Pop on empty set reports a value")
	}
}

func TestValuesBreak(t *testing.T) {
	n := 0
	for range set.New(1, 2, 3).Values() {
		n++
		break
	}
	if n != 1 {
		t.Errorf("iterations exp 1, got %d", n)
	}
}

// ----- String and WriteTo (order of values is unspecified)

func TestString(t *testing.T) {
	var zero set.Set[int]
	if got := zero.String(); got != "" {
		t.Error("zero string got:", got)
	}
	if got := set.New[int]().String(); got != "" {
		t.Error("empty string got:", got)
	}
	if got := set.New(3).String(); got != "3" {
		t.Error("{3} string got:", got)
	}

	parts := strings.Split(set.New(10, 2, 7).String(), " ")
	slices.Sort(parts)
	if exp := []string{"10", "2", "7"}; !slices.Equal(parts, exp) {
		t.Errorf("ints exp %v, got %v", exp, parts)
	}
	parts = strings.Split(set.New("b", "a", "c").String(), " ")
	slices.Sort(parts)
	if exp := []string{"a", "b", "c"}; !slices.Equal(parts, exp) {
		t.Errorf("strings exp %v, got %v", exp, parts)
	}
}

type myInt int

type stringer struct{ n int }

func (s stringer) String() string { return fmt.Sprintf("<%d>", s.n) }

// checkString verifies the output of String for a set built from vals.
// The order of values is unspecified, so the parts are compared sorted.
// The expected parts must not contain spaces.
func checkString[V comparable](t *testing.T, name string, vals []V, exp ...string) {
	t.Helper()
	got := set.New(vals...).String()
	if len(exp) == 0 {
		if got != "" {
			t.Errorf("%s: exp empty string, got %q", name, got)
		}
		return
	}
	parts := strings.Split(got, " ")
	slices.Sort(parts)
	slices.Sort(exp)
	if !slices.Equal(parts, exp) {
		t.Errorf("%s: exp %v, got %v (string %q)", name, exp, parts, got)
	}
}

// TestStringValueTypes covers every case of appendValue.
func TestStringValueTypes(t *testing.T) {
	// case int
	checkString(t, "int", []int{-5, 0, 42}, "-5", "0", "42")

	// case uint, including the largest value that does not fit into int
	maxUint := strconv.FormatUint(uint64(math.MaxUint), 10)
	checkString(t, "uint", []uint{7, 0, math.MaxUint}, "7", "0", maxUint)

	// case string
	checkString(t, "string", []string{"b", "", "a"}, "a", "b", "")

	// default: other predeclared types
	checkString(t, "int64", []int64{-9, 10}, "-9", "10")
	checkString(t, "uint8", []uint8{3, 200}, "3", "200")
	checkString(t, "float64", []float64{1.5, -2}, "1.5", "-2")
	checkString(t, "bool", []bool{true, false}, "true", "false")

	// default: named type with underlying type int does not match case int
	checkString(t, "named int", []myInt{3, -4}, "3", "-4")

	// default: fmt uses the String method
	checkString(t, "stringer", []stringer{{1}, {22}}, "<1>", "<22>")

	// default: struct without a space in its formatted form
	checkString(t, "struct", []struct{ n int }{{1}, {2}}, "{1}", "{2}")
}

// TestStringLongValues verifies that values longer than the initial
// buffer of WriteTo are written completely and do not corrupt later values.
func TestStringLongValues(t *testing.T) {
	long := strings.Repeat("x", 100)
	mid := strings.Repeat("y", 40)
	checkString(t, "long strings", []string{long, "s", mid}, long, "s", mid)
}

var errFail = errors.New("write failed")

// countWriter counts Write calls; the call number failAt (1-based) fails.
type countWriter struct {
	writes, failAt int
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		return 0, errFail
	}
	return len(p), nil
}

func TestWriteTo(t *testing.T) {
	s := set.New(1, 2, 3)

	var buf bytes.Buffer
	n, err := s.WriteTo(&buf)
	if err != nil || n != 5 || buf.Len() != 5 {
		t.Errorf("WriteTo: n=%d, len=%d, err=%v", n, buf.Len(), err)
	}

	cw := &countWriter{}
	if _, err = s.WriteTo(cw); err != nil || cw.writes != s.Count() {
		t.Errorf("exp one Write per value (%d), got %d, err=%v", s.Count(), cw.writes, err)
	}

	var empty set.Set[int]
	cw = &countWriter{}
	if n, err = empty.WriteTo(cw); n != 0 || err != nil || cw.writes != 0 {
		t.Errorf("empty: n=%d, writes=%d, err=%v", n, cw.writes, err)
	}

	cw = &countWriter{failAt: 2}
	n, err = s.WriteTo(cw)
	if !errors.Is(err, errFail) || n != 1 || cw.writes != 2 {
		t.Errorf("failing writer: n=%d, writes=%d, err=%v", n, cw.writes, err)
	}
}

// ----- Binary operations and relations

type opCase struct {
	name         string
	a, b         []int
	union, inter []int
	onlyA, onlyB []int // a \ b and b \ a
}

var opCases = []opCase{
	{name: "both-empty"},
	{name: "a-empty", b: []int{1, 2}, union: []int{1, 2}, onlyB: []int{1, 2}},
	{name: "b-empty", a: []int{1, 2}, union: []int{1, 2}, onlyA: []int{1, 2}},
	{name: "equal", a: []int{1, 2, 3}, b: []int{3, 2, 1}, union: []int{1, 2, 3}, inter: []int{1, 2, 3}},
	{name: "disjoint", a: []int{1, 2}, b: []int{3, 4}, union: []int{1, 2, 3, 4}, onlyA: []int{1, 2}, onlyB: []int{3, 4}},
	{name: "overlap", a: []int{1, 2, 3}, b: []int{3, 4, 5}, union: []int{1, 2, 3, 4, 5}, inter: []int{3}, onlyA: []int{1, 2}, onlyB: []int{4, 5}},
	{name: "a-subset", a: []int{1, 2}, b: []int{1, 2, 3}, union: []int{1, 2, 3}, inter: []int{1, 2}, onlyB: []int{3}},
	{name: "b-subset", a: []int{1, 2, 3}, b: []int{1, 2}, union: []int{1, 2, 3}, inter: []int{1, 2}, onlyA: []int{3}},
	{name: "different-sizes", a: []int{5}, b: ints(7), union: ints(7), inter: []int{5}, onlyB: []int{0, 1, 2, 3, 4, 6}},
}

// checkOps verifies all binary operations and relations for one case.
func checkOps(t *testing.T, tc opCase) {
	t.Helper()
	a, b := set.New(tc.a...), set.New(tc.b...)
	symdiff := slices.Concat(tc.onlyA, tc.onlyB)

	// Non-mutating operations, in both operand orders where symmetric.
	checkSet(t, "Union", a.Union(b), tc.union...)
	checkSet(t, "Union reversed", b.Union(a), tc.union...)
	checkSet(t, "Intersection", a.Intersection(b), tc.inter...)
	checkSet(t, "Intersection reversed", b.Intersection(a), tc.inter...)
	checkSet(t, "Difference", a.Difference(b), tc.onlyA...)
	checkSet(t, "Difference reversed", b.Difference(a), tc.onlyB...)
	checkSet(t, "SymmetricDifference", a.SymmetricDifference(b), symdiff...)
	checkSet(t, "SymmetricDifference reversed", b.SymmetricDifference(a), symdiff...)
	onlyA, onlyB := a.Delta(b)
	checkSet(t, "Delta onlyA", onlyA, tc.onlyA...)
	checkSet(t, "Delta onlyB", onlyB, tc.onlyB...)
	onlyB, onlyA = b.Delta(a)
	checkSet(t, "Delta reversed onlyA", onlyA, tc.onlyA...)
	checkSet(t, "Delta reversed onlyB", onlyB, tc.onlyB...)
	checkSet(t, "a unchanged", a, tc.a...)
	checkSet(t, "b unchanged", b, tc.b...)

	// Mutating operations, on a clone of a and on a zero value.
	mutating := []struct {
		name   string
		op     func(*set.Set[int], set.Set[int])
		onA    []int
		onZero []int
	}{
		{"Or", (*set.Set[int]).Or, tc.union, tc.b},
		{"And", (*set.Set[int]).And, tc.inter, nil},
		{"AndNot", (*set.Set[int]).AndNot, tc.onlyA, nil},
		{"Xor", (*set.Set[int]).Xor, symdiff, tc.b},
	}
	for _, m := range mutating {
		s := a.Clone()
		m.op(&s, b)
		checkSet(t, m.name, s, m.onA...)

		var z set.Set[int]
		m.op(&z, b)
		checkSet(t, m.name+" on zero value", z, m.onZero...)
		z.Insert(99) // must not leak into b
		checkSet(t, m.name+": b unchanged", b, tc.b...)
	}
	checkSet(t, "a unchanged after mutating ops", a, tc.a...)

	// Relations.
	noA, noB := len(tc.onlyA) == 0, len(tc.onlyB) == 0
	if got := a.Equal(b); got != (noA && noB) {
		t.Errorf("Equal exp %v, got %v", noA && noB, got)
	}
	if got := b.Equal(a); got != (noA && noB) {
		t.Errorf("Equal reversed exp %v, got %v", noA && noB, got)
	}
	if got := a.IsSubset(b); got != noA {
		t.Errorf("IsSubset exp %v, got %v", noA, got)
	}
	if got := b.IsSubset(a); got != noB {
		t.Errorf("IsSubset reversed exp %v, got %v", noB, got)
	}
	inter := len(tc.inter) > 0
	if got := a.Intersects(b); got != inter {
		t.Errorf("Intersects exp %v, got %v", inter, got)
	}
	if got := b.Intersects(a); got != inter {
		t.Errorf("Intersects reversed exp %v, got %v", inter, got)
	}
}

func TestBinaryOps(t *testing.T) {
	for _, tc := range opCases {
		t.Run(tc.name, func(t *testing.T) { checkOps(t, tc) })
	}
}

// TestRandomAgainstModel compares all operations with a trivial reference
// model (membership via slices.Contains) on random sets.
func TestRandomAgainstModel(t *testing.T) {
	const universe = 16
	rnd := rand.New(rand.NewPCG(1, 2))
	gen := func() []int {
		vals := make([]int, rnd.IntN(12))
		for i := range vals {
			vals[i] = rnd.IntN(universe)
		}
		return vals
	}
	for range 500 {
		tc := opCase{name: "random", a: gen(), b: gen()}
		for v := range universe {
			inA, inB := slices.Contains(tc.a, v), slices.Contains(tc.b, v)
			if inA || inB {
				tc.union = append(tc.union, v)
			}
			switch {
			case inA && inB:
				tc.inter = append(tc.inter, v)
			case inA:
				tc.onlyA = append(tc.onlyA, v)
			case inB:
				tc.onlyB = append(tc.onlyB, v)
			}
		}
		checkOps(t, tc)
		if t.Failed() {
			t.Fatalf("failing input: a=%v, b=%v", tc.a, tc.b)
		}
	}
}

func TestSelfOperands(t *testing.T) {
	s := set.New(1, 2, 3)
	s.Or(s)
	checkSet(t, "Or self", s, 1, 2, 3)
	s.And(s)
	checkSet(t, "And self", s, 1, 2, 3)
	s.AndNot(s)
	checkSet(t, "AndNot self", s)

	s = set.New(1, 2, 3)
	s.Xor(s)
	checkSet(t, "Xor self", s)

	a := set.New(1, 2, 3)
	checkSet(t, "Union self", a.Union(a), 1, 2, 3)
	checkSet(t, "Intersection self", a.Intersection(a), 1, 2, 3)
	checkSet(t, "Difference self", a.Difference(a))
	checkSet(t, "SymmetricDifference self", a.SymmetricDifference(a))
	onlyA, onlyB := a.Delta(a)
	checkSet(t, "Delta self onlyA", onlyA)
	checkSet(t, "Delta self onlyB", onlyB)
	if !a.Equal(self(a)) || !a.IsSubset(self(a)) || !a.Intersects(self(a)) {
		t.Error("relations with self")
	}
	checkSet(t, "a unchanged", a, 1, 2, 3)
}

// self returns r unchanged. It marks an intentional self-application such as
// a.IsSubset(self(a)) and keeps linters from reporting it as a typo.
func self[T any](v T) T { return v }

// TestResultsIndependent verifies that results share no storage with operands.
func TestResultsIndependent(t *testing.T) {
	a, b := set.New(1, 2, 3), set.New(3, 4)
	onlyA, onlyB := a.Delta(b)
	results := []set.Set[int]{
		a.Union(b), a.Intersection(b), a.Difference(b), a.SymmetricDifference(b),
		a.Union(set.Set[int]{}), set.Set[int]{}.Union(b), a.Clone(), onlyA, onlyB,
	}
	for i := range results {
		results[i].Insert(99)
		results[i].Delete(3)
	}
	checkSet(t, "a", a, 1, 2, 3)
	checkSet(t, "b", b, 3, 4)
}

// ----- Memory management

func TestClone(t *testing.T) {
	a := set.New(1, 2, 3)
	c := a.Clone()
	c.Insert(4)
	c.Delete(1)
	checkSet(t, "original", a, 1, 2, 3)
	checkSet(t, "clone", c, 2, 3, 4)

	e := set.New(1)
	e.Delete(1) // empty, but allocated
	c = e.Clone()
	c.Insert(5)
	checkSet(t, "empty original", e)
	checkSet(t, "clone of empty", c, 5)
}

func TestGrowShrink(t *testing.T) {
	var s set.Set[int]
	s.Grow(10)
	checkSet(t, "Grow empty", s)
	for i := range 10 {
		s.Insert(i)
	}
	checkSet(t, "after inserts", s, ints(10)...)
	s.Grow(100)
	checkSet(t, "Grow keeps values", s, ints(10)...)
	s.Grow(0)
	checkSet(t, "Grow(0) keeps values", s, ints(10)...)

	for i := 2; i < 10; i++ {
		s.Delete(i)
	}
	s.Shrink()
	checkSet(t, "Shrink keeps values", s, 0, 1)

	s.DeleteAll()
	s.Shrink()
	checkSet(t, "Shrink empty", s)
	s.Insert(7)
	checkSet(t, "Insert after Shrink", s, 7)
}

func TestGrowNegative(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Grow(-1) did not panic")
		}
	}()
	var s set.Set[int]
	s.Grow(-1)
}

// ----- Floating-point semantics (documented in package comment)

func TestFloat(t *testing.T) {
	s := set.New(0.0, math.Copysign(0, -1))
	if got := s.Count(); got != 1 {
		t.Errorf("0.0 and -0.0 exp one element, got %d", got)
	}

	nan := math.NaN()
	var n set.Set[float64]
	n.Insert(nan)
	n.Insert(nan)
	if n.Contains(nan) {
		t.Error("NaN is contained, but NaN != NaN")
	}
	if got := n.Count(); got != 2 {
		t.Errorf("two NaN inserts exp two elements, got %d", got)
	}
}

// ----- Examples

func ExampleSet_Delta() {
	oldWords := set.New("a", "b", "c")
	newWords := set.New("b", "c", "d")
	removed, added := oldWords.Delta(newWords)
	fmt.Println(slices.Sorted(removed.Values()), slices.Sorted(added.Values()))
	// Output:
	// [a] [d]
}
