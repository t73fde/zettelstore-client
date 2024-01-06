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

	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

// Encoder is the structure to hold relevant data to execute the encoding.
type Encoder struct {
	sb strings.Builder
}

func NewEncoder() *Encoder {
	enc := &Encoder{
		sb: strings.Builder{},
	}
	return enc
}

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
	for elem := lst; elem != nil; elem = elem.Tail() {
		enc.execute(elem.Car())
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
	if sym.IsEqual(sz.SymText) {
		args := cmd.Tail()
		if args == nil {
			return
		}
		if val, isString := sx.GetString(args.Car()); isString {
			enc.sb.WriteString(string(val))
		}
	} else if sym.IsEqual(sz.SymSpace) || sym.IsEqual(sz.SymSoft) {
		enc.sb.WriteByte(' ')
	} else if sym.IsEqual(sz.SymHard) {
		enc.sb.WriteByte('\n')
	} else if !sym.IsEqual(sx.SymbolQuote) {
		enc.executeList(cmd.Tail())
	}
}
