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

package graph_test

import (
	"slices"
	"testing"

	"t73f.de/r/zero/graph"
	"t73f.de/r/zero/set"
)

type zps = graph.EdgeSlice[int]

func createDigraph(pairs zps) (dg graph.Digraph[int]) {
	return dg.AddEdges(pairs)
}

func TestDigraphEqual(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		equal    bool
		dg1, dg2 zps
	}{
		{"both empty", true, nil, nil},
		{"one empty", false, nil, zps{{1, 2}}},
		{"reverse", false, zps{{2, 1}}, zps{{1, 2}}},
		{"seq", true, zps{{1, 2}, {3, 4}}, zps{{3, 4}, {1, 2}}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg1, dg2 := createDigraph(tc.dg1), createDigraph(tc.dg2)
			if got := dg1.Equal(dg2); got != tc.equal {
				t.Errorf("%v.Equal(%v): expected %v, but got %v", tc.dg1, tc.dg2, tc.equal, got)
			}
			if got := dg2.Equal(dg1); got != tc.equal {
				t.Errorf("%v.Equal(%v): expected %v, but got %v", tc.dg2, tc.dg1, tc.equal, got)
			}
		})
	}
}

func TestDigraphHasVertex(t *testing.T) {
	t.Parallel()
	var dg graph.Digraph[int]
	for i := range 10 {
		if dg.HasVertex(i) {
			t.Error("nil digraph has vertex:", i)
		}
	}
	dg = createDigraph(zps{{0, 6}, {1, 2}, {2, 3}, {4, 5}})
	for i := range 7 {
		if !dg.HasVertex(i) {
			t.Error("digraph should have vertex:", i)
		}
	}
	for i := range 7 {
		if dg.HasVertex(i + 7) {
			t.Error("digraph must not have vertex:", i+7)
		}
	}
}

func TestDigraphAddVertex(t *testing.T) {
	t.Parallel()
	var dg graph.Digraph[int]
	dg = dg.AddVertex(1)
	dg = dg.AddVertex(1)
	dg = dg.AddVertex(2)
	if exp := set.New(1, 2); !dg.Vertices().Equal(exp) {
		t.Errorf("expected vertices %v, but got %v", exp, dg.Vertices())
	}
	if edges := dg.Edges(); len(edges) != 0 {
		t.Errorf("expected no edges, but got %v", edges)
	}
}

func TestDigraphAddEdge(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		start    zps
		from, to int
		exp      zps // must be sorted
		expVerts set.Set[int]
	}{
		{"nil", nil, 1, 2, zps{{1, 2}}, set.New(1, 2)},
		{"nil self-loop", nil, 1, 1, zps{{1, 1}}, set.New(1)},
		{"unknown vertices", zps{{1, 2}}, 3, 4, zps{{1, 2}, {3, 4}}, set.New(1, 2, 3, 4)},
		{"known vertices", zps{{1, 2}, {3, 4}}, 2, 3, zps{{1, 2}, {2, 3}, {3, 4}}, set.New(1, 2, 3, 4)},
		{"duplicate", zps{{1, 2}}, 1, 2, zps{{1, 2}}, set.New(1, 2)},
		{"self-loop", zps{{1, 2}}, 2, 2, zps{{1, 2}, {2, 2}}, set.New(1, 2)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg := createDigraph(tc.start).AddEdge(tc.from, tc.to)
			if got := dg.Edges().Sort(); !got.Equal(tc.exp) {
				t.Errorf("edges: expected %v, but got %v", tc.exp, got)
			}
			if got := dg.Vertices(); !got.Equal(tc.expVerts) {
				t.Errorf("vertices: expected %v, but got %v", tc.expVerts, got)
			}
		})
	}
}

func TestDigraphAddEdgesEmpty(t *testing.T) {
	t.Parallel()
	var dg graph.Digraph[int]
	if got := dg.AddEdges(nil); got != nil {
		t.Errorf("expected nil, but got %v", got)
	}
}

func TestDigraphRemoveVertex(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name   string
		dg     zps
		remove int
		exp    zps // must be sorted
		verts  set.Set[int]
	}{
		{"nil", nil, 1, nil, set.Set[int]{}},
		{"unknown", zps{{1, 2}}, 7, zps{{1, 2}}, set.New(1, 2)},
		{"middle", zps{{1, 2}, {2, 3}, {1, 3}}, 2, zps{{1, 3}}, set.New(1, 3)},
		{"loop", zps{{1, 1}, {1, 2}}, 1, nil, set.New(2)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg := createDigraph(tc.dg)
			dg.RemoveVertex(tc.remove)
			if got := dg.Edges().Sort(); !got.Equal(tc.exp) {
				t.Errorf("edges: expected %v, but got %v", tc.exp, got)
			}
			if got := dg.Vertices(); !got.Equal(tc.verts) {
				t.Errorf("vertices: expected %v, but got %v", tc.verts, got)
			}
		})
	}
}

