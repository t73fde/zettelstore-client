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
	"fmt"

	"zettelstore.de/client.fossil/input"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

// parseBlock parses one block.
func (cp *zmkP) parseBlock(lastPara *sx.Pair) (res *sx.Pair, cont bool) {
	inp := cp.inp
	pos := inp.Pos
	if cp.nestingLevel <= maxNestingLevel {
		cp.nestingLevel++
		defer func() { cp.nestingLevel-- }()

		var bn *sx.Pair
		success := false

		switch inp.Ch {
		case input.EOS:
			return nil, false
		case '\n', '\r':
			inp.EatEOL()
			cp.cleanupListsAfterEOL()
			return nil, false
		case ':':
			bn, success = cp.parseColon()
		case '@', '`', runeModGrave, '%', '~', '$':
			cp.clearStacked()
			bn, success = cp.parseVerbatim()
		case '"', '<':
			cp.clearStacked()
			bn, success = cp.parseRegion()
		case '=':
			cp.clearStacked()
			bn, success = cp.parseHeading()
		case '-':
			cp.clearStacked()
			bn, success = cp.parseHRule()
		case '*', '#', '>':
			cp.lastRow = nil
			// cp.descrl = nil
			bn, success = cp.parseNestedList()
		case ';':
			cp.lists = nil
			cp.lastRow = nil
			bn, success = cp.parseDefTerm()
		case ' ':
			cp.lastRow = nil
			bn, success = nil, cp.parseIndent()
		case '|':
			cp.lists = nil
			// cp.descrl = nil
			bn, success = cp.parseRow(), true
		case '{':
			cp.clearStacked()
			bn, success = cp.parseTransclusion()
		}

		if success {
			return bn, false
		}
	}
	inp.SetPos(pos)
	cp.clearStacked()
	ins := cp.parsePara()
	if startsWithSpaceSoftBreak(ins) {
		ins = ins[2:]
	} else if lastPara != nil {
		lastPair := lastPara.LastPair()
		lastPair.ExtendBang(sx.MakeList(ins...))
		return nil, true
	}
	return sx.MakeList(ins...).Cons(sz.SymPara), false
}

func startsWithSpaceSoftBreak(ins sx.Vector) bool {
	if len(ins) < 2 {
		return false
	}
	pair0, isPair0 := sx.GetPair(ins[0])
	pair1, isPair1 := sx.GetPair(ins[0])
	if !isPair0 || !isPair1 {
		return false
	}
	car1 := pair1.Car()
	return pair0.Car().IsEqual(sz.SymSpace) && (car1.IsEqual(sz.SymSoft) || car1.IsEqual(sz.SymHard))
}

func (cp *zmkP) cleanupListsAfterEOL() {
	// for _, l := range cp.lists {
	// 	if lits := len(l.Items); lits > 0 {
	// 		l.Items[lits-1] = append(l.Items[lits-1], &nullItemNode{})
	// 	}
	// }
	// if cp.descrl != nil {
	// 	defPos := len(cp.descrl.Descriptions) - 1
	// 	if ldds := len(cp.descrl.Descriptions[defPos].Descriptions); ldds > 0 {
	// 		cp.descrl.Descriptions[defPos].Descriptions[ldds-1] = append(
	// 			cp.descrl.Descriptions[defPos].Descriptions[ldds-1], &nullDescriptionNode{})
	// 	}
	// }
}

// parseColon determines which element should be parsed.
func (cp *zmkP) parseColon() (*sx.Pair, bool) {
	inp := cp.inp
	if inp.PeekN(1) == ':' {
		cp.clearStacked()
		return cp.parseRegion()
	}
	return cp.parseDefDescr()
}

// parsePara parses paragraphed inline material as a slice of sx.Object.
func (cp *zmkP) parsePara() (result sx.Vector) {
	for {
		in := cp.parseInline()
		if in == nil {
			return result
		}
		result = append(result, in)
		if sym := in.Car(); sym.IsEqual(sz.SymSoft) || sym.IsEqual(sz.SymHard) {
			ch := cp.inp.Ch
			switch ch {
			// Must contain all cases from above switch in parseBlock.
			case input.EOS, '\n', '\r', '@', '`', runeModGrave, '%', '~', '$', '"', '<', '=', '-', '*', '#', '>', ';', ':', ' ', '|', '{':
				return result
			}
		}
	}
}

