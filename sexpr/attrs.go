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
		p, ok := elem.Head().(*sxpf.List)
		if !ok || p == nil {
			continue
		}
		key := p.Head()
		if !sxpf.IsAtom(key) {
			continue
		}
		var val string
		q := p.Tail()
		if q != nil {
			v := q.Head()
			if !sxpf.IsAtom(v) {
				continue
			}
			val = v.String()
		}
		result = result.Set(key.String(), val)
	}
	return result
}
