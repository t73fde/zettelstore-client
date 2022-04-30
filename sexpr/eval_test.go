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

func (e *testEnv) Continue() (sexpr.Value, error) { return nil, nil }

func (e *testEnv) Lookup(sym *sexpr.Symbol) (sexpr.PrimitiveFn, bool, bool) {
	switch sym.GetValue() {
	case "CAT":
		return func(env sexpr.Environment, args []sexpr.Value) (sexpr.Value, error) {
			var buf bytes.Buffer
			for _, arg := range args {
				buf.WriteString(arg.String())
			}
			return sexpr.NewString(buf.String()), nil
		}, false, true
	case "QUOTE":
		return func(env sexpr.Environment, args []sexpr.Value) (sexpr.Value, error) {
			return sexpr.NewList(args...), nil
		}, true, true
	}
	return nil, false, false
}

func (e *testEnv) EvaluateSymbol(sym *sexpr.Symbol) (sexpr.Value, error) {
	return sym, nil
}
func (e *testEnv) EvaluateString(str *sexpr.String) (sexpr.Value, error) { return str, nil }
func (e *testEnv) EvaluateList(lst *sexpr.List) (sexpr.Value, error) {
	res, err, done := sexpr.EvaluateCall(e, lst.GetValue())
	if done {
		return res, err
	}
	return sexpr.EvaluateSlice(e, lst.GetValue())
}
