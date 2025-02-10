//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of App.
//
// App is licensed under the latest version of the EUPL (European Union Public
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

	appiter "t73f.de/r/app/iter"
)

func TestCatSeq(t *testing.T) {
	if exp, got := []int{}, slices.Collect(appiter.CatSeq[int]()); !slices.Equal(exp, got) {
		t.Error(got)
	}
	ones := []int{1}
	if exp, got := ones, slices.Collect(appiter.CatSeq(slices.Values(ones))); !slices.Equal(exp, got) {
		t.Error(got)
	}
	if exp, got := []int{1, 1}, slices.Collect(appiter.CatSeq(slices.Values(ones), slices.Values(ones))); !slices.Equal(exp, got) {
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
			got := slices.Collect(appiter.MapSeq(slices.Values(tc.inp), strconv.Itoa))
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
	got := slices.Collect(appiter.FilterSeq(slices.Values(nums), isPrime))
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

func TestMapReduce(t *testing.T) {
	sData := []string{"1", "2", "3", "4", "5", "6"}
	intSeq := appiter.MapSeq(slices.Values(sData), func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		return i
	})
	sum := appiter.ReduceSeq(intSeq, 0, func(x, y int) int { return x + y })
	if sum != 21 {
		t.Error(sum)
	}
	prod := appiter.ReduceSeq(intSeq, 1, func(x, y int) int { return x * y })
	if prod != 720 {
		t.Error(prod)
	}
}

func TestTakeSeq(t *testing.T) {
	if exp, got := []int{}, slices.Collect(appiter.TakeSeq(0, appiter.CountSeq())); !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
	if exp, got := []int{0}, slices.Collect(appiter.TakeSeq(1, appiter.CountSeq())); !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
	if exp, got := []int{0, 1}, slices.Collect(appiter.TakeSeq(2, appiter.CountSeq())); !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
}

func TestKeySeq(t *testing.T) {
	sl := []string{"a", "b", "c", "d"}
	zipseq := appiter.ZipSeq(appiter.CountSeq(), slices.Values(sl))
	got := slices.Collect(appiter.KeySeq(zipseq))
	exp := []int{0, 1, 2, 3}
	if !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
}

func TestValSeq(t *testing.T) {
	sl := []string{"a", "b", "c", "d"}
	zipseq := appiter.ZipSeq(appiter.CountSeq(), slices.Values(sl))
	got := slices.Collect(appiter.ValSeq(zipseq))
	exp := sl
	if !slices.Equal(exp, got) {
		t.Error(exp, got)
	}
}
