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
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/zsx"
)

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

// Meta map metadata keys to MetaValue.
type Meta map[string]MetaValue

// MetaValue is an extended metadata value:
//
//   - Type: the type assiciated with the metata key
//   - Key: the metadata key itself
//   - Value: the metadata value as an (sx-) object.
type MetaValue struct {
	Type  string
	Key   string
	Value sx.Object
}

// MakeMeta build a Meta based on a list of metadata objects.
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
	for node := range lst.Tail().Pairs() {
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

// GetString return the metadata string value associated with the given key.
func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return zsx.GoValue(v.Value)
	}
	return ""
}

// GetPair return the metadata value associated with the given key,
// as a list of objects.
func (m Meta) GetPair(key string) *sx.Pair {
	if mv, found := m[key]; found {
		if pair, isPair := sx.GetPair(mv.Value); isPair {
			return pair
		}
	}
	return nil
}

// NormalizedSpacedText returns the given string, but normalize multiple spaces to one space.
func NormalizedSpacedText(s string) string { return strings.Join(strings.Fields(s), " ") }
