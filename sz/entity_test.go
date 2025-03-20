//-----------------------------------------------------------------------------
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
//-----------------------------------------------------------------------------

package sz_test

import (
	"testing"

	"t73f.de/r/zsc/sz"
	"t73f.de/r/zsx/input"
)

func TestScanEntity(t *testing.T) {
	t.Parallel()
	var testcases = []struct {
		text string
		exp  string
	}{
		{"", ""},
		{"a", ""},
		{"&amp;", "&"},
		{"&#33;", "!"},
		{"&#x33;", "3"},
		{"&quot;", "\""},
	}
	for id, tc := range testcases {
		inp := input.NewInput([]byte(tc.text))
		got, ok := sz.ScanEntity(inp)
		if !ok {
			if tc.exp != "" {
				t.Errorf("ID=%d, text=%q: expected error, but got %q", id, tc.text, got)
			}
			if inp.Pos != 0 {
				t.Errorf("ID=%d, text=%q: input position advances to %d", id, tc.text, inp.Pos)
			}
			continue
		}
		if tc.exp != got {
			t.Errorf("ID=%d, text=%q: expected %q, but got %q", id, tc.text, tc.exp, got)
		}
	}
}

func TestScanIllegalEntity(t *testing.T) {
	t.Parallel()
	testcases := []string{"", "a", "& Input &rarr;", "&#9;", "&#x1f;"}
	for i, tc := range testcases {
		got, ok := sz.ScanEntity(input.NewInput([]byte(tc)))
		if ok {
			t.Errorf("%d: scanning %q was unexpected successful, got %q", i, tc, got)
			continue
		}
	}
}
