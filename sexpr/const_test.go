//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
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

	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/sexpr"
)

func BenchmarkInitializeZettelSymbols(b *testing.B) {
	sf := sxpf.MakeMappedFactory()
	for i := 0; i < b.N; i++ {
		var zs sexpr.ZettelSymbols
		zs.InitializeZettelSymbols(sf)
	}
}
