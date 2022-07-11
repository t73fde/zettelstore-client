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

	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/sexpr"
)

// EvaluateInlineString returns the text content of the given inline list as a string.
func EvaluateInlineString(pl *sxpf.Pair) string {
	var buf bytes.Buffer
	env := newTextEnvironment(&buf)
	env.EvaluatePair(pl)
	return buf.String()
}

type textEnvironment struct {
	err error
	w   io.Writer
	sm  *sxpf.SymbolMap
}

func newTextEnvironment(w io.Writer) textEnvironment {
	sm := sxpf.NewSymbolMap(sexpr.Smk, nil)
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

func (*textEnvironment) MakeSymbol(s string) *sxpf.Symbol { return sexpr.Smk.MakeSymbol(s) }

func (env *textEnvironment) LookupForm(sym *sxpf.Symbol) (sxpf.Form, error) {
	return env.sm.LookupForm(sym)
}

func (*textEnvironment) EvaluateSymbol(*sxpf.Symbol) (sxpf.Value, error) { return sxpf.Nil(), nil }

func (env *textEnvironment) EvaluatePair(p *sxpf.Pair) (sxpf.Value, error) {
	if p.IsEmpty() {
		return p, nil
	}
	if sym, ok := p.GetFirst().(*sxpf.Symbol); ok {
		if form, err := env.LookupForm(sym); err == nil {
			if rest, ok := p.GetSecond().(*sxpf.Pair); ok {
				form.Call(env, rest)
				return nil, nil
			}
		}
	}
	sxpf.EvaluateList(env, p)
	return sxpf.Nil(), nil
}

func (env *textEnvironment) EvaluateOther(val sxpf.Value) (sxpf.Value, error) {
	if strVal, ok := val.(*sxpf.String); ok {
		env.WriteString(strVal.GetValue())
		return sxpf.Nil(), nil
	}
	return val, nil
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
