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

package iter_test

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"testing"

	zeroiter "t73f.de/r/zero/iter"
)

func TestEmptySeq(t *testing.T) {
	if got := slices.Collect(zeroiter.EmptySeq[int]()); len(got) != 0 {
		t.Error("EmptySeq is not empty:", got)
	}
}

func TestOneSeq(t *testing.T) {
	exp, got := []int{5}, slices.Collect(zeroiter.OneSeq(5))
	if !slices.Equal(exp, got) {
		t.Error("exp:", exp, "got:", got)
	}
}

func TestCatSeq(t *testing.T) {
	ones := []int{1}
	testdata := []struct {
		name string
		seqs []iter.Seq[int]
		exp  []int
	}{
		{"none", nil, nil},
		{"one", []iter.Seq[int]{slices.Values(ones)}, ones},
		{"two", []iter.Seq[int]{slices.Values(ones), slices.Values([]int{2, 3})}, []int{1, 2, 3}},
		{"empty-between", []iter.Seq[int]{slices.Values(ones), zeroiter.EmptySeq[int](), slices.Values(ones)}, []int{1, 1}},
	}
	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			if got := slices.Collect(zeroiter.CatSeq(tc.seqs...)); !slices.Equal(tc.exp, got) {
				t.Errorf("exp: %v, got: %v", tc.exp, got)
			}
		})
	}
}

// A consumer that stops early must not be called again (would panic).
func TestCatSeqBreak(t *testing.T) {
	vals := slices.Values([]int{1, 2})
	n := 0
	for range zeroiter.CatSeq(vals, vals) {
		n++
		break
	}
	if n != 1 {
		t.Error("expected exactly one iteration, got:", n)
	}
}

func TestMapSeq(t *testing.T) {
	testdata := []struct {
		name string
		inp  []int
		exp  []string
	}{
		{"nil", nil, nil},
		{"one", []int{1}, []string{"1"}},
		{"two", []int{1, 2}, []string{"1", "2"}},
		{"three", []int{1, 2, 3}, []string{"1", "2", "3"}},
	}

	for i, tc := range testdata {
		t.Run(fmt.Sprintf("%s-%d", tc.name, i), func(t *testing.T) {
			got := slices.Collect(zeroiter.MapSeq(slices.Values(tc.inp), strconv.Itoa))
			if !slices.Equal(tc.exp, got) {
				t.Errorf("exp: %v, got: %v", tc.exp, got)
			}
		})
	}
}

func TestFilterSeq(t *testing.T) {
	nums := make([]int, 50)
	for i := range nums {
		nums[i] = i
	}
	exp := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}
	got := slices.Collect(zeroiter.FilterSeq(slices.Values(nums), isPrime))
	if !slices.Equal(exp, got) {
		t.Errorf("exp: %v, got: %v", exp, got)
	}
}

func isPrime(i int) bool {
	if i < 2 {
		return false
	}
	if i < 4 {
		return true
	}
	if i%2 == 0 {
		return false
	}
	for factor := 3; factor*factor <= i; factor += 2 {
		if i%factor == 0 {
			return false
		}
	}
	return true
}

func TestMapFilterSeq(t *testing.T) {
	exp := []int{0, 6, 12, 18, 24, 30, 36}
	got := slices.Collect(zeroiter.MapFilterSeq(
		zeroiter.TakeSeq(20, zeroiter.CountSeq()),
		func(val int) (int, bool) {
			if val%3 == 0 {
				return val * 2, true
			}
			return -1, false
		}),
	)
	if !slices.Equal(exp, got) {
		t.Error(got)
	}
}

func TestReduceSeq(t *testing.T) {
	intSeq := zeroiter.MapSeq(slices.Values([]string{"1", "2", "3", "4", "5", "6"}),
		func(s string) int {
			i, err := strconv.Atoi(s)
			if err != nil {
				t.Fatal(err)
			}
			return i
		})
	if sum := zeroiter.ReduceSeq(intSeq, 0, func(x, y int) int { return x + y }); sum != 21 {
		t.Error("sum:", sum)
	}
	if prod := zeroiter.ReduceSeq(intSeq, 1, func(x, y int) int { return x * y }); prod != 720 {
		t.Error("prod:", prod)
	}
}

