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

	"t73f.de/r/zsc/sz"
	"t73f.de/r/zsx/input"
)

func TestParseNone(t *testing.T) {
	if got := sz.ParseNoneBlocks(nil); got != nil {
		t.Error("GOTB", got)
	}

	inp := input.NewInput([]byte("1234\n6789"))
	if got := sz.ParseNoneBlocks(inp); got != nil {
		t.Error("GOTI", got)
	}
}

func TestParsePlani(t *testing.T) {
	testcases := []struct {
		src       string
		syntax    string
		expBlocks string
	}{
		{"abc", "html", "(VERBATIM-HTML ((\"\" . \"html\")) \"abc\")"},
		{"abc\ndef", "html", "(VERBATIM-HTML ((\"\" . \"html\")) \"abc\\ndef\")"},
		{"abc", "text", "(VERBATIM-CODE ((\"\" . \"text\")) \"abc\")"},
		{"abc\nDEF", "text", "(VERBATIM-CODE ((\"\" . \"text\")) \"abc\\nDEF\")"},
	}
	for i, tc := range testcases {
		t.Run(tc.syntax+":"+tc.src, func(t *testing.T) {
			inp := input.NewInput([]byte(tc.src))
			if got := sz.ParsePlainBlocks(inp, tc.syntax).String(); tc.expBlocks != got {
				t.Errorf("%d: %q/%v\nexpected: %q\ngot     : %q", i, tc.src, tc.syntax, tc.expBlocks, got)
			}
		})
	}
}
