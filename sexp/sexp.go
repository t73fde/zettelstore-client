//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package sexp contains helper function to work with s-expression in an alien
// environment.
package sexp

import (
	"errors"
	"fmt"

	"zettelstore.de/sx.fossil"
)

// ParseList parses the given object as a proper list, based on a type specification.
func ParseList(obj sx.Object, spec string) ([]sx.Object, error) {
	pair, isPair := sx.GetPair(obj)
	if !isPair {
		return nil, fmt.Errorf("not a list: %T/%v", obj, obj)
	}
	if pair == nil {
		if spec == "" {
			return nil, nil
		}
		return nil, ErrElementsMissing
	}

	result := make([]sx.Object, 0, len(spec))
	node, i := pair, 0
	for ; node != nil; i++ {
		if i >= len(spec) {
			return nil, ErrNoSpec
		}
		var val sx.Object
		var ok bool
		car := node.Car()
		switch spec[i] {
		case 'b':
			val, ok = sx.GetBoolean(car)
		case 'i':
			val, ok = car.(sx.Int64)
		case 'o':
			val, ok = car, true
		case 'p':
			val, ok = sx.GetPair(car)
		case 's':
			val, ok = sx.GetString(car)
		case 'y':
			val, ok = sx.GetSymbol(car)
		default:
			return nil, fmt.Errorf("unknown spec '%c'", spec[i])
		}
		if !ok {
			return nil, fmt.Errorf("does not match spec '%v': %v", spec[i], car)
		}
		result = append(result, val)
		next, isNextPair := sx.GetPair(node.Cdr())
		if !isNextPair {
			return nil, sx.ErrImproper{Pair: pair}
		}
		node = next
	}
	if i < len(spec) {
		return nil, ErrElementsMissing
	}
	return result, nil
}

var ErrElementsMissing = errors.New("spec contains more data")
var ErrNoSpec = errors.New("no spec for elements")
