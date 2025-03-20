//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

// Package graph implements a (directed) graph of orderable values.
package graph

import (
	"cmp"
	"maps"
	"slices"

	"t73f.de/r/zero/set"
)

// Digraph relates orderable values in a directional way.
type Digraph[T cmp.Ordered] map[T]*set.Set[T]

// AddVertex adds an edge / vertex to the digraph.
func (dg Digraph[T]) AddVertex(v T) Digraph[T] {
	if dg == nil {
		return Digraph[T]{v: nil}
	}
	if _, found := dg[v]; !found {
		dg[v] = nil
	}
	return dg
}

// RemoveVertex removes a vertex and all its edges from the digraph.
func (dg Digraph[T]) RemoveVertex(v T) {
	if len(dg) > 0 {
		delete(dg, v)
		for vertex, closure := range dg {
			dg[vertex] = closure.Remove(v)
		}
	}
}

// AddEdge adds a connection from `from` to `to`.
// Both vertices must be added before. Otherwise the function may panic.
func (dg Digraph[T]) AddEdge(from, to T) Digraph[T] {
	if dg == nil {
		return Digraph[T]{from: set.New(to), to: nil}
	}
	dg[from] = dg[from].Add(to)
	return dg
}

// AddEgdes adds all given `Edge`s to the digraph.
//
// In contrast to `AddEdge` the vertices must not exist before.
func (dg Digraph[T]) AddEgdes(edges EdgeSlice[T]) Digraph[T] {
	if dg == nil {
		if len(edges) == 0 {
			return nil
		}
		dg = make(Digraph[T], len(edges))
	}
	for _, edge := range edges {
		dg = dg.AddVertex(edge.From)
		dg = dg.AddVertex(edge.To)
		dg = dg.AddEdge(edge.From, edge.To)
	}
	return dg
}

// Equal returns true if both digraphs have the same vertices and edges.
func (dg Digraph[T]) Equal(other Digraph[T]) bool {
	return maps.EqualFunc(dg, other, func(cg, co *set.Set[T]) bool { return cg.Equal(co) })
}

// Clone a digraph.
func (dg Digraph[T]) Clone() Digraph[T] {
	if len(dg) == 0 {
		return nil
	}
	copyDG := make(Digraph[T], len(dg))
	for vertex, closure := range dg {
		copyDG[vertex] = closure.Clone()
	}
	return copyDG
}

// HasVertex returns true, if `v` is a vertex of the digraph.
func (dg Digraph[T]) HasVertex(v T) bool {
	if len(dg) == 0 {
		return false
	}
	_, found := dg[v]
	return found
}

// Vertices returns the set of all vertices.
func (dg Digraph[T]) Vertices() *set.Set[T] {
	if len(dg) == 0 {
		return nil
	}
	verts := set.New[T]()
	for vert := range dg {
		verts.Add(vert)
	}
	return verts
}

// Edges returns an unsorted slice of the edges of the digraph.
func (dg Digraph[T]) Edges() (es EdgeSlice[T]) {
	for vert, closure := range dg {
		for next := range closure.Values() {
			es = append(es, Edge[T]{From: vert, To: next})
		}
	}
	return es
}

// Originators will return the set of all vertices that are not referenced
// at the to-part of an edge.
func (dg Digraph[T]) Originators() *set.Set[T] {
	if len(dg) == 0 {
		return nil
	}
	origs := dg.Vertices()
	for _, closure := range dg {
		for c := range closure.Values() {
			origs.Remove(c)
		}
	}
	return origs
}

// Terminators returns the set of all vertices that does not reference
// other vertices.
func (dg Digraph[T]) Terminators() (terms *set.Set[T]) {
	for vert, closure := range dg {
		if closure.Length() == 0 {
			terms = terms.Add(vert)
		}
	}
	return terms
}

// TransitiveClosure calculates the sub-graph that is reachable from `v`.
func (dg Digraph[T]) TransitiveClosure(v T) (tc Digraph[T]) {
	if len(dg) == 0 {
		return nil
	}
	var marked *set.Set[T]
	stack := []T{v}
	for pos := len(stack) - 1; pos >= 0; pos = len(stack) - 1 {
		curr := stack[pos]
		stack = stack[:pos]
		if marked.Contains(curr) {
			continue
		}
		tc = tc.AddVertex(curr)
		for next := range dg[curr].Values() {
			tc = tc.AddVertex(next)
			tc = tc.AddEdge(curr, next)
			stack = append(stack, next)
		}
		marked = marked.Add(curr)
	}
	return tc
}

// ReachableVertices calculates the set of all vertices that are reachable
// from the given vertex `startV`.
func (dg Digraph[T]) ReachableVertices(startV T) (tc *set.Set[T]) {
	if len(dg) == 0 {
		return nil
	}
	stack := slices.Collect(dg[startV].Values())
	for last := len(stack) - 1; last >= 0; last = len(stack) - 1 {
		curr := stack[last]
		stack = stack[:last]
		if tc.Contains(curr) {
			continue
		}
		closure, found := dg[curr]
		if !found {
			continue
		}
		tc = tc.Add(curr)
		for next := range closure.Values() {
			stack = append(stack, next)
		}
	}
	return tc
}

// IsDAG returns a vertex and false, if the graph has a cycle containing the vertex.
func (dg Digraph[T]) IsDAG() (T, bool) {
	for vertex := range dg {
		if dg.ReachableVertices(vertex).Contains(vertex) {
			return vertex, false
		}
	}
	var zeroT T
	return zeroT, true
}

// Reverse returns a graph with reversed edges.
func (dg Digraph[T]) Reverse() (revDg Digraph[T]) {
	for vertex, closure := range dg {
		revDg = revDg.AddVertex(vertex)
		for next := range closure.Values() {
			revDg = revDg.AddVertex(next)
			revDg = revDg.AddEdge(next, vertex)
		}
	}
	return revDg
}

// SortReverse returns a deterministic, topological, reverse sort of the digraph.
//
// Works only if digraph is a DAG. Otherwise the algorithm will not terminate
// or returns an arbitrary value.
func (dg Digraph[T]) SortReverse() (sl []T) {
	if len(dg) == 0 {
		return nil
	}
	tempDg := dg.Clone()
	for len(tempDg) > 0 {
		terms := tempDg.Terminators()
		if terms.Length() == 0 {
			break
		}
		termSlice := slices.Sorted(terms.Values())
		slices.Reverse(termSlice)
		sl = append(sl, termSlice...)
		for t := range terms.Values() {
			tempDg.RemoveVertex(t)
		}
	}
	return sl
}
