//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr_test

import (
	"testing"

	"zettelstore.de/c/sexpr"
)

func TestReadString(t *testing.T) {
	testcases := []struct {
		src string
		exp string
	}{
		{"a", "A"},
		{`""`, `""`},
		{`"a"`, `"a"`},
		{"()", "()"},
		{"(a)", "(A)"},
		{"((a))", "((A))"},
		{"(a b c)", "(A B C)"},
		{`("a" b "c")`, `("a" B "c")`},
		{"(A ((b c) d) (e f))", "(A ((B C) D) (E F))"},
	}
	for i, tc := range testcases {
		val, err := sexpr.ReadString(tc.src)
		if err != nil {
			t.Errorf("%d: ReadString(%q) resulted in error: %v", i, tc.src, err)
			continue
		}
		if val == nil {
			t.Errorf("%d: ReadString(%q) resulted in nil value", i, tc.src)
			continue
		}
		got := val.String()
		if tc.exp != got {
			t.Errorf("%d: ReadString(%q) should return %q, but got %q", i, tc.src, tc.exp, got)
		}
	}
}
