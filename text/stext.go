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
func SEncodeBlock(w io.Writer, lst *sxpf.List) {
	env := textEnvironment{w: w}
	env.Encode(lst)
}

// SEncodeInlineString returns the text content of the given inline list as a string.
func SEncodeInlineString(vals []sxpf.Value) string {
	var buf bytes.Buffer
	env := textEnvironment{w: &buf}
	env.EncodeList(vals)
	return buf.String()
}

type textEnvironment struct {
	err error
	w   io.Writer
}

func (env *textEnvironment) WriteString(s string) {
	if env.err == nil {
		_, env.err = io.WriteString(env.w, s)
	}
}

func (env *textEnvironment) GetString(args []sxpf.Value, idx int) (res string) {
	if env.err == nil {
		res, env.err = sxpf.GetString(args, idx)
		return res
	}
	return ""
}

func (env *textEnvironment) Encode(value sxpf.Value) {
	if env.err != nil {
		return
	}
	switch val := value.(type) {
	case *sxpf.Symbol:
		// Do nothing: there is no relevant text in a symbol
	case *sxpf.String:
		env.WriteString(val.GetValue())
	case *sxpf.List:
		env.EncodeList(val.GetValue())
	}
}

func (env *textEnvironment) EncodeList(lst []sxpf.Value) {
	if len(lst) == 0 {
		return
	}
	if sym, ok := lst[0].(*sxpf.Symbol); ok {
		if f, found := builtins[sym]; found && f != nil {
			f(env, lst[1:])
			return
		}
	}
	for _, value := range lst {
		env.Encode(value)
	}
}

var builtins = map[*sxpf.Symbol]func(env *textEnvironment, args []sxpf.Value){
	sexpr.SymText:  func(env *textEnvironment, args []sxpf.Value) { env.WriteString(env.GetString(args, 0)) },
	sexpr.SymTag:   func(env *textEnvironment, args []sxpf.Value) { env.WriteString(env.GetString(args, 0)) },
	sexpr.SymSpace: func(env *textEnvironment, args []sxpf.Value) { env.WriteString(" ") },
	sexpr.SymSoft:  func(env *textEnvironment, args []sxpf.Value) { env.WriteString(" ") },
	sexpr.SymHard:  func(env *textEnvironment, args []sxpf.Value) { env.WriteString("\n") },
}
