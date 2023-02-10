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
	"strings"

	"codeberg.org/t73fde/sxpf"
	"codeberg.org/t73fde/sxpf/eval"
	"zettelstore.de/c/sexpr"
)

// EvaluateInlineString returns the text content of the given inline list as a string.
func EvaluateInlineString(lst *sxpf.List) string {
	sf := sxpf.FindSymbolFactory(lst)
	if sf == nil {
		return ""
	}
	var sb strings.Builder
	env := sxpf.MakeRootEnvironment()
	env.Bind(sf.Make(sexpr.NameSymText), eval.MakeSpecial(
		sexpr.NameSymText,
		func(_ sxpf.Environment, args *sxpf.List) (sxpf.Value, error) {
			if args != nil {
				if val, ok := args.Head().(sxpf.String); ok {
					sb.WriteString(val.String())
				}
			}
			return sxpf.Nil(), nil
		},
	))
	env.Bind(sf.Make(sexpr.NameSymSpace), eval.MakeSpecial(
		sexpr.NameSymSpace,
		func(sxpf.Environment, *sxpf.List) (sxpf.Value, error) {
			sb.WriteByte(' ')
			return sxpf.Nil(), nil
		},
	))
	env.Bind(sf.Make(sexpr.NameSymSoft), eval.MakeSpecial(
		sexpr.NameSymSoft,
		func(sxpf.Environment, *sxpf.List) (sxpf.Value, error) {
			sb.WriteByte(' ')
			return sxpf.Nil(), nil
		},
	))
	env.Bind(sf.Make(sexpr.NameSymHard), eval.MakeSpecial(
		sexpr.NameSymHard,
		func(sxpf.Environment, *sxpf.List) (sxpf.Value, error) {
			sb.WriteByte('\n')
			return sxpf.Nil(), nil
		},
	))
	sexpr.BindOther(env, sf)

	eval.Eval(env, lst)
	return sb.String()
}
