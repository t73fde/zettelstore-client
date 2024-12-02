//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

// Package sexp contains helper function to work with s-expression in an alien
// environment.
package sexp

import (
	"errors"
	"fmt"
	"sort"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/api"
)

// EncodeZettel transforms zettel data into a sx object.
func EncodeZettel(zettel api.ZettelData) sx.Object {
	return sx.MakeList(
		sx.MakeSymbol("zettel"),
		meta2sz(zettel.Meta),
		sx.MakeList(sx.MakeSymbol("rights"), sx.Int64(int64(zettel.Rights))),
		sx.MakeList(sx.MakeSymbol("encoding"), sx.MakeString(zettel.Encoding)),
		sx.MakeList(sx.MakeSymbol("content"), sx.MakeString(zettel.Content)),
	)
}

// ParseZettel parses an object to contain all needed data for a zettel.
func ParseZettel(obj sx.Object) (api.ZettelData, error) {
	vals, err := ParseList(obj, "ypppp")
	if err != nil {
		return api.ZettelData{}, err
	}
	if errSym := CheckSymbol(vals[0], "zettel"); errSym != nil {
		return api.ZettelData{}, errSym
	}

	meta, err := ParseMeta(vals[1].(*sx.Pair))
	if err != nil {
		return api.ZettelData{}, err
	}

	rights, err := ParseRights(vals[2])
	if err != nil {
		return api.ZettelData{}, err
	}

	encVals, err := ParseList(vals[3], "ys")
	if err != nil {
		return api.ZettelData{}, err
	}
	if errSym := CheckSymbol(encVals[0], "encoding"); errSym != nil {
		return api.ZettelData{}, errSym
	}

	contentVals, err := ParseList(vals[4], "ys")
	if err != nil {
		return api.ZettelData{}, err
	}
	if errSym := CheckSymbol(contentVals[0], "content"); errSym != nil {
		return api.ZettelData{}, errSym
	}

	return api.ZettelData{
		Meta:     meta,
		Rights:   rights,
		Encoding: encVals[1].(sx.String).GetValue(),
		Content:  contentVals[1].(sx.String).GetValue(),
	}, nil
}

// EncodeMetaRights translates metadata/rights into a sx object.
func EncodeMetaRights(mr api.MetaRights) *sx.Pair {
	return sx.MakeList(
		sx.SymbolList,
		meta2sz(mr.Meta),
		sx.MakeList(sx.MakeSymbol("rights"), sx.Int64(int64(mr.Rights))),
	)
}

func meta2sz(m api.ZettelMeta) sx.Object {
	var result sx.ListBuilder
	result.Add(sx.MakeSymbol("meta"))
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		val := sx.MakeList(sx.MakeSymbol(k), sx.MakeString(m[k]))
		result.Add(val)
	}
	return result.List()
}

// ParseMeta translates the given list to metadata.
func ParseMeta(pair *sx.Pair) (api.ZettelMeta, error) {
	if err := CheckSymbol(pair.Car(), "meta"); err != nil {
		return nil, err
	}
	res := api.ZettelMeta{}
	for node := pair.Tail(); node != nil; node = node.Tail() {
		mVals, err := ParseList(node.Car(), "ys")
		if err != nil {
			return nil, err
		}
		res[(mVals[0].(*sx.Symbol)).GetValue()] = mVals[1].(sx.String).GetValue()
	}
	return res, nil
}

// ParseRights returns the rights values of the given object.
func ParseRights(obj sx.Object) (api.ZettelRights, error) {
	rVals, err := ParseList(obj, "yi")
	if err != nil {
		return api.ZettelMaxRight, err
	}
	if errSym := CheckSymbol(rVals[0], "rights"); errSym != nil {
		return api.ZettelMaxRight, errSym
	}
	i64 := int64(rVals[1].(sx.Int64))
	if i64 < 0 && i64 >= int64(api.ZettelMaxRight) {
		return api.ZettelMaxRight, fmt.Errorf("invalid zettel right value: %v", i64)
	}
	return api.ZettelRights(i64), nil
}

// ParseList parses the given object as a proper list, based on a type specification.
func ParseList(obj sx.Object, spec string) (sx.Vector, error) {
	pair, isPair := sx.GetPair(obj)
	if !isPair {
		return nil, fmt.Errorf("not a list: %T/%v", obj, obj)
	}
	if pair == nil {
		if spec == "" {
			return nil, nil
		}
		return nil, ErrElementsMissing
	}

	result := make(sx.Vector, 0, len(spec))
	node, i := pair, 0
	for ; node != nil; i++ {
		if i >= len(spec) {
			return nil, ErrNoSpec
		}
		var val sx.Object
		var ok bool
		car := node.Car()
		switch spec[i] {
		case 'b':
			val, ok = sx.MakeBoolean(!sx.IsNil(car)), true
		case 'i':
			val, ok = car.(sx.Int64)
		case 'o':
			val, ok = car, true
		case 'p':
			val, ok = sx.GetPair(car)
		case 's':
			val, ok = sx.GetString(car)
		case 'y':
			val, ok = sx.GetSymbol(car)
		default:
			return nil, fmt.Errorf("unknown spec '%c'", spec[i])
		}
		if !ok {
			return nil, fmt.Errorf("does not match spec '%v': %v", spec[i], car)
		}
		result = append(result, val)
		next, isNextPair := sx.GetPair(node.Cdr())
		if !isNextPair {
			return nil, sx.ErrImproper{Pair: pair}
		}
		node = next
	}
	if i < len(spec) {
		return nil, ErrElementsMissing
	}
	return result, nil
}

// ErrElementsMissing is returned,
// if ParseList is called with a list smaller than the number of type specifications.
var ErrElementsMissing = errors.New("spec contains more data")

// ErrNoSpec is returned,
// if ParseList if called with a list greater than the number of type specifications.
var ErrNoSpec = errors.New("no spec for elements")

// CheckSymbol ensures that the given object is a symbol with the given name.
func CheckSymbol(obj sx.Object, name string) error {
	sym, isSymbol := sx.GetSymbol(obj)
	if !isSymbol {
		return fmt.Errorf("object %v/%T is not a symbol", obj, obj)
	}
	if got := sym.GetValue(); got != name {
		return fmt.Errorf("symbol %q expected, but got: %q", name, got)
	}
	return nil
}
