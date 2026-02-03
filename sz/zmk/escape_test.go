//-----------------------------------------------------------------------------
// Copyright (c) 2026-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2026-present Detlef Stern
//-----------------------------------------------------------------------------

package zmk_test

import (
	"testing"

	"t73f.de/r/zsc/sz/zmk"
)

func TestEscapeZmkSyntax(t *testing.T) {
	testcases := []struct {
		in, exp string
	}{
		{"", ""},
		{"#", "#"}, {"##", `#\#`}, {"###", `#\##`}, {"####", `#\##\#`},
		{`\`, `\`}, {`\\`, `\\\`}, {`\\\`, `\\\\`}, {`\\\\`, `\\\\\\`},
	}
	for _, tc := range testcases {
		t.Run(tc.in, func(t *testing.T) {
			got := zmk.EscapeZmkSyntax(tc.in)
			if got != tc.exp {
				t.Errorf("expected %q, but got %q", tc.exp, got)
			}
		})
	}
}
