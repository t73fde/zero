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
	"slices"
	"strconv"
	"testing"

	zeroiter "t73f.de/r/zero/iter"
)

func TestEmptySeq(t *testing.T) {
	count := 0
	for i := range zeroiter.EmptySeq[int]() {
		count += count + i + 1
	}
	if count > 0 {
		t.Error("EmptySeq is not empty: ", count)
	}
}

func TestCatSeq(t *testing.T) {
	if exp, got := []int{}, slices.Collect(zeroiter.CatSeq[int]()); !slices.Equal(exp, got) {
		t.Error(got)
	}
	ones := []int{1}
	if exp, got := ones, slices.Collect(zeroiter.CatSeq(slices.Values(ones))); !slices.Equal(exp, got) {
		t.Error(got)
	}
	if exp, got := []int{1, 1}, slices.Collect(zeroiter.CatSeq(slices.Values(ones), slices.Values(ones))); !slices.Equal(exp, got) {
		t.Error(exp, got)
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
	exp := []int{1, 2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}
	got := slices.Collect(zeroiter.FilterSeq(slices.Values(nums), isPrime))
	if !slices.Equal(exp, got) {
		t.Error(got)
	}
}
func isPrime(i int) bool {
	if i <= 0 {
		return false
	}
	if i <= 3 {
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

func TestMapReduce(t *testing.T) {
	sData := []string{"1", "2", "3", "4", "5", "6"}
	intSeq := zeroiter.MapSeq(slices.Values(sData), func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		return i
	})
	sum := zeroiter.ReduceSeq(intSeq, 0, func(x, y int) int { return x + y })
	if sum != 21 {
		t.Error(sum)
	}
	prod := zeroiter.ReduceSeq(intSeq, 1, func(x, y int) int { return x * y })
	if prod != 720 {
		t.Error(prod)
	}
}

func TestTakeSeq(t *testing.T) {
	if exp, got := []int{}, slices.Collect(zeroiter.TakeSeq(0, zeroiter.CountSeq())); !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
	if exp, got := []int{0}, slices.Collect(zeroiter.TakeSeq(1, zeroiter.CountSeq())); !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
	if exp, got := []int{0, 1}, slices.Collect(zeroiter.TakeSeq(2, zeroiter.CountSeq())); !slices.Equal(exp, got) {
		t.Error(exp, got)
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
