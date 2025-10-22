// -----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
// -----------------------------------------------------------------------------

package sz_test

import (
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/sz"
	"t73f.de/r/zsx"
)

func TestParseReference(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		s   string
		exp string
	}{
		{"", `(INVALID "")`},
		{"abc", `(HOSTED "abc")`},
		{"abc def", `(INVALID "abc def")`},
		{"/hosted", `(HOSTED "/hosted")`},
		{"/hosted ref", `(INVALID "/hosted ref")`},
		{"./", `(HOSTED "./")`},
		{"./12345678901234", `(HOSTED "./12345678901234")`},
		{"../", `(HOSTED "../")`},
		{"../12345678901234", `(HOSTED "../12345678901234")`},
		{"abc#frag", `(HOSTED "abc#frag")`},
		{"abc#frag space", `(INVALID "abc#frag space")`},
		{"abc#", `(INVALID "abc#")`},
		{"abc# ", `(INVALID "abc# ")`},
		{"/hosted#frag", `(HOSTED "/hosted#frag")`},
		{"./#frag", `(HOSTED "./#frag")`},
		{"./12345678901234#frag", `(HOSTED "./12345678901234#frag")`},
		{"../#frag", `(HOSTED "../#frag")`},
		{"../12345678901234#frag", `(HOSTED "../12345678901234#frag")`},
		{"#frag", `(SELF "#frag")`},
		{"#", `(INVALID "#")`},
		{"# ", `(INVALID "# ")`},
		{"https://t73f.de", `(EXTERNAL "https://t73f.de")`},
		{"https://t73f.de/12345678901234", `(EXTERNAL "https://t73f.de/12345678901234")`},
		{"http://t73f.de/1234567890", `(EXTERNAL "http://t73f.de/1234567890")`},
		{"mailto:ds@zettelstore.de", `(EXTERNAL "mailto:ds@zettelstore.de")`},
		{",://", `(INVALID ",://")`},

		// ZS specific
		{"00000000000000", `(INVALID "00000000000000")`},
		{"00000000000000#frag", `(INVALID "00000000000000#frag")`},
		{"12345678901234", `(ZETTEL "12345678901234")`},
		{"12345678901234#frag", `(ZETTEL "12345678901234#frag")`},
		{"12345678901234#", `(INVALID "12345678901234#")`},
		{"12345678901234# space", `(INVALID "12345678901234# space")`},
		{"12345678901234#frag ", `(INVALID "12345678901234#frag ")`},
		{"12345678901234#frag space", `(INVALID "12345678901234#frag space")`},
		{"query:role:zettel LIMIT 13", `(QUERY "role:zettel LIMIT 13")`},
		{"//based", `(BASED "/based")`},
		{"//based#frag", `(BASED "/based#frag")`},
		{"//based#", `(INVALID "//based#")`},
	}
	for _, tc := range testcases {
		t.Run(tc.s, func(t *testing.T) {
			if got := sz.ScanReference(tc.s); got.String() != tc.exp {
				t.Errorf("%q should be %q, but got %q", tc.s, tc.exp, got)
			}
		})
	}
}

func TestWriteReference(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		src *sx.Pair
		exp string
	}{
		{nil, ""},
		{zsx.MakeReference(sz.SymRefStateZettel, "12345678901234"), "12345678901234"},
		{zsx.MakeReference(sz.SymRefStateQuery, "12345678901234"), "query:12345678901234"},
		{zsx.MakeReference(sz.SymRefStateBased, "/based"), "//based"},
		{zsx.MakeReference(zsx.SymRefStateHosted, "/hosted"), "/hosted"},
	}
	for _, tc := range testcases {
		t.Run(tc.src.String(), func(t *testing.T) {
			var sb strings.Builder
			err := sz.WriteReference(&sb, tc.src)
			if err != nil {
				t.Error(err)
				return
			}
			if got := sb.String(); got != tc.exp {
				t.Errorf("expect %q, but got %q", tc.exp, got)
			}
			if got := sz.ReferenceString(tc.src); got != tc.exp {
				t.Errorf("expect %q, but got %q", tc.exp, got)
			}
			if tc.src != nil {
				if got := sz.ScanReference(tc.exp); !got.IsEqual(tc.src) {
					t.Errorf("expect %v, but got %v", tc.src, got)
				}
			}
		})
	}
}