// countDelim read from input until a non-delimiter is found and returns number of delimiter chars.
func (cp *zmkP) countDelim(delim rune) int {
	inp := cp.inp
	cnt := 0
	for inp.Ch == delim {
		cnt++
		inp.Next()
	}
	return cnt
}

// parseVerbatim parses a verbatim block.
func (cp *zmkP) parseVerbatim() (rn *sx.Pair, success bool) {
	inp := cp.inp
	fch := inp.Ch
	cnt := cp.countDelim(fch)
	if cnt < 3 {
		return nil, false
	}
	attrs := cp.parseBlockAttributes()
	inp.SkipToEOL()
	if inp.Ch == input.EOS {
		return nil, false
	}
	var sym sx.Symbol
	switch fch {
	case '@':
		sym = sz.SymVerbatimZettel
	case '`', runeModGrave:
		sym = sz.SymVerbatimProg
	case '%':
		sym = sz.SymVerbatimComment
	case '~':
		sym = sz.SymVerbatimEval
	case '$':
		sym = sz.SymVerbatimMath
	default:
		panic(fmt.Sprintf("%q is not a verbatim char", fch))
	}
	content := make([]byte, 0, 512)
	for {
		inp.EatEOL()
		posL := inp.Pos
		switch inp.Ch {
		case fch:
			if cp.countDelim(fch) >= cnt {
				inp.SkipToEOL()
				rn = sx.MakeList(sym, attrs, sx.String(content))
				return rn, true
			}
			inp.SetPos(posL)
		case input.EOS:
			return nil, false
		}
		inp.SkipToEOL()
		if len(content) > 0 {
			content = append(content, '\n')
		}
		content = append(content, inp.Src[posL:inp.Pos]...)
	}
}

// parseRegion parses a block region.
func (cp *zmkP) parseRegion() (rn *sx.Pair, success bool) {
	inp := cp.inp
	fch := inp.Ch
	cnt := cp.countDelim(fch)
	if cnt < 3 {
		return nil, false
	}

	var sym sx.Symbol
	oldInVerse := cp.inVerse
	defer func() { cp.inVerse = oldInVerse }()
	switch fch {
	case ':':
		sym = sz.SymRegionBlock
	case '<':
		sym = sz.SymRegionQuote
	case '"':
		sym = sz.SymRegionVerse
		cp.inVerse = true
	default:
		panic(fmt.Sprintf("%q is not a region char", fch))
	}
	attrs := cp.parseBlockAttributes()
	inp.SkipToEOL()
	if inp.Ch == input.EOS {
		return nil, false
	}
	var blocksBuilder pairBuilder
	var lastPara *sx.Pair
	inp.EatEOL()
	for {
		posL := inp.Pos
		switch inp.Ch {
		case fch:
			if cp.countDelim(fch) >= cnt {
				ins := cp.parseRegionLastLine()
				rn = ins.Cons(blocksBuilder.result).Cons(attrs).Cons(sym)
				return rn, true
			}
			inp.SetPos(posL)
		case input.EOS:
			return nil, false
		}
		bn, cont := cp.parseBlock(lastPara)
		if bn != nil {
			blocksBuilder.appendBang(bn)
		}
		if !cont {
			lastPara = bn
		}
	}
}

// parseRegionLastLine parses the last line of a region and returns its inline text.
func (cp *zmkP) parseRegionLastLine() *sx.Pair {
	cp.clearStacked() // remove any lists defined in the region
	cp.skipSpace()
	var region pairBuilder
	for {
		switch cp.inp.Ch {
		case input.EOS, '\n', '\r':
			return region.result
		}
		in := cp.parseInline()
		if in == nil {
			return region.result
		}
		region.appendBang(in)
	}
}

