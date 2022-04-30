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

// Environment provides data to evaluate a s-expression.
type Environment interface {
	EvaluateString(*String) (Value, error)
	EvaluateSymbol(*Symbol) (Value, error)
	EvaluateList(*List) (Value, error)
}

func Evaluate(env Environment, value Value) (Value, error) {
	switch val := value.(type) {
	case *Symbol:
		return env.EvaluateSymbol(val)
	case *String:
		return env.EvaluateString(val)
	case *List:
		return env.EvaluateList(val)
	default:
		// Other types evaluate to themself
		return value, nil
	}
}

func EvaluateCall(env Environment, vals []Value) (Value, error, bool) {
	if len(vals) == 0 {
		return nil, nil, false
	}
	if sym, ok := vals[0].(*Symbol); ok {
		sval, err := env.EvaluateSymbol(sym)
		if err != nil {
			return nil, err, true
		}
		fn, ok := sval.(*Function)
		if !ok {
			return nil, fmt.Errorf("unbound identifier: %q", sym.GetValue()), true
		}
		params := vals[1:]
		if !fn.IsSpecial() {
			var err error
			params, err = EvaluateSlice(env, params)
			if err != nil {
				return nil, err, true
			}
		}
		res, err := fn.Call(env, params)
		return res, err, true
	}
	return nil, nil, false
}

func EvaluateSlice(env Environment, vals []Value) (res []Value, err error) {
	res = make([]Value, len(vals))
	for i, value := range vals {
		res[i], err = Evaluate(env, value)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
