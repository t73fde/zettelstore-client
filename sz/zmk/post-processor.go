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

	"t73f.de/r/sx"
	"t73f.de/r/zsx"
)

var symInVerse = sx.MakeSymbol("in-verse")
var symNoBlock = sx.MakeSymbol("no-block")

type postProcessor struct{}

func (pp *postProcessor) VisitBefore(lst *sx.Pair, alst *sx.Pair) (sx.Object, bool) {
	if lst == nil {
		return nil, true
	}
	sym, isSym := sx.GetSymbol(lst.Car())
	if !isSym {
		panic(lst)
	}
	if fn, found := symMap[sym]; found {
		return fn(pp, lst, alst), true
	}
	return nil, false
}

func (pp *postProcessor) VisitAfter(lst *sx.Pair, _ *sx.Pair) sx.Object { return lst }

func (pp *postProcessor) visitPairList(lst *sx.Pair, alst *sx.Pair) *sx.Pair {
	var pList sx.ListBuilder
	for node := range lst.Pairs() {
		if elem, isPair := sx.GetPair(zsx.WalkBang(pp, node.Head(), alst)); isPair && elem != nil {
			pList.Add(elem)
		}
	}
	return pList.List()
}

var symMap map[*sx.Symbol]func(*postProcessor, *sx.Pair, *sx.Pair) *sx.Pair

func init() {
	symMap = map[*sx.Symbol]func(*postProcessor, *sx.Pair, *sx.Pair) *sx.Pair{
		zsx.SymBlock:           postProcessBlockList,
		zsx.SymPara:            postProcessInlineList,
		zsx.SymRegionBlock:     postProcessRegion,
		zsx.SymRegionQuote:     postProcessRegion,
		zsx.SymRegionVerse:     postProcessRegionVerse,
		zsx.SymVerbatimComment: postProcessVerbatim,
		zsx.SymVerbatimEval:    postProcessVerbatim,
		zsx.SymVerbatimMath:    postProcessVerbatim,
		zsx.SymVerbatimCode:    postProcessVerbatim,
		zsx.SymVerbatimZettel:  postProcessVerbatim,
		zsx.SymHeading:         postProcessHeading,
		zsx.SymListOrdered:     postProcessItemList,
		zsx.SymListUnordered:   postProcessItemList,
		zsx.SymListQuote:       postProcessQuoteList,
		zsx.SymDescription:     postProcessDescription,
		zsx.SymTable:           postProcessTable,

		zsx.SymInline:       postProcessInlineList,
		zsx.SymText:         postProcessText,
		zsx.SymSoft:         postProcessSoft,
		zsx.SymEndnote:      postProcessEndnote,
		zsx.SymMark:         postProcessMark,
		zsx.SymLink:         postProcessInlines4,
		zsx.SymEmbed:        postProcessEmbed,
		zsx.SymCite:         postProcessInlines4,
		zsx.SymFormatDelete: postProcessFormat,
		zsx.SymFormatEmph:   postProcessFormat,
		zsx.SymFormatInsert: postProcessFormat,
		zsx.SymFormatMark:   postProcessFormat,
		zsx.SymFormatQuote:  postProcessFormat,
		zsx.SymFormatStrong: postProcessFormat,
		zsx.SymFormatSpan:   postProcessFormat,
		zsx.SymFormatSub:    postProcessFormat,
		zsx.SymFormatSuper:  postProcessFormat,

		symSeparator: ignoreProcess,
	}
}

func ignoreProcess(*postProcessor, *sx.Pair, *sx.Pair) *sx.Pair { return nil }

func postProcessBlockList(pp *postProcessor, lst *sx.Pair, alst *sx.Pair) *sx.Pair {
	result := pp.visitPairList(lst.Tail(), alst)
	if result == nil {
		if noBlockPair := alst.Assoc(symNoBlock); noBlockPair == nil || sx.IsTrue(noBlockPair.Cdr()) {
			return nil
		}
	}
	return result.Cons(lst.Car())
}

