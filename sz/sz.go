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
		cell, isCell := sxpf.GetCell(elem.Car())
		if !isCell || cell == nil {
			continue
		}
		key := cell.Car()
		if !key.IsAtom() {
			continue
		}
		val := cell.Cdr()
		if tail, isTailCell := sxpf.GetCell(val); isTailCell {
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
	if cell, isCell := sxpf.GetCell(zettel); isCell {
		m := cell.Car()
		if s := cell.Tail(); s != nil {
			if content, isContentCell := sxpf.GetCell(s.Car()); isContentCell {
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
		cell, isCell := sxpf.GetCell(obj)
		if !isCell {
			return result
		}
		if mv, ok2 := makeMetaValue(cell); ok2 {
			result[mv.Key] = mv
		}
		obj = cell.Cdr()
	}
}
func makeMetaValue(mnode *sxpf.Cell) (MetaValue, bool) {
	var result MetaValue
	mval, isCell := sxpf.GetCell(mnode.Car())
	if !isCell {
		return result, false
	}
	typeSym, isSymbol := sxpf.GetSymbol(mval.Car())
	if !isSymbol {
		return result, false
	}
	keyPair, isCell := sxpf.GetCell(mval.Cdr())
	if !isCell {
		return result, false
	}
	keyList, isCell := sxpf.GetCell(keyPair.Car())
	if !isCell {
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
	valPair, isCell := sxpf.GetCell(keyPair.Cdr())
	if !isCell {
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

func (m Meta) GetCell(key string) *sxpf.Cell {
	if mv, found := m[key]; found {
		if cell, isCell := sxpf.GetCell(mv.Value); isCell {
			return cell
		}
	}
	return nil
}
