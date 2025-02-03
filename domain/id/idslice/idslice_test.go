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

package idslice_test

import (
	"testing"

	"t73f.de/r/zsc/domain/id/idslice"
)

func TestSliceSort(t *testing.T) {
	t.Parallel()
	zs := idslice.Slice{9, 4, 6, 1, 7}
	zs.Sort()
	exp := idslice.Slice{1, 4, 6, 7, 9}
	if !zs.Equal(exp) {
		t.Errorf("Slice.Sort did not work. Expected %v, got %v", exp, zs)
	}
}

func TestCopy(t *testing.T) {
	t.Parallel()
	var orig idslice.Slice
	got := orig.Clone()
	if got != nil {
		t.Errorf("Nil copy resulted in %v", got)
	}
	orig = idslice.Slice{9, 4, 6, 1, 7}
	got = orig.Clone()
	if !orig.Equal(got) {
		t.Errorf("Slice.Copy did not work. Expected %v, got %v", orig, got)
	}
}

func TestSliceEqual(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s1, s2 idslice.Slice
		exp    bool
	}{
		{nil, nil, true},
		{nil, idslice.Slice{}, true},
		{nil, idslice.Slice{1}, false},
		{idslice.Slice{1}, idslice.Slice{1}, true},
		{idslice.Slice{1}, idslice.Slice{2}, false},
		{idslice.Slice{1, 2}, idslice.Slice{2, 1}, false},
		{idslice.Slice{1, 2}, idslice.Slice{1, 2}, true},
	}
	for i, tc := range testcases {
		got := tc.s1.Equal(tc.s2)
		if got != tc.exp {
			t.Errorf("%d/%v.Equal(%v)==%v, but got %v", i, tc.s1, tc.s2, tc.exp, got)
		}
		got = tc.s2.Equal(tc.s1)
		if got != tc.exp {
			t.Errorf("%d/%v.Equal(%v)==%v, but got %v", i, tc.s2, tc.s1, tc.exp, got)
		}
	}
}

func TestSliceMetaString(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		in  idslice.Slice
		exp string
	}{
		{nil, ""},
		{idslice.Slice{}, ""},
		{idslice.Slice{1}, "00000000000001"},
		{idslice.Slice{1, 2}, "00000000000001 00000000000002"},
	}
	for i, tc := range testcases {
		got := tc.in.MetaString()
		if got != tc.exp {
			t.Errorf("%d/%v: expected %q, but got %q", i, tc.in, tc.exp, got)
		}
	}
}
