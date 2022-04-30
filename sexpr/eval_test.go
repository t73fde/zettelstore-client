//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr_test

import (
	"bytes"
	"testing"

	"zettelstore.de/c/sexpr"
)

func TestEvaluate(t *testing.T) {
	testcases := []struct {
		src string
		exp string
	}{
		{"a", "A"},
		{`"a"`, `"a"`},
		{"(CAT a b)", `"AB"`},
		{"(QUOTE (A b) c)", "((A B) C)"},
	}
	env := testEnv{}
	for i, tc := range testcases {
		expr, err := sexpr.ReadString(tc.src)
		if err != nil {
			t.Error(err)
			continue
		}
		val, err := sexpr.Evaluate(&env, expr)
		if err != nil {
			t.Error(err)
			continue
		}
		got := val.String()
		if got != tc.exp {
			t.Errorf("%d: %v should evaluate to %v, but got: %v", i, tc.src, tc.exp, got)
		}
	}
}

type testEnv struct{}

var testFns = []*sexpr.Function{
	sexpr.NewPrimitive(
		"CAT",
		false,
		func(env sexpr.Environment, args []sexpr.Value) (sexpr.Value, error) {
			var buf bytes.Buffer
			for _, arg := range args {
				buf.WriteString(arg.String())
			}
			return sexpr.NewString(buf.String()), nil
		},
	),
	sexpr.NewPrimitive(
		"QUOTE",
		true,
		func(env sexpr.Environment, args []sexpr.Value) (sexpr.Value, error) {
			return sexpr.NewList(args...), nil
		},
	),
}

var testFnMap = map[string]*sexpr.Function{}

func init() {
	for _, fn := range testFns {
		testFnMap[fn.Name()] = fn
	}
}

func (e *testEnv) EvaluateSymbol(sym *sexpr.Symbol) (sexpr.Value, error) {
	if fn, found := testFnMap[sym.GetValue()]; found {
		return fn, nil
	}
	return sym, nil
}

func (e *testEnv) EvaluateString(str *sexpr.String) (sexpr.Value, error) { return str, nil }
func (e *testEnv) EvaluateList(lst *sexpr.List) (sexpr.Value, error) {
	vals := lst.GetValue()
	res, err, done := sexpr.EvaluateCall(e, vals)
	if done {
		return res, err
	}
	result, err := sexpr.EvaluateSlice(e, vals)
	if err != nil {
		return nil, err
	}
	return sexpr.NewList(result...), nil
}
