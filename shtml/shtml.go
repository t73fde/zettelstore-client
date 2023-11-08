//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package shtml transforms a s-expr encoded zettel AST into a s-expr representation of HTML.
package shtml

import (
	"strconv"
	"strings"

	"zettelstore.de/client.fossil/attrs"
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxhtml"
)

// Transformer will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Transformer struct {
	sf            sx.SymbolFactory
	rebinder      RebindProc
	headingOffset int64
	unique        string
	endnotes      []endnoteInfo
	// noLinks       bool // true iff output must not include links
	symAttr  *sx.Symbol
	symClass *sx.Symbol
	symMeta  *sx.Symbol
	symA     *sx.Symbol
	// symSpan  *sx.Symbol
}

type endnoteInfo struct {
	// noteAST *sx.Pair // Endnote as AST
	noteHx *sx.Pair // Endnote as SxHTML
	attrs  *sx.Pair // attrs a-list
}

// NewTransformer creates a new transformer object.
func NewTransformer(headingOffset int, sf sx.SymbolFactory) *Transformer {
	if sf == nil {
		sf = sx.MakeMappedFactory(128)
	}
	return &Transformer{
		sf:            sf,
		rebinder:      nil,
		headingOffset: int64(headingOffset),
		symAttr:       sf.MustMake(sxhtml.NameSymAttr),
		symClass:      sf.MustMake("class"),
		symMeta:       sf.MustMake("meta"),
		symA:          sf.MustMake("a"),
		// 	symSpan:       sf.MustMake("span"),
	}
}

// SetUnique sets a prefix to make several HTML ids unique.
func (tr *Transformer) SetUnique(s string) { tr.unique = s }

// IsValidName returns true, if name is a valid symbol name.
func (tr *Transformer) IsValidName(s string) bool { return tr.sf.IsValidName(s) }

// Make a new HTML symbol.
func (tr *Transformer) Make(s string) *sx.Symbol { return tr.sf.MustMake(s) }

// RebindProc is a procedure which is called every time before a tranformation takes place.
type RebindProc func(*TransformEnv)

// SetRebinder sets the rebinder procedure.
func (tr *Transformer) SetRebinder(rb RebindProc) { tr.rebinder = rb }

// TransformAttrbute transforms the given attributes into a HTML s-expression.
func (tr *Transformer) TransformAttrbute(a attrs.Attributes) *sx.Pair {
	if len(a) == 0 {
		return nil
	}
	plist := sx.Nil()
	keys := a.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		if key != attrs.DefaultAttribute && tr.IsValidName(key) {
			plist = plist.Cons(sx.Cons(tr.Make(key), sx.String(a[key])))
		}
	}
	if plist == nil {
		return nil
	}
	return plist.Cons(tr.symAttr)
}

// TransformMeta creates a HTML meta s-expression
func (tr *Transformer) TransformMeta(a attrs.Attributes) *sx.Pair {
	return sx.Nil().Cons(tr.TransformAttrbute(a)).Cons(tr.symMeta)
}

// Transform an AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) Transform(lst *sx.Pair) (*sx.Pair, error) {
	return sx.Nil(), nil
}

// Endnotes returns a SHTML object with all collected endnotes.
func (tr *Transformer) Endnotes() *sx.Pair {
	if len(tr.endnotes) == 0 {
		return nil
	}
	result := sx.Nil().Cons(tr.Make("ol"))
	currResult := result.AppendBang(sx.Nil().Cons(sx.Cons(tr.symClass, sx.String("zs-endnotes"))).Cons(tr.symAttr))
	for i, fni := range tr.endnotes {
		noteNum := strconv.Itoa(i + 1)
		noteID := tr.unique + noteNum

		attrs := fni.attrs.Cons(sx.Cons(tr.symClass, sx.String("zs-endnote"))).
			Cons(sx.Cons(tr.Make("value"), sx.String(noteNum))).
			Cons(sx.Cons(tr.Make("id"), sx.String("fn:"+noteID))).
			Cons(sx.Cons(tr.Make("role"), sx.String("doc-endnote"))).
			Cons(tr.symAttr)

		backref := sx.Nil().Cons(sx.String("\u21a9\ufe0e")).
			Cons(sx.Nil().
				Cons(sx.Cons(tr.symClass, sx.String("zs-endnote-backref"))).
				Cons(sx.Cons(tr.Make("href"), sx.String("#fnref:"+noteID))).
				Cons(sx.Cons(tr.Make("role"), sx.String("doc-backlink"))).
				Cons(tr.symAttr)).
			Cons(tr.symA)

		li := sx.Nil().Cons(tr.Make("li"))
		li.AppendBang(attrs).
			ExtendBang(fni.noteHx).
			AppendBang(sx.String(" ")).AppendBang(backref)
		currResult = currResult.AppendBang(li)
	}
	tr.endnotes = nil
	return result
}

// TransformEnv is the environment where the actual transformation takes places.
type TransformEnv struct {
	tr *Transformer
}

func (te *TransformEnv) Make(name string) *sx.Symbol { return te.tr.Make(name) }

var unsafeSnippets = []string{
	"<script", "</script",
	"<iframe", "</iframe",
}

// IsSafe returns true if the given string does not contain unsafe HTML elements.
func IsSafe(s string) bool {
	lower := strings.ToLower(s)
	for _, snippet := range unsafeSnippets {
		if strings.Contains(lower, snippet) {
			return false
		}
	}
	return true
}
