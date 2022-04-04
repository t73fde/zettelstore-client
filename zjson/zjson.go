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

// GetMetaContent returns the metadata and the content of a zettel ZJSON.
func GetMetaContent(zjZettel Value) (Meta, Array) {
	if zettel, ok := zjZettel.(Object); ok {
		meta := MakeMeta(zettel["meta"])
		content := MakeArray(zettel["content"])
		return meta, content
	}
	return nil, nil
}

// Visitor provides functionality when a Value is traversed.
type Visitor interface {
	BlockArray(a Array, pos int) CloseFunc
	InlineArray(a Array, pos int) CloseFunc
	ItemArray(a Array, pos int) CloseFunc

	BlockObject(t string, obj Object, pos int) (bool, CloseFunc)
	InlineObject(t string, obj Object, pos int) (bool, CloseFunc)

	Unexpected(val Value, pos int, exp string)
}

// CloseFunc is a function that executes after a ZJSON element is visited.
type CloseFunc func()

// WalkBlock traverses a block array.
func WalkBlock(v Visitor, a Array, pos int) {
	ef := v.BlockArray(a, pos)
	for i, elem := range a {
		WalkBlockObject(v, elem, i)
	}
	if ef != nil {
		ef()
	}
}

// WalkInline traverses an inline array.
func WalkInline(v Visitor, a Array, pos int) {
	ef := v.InlineArray(a, pos)
	for i, elem := range a {
		WalkInlineObject(v, elem, i)
	}
	if ef != nil {
		ef()
	}
}

// WalkBlockObject traverses a value as a JSON object in a block array.
func WalkBlockObject(v Visitor, val Value, pos int) { walkObject(v, val, pos, v.BlockObject) }

// WalkInlineObject traverses a value as a JSON object in an inline array.
func WalkInlineObject(v Visitor, val Value, pos int) { walkObject(v, val, pos, v.InlineObject) }

func walkObject(v Visitor, val Value, pos int, objFunc func(string, Object, int) (bool, CloseFunc)) {
	obj, ok := val.(Object)
	if !ok {
		v.Unexpected(val, pos, "Object")
		return
	}

	tVal, ok := obj[NameType]
	if !ok {
		v.Unexpected(obj, pos, "Object type")
		return
	}
	t, ok := tVal.(string)
	if !ok {
		v.Unexpected(obj, pos, "Object type value")
		return
	}

	doChilds, ef := objFunc(t, obj, pos)
	if doChilds {
		WalkBlockChild(v, obj, pos)
		WalkItemChild(v, obj, pos)
		WalkInlineChild(v, obj, pos)
		walkDescriptionList(v, obj)
		walkTable(v, obj, pos)
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
			v.Unexpected(iVal, pos, "Inline array")
		}
	}
}

// WalkBlockChild traverses the array found at the name NameBlock ('b').
func WalkBlockChild(v Visitor, obj Object, pos int) {
	if bVal, ok := obj[NameBlock]; ok {
		if bl, ok := bVal.(Array); ok {
			WalkBlock(v, bl, 0)
		} else {
			v.Unexpected(bVal, pos, "Block array")
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
		v.Unexpected(iVal, pos, "Item array")
		return
	}
	for i, l := range it {
		ef := v.ItemArray(it, i)
		if bl, ok := l.(Array); ok {
			WalkBlock(v, bl, i)
		} else {
			v.Unexpected(l, i, "Item block array")
		}
		if ef != nil {
			ef()
		}
	}
}

func walkDescriptionList(v Visitor, obj Object) {
	descrs := GetArray(obj, NameDescrList)
	if len(descrs) == 0 {
		return
	}
	for i, elem := range descrs {
		dObj := MakeObject(elem)
		if dObj == nil {
			continue
		}
		WalkInlineChild(v, dObj, i)
		descr := GetArray(dObj, NameDescription)
		if len(descr) == 0 {
			continue
		}
		for j, ddv := range descr {
			dd := MakeArray(ddv)
			if len(dd) == 0 {
				continue
			}
			WalkBlock(v, dd, j)
		}
	}
}

func walkTable(v Visitor, obj Object, pos int) {
	tdata := GetArray(obj, NameTable)
	if len(tdata) == 0 {
		return
	}
	if len(tdata) != 2 {
		v.Unexpected(obj, pos, "Table header/rows")
		return
	}
	walkRow(v, MakeArray(tdata[0]))
	if bArray := MakeArray(tdata[1]); len(bArray) > 0 {
		for _, row := range bArray {
			if rArray := MakeArray(row); rArray != nil {
				walkRow(v, rArray)
			}
		}
	}
}
func walkRow(v Visitor, row Array) {
	if len(row) > 0 {
		for _, cell := range row {
			if cObj := MakeObject(cell); cObj != nil {
				WalkInlineChild(v, cObj, 0)
			}
		}
	}
}

// GetArray returns the array-typed value under the given name.
func GetArray(obj Object, name string) Array {
	if v, ok := obj[name]; ok && v != nil {
		return MakeArray(v)
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
		return MakeString(v)
	}
	return ""
}

// MakeArray returns the given value as a JSON array.
func MakeArray(val Value) Array {
	if a, ok := val.(Array); ok {
		return a
	}
	return nil
}

// MakeString returns the given value as a string.
func MakeString(val Value) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetAttribute returns a copy of the attributes of the given object.
func GetAttributes(obj Object) Attributes {
	a := GetObject(obj, NameAttribute)
	if len(a) == 0 {
		return nil
	}
	result := make(Attributes, len(a))
	for n, v := range a {
		if val, ok := v.(string); ok {
			result[n] = val
		}
	}
	return result
}

// SetAttributes copies the attributes to the given object.
func SetAttributes(obj Object, a Attributes) {
	if len(a) == 0 {
		delete(obj, NameAttribute)
	}
	val := make(Object)
	for k, v := range a {
		val[k] = v
	}
	obj[NameAttribute] = val
}

// GetObject returns the object found at the given object with the given name.
func GetObject(obj Object, name string) Object {
	if v, ok := obj[name]; ok && v != nil {
		return MakeObject(v)
	}
	return nil
}

// MakeObject returns the given value as a JSON object.
func MakeObject(val Value) Object {
	if o, ok := val.(Object); ok {
		return o
	}
	return nil
}

// GetParagraphInline return the inline list of the first paragraph (or nil if there is no such thing)
func GetParagraphInline(a Array) Array {
	if len(a) != 1 {
		return nil
	}
	if o := MakeObject(a[0]); o != nil {
		if GetString(o, NameType) == TypeParagraph {
			return GetArray(o, NameInline)
		}
	}
	return nil
}
