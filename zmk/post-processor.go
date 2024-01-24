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
	"strings"

	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

const symInVerse = sx.Symbol("in-verse")

func postProcess(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	if lst == nil {
		return nil
	}
	sym, isSym := sx.GetSymbol(lst.Car())
	if !isSym {
		panic(lst)
	}
	if fn, found := symMap[sym]; found {
		return fn(lst, env)
	}
	if _, found := ignoreMap[sym]; found {
		return lst
	}
	panic(lst)
}

func postProcessList(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	var result, curr *sx.Pair
	for node := lst; node != nil; node = node.Tail() {
		elem, isPair := sx.GetPair(node.Car())
		if isPair {
			elem = postProcess(elem, env)
		}
		if elem == nil {
			continue
		}
		if result == nil {
			result = sx.Cons(elem, nil)
			curr = result
		} else {
			curr = curr.AppendBang(elem)
		}
	}
	return result
}

var ignoreMap = map[sx.Symbol]struct{}{
	sz.SymThematic:   {},
	sz.SymTransclude: {},

	sz.SymLiteralComment: {},
	sz.SymLiteralHTML:    {},
	sz.SymLiteralInput:   {},
	sz.SymLiteralMath:    {},
	sz.SymLiteralProg:    {},
	sz.SymLiteralOutput:  {},
	sz.SymLiteralZettel:  {},
	sz.SymSpace:          {},
	sz.SymHard:           {},
}

var symMap map[sx.Symbol]func(*sx.Pair, *sx.Pair) *sx.Pair

