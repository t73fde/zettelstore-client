//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package text_test

import (
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxreader"
	"t73f.de/r/zsc/text"
)

func TestSzText(t *testing.T) {
	testcases := []struct {
		src string
		exp string
	}{
		{"()", ""},
		{`(INLINE (TEXT "a"))`, "a"},
		{`(INLINE (TEXT " "))`, " "},
		{`(INLINE (TEXT "  "))`, " "},
	}
	for i, tc := range testcases {
		sval, err := sxreader.MakeReader(strings.NewReader(tc.src)).Read()
		if err != nil {
			t.Error(err)
			continue
		}
		seq, isPair := sx.GetPair(sval)
		if !isPair {
			t.Errorf("%d: not a list: %v", i, sval)
		}
		got := text.EvaluateInlineString(seq)
		if got != tc.exp {
			t.Errorf("%d: EncodeBlock(%q) == %q, but got %q", i, tc.src, tc.exp, got)
		}
	}
}
