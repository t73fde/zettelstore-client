//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
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
func GetAttributes(seq *sxpf.List) (result attrs.Attributes) {
	for elem := seq; elem != nil; elem = elem.Tail() {
		p, ok := elem.Head().(*sxpf.Pair)
		if !ok {
			continue
		}
		key, ok := p.Car().(*sxpf.Symbol)
		if !ok {
			continue
		}
		val := p.Cdr()
		switch val.(type) {
		case *sxpf.Symbol:
		case sxpf.String:
		case sxpf.Keyword:
		case *sxpf.Number:
		default:
			continue
		}
		result = result.Set(key.String(), val.String())
	}
	return result
}
