//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package html_test

import (
	"bytes"
	"testing"

	"zettelstore.de/c/html"
)

func TestEscape(t *testing.T) {
	testcases := []struct {
		in, exp string
	}{
		{"", ""},
		{"<", "&lt;"},
	}
	for _, tc := range testcases {
		var buf bytes.Buffer
		_, err := html.Escape(&buf, tc.in)
		if err != nil {
			t.Errorf("Escape(%q) got error: %v", tc.in, err)
		}
		if got := buf.String(); tc.exp != got {
			t.Errorf("Escape(%q) == %q, but got %q", tc.in, tc.exp, got)
		}
	}
}

func TestEscapeVisible(t *testing.T) {
	testcases := []struct {
		in, exp string
	}{
		{"", ""},
		{"<", "&lt;"},
		{" a  b ", "␣a␣␣b␣"},
	}
	for _, tc := range testcases {
		var buf bytes.Buffer
		_, err := html.EscapeVisible(&buf, tc.in)
		if err != nil {
			t.Errorf("Escape(%q) got error: %v", tc.in, err)
		}
		if got := buf.String(); tc.exp != got {
			t.Errorf("Escape(%q) == %q, but got %q", tc.in, tc.exp, got)
		}
	}
}
