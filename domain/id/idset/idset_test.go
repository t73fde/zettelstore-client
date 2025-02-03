//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore Client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2021-present Detlef Stern
//-----------------------------------------------------------------------------

package idset_test

import (
	"testing"

	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/domain/id/idset"
	"t73f.de/r/zsc/domain/id/idslice"
)

func TestSetContainsOrNil(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s   *idset.Set
		zid id.Zid
		exp bool
	}{
		{nil, id.Invalid, true},
		{nil, 14, true},
		{idset.NewSet(), id.Invalid, false},
		{idset.NewSet(), 1, false},
		{idset.NewSet(), id.Invalid, false},
		{idset.NewSet(1), 1, true},
	}
	for i, tc := range testcases {
		got := tc.s.ContainsOrNil(tc.zid)
		if got != tc.exp {
			t.Errorf("%d: %v.ContainsOrNil(%v) == %v, but got %v", i, tc.s, tc.zid, tc.exp, got)
		}
	}
}

func TestSetAdd(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s1, s2 *idset.Set
		exp    idslice.Slice
	}{
		{nil, nil, nil},
		{idset.NewSet(), nil, nil},
		{idset.NewSet(), idset.NewSet(), nil},
		{nil, idset.NewSet(1), idslice.Slice{1}},
		{idset.NewSet(1), nil, idslice.Slice{1}},
		{idset.NewSet(1), idset.NewSet(), idslice.Slice{1}},
		{idset.NewSet(1), idset.NewSet(2), idslice.Slice{1, 2}},
		{idset.NewSet(1), idset.NewSet(1), idslice.Slice{1}},
	}
	for i, tc := range testcases {
		sl1 := tc.s1.SafeSorted()
		sl2 := tc.s2.SafeSorted()
		got := tc.s1.IUnion(tc.s2).SafeSorted()
		if !got.Equal(tc.exp) {
			t.Errorf("%d: %v.Add(%v) should be %v, but got %v", i, sl1, sl2, tc.exp, got)
		}
	}
}

func TestSetSafeSorted(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		set *idset.Set
		exp idslice.Slice
	}{
		{nil, nil},
		{idset.NewSet(), nil},
		{idset.NewSet(9, 4, 6, 1, 7), idslice.Slice{1, 4, 6, 7, 9}},
	}
	for i, tc := range testcases {
		got := tc.set.SafeSorted()
		if !got.Equal(tc.exp) {
			t.Errorf("%d: %v.SafeSorted() should be %v, but got %v", i, tc.set, tc.exp, got)
		}
	}
}

func TestSetIntersectOrSet(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s1, s2 *idset.Set
		exp    idslice.Slice
	}{
		{nil, nil, nil},
		{idset.NewSet(), nil, nil},
		{nil, idset.NewSet(), nil},
		{idset.NewSet(), idset.NewSet(), nil},
		{idset.NewSet(1), nil, nil},
		{nil, idset.NewSet(1), idslice.Slice{1}},
		{idset.NewSet(1), idset.NewSet(), nil},
		{idset.NewSet(), idset.NewSet(1), nil},
		{idset.NewSet(1), idset.NewSet(2), nil},
		{idset.NewSet(2), idset.NewSet(1), nil},
		{idset.NewSet(1), idset.NewSet(1), idslice.Slice{1}},
	}
	for i, tc := range testcases {
		sl1 := tc.s1.SafeSorted()
		sl2 := tc.s2.SafeSorted()
		got := tc.s1.IntersectOrSet(tc.s2).SafeSorted()
		if !got.Equal(tc.exp) {
			t.Errorf("%d: %v.IntersectOrSet(%v) should be %v, but got %v", i, sl1, sl2, tc.exp, got)
		}
	}
}

func TestSetIUnion(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s1, s2 *idset.Set
		exp    *idset.Set
	}{
		{nil, nil, nil},
		{idset.NewSet(), nil, nil},
		{nil, idset.NewSet(), nil},
		{idset.NewSet(), idset.NewSet(), nil},
		{idset.NewSet(1), nil, idset.NewSet(1)},
		{nil, idset.NewSet(1), idset.NewSet(1)},
		{idset.NewSet(1), idset.NewSet(), idset.NewSet(1)},
		{idset.NewSet(), idset.NewSet(1), idset.NewSet(1)},
		{idset.NewSet(1), idset.NewSet(2), idset.NewSet(1, 2)},
		{idset.NewSet(2), idset.NewSet(1), idset.NewSet(2, 1)},
		{idset.NewSet(1), idset.NewSet(1), idset.NewSet(1)},
		{idset.NewSet(1, 2, 3), idset.NewSet(2, 3, 4), idset.NewSet(1, 2, 3, 4)},
	}
	for i, tc := range testcases {
		s1 := tc.s1.Clone()
		sl1 := s1.SafeSorted()
		sl2 := tc.s2.SafeSorted()
		got := s1.IUnion(tc.s2)
		if !got.Equal(tc.exp) {
			t.Errorf("%d: %v.IUnion(%v) should be %v, but got %v", i, sl1, sl2, tc.exp, got)
		}
	}
}

