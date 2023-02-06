//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
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
	"io"
	"strings"

	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/sexpr"
)

// EvaluateInlineString returns the text content of the given inline list as a string.
func EvaluateInlineString(pl *sxpf.Pair) string {
	var sb strings.Builder
	env := newTextEnvironment(&sb)
	env.EvalPair(pl)
	return sb.String()
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
			func(env sxpf.Environment, args *sxpf.Pair, _ int) (sxpf.Value, error) {
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

func (env *textEnvironment) GetString(p *sxpf.Pair) (res string) {
	if env.err == nil {
		res, env.err = p.GetString()
		return res
	}
	return ""
}

func (env *textEnvironment) WriteString(s string) {
	if env.err == nil {
		_, env.err = io.WriteString(env.w, s)
	}
}

func (env *textEnvironment) LookupForm(sym *sxpf.Symbol) (sxpf.Form, error) {
	return env.sm.LookupForm(sym)
}

func (*textEnvironment) EvalSymbol(*sxpf.Symbol) (sxpf.Value, error) { return nil, nil }
func (env *textEnvironment) EvalPair(p *sxpf.Pair) (sxpf.Value, error) {
	return sxpf.EvalCallOrSeq(env, p)
}
func (env *textEnvironment) EvalOther(val sxpf.Value) (sxpf.Value, error) {
	if strVal, ok := val.(*sxpf.String); ok {
		env.WriteString(strVal.GetValue())
		return nil, nil
	}
	return val, nil
}

var builtins = []struct {
	sym     *sxpf.Symbol
	minArgs int
	fn      func(env *textEnvironment, args *sxpf.Pair)
}{
	{sexpr.SymText, 1, func(env *textEnvironment, args *sxpf.Pair) { env.WriteString(env.GetString(args)) }},
	{sexpr.SymSpace, 0, func(env *textEnvironment, _ *sxpf.Pair) { env.WriteString(" ") }},
	{sexpr.SymSoft, 0, func(env *textEnvironment, _ *sxpf.Pair) { env.WriteString(" ") }},
	{sexpr.SymHard, 0, func(env *textEnvironment, _ *sxpf.Pair) { env.WriteString("\n") }},
}
