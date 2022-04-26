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
	"fmt"
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

func (env *EncEnvironment) GetString(args []sexpr.Value, idx int) string {
	if env.err != nil {
		return ""
	}
	if idx < 0 && len(args) <= idx {
		env.SetError(fmt.Errorf("index %d out of bounds: %v", idx, args))
		return ""
	}
	if val, ok := args[idx].(*sexpr.String); ok {
		return val.GetValue()
	}
	if val, ok := args[idx].(*sexpr.Symbol); ok {
		return val.GetValue()
	}
	env.SetError(fmt.Errorf("%v / %d is not a string", args[idx], idx))
	return ""
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
		key := env.GetString(attrs, 0)
		if key == "" || key == "-" {
			continue
		}
		val := env.GetString(attrs, 1)
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
		env.EncodeList(val.GetValue())
	}
}
func (env *EncEnvironment) EncodeList(lst []sexpr.Value) {
	if len(lst) == 0 {
		return
	}
	if sym, ok := lst[0].(*sexpr.Symbol); ok {
		symStr := sym.GetValue()
		if f, found := env.builtins[symStr]; found {
			f(env, lst[1:])
			return
		}
		env.SetError(fmt.Errorf("unbound identifier: %q", symStr))
		return
	}
	for _, value := range lst {
		env.Encode(value)
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
		nLevel, err := strconv.Atoi(env.GetString(args, 0))
		if err != nil {
			env.SetError(err)
			return
		}
		level := strconv.Itoa(nLevel + env.headingOffset)
		fragment := env.GetString(args, 3)

		env.WriteString("<h")
		env.WriteString(level)
		env.WriteAttributes(args[1])
		if fragment != "" {
			env.WriteString(` id="`)
			env.WriteString(fragment)
			env.WriteString(`">`)
		} else {
			env.WriteString(">")
		}
		env.EncodeList(args[4:])
		env.WriteString("</h")
		env.WriteString(level)
		env.WriteString(">")
	},
	sexpr.SymThematic.GetValue(): func(env *EncEnvironment, _ []sexpr.Value) { env.WriteString("<hr>") },
	sexpr.SymText.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) > 0 {
			env.WriteEscaped(env.GetString(args, 0))
		}
	},
	sexpr.SymSpace.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) == 0 {
			env.WriteString(" ")
			return
		}
		env.WriteEscaped(env.GetString(args, 0))
	},
	sexpr.SymSoft.GetValue(): func(env *EncEnvironment, _ []sexpr.Value) { env.WriteString("\n") },
	sexpr.SymHard.GetValue(): func(env *EncEnvironment, _ []sexpr.Value) { env.WriteString("<br>\n") },
	sexpr.SymTag.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if len(args) > 0 {
			env.WriteEscaped(env.GetString(args, 0))
		}
	},
}