func TestSetISubtract(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s1, s2 *idset.Set
		exp    idslice.Slice
	}{
		{nil, nil, nil},
		{idset.NewSet(), nil, nil},
		{nil, idset.NewSet(), nil},
		{idset.NewSet(), idset.NewSet(), nil},
		{idset.NewSet(1), nil, idslice.Slice{1}},
		{nil, idset.NewSet(1), nil},
		{idset.NewSet(1), idset.NewSet(), idslice.Slice{1}},
		{idset.NewSet(), idset.NewSet(1), nil},
		{idset.NewSet(1), idset.NewSet(2), idslice.Slice{1}},
		{idset.NewSet(2), idset.NewSet(1), idslice.Slice{2}},
		{idset.NewSet(1), idset.NewSet(1), nil},
		{idset.NewSet(1, 2, 3), idset.NewSet(1), idslice.Slice{2, 3}},
		{idset.NewSet(1, 2, 3), idset.NewSet(2), idslice.Slice{1, 3}},
		{idset.NewSet(1, 2, 3), idset.NewSet(3), idslice.Slice{1, 2}},
		{idset.NewSet(1, 2, 3), idset.NewSet(1, 2), idslice.Slice{3}},
		{idset.NewSet(1, 2, 3), idset.NewSet(1, 3), idslice.Slice{2}},
		{idset.NewSet(1, 2, 3), idset.NewSet(2, 3), idslice.Slice{1}},
	}
	for i, tc := range testcases {
		s1 := tc.s1.Clone()
		sl1 := s1.SafeSorted()
		sl2 := tc.s2.SafeSorted()
		s1.ISubstract(tc.s2)
		got := s1.SafeSorted()
		if !got.Equal(tc.exp) {
			t.Errorf("%d: %v.ISubstract(%v) should be %v, but got %v", i, sl1, sl2, tc.exp, got)
		}
	}
}

func TestSetDiff(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in1, in2   *idset.Set
		exp1, exp2 *idset.Set
	}{
		{nil, nil, nil, nil},
		{idset.NewSet(1), nil, nil, idset.NewSet(1)},
		{nil, idset.NewSet(1), idset.NewSet(1), nil},
		{idset.NewSet(1), idset.NewSet(1), nil, nil},
		{idset.NewSet(1, 2), idset.NewSet(1), nil, idset.NewSet(2)},
		{idset.NewSet(1), idset.NewSet(1, 2), idset.NewSet(2), nil},
		{idset.NewSet(1, 2), idset.NewSet(1, 3), idset.NewSet(3), idset.NewSet(2)},
		{idset.NewSet(1, 2, 3), idset.NewSet(2, 3, 4), idset.NewSet(4), idset.NewSet(1)},
		{idset.NewSet(2, 3, 4), idset.NewSet(1, 2, 3), idset.NewSet(1), idset.NewSet(4)},
	}
	for i, tc := range testcases {
		gotN, gotO := tc.in1.Diff(tc.in2)
		if !tc.exp1.Equal(gotN) {
			t.Errorf("%d: expected %v, but got: %v", i, tc.exp1, gotN)
		}
		if !tc.exp2.Equal(gotO) {
			t.Errorf("%d: expected %v, but got: %v", i, tc.exp2, gotO)
		}
	}
}

func TestSetRemove(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s1, s2 *idset.Set
		exp    idslice.Slice
	}{
		{nil, nil, nil},
		{idset.NewSet(), nil, nil},
		{idset.NewSet(), idset.NewSet(), nil},
		{idset.NewSet(1), nil, idslice.Slice{1}},
		{idset.NewSet(1), idset.NewSet(), idslice.Slice{1}},
		{idset.NewSet(1), idset.NewSet(2), idslice.Slice{1}},
		{idset.NewSet(1), idset.NewSet(1), idslice.Slice{}},
	}
	for i, tc := range testcases {
		sl1 := tc.s1.SafeSorted()
		sl2 := tc.s2.SafeSorted()
		newS1 := idset.NewSet(sl1...)
		newS1.ISubstract(tc.s2)
		got := newS1.SafeSorted()
		if !got.Equal(tc.exp) {
			t.Errorf("%d: %v.Remove(%v) should be %v, but got %v", i, sl1, sl2, tc.exp, got)
		}
	}
}

func BenchmarkSet(b *testing.B) {
	s := idset.NewSetCap(b.N)
	for i := range b.N {
		s.Add(id.Zid(i))
	}
}
