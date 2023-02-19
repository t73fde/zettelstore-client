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

// Encoder is the structure to hold relevant data to execute the encoding.
type Encoder struct {
	sf  sxpf.SymbolFactory
	env sxpf.Environment
	sb  strings.Builder
}

func NewEncoder(sf sxpf.SymbolFactory) *Encoder {
	if sf == nil {
		return nil
	}
	enc := &Encoder{
		sf:  sf,
		env: nil,
		sb:  strings.Builder{},
	}
	env := sxpf.MakeRootEnvironment()
	env.Bind(sf.Make(sexpr.NameSymText), eval.MakeSpecial(
		sexpr.NameSymText,
		func(_ sxpf.Environment, args *sxpf.List) (sxpf.Value, error) {
			if args != nil {
				if val, ok := args.Car().(sxpf.String); ok {
					enc.sb.WriteString(val.String())
				}
			}
			return sxpf.Nil(), nil
		},
	))
	env.Bind(sf.Make(sexpr.NameSymSpace), eval.MakeSpecial(
		sexpr.NameSymSpace,
		func(sxpf.Environment, *sxpf.List) (sxpf.Value, error) {
			enc.sb.WriteByte(' ')
			return sxpf.Nil(), nil
		},
	))
	env.Bind(sf.Make(sexpr.NameSymSoft), eval.MakeSpecial(
		sexpr.NameSymSoft,
		func(sxpf.Environment, *sxpf.List) (sxpf.Value, error) {
			enc.sb.WriteByte(' ')
			return sxpf.Nil(), nil
		},
	))
	env.Bind(sf.Make(sexpr.NameSymHard), eval.MakeSpecial(
		sexpr.NameSymHard,
		func(sxpf.Environment, *sxpf.List) (sxpf.Value, error) {
			enc.sb.WriteByte('\n')
			return sxpf.Nil(), nil
		},
	))
	sexpr.BindOther(env, sf)

	enc.env = env
	return enc
}

func (enc *Encoder) Encode(lst *sxpf.List) string {
	eval.Eval(enc.env, lst)
	result := enc.sb.String()
	enc.sb.Reset()
	return result
}

// EvaluateInlineString returns the text content of the given inline list as a string.
func EvaluateInlineString(lst *sxpf.List) string {
	if sf := sxpf.FindSymbolFactory(lst); sf != nil {
		return NewEncoder(sf).Encode(lst)
	}
	return ""
}
