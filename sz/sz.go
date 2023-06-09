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
	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/attrs"
)

// GetAttributes traverses a s-expression list and returns an attribute structure.
func GetAttributes(seq *sxpf.Cell) (result attrs.Attributes) {
	for elem := seq; elem != nil; elem = elem.Tail() {
		p, ok := sxpf.GetList(elem.Car())
		if !ok || p == nil {
			continue
		}
		key := p.Car()
		if !key.IsAtom() {
			continue
		}
		val := p.Cdr()
		if tail, ok2 := sxpf.GetList(val); ok2 {
			val = tail.Car()
		}
		if !val.IsAtom() {
			continue
		}
		result = result.Set(key.String(), val.String())
	}
	return result
}

// GetMetaContent returns the metadata and the content of a sz encoded zettel.
func GetMetaContent(zettel sxpf.Object) (Meta, *sxpf.Cell) {
	if pair, ok := sxpf.GetList(zettel); ok {
		m := pair.Car()
		if s := pair.Tail(); s != nil {
			if content, ok2 := sxpf.GetList(s.Car()); ok2 {
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
	Value sxpf.Object
}

func MakeMeta(obj sxpf.Object) Meta {
	if result := doMakeMeta(obj); len(result) > 0 {
		return result
	}
	return nil
}
func doMakeMeta(obj sxpf.Object) Meta {
	result := make(map[string]MetaValue)
	for {
		if sxpf.IsNil(obj) {
			return result
		}
		lst, ok := sxpf.GetList(obj)
		if !ok {
			return result
		}
		if mv, ok2 := makeMetaValue(lst); ok2 {
			result[mv.Key] = mv
		}
		obj = lst.Cdr()
	}
}
func makeMetaValue(mnode *sxpf.Cell) (MetaValue, bool) {
	var result MetaValue
	mval, ok := sxpf.GetList(mnode.Car())
	if !ok {
		return result, false
	}
	typeSym, ok := sxpf.GetSymbol(mval.Car())
	if !ok {
		return result, false
	}
	keyPair, ok := sxpf.GetList(mval.Cdr())
	if !ok {
		return result, false
	}
	keyList, ok := sxpf.GetList(keyPair.Car())
	if !ok {
		return result, false
	}
	quoteSym, ok := sxpf.GetSymbol(keyList.Car())
	if !ok || quoteSym.Name() != "quote" {
		return result, false
	}
	keySym, ok := sxpf.GetSymbol(keyList.Tail().Car())
	if !ok {
		return result, false
	}
	valPair, ok := sxpf.GetList(keyPair.Cdr())
	if !ok {
		return result, false
	}
	result.Type = typeSym.Name()
	result.Key = keySym.Name()
	result.Value = valPair.Car()
	return result, true
}

func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return v.Value.String()
	}
	return ""
}

func (m Meta) GetList(key string) *sxpf.Cell {
	if mv, found := m[key]; found {
		if seq, ok := sxpf.GetList(mv.Value); ok {
			return seq
		}
	}
	return nil
}
