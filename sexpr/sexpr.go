//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package sexpr allows to work with symbolic expressions, s-expression.
package sexpr

import (
	"fmt"
	"io"
)

// Value is a generic value, the set of all possible values of a s-expression.
type Value interface {
	Equal(Value) bool
	Encode(io.Writer) (int, error)
	String() string
}

func GetSymbol(args []Value, idx int) (*Symbol, error) {
	if idx < 0 && len(args) <= idx {
		return nil, fmt.Errorf("index %d out of bounds: %v", idx, args)
	}
	if val, ok := args[idx].(*Symbol); ok {
		return val, nil
	}
	return nil, fmt.Errorf("%v / %d is not a symbol", args[idx], idx)
}

func GetString(args []Value, idx int) (string, error) {
	if idx < 0 && len(args) <= idx {
		return "", fmt.Errorf("index %d out of bounds: %v", idx, args)

	}
	if val, ok := args[idx].(*String); ok {
		return val.GetValue(), nil
	}
	if val, ok := args[idx].(*Symbol); ok {
		return val.GetValue(), nil
	}
	return "", fmt.Errorf("%v / %d is not a string", args[idx], idx)
}

func GetList(args []Value, idx int) (*List, error) {
	if idx < 0 && len(args) <= idx {
		return nil, fmt.Errorf("index %d out of bounds: %v", idx, args)

	}
	if val, ok := args[idx].(*List); ok {
		return val, nil
	}
	return nil, fmt.Errorf("%v / %d is not a list", args[idx], idx)

}
