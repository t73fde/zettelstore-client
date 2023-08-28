//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sz_test

import (
	"strconv"
	"testing"

	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

func BenchmarkInitZttlSyms(b *testing.B) {
	for i := 0; i < 20; i++ {
		b.Run("hint-"+strconv.Itoa(i), func(b *testing.B) {
			sf := sx.MakeMappedFactory(1 << i)
			for i := 0; i < b.N; i++ {
				var zs sz.ZettelSymbols
				zs.InitializeZettelSymbols(sf)
			}
		})
	}
}
