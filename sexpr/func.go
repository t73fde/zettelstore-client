//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr

import "io"

// Function is a wrapper for a primitive or a user defined function.
// Currently, only primitive functions are allowed.
type Function struct {
	name      string
	primitive PrimitiveFn
	special   bool
}

// PrimitiveFn is a primitve function that is implemented in Go.
type PrimitiveFn func(Environment, []Value) (Value, error)

// NewPrimitive returns a new primitive function.
func NewPrimitive(name string, special bool, fn PrimitiveFn) *Function {
	return &Function{name, fn, special}
}

func (fn *Function) Equal(other Value) bool {
	if fn == nil || other == nil {
		return Value(fn) == other
	}
	if o, ok := other.(*Function); ok {
		return fn.name == o.name
	}
	return false
}

func (fn *Function) Encode(w io.Writer) (int, error) { return io.WriteString(w, fn.String()) }

func (fn *Function) String() string { return "#" + fn.name }

func (fn *Function) IsSpecial() bool { return fn != nil && fn.special }
func (fn *Function) Name() string {
	if fn == nil {
		return ""
	}
	return fn.name
}

func (fn *Function) Call(env Environment, args []Value) (Value, error) {
	return fn.primitive(env, args)
}
