//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package text_test

import (
	"bytes"
	"testing"

	"github.com/t73fde/sxpf"
	"zettelstore.de/c/sexpr"
	"zettelstore.de/c/text"
)

func TestSexprText(t *testing.T) {
	testcases := []struct {
		src string
		exp string
	}{
		{"[]", ""},
		{`[TEXT "a"]`, "a"},
		{`[SPACE "a"]`, " "},
	}
	for i, tc := range testcases {
		sval, err := sxpf.ReadString(sexpr.Smk, tc.src)
		if err != nil {
			t.Error(err)
			continue
		}
		seq, ok := sval.(sxpf.Sequence)
		if !ok {
			t.Errorf("%d: not a list: %v", i, sval)
		}
		var buf bytes.Buffer
		text.SEncodeBlock(&buf, seq)
		got := buf.String()
		if got != tc.exp {
			t.Errorf("%d: EncodeBlock(%q) == %q, but got %q", i, tc.src, tc.exp, got)
		}
	}
}
