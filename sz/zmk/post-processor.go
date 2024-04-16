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
	"t73f.de/r/sx"
	"zettelstore.de/client.fossil/sz"
)

var symInVerse = sx.MakeSymbol("in-verse")
var symNoBlock = sx.MakeSymbol("no-block")

type postProcessor struct{}

func (pp *postProcessor) Visit(lst *sx.Pair, env *sx.Pair) sx.Object {
	if lst == nil {
		return nil
	}
	sym, isSym := sx.GetSymbol(lst.Car())
	if !isSym {
		panic(lst)
	}
	symVal := sym.GetValue()
	if fn, found := symMap[symVal]; found {
		return fn(pp, lst, env)
	}
	return sx.Int64(0)
}

func (pp *postProcessor) visitPairList(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	var pList sx.ListBuilder
	for node := lst; node != nil; node = node.Tail() {
		if elem := sz.Walk(pp, node.Head(), env); elem != nil {
			pList.Add(elem)
		}
	}
	return pList.List()
}

var symMap map[string]func(*postProcessor, *sx.Pair, *sx.Pair) *sx.Pair

func init() {
	symMap = map[string]func(*postProcessor, *sx.Pair, *sx.Pair) *sx.Pair{
		sz.NameBlock:           postProcessBlockList,
		sz.NamePara:            postProcessInlineList,
		sz.NameRegionBlock:     postProcessRegion,
		sz.NameRegionQuote:     postProcessRegion,
		sz.NameRegionVerse:     postProcessRegionVerse,
		sz.NameVerbatimComment: postProcessVerbatim,
		sz.NameVerbatimEval:    postProcessVerbatim,
		sz.NameVerbatimMath:    postProcessVerbatim,
		sz.NameVerbatimProg:    postProcessVerbatim,
		sz.NameVerbatimZettel:  postProcessVerbatim,
		sz.NameHeading:         postProcessHeading,
		sz.NameListOrdered:     postProcessItemList,
		sz.NameListUnordered:   postProcessItemList,
		sz.NameListQuote:       postProcessQuoteList,
		sz.NameDescription:     postProcessDescription,
		sz.NameTable:           postProcessTable,

		sz.NameInline:       postProcessInlineList,
		sz.NameText:         postProcessText,
		sz.NameSoft:         postProcessSoft,
		sz.NameEndnote:      postProcessEndnote,
		sz.NameMark:         postProcessMark,
		sz.NameLinkBased:    postProcessInlines4,
		sz.NameLinkBroken:   postProcessInlines4,
		sz.NameLinkExternal: postProcessInlines4,
		sz.NameLinkFound:    postProcessInlines4,
		sz.NameLinkHosted:   postProcessInlines4,
		sz.NameLinkInvalid:  postProcessInlines4,
		sz.NameLinkQuery:    postProcessInlines4,
		sz.NameLinkSelf:     postProcessInlines4,
		sz.NameLinkZettel:   postProcessInlines4,
		sz.NameEmbed:        postProcessInlines4,
		sz.NameCite:         postProcessInlines4,
		sz.NameFormatDelete: postProcessFormat,
		sz.NameFormatEmph:   postProcessFormat,
		sz.NameFormatInsert: postProcessFormat,
		sz.NameFormatMark:   postProcessFormat,
		sz.NameFormatQuote:  postProcessFormat,
		sz.NameFormatStrong: postProcessFormat,
		sz.NameFormatSpan:   postProcessFormat,
		sz.NameFormatSub:    postProcessFormat,
		sz.NameFormatSuper:  postProcessFormat,

		symSeparator.GetValue(): ignoreProcess,
	}
}

func ignoreProcess(*postProcessor, *sx.Pair, *sx.Pair) *sx.Pair { return nil }

func postProcessBlockList(pp *postProcessor, lst *sx.Pair, env *sx.Pair) *sx.Pair {
	result := pp.visitPairList(lst.Tail(), env)
	if result == nil {
		if noBlockPair := env.Assoc(symNoBlock); noBlockPair == nil || sx.IsTrue(noBlockPair.Cdr()) {
			return nil
		}
	}
	return result.Cons(lst.Car())
}

func postProcessInlineList(pp *postProcessor, lst *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := lst.Car()
	if rest := pp.visitInlines(lst.Tail(), env); rest != nil {
		return rest.Cons(sym)
	}
	return nil
}

func postProcessRegion(pp *postProcessor, rn *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := rn.Car()
	next := rn.Tail()
	attrs := next.Car()
	next = next.Tail()
	blocks := pp.visitPairList(next.Head(), env)
	text := pp.visitInlines(next.Tail(), env)
	if blocks == nil && text == nil {
		return nil
	}
	return text.Cons(blocks).Cons(attrs).Cons(sym)
}

func postProcessRegionVerse(pp *postProcessor, rn *sx.Pair, env *sx.Pair) *sx.Pair {
	return postProcessRegion(pp, rn, env.Cons(sx.Cons(symInVerse, nil)))
}

