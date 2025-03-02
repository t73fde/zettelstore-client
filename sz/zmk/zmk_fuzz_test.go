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

package zmk_test

import (
	"testing"

	"t73f.de/r/zsc/input"
	"t73f.de/r/zsc/sz/zmk"
)

func FuzzParseBlocks(f *testing.F) {
	f.Fuzz(func(t *testing.T, src []byte) {
		t.Parallel()
		inp := input.NewInput(src)
		zmk.Parse(inp)
	})
}
