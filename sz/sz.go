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

// Package sz contains zettel data handling as sx expressions.
package sz

import (
	"t73f.de/r/sx"
	"t73f.de/r/zsc/attrs"
)

// GetAttributes traverses a s-expression list and returns an attribute structure.
func GetAttributes(seq *sx.Pair) (result attrs.Attributes) {
	for obj := range seq.Values() {
		pair, isPair := sx.GetPair(obj)
		if !isPair || pair == nil {
			continue
		}
		key := pair.Car()
		if !key.IsAtom() {
			continue
		}
		val := pair.Cdr()
		if tail, isTailPair := sx.GetPair(val); isTailPair {
			val = tail.Car()
		}
		if !val.IsAtom() {
			continue
		}
		result = result.Set(GoValue(key), GoValue(val))
	}
	return result
}

// GoValue returns the string value of the sx.Object suitable for Go processing.
func GoValue(obj sx.Object) string {
	switch o := obj.(type) {
	case sx.String:
		return o.GetValue()
	case *sx.Symbol:
		return o.GetValue()
	}
	return obj.String()
}
