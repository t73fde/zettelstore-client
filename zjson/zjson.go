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

import (
	"encoding/json"
	"fmt"
)

// Value is the gerneric JSON value.
type Value = interface{}

// Array represents a JSON array.
type Array = []Value

// Object represents a JSON object.
type Object = map[string]Value

// Visitor provides functionality when a Value is traversed.
type Visitor interface {
	Block(a Array, pos int) (bool, EndFunc)
	Inline(a Array, pos int) (bool, EndFunc)
	Item(a Array, pos int) (bool, EndFunc)
	Object(t string, obj Object, pos int) (bool, EndFunc)

	NoValue(val Value, pos int)
	NoArray(val Value, pos int)
	NoObject(obj Object, pos int)
}

// EndFunc is a function that executes after a ZJSON element is visited.
type EndFunc func()

// WalkBlock traverses a block array.
func WalkBlock(v Visitor, a Array, pos int) {
	children, ef := v.Block(a, pos)
	if children {
		for i, elem := range a {
			WalkObject(v, elem, i)
		}
	}
	if ef != nil {
		ef()
	}
}

// WalkInline traverses an inline array.
func WalkInline(v Visitor, a Array, pos int) {
	children, ef := v.Inline(a, pos)
	if children {
		for i, elem := range a {
			WalkObject(v, elem, i)
		}
	}
	if ef != nil {
		ef()
	}
}

// WalkObject traverses a value as a JSON object.
func WalkObject(v Visitor, val Value, pos int) {
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
		WalkInlineChild(v, obj, pos)
		WalkBlockChild(v, obj, pos)
		WalkItemChild(v, obj, pos)
	}
	if ef != nil {
		ef()
	}
}

// WalkInlineChild traverses the array found at the name NameInline ('i').
func WalkInlineChild(v Visitor, obj Object, pos int) {
	if iVal, ok := obj[NameInline]; ok {
		if il, ok := iVal.(Array); ok {
			WalkInline(v, il, 0)
		} else {
			v.NoArray(iVal, pos)
		}
	}
}

// WalkBlockChild traverses the array found at the name NameBlock ('b').
func WalkBlockChild(v Visitor, obj Object, pos int) {
	if bVal, ok := obj[NameBlock]; ok {
		if bl, ok := bVal.(Array); ok {
			WalkBlock(v, bl, 0)
		} else {
			v.NoArray(bVal, pos)
		}
	}
}

// WalkItemChild traverses the arrays found at the name NameList ('c').
func WalkItemChild(v Visitor, obj Object, pos int) {
	iVal, ok := obj[NameList]
	if !ok {
		return
	}
	it, ok := iVal.(Array)
	if !ok {
		v.NoArray(iVal, pos)
		return
	}
	for i, l := range it {
		children, ef := v.Item(it, i)
		if !children {
			continue
		}
		if bl, ok := l.(Array); ok {
			WalkBlock(v, bl, 0)
		} else {
			v.NoArray(l, i)
		}
		if ef != nil {
			ef()
		}
	}
}

// GetArray returns the array-typed value under the given name.
func GetArray(obj Object, name string) Array {
	if v, ok := obj[name]; ok && v != nil {
		if a, ok := v.(Array); ok {
			return a
		}
	}
	return nil
}

// GetNumber returns the numeric value at NameNumberic ('n') as a string.
func GetNumber(obj Object) string {
	if v, ok := obj[NameNumeric]; ok {
		if n, ok := v.(json.Number); ok {
			return string(n)
		}
		if f, ok := v.(float64); ok {
			return fmt.Sprint(f)
		}
	}
	return ""
}

// GetString returns the string value at the given name.
func GetString(obj Object, name string) string {
	if v, ok := obj[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
