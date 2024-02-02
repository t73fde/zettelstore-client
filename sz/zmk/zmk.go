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

// Package zmk provides a parser for zettelmarkup.
package zmk

import (
	"slices"
	"strings"
	"unicode"

	"zettelstore.de/client.fossil/input"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
)

func ParseBlocks(inp *input.Input) *sx.Pair {
	parser := zmkP{inp: inp}

	var lastPara *sx.Pair
	var blkBuild sx.ListBuilder
	for inp.Ch != input.EOS {
		bn, cont := parser.parseBlock(lastPara)
		if bn != nil {
			blkBuild.Add(bn)
		}
		if !cont {
			if bn.Car().IsEqual(sz.SymPara) {
				lastPara = bn
			} else {
				lastPara = nil
			}
		}
	}
	if parser.nestingLevel != 0 {
		panic("Nesting level was not decremented")
	}

	if bs := postProcessPairList(blkBuild.List(), nil); bs != nil {
		return bs.Cons(sz.SymBlock)
	}
	return nil
}

func ParseInlines(inp *input.Input) *sx.Pair {
	parser := zmkP{inp: inp}
	var ins sx.Vector
	for inp.Ch != input.EOS {
		in := parser.parseInline()
		if in == nil {
			break
		}
		ins = append(ins, in)
	}

	if is := postProcess(ins, nil); is != nil {
		return is.Cons(sz.SymInline)
	}
	return nil
}

type zmkP struct {
	inp          *input.Input // Input stream
	lists        []*sx.Pair   // Stack of lists
	lastRow      *sx.Pair     // Last row of table, or nil if not in table.
	descrl       *sx.Pair     // Current description list
	nestingLevel int          // Count nesting of block and inline elements

	inVerse bool // Currently in a vers region?
}

// runeModGrave is Unicode code point U+02CB (715) called "MODIFIER LETTER
// GRAVE ACCENT". On the iPad it is much more easier to type in this code point
// than U+0060 (96) "Grave accent" (aka backtick). Therefore, U+02CB will be
// considered equivalent to U+0060.
const runeModGrave = 'ˋ' // This is NOT '`'!

const maxNestingLevel = 50

// clearStacked removes all multi-line nodes from parser.
func (cp *zmkP) clearStacked() {
	cp.lists = nil
	cp.lastRow = nil
	cp.descrl = nil
}

type attrMap map[string]string

func (attrs attrMap) updateAttrs(key, val string) {
	if prevVal := attrs[key]; len(prevVal) > 0 {
		attrs[key] = prevVal + " " + val
	} else {
		attrs[key] = val
	}
}

func (attrs attrMap) asPairAssoc() *sx.Pair {
	names := make([]string, 0, len(attrs))
	for n := range attrs {
		names = append(names, n)
	}
	slices.Sort(names)
	var assoc *sx.Pair = nil
	for i := len(names) - 1; i >= 0; i-- {
		n := names[i]
		assoc = assoc.Cons(sx.Cons(sx.String(n), sx.String(attrs[n])))
	}
	return assoc
}

func (cp *zmkP) parseNormalAttribute(attrs attrMap) bool {
	inp := cp.inp
	posK := inp.Pos
	for isNameRune(inp.Ch) {
		inp.Next()
	}
	if posK == inp.Pos {
		return false
	}
	key := string(inp.Src[posK:inp.Pos])
	if inp.Ch != '=' {
		attrs[key] = ""
		return true
	}
	return cp.parseAttributeValue(key, attrs)
}

func (cp *zmkP) parseAttributeValue(key string, attrs attrMap) bool {
	inp := cp.inp
	if inp.Next() == '"' {
		return cp.parseQuotedAttributeValue(key, attrs)
	}
	posV := inp.Pos
	for {
		switch inp.Ch {
		case input.EOS:
			return false
		case '\n', '\r', ' ', ',', '}':
			attrs.updateAttrs(key, string(inp.Src[posV:inp.Pos]))
			return true
		}
		inp.Next()
	}
}

func (cp *zmkP) parseQuotedAttributeValue(key string, attrs attrMap) bool {
	inp := cp.inp
	inp.Next()
	var sb strings.Builder
	for {
		switch inp.Ch {
		case input.EOS:
			return false
		case '"':
			attrs.updateAttrs(key, sb.String())
			inp.Next()
			return true
		case '\\':
			switch inp.Next() {
			case input.EOS, '\n', '\r':
				return false
			}
			fallthrough
		default:
			sb.WriteRune(inp.Ch)
			inp.Next()
		}
	}

}

func (cp *zmkP) parseBlockAttributes() *sx.Pair {
	inp := cp.inp
	pos := inp.Pos
	for isNameRune(inp.Ch) {
		inp.Next()
	}
	if pos < inp.Pos {
		return attrMap{"": string(inp.Src[pos:inp.Pos])}.asPairAssoc()
	}

	// No immediate name: skip spaces
	cp.skipSpace()
	return cp.parseInlineAttributes()
}

func (cp *zmkP) parseInlineAttributes() *sx.Pair {
	inp := cp.inp
	pos := inp.Pos
	if attrs, success := cp.doParseAttributes(); success {
		return attrs
	}
	inp.SetPos(pos)
	return nil
}

// doParseAttributes reads attributes.
func (cp *zmkP) doParseAttributes() (res *sx.Pair, success bool) {
	inp := cp.inp
	if inp.Ch != '{' {
		return nil, false
	}
	inp.Next()
	a := attrMap{}
	if !cp.parseAttributeValues(a) {
		return nil, false
	}
	inp.Next()
	return a.asPairAssoc(), true
}

func (cp *zmkP) parseAttributeValues(a attrMap) bool {
	inp := cp.inp
	for {
		cp.skipSpaceLine()
		switch inp.Ch {
		case input.EOS:
			return false
		case '}':
			return true
		case '.':
			inp.Next()
			posC := inp.Pos
			for isNameRune(inp.Ch) {
				inp.Next()
			}
			if posC == inp.Pos {
				return false
			}
			a.updateAttrs("class", string(inp.Src[posC:inp.Pos]))
		case '=':
			delete(a, "")
			if !cp.parseAttributeValue("", a) {
				return false
			}
		default:
			if !cp.parseNormalAttribute(a) {
				return false
			}
		}

		switch inp.Ch {
		case '}':
			return true
		case '\n', '\r':
		case ' ', ',':
			inp.Next()
		default:
			return false
		}
	}
}

func (cp *zmkP) skipSpaceLine() {
	for inp := cp.inp; ; {
		switch inp.Ch {
		case ' ':
			inp.Next()
		case '\n', '\r':
			inp.EatEOL()
		default:
			return
		}
	}
}

func (cp *zmkP) skipSpace() {
	for inp := cp.inp; inp.Ch == ' '; {
		inp.Next()
	}
}

func isNameRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-' || ch == '_'
}
