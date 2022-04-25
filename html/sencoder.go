//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package html

import (
	"io"
	"strconv"

	"zettelstore.de/c/sexpr"
)

type EncodingFunc func(env *EncEnvironment, args []sexpr.Value)
type encodingMap map[string]EncodingFunc

func (m encodingMap) Clone() encodingMap {
	if l := len(m); l > 0 {
		result := make(encodingMap, l)
		for k, v := range m {
			result[k] = v
		}
		return result
	}
	return nil

}

type EncEnvironment struct {
	err           error
	builtins      encodingMap
	w             io.Writer
	headingOffset int
}

func NewEncEnvironment(w io.Writer, headingOffset int) *EncEnvironment {
	return &EncEnvironment{
		builtins:      defaultEncodingFunctions.Clone(),
		w:             w,
		headingOffset: headingOffset,
	}
}

func (env *EncEnvironment) SetError(err error) {
	if env.err == nil {
		env.err = err
	}
}
func (env *EncEnvironment) GetError() error { return env.err }

func (env *EncEnvironment) WriteString(s string) {
	if env.err == nil {
		_, env.err = io.WriteString(env.w, s)
	}
}
func (env *EncEnvironment) WriteEscaped(s string) {
	if env.err == nil {
		_, env.err = Escape(env.w, s)
	}
}

func (env *EncEnvironment) WriteAttributes(value sexpr.Value) {
	attrList, ok := value.(*sexpr.List)
	if !ok {
		return
	}
	for _, attrVal := range attrList.GetValue() {
		attrPair, ok := attrVal.(*sexpr.List)
		if !ok {
			continue
		}
		attrs := attrPair.GetValue()
		if len(attrs) < 2 {
			continue
		}
		keyV, ok := attrs[0].(*sexpr.String)
		if !ok {
			continue
		}
		key := keyV.GetValue()
		if key == "" || key == "-" {
			continue
		}
		valV, ok := attrs[1].(*sexpr.String)
		if !ok {
			continue
		}
		val := valV.GetValue()
		env.WriteString(" ")
		env.WriteString(key)
		if val != "" {
			env.WriteString(`="`)
			if env.err == nil {
				_, env.err = AttributeEscape(env.w, val)
			}
			env.WriteString(`"`)
		}
	}
}

func (env *EncEnvironment) Encode(value sexpr.Value) {
	if env.err != nil {
		return
	}
	switch val := value.(type) {
	case *sexpr.Symbol:
		env.WriteEscaped(val.GetValue())
	case *sexpr.String:
		env.WriteEscaped(val.GetValue())
	case *sexpr.List:
		lstVals := val.GetValue()
		if len(lstVals) == 0 {
			return
		}
		if sym, ok := lstVals[0].(*sexpr.Symbol); ok {
			symStr := sym.GetValue()
			if f, found := env.builtins[symStr]; found {
				f(env, lstVals[1:])
				return
			}
		}
		for _, value := range lstVals {
			env.Encode(value)
		}
	}
}

var defaultEncodingFunctions = encodingMap{
	sexpr.SymPara.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		env.WriteString("<p>")
		env.Encode(sexpr.NewList(args...))
		env.WriteString("</p>")
	},
	sexpr.SymHeading.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) < 5 {
			return
		}
		levelSym, ok := args[0].(*sexpr.Symbol)
		if !ok {
			return
		}
		level, err := strconv.Atoi(levelSym.GetValue())
		if err != nil {
			env.SetError(err)
			return
		}
		levelS := strconv.Itoa(level + env.headingOffset)
		fragmentStr, ok := args[3].(*sexpr.String)
		if !ok {
			return
		}
		fragmentS := fragmentStr.GetValue()

		env.WriteString("<h")
		env.WriteString(levelS)
		env.WriteAttributes(args[1])
		if fragmentS != "" {
			env.WriteString(` id="`)
			env.WriteString(fragmentS)
			env.WriteString(`">`)
		} else {
			env.WriteString(">")
		}
		env.Encode(sexpr.NewList(args[4:]...))
		env.WriteString("</h")
		env.WriteString(levelS)
		env.WriteString(">")
	},
	sexpr.SymThematic.GetValue(): func(env *EncEnvironment, _ []sexpr.Value) { env.WriteString("<hr>") },
	sexpr.SymText.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) > 0 {
			if arg, ok := args[0].(*sexpr.String); ok {
				env.WriteEscaped(arg.GetValue())
			}
		}
	},
	sexpr.SymSpace.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) == 0 {
			env.WriteString(" ")
			return
		}
		if arg, ok := args[0].(*sexpr.String); ok {
			env.WriteString(arg.GetValue())
		}
	},
	sexpr.SymSoft.GetValue(): func(env *EncEnvironment, _ []sexpr.Value) { env.WriteString("\n") },
	sexpr.SymHard.GetValue(): func(env *EncEnvironment, _ []sexpr.Value) { env.WriteString("<br>\n") },
	sexpr.SymTag.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) > 0 {
			if arg, ok := args[0].(*sexpr.String); ok {
				env.WriteEscaped(arg.GetValue())
			}
		}
	},
}
