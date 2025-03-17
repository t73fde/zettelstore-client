//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import (
	"t73f.de/r/sx"
)

// Visitor is walking the sx-based AST.
type Visitor interface {
	VisitBefore(node *sx.Pair, env *sx.Pair) (sx.Object, bool)
	VisitAfter(node *sx.Pair, env *sx.Pair) sx.Object
}

// Walk a sx-based AST through a Visitor.
func Walk(v Visitor, node *sx.Pair, env *sx.Pair) sx.Object {
	if node == nil {
		return nil
	}
	if result, ok := v.VisitBefore(node, env); ok {
		return result
	}

	if sym, isSymbol := sx.GetSymbol(node.Car()); isSymbol {
		if fn, found := mapChildrenWalk[sym]; found {
			node = fn(v, node, env)
		}
	}
	return v.VisitAfter(node, env)
}

var mapChildrenWalk map[*sx.Symbol]func(Visitor, *sx.Pair, *sx.Pair) *sx.Pair

func init() {
	mapChildrenWalk = map[*sx.Symbol]func(Visitor, *sx.Pair, *sx.Pair) *sx.Pair{
		SymBlock:         walkChildrenTail,
		SymPara:          walkChildrenTail,
		SymRegionBlock:   walkChildrenRegion,
		SymRegionQuote:   walkChildrenRegion,
		SymRegionVerse:   walkChildrenRegion,
		SymHeading:       walkHeadingChildren,
		SymListOrdered:   walkListChildren,
		SymListUnordered: walkListChildren,
		SymListQuote:     walkListChildren,
		SymDescription:   walkDescriptionChildren,
		SymTable:         walkTableChildren,
		SymCell:          walkCellChildren,
		SymTransclude:    walkChildrenInlines4,
		SymBLOB:          walkBLOBChildren,

		SymInline:       walkChildrenTail,
		SymEndnote:      walkChildrenInlines3,
		SymMark:         walkMarkChildren,
		SymLink:         walkChildrenInlines4,
		SymEmbed:        walkEmbedChildren,
		SymCite:         walkChildrenInlines4,
		SymFormatDelete: walkChildrenInlines3,
		SymFormatEmph:   walkChildrenInlines3,
		SymFormatInsert: walkChildrenInlines3,
		SymFormatMark:   walkChildrenInlines3,
		SymFormatQuote:  walkChildrenInlines3,
		SymFormatStrong: walkChildrenInlines3,
		SymFormatSpan:   walkChildrenInlines3,
		SymFormatSub:    walkChildrenInlines3,
		SymFormatSuper:  walkChildrenInlines3,
	}
}

func walkChildrenTail(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	hasNil := false
	for n := range node.Tail().Pairs() {
		obj := Walk(v, n.Head(), env)
		if sx.IsNil(obj) {
			hasNil = true
		}
		n.SetCar(obj)
	}
	if !hasNil {
		return node
	}
	for n := node; ; {
		next := n.Tail()
		if next == nil {
			break
		}
		if sx.IsNil(next.Car()) {
			n.SetCdr(next.Cdr())
			continue
		}
		n = next
	}
	return node
}

func walkChildrenList(v Visitor, lst *sx.Pair, env *sx.Pair) *sx.Pair {
	hasNil := false
	for n := range lst.Pairs() {
		obj := Walk(v, n.Head(), env)
		if sx.IsNil(obj) {
			hasNil = true
		}
		n.SetCar(obj)
	}
	if !hasNil {
		return lst
	}
	var result sx.ListBuilder
	for obj := range lst.Values() {
		if !sx.IsNil(obj) {
			result.Add(obj)
		}
	}
	return result.List()
}

func walkChildrenRegion(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := node.Car()
	next := node.Tail()
	// attrs := next.Car()
	next = next.Tail()
	next.SetCar(walkChildrenList(v, next.Head(), env))
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return node
}

func walkHeadingChildren(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := node.Car()
	next := node.Tail()
	// level := next.Car()
	next = next.Tail()
	// attrs := next.Car()
	next = next.Tail()
	// slug := next.Car()
	next = next.Tail()
	// fragment := next.Car()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return node
}
func walkListChildren(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := node.Car()
	next := node.Tail()
	// attrs := next.Car()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return node
}

func walkDescriptionChildren(v Visitor, dn *sx.Pair, env *sx.Pair) *sx.Pair {
	for n := dn.Tail().Tail(); n != nil; n = n.Tail() {
		n.SetCar(walkChildrenList(v, n.Head(), env))
		n = n.Tail()
		if n == nil {
			break
		}
		n.SetCar(Walk(v, n.Head(), env))
	}
	return dn
}

func walkTableChildren(v Visitor, tn *sx.Pair, env *sx.Pair) *sx.Pair {
	for row := range tn.Tail().Pairs() {
		row.SetCar(walkChildrenList(v, row.Head(), env))
	}
	return tn
}

func walkCellChildren(v Visitor, cn *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := cn.Car()
	next := cn.Tail()
	// attrs := next.Head()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return cn
}

func walkBLOBChildren(v Visitor, bn *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := bn.Car()
	next := bn.Tail()
	// attrs := next.Car()
	next = next.Tail()
	// description := next.Car()
	next.SetCar(walkChildrenList(v, next.Head(), env))
	return bn
}

func walkMarkChildren(v Visitor, mn *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := mn.Car()
	next := mn.Tail()
	// mark := next.Car()
	next = next.Tail()
	// slug := next.Car()
	next = next.Tail()
	// fragment := next.Car()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return mn
}

func walkEmbedChildren(v Visitor, en *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := en.Car()
	next := en.Tail()
	// attr := next.Car()
	next = next.Tail()
	// ref := next.Car()
	next = next.Tail()
	// syntax := next.Car()

	// text-list := next.Tail()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return en
}

func walkChildrenInlines4(v Visitor, ln *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := ln.Car()
	next := ln.Tail()
	// attrs := next.Car()
	next = next.Tail()
	// val3 := next.Car()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return ln
}

func walkChildrenInlines3(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := node.Car()
	next := node.Tail() // Attrs
	// attrs := next.Car()
	next.SetCdr(walkChildrenList(v, next.Tail(), env))
	return node
}