func postProcessVerbatim(pp *postProcessor, verb *sx.Pair, _ *sx.Pair) *sx.Pair {
	if content, isString := sx.GetString(verb.Tail().Tail().Car()); isString && content.GetValue() != "" {
		return verb
	}
	return nil
}

func postProcessHeading(pp *postProcessor, hn *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := hn.Car()
	next := hn.Tail()
	level := next.Car()
	next = next.Tail()
	attrs := next.Car()
	next = next.Tail()
	slug := next.Car()
	next = next.Tail()
	fragment := next.Car()
	if text := pp.visitInlines(next.Tail(), env); text != nil {
		return text.Cons(fragment).Cons(slug).Cons(attrs).Cons(level).Cons(sym)
	}
	return nil
}

func postProcessItemList(pp *postProcessor, ln *sx.Pair, env *sx.Pair) *sx.Pair {
	elems := pp.visitListElems(ln, env)
	if elems == nil {
		return nil
	}
	return elems.Cons(ln.Car())
}

func postProcessQuoteList(pp *postProcessor, ln *sx.Pair, env *sx.Pair) *sx.Pair {
	elems := pp.visitListElems(ln, env.Cons(sx.Cons(symNoBlock, nil)))

	// Collect multiple paragraph items into one item.

	var newElems sx.ListBuilder
	var newPara sx.ListBuilder

	addtoParagraph := func() {
		if !newPara.IsEmpty() {
			newElems.Add(sx.MakeList(sz.SymBlock, newPara.List().Cons(sz.SymPara)))
		}
	}
	for node := elems; node != nil; node = node.Tail() {
		item := node.Head()
		if !item.Car().IsEqual(sz.SymBlock) {
			continue
		}
		itemTail := item.Tail()
		if itemTail == nil || itemTail.Tail() != nil {
			addtoParagraph()
			newElems.Add(item)
			continue
		}
		if pn := itemTail.Head(); pn.Car().IsEqual(sz.SymPara) {
			if !newPara.IsEmpty() {
				newPara.Add(sx.Cons(sz.SymSoft, nil))
			}
			newPara.ExtendBang(pn.Tail())
			continue
		}
		addtoParagraph()
		newElems.Add(item)
	}
	addtoParagraph()
	return newElems.List().Cons(ln.Car())
}

func (pp *postProcessor) visitListElems(ln *sx.Pair, env *sx.Pair) *sx.Pair {
	var pList sx.ListBuilder
	for node := ln.Tail(); node != nil; node = node.Tail() {
		if elem := sz.Walk(pp, node.Head(), env); elem != nil {
			pList.Add(elem)
		}
	}
	return pList.List()
}

func postProcessDescription(pp *postProcessor, dl *sx.Pair, env *sx.Pair) *sx.Pair {
	var dList sx.ListBuilder
	isTerm := false
	for node := dl.Tail(); node != nil; node = node.Tail() {
		isTerm = !isTerm
		if isTerm {
			dList.Add(pp.visitInlines(node.Head(), env))
		} else {
			dList.Add(sz.Walk(pp, node.Head(), env))
		}
	}
	return dList.List().Cons(dl.Car())
}

func postProcessTable(pp *postProcessor, tbl *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := tbl.Car()
	next := tbl.Tail()
	header := next.Head()
	if header != nil {
		// Already post-processed
		return tbl
	}
	rows, width := pp.visitRows(next.Tail(), env)
	if rows == nil {
		// Header and row are nil -> no table
		return nil
	}
	header, rows, align := splitTableHeader(rows, width)
	alignRow(header, align)
	for node := rows; node != nil; node = node.Tail() {
		alignRow(node.Head(), align)
	}
	return rows.Cons(header).Cons(sym)
}

func (pp *postProcessor) visitRows(rows *sx.Pair, env *sx.Pair) (*sx.Pair, int) {
	maxWidth := 0
	var pRows sx.ListBuilder
	for node := rows; node != nil; node = node.Tail() {
		row := node.Head()
		row, width := pp.visitCells(row, env)
		if maxWidth < width {
			maxWidth = width
		}
		pRows.Add(row)
	}
	return pRows.List(), maxWidth
}

func (pp *postProcessor) visitCells(cells *sx.Pair, env *sx.Pair) (*sx.Pair, int) {
	width := 0
	var pCells sx.ListBuilder
	for node := cells; node != nil; node = node.Tail() {
		cell := node.Head()
		ins := pp.visitInlines(cell.Tail(), env)
		newCell := ins.Cons(cell.Car())
		pCells.Add(newCell)
		width++
	}
	return pCells.List(), width
}

