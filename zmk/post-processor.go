//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package zmk

import (
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

func postProcess(lst *sx.Pair) *sx.Pair {
	if lst == nil {
		return nil
	}
	sym, isSym := sx.GetSymbol(lst.Car())
	if !isSym {
		panic(lst)
	}
	if fn, found := symMap[sym]; found {
		return fn(lst)
	}
	if _, found := ignoreMap[sym]; found {
		return lst
	}
	panic(lst)
}

var ignoreMap = map[sx.Symbol]struct{}{
	sz.SymLiteralComment: {},
	sz.SymSpace:          {},
	sz.SymSoft:           {},
	sz.SymHard:           {},
}

var symMap map[sx.Symbol]func(*sx.Pair) *sx.Pair

func init() {
	symMap = map[sx.Symbol]func(*sx.Pair) *sx.Pair{
		sz.SymBlock:        postProcessBlockList,
		sz.SymPara:         postProcessInlineList,
		sz.SymInline:       postProcessInlineList,
		sz.SymText:         postProcessText,
		sz.SymFormatDelete: postProcessFormat,
		sz.SymFormatEmph:   postProcessFormat,
		sz.SymFormatInsert: postProcessFormat,
		sz.SymFormatMark:   postProcessFormat,
		sz.SymFormatQuote:  postProcessFormat,
		sz.SymFormatStrong: postProcessFormat,
		sz.SymFormatSpan:   postProcessFormat,
		sz.SymFormatSub:    postProcessFormat,
		sz.SymFormatSuper:  postProcessFormat,
	}
}

func postProcessBlockList(lst *sx.Pair) *sx.Pair {
	result := sx.Cons(lst.Car(), sx.Nil())
	curr := result
	for node := lst.Tail(); node != nil; node = node.Tail() {
		elem, isPair := sx.GetPair(node.Car())
		if isPair {
			elem = postProcess(elem)
		}
		if elem == nil {
			continue
		}
		curr = curr.AppendBang(elem)
	}

	if curr == result {
		// Empty block
		return nil
	}
	return result
}

func postProcessInlineList(lst *sx.Pair) *sx.Pair {
	temp := sx.Cons(lst.Car(), sx.Nil())
	curr := temp
	// 1st phase: process all childs and merge some elements
	for node := lst.Tail(); node != nil; node = node.Tail() {
		elem, isPair := sx.GetPair(node.Car())
		if isPair {
			elem = postProcess(elem)
		}
		if elem == nil {
			continue
		}
		if curr == temp {
			// The 1st element is always moved.
			curr = curr.AppendBang(elem)
			continue
		}
		last := curr.Head()
		lastSym := last.Car()
		elemSym := elem.Car()

		if lastSym.IsEqual(sz.SymText) && elemSym.IsEqual(sz.SymText) {
			// Merge two TEXT elements into one
			lastText := last.Tail().Car().(sx.String)
			elemText := elem.Tail().Car().(sx.String)
			last.SetCdr(sx.Cons(lastText+elemText, sx.Nil()))
			continue
		}

		if lastSym.IsEqual(sz.SymSpace) && elemSym.IsEqual(sz.SymSoft) {
			// Merge (SPACE) (SOFT) to (HARD)
			curr.SetCar(sx.Cons(sz.SymHard, sx.Nil()))
			continue
		}

		curr = curr.AppendBang(elem)
	}

	result := sx.Cons(lst.Car(), sx.Nil())
	curr = result
	// 2nd phase: remove (SPACE) at the start, and (SOFT), (HARD), (SPACE) at the end
	for node := temp.Tail(); node != nil; node = node.Tail() {
		elem := node.Head()
		elemSym := elem.Car()
		if curr == result {
			// We are at the start
			if elemSym.IsEqual(sz.SymSpace) {
				continue
			}
		}
		if node.Tail() != nil {
			// Not at the end, continue
			curr = curr.AppendBang(elem)
			continue
		}
		if elemSym.IsEqual(sz.SymSpace) || elemSym.IsEqual(sz.SymSoft) || elemSym.IsEqual(sz.SymHard) {
			break
		}
		curr = curr.AppendBang(elem)
		break
	}

	if curr == result {
		// Empty inline
		return nil
	}
	return result
}

func postProcessText(txt *sx.Pair) *sx.Pair {
	if tail := txt.Tail(); tail != nil {
		if content, isString := sx.GetString(tail.Car()); isString && content != "" {
			return txt
		}
	}
	return nil
}

func postProcessFormat(fn *sx.Pair) *sx.Pair {
	symFormat := fn.Car()
	next := fn.Tail() // Attrs
	attrs := next.Car()
	next = next.Tail() // Possible inlines
	if next == nil {
		return fn
	}
	inlines := postProcess(next.Cons(sz.SymInline))
	return inlines.Tail().Cons(attrs).Cons(symFormat)
}