func postProcessInlineList(pp *postProcessor, lst *sx.Pair, alst *sx.Pair) *sx.Pair {
	sym := lst.Car()
	if rest := pp.visitInlines(lst.Tail(), alst); rest != nil {
		return rest.Cons(sym)
	}
	return nil
}

func postProcessRegion(pp *postProcessor, rn *sx.Pair, alst *sx.Pair) *sx.Pair {
	return doPostProcessRegion(pp, rn, alst, alst)
}

func postProcessRegionVerse(pp *postProcessor, rn *sx.Pair, alst *sx.Pair) *sx.Pair {
	return doPostProcessRegion(pp, rn, alst.Cons(sx.Cons(symInVerse, nil)), alst)
}

func doPostProcessRegion(pp *postProcessor, rn *sx.Pair, alstBlock, alstInline *sx.Pair) *sx.Pair {
	sym := rn.Car().(*sx.Symbol)
	next := rn.Tail()
	attrs := next.Car().(*sx.Pair)
	next = next.Tail()
	blocks := pp.visitPairList(next.Head(), alstBlock)
	text := pp.visitInlines(next.Tail(), alstInline)
	if blocks == nil && text == nil {
		return nil
	}
	return zsx.MakeRegion(sym, attrs, blocks, text)
}

func postProcessVerbatim(_ *postProcessor, verb *sx.Pair, _ *sx.Pair) *sx.Pair {
	if content, isString := sx.GetString(verb.Tail().Tail().Car()); isString && content.GetValue() != "" {
		return verb
	}
	return nil
}

func postProcessHeading(pp *postProcessor, hn *sx.Pair, alst *sx.Pair) *sx.Pair {
	next := hn.Tail()
	level := next.Car().(sx.Int64)
	next = next.Tail()
	attrs := next.Car().(*sx.Pair)
	next = next.Tail()
	slug := next.Car().(sx.String)
	next = next.Tail()
	fragment := next.Car().(sx.String)
	if text := pp.visitInlines(next.Tail(), alst); text != nil {
		return zsx.MakeHeading(int(level), attrs, text, slug.GetValue(), fragment.GetValue())
	}
	return nil
}

func postProcessItemList(pp *postProcessor, ln *sx.Pair, alst *sx.Pair) *sx.Pair {
	attrs := ln.Tail().Head()
	elems := pp.visitListElems(ln.Tail(), alst)
	if elems == nil {
		return nil
	}
	return zsx.MakeList(ln.Car().(*sx.Symbol), attrs, elems)
}

func postProcessQuoteList(pp *postProcessor, ln *sx.Pair, alst *sx.Pair) *sx.Pair {
	attrs := ln.Tail().Head()
	elems := pp.visitListElems(ln.Tail(), alst.Cons(sx.Cons(symNoBlock, nil)))

	// Collect multiple paragraph items into one item.

	var newElems sx.ListBuilder
	var newPara sx.ListBuilder

	addtoParagraph := func() {
		if !newPara.IsEmpty() {
			newElems.Add(sx.MakeList(zsx.SymBlock, newPara.List().Cons(zsx.SymPara)))
			newPara.Reset()
		}
	}
	for node := range elems.Pairs() {
		item := node.Head()
		if !item.Car().IsEqual(zsx.SymBlock) {
			continue
		}
		itemTail := item.Tail()
		if itemTail == nil || itemTail.Tail() != nil {
			addtoParagraph()
			newElems.Add(item)
			continue
		}
		if pn := itemTail.Head(); pn.Car().IsEqual(zsx.SymPara) {
			if !newPara.IsEmpty() {
				newPara.Add(sx.Cons(zsx.SymSoft, nil))
			}
			newPara.ExtendBang(pn.Tail())
			continue
		}
		addtoParagraph()
		newElems.Add(item)
	}
	addtoParagraph()
	return zsx.MakeList(ln.Car().(*sx.Symbol), attrs, newElems.List())
}

func (pp *postProcessor) visitListElems(ln *sx.Pair, alst *sx.Pair) *sx.Pair {
	var pList sx.ListBuilder
	for node := range ln.Tail().Pairs() {
		if elem := zsx.WalkBang(pp, node.Head(), alst); elem != nil {
			pList.Add(elem)
		}
	}
	return pList.List()
}

