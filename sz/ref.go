//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import "t73f.de/r/sx"

// MakeReference builds a reference node.
func MakeReference(sym *sx.Symbol, val string) *sx.Pair {
	return sx.Cons(sym, sx.Cons(sx.MakeString(val), sx.Nil()))
}

// GetReference returns the reference symbol and value.
func GetReference(ref *sx.Pair) (*sx.Symbol, string) {
	if ref != nil {
		if sym, isSymbol := sx.GetSymbol(ref.Car()); isSymbol {
			val, isString := sx.GetString(ref.Cdr())
			if !isString {
				val, isString = sx.GetString(ref.Tail().Car())
			}
			if isString {
				return sym, val.GetValue()
			}
		}
	}
	return nil, ""
}
