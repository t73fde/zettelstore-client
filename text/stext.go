//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package text provides types, constants and function to work with text output.
package text

import (
	"bytes"
	"io"

	"github.com/t73fde/sxpf"
	"zettelstore.de/c/sexpr"
)

// SEncodeBlock writes the text of the given block list to the given writer.
func SEncodeBlock(w io.Writer, lst *sxpf.Array) {
	env := newTextEnvironment(w)
	env.EvaluateArray(lst)
}

// SEncodeInlineString returns the text content of the given inline list as a string.
func SEncodeInlineString(vals []sxpf.Value) string {
	var buf bytes.Buffer
	env := newTextEnvironment(&buf)
	sxpf.EvaluateSlice(&env, vals)
	return buf.String()
}

type textEnvironment struct {
	err error
	w   io.Writer
	sm  *sxpf.SymbolMap
}

func newTextEnvironment(w io.Writer) textEnvironment {
	sm := sxpf.NewSymbolMap(nil)
	for _, bFn := range builtins {
		sym := bFn.sym
		minArgs := bFn.minArgs
		fn := bFn.fn
		sm.Set(sym, sxpf.NewBuiltin(
			sym.GetValue(),
			true, minArgs, -1,
			func(env sxpf.Environment, args []sxpf.Value) (sxpf.Value, error) {
				fn(env.(*textEnvironment), args)
				return sxpf.Nil(), nil
			},
		))
	}

	return textEnvironment{
		w:  w,
		sm: sm,
	}
}

func (env *textEnvironment) GetString(args []sxpf.Value, idx int) (res string) {
	if env.err == nil {
		res, env.err = sxpf.GetString(args, idx)
		return res
	}
	return ""
}
func (env *textEnvironment) WriteString(s string) {
	if env.err == nil {
		_, env.err = io.WriteString(env.w, s)
	}
}

func (env *textEnvironment) MakeSymbol(s string) *sxpf.Symbol { return sexpr.Smk.MakeSymbol(s) }

// LookupForm returns the form associated with the given symbol.
func (env *textEnvironment) LookupForm(sym *sxpf.Symbol) (sxpf.Form, error) {
	return env.sm.LookupForm(sym)
}

// Evaluate the string. In many cases, strings evaluate to itself.
func (env *textEnvironment) EvaluateString(str *sxpf.String) (sxpf.Value, error) {
	env.WriteString(str.GetValue())
	return sxpf.Nil(), nil
}

// Evaluate the symbol. In many cases this result in returning a value
// found in some internal lookup tables.
func (env *textEnvironment) EvaluateSymbol(*sxpf.Symbol) (sxpf.Value, error) {
	return sxpf.Nil(), nil
}

// Evaluate the given list. In many cases this means to evaluate the first
// element to a form and then call the form with the remaning elements
// (possibly evaluated) as parameters.
func (env *textEnvironment) EvaluateArray(lst *sxpf.Array) (sxpf.Value, error) {
	args := lst.GetValue()
	if sym, err := sxpf.GetSymbol(args, 0); err == nil {
		if form, err := env.LookupForm(sym); err == nil {
			form.Call(env, args[1:])
			return nil, nil
		}
	}
	sxpf.EvaluateSlice(env, args)
	return sxpf.Nil(), nil
}

var builtins = []struct {
	sym     *sxpf.Symbol
	minArgs int
	fn      func(env *textEnvironment, args []sxpf.Value)
}{
	{sexpr.SymText, 1, func(env *textEnvironment, args []sxpf.Value) { env.WriteString(env.GetString(args, 0)) }},
	{sexpr.SymTag, 1, func(env *textEnvironment, args []sxpf.Value) { env.WriteString(env.GetString(args, 0)) }},
	{sexpr.SymSpace, 0, func(env *textEnvironment, args []sxpf.Value) { env.WriteString(" ") }},
	{sexpr.SymSoft, 0, func(env *textEnvironment, args []sxpf.Value) { env.WriteString(" ") }},
	{sexpr.SymHard, 0, func(env *textEnvironment, args []sxpf.Value) { env.WriteString("\n") }},
}