func postProcessDescription(pp *postProcessor, dl *sx.Pair, alst *sx.Pair) *sx.Pair {
	attrs := dl.Tail().Head()
	var dList sx.ListBuilder
	isTerm := false
	for node := range dl.Tail().Tail().Pairs() {
		isTerm = !isTerm
		if isTerm {
			dList.Add(pp.visitInlines(node.Head(), alst))
		} else {
			dList.Add(zsx.WalkBang(pp, node.Head(), alst.Cons(sx.Cons(symNoBlock, nil))))
		}
	}
	return dList.List().Cons(attrs).Cons(dl.Car())
}

func postProcessTable(pp *postProcessor, tbl *sx.Pair, alst *sx.Pair) *sx.Pair {
	sym, next := tbl.Car(), tbl.Tail()
	attrs := next.Head()
	next = next.Tail()
	header := next.Head()
	if header != nil {
		// Already post-processed
		return tbl
	}
	rows, width := pp.visitRows(next.Tail(), alst)
	if rows == nil {
		// Header and row are nil -> no table
		return nil
	}
	header, rows, align := splitTableHeader(rows, width)
	alignRow(header, align)
	for node := range rows.Pairs() {
		alignRow(node.Head(), align)
	}
	return rows.Cons(header).Cons(attrs).Cons(sym)
}

func (pp *postProcessor) visitRows(rows *sx.Pair, alst *sx.Pair) (*sx.Pair, int) {
	maxWidth := 0
	var pRows sx.ListBuilder
	for node := range rows.Pairs() {
		row := node.Head()
		row, width := pp.visitCells(row, alst)
		if maxWidth < width {
			maxWidth = width
		}
		pRows.Add(row)
	}
	return pRows.List(), maxWidth
}

func (pp *postProcessor) visitCells(cells *sx.Pair, alst *sx.Pair) (*sx.Pair, int) {
	width := 0
	var pCells sx.ListBuilder
	for node := range cells.Pairs() {
		cell := node.Head()
		rest := cell.Tail()
		attrs := rest.Head()
		ins := pp.visitInlines(rest.Tail(), alst)
		pCells.Add(zsx.MakeCell(attrs, ins))
		width++
	}
	return pCells.List(), width
}

func splitTableHeader(rows *sx.Pair, width int) (header, realRows *sx.Pair, align []byte) {
	align = make([]byte, width)

	foundHeader := false
	cellCount := 0

	// assert: rows != nil (checked in postProcessTable)
	for node := range rows.Head().Pairs() {
		cell := node.Head()
		cellCount++
		rest := cell.Tail() // attrs := rest.Head()
		cellInlines := rest.Tail()
		if cellInlines == nil {
			continue
		}

		// elem is first cell inline element
		elem := cellInlines.Head()
		if elem.Car().IsEqual(zsx.SymText) {
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
			next := cellInlines.Tail()
			if next == nil {
				break
			}
			cellInlines = next
		}

		elem = cellInlines.Head()
		if elem.Car().IsEqual(zsx.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString && s.GetValue() != "" {
				str := s.GetValue()
				lastByte := str[len(str)-1]
				if cellAlign, isValid := getCellAlignment(lastByte); isValid {
					elem.SetCdr(sx.Cons(sx.MakeString(str[0:len(str)-1]), nil))
					rest.SetCar(makeCellAttrs(cellAlign))
				}
				align[cellCount-1] = lastByte
			}
		}
	}

	if !foundHeader {
		return nil, rows, align
	}

	return rows.Head(), rows.Tail(), align
}

