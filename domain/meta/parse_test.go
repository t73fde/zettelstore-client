//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore Client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package meta_test

import (
	"iter"
	"strings"
	"testing"

	"t73f.de/r/zsc/domain/meta"
	"t73f.de/r/zsc/input"
)

func parseMetaStr(src string) *meta.Meta {
	return meta.NewFromInput(testID, input.NewInput([]byte(src)))
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	m := parseMetaStr("")
	if got, ok := m.Get(meta.KeySyntax); ok || got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	if got, ok := m.GetList(meta.KeyTags); ok || len(got) > 0 {
		t.Errorf("Tags are not nil, but %v", got)
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()
	td := []struct {
		s string
		e meta.Value
	}{
		{meta.KeyTitle + ": a title", "a title"},
		{meta.KeyTitle + ": a\n\t title", "a title"},
		{meta.KeyTitle + ": a\n\t title\r\n  x", "a title x"},
		{meta.KeyTitle + " AbC", "AbC"},
		{meta.KeyTitle + " AbC\n ded", "AbC ded"},
		{meta.KeyTitle + ": o\ntitle: p", "o p"},
		{meta.KeyTitle + ": O\n\ntitle: P", "O"},
		{meta.KeyTitle + ": b\r\ntitle: c", "b c"},
		{meta.KeyTitle + ": B\r\n\r\ntitle: C", "B"},
		{meta.KeyTitle + ": r\rtitle: q", "r q"},
		{meta.KeyTitle + ": R\r\rtitle: Q", "R"},
	}
	for i, tc := range td {
		m := parseMetaStr(tc.s)
		if got, ok := m.Get(meta.KeyTitle); !ok || got != tc.e {
			t.Log(m)
			t.Errorf("TC=%d: expected %q, got %q", i, tc.e, got)
		}
	}
}

func TestTags(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		src string
		exp string
	}{
		{"", ""},
		{meta.KeyTags + ":", ""},
		{meta.KeyTags + ": c", ""},
		{meta.KeyTags + ": #", ""},
		{meta.KeyTags + ": #c", "c"},
		{meta.KeyTags + ": #c #", "c"},
		{meta.KeyTags + ": #c #b", "b c"},
		{meta.KeyTags + ": #c # #", "c"},
		{meta.KeyTags + ": #c # #b", "b c"},
	}
	for i, tc := range testcases {
		m := parseMetaStr(tc.src)
		tagsString, found := m.Get(meta.KeyTags)
		if !found {
			if tc.exp != "" {
				t.Errorf("%d / %q: no %s found", i, tc.src, meta.KeyTags)
			}
			continue
		}
		tags := tagsString.AsTags()
		if tc.exp == "" && len(tags) > 0 {
			t.Errorf("%d / %q: expected no %s, but got %v", i, tc.src, meta.KeyTags, tags)
			continue
		}
		got := strings.Join(tags, " ")
		if tc.exp != got {
			t.Errorf("%d / %q: expected %q, got: %q", i, tc.src, tc.exp, got)
		}
	}
}

func TestNewFromInput(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		input string
		exp   []pair
	}{
		{"", []pair{}},
		{" a:b", []pair{{"a", "b"}}},
		{"%a:b", []pair{}},
		{"a:b\r\n\r\nc:d", []pair{{"a", "b"}}},
		{"a:b\r\n%c:d", []pair{{"a", "b"}}},
		{"% a:b\r\n c:d", []pair{{"c", "d"}}},
		{"---\r\na:b\r\n", []pair{{"a", "b"}}},
		{"---\r\na:b\r\n--\r\nc:d", []pair{{"a", "b"}, {"c", "d"}}},
		{"---\r\na:b\r\n---\r\nc:d", []pair{{"a", "b"}}},
		{"---\r\na:b\r\n----\r\nc:d", []pair{{"a", "b"}}},
		{"new-title:\nnew-url:", []pair{{"new-title", ""}, {"new-url", ""}}},
	}
	for i, tc := range testcases {
		meta := parseMetaStr(tc.input)
		if got := iter2pairs(meta.All()); !equalPairs(tc.exp, got) {
			t.Errorf("TC=%d: expected=%v, got=%v", i, tc.exp, got)
		}
	}

	// Test, whether input position is correct.
	inp := input.NewInput([]byte("---\na:b\n---\nX"))
	m := meta.NewFromInput(testID, inp)
	exp := []pair{{"a", "b"}}
	if got := iter2pairs(m.All()); !equalPairs(exp, got) {
		t.Errorf("Expected=%v, got=%v", exp, got)
	}
	expCh := 'X'
	if gotCh := inp.Ch; gotCh != expCh {
		t.Errorf("Expected=%v, got=%v", expCh, gotCh)
	}
}

type pair struct {
	key meta.Key
	val meta.Value
}

func iter2pairs(it iter.Seq2[meta.Key, meta.Value]) (result []pair) {
	it(func(key meta.Key, val meta.Value) bool {
		result = append(result, pair{key, val})
		return true
	})
	return result
}

func equalPairs(one, two []pair) bool {
	if len(one) != len(two) {
		return false
	}
	for i := range len(one) {
		if one[i].key != two[i].key || one[i].val != two[i].val {
			return false
		}
	}
	return true
}

func TestPrecursorIDSet(t *testing.T) {
	t.Parallel()
	var testdata = []struct {
		inp string
		exp meta.Value
	}{
		{"", ""},
		{"123", ""},
		{"12345678901234", "12345678901234"},
		{"123 12345678901234", "12345678901234"},
		{"12345678901234 123", "12345678901234"},
		{"01234567890123 123 12345678901234", "01234567890123 12345678901234"},
		{"12345678901234 01234567890123", "01234567890123 12345678901234"},
	}
	for i, tc := range testdata {
		m := parseMetaStr(meta.KeyPrecursor + ": " + tc.inp)
		if got, ok := m.Get(meta.KeyPrecursor); (!ok && tc.exp != "") || tc.exp != got {
			t.Errorf("TC=%d: expected %q, but got %q when parsing %q", i, tc.exp, got, tc.inp)
		}
	}
}
