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

	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

// Encoder is the structure to hold relevant data to execute the encoding.
type Encoder struct {
	sf sx.SymbolFactory
	sb strings.Builder

	symText  *sx.Symbol
	symSpace *sx.Symbol
	symSoft  *sx.Symbol
	symHard  *sx.Symbol
	symQuote *sx.Symbol
}

func NewEncoder(sf sx.SymbolFactory) *Encoder {
	if sf == nil {
		return nil
	}
	enc := &Encoder{
		sf:       sf,
		sb:       strings.Builder{},
		symText:  sf.MustMake(sz.NameSymText),
		symSpace: sf.MustMake(sz.NameSymSpace),
		symSoft:  sf.MustMake(sz.NameSymSoft),
		symHard:  sf.MustMake(sz.NameSymHard),
		symQuote: sf.MustMake(sz.NameSymQuote),
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
	if sf := sx.FindSymbolFactory(lst); sf != nil {
		return NewEncoder(sf).Encode(lst)
	}
	return ""
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
	if sym.IsEqual(enc.symText) {
		args := cmd.Tail()
		if args == nil {
			return
		}
		if val, isString := sx.GetString(args.Car()); isString {
			enc.sb.WriteString(val.String())
		}
	} else if sym.IsEqual(enc.symSpace) || sym.IsEqual(enc.symSoft) {
		enc.sb.WriteByte(' ')
	} else if sym.IsEqual(enc.symHard) {
		enc.sb.WriteByte('\n')
	} else if !sym.IsEqual(enc.symQuote) {
		enc.executeList(cmd.Tail())
	}
}
