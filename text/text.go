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
	"log"

	"zettelstore.de/c/zjson"
)

// EncodeBlock writes the text of the given block list to the given writer.
func EncodeBlock(w io.Writer, a zjson.Array) {
	zjson.WalkBlock(&textV{w: w}, a, 0)
}

// EncodeInlineString returns the text content of the given inline list as a string.
func EncodeInlineString(a zjson.Array) string {
	var buf bytes.Buffer
	zjson.WalkInline(&textV{w: &buf}, a, 0)
	return buf.String()
}

type textV struct {
	w        io.Writer
	canSpace bool // It is allowed to write a space character
}

func (v *textV) WriteSpace() {
	if v.canSpace {
		v.w.Write([]byte{' '})
		v.canSpace = false
	}
}
func (v *textV) WriteString(s string) { io.WriteString(v.w, s); v.canSpace = true }

func (v *textV) BlockArray(a zjson.Array, pos int) zjson.CloseFunc  { return nil }
func (v *textV) InlineArray(a zjson.Array, pos int) zjson.CloseFunc { return nil }
func (v *textV) ItemArray(a zjson.Array, pos int) zjson.CloseFunc   { return nil }

func (v *textV) BlockObject(t string, obj zjson.Object, pos int) (bool, zjson.CloseFunc) {
	v.WriteSpace()
	return true, nil
}

func (v *textV) InlineObject(t string, obj zjson.Object, pos int) (bool, zjson.CloseFunc) {
	v.WriteSpace()
	switch t {
	case zjson.TypeText, zjson.TypeTag:
		v.WriteString(zjson.GetString(obj, zjson.NameString))
	case zjson.TypeSpace, zjson.TypeBreakSoft, zjson.TypeBreakHard:
		v.WriteSpace()
	default:
		return true, nil
	}
	return false, nil
}

func (v *textV) Unexpected(val zjson.Value, pos int, exp string) {
	log.Printf("?%v %d %T %v\n", exp, pos, val, val)
}
