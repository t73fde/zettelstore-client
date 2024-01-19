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
	sz.SymLiteralHTML:    {},
	sz.SymLiteralInput:   {},
	sz.SymLiteralMath:    {},
	sz.SymLiteralProg:    {},
	sz.SymLiteralOutput:  {},
	sz.SymLiteralZettel:  {},
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
	length := lst.Length() - 1
	if length < 0 {
		return nil
	}
	vector := make([]*sx.Pair, 0, length)
	// 1st phase: process all childs, ignore SPACE at start, and merge some elements
	for node := lst.Tail(); node != nil; node = node.Tail() {
		elem, isPair := sx.GetPair(node.Car())
		if isPair {
			elem = postProcess(elem)
		}
		if elem == nil {
			continue
		}
		elemSym := elem.Car()
		if len(vector) == 0 {
			// The 1st element is always moved, except for a SPACE
			if elemSym.IsEqual(sz.SymSpace) {
				continue
			}
			vector = append(vector, elem)
			continue
		}
		last := vector[len(vector)-1]
		lastSym := last.Car()

		if lastSym.IsEqual(sz.SymText) && elemSym.IsEqual(sz.SymText) {
			// Merge two TEXT elements into one
			lastText := last.Tail().Car().(sx.String)
			elemText := elem.Tail().Car().(sx.String)
			last.SetCdr(sx.Cons(lastText+elemText, sx.Nil()))
			continue
		}

		if lastSym.IsEqual(sz.SymSpace) && elemSym.IsEqual(sz.SymSoft) {
			// Merge (SPACE) (SOFT) to (HARD)
			vector[len(vector)-1] = sx.Cons(sz.SymHard, sx.Nil())
			continue
		}

		vector = append(vector, elem)
	}
	if len(vector) == 0 {
		return nil
	}

	// 2nd phase: remove (SOFT), (HARD), (SPACE) at the end
	lastPos := len(vector) - 1
	for lastPos >= 0 {
		elem := vector[lastPos]
		elemSym := elem.Car()
		if !elemSym.IsEqual(sz.SymSpace) && !elemSym.IsEqual(sz.SymSoft) && !elemSym.IsEqual(sz.SymHard) {
			break
		}
		lastPos--
	}
	if lastPos < 0 {
		return nil
	}

	result := sx.Cons(lst.Car(), sx.Nil())
	curr := result
	for i := 0; i <= lastPos; i++ {
		curr = curr.AppendBang(vector[i])
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
