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
	"codeberg.org/t73fde/sxpf/eval"
)

// BindOther bind all unbound symbols to a special function that just recursively
// traverses through the argument lists.
func BindOther(env sxpf.Environment, sf sxpf.SymbolFactory) {
	for _, sym := range sf.Symbols() {
		if _, found := env.Resolve(sym); !found {
			env.Bind(sym, eval.MakeSpecial(sym.String(), DoNothing))
		}
	}

}

// DoNothing just traverses though all (sub-) lists.
func DoNothing(env sxpf.Environment, args *sxpf.List) (sxpf.Value, error) {
	for elem := args; elem != nil; elem = elem.Tail() {
		if lst, ok := elem.Head().(*sxpf.List); ok {
			if _, err := eval.Eval(env, lst); err != nil {
				return sxpf.Nil(), err
			}
		}
	}
	return sxpf.Nil(), nil
}
