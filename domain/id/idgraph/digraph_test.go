//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore Client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package idgraph_test

import (
	"testing"

	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/domain/id/idgraph"
	"t73f.de/r/zsc/domain/id/idset"
	"t73f.de/r/zsc/domain/id/idslice"
)

type zps = idgraph.EdgeSlice

func createDigraph(pairs zps) (dg idgraph.Digraph) {
	return dg.AddEgdes(pairs)
}

func TestDigraphOriginators(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   idgraph.EdgeSlice
		orig *idset.Set
		term *idset.Set
	}{
		{"empty", nil, nil, nil},
		{"single", zps{{0, 1}}, idset.New(0), idset.New(1)},
		{"chain", zps{{0, 1}, {1, 2}, {2, 3}}, idset.New(0), idset.New(3)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dg := createDigraph(tc.dg)
			if got := dg.Originators(); !tc.orig.Equal(got) {
				t.Errorf("Originators: expected:\n%v, but got:\n%v", tc.orig, got)
			}
			if got := dg.Terminators(); !tc.term.Equal(got) {
				t.Errorf("Termintors: expected:\n%v, but got:\n%v", tc.orig, got)
			}
		})
	}
}

func TestDigraphReachableVertices(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name  string
		pairs idgraph.EdgeSlice
		start id.Zid
		exp   *idset.Set
	}{
		{"nil", nil, 0, nil},
		{"0-2", zps{{1, 2}, {2, 3}}, 1, idset.New(2, 3)},
		{"1,2", zps{{1, 2}, {2, 3}}, 2, idset.New(3)},
		{"0-2,1-2", zps{{1, 2}, {2, 3}, {1, 3}}, 1, idset.New(2, 3)},
		{"0-2,1-2/1", zps{{1, 2}, {2, 3}, {1, 3}}, 2, idset.New(3)},
		{"0-2,1-2/2", zps{{1, 2}, {2, 3}, {1, 3}}, 3, nil},
		{"0-2,1-2,3*", zps{{1, 2}, {2, 3}, {1, 3}, {4, 4}}, 1, idset.New(2, 3)},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dg := createDigraph(tc.pairs)
			if got := dg.ReachableVertices(tc.start); !got.Equal(tc.exp) {
				t.Errorf("\n%v, but got:\n%v", tc.exp, got)
			}

		})
	}
}

func TestDigraphTransitiveClosure(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name  string
		pairs idgraph.EdgeSlice
		start id.Zid
		exp   idgraph.EdgeSlice
	}{
		{"nil", nil, 0, nil},
		{"1-3", zps{{1, 2}, {2, 3}}, 1, zps{{1, 2}, {2, 3}}},
		{"1,2", zps{{1, 1}, {2, 3}}, 2, zps{{2, 3}}},
		{"0-2,1-2", zps{{1, 2}, {2, 3}, {1, 3}}, 1, zps{{1, 2}, {1, 3}, {2, 3}}},
		{"0-2,1-2/1", zps{{1, 2}, {2, 3}, {1, 3}}, 1, zps{{1, 2}, {1, 3}, {2, 3}}},
		{"0-2,1-2/2", zps{{1, 2}, {2, 3}, {1, 3}}, 2, zps{{2, 3}}},
		{"0-2,1-2,3*", zps{{1, 2}, {2, 3}, {1, 3}, {4, 4}}, 1, zps{{1, 2}, {1, 3}, {2, 3}}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dg := createDigraph(tc.pairs)
			if got := dg.TransitiveClosure(tc.start).Edges().Sort(); !got.Equal(tc.exp) {
				t.Errorf("\n%v, but got:\n%v", tc.exp, got)
			}
		})
	}
}

func TestIsDAG(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   idgraph.EdgeSlice
		exp  bool
	}{
		{"empty", nil, true},
		{"single-edge", zps{{1, 2}}, true},
		{"single-loop", zps{{1, 1}}, false},
		{"long-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 2}}, false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if zid, got := createDigraph(tc.dg).IsDAG(); got != tc.exp {
				t.Errorf("expected %v, but got %v (%v)", tc.exp, got, zid)
			}
		})
	}
}

func TestDigraphReverse(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   idgraph.EdgeSlice
		exp  idgraph.EdgeSlice
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
			dg := createDigraph(tc.dg)
			if got := dg.Reverse().Edges().Sort(); !got.Equal(tc.exp) {
				t.Errorf("\n%v, but got:\n%v", tc.exp, got)
			}
		})
	}
}

func TestDigraphSortReverse(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name string
		dg   idgraph.EdgeSlice
		exp  idslice.Slice
	}{
		{"empty", nil, nil},
		{"single-edge", zps{{1, 2}}, idslice.Slice{2, 1}},
		{"single-loop", zps{{1, 1}}, nil},
		{"end-loop", zps{{1, 2}, {2, 2}}, idslice.Slice{}},
		{"long-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 2}}, idslice.Slice{}},
		{"sect-loop", zps{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {4, 2}}, idslice.Slice{5}},
		{"two-islands", zps{{1, 2}, {2, 3}, {4, 5}}, idslice.Slice{5, 3, 4, 2, 1}},
		{"direct-indirect", zps{{1, 2}, {1, 3}, {3, 2}}, idslice.Slice{2, 3, 1}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := createDigraph(tc.dg).SortReverse(); !got.Equal(tc.exp) {
				t.Errorf("expected:\n%v, but got:\n%v", tc.exp, got)
			}
		})
	}
}
