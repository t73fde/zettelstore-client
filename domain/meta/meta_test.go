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

package meta

import (
	"iter"
	"slices"
	"strings"
	"testing"

	"t73f.de/r/zsc/domain/id"
)

const testID = id.Zid(98765432101234)

func TestKeyIsValid(t *testing.T) {
	t.Parallel()
	validKeys := []string{"0", "a", "0-", "title", "title-----", strings.Repeat("r", 255)}
	for _, key := range validKeys {
		if !KeyIsValid(key) {
			t.Errorf("Key %q wrongly identified as invalid key", key)
		}
	}
	invalidKeys := []string{"", "-", "-a", "Title", "a_b", strings.Repeat("e", 256)}
	for _, key := range invalidKeys {
		if KeyIsValid(key) {
			t.Errorf("Key %q wrongly identified as valid key", key)
		}
	}
}

func TestTitleHeader(t *testing.T) {
	t.Parallel()
	m := New(testID)
	if got, ok := m.Get(KeyTitle); ok && got != "" {
		t.Errorf("Title is not empty, but %q", got)
	}
	addToMeta(m, KeyTitle, " ")
	if got, ok := m.Get(KeyTitle); ok && got != "" {
		t.Errorf("Title is not empty, but %q", got)
	}
	const st = "A simple text"
	addToMeta(m, KeyTitle, " "+st+"  ")
	if got, ok := m.Get(KeyTitle); !ok || got != st {
		t.Errorf("Title is not %q, but %q", st, got)
	}
	addToMeta(m, KeyTitle, "  "+st+"\t")
	const exp = st + " " + st
	if got, ok := m.Get(KeyTitle); !ok || got != exp {
		t.Errorf("Title is not %q, but %q", exp, got)
	}

	m = New(testID)
	const at = "A Title"
	addToMeta(m, KeyTitle, at)
	addToMeta(m, KeyTitle, " ")
	if got, ok := m.Get(KeyTitle); !ok || got != at {
		t.Errorf("Title is not %q, but %q", at, got)
	}
}

func checkTags(t *testing.T, exp []string, m *Meta) {
	t.Helper()
	got := slices.Collect(m.GetFields(KeyTags))
	for i, tag := range exp {
		if i < len(got) {
			if tag != got[i] {
				t.Errorf("Pos=%d, expected %q, got %q", i, exp[i], got[i])
			}
		} else {
			t.Errorf("Expected %q, but is missing", exp[i])
		}
	}
	if len(exp) < len(got) {
		t.Errorf("Extra tags: %q", got[len(exp):])
	}
}

func TestTagsHeader(t *testing.T) {
	t.Parallel()
	m := New(testID)
	checkTags(t, []string{}, m)

	addToMeta(m, KeyTags, "")
	checkTags(t, []string{}, m)

	addToMeta(m, KeyTags, "  #t1 #t2  #t3 #t4  ")
	checkTags(t, []string{"#t1", "#t2", "#t3", "#t4"}, m)

	addToMeta(m, KeyTags, "#t5")
	checkTags(t, []string{"#t1", "#t2", "#t3", "#t4", "#t5"}, m)

	addToMeta(m, KeyTags, "t6")
	checkTags(t, []string{"#t1", "#t2", "#t3", "#t4", "#t5"}, m)
}

func TestSyntax(t *testing.T) {
	t.Parallel()
	m := New(testID)
	if got, ok := m.Get(KeySyntax); ok || got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	addToMeta(m, KeySyntax, " ")
	if got, _ := m.Get(KeySyntax); got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
	addToMeta(m, KeySyntax, "MarkDown")
	const exp = "markdown"
	if got, ok := m.Get(KeySyntax); !ok || got != exp {
		t.Errorf("Syntax is not %q, but %q", exp, got)
	}
	addToMeta(m, KeySyntax, " ")
	if got, _ := m.Get(KeySyntax); got != "" {
		t.Errorf("Syntax is not %q, but %q", "", got)
	}
}

func checkHeader(t *testing.T, exp map[string]string, gotI iter.Seq2[Key, Value]) {
	t.Helper()
	got := make(map[string]string)
	gotI(func(key Key, val Value) bool {
		got[key] = string(val)
		if _, ok := exp[key]; !ok {
			t.Errorf("Key %q is not expected, but has value %q", key, val)
		}
		return true
	})
	for k, v := range exp {
		if gv, ok := got[k]; !ok || v != gv {
			if ok {
				t.Errorf("Key %q is not %q, but %q", k, v, got[k])
			} else {
				t.Errorf("Key %q missing, should have value %q", k, v)
			}
		}
	}
}