func TestDeduplicateSeq(t *testing.T) {
	inp := []int{0, 1, 0, 1, 2, 0, 1, 2, 3, 0, 1, 2, 3, 4, 0}
	exp := []int{0, 1, 2, 3, 4}
	if got := slices.Collect(zeroiter.DeduplicateSeq(slices.Values(inp))); !slices.Equal(exp, got) {
		t.Errorf("exp: %v, got: %v", exp, got)
	}
}

func TestCountSeq(t *testing.T) {
	exp := []int{0, 1, 2}
	for i, got := range slices.Collect(zeroiter.TakeSeq(3, zeroiter.CountSeq())) {
		if got != exp[i] {
			t.Errorf("index %d: exp: %d, got: %d", i, exp[i], got)
		}
	}
}

func TestTakeSeq(t *testing.T) {
	finite := []int{0, 1}
	testdata := []struct {
		name string
		num  int
		seq  iter.Seq[int]
		exp  []int
	}{
		{"negative", -1, zeroiter.CountSeq(), nil},
		{"zero", 0, zeroiter.CountSeq(), nil},
		{"one", 1, zeroiter.CountSeq(), []int{0}},
		{"two", 2, zeroiter.CountSeq(), []int{0, 1}},
		{"more-than-available", 5, slices.Values(finite), finite},
	}
	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			if got := slices.Collect(zeroiter.TakeSeq(tc.num, tc.seq)); !slices.Equal(tc.exp, got) {
				t.Errorf("exp: %v, got: %v", tc.exp, got)
			}
		})
	}
}

// TakeSeq must not request more elements from its source than needed.
func TestTakeSeqNoOverpull(t *testing.T) {
	produced := 0
	src := func(yield func(int) bool) {
		for i := 0; ; i++ {
			produced++
			if !yield(i) {
				return
			}
		}
	}
	_ = slices.Collect(zeroiter.TakeSeq(3, src))
	if produced != 3 {
		t.Error("source produced elements:", produced)
	}
}

func TestZipSeq(t *testing.T) {
	collect := func(k iter.Seq[int], v iter.Seq[string]) []string {
		var res []string
		for key, val := range zeroiter.ZipSeq(k, v) {
			res = append(res, fmt.Sprintf("%d=%s", key, val))
		}
		return res
	}
	testdata := []struct {
		name string
		k    []int
		v    []string
		exp  []string
	}{
		{"both-empty", nil, nil, nil},
		{"k-empty", nil, []string{"a"}, nil},
		{"v-empty", []int{1}, nil, nil},
		{"equal", []int{1, 2}, []string{"a", "b"}, []string{"1=a", "2=b"}},
		{"k-shorter", []int{1}, []string{"a", "b"}, []string{"1=a"}},
		{"v-shorter", []int{1, 2}, []string{"a"}, []string{"1=a"}},
	}
	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			got := collect(slices.Values(tc.k), slices.Values(tc.v))
			if !slices.Equal(tc.exp, got) {
				t.Errorf("exp: %v, got: %v", tc.exp, got)
			}
		})
	}
}

// Both sources must be terminated, also if the shorter one ends first.
func TestZipSeqStopsSources(t *testing.T) {
	stopped := false
	infinite := func(yield func(int) bool) {
		defer func() { stopped = true }()
		for i := 0; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
	_ = slices.Collect(zeroiter.KeySeq(zeroiter.ZipSeq(infinite, slices.Values([]string{"a"}))))
	if !stopped {
		t.Error("infinite source was not stopped")
	}
}

func TestKeySeq(t *testing.T) {
	sl := []string{"a", "b", "c", "d"}
	zipseq := zeroiter.ZipSeq(zeroiter.CountSeq(), slices.Values(sl))
	got := slices.Collect(zeroiter.KeySeq(zipseq))
	exp := []int{0, 1, 2, 3}
	if !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
}

func TestValSeq(t *testing.T) {
	sl := []string{"a", "b", "c", "d"}
	zipseq := zeroiter.ZipSeq(zeroiter.CountSeq(), slices.Values(sl))
	got := slices.Collect(zeroiter.ValSeq(zipseq))
	exp := sl
	if !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
}

func TestEnumerateSeq(t *testing.T) {
	cnt := 0
	for i, j := range zeroiter.EnumerateSeq(zeroiter.CountSeq()) {
		if i != j {
			t.Fatalf("enumerateCount fails, i=%d, j=%d", i, j)
		}
		cnt++
		if cnt > 1000 {
			break
		}
	}
}
