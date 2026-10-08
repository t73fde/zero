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
//
// Conventions: Every vertex is a key of the map, including vertices without
// outgoing edges. Methods returning a digraph return nil for an empty result.
// Methods returning a set always return a non-nil set.
type Digraph[T cmp.Ordered] map[T]set.Set[T]

// AddVertex adds a vertex to the digraph.
func (dg Digraph[T]) AddVertex(v T) Digraph[T] {
	if dg == nil {
		return Digraph[T]{v: set.Set[T]{}}
	}
	if _, found := dg[v]; !found {
		dg[v] = set.Set[T]{}
	}
	return dg
}

// RemoveVertex removes a vertex and all its edges from the digraph.
func (dg Digraph[T]) RemoveVertex(v T) {
	if len(dg) > 0 {
		delete(dg, v)
		for vertex, closure := range dg {
			closure.Delete(v)
			dg[vertex] = closure
		}
	}
}

// AddEdge adds a connection from `from` to `to`.
// Both vertices must be added before. Otherwise the function may panic.
func (dg Digraph[T]) AddEdge(from, to T) Digraph[T] {
	dg = dg.AddVertex(from).AddVertex(to)
	fromSet := dg[from]
	fromSet.Insert(to)
	dg[from] = fromSet
	return dg
}

// AddEdges adds all given `Edge`s to the digraph.
//
// In contrast to `AddEdge` the vertices must not exist before.
func (dg Digraph[T]) AddEdges(edges EdgeSlice[T]) Digraph[T] {
	if dg == nil {
		if len(edges) == 0 {
			return nil
		}
		dg = make(Digraph[T], len(edges))
	}
	for _, edge := range edges {
		dg = dg.AddEdge(edge.From, edge.To)
	}
	return dg
}

// Equal returns true if both digraphs have the same vertices and edges.
func (dg Digraph[T]) Equal(other Digraph[T]) bool {
	return maps.EqualFunc(dg, other, func(cg, co set.Set[T]) bool { return cg.Equal(co) })
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
	_, found := dg[v]
	return found
}

// Vertices returns the set of all vertices.
func (dg Digraph[T]) Vertices() (verts set.Set[T]) {
	for vert := range dg {
		verts.Insert(vert)
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
func (dg Digraph[T]) Originators() set.Set[T] {
	origs := dg.Vertices()
	for _, closure := range dg {
		for c := range closure.Values() {
			origs.Delete(c)
		}
	}
	return origs
}

// Terminators returns the set of all vertices that does not reference
// other vertices.
func (dg Digraph[T]) Terminators() (terms set.Set[T]) {
	for vert, closure := range dg {
		if closure.IsEmpty() {
			terms.Insert(vert)
		}
	}
	return terms
}

// TransitiveClosure calculates the sub-graph that is reachable from `v`.
//
// The result contains `v` and all vertices reachable from it, together with
// the edges between them. If `v` is not a vertex, nil is returned.
func (dg Digraph[T]) TransitiveClosure(v T) (tc Digraph[T]) {
	if !dg.HasVertex(v) {
		return nil
	}
	var marked set.Set[T]
	stack := []T{v}
	for len(stack) > 0 {
		last := len(stack) - 1
		curr := stack[last]
		stack = stack[:last]
		tc = tc.AddVertex(curr)
		for next := range dg[curr].Values() {
			tc = tc.AddEdge(curr, next)
			if !marked.Contains(next) {
				marked.Insert(next)
				stack = append(stack, next)
			}
		}
	}
	return tc
}

// ReachableVertices calculates the set of all vertices that are reachable
// via at least one edge from the given vertex `startV`.
//
// `startV` is part of the result only if it lies on a cycle.
func (dg Digraph[T]) ReachableVertices(startV T) (reached set.Set[T]) {
	if len(dg) == 0 {
		return set.Set[T]{}
	}
	stack := slices.Collect(dg[startV].Values())
	for len(stack) > 0 {
		last := len(stack) - 1
		curr := stack[last]
		stack = stack[:last]
		if reached.Contains(curr) {
			continue
		}
		reached.Insert(curr)
		for next := range dg[curr].Values() {
			stack = append(stack, next)
		}
	}
	return reached
}

// IsDAG returns a vertex and false, if the graph has a cycle containing the
// vertex. Otherwise it returns the zero value and true.
//
// Runs in O(V+E).
func (dg Digraph[T]) IsDAG() (T, bool) {
	const (
		unvisited int8 = iota
		active
		done
	)
	state := make(map[T]int8, len(dg))
	var cycleVertex T
	var visit func(T) bool
	visit = func(v T) bool {
		state[v] = active
		for next := range dg[v].Values() {
			switch state[next] {
			case active:
				cycleVertex = next
				return false
			case unvisited:
				if !visit(next) {
					return false
				}
			}
		}
		state[v] = done
		return true
	}
	for v := range dg {
		if state[v] == unvisited && !visit(v) {
			return cycleVertex, false
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
			revDg = revDg.AddEdge(next, vertex)
		}
	}
	return revDg
}

// SortReverse returns a deterministic, topological, reverse sort of the digraph.
//
// Vertices are emitted level by level: first all vertices without outgoing
// edges (descending), then all vertices whose successors have been emitted, etc.
//
// If the digraph contains a cycle, the vertices on a cycle and all vertices
// that reach a cycle are omitted. Check with `IsDAG` beforehand if a complete
// result is required.
//
// Runs in O(V+E) plus the cost of sorting each level.
func (dg Digraph[T]) SortReverse() (sl []T) {
	if len(dg) == 0 {
		return nil
	}
	outDegree := make(map[T]int, len(dg))
	var level []T
	for v, closure := range dg {
		n := closure.Count()
		outDegree[v] = n
		if n == 0 {
			level = append(level, v)
		}
	}
	rev := dg.Reverse()
	for len(level) > 0 {
		slices.Sort(level)
		slices.Reverse(level)
		sl = append(sl, level...)
		var nextLevel []T
		for _, v := range level {
			for pred := range rev[v].Values() {
				outDegree[pred]--
				if outDegree[pred] == 0 {
					nextLevel = append(nextLevel, pred)
				}
			}
		}
		level = nextLevel
	}
	return sl
}
