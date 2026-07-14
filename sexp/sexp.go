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
	"t73f.de/r/sx/sxbuiltins"
	"t73f.de/r/zsc/webapi"
	"t73f.de/r/zsx"
)

// Often-used symbols
var (
	SymZettel      = sx.MakeSymbol("zettel")
	symRights      = sx.MakeSymbol("rights")
	symRightCreate = sx.MakeSymbol("create")
	symRightRead   = sx.MakeSymbol("read")
	symRightUpdate = sx.MakeSymbol("update")
	symRightDelete = sx.MakeSymbol("delete")
	symContent     = sx.MakeSymbol("content")
	SymList        = sx.MakeSymbol(sxbuiltins.List.Name)
	symMeta        = sx.MakeSymbol("meta")
)

// EncodeZettel transforms zettel data into a sx object.
func EncodeZettel(zettel webapi.ZettelData) sx.Object {
	return sx.MakeList(
		SymZettel,
		meta2sz(zettel.Meta),
		EncodeRights(zettel.Rights),
		EncodeContent(zettel.Content, zettel.Encoding),
	)
}

// ParseZettel parses an object to contain all needed data for a zettel.
func ParseZettel(obj sx.Object) (webapi.ZettelData, error) {
	vals, err := ParseList(obj, "yppp")
	if err != nil {
		return webapi.ZettelData{}, err
	}
	if errSym := CheckSymbol(vals[0], SymZettel); errSym != nil {
		return webapi.ZettelData{}, errSym
	}

	meta, err := ParseMeta(vals[1].(*sx.Pair))
	if err != nil {
		return webapi.ZettelData{}, err
	}

	rights, err := ParseRights(vals[2])
	if err != nil {
		return webapi.ZettelData{}, err
	}

	content, encoding, err := ParseContent(vals[3])
	if err != nil {
		return webapi.ZettelData{}, err
	}

	return webapi.ZettelData{
		Meta:     meta,
		Rights:   rights,
		Encoding: encoding,
		Content:  content,
	}, nil
}

// EncodeMetaRights translates metadata/rights into a sx object.
func EncodeMetaRights(mr webapi.MetaRights) *sx.Pair {
	return sx.MakeList(
		SymList,
		meta2sz(mr.Meta),
		EncodeRights(mr.Rights),
	)
}

// EncodeRights translate zettel rights in a sx object.
func EncodeRights(r webapi.ZettelRights) *sx.Pair {
	var lb sx.ListBuilder
	lb.Add(symRights)
	if r&webapi.ZettelCanCreate != 0 {
		lb.Add(symRightCreate)
	}
	if r&webapi.ZettelCanRead != 0 {
		lb.Add(symRightRead)
	}
	if r&webapi.ZettelCanWrite != 0 {
		lb.Add(symRightUpdate)
	}
	if r&webapi.ZettelCanDelete != 0 {
		lb.Add(symRightDelete)
	}
	return lb.List()
}

func meta2sz(m webapi.ZettelMeta) sx.Object {
	var result sx.ListBuilder
	result.Add(symMeta)
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
func ParseMeta(pair *sx.Pair) (webapi.ZettelMeta, error) {
	if err := CheckSymbol(pair.Car(), symMeta); err != nil {
		return nil, err
	}
	res := webapi.ZettelMeta{}
	for obj := range pair.Tail().Values() {
		mVals, err := ParseList(obj, "ys")
		if err != nil {
			return nil, err
		}
		res[(mVals[0].(*sx.Symbol)).GetValue()] = mVals[1].(sx.String).GetValue()
	}
	return res, nil
}

// ParseRights returns the rights values of the given object.
func ParseRights(obj sx.Object) (webapi.ZettelRights, error) {
	lst, isPair := sx.GetPair(obj)
	if !isPair {
		return webapi.ZettelCanNone, fmt.Errorf("no rights list, found %v/%T", obj, obj)
	}
	if err := CheckSymbol(lst.Car(), symRights); err != nil {
		return webapi.ZettelCanNone, err
	}
	result := webapi.ZettelRights(0)
	for node := lst.Tail(); node != nil; node = node.Tail() {
		switch sym := zsx.NodeSymbol(node); sym {
		case symRightCreate:
			result |= webapi.ZettelCanCreate
		case symRightRead:
			result |= webapi.ZettelCanRead
		case symRightUpdate:
			result |= webapi.ZettelCanWrite
		case symRightDelete:
			result |= webapi.ZettelCanDelete
		}
	}
	return result, nil
}

// EncodeContent transforms zettel content into a sx object.
func EncodeContent(content, encoding string) sx.Object {
	return sx.MakeList(symContent, sx.MakeString(encoding), sx.MakeString(content))
}

// ParseContent translates the given list to zettel content and encoding.
func ParseContent(obj sx.Object) (content string, encoding string, err error) {
	vals, err := ParseList(obj, "yss")
	if err != nil {
		return "", "", err
	}
	if errSym := CheckSymbol(vals[0], symContent); errSym != nil {
		return "", "", errSym
	}
	return vals[2].(sx.String).GetValue(), vals[1].(sx.String).GetValue(), nil
}

// ParseList parses the given object as a proper list, based on a type specification.
//
// 'b' expects a boolean, 'i' an int64, 'o' any object, 'p' a pair, 's' a string,
// and 'y' expects a symbol. A 'r' as the last type spracification matches all
// remaining values, including a non existent object.
func ParseList(obj sx.Object, spec string) (sx.Vector, error) {
	pair, isPair := sx.GetPair(obj)
	if !isPair {
		return nil, fmt.Errorf("not a list: %T/%v", obj, obj)
	}
	if pair == nil {
		if spec == "r" {
			return sx.Vector{sx.Nil()}, nil
		}
		if spec == "" {
			return nil, nil
		}
		return nil, ErrElementsMissing
	}

	specLen := len(spec)
	result := make(sx.Vector, 0, specLen)
	node, i := pair, 0
loop:
	for ; node != nil; i++ {
		if i >= specLen {
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
		case 'r':
			if i < specLen-1 {
				return nil, fmt.Errorf("spec 'r' must be the last: %q", spec)
			}
			result = append(result, node)
			i++
			break loop
		case 's':
			val, ok = sx.GetString(car)
		case 'y':
			val, ok = sx.GetSymbol(car)
		default:
			return nil, fmt.Errorf("unknown spec %d: '%c'", i, spec[i])
		}
		if !ok {
			return nil, fmt.Errorf("does not match spec %d '%c': %v", i, spec[i], car)
		}
		result = append(result, val)
		next, isNextPair := sx.GetPair(node.Cdr())
		if !isNextPair {
			return nil, sx.ErrImproper{Pair: pair}
		}
		node = next
	}
	if i < specLen {
		if lastSpec := specLen - 1; i < lastSpec || spec[lastSpec] != 'r' {
			return nil, ErrElementsMissing
		}
		result = append(result, sx.Nil())
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
func CheckSymbol(obj sx.Object, sym *sx.Symbol) error {
	if !sym.IsEqual(obj) {
		return fmt.Errorf("symbol %q expected, but got %v/%T is not a symbol", sym.GetValue(), obj, obj)
	}
	return nil
}
