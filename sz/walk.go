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
	"log"

	"zettelstore.de/sx.fossil"
)

// Visitor is walking the sx-based AST.
type Visitor interface {
	Visit(node *sx.Pair, env *sx.Pair) sx.Object
}

// Walk a sx-based AST through a Visitor.
func Walk(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	if node == nil {
		return nil
	}
	if result, isPair := sx.GetPair(v.Visit(node, env)); isPair {
		return result
	}
	return WalkChildren(v, node, env)
}

// WalkChildren will walk all child nodes.
func WalkChildren(v Visitor, node *sx.Pair, env *sx.Pair) *sx.Pair {
	if sym, isSymbol := sx.GetSymbol(node.Car()); isSymbol {
		if fn, found := mapChildrenWalk[sym.GetValue()]; found {
			return fn(v, node, env)
		}
		log.Println("MISS", sym, node)
		return node
	}
	panic(node)
}

var mapChildrenWalk map[string]func(Visitor, *sx.Pair, *sx.Pair) *sx.Pair

func init() {
	mapChildrenWalk = map[string]func(Visitor, *sx.Pair, *sx.Pair) *sx.Pair{
		NameBlock:         walkChildrenTail,
		NamePara:          walkChildrenTail,
		NameRegionBlock:   walkChildrenRegion,
		NameRegionQuote:   walkChildrenRegion,
		NameRegionVerse:   walkChildrenRegion,
		NameHeading:       walkChildrenHeading,
		NameListOrdered:   walkChildrenTail,
		NameListUnordered: walkChildrenTail,
		NameListQuote:     walkChildrenTail,
		NameDescription:   walkChildrenDescription,
		NameTable:         walkChildrenTable,

		NameInline:       walkChildrenTail,
		NameEndnote:      walkChildrenInlines3,
		NameMark:         walkChildrenMark,
		NameLinkBased:    walkChildrenInlines4,
		NameLinkBroken:   walkChildrenInlines4,
		NameLinkExternal: walkChildrenInlines4,
		NameLinkFound:    walkChildrenInlines4,
		NameLinkHosted:   walkChildrenInlines4,
		NameLinkInvalid:  walkChildrenInlines4,
		NameLinkQuery:    walkChildrenInlines4,
		NameLinkSelf:     walkChildrenInlines4,
		NameLinkZettel:   walkChildrenInlines4,
		NameEmbed:        walkChildrenInlines4,
		NameCite:         walkChildrenInlines4,
		NameFormatDelete: walkChildrenInlines3,
		NameFormatEmph:   walkChildrenInlines3,
		NameFormatInsert: walkChildrenInlines3,
		NameFormatMark:   walkChildrenInlines3,
		NameFormatQuote:  walkChildrenInlines3,
		NameFormatStrong: walkChildrenInlines3,
		NameFormatSpan:   walkChildrenInlines3,
		NameFormatSub:    walkChildrenInlines3,
		NameFormatSuper:  walkChildrenInlines3,
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
	header := tn.Tail()
	header.SetCar(walkChildrenList(v, header.Tail(), env))
	for row := header.Tail(); row != nil; row = row.Tail() {
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
