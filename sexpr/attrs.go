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
	seq = seq.Tail() // TODO: check for "ATTR" symbol as head.
	for elem := seq; elem != nil; elem = elem.Tail() {
		p, ok := elem.Car().(*sxpf.List)
		if !ok || p == nil {
			continue
		}
		key := p.Car()
		if !sxpf.IsAtom(key) {
			continue
		}
		val := p.Cdr()
		if tail, ok := val.(*sxpf.List); ok {
			val = tail.Car()
		}
		if !sxpf.IsAtom(val) {
			continue
		}
		result = result.Set(key.String(), val.String())
	}
	return result
}
