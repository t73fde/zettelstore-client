//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package zjson provides types, constants and function to work with the ZJSON
// encoding of zettel.
package zjson

// Value is the gerneric JSON value.
type Value = interface{}

// Array represents a JSON array.
type Array = []Value

// Object represents a JSON object.
type Object = map[string]Value

// Visitor provides functionality when a Value is traversed.
type Visitor interface {
	Block(a Array, pos int) EndFunc
	Inline(a Array, pos int) EndFunc
	Item(a Array, pos int) EndFunc
	Object(t string, obj Object, pos int) (bool, EndFunc)

	NoValue(val Value, pos int)
	NoArray(val Value, pos int)
	NoObject(obj Object, pos int)
}

// EndFunc is a function that executes after a ZJSON element is visited.
type EndFunc func()

// WalkBlock traverses a block array.
func WalkBlock(v Visitor, a Array, pos int) {
	ef := v.Block(a, pos)
	for i, elem := range a {
		walkObject(v, elem, i)
	}
	if ef != nil {
		ef()
	}
}

// WalkInline traverses an inline array.
func WalkInline(v Visitor, a Array, pos int) {
	ef := v.Inline(a, pos)
	for i, elem := range a {
		walkObject(v, elem, i)
	}
	if ef != nil {
		ef()
	}
}

func walkObject(v Visitor, val Value, pos int) {
	obj, ok := val.(Object)
	if !ok {
		v.NoValue(val, pos)
		return
	}

	tVal, ok := obj[NameType]
	if !ok {
		v.NoObject(obj, pos)
		return
	}
	t, ok := tVal.(string)
	if !ok {
		v.NoObject(obj, pos)
		return
	}

	doChilds, ef := v.Object(t, obj, pos)
	if doChilds {
		inlineChild(v, obj, pos)
		blockChild(v, obj, pos)
		itemChild(v, obj, pos)
	}
	if ef != nil {
		ef()
	}
}

func inlineChild(v Visitor, obj Object, pos int) {
	if iVal, ok := obj[NameInline]; ok {
		if il, ok := iVal.(Array); ok {
			WalkInline(v, il, 0)
		} else {
			v.NoArray(iVal, pos)
		}
	}
}
func blockChild(v Visitor, obj Object, pos int) {
	if bVal, ok := obj[NameBlock]; ok {
		if bl, ok := bVal.(Array); ok {
			WalkBlock(v, bl, 0)
		} else {
			v.NoArray(bVal, pos)
		}
	}
}
func itemChild(v Visitor, obj Object, pos int) {
	if iVal, ok := obj[NameList]; ok {
		if it, ok := iVal.(Array); ok {
			for i, l := range it {
				ef := v.Item(it, i)
				if bl, ok := l.(Array); ok {
					WalkBlock(v, bl, i)
				} else {
					v.NoArray(l, i)
				}
				if ef != nil {
					ef()
				}
			}
		} else {
			v.NoArray(iVal, pos)
		}
	}
}