// parseHeading parses a head line.
func (cp *zmkP) parseHeading() (hn *sx.Pair, success bool) {
	inp := cp.inp
	delims := cp.countDelim(inp.Ch)
	if delims < 3 {
		return nil, false
	}
	if inp.Ch != ' ' {
		return nil, false
	}
	inp.Next()
	cp.skipSpace()
	if delims > 7 {
		delims = 7
	}
	level := int64(delims - 2)
	var attrs *sx.Pair
	var text pairBuilder
	for {
		if input.IsEOLEOS(inp.Ch) {
			return createHeading(level, attrs, text.result), true
		}
		in := cp.parseInline()
		if in == nil {
			return createHeading(level, attrs, text.result), true
		}
		text.appendBang(in)
		if inp.Ch == '{' && inp.Peek() != '{' {
			attrs = cp.parseBlockAttributes()
			inp.SkipToEOL()
			return createHeading(level, attrs, text.result), true
		}
	}
}
func createHeading(level int64, attrs, text *sx.Pair) *sx.Pair {
	return text.
		Cons(sx.String("")). // Fragment
		Cons(sx.String("")). // Slug
		Cons(attrs).
		Cons(sx.Int64(level)).
		Cons(sz.SymHeading)
}

// parseHRule parses a horizontal rule.
func (cp *zmkP) parseHRule() (hn *sx.Pair, success bool) {
	inp := cp.inp

	if cp.countDelim(inp.Ch) < 3 {
		return nil, false
	}

	attrs := cp.parseBlockAttributes()
	inp.SkipToEOL()
	return sx.MakeList(sz.SymThematic, attrs), true
}

// parseNestedList parses a list.
func (cp *zmkP) parseNestedList() (res *sx.Pair /*ast.BlockNode*/, success bool) {
	inp := cp.inp
	kinds := cp.parseNestedListKinds()
	if kinds == nil {
		return nil, false
	}
	cp.skipSpace()
	if kinds[len(kinds)-1] != sz.SymListQuote && input.IsEOLEOS(inp.Ch) {
		return nil, false
	}

	if len(kinds) < len(cp.lists) {
		cp.lists = cp.lists[:len(kinds)]
	}
	ln, newLnCount := cp.buildNestedList(kinds)
	pn := cp.parseLinePara()
	bn := sx.Cons(sz.SymBlock, nil)
	if pn != nil {
		bn.AppendBang(pn.Cons(sz.SymPara))
	}
	lastItemPair := ln.LastPair()
	lastItemPair.AppendBang(bn)
	return cp.cleanupParsedNestedList(newLnCount)
}

func (cp *zmkP) parseNestedListKinds() (result []sx.Symbol) {
	inp := cp.inp
	for {
		var sym sx.Symbol
		switch inp.Ch {
		case '*':
			sym = sz.SymListUnordered
		case '#':
			sym = sz.SymListOrdered
		case '>':
			sym = sz.SymListQuote
		default:
			panic(fmt.Sprintf("%q is not a region char", inp.Ch))
		}
		result = append(result, sym)
		switch inp.Next() {
		case '*', '#', '>':
		case ' ', input.EOS, '\n', '\r':
			return result
		default:
			return nil
		}
	}
}

func (cp *zmkP) buildNestedList(kinds []sx.Symbol) (ln *sx.Pair, newLnCount int) {
	for i, kind := range kinds {
		if i < len(cp.lists) {
			if !cp.lists[i].Car().IsEqual(kind) {
				ln = sx.Cons(kind, nil)
				newLnCount++
				cp.lists[i] = ln
				cp.lists = cp.lists[:i+1]
			} else {
				ln = cp.lists[i]
			}
		} else {
			ln = sx.Cons(kind, nil)
			newLnCount++
			cp.lists = append(cp.lists, ln)
		}
	}
	return ln, newLnCount
}

func (cp *zmkP) cleanupParsedNestedList(newLnCount int) (res *sx.Pair /*ast.BlockNode*/, success bool) {
	listDepth := len(cp.lists)
	for i := 0; i < newLnCount; i++ {
		childPos := listDepth - i - 1
		parentPos := childPos - 1
		if parentPos < 0 {
			return cp.lists[0], true
		}
		// 	if prevItems := cp.lists[parentPos].Items; len(prevItems) > 0 {
		// 		lastItem := len(prevItems) - 1
		// 		prevItems[lastItem] = append(prevItems[lastItem], cp.lists[childPos])
		// 	} else {
		// 		cp.lists[parentPos].Items = []ast.ItemSlice{{cp.lists[childPos]}}
		// 	}
	}
	return nil, true
}

