//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

// Package text provides types, constants and function to work with text output.
package text

import (
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/zsx"
	"t73f.de/r/zsx/input"
)

// Encoder is the structure to hold relevant data to execute the encoding.
type Encoder struct {
	sb strings.Builder
}

// NewEncoder returns a new text encoder.
func NewEncoder() *Encoder {
	enc := &Encoder{
		sb: strings.Builder{},
	}
	return enc
}

// Encode the object list as a string.
func (enc *Encoder) Encode(lst *sx.Pair) string {
	enc.executeList(lst)
	result := enc.sb.String()
	enc.sb.Reset()
	return result
}

// EvaluateInlineString returns the text content of the given inline list as a string.
func EvaluateInlineString(lst *sx.Pair) string {
	return NewEncoder().Encode(lst)
}

func (enc *Encoder) executeList(lst *sx.Pair) {
	for obj := range lst.Values() {
		enc.execute(obj)
	}
}
func (enc *Encoder) execute(obj sx.Object) {
	cmd, isPair := sx.GetPair(obj)
	if !isPair {
		return
	}
	sym := cmd.Car()
	if sx.IsNil(sym) {
		return
	}
	if sym.IsEqual(zsx.SymText) {
		args := cmd.Tail()
		if args == nil {
			return
		}
		if val, isString := sx.GetString(args.Car()); isString {
			hadSpace := false
			for _, ch := range val.GetValue() {
				if input.IsSpace(ch) {
					if !hadSpace {
						enc.sb.WriteByte(' ')
						hadSpace = true
					}
				} else {
					enc.sb.WriteRune(ch)
					hadSpace = false
				}
			}
		}
	} else if sym.IsEqual(zsx.SymSoft) {
		enc.sb.WriteByte(' ')
	} else if sym.IsEqual(zsx.SymHard) {
		enc.sb.WriteByte('\n')
	} else if !sym.IsEqual(sx.SymbolQuote) {
		enc.executeList(cmd.Tail())
	}
}
