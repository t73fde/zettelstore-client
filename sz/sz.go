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
func GetAttributes(seq *sxpf.List) (result attrs.Attributes) {
	for elem := seq; elem != nil; elem = elem.Tail() {
		p, ok := elem.Car().(*sxpf.List)
		if !ok || p == nil {
			continue
		}
		key := p.Car()
		if !sxpf.IsAtom(key) {
			continue
		}
		val := p.Cdr()
		if tail, ok2 := val.(*sxpf.List); ok2 {
			val = tail.Car()
		}
		if !sxpf.IsAtom(val) {
			continue
		}
		result = result.Set(key.String(), val.String())
	}
	return result
}

// GetMetaContent returns the metadata and the content of a sz encoded zettel.
func GetMetaContent(zettel sxpf.Object) (Meta, *sxpf.List) {
	if pair, ok := zettel.(*sxpf.List); ok {
		m := pair.Car()
		if s := pair.Tail(); s != nil {
			if content, ok2 := s.Car().(*sxpf.List); ok2 {
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

func MakeMeta(val sxpf.Object) Meta {
	if result := doMakeMeta(val); len(result) > 0 {
		return result
	}
	return nil
}
func doMakeMeta(val sxpf.Object) Meta {
	result := make(map[string]MetaValue)
	for {
		if sxpf.IsNil(val) {
			return result
		}
		lst, ok := val.(*sxpf.List)
		if !ok {
			return result
		}
		if mv, ok2 := makeMetaValue(lst); ok2 {
			result[mv.Key] = mv
		}
		val = lst.Cdr()
	}
}
func makeMetaValue(pair *sxpf.List) (MetaValue, bool) {
	var result MetaValue
	typePair, ok := pair.Car().(*sxpf.List)
	if !ok {
		return result, false
	}
	typeVal, ok := typePair.Car().(*sxpf.Symbol)
	if !ok {
		return result, false
	}
	keyPair, ok := typePair.Cdr().(*sxpf.List)
	if !ok {
		return result, false
	}
	keyStr, ok := keyPair.Car().(sxpf.String)
	if !ok {
		return result, false
	}
	valPair, ok := keyPair.Cdr().(*sxpf.List)
	if !ok {
		return result, false
	}
	result.Type = typeVal.Name()
	result.Key = keyStr.String()
	result.Value = valPair.Car()
	return result, true
}

func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return v.Value.String()
	}
	return ""
}

func (m Meta) GetList(key string) *sxpf.List {
	if mv, found := m[key]; found {
		if seq, ok := mv.Value.(*sxpf.List); ok {
			return seq
		}
	}
	return nil
}