// parseDefTerm parses a term of a definition list.
func (cp *zmkP) parseDefTerm() (res *sx.Pair /*ast.BlockNode*/, success bool) {
	// inp := cp.inp
	// inp.Next()
	// if inp.Ch != ' ' {
	// 	return nil, false
	// }
	// inp.Next()
	// cp.skipSpace()
	// descrl := cp.descrl
	// if descrl == nil {
	// 	descrl = &ast.DescriptionListNode{}
	// 	cp.descrl = descrl
	// }
	// descrl.Descriptions = append(descrl.Descriptions, ast.Description{})
	// defPos := len(descrl.Descriptions) - 1
	// if defPos == 0 {
	// 	res = descrl
	// }
	// for {
	// 	in := cp.parseInline()
	// 	if in == nil {
	// 		if len(descrl.Descriptions[defPos].Term) == 0 {
	// 			return nil, false
	// 		}
	// 		return res, true
	// 	}
	// 	descrl.Descriptions[defPos].Term = append(descrl.Descriptions[defPos].Term, in)
	// 	if _, ok := in.(*ast.BreakNode); ok {
	// 		return res, true
	// 	}
	// }
	return nil, false
}

// parseDefDescr parses a description of a definition list.
func (cp *zmkP) parseDefDescr() (res *sx.Pair /*ast.BlockNode*/, success bool) {
	// inp := cp.inp
	// inp.Next()
	//
	//	if inp.Ch != ' ' {
	//		return nil, false
	//	}
	//
	// inp.Next()
	// cp.skipSpace()
	// descrl := cp.descrl
	//
	//	if descrl == nil || len(descrl.Descriptions) == 0 {
	//		return nil, false
	//	}
	//
	// defPos := len(descrl.Descriptions) - 1
	//
	//	if len(descrl.Descriptions[defPos].Term) == 0 {
	//		return nil, false
	//	}
	//
	// pn := cp.parseLinePara()
	//
	//	if pn == nil {
	//		return nil, false
	//	}
	//
	// cp.lists = nil
	// cp.lastRow = nil
	// descrl.Descriptions[defPos].Descriptions = append(descrl.Descriptions[defPos].Descriptions, ast.DescriptionSlice{pn})
	// return nil, true
	return nil, false
}

// parseIndent parses initial spaces to continue a list.
func (cp *zmkP) parseIndent() bool {
	// inp := cp.inp
	// cnt := 0
	// for {
	// 	inp.Next()
	// 	if inp.Ch != ' ' {
	// 		break
	// 	}
	// 	cnt++
	// }
	// if cp.lists != nil {
	// 	return cp.parseIndentForList(cnt)
	// }
	// if cp.descrl != nil {
	// 	return cp.parseIndentForDescription(cnt)
	// }
	return false
}

func (cp *zmkP) parseIndentForList(cnt int) bool {
	// if len(cp.lists) < cnt {
	// 	cnt = len(cp.lists)
	// }
	// cp.lists = cp.lists[:cnt]
	// if cnt == 0 {
	// 	return false
	// }
	// ln := cp.lists[cnt-1]
	// pn := cp.parseLinePara()
	// if pn == nil {
	// 	pn = &ast.ParaNode{}
	// }
	// lbn := ln.Items[len(ln.Items)-1]
	// if lpn, ok := lbn[len(lbn)-1].(*ast.ParaNode); ok {
	// 	lpn.Inlines = append(lpn.Inlines, pn.Inlines...)
	// } else {
	// 	ln.Items[len(ln.Items)-1] = append(ln.Items[len(ln.Items)-1], pn)
	// }
	// return true
	return false
}