func init() {
	symMap = map[sx.Symbol]func(*sx.Pair, *sx.Pair) *sx.Pair{
		sz.SymBlock:           postProcessBlockList,
		sz.SymPara:            postProcessInlineList,
		sz.SymRegionBlock:     postProcessRegion,
		sz.SymRegionQuote:     postProcessRegion,
		sz.SymRegionVerse:     postProcessRegionVerse,
		sz.SymVerbatimComment: postProcessVerbatim,
		sz.SymVerbatimEval:    postProcessVerbatim,
		sz.SymVerbatimMath:    postProcessVerbatim,
		sz.SymVerbatimProg:    postProcessVerbatim,
		sz.SymVerbatimZettel:  postProcessVerbatim,
		sz.SymHeading:         postProcessHeading,
		sz.SymTable:           postProcessTable,

		sz.SymInline:       postProcessInlineList,
		sz.SymText:         postProcessText,
		sz.SymSoft:         postProcessSoft,
		sz.SymEndnote:      postProcessEndnote,
		sz.SymMark:         postProcessMark,
		sz.SymLinkBased:    postProcessInlines4,
		sz.SymLinkBroken:   postProcessInlines4,
		sz.SymLinkExternal: postProcessInlines4,
		sz.SymLinkFound:    postProcessInlines4,
		sz.SymLinkHosted:   postProcessInlines4,
		sz.SymLinkInvalid:  postProcessInlines4,
		sz.SymLinkQuery:    postProcessInlines4,
		sz.SymLinkSelf:     postProcessInlines4,
		sz.SymLinkZettel:   postProcessInlines4,
		sz.SymEmbed:        postProcessInlines4,
		sz.SymCite:         postProcessInlines4,
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

func postProcessBlockList(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	result := postProcessList(lst.Tail(), env)
	if result == nil {
		return nil
	}
	return result.Cons(lst.Car())
}

func postProcessInlineList(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := lst.Car()
	if rest := postProcessInlines(lst.Tail(), env); rest != nil {
		return rest.Cons(sym)
	}
	return nil
}

func postProcessRegion(rn *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := rn.Car()
	next := rn.Tail()
	attrs := next.Car()
	next = next.Tail()
	blocks := postProcessList(next.Head(), env)
	text := postProcessInlines(next.Tail(), env)
	if blocks == nil && text == nil {
		return nil
	}
	return text.Cons(blocks).Cons(attrs).Cons(sym)
}

func postProcessRegionVerse(rn *sx.Pair, env *sx.Pair) *sx.Pair {
	return postProcessRegion(rn, env.Cons(sx.Cons(symInVerse, nil)))
}

func postProcessVerbatim(verb *sx.Pair, _ *sx.Pair) *sx.Pair {
	if content, isString := sx.GetString(verb.Tail().Tail().Car()); isString && content != "" {
		return verb
	}
	return nil
}

func postProcessHeading(hn *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := hn.Car()
	next := hn.Tail()
	level := next.Car()
	next = next.Tail()
	attrs := next.Car()
	next = next.Tail()
	slug := next.Car()
	next = next.Tail()
	fragment := next.Car()
	if text := postProcessInlines(next.Tail(), env); text != nil {
		return text.Cons(fragment).Cons(slug).Cons(attrs).Cons(level).Cons(sym)
	}
	return nil
}

func postProcessTable(tbl *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := tbl.Car()
	next := tbl.Tail()
	header := next.Car()
	if header != nil {
		// Already post-processed
		return tbl
	}
	rows, width := postProcessRows(next.Tail(), env)
	if rows == nil {
		// Header and row are nil -> no table
		return nil
	}
	header, rows, _ = splitTableHeader(rows, width)
	return rows.Cons(header).Cons(sym)
}

func postProcessRows(rows *sx.Pair, env *sx.Pair) (*sx.Pair, int) {
	maxWidth := 0
	var result, curr *sx.Pair
	for node := rows; node != nil; node = node.Tail() {
		row := node.Head()
		row, width := postProcessCells(row, env)
		if maxWidth < width {
			maxWidth = width
		}
		if result == nil {
			result = sx.Cons(row, nil)
			curr = result
		} else {
			curr = curr.AppendBang(row)
		}
	}
	return result, maxWidth
}

func postProcessCells(cells *sx.Pair, env *sx.Pair) (*sx.Pair, int) {
	width := 0
	var result, curr *sx.Pair
	for node := cells; node != nil; node = node.Tail() {
		cell := node.Head()
		ins := postProcessInlines(cell.Tail(), env)
		newCell := ins.Cons(cell.Car())
		if result == nil {
			result = sx.Cons(newCell, nil)
			curr = result
		} else {
			curr = curr.AppendBang(newCell)
		}
		width++
	}
	return result, width
}

func splitTableHeader(rows *sx.Pair, width int) (header, realRows *sx.Pair, align []sx.Symbol) {
	align = make([]sx.Symbol, width)

	foundHeader := false

	var lastCellNode *sx.Pair
	var cellCount int

	// assert: rows != nil (checked in postProcessTable)
	for node := rows.Head(); node != nil; node = node.Tail() {
		lastCellNode = node
		cellCount++
		cell := node.Head()
		cellTail := cell.Tail()
		if cellTail == nil {
			continue
		}

		// elem is first cell inline element
		elem := cellTail.Head()
		if elem.Car().IsEqual(sz.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString {
				if strings.HasPrefix(string(s), "=") {
					foundHeader = true
					elem.SetCdr(sx.Cons(sx.String(strings.TrimPrefix(string(s), "=")), nil))
				}
			}
		}

		// move to the last cell inline element
		for {
			next := elem.Tail()
			if next == nil {
				break
			}
			elem = next
		}

		if elem.Car().IsEqual(sz.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString && s != "" {
				cellAlign := getCellAlignment(s[len(s)-1])
				if cellAlign != sz.SymCell {
					elem.SetCdr(s[0 : len(s)-1])
				}
				align[cellCount-1] = cellAlign
			}
		}
	}

	if !foundHeader {
		for i := 0; i < width; i++ {
			align[i] = sz.SymCell // Default alignment
		}
		return nil, rows, align
	}

	for cellCount < width {
		lastCellNode = lastCellNode.AppendBang(sx.Cons(sz.SymCell, nil))
	}
	for i := 0; i < width; i++ {
		if align[i] == "" {
			align[i] = sz.SymCell // Default alignment
		}
	}
	return rows.Head(), rows.Tail(), align
}

func getCellAlignment(ch byte) sx.Symbol {
	switch ch {
	case ':':
		return sz.SymCellCenter
	case '<':
		return sz.SymCellLeft
	case '>':
		return sz.SymCellRight
	default:
		return sz.SymCell
	}
}

func postProcessInlines(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	length := lst.Length()
	if length <= 0 {
		return nil
	}
	inVerse := env.Assoc(symInVerse) != nil
	vector := make([]*sx.Pair, 0, length)
	// 1st phase: process all childs, ignore SPACE at start, and merge some elements
	for node := lst; node != nil; node = node.Tail() {
		elem, isPair := sx.GetPair(node.Car())
		if isPair {
			elem = postProcess(elem, env)
		}
		if elem == nil {
			continue
		}
		elemSym := elem.Car()
		if len(vector) == 0 {
			// The 1st element is always moved, except for a SPACE outside a verse block
			if !inVerse && elemSym.IsEqual(sz.SymSpace) {
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

	result := sx.Cons(vector[0], nil)
	curr := result
	for i := 1; i <= lastPos; i++ {
		curr = curr.AppendBang(vector[i])
	}
	return result
}

func postProcessText(txt *sx.Pair, _ *sx.Pair) *sx.Pair {
	if tail := txt.Tail(); tail != nil {
		if content, isString := sx.GetString(tail.Car()); isString && content != "" {
			return txt
		}
	}
	return nil
}

func postProcessSoft(sn *sx.Pair, env *sx.Pair) *sx.Pair {
	if env.Assoc(symInVerse) == nil {
		return sn
	}
	return sx.Cons(sz.SymHard, nil)
}

func postProcessEndnote(en *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := en.Car()
	next := en.Tail()
	attrs := next.Car()
	if text := postProcessInlines(next.Tail(), env); text != nil {
		return text.Cons(attrs).Cons(sym)
	}
	return sx.MakeList(sym, attrs)
}

func postProcessMark(en *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := en.Car()
	next := en.Tail()
	mark := next.Car()
	next = next.Tail()
	slug := next.Car()
	next = next.Tail()
	fragment := next.Car()
	text := postProcessInlines(next.Tail(), env)
	return text.Cons(fragment).Cons(slug).Cons(mark).Cons(sym)
}

func postProcessInlines4(ln *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := ln.Car()
	next := ln.Tail()
	attrs := next.Car()
	next = next.Tail()
	val3 := next.Car()
	text := postProcessInlines(next.Tail(), env)
	return text.Cons(val3).Cons(attrs).Cons(sym)
}

func postProcessFormat(fn *sx.Pair, env *sx.Pair) *sx.Pair {
	symFormat := fn.Car()
	next := fn.Tail() // Attrs
	attrs := next.Car()
	next = next.Tail() // Possible inlines
	if next == nil {
		return fn
	}
	inlines := postProcessInlines(next, env)
	return inlines.Cons(attrs).Cons(symFormat)
}
