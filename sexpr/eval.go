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

import "fmt"

// PrimitiveFn is a primitve function that is implemented in Go.
type PrimitiveFn func(Environment, []Value) (Value, error)

// Lookuper can look-up symbols and return a function plus indication of special value.
type Lookuper interface {
	Lookup(*Symbol) (PrimitiveFn, bool, bool)
}

// Environment provides data to evaluate a s-expression.
type Environment interface {
	Lookuper

	Continue() (Value, error)
	EvaluateString(*String) (Value, error)
	EvaluateSymbol(*Symbol) (Value, error)
	EvaluateList(*List) (Value, error)
}

func Evaluate(env Environment, value Value) (Value, error) {
	if res, err := env.Continue(); res != nil || err != nil {
		return res, err
	}
	switch val := value.(type) {
	case *Symbol:
		return env.EvaluateSymbol(val)
	case *String:
		return env.EvaluateString(val)
	case *List:
		return env.EvaluateList(val)
	}
	return nil, nil // error
}

func EvaluateCall(env Environment, vals []Value) (Value, error, bool) {
	if len(vals) == 0 {
		return nil, nil, false
	}
	if sym, ok := vals[0].(*Symbol); ok {
		fn, primitive, found := env.Lookup(sym)
		if !found {
			return nil, fmt.Errorf("unbound identifier: %q", sym.GetValue()), true
		}
		params := vals[1:]
		if !primitive {
			args := make([]Value, len(params))
			for i, param := range params {
				val, err := Evaluate(env, param)
				if val == nil || err != nil {
					return nil, err, true
				}
				args[i] = val
			}
			params = args
		}
		res, err := fn(env, params)
		return res, err, true
	}
	return nil, nil, false
}

func EvaluateSlice(env Environment, vals []Value) (res *List, err error) {
	result := make([]Value, len(vals))
	for i, value := range vals {
		result[i], err = Evaluate(env, value)
		if err != nil {
			return nil, err
		}
	}
	return NewList(result...), nil
}