func TestDefaultHeader(t *testing.T) {
	t.Parallel()
	m := New(testID)
	addToMeta(m, "h1", "d1")
	addToMeta(m, "H2", "D2")
	addToMeta(m, "H1", "D1.1")
	exp := map[string]string{"h1": "d1 D1.1", "h2": "D2"}
	checkHeader(t, exp, m.All())
	addToMeta(m, "", "d0")
	checkHeader(t, exp, m.All())
	addToMeta(m, "h3", "")
	exp["h3"] = ""
	checkHeader(t, exp, m.All())
	addToMeta(m, "h3", "  ")
	checkHeader(t, exp, m.All())
	addToMeta(m, "h4", " ")
	exp["h4"] = ""
	checkHeader(t, exp, m.All())
}

func TestDelete(t *testing.T) {
	t.Parallel()
	m := New(testID)
	m.Set("key", "val")
	if got, ok := m.Get("key"); !ok || got != "val" {
		t.Errorf("Value != %q, got: %v/%q", "val", ok, got)
	}
	m.Set("key", "")
	if got, ok := m.Get("key"); !ok || got != "" {
		t.Errorf("Value != %q, got: %v/%q", "", ok, got)
	}
	m.Delete("key")
	if got, ok := m.Get("key"); ok || got != "" {
		t.Errorf("Value != %q, got: %v/%q", "", ok, got)
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		pairs1, pairs2 []string
		allowComputed  bool
		exp            bool
	}{
		{nil, nil, true, true},
		{nil, nil, false, true},
		{[]string{"a", "a"}, nil, false, false},
		{[]string{"a", "a"}, nil, true, false},
		{[]string{KeyFolge, "0"}, nil, true, false},
		{[]string{KeyFolge, "0"}, nil, false, true},
		{[]string{KeyFolge, "0"}, []string{KeyFolge, "0"}, true, true},
		{[]string{KeyFolge, "0"}, []string{KeyFolge, "0"}, false, true},
	}
	for i, tc := range testcases {
		m1 := pairs2meta(tc.pairs1)
		m2 := pairs2meta(tc.pairs2)
		got := m1.Equal(m2, tc.allowComputed)
		if tc.exp != got {
			t.Errorf("%d: %v =?= %v: expected=%v, but got=%v", i, tc.pairs1, tc.pairs2, tc.exp, got)
		}
		got = m2.Equal(m1, tc.allowComputed)
		if tc.exp != got {
			t.Errorf("%d: %v =!= %v: expected=%v, but got=%v", i, tc.pairs1, tc.pairs2, tc.exp, got)
		}
	}

	// Pathologic cases
	var m1, m2 *Meta
	if !m1.Equal(m2, true) {
		t.Error("Nil metas should be treated equal")
	}
	m1 = New(testID)
	if m1.Equal(m2, true) {
		t.Error("Empty meta should not be equal to nil")
	}
	if m2.Equal(m1, true) {
		t.Error("Nil meta should should not be equal to empty")
	}
	m2 = New(testID + 1)
	if m1.Equal(m2, true) {
		t.Error("Different ID should differentiate")
	}
	if m2.Equal(m1, true) {
		t.Error("Different ID should differentiate")
	}
}

func pairs2meta(pairs []string) *Meta {
	m := New(testID)
	for i := 0; i < len(pairs); i += 2 {
		m.Set(pairs[i], Value(pairs[i+1]))
	}
	return m
}

func TestRemoveNonGraphic(t *testing.T) {
	testCases := []struct {
		inp string
		exp string
	}{
		{"", ""},
		{" ", ""},
		{"a", "a"},
		{"a ", "a"},
		{"a b", "a b"},
		{"\n", ""},
		{"a\n", "a"},
		{"a\nb", "a b"},
		{"a\tb", "a b"},
	}
	for i, tc := range testCases {
		got := RemoveNonGraphic(tc.inp)
		if tc.exp != got {
			t.Errorf("%q/%d: expected %q, but got %q", tc.inp, i, tc.exp, got)
		}
	}
}
