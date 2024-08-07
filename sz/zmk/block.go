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

	"t73f.de/r/sx"
	"t73f.de/r/zsc/input"
	"t73f.de/r/zsc/sz"
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
			cp.descrl = nil
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
			cp.descrl = nil
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
	if pair0.Car().IsEqual(sz.SymText) && sz.IsBreakSym(pair1.Car()) {
		if args := pair0.Tail(); args != nil {
			if val, isString := sx.GetString(args.Car()); isString {
				for _, ch := range val.GetValue() {
					if !input.IsSpace(ch) {
						return false
					}
				}
				return true
			}
		}
	}
	return false
}

var symSeparator = sx.MakeSymbol("sEpArAtOr")

func (cp *zmkP) cleanupListsAfterEOL() {
	for _, l := range cp.lists {
		l.LastPair().Head().LastPair().AppendBang(sx.Cons(symSeparator, nil))
	}
	if descrl := cp.descrl; descrl != nil {
		if lastPair, pos := lastPairPos(descrl); pos > 1 && pos%2 == 0 {
			lastPair.Head().LastPair().AppendBang(sx.Cons(symSeparator, nil))
		}
	}
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

// parsePara parses paragraphed inline material as a sx.Vector.
func (cp *zmkP) parsePara() (result sx.Vector) {
	for {
		in := cp.parseInline()
		if in == nil {
			return result
		}
		result = append(result, in)
		if sz.IsBreakSym(in.Car()) {
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
	var sym *sx.Symbol
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
				rn = sx.MakeList(sym, attrs, sx.MakeString(string(content)))
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

	var sym *sx.Symbol
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
	var blocksBuilder sx.ListBuilder
	var lastPara *sx.Pair
	inp.EatEOL()
	for {
		posL := inp.Pos
		switch inp.Ch {
		case fch:
			if cp.countDelim(fch) >= cnt {
				ins := cp.parseRegionLastLine()
				rn = ins.Cons(blocksBuilder.List()).Cons(attrs).Cons(sym)
				return rn, true
			}
			inp.SetPos(posL)
		case input.EOS:
			return nil, false
		}
		bn, cont := cp.parseBlock(lastPara)
		if bn != nil {
			blocksBuilder.Add(bn)
		}
		if !cont {
			lastPara = bn
		}
	}
}

// parseRegionLastLine parses the last line of a region and returns its inline text.
func (cp *zmkP) parseRegionLastLine() *sx.Pair {
	inp := cp.inp
	cp.clearStacked() // remove any lists defined in the region
	inp.SkipSpace()
	var region sx.ListBuilder
	for {
		switch inp.Ch {
		case input.EOS, '\n', '\r':
			return region.List()
		}
		in := cp.parseInline()
		if in == nil {
			return region.List()
		}
		region.Add(in)
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
	inp.SkipSpace()
	if delims > 7 {
		delims = 7
	}
	level := int64(delims - 2)
	var attrs *sx.Pair
	var text sx.ListBuilder
	for {
		if input.IsEOLEOS(inp.Ch) {
			return createHeading(level, attrs, text.List()), true
		}
		in := cp.parseInline()
		if in == nil {
			return createHeading(level, attrs, text.List()), true
		}
		text.Add(in)
		if inp.Ch == '{' && inp.Peek() != '{' {
			attrs = cp.parseBlockAttributes()
			inp.SkipToEOL()
			return createHeading(level, attrs, text.List()), true
		}
	}
}
func createHeading(level int64, attrs, text *sx.Pair) *sx.Pair {
	return text.
		Cons(sx.MakeString("")). // Fragment
		Cons(sx.MakeString("")). // Slug
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
func (cp *zmkP) parseNestedList() (res *sx.Pair, success bool) {
	kinds := cp.parseNestedListKinds()
	if len(kinds) == 0 {
		return nil, false
	}
	inp := cp.inp
	inp.SkipSpace()
	if !kinds[len(kinds)-1].IsEqual(sz.SymListQuote) && input.IsEOLEOS(inp.Ch) {
		return nil, false
	}

	if len(kinds) < len(cp.lists) {
		cp.lists = cp.lists[:len(kinds)]
	}
	ln, newLnCount := cp.buildNestedList(kinds)
	pv := cp.parseLinePara()
	bn := sx.Cons(sz.SymBlock, nil)
	if len(pv) != 0 {
		bn.AppendBang(pv.MakeList().Cons(sz.SymPara))
	}
	lastItemPair := ln.LastPair()
	lastItemPair.AppendBang(bn)
	return cp.cleanupParsedNestedList(newLnCount)
}

func (cp *zmkP) parseNestedListKinds() []*sx.Symbol {
	inp := cp.inp
	result := make([]*sx.Symbol, 0, 8)
	for {
		var sym *sx.Symbol
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

func (cp *zmkP) buildNestedList(kinds []*sx.Symbol) (ln *sx.Pair, newLnCount int) {
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

func (cp *zmkP) cleanupParsedNestedList(newLnCount int) (res *sx.Pair, success bool) {
	childPos := len(cp.lists) - 1
	parentPos := childPos - 1
	for i := 0; i < newLnCount; i++ {
		if parentPos < 0 {
			return cp.lists[0], true
		}
		parentLn := cp.lists[parentPos]
		childLn := cp.lists[childPos]
		if firstParent := parentLn.Tail(); firstParent != nil {
			// Add list to last item of the parent list
			lastParent := firstParent.LastPair()
			lastParent.Head().LastPair().AppendBang(childLn)
		} else {
			// Set list to first child of parent.
			childBlock := sx.MakeList(sz.SymBlock, cp.lists[childPos])
			parentLn.LastPair().AppendBang(childBlock)
		}
		childPos--
		parentPos--
	}
	return nil, true
}

// parseDefTerm parses a term of a definition list.
func (cp *zmkP) parseDefTerm() (res *sx.Pair, success bool) {
	inp := cp.inp
	if inp.Next() != ' ' {
		return nil, false
	}
	inp.Next()
	inp.SkipSpace()
	descrl := cp.descrl
	if descrl == nil {
		descrl = sx.Cons(sz.SymDescription, nil)
		cp.descrl = descrl
		res = descrl
	}
	lastPair, pos := lastPairPos(descrl)
	for {
		in := cp.parseInline()
		if in == nil {
			if pos%2 == 0 {
				// lastPair is either the empty description list or the last block of definitions
				return nil, false
			}
			// lastPair is the definition term
			return res, true
		}
		if pos%2 == 0 {
			// lastPair is either the empty description list or the last block of definitions
			lastPair = lastPair.AppendBang(sx.Cons(in, nil))
			pos++
		} else {
			// lastPair is the term part and we need to append the inline list just read
			lastPair.Head().LastPair().AppendBang(in)
		}
		if sz.IsBreakSym(in.Car()) {
			return res, true
		}
	}
}

// parseDefDescr parses a description of a definition list.
func (cp *zmkP) parseDefDescr() (res *sx.Pair, success bool) {
	inp := cp.inp
	if inp.Next() != ' ' {
		return nil, false
	}
	inp.Next()
	inp.SkipSpace()
	descrl := cp.descrl
	lastPair, pos := lastPairPos(descrl)
	if descrl == nil || pos <= 0 {
		// No term given
		return nil, false
	}

	pn := cp.parseLinePara()
	if len(pn) == 0 {
		return nil, false
	}

	newDef := sx.MakeList(sz.SymBlock, pn.MakeList().Cons(sz.SymPara))
	if pos%2 == 1 {
		// Just a term, but no definitions
		lastPair.AppendBang(sx.MakeList(sz.SymBlock, newDef))
	} else {
		// lastPara points a the last definition
		lastPair.Head().LastPair().AppendBang(newDef)
	}
	return nil, true
}

func lastPairPos(p *sx.Pair) (*sx.Pair, int) {
	cnt := 0
	for node := p; node != nil; {
		next := node.Tail()
		if next == nil {
			return node, cnt
		}
		node = next
		cnt++
	}
	return nil, -1
}

// parseIndent parses initial spaces to continue a list.
func (cp *zmkP) parseIndent() bool {
	inp := cp.inp
	cnt := 0
	for {
		if inp.Next() != ' ' {
			break
		}
		cnt++
	}
	if cp.lists != nil {
		return cp.parseIndentForList(cnt)
	}
	if cp.descrl != nil {
		return cp.parseIndentForDescription(cnt)
	}
	return false
}

func (cp *zmkP) parseIndentForList(cnt int) bool {
	if len(cp.lists) < cnt {
		cnt = len(cp.lists)
	}
	cp.lists = cp.lists[:cnt]
	if cnt == 0 {
		return false
	}
	pv := cp.parseLinePara()
	if len(pv) == 0 {
		return false
	}
	ln := cp.lists[cnt-1]
	lbn := ln.LastPair().Head()
	lpn := lbn.LastPair().Head()
	pvList := pv.MakeList()
	if lpn.Car().IsEqual(sz.SymPara) {
		lpn.LastPair().SetCdr(pvList)
	} else {
		lbn.LastPair().AppendBang(pvList.Cons(sz.SymPara))
	}
	return true
}

func (cp *zmkP) parseIndentForDescription(cnt int) bool {
	descrl := cp.descrl
	lastPair, pos := lastPairPos(descrl)
	if cnt < 1 || pos < 1 {
		return false
	}
	if pos%2 == 1 {
		// Continuation of a definition term
		for {
			in := cp.parseInline()
			if in == nil {
				return true
			}
			lastPair.Head().LastPair().AppendBang(in)
			if sz.IsBreakSym(in.Car()) {
				return true
			}
		}
	}

	// Continuation of a definition description
	pn := cp.parseLinePara()
	if len(pn) == 0 {
		return false
	}
	bn := lastPair.Head()
	para := bn.LastPair().Head().LastPair().Head()
	pnList := pn.MakeList()
	if para.Car().IsEqual(sz.SymPara) {
		para.LastPair().SetCdr(pnList)
	} else {
		bn.LastPair().AppendBang(pnList.Cons(sz.SymPara))
	}
	return true
}

// parseLinePara parses one paragraph of inline material.
func (cp *zmkP) parseLinePara() sx.Vector {
	var ins sx.Vector
	for {
		in := cp.parseInline()
		if in == nil {
			return ins
		}
		ins = append(ins, in)
		if sz.IsBreakSym(in.Car()) {
			return ins
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
	var row sx.ListBuilder
	for {
		inp.Next()
		cell := cp.parseCell()
		if cell != nil {
			row.Add(cell)
		}
		switch inp.Ch {
		case '\n', '\r':
			inp.EatEOL()
			fallthrough
		case input.EOS:
			// add to table
			if cp.lastRow == nil {
				if row.IsEmpty() {
					return nil
				}
				cp.lastRow = sx.Cons(row.List(), nil)
				return cp.lastRow.Cons(nil).Cons(sz.SymTable)
			}
			cp.lastRow = cp.lastRow.AppendBang(row.List())
			return nil
		}
		// inp.Ch must be '|'
	}
}

// parseCell parses one single cell of a table row.
func (cp *zmkP) parseCell() *sx.Pair {
	inp := cp.inp
	var cell sx.ListBuilder
	for {
		if input.IsEOLEOS(inp.Ch) {
			if cell.IsEmpty() {
				return nil
			}
			return cell.List().Cons(sz.SymCell)
		}
		if inp.Ch == '|' {
			return cell.List().Cons(sz.SymCell)
		}

		in := cp.parseInline()
		cell.Add(in)
	}
}

// parseTransclusion parses '{' '{' '{' ZID '}' '}' '}'
func (cp *zmkP) parseTransclusion() (*sx.Pair, bool) {
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