func alignRow(row *sx.Pair, defaultAlign []byte) {
	if row == nil {
		return
	}
	var lastCellNode *sx.Pair
	cellColumnNo := 0
	for node := range row.Pairs() {
		lastCellNode = node
		cell := node.Head()
		cellColumnNo++
		rest := cell.Tail() // attrs := rest.Head()
		if cellAlign, isValid := getCellAlignment(defaultAlign[cellColumnNo-1]); isValid {
			rest.SetCar(makeCellAttrs(cellAlign))
		}
		cellInlines := rest.Tail()
		if cellInlines == nil {
			continue
		}

		// elem is first cell inline element
		elem := cellInlines.Head()
		if elem.Car().IsEqual(zsx.SymText) {
			if s, isString := sx.GetString(elem.Tail().Car()); isString && s.GetValue() != "" {
				str := s.GetValue()
				cellAlign, isValid := getCellAlignment(str[0])
				if isValid {
					elem.SetCdr(sx.Cons(sx.MakeString(str[1:]), nil))
					rest.SetCar(makeCellAttrs(cellAlign))
				}
			}
		}
	}

	for cellColumnNo < len(defaultAlign) {
		var attrs *sx.Pair
		if cellAlign, isValid := getCellAlignment(defaultAlign[cellColumnNo]); isValid {
			attrs = makeCellAttrs(cellAlign)
		}
		lastCellNode = lastCellNode.AppendBang(zsx.MakeCell(attrs, nil))
		cellColumnNo++
	}
}

func makeCellAttrs(align sx.String) *sx.Pair {
	return sx.Cons(sx.Cons(zsx.SymAttrAlign, align), sx.Nil())
}

func getCellAlignment(ch byte) (sx.String, bool) {
	switch ch {
	case ':':
		return zsx.AttrAlignCenter, true
	case '<':
		return zsx.AttrAlignLeft, true
	case '>':
		return zsx.AttrAlignRight, true
	default:
		return sx.MakeString(""), false
	}
}