func TestDigraphClone(t *testing.T) {
	t.Parallel()
	if got := graph.Digraph[int](nil).Clone(); got != nil {
		t.Errorf("clone of nil must be nil, but got %v", got)
	}
	orig := createDigraph(zps{{1, 2}})
	clone := orig.Clone()
	if !orig.Equal(clone) {
		t.Fatalf("clone differs: %v vs. %v", orig, clone)
	}
	orig.AddEdge(2, 3)
	if exp := createDigraph(zps{{1, 2}}); !clone.Equal(exp) {
		t.Errorf("clone was modified through original: %v", clone)
	}
	if orig.Equal(clone) {
		t.Error("original and clone must differ after modification")
	}
}

func TestDigraphVerticesEdges(t *testing.T) {
	t.Parallel()
	var empty graph.Digraph[int]
	if got := empty.Vertices(); !got.Equal(set.Set[int]{}) {
		t.Errorf("empty graph: vertices %v", got)
	}
	if got := empty.Edges(); len(got) != 0 {
		t.Errorf("empty graph: edges %v", got)
	}
	dg := createDigraph(zps{{2, 3}, {1, 2}, {4, 4}})
	if exp, got := set.New(1, 2, 3, 4), dg.Vertices(); !got.Equal(exp) {
		t.Errorf("vertices: expected %v, but got %v", exp, got)
	}
	if exp, got := (zps{{1, 2}, {2, 3}, {4, 4}}), dg.Edges().Sort(); !got.Equal(exp) {
		t.Errorf("edges: expected %v, but got %v", exp, got)
	}
}

func TestDigraphOriginators(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   zps
		orig set.Set[int]
		term set.Set[int]
	}{
		{"empty", nil, set.Set[int]{}, set.Set[int]{}},
		{"single", zps{{0, 1}}, set.New(0), set.New(1)},
		{"chain", zps{{0, 1}, {1, 2}, {2, 3}}, set.New(0), set.New(3)},
		{"loop", zps{{0, 1}, {1, 1}}, set.New(0), set.Set[int]{}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg := createDigraph(tc.dg)
			if got := dg.Originators(); !tc.orig.Equal(got) {
				t.Errorf("Originators: expected %v, but got %v", tc.orig, got)
			}
			if got := dg.Terminators(); !tc.term.Equal(got) {
				t.Errorf("Terminators: expected %v, but got %v", tc.term, got)
			}
		})
	}
}

func TestDigraphReachableVertices(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name  string
		pairs zps
		start int
		exp   set.Set[int]
	}{
		{"nil", nil, 0, set.Set[int]{}},
		{"unknown", zps{{1, 2}}, 9, set.Set[int]{}},
		{"0-2", zps{{1, 2}, {2, 3}}, 1, set.New(2, 3)},
		{"1,2", zps{{1, 2}, {2, 3}}, 2, set.New(3)},
		{"0-2,1-2", zps{{1, 2}, {2, 3}, {1, 3}}, 1, set.New(2, 3)},
		{"0-2,1-2/1", zps{{1, 2}, {2, 3}, {1, 3}}, 2, set.New(3)},
		{"0-2,1-2/2", zps{{1, 2}, {2, 3}, {1, 3}}, 3, set.Set[int]{}},
		{"0-2,1-2,3*", zps{{1, 2}, {2, 3}, {1, 3}, {4, 4}}, 1, set.New(2, 3)},
		{"cycle", zps{{1, 2}, {2, 1}}, 1, set.New(1, 2)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg := createDigraph(tc.pairs)
			if got := dg.ReachableVertices(tc.start); !got.Equal(tc.exp) {
				t.Errorf("expected %v, but got %v", tc.exp, got)
			}
		})
	}
}

