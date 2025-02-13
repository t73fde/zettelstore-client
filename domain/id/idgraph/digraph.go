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

// Package idgraph implements a graph of zettel identifier.
package idgraph

import (
	"maps"
	"slices"

	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/domain/id/idset"
)

// Digraph relates zettel identifier in a directional way.
type Digraph map[id.Zid]*idset.Set

// AddVertex adds an edge / vertex to the digraph.
func (dg Digraph) AddVertex(zid id.Zid) Digraph {
	if dg == nil {
		return Digraph{zid: nil}
	}
	if _, found := dg[zid]; !found {
		dg[zid] = nil
	}
	return dg
}

// RemoveVertex removes a vertex and all its edges from the digraph.
func (dg Digraph) RemoveVertex(zid id.Zid) {
	if len(dg) > 0 {
		delete(dg, zid)
		for vertex, closure := range dg {
			dg[vertex] = closure.Remove(zid)
		}
	}
}

// AddEdge adds a connection from `zid1` to `zid2`.
// Both vertices must be added before. Otherwise the function may panic.
func (dg Digraph) AddEdge(fromZid, toZid id.Zid) Digraph {
	if dg == nil {
		return Digraph{fromZid: (*idset.Set)(nil).Add(toZid), toZid: nil}
	}
	dg[fromZid] = dg[fromZid].Add(toZid)
	return dg
}

// AddEgdes adds all given `Edge`s to the digraph.
//
// In contrast to `AddEdge` the vertices must not exist before.
func (dg Digraph) AddEgdes(edges EdgeSlice) Digraph {
	if dg == nil {
		if len(edges) == 0 {
			return nil
		}
		dg = make(Digraph, len(edges))
	}
	for _, edge := range edges {
		dg = dg.AddVertex(edge.From)
		dg = dg.AddVertex(edge.To)
		dg = dg.AddEdge(edge.From, edge.To)
	}
	return dg
}

// Equal returns true if both digraphs have the same vertices and edges.
func (dg Digraph) Equal(other Digraph) bool {
	return maps.EqualFunc(dg, other, func(cg, co *idset.Set) bool { return cg.Equal(co) })
}

// Clone a digraph.
func (dg Digraph) Clone() Digraph {
	if len(dg) == 0 {
		return nil
	}
	copyDG := make(Digraph, len(dg))
	for vertex, closure := range dg {
		copyDG[vertex] = closure.Clone()
	}
	return copyDG
}

// HasVertex returns true, if `zid` is a vertex of the digraph.
func (dg Digraph) HasVertex(zid id.Zid) bool {
	if len(dg) == 0 {
		return false
	}
	_, found := dg[zid]
	return found
}

// Vertices returns the set of all vertices.
func (dg Digraph) Vertices() *idset.Set {
	if len(dg) == 0 {
		return nil
	}
	verts := idset.NewCap(len(dg))
	for vert := range dg {
		verts.Add(vert)
	}
	return verts
}

// Edges returns an unsorted slice of the edges of the digraph.
func (dg Digraph) Edges() (es EdgeSlice) {
	for vert, closure := range dg {
		closure.ForEach(func(next id.Zid) {
			es = append(es, Edge{From: vert, To: next})
		})
	}
	return es
}

// Originators will return the set of all vertices that are not referenced
// a the to-part of an edge.
func (dg Digraph) Originators() *idset.Set {
	if len(dg) == 0 {
		return nil
	}
	origs := dg.Vertices()
	for _, closure := range dg {
		origs.ISubstract(closure)
	}
	return origs
}

// Terminators returns the set of all vertices that does not reference
// other vertices.
func (dg Digraph) Terminators() (terms *idset.Set) {
	for vert, closure := range dg {
		if closure.IsEmpty() {
			terms = terms.Add(vert)
		}
	}
	return terms
}

// TransitiveClosure calculates the sub-graph that is reachable from `zid`.
func (dg Digraph) TransitiveClosure(zid id.Zid) (tc Digraph) {
	if len(dg) == 0 {
		return nil
	}
	var marked *idset.Set
	stack := []id.Zid{zid}
	for pos := len(stack) - 1; pos >= 0; pos = len(stack) - 1 {
		curr := stack[pos]
		stack = stack[:pos]
		if marked.Contains(curr) {
			continue
		}
		tc = tc.AddVertex(curr)
		dg[curr].ForEach(func(next id.Zid) {
			tc = tc.AddVertex(next)
			tc = tc.AddEdge(curr, next)
			stack = append(stack, next)
		})
		marked = marked.Add(curr)
	}
	return tc
}

// ReachableVertices calculates the set of all vertices that are reachable
// from the given `zid`.
func (dg Digraph) ReachableVertices(zid id.Zid) (tc *idset.Set) {
	if len(dg) == 0 {
		return nil
	}
	stack := dg[zid].SafeSorted()
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
		closure.ForEach(func(next id.Zid) {
			stack = append(stack, next)
		})
	}
	return tc
}

// IsDAG returns a vertex and false, if the graph has a cycle containing the vertex.
func (dg Digraph) IsDAG() (id.Zid, bool) {
	for vertex := range dg {
		if dg.ReachableVertices(vertex).Contains(vertex) {
			return vertex, false
		}
	}
	return id.Invalid, true
}

// Reverse returns a graph with reversed edges.
func (dg Digraph) Reverse() (revDg Digraph) {
	for vertex, closure := range dg {
		revDg = revDg.AddVertex(vertex)
		closure.ForEach(func(next id.Zid) {
			revDg = revDg.AddVertex(next)
			revDg = revDg.AddEdge(next, vertex)
		})
	}
	return revDg
}

// SortReverse returns a deterministic, topological, reverse sort of the
// digraph.
//
// Works only if digraph is a DAG. Otherwise the algorithm will not terminate
// or returns an arbitrary value.
func (dg Digraph) SortReverse() (sl []id.Zid) {
	if len(dg) == 0 {
		return nil
	}
	tempDg := dg.Clone()
	for len(tempDg) > 0 {
		terms := tempDg.Terminators()
		if terms.IsEmpty() {
			break
		}
		termSlice := terms.SafeSorted()
		slices.Reverse(termSlice)
		sl = append(sl, termSlice...)
		terms.ForEach(func(t id.Zid) {
			tempDg.RemoveVertex(t)
		})
	}
	return sl
}