func (cp *zmkP) parseIndentForDescription(cnt int) bool {
	// defPos := len(cp.descrl.Descriptions) - 1
	// if cnt < 1 || defPos < 0 {
	// 	return false
	// }
	// if len(cp.descrl.Descriptions[defPos].Descriptions) == 0 {
	// 	// Continuation of a definition term
	// 	for {
	// 		in := cp.parseInline()
	// 		if in == nil {
	// 			return true
	// 		}
	// 		cp.descrl.Descriptions[defPos].Term = append(cp.descrl.Descriptions[defPos].Term, in)
	// 		if _, ok := in.(*ast.BreakNode); ok {
	// 			return true
	// 		}
	// 	}
	// }

	// // Continuation of a definition description
	// pn := cp.parseLinePara()
	// if pn == nil {
	// 	return false
	// }
	// descrPos := len(cp.descrl.Descriptions[defPos].Descriptions) - 1
	// lbn := cp.descrl.Descriptions[defPos].Descriptions[descrPos]
	// if lpn, ok := lbn[len(lbn)-1].(*ast.ParaNode); ok {
	// 	lpn.Inlines = append(lpn.Inlines, pn.Inlines...)
	// } else {
	// 	descrPos = len(cp.descrl.Descriptions[defPos].Descriptions) - 1
	// 	cp.descrl.Descriptions[defPos].Descriptions[descrPos] = append(cp.descrl.Descriptions[defPos].Descriptions[descrPos], pn)
	// }
	// return true
	return false
}

// parseLinePara parses one line of inline material.
func (cp *zmkP) parseLinePara() *sx.Pair /**ast.ParaNode*/ {
	ins := sx.Vector{}
	for {
		in := cp.parseInline()
		if in == nil {
			if len(ins) == 0 {
				return nil
			}
			return sx.MakeList(ins...)
		}
		ins = append(ins, in)
		if sym := in.Car(); sym.IsEqual(sz.SymSoft) || sym.IsEqual(sz.SymHard) {
			return sx.MakeList(ins...)
		}
	}
}

// parseRow parse one table row.
func (cp *zmkP) parseRow() *sx.Pair {
	inp := cp.inp
	if inp.Peek() == '%' {
		inp.SkipToEOL()
		return nil
	}
	//var row, curr *sx.Pair
	var row pairBuilder
	for {
		inp.Next()
		cell := cp.parseCell()
		if cell != nil {
			row.appendBang(cell)
		}
		switch inp.Ch {
		case '\n', '\r':
			inp.EatEOL()
			fallthrough
		case input.EOS:
			// add to table
			if cp.lastRow == nil {
				if row.result == nil {
					return nil
				}
				cp.lastRow = sx.Cons(row.result, nil)
				return cp.lastRow.Cons(nil).Cons(sz.SymTable)
			}
			cp.lastRow = cp.lastRow.AppendBang(row.result)
			return nil
		}
		// inp.Ch must be '|'
	}
}

// parseCell parses one single cell of a table row.
func (cp *zmkP) parseCell() *sx.Pair {
	inp := cp.inp
	var cell pairBuilder
	for {
		if input.IsEOLEOS(inp.Ch) {
			if cell.result == nil {
				return nil
			}
			return cell.result.Cons(sz.SymCell)
		}
		if inp.Ch == '|' {
			return cell.result.Cons(sz.SymCell)
		}

		in := cp.parseInline()
		cell.appendBang(in)
	}
}

// parseTransclusion parses '{' '{' '{' ZID '}' '}' '}'
func (cp *zmkP) parseTransclusion() (*sx.Pair /*ast.BlockNode*/, bool) {
	if cp.countDelim('{') != 3 {
		return nil, false
	}
	inp := cp.inp
	posA, posE := inp.Pos, 0

loop:

	for {
		switch inp.Ch {
		case input.EOS:
			return nil, false
		case '\n', '\r', ' ', '\t':
			if !hasQueryPrefix(inp.Src[posA:]) {
				return nil, false
			}
		case '\\':
			switch inp.Next() {
			case input.EOS, '\n', '\r':
				return nil, false
			}
		case '}':
			posE = inp.Pos
			if posA >= posE {
				return nil, false
			}
			if inp.Next() != '}' {
				continue
			}
			if inp.Next() != '}' {
				continue
			}
			break loop
		}
		inp.Next()
	}
	inp.Next() // consume last '}'
	a := cp.parseBlockAttributes()
	inp.SkipToEOL()
	refText := string(inp.Src[posA:posE])
	ref := ParseReference(refText)
	return sx.MakeList(sz.SymTransclude, a, ref), true
}