func TestDigraphTransitiveClosure(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name  string
		pairs zps
		start int
		exp   zps
	}{
		{"nil", nil, 0, nil},
		{"unknown", zps{{1, 2}}, 7, nil},
		{"1-3", zps{{1, 2}, {2, 3}}, 1, zps{{1, 2}, {2, 3}}},
		{"1,2", zps{{1, 1}, {2, 3}}, 2, zps{{2, 3}}},
		{"0-2,1-2", zps{{1, 2}, {2, 3}, {1, 3}}, 1, zps{{1, 2}, {1, 3}, {2, 3}}},
		{"0-2,1-2/2", zps{{1, 2}, {2, 3}, {1, 3}}, 2, zps{{2, 3}}},
		{"0-2,1-2,3*", zps{{1, 2}, {2, 3}, {1, 3}, {4, 4}}, 1, zps{{1, 2}, {1, 3}, {2, 3}}},
		{"cycle", zps{{1, 2}, {2, 1}, {3, 1}}, 1, zps{{1, 2}, {2, 1}}},
		{"terminator", zps{{1, 2}}, 2, nil},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg := createDigraph(tc.pairs)
			if got := dg.TransitiveClosure(tc.start).Edges().Sort(); !got.Equal(tc.exp) {
				t.Errorf("expected %v, but got %v", tc.exp, got)
			}
		})
	}
}

func TestIsDAG(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name  string
		dg    zps
		exp   bool
		cycle set.Set[int] // vertices that may be reported if !exp
	}{
		{"empty", nil, true, set.Set[int]{}},
		{"single-edge", zps{{1, 2}}, true, set.Set[int]{}},
		{"diamond", zps{{1, 2}, {1, 3}, {2, 4}, {3, 4}}, true, set.Set[int]{}},
		{"single-loop", zps{{1, 1}}, false, set.New(1)},
		{"long-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 2}}, false, set.New(2, 3, 4, 5)},
		{"loop-in-island", zps{{1, 2}, {3, 4}, {4, 3}}, false, set.New(3, 4)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v, got := createDigraph(tc.dg).IsDAG()
			if got != tc.exp {
				t.Fatalf("expected %v, but got %v (%v)", tc.exp, got, v)
			}
			if !tc.exp && !tc.cycle.Contains(v) {
				t.Errorf("reported vertex %v is not on a cycle %v", v, tc.cycle)
			}
		})
	}
}

func TestDigraphReverse(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   zps
		exp  zps
	}{
		{"empty", nil, nil},
		{"single-edge", zps{{1, 2}}, zps{{2, 1}}},
		{"single-loop", zps{{1, 1}}, zps{{1, 1}}},
		{"end-loop", zps{{1, 2}, {2, 2}}, zps{{2, 1}, {2, 2}}},
		{"long-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 2}}, zps{{2, 1}, {2, 5}, {3, 2}, {4, 3}, {5, 4}}},
		{"sect-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {4, 2}}, zps{{2, 1}, {2, 4}, {3, 2}, {4, 3}, {5, 4}}},
		{"two-islands", zps{{1, 2}, {2, 3}, {4, 5}}, zps{{2, 1}, {3, 2}, {5, 4}}},
		{"direct-indirect", zps{{1, 2}, {1, 3}, {3, 2}}, zps{{2, 1}, {2, 3}, {3, 1}}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dg := createDigraph(tc.dg)
			rev := dg.Reverse()
			if got := rev.Edges().Sort(); !got.Equal(tc.exp) {
				t.Errorf("expected %v, but got %v", tc.exp, got)
			}
			if got := rev.Vertices(); !got.Equal(dg.Vertices()) {
				t.Errorf("vertices differ: expected %v, but got %v", dg.Vertices(), got)
			}
		})
	}
}

func TestDigraphSortReverse(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   zps
		exp  []int
	}{
		{"empty", nil, nil},
		{"single-edge", zps{{1, 2}}, []int{2, 1}},
		{"single-loop", zps{{1, 1}}, nil},
		{"end-loop", zps{{1, 2}, {2, 2}}, nil},
		{"long-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 2}}, nil},
		{"sect-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {4, 2}}, []int{5}},
		{"two-islands", zps{{1, 2}, {2, 3}, {4, 5}}, []int{5, 3, 4, 2, 1}},
		{"direct-indirect", zps{{1, 2}, {1, 3}, {3, 2}}, []int{2, 3, 1}},
		{"diamond", zps{{1, 2}, {1, 3}, {2, 4}, {3, 4}}, []int{4, 3, 2, 1}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := createDigraph(tc.dg).SortReverse()
			if !slices.Equal(got, tc.exp) || (got == nil) != (tc.exp == nil) {
				t.Errorf("expected %v, but got %v", tc.exp, got)
			}
		})
	}
}

func TestEdgeSliceSort(t *testing.T) {
	t.Parallel()
	es := zps{{2, 1}, {1, 3}, {1, 2}, {2, 0}}
	exp := zps{{1, 2}, {1, 3}, {2, 0}, {2, 1}}
	if got := es.Sort(); !got.Equal(exp) {
		t.Errorf("expected %v, but got %v", exp, got)
	}
}