func splitTableHeader(rows *sx.Pair, width int) (header, realRows *sx.Pair, align []*sx.Symbol) {
	align = make([]*sx.Symbol, width)

	foundHeader := false
	cellCount := 0

	// assert: rows != nil (checked in postProcessTable)
	for node := rows.Head(); node != nil; node = node.Tail() {
		cellCount++
		cell := node.Head()
		cellTail := cell.Tail()
		if cellTail == nil {
			continue
		}

		// elem is first cell inline element
		elem := cellTail.Head()
		if elem.Car().IsEqual(sz.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString && s.GetValue() != "" {
				str := s.GetValue()
				if str[0] == '=' {
					foundHeader = true
					elem.SetCdr(sx.Cons(sx.MakeString(str[1:]), nil))
				}
			}
		}

		// move to the last cell inline element
		for {
			next := cellTail.Tail()
			if next == nil {
				break
			}
			cellTail = next
		}

		elem = cellTail.Head()
		if elem.Car().IsEqual(sz.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString && s.GetValue() != "" {
				str := s.GetValue()
				cellAlign := getCellAlignment(str[len(str)-1])
				if !cellAlign.IsEqual(sz.SymCell) {
					elem.SetCdr(sx.Cons(sx.MakeString(str[0:len(str)-1]), nil))
				}
				align[cellCount-1] = cellAlign
				cell.SetCar(cellAlign)
			}
		}
	}

	if !foundHeader {
		for i := 0; i < width; i++ {
			align[i] = sz.SymCell // Default alignment
		}
		return nil, rows, align
	}

	for i := 0; i < width; i++ {
		if align[i] == nil {
			align[i] = sz.SymCell // Default alignment
		}
	}
	return rows.Head(), rows.Tail(), align
}

func alignRow(row *sx.Pair, align []*sx.Symbol) {
	if row == nil {
		return
	}
	var lastCellNode *sx.Pair
	cellCount := 0
	for node := row; node != nil; node = node.Tail() {
		lastCellNode = node
		cell := node.Head()
		cell.SetCar(align[cellCount])
		cellCount++
		cellTail := cell.Tail()
		if cellTail == nil {
			continue
		}

		// elem is first cell inline element
		elem := cellTail.Head()
		if elem.Car().IsEqual(sz.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString && s.GetValue() != "" {
				str := s.GetValue()
				cellAlign := getCellAlignment(str[0])
				if !cellAlign.IsEqual(sz.SymCell) {
					elem.SetCdr(sx.Cons(sx.MakeString(str[1:]), nil))
					cell.SetCar(cellAlign)
				}
			}
		}
	}

	for cellCount < len(align) {
		lastCellNode = lastCellNode.AppendBang(sx.Cons(align[cellCount], nil))
		cellCount++
	}
}

func getCellAlignment(ch byte) *sx.Symbol {
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

func (pp *postProcessor) visitInlines(lst *sx.Pair, env *sx.Pair) *sx.Pair {
	length := lst.Length()
	if length <= 0 {
		return nil
	}
	inVerse := env.Assoc(symInVerse) != nil
	vector := make([]*sx.Pair, 0, length)
	// 1st phase: process all childs, ignore SPACE at start, and merge some elements
	for node := lst; node != nil; node = node.Tail() {
		elem := sz.Walk(pp, node.Head(), env)
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
			lastText := last.Tail().Car().(sx.String).GetValue()
			elemText := elem.Tail().Car().(sx.String).GetValue()
			last.SetCdr(sx.Cons(sx.MakeString(lastText+elemText), sx.Nil()))
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
		if !elemSym.IsEqual(sz.SymSpace) && !sz.IsBreakSym(elemSym) {
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

func postProcessText(_ *postProcessor, txt *sx.Pair, _ *sx.Pair) *sx.Pair {
	if tail := txt.Tail(); tail != nil {
		if content, isString := sx.GetString(tail.Car()); isString && content.GetValue() != "" {
			return txt
		}
	}
	return nil
}

func postProcessSoft(_ *postProcessor, sn *sx.Pair, env *sx.Pair) *sx.Pair {
	if env.Assoc(symInVerse) == nil {
		return sn
	}
	return sx.Cons(sz.SymHard, nil)
}

func postProcessEndnote(pp *postProcessor, en *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := en.Car()
	next := en.Tail()
	attrs := next.Car()
	if text := pp.visitInlines(next.Tail(), env); text != nil {
		return text.Cons(attrs).Cons(sym)
	}
	return sx.MakeList(sym, attrs)
}

func postProcessMark(pp *postProcessor, en *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := en.Car()
	next := en.Tail()
	mark := next.Car()
	next = next.Tail()
	slug := next.Car()
	next = next.Tail()
	fragment := next.Car()
	text := pp.visitInlines(next.Tail(), env)
	return text.Cons(fragment).Cons(slug).Cons(mark).Cons(sym)
}

func postProcessInlines4(pp *postProcessor, ln *sx.Pair, env *sx.Pair) *sx.Pair {
	sym := ln.Car()
	next := ln.Tail()
	attrs := next.Car()
	next = next.Tail()
	val3 := next.Car()
	text := pp.visitInlines(next.Tail(), env)
	return text.Cons(val3).Cons(attrs).Cons(sym)
}

func postProcessFormat(pp *postProcessor, fn *sx.Pair, env *sx.Pair) *sx.Pair {
	symFormat := fn.Car()
	next := fn.Tail() // Attrs
	attrs := next.Car()
	next = next.Tail() // Possible inlines
	if next == nil {
		return fn
	}
	inlines := pp.visitInlines(next, env)
	return inlines.Cons(attrs).Cons(symFormat)
}
