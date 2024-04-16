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

// Package sz contains zettel data handling as sx expressions.
package sz

import (
	"t73f.de/r/sx"
	"zettelstore.de/client.fossil/attrs"
)

// GetAttributes traverses a s-expression list and returns an attribute structure.
func GetAttributes(seq *sx.Pair) (result attrs.Attributes) {
	for elem := seq; elem != nil; elem = elem.Tail() {
		pair, isPair := sx.GetPair(elem.Car())
		if !isPair || pair == nil {
			continue
		}
		key := pair.Car()
		if !key.IsAtom() {
			continue
		}
		val := pair.Cdr()
		if tail, isTailPair := sx.GetPair(val); isTailPair {
			val = tail.Car()
		}
		if !val.IsAtom() {
			continue
		}
		result = result.Set(GoValue(key), GoValue(val))
	}
	return result
}

// GoValue returns the string value of the sx.Object suitable for Go processing.
func GoValue(obj sx.Object) string {
	switch o := obj.(type) {
	case sx.String:
		return o.GetValue()
	case *sx.Symbol:
		return o.GetValue()
	}
	return obj.String()
}

// GetMetaContent returns the metadata and the content of a sz encoded zettel.
func GetMetaContent(zettel sx.Object) (Meta, *sx.Pair) {
	if pair, isPair := sx.GetPair(zettel); isPair {
		m := pair.Car()
		if s := pair.Tail(); s != nil {
			if content, isContentPair := sx.GetPair(s.Car()); isContentPair {
				return MakeMeta(m), content
			}
		}
		return MakeMeta(m), nil
	}
	return nil, nil
}

type Meta map[string]MetaValue
type MetaValue struct {
	Type  string
	Key   string
	Value sx.Object
}

func MakeMeta(obj sx.Object) Meta {
	if result := doMakeMeta(obj); len(result) > 0 {
		return result
	}
	return nil
}
func doMakeMeta(obj sx.Object) Meta {
	lst, isList := sx.GetPair(obj)
	if !isList || !lst.Car().IsEqual(SymMeta) {
		return nil
	}
	result := make(map[string]MetaValue)
	for node := lst.Tail(); node != nil; node = node.Tail() {
		if mv, found := makeMetaValue(node.Head()); found {
			result[mv.Key] = mv
		}
	}
	return result
}
func makeMetaValue(mnode *sx.Pair) (MetaValue, bool) {
	var result MetaValue
	typeSym, isSymbol := sx.GetSymbol(mnode.Car())
	if !isSymbol {
		return result, false
	}
	next := mnode.Tail()
	keySym, isSymbol := sx.GetSymbol(next.Car())
	if !isSymbol {
		return result, false
	}
	next = next.Tail()
	result.Type = typeSym.GetValue()
	result.Key = keySym.GetValue()
	result.Value = next.Car()
	return result, true
}

func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return GoValue(v.Value)
	}
	return ""
}

func (m Meta) GetPair(key string) *sx.Pair {
	if mv, found := m[key]; found {
		if pair, isPair := sx.GetPair(mv.Value); isPair {
			return pair
		}
	}
	return nil
}

// MapRefStateToLinkEmbed maps a reference state symbol to a link symbol or to
// an embed symbol, depending on 'forLink'.
func MapRefStateToLinkEmbed(symRefState *sx.Symbol, forLink bool) *sx.Symbol {
	if !forLink {
		return SymEmbed
	}
	if sym, found := mapRefStateLink[symRefState.GetValue()]; found {
		return sym
	}
	return SymLinkInvalid
}

var mapRefStateLink = map[string]*sx.Symbol{
	NameRefStateInvalid:  SymLinkInvalid,
	NameRefStateZettel:   SymLinkZettel,
	NameRefStateSelf:     SymLinkSelf,
	NameRefStateFound:    SymLinkFound,
	NameRefStateBroken:   SymLinkBroken,
	NameRefStateHosted:   SymLinkHosted,
	NameRefStateBased:    SymLinkBased,
	NameRefStateQuery:    SymLinkQuery,
	NameRefStateExternal: SymLinkExternal,
}

// IsBreakSym return true if the object is either a soft or a hard break symbol.
func IsBreakSym(obj sx.Object) bool {
	return SymSoft.IsEqual(obj) || SymHard.IsEqual(obj)
}
