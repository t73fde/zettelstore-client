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
	"zettelstore.de/c/sexpr"
)

// Encoder is the structure to hold relevant data to execute the encoding.
type Encoder struct {
	sf sxpf.SymbolFactory
	sb strings.Builder

	symText  *sxpf.Symbol
	symSpace *sxpf.Symbol
	symSoft  *sxpf.Symbol
	symHard  *sxpf.Symbol
	symQuote *sxpf.Symbol
}

func NewEncoder(sf sxpf.SymbolFactory) *Encoder {
	if sf == nil {
		return nil
	}
	enc := &Encoder{
		sf:       sf,
		sb:       strings.Builder{},
		symText:  sf.MustMake(sexpr.NameSymText),
		symSpace: sf.MustMake(sexpr.NameSymSpace),
		symSoft:  sf.MustMake(sexpr.NameSymSoft),
		symHard:  sf.MustMake(sexpr.NameSymHard),
		symQuote: sf.MustMake(sexpr.NameSymQuote),
	}
	return enc
}

func (enc *Encoder) Encode(lst *sxpf.List) string {
	enc.executeList(lst)
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

func (enc *Encoder) executeList(lst *sxpf.List) {
	for elem := lst; elem != nil; elem = elem.Tail() {
		enc.execute(elem.Car())
	}
}
func (enc *Encoder) execute(obj sxpf.Object) {
	cmd, ok := obj.(*sxpf.List)
	if !ok {
		return
	}
	sym := cmd.Car()
	if sxpf.IsNil(sym) {
		return
	}
	if sym.IsEqual(enc.symText) {
		args := cmd.Tail()
		if args == nil {
			return
		}
		if val, ok2 := args.Car().(sxpf.String); ok2 {
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
