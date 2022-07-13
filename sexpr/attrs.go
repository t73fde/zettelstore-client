//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr

import (
	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/attrs"
)

// GetAttributes traverses a s-expression list and returns an attribute structure.
func GetAttributes(seq *sxpf.Pair) (result attrs.Attributes) {
	for elem := seq; !elem.IsNil(); elem = elem.GetTail() {
		attr, err := elem.GetPair()
		if err != nil {
			continue
		}
		key, err := attr.GetString()
		if err != nil {
			continue
		}
		val, err := attr.GetTail().GetString()
		if err != nil {
			continue
		}
		result = result.Set(key, val)
	}
	return result
}
