//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of Zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

// cleaner provides functions to clean up the parsed AST.

import (
	"strconv"

	"t73f.de/r/sx"
	zerostrings "t73f.de/r/zero/strings"
	"t73f.de/r/zsc/text"
	"t73f.de/r/zsx"
)

// AssignIdentifier the given SZ syntax tree.
func AssignIdentifier(node *sx.Pair) {
	v1 := assignPhase1{ids: idsNode{}}
	zsx.WalkIt(&v1, node, nil)
	if v1.hasMark {
		v2 := assignPhase2{ids: v1.ids}
		zsx.WalkIt(&v2, node, nil)
	}
}

type assignPhase1 struct {
	ids     idsNode
	hasMark bool // Mark nodes will be cleaned in phase 2 only
}

func (v *assignPhase1) VisitItBefore(node *sx.Pair, _ *sx.Pair) bool {
	if sym, isSymbol := sx.GetSymbol(node.Car()); isSymbol {
		switch sym {
		case zsx.SymHeading:
			levelNode := node.Tail().Tail()
			textNode := levelNode.Tail()
			if s := text.EvaluateInlineString(textNode); s != "" {
				v.ids.setNodeID(node, s)
			}
		case zsx.SymMark:
			v.hasMark = true
		}
	}
	return false
}
func (v *assignPhase1) VisitItAfter(*sx.Pair, *sx.Pair) {}

type assignPhase2 struct {
	ids idsNode
}

func (v *assignPhase2) VisitItBefore(node *sx.Pair, _ *sx.Pair) bool {
	if sym, isSymbol := sx.GetSymbol(node.Car()); isSymbol && sym.IsEqualSymbol(zsx.SymMark) {
		stringNode := node.Tail().Tail()
		if markString, isString := sx.GetString(stringNode.Car()); isString {
			v.ids.setNodeID(node, markString.GetValue())
		}
	}
	return false
}
func (v *assignPhase2) VisitItAfter(*sx.Pair, *sx.Pair) {}

type idsNode map[string]*sx.Pair

func (ids idsNode) setNodeID(node *sx.Pair, text string) {
	attrsNode := node.Tail()
	slugText := zerostrings.Slugify(text)
	fragText := ids.addIdentifier(slugText, node)
	attrs := attrsNode.Head().RemoveAssoc(zsx.SymSpecialID)
	attrs = sx.Cons(sx.Cons(zsx.SymSpecialID, sx.MakeString(fragText)), attrs)
	attrsNode.SetCar(attrs)
}

func (ids idsNode) addIdentifier(id string, node *sx.Pair) string {
	if n, ok := ids[id]; ok && n != node {
		prefix := id + "-"
		for count := 1; ; count++ {
			newID := prefix + strconv.Itoa(count)
			if n2, ok2 := ids[newID]; !ok2 || n2 == node {
				ids[newID] = node
				return newID
			}
		}
	}
	ids[id] = node
	return id
}
