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

import "t73f.de/r/sx"

// Visitor is walking the sx-based AST.
type Visitor interface {
	VisitBefore(node *sx.Pair, env *sx.Pair) (sx.Object, bool)
	VisitAfter(node *sx.Pair, env *sx.Pair) (sx.Object, bool)
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
			if result, ok := v.VisitAfter(node, env); ok {
				return result
			}
		}
	}
	return node
}

var mapChildrenWalk map[*sx.Symbol]func(Visitor, *sx.Pair, *sx.Pair) *sx.Pair

func init() {
	mapChildrenWalk = map[*sx.Symbol]func(Visitor, *sx.Pair, *sx.Pair) *sx.Pair{
		SymBlock:         walkChildrenTail,
		SymPara:          walkChildrenTail,
		SymRegionBlock:   walkChildrenRegion,
		SymRegionQuote:   walkChildrenRegion,
		SymRegionVerse:   walkChildrenRegion,
		SymHeading:       walkChildrenHeading,
		SymListOrdered:   walkChildrenTail,
		SymListUnordered: walkChildrenTail,
		SymListQuote:     walkChildrenTail,
		SymDescription:   walkChildrenDescription,
		SymTable:         walkChildrenTable,

		SymInline:       walkChildrenTail,
		SymEndnote:      walkChildrenInlines3,
		SymMark:         walkChildrenMark,
		SymLinkBased:    walkChildrenInlines4,
		SymLinkBroken:   walkChildrenInlines4,
		SymLinkExternal: walkChildrenInlines4,
		SymLinkFound:    walkChildrenInlines4,
		SymLinkHosted:   walkChildrenInlines4,
		SymLinkInvalid:  walkChildrenInlines4,
		SymLinkQuery:    walkChildrenInlines4,
		SymLinkSelf:     walkChildrenInlines4,
		SymLinkZettel:   walkChildrenInlines4,
		SymEmbed:        walkChildrenEmbed,
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
	for n := node.Tail(); n != nil; n = n.Tail() {
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
	for n := lst; n != nil; n = n.Tail() {
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
	for n := lst; n != nil; n = n.Tail() {
		obj := n.Car()
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

func walkChildrenHeading(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
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

func walkChildrenDescription(v Visitor, dn *sx.Pair, env *sx.Pair) *sx.Pair {
	for n := dn.Tail(); n != nil; n = n.Tail() {
		n.SetCar(walkChildrenList(v, n.Head(), env))
		n = n.Tail()
		if n == nil {
			break
		}
		n.SetCar(Walk(v, n.Head(), env))
	}
	return dn
}

func walkChildrenTable(v Visitor, tn *sx.Pair, env *sx.Pair) *sx.Pair {
	for row := tn.Tail(); row != nil; row = row.Tail() {
		row.SetCar(walkChildrenList(v, row.Head(), env))
	}
	return tn
}

func walkChildrenMark(v Visitor, mn *sx.Pair, env *sx.Pair) *sx.Pair {
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

func walkChildrenEmbed(v Visitor, en *sx.Pair, env *sx.Pair) *sx.Pair {
	// sym := en.Car()
	next := en.Tail()
	// attr := next.Car()
	next = next.Tail()
	// ref := next.Car()
	next = next.Tail()
	// syntax := next.Car()
	next = next.Tail()
	if next != nil {
		// text := next.Car()
		next.SetCar(Walk(v, next.Head(), env))
	}
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
