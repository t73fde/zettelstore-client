//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sz

import (
	"zettelstore.de/client.fossil/attrs"
	"zettelstore.de/sx.fossil"
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
		result = result.Set(goString(key), goString(val))
	}
	return result
}

func goString(obj sx.Object) string {
	switch o := obj.(type) {
	case sx.String:
		return string(o)
	case sx.Symbol:
		return string(o)
	default:
		return obj.String()
	}
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
	result := make(map[string]MetaValue)
	for {
		if sx.IsNil(obj) {
			return result
		}
		pair, isPair := sx.GetPair(obj)
		if !isPair {
			return result
		}
		if mv, ok2 := makeMetaValue(pair); ok2 {
			result[mv.Key] = mv
		}
		obj = pair.Cdr()
	}
}
func makeMetaValue(mnode *sx.Pair) (MetaValue, bool) {
	var result MetaValue
	mval, isPair := sx.GetPair(mnode.Car())
	if !isPair {
		return result, false
	}
	typeSym, isSymbol := sx.GetSymbol(mval.Car())
	if !isSymbol {
		return result, false
	}
	keyPair, isPair := sx.GetPair(mval.Cdr())
	if !isPair {
		return result, false
	}
	keyList, isPair := sx.GetPair(keyPair.Car())
	if !isPair {
		return result, false
	}
	quoteSym, isSymbol := sx.GetSymbol(keyList.Car())
	if !isSymbol || quoteSym != "quote" {
		return result, false
	}
	keySym, isSymbol := sx.GetSymbol(keyList.Tail().Car())
	if !isSymbol {
		return result, false
	}
	valPair, isPair := sx.GetPair(keyPair.Cdr())
	if !isPair {
		return result, false
	}
	result.Type = string(typeSym)
	result.Key = string(keySym)
	result.Value = valPair.Car()
	return result, true
}

func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return goString(v.Value)
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
func MapRefStateToLinkEmbed(symRefState sx.Symbol, forLink bool) sx.Symbol {
	if !forLink {
		return SymEmbed
	}
	if sym, found := mapRefStateLink[symRefState]; found {
		return sym
	}
	return SymLinkInvalid
}

var mapRefStateLink = map[sx.Symbol]sx.Symbol{
	SymRefStateInvalid:  SymLinkInvalid,
	SymRefStateZettel:   SymLinkZettel,
	SymRefStateSelf:     SymLinkSelf,
	SymRefStateFound:    SymLinkFound,
	SymRefStateBroken:   SymLinkBroken,
	SymRefStateHosted:   SymLinkHosted,
	SymRefStateBased:    SymLinkBased,
	SymRefStateQuery:    SymLinkQuery,
	SymRefStateExternal: SymLinkExternal,
}

// IsBreakSym return true if the object is either a soft or a hard break symbol.
func IsBreakSym(obj sx.Object) bool {
	return SymSoft.IsEqual(obj) || SymHard.IsEqual(obj)
}
