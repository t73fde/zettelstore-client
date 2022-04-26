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
	"log"
	"strconv"

	"zettelstore.de/c/attrs"
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
	noLinks       bool // true iff output must not include links
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

func (env *EncEnvironment) MissingArgs(args []sexpr.Value, minArgs int) bool {
	if len(args) < minArgs {
		env.SetError(fmt.Errorf("required args: %d, but got only: %d", minArgs, len(args)))
		return true
	}
	return false
}
func (env *EncEnvironment) GetSymbol(args []sexpr.Value, idx int) (res *sexpr.Symbol) {
	if env.err != nil {
		return nil
	}
	res, env.err = sexpr.GetSymbol(args, idx)
	return res
}
func (env *EncEnvironment) GetString(args []sexpr.Value, idx int) (res string) {
	if env.err != nil {
		return ""
	}
	res, env.err = sexpr.GetString(args, idx)
	return res
}
func (env *EncEnvironment) GetList(args []sexpr.Value, idx int) (res *sexpr.List) {
	if env.err != nil {
		return nil
	}
	res, env.err = sexpr.GetList(args, idx)
	return res
}

func (env *EncEnvironment) WriteAttributes(a attrs.Attributes) {
	if len(a) == 0 {
		return
	}
	for _, key := range a.Keys() {
		if key == "" || key == attrs.DefaultAttribute {
			continue
		}
		val, found := a.Get(key)
		if !found {
			continue
		}
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
		if env.MissingArgs(args, 5) {
			return
		}
		nLevel, err := strconv.Atoi(env.GetString(args, 0))
		if err != nil {
			env.SetError(err)
			return
		}
		level := strconv.Itoa(nLevel + env.headingOffset)

		a := sexpr.GetAttributes(env.GetList(args, 1))
		if fragment := env.GetString(args, 3); fragment != "" {
			a = a.Set("id", fragment)
		}

		env.WriteString("<h")
		env.WriteString(level)
		env.WriteAttributes(a)
		env.WriteString(">")
		env.EncodeList(args[4:])
		env.WriteString("</h")
		env.WriteString(level)
		env.WriteString(">")
	},
	sexpr.SymThematic.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		env.WriteString("<hr")
		if len(args) > 0 {
			env.WriteAttributes(sexpr.GetAttributes(env.GetList(args, 0)))
		}
		env.WriteString(">")
	},
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
	sexpr.SymLink.GetValue(): func(env *EncEnvironment, args []sexpr.Value) {
		if env.noLinks {
			spanList := sexpr.NewList(sexpr.SymFormatSpan)
			spanList.Append(args...)
			env.Encode(spanList)
			return
		}
		if env.MissingArgs(args, 2) {
			return
		}
		a := sexpr.GetAttributes(env.GetList(args, 0))
		ref := env.GetList(args, 1)
		if ref == nil {
			return
		}
		refPair := ref.GetValue()
		refKind := env.GetSymbol(refPair, 0)
		if refKind == nil {
			return
		}
		refValue := env.GetString(refPair, 1)
		switch {
		case sexpr.SymRefStateExternal.Equal(refKind):
			a = a.Set("href", refValue).AddClass("external")
		case sexpr.SymRefStateZettel.Equal(refKind), sexpr.SymRefStateBased.Equal(refKind), sexpr.SymRefStateHosted.Equal(refKind), sexpr.SymRefStateSelf.Equal(refKind):
			a = a.Set("href", refValue)
		case sexpr.SymRefStateBroken.Equal(refKind):
			a = a.AddClass("broken")
		default:
			log.Println("LINK", sexpr.NewList(args...))
		}
		env.WriteString("<a")
		env.WriteAttributes(a)
		env.WriteString(">")

		if in := args[2:]; len(in) != 0 {
			env.WriteString(refValue)
		} else {
			env.EncodeList(in)
		}
		env.WriteString("</a>")
	},
}