func (pp *postProcessor) visitInlines(lst *sx.Pair, alst *sx.Pair) *sx.Pair {
	length := lst.Length()
	if length <= 0 {
		return nil
	}
	inVerse := alst.Assoc(symInVerse) != nil
	vector := make([]*sx.Pair, 0, length)
	// 1st phase: process all childs, ignore ' ' / '\t' at start, and merge some elements
	for node := range lst.Pairs() {
		elem, isPair := sx.GetPair(zsx.WalkBang(pp, node.Head(), alst))
		if !isPair || elem == nil {
			continue
		}
		elemSym := elem.Car()
		elemTail := elem.Tail()

		if inVerse && elemSym.IsEqual(zsx.SymText) {
			if s, isString := sx.GetString(elemTail.Car()); isString {
				verseText := s.GetValue()
				verseText = strings.ReplaceAll(verseText, " ", "\u00a0")
				elemTail.SetCar(sx.MakeString(verseText))
			}
		}

		if len(vector) == 0 {
			// If the 1st element is a TEXT, remove all ' ', '\t' at the beginning, if outside a verse block.
			if !elemSym.IsEqual(zsx.SymText) {
				vector = append(vector, elem)
				continue
			}

			elemText := elemTail.Car().(sx.String).GetValue()
			if elemText != "" && (elemText[0] == ' ' || elemText[0] == '\t') {
				for elemText != "" {
					if ch := elemText[0]; ch != ' ' && ch != '\t' {
						break
					}
					elemText = elemText[1:]
				}
				elemTail.SetCar(sx.MakeString(elemText))
			}
			if elemText != "" {
				vector = append(vector, elem)
			}
			continue
		}
		last := vector[len(vector)-1]
		lastSym := last.Car()

		if lastSym.IsEqual(zsx.SymText) && elemSym.IsEqual(zsx.SymText) {
			// Merge two TEXT elements into one
			lastText := last.Tail().Car().(sx.String).GetValue()
			elemText := elem.Tail().Car().(sx.String).GetValue()
			last.SetCdr(sx.Cons(sx.MakeString(lastText+elemText), sx.Nil()))
			continue
		}

		if lastSym.IsEqual(zsx.SymText) && elemSym.IsEqual(zsx.SymSoft) {
			// Merge (TEXT "... ") (SOFT) to (TEXT "...") (HARD)
			lastTail := last.Tail()
			if lastText := lastTail.Car().(sx.String).GetValue(); strings.HasSuffix(lastText, " ") {
				newText := removeTrailingSpaces(lastText)
				if newText == "" {
					vector[len(vector)-1] = sx.Cons(zsx.SymHard, sx.Nil())
					continue
				}
				lastTail.SetCar(sx.MakeString(newText))
				elemSym = zsx.SymHard
				elem.SetCar(elemSym)
			}
		}

		vector = append(vector, elem)
	}
	if len(vector) == 0 {
		return nil
	}

	// 2nd phase: remove (SOFT), (HARD) at the end, remove trailing spaces in (TEXT "...")
	lastPos := len(vector) - 1
	for lastPos >= 0 {
		elem := vector[lastPos]
		elemSym := elem.Car()
		if elemSym.IsEqual(zsx.SymText) {
			elemTail := elem.Tail()
			elemText := elemTail.Car().(sx.String).GetValue()
			newText := removeTrailingSpaces(elemText)
			if newText != "" {
				elemTail.SetCar(sx.MakeString(newText))
				break
			}
			lastPos--
		} else if isBreakSym(elemSym) {
			lastPos--
		} else {
			break
		}
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

func removeTrailingSpaces(s string) string {
	for len(s) > 0 {
		if ch := s[len(s)-1]; ch != ' ' && ch != '\t' {
			return s
		}
		s = s[0 : len(s)-1]
	}
	return ""
}

func postProcessText(_ *postProcessor, txt *sx.Pair, _ *sx.Pair) *sx.Pair {
	if tail := txt.Tail(); tail != nil {
		if content, isString := sx.GetString(tail.Car()); isString && content.GetValue() != "" {
			return txt
		}
	}
	return nil
}

func postProcessSoft(_ *postProcessor, sn *sx.Pair, alst *sx.Pair) *sx.Pair {
	if alst.Assoc(symInVerse) == nil {
		return sn
	}
	return sx.Cons(zsx.SymHard, nil)
}

func postProcessEndnote(pp *postProcessor, en *sx.Pair, alst *sx.Pair) *sx.Pair {
	next := en.Tail()
	attrs := next.Car().(*sx.Pair)
	if text := pp.visitInlines(next.Tail(), alst); text != nil {
		return zsx.MakeEndnote(attrs, text)
	}
	return zsx.MakeEndnote(attrs, sx.Nil())
}

func postProcessMark(pp *postProcessor, en *sx.Pair, alst *sx.Pair) *sx.Pair {
	next := en.Tail()
	mark := next.Car().(sx.String)
	next = next.Tail()
	slug := next.Car().(sx.String)
	next = next.Tail()
	fragment := next.Car().(sx.String)
	text := pp.visitInlines(next.Tail(), alst)
	return zsx.MakeMark(mark.GetValue(), slug.GetValue(), fragment.GetValue(), text)
}

func postProcessInlines4(pp *postProcessor, ln *sx.Pair, alst *sx.Pair) *sx.Pair {
	sym := ln.Car()
	next := ln.Tail()
	attrs := next.Car()
	next = next.Tail()
	val3 := next.Car()
	text := pp.visitInlines(next.Tail(), alst)
	return text.Cons(val3).Cons(attrs).Cons(sym)
}

func postProcessEmbed(pp *postProcessor, ln *sx.Pair, alst *sx.Pair) *sx.Pair {
	next := ln.Tail()
	attrs := next.Car().(*sx.Pair)
	next = next.Tail()
	ref := next.Car()
	next = next.Tail()
	syntax := next.Car().(sx.String)
	text := pp.visitInlines(next.Tail(), alst)
	return zsx.MakeEmbed(attrs, ref, syntax.GetValue(), text)
}

func postProcessFormat(pp *postProcessor, fn *sx.Pair, alst *sx.Pair) *sx.Pair {
	symFormat := fn.Car().(*sx.Symbol)
	next := fn.Tail() // Attrs
	attrs := next.Car().(*sx.Pair)
	next = next.Tail() // Possible inlines
	if next == nil {
		return fn
	}
	inlines := pp.visitInlines(next, alst)
	return zsx.MakeFormat(symFormat, attrs, inlines)
}
