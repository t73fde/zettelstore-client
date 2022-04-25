//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package sexpr allows to work with symbolic expressions, s-expression.
package sexpr_test

import (
	"bytes"
	"testing"

	"zettelstore.de/c/sexpr"
)

func TestSymbol(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		val string
		ok  bool
		exp string
	}{
		{"", false, ""},
		{"a", true, "A"},
	}
	for i, tc := range testcases {
		s := sexpr.GetSymbol(tc.val)
		if (s != nil) != tc.ok {
			if s == nil {
				t.Errorf("%d: GetSymbol(%q) must not be nil, but is", i, tc.val)
			} else {
				t.Errorf("%d: GetSymbol(%q) must be nil, but is not: %q", i, tc.val, s.GetValue())
			}
			continue
		}
		if s == nil {
			continue
		}
		got := s.GetValue()
		if tc.exp != got {
			t.Errorf("%d: GetValue(%q) != %q, but got %q", i, tc.val, tc.exp, got)
		}
		if !s.Equal(s) {
			t.Errorf("%d: %q is not equal to itself", i, got)
		}

		s2 := sexpr.GetSymbol(tc.val)
		if s2 != s {
			t.Errorf("%d: GetSymbol(%q) produces different values if called multiple times", i, tc.val)
		}
	}
}

func FuzzSymbol(f *testing.F) {
	f.Fuzz(func(t *testing.T, in string) {
		t.Parallel()
		s := sexpr.GetSymbol(in)
		if !s.Equal(s) {
			if s == nil {
				t.Errorf("nil symbol is not equal to itself")
			} else {
				t.Errorf("%q is not equal to itself", s.GetValue())
			}
		}
	})
}

func TestStringEncode(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		val string
		exp string
	}{
		{"", ""},
		{"a", "a"},
		{"\n", "\\n"},
	}
	for i, tc := range testcases {
		var buf bytes.Buffer
		s := sexpr.NewString(tc.val)
		if s == nil {
			t.Errorf("%d: NewString(%q) == nil", i, tc.val)
			continue
		}
		sVal := s.GetValue()
		if sVal != tc.val {
			t.Errorf("%d: NewString(%q) changed value to %q", i, tc.val, sVal)
			continue
		}
		length, err := s.Encode(&buf)
		if err != nil {
			t.Errorf("%d: Encode(%q) -> %v", i, tc.val, err)
			continue
		}
		got := buf.String()
		if length < 2 {
			t.Errorf("%d: Encode(%q).Length < 2: %q (%d)", i, tc.val, got, length)
			continue
		}
		exp := "\"" + tc.exp + "\""
		if got != exp {
			t.Errorf("%d: Encode(%q) expected %q, but got %q", i, tc.val, exp, got)
		}
	}
}
