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

// Package set provides a simple set type.
package set

import (
	"fmt"
	"iter"
	"strings"
)

// Set is an unordered collection of non-duplicate elements.
type Set[E comparable] struct {
	m map[E]struct{}
}

// New creates a new set with the given elements.
func New[E comparable](elems ...E) *Set[E] {
	m := make(map[E]struct{}, min(3, len(elems)))
	for _, elem := range elems {
		m[elem] = struct{}{}
	}
	return &Set[E]{m}
}

// String returns a string representation.
func (s *Set[E]) String() string {
	var sb strings.Builder
	sb.WriteByte('{')
	if s != nil && s.m != nil {
		comma := false
		for elem := range s.m {
			if comma {
				sb.WriteString(", ")
			}
			comma = true
			sb.WriteString(fmt.Sprintf("%v", elem))
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

// Add an elements to the set.
func (s *Set[E]) Add(elem E) *Set[E] {
	s = s.ensure()
	s.m[elem] = struct{}{}
	return s
}

// Contains returns true, if the set contains the element.
func (s *Set[E]) Contains(elem E) bool {
	if s != nil && s.m != nil {
		_, ok := s.m[elem]
		return ok
	}
	return false
}

// Length returns the number of elements in the set.
func (s *Set[E]) Length() int {
	if s != nil {
		return len(s.m)
	}
	return 0
}

// Values returns an iterator of all elements of the set.
func (s *Set[E]) Values() iter.Seq[E] {
	return func(yield func(E) bool) {
		if s != nil && s.m != nil {
			for elem := range s.m {
				if !yield(elem) {
					return
				}
			}
		}
	}
}

// Remove an element from the set.
func (s *Set[E]) Remove(elem E) *Set[E] {
	if s != nil && s.m != nil {
		delete(s.m, elem)
	}
	return s
}

// ensure a valid zero value.
func (s *Set[E]) ensure() *Set[E] {
	if s == nil {
		return New[E]()
	}
	if s.m == nil {
		s.m = map[E]struct{}{}
	}
	return s
}
