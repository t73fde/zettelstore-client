//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sz_test

import (
	"testing"

	"t73f.de/r/zsc/input"
	"t73f.de/r/zsc/sz"
)

func TestParseNone(t *testing.T) {
	if got := sz.ParseNoneBlocks(nil); got != nil {
		t.Error("GOTB", got)
	}

	inp := input.NewInput([]byte("1234\n6789"))
	if got := sz.ParseNoneInlines(inp); got != nil {
		t.Error("GOTI", got)
	}
	if got := inp.Pos; got != 4 {
		t.Errorf("input should be on position 4, but is %d", got)
	}
	if got := inp.Ch; got != '\n' {
		t.Errorf("input character should be 10, but is %d", got)
	}
}

func TestParsePlani(t *testing.T) {
	testcases := []struct {
		src        string
		syntax     string
		expBlocks  string
		expInlines string
	}{
		{"abc", "html",
			"(VERBATIM-HTML ((\"\" . \"html\")) \"abc\")",
			"(LITERAL-HTML ((\"\" . \"html\")) \"abc\")"},
		{"abc\ndef", "html",
			"(VERBATIM-HTML ((\"\" . \"html\")) \"abc\\ndef\")",
			"(LITERAL-HTML ((\"\" . \"html\")) \"abc\")"},
		{"abc", "text",
			"(VERBATIM-CODE ((\"\" . \"text\")) \"abc\")",
			"(LITERAL-CODE ((\"\" . \"text\")) \"abc\")"},
		{"abc\nDEF", "text",
			"(VERBATIM-CODE ((\"\" . \"text\")) \"abc\\nDEF\")",
			"(LITERAL-CODE ((\"\" . \"text\")) \"abc\")"},
	}
	for i, tc := range testcases {
		t.Run(tc.syntax+":"+tc.src, func(t *testing.T) {
			inp := input.NewInput([]byte(tc.src))
			if got := sz.ParsePlainBlocks(inp, tc.syntax).String(); tc.expBlocks != got {
				t.Errorf("%d: %q/%v\nexpected: %q\ngot     : %q", i, tc.src, tc.syntax, tc.expBlocks, got)
			}
			inp.SetPos(0)
			if got := sz.ParsePlainInlines(inp, tc.syntax).String(); tc.expInlines != got {
				t.Errorf("%d: %q/%v\nexpected: %q\ngot     : %q", i, tc.src, tc.syntax, tc.expInlines, got)
			}
		})
	}
}
