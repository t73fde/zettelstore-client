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
	"zettelstore.de/sx.fossil/sxpf"
)

// GetAttributes traverses a s-expression list and returns an attribute structure.
func GetAttributes(seq *sxpf.Pair) (result attrs.Attributes) {
	for elem := seq; elem != nil; elem = elem.Tail() {
		pair, isPair := sxpf.GetPair(elem.Car())
		if !isPair || pair == nil {
			continue
		}
		key := pair.Car()
		if !key.IsAtom() {
			continue
		}
		val := pair.Cdr()
		if tail, isTailPair := sxpf.GetPair(val); isTailPair {
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
func GetMetaContent(zettel sxpf.Object) (Meta, *sxpf.Pair) {
	if pair, isPair := sxpf.GetPair(zettel); isPair {
		m := pair.Car()
		if s := pair.Tail(); s != nil {
			if content, isContentPair := sxpf.GetPair(s.Car()); isContentPair {
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
		pair, isPair := sxpf.GetPair(obj)
		if !isPair {
			return result
		}
		if mv, ok2 := makeMetaValue(pair); ok2 {
			result[mv.Key] = mv
		}
		obj = pair.Cdr()
	}
}
func makeMetaValue(mnode *sxpf.Pair) (MetaValue, bool) {
	var result MetaValue
	mval, isPair := sxpf.GetPair(mnode.Car())
	if !isPair {
		return result, false
	}
	typeSym, isSymbol := sxpf.GetSymbol(mval.Car())
	if !isSymbol {
		return result, false
	}
	keyPair, isPair := sxpf.GetPair(mval.Cdr())
	if !isPair {
		return result, false
	}
	keyList, isPair := sxpf.GetPair(keyPair.Car())
	if !isPair {
		return result, false
	}
	quoteSym, isSymbol := sxpf.GetSymbol(keyList.Car())
	if !isSymbol || quoteSym.Name() != "quote" {
		return result, false
	}
	keySym, isSymbol := sxpf.GetSymbol(keyList.Tail().Car())
	if !isSymbol {
		return result, false
	}
	valPair, isPair := sxpf.GetPair(keyPair.Cdr())
	if !isPair {
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

func (m Meta) GetPair(key string) *sxpf.Pair {
	if mv, found := m[key]; found {
		if pair, isPair := sxpf.GetPair(mv.Value); isPair {
			return pair
		}
	}
	return nil
}
