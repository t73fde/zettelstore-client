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
	"maps"
	"slices"
	"strings"
	"unicode"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/sz"
	"t73f.de/r/zsx"
	"t73f.de/r/zsx/input"
)

// Parser allows to parse its plain text input into Zettelmarkup.
type Parser struct {
	inp               *input.Input // Input stream
	lists             []*sx.Pair   // Stack of lists
	lastRow           *sx.Pair     // Last row of table, or nil if not in table.
	descrl            *sx.Pair     // Current description list
	nestingLevel      int          // Count nesting of block and inline elements
	linkLikeRestLevel int          // Count nesting of link-like rests

	scanReference    func(string) *sx.Pair // Builds a reference node from a given string reference
	isSpaceReference func([]byte) bool     // Returns true, if src starts with a reference that allows white space
}

// Initialize the parser with the input stream and a reference scanner.
func (cp *Parser) Initialize(inp *input.Input) {
	cp.inp = inp
	cp.scanReference = sz.ScanReference
	cp.isSpaceReference = withQueryPrefix
}

// Parse tries to parse the input as a block element.
func (cp *Parser) Parse() *sx.Pair {
	cp.lists = nil
	cp.lastRow = nil
	cp.descrl = nil
	cp.nestingLevel = 0
	cp.linkLikeRestLevel = 0

	var lastPara *sx.Pair
	var blkBuild sx.ListBuilder
	for cp.inp.Ch != input.EOS {
		lastPara = cp.parseBlock(&blkBuild, lastPara)
	}
	if cp.nestingLevel != 0 {
		panic("Nesting level was not decremented")
	}
	if cp.linkLikeRestLevel != 0 {
		panic("Link nesting level was not decremented")
	}

	var pp postProcessor
	return pp.visitPairList(blkBuild.List(), nil).Cons(zsx.SymBlock)
}

func withQueryPrefix(src []byte) bool {
	return len(src) > len(api.QueryPrefix) && string(src[:len(api.QueryPrefix)]) == api.QueryPrefix
}

// runeModGrave is Unicode code point U+02CB (715) called "MODIFIER LETTER
// GRAVE ACCENT". On the iPad it is much more easier to type in this code point
// than U+0060 (96) "Grave accent" (aka backtick). Therefore, U+02CB will be
// considered equivalent to U+0060.
const runeModGrave = 'ˋ' // This is NOT '`'!

const maxNestingLevel = 50
const maxLinkLikeRest = 5

// clearStacked removes all multi-line nodes from parser.
func (cp *Parser) clearStacked() {
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
	var lb sx.ListBuilder
	for _, key := range slices.Sorted(maps.Keys(attrs)) {
		lb.Add(sx.Cons(sx.MakeString(key), sx.MakeString(attrs[key])))
	}
	return lb.List()
}

func parseNormalAttribute(inp *input.Input, attrs attrMap) bool {
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
	return parseAttributeValue(inp, key, attrs)
}

func parseAttributeValue(inp *input.Input, key string, attrs attrMap) bool {
	if inp.Next() == '"' {
		return parseQuotedAttributeValue(inp, key, attrs)
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

func parseQuotedAttributeValue(inp *input.Input, key string, attrs attrMap) bool {
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

func parseBlockAttributes(inp *input.Input) *sx.Pair {
	pos := inp.Pos
	for isNameRune(inp.Ch) {
		inp.Next()
	}
	if pos < inp.Pos {
		return attrMap{"": string(inp.Src[pos:inp.Pos])}.asPairAssoc()
	}

	// No immediate name: skip spaces
	inp.SkipSpace()
	return parseInlineAttributes(inp)
}

func parseInlineAttributes(inp *input.Input) *sx.Pair {
	pos := inp.Pos
	if attrs, success := doParseAttributes(inp); success {
		return attrs
	}
	inp.SetPos(pos)
	return nil
}

// doParseAttributes reads attributes.
func doParseAttributes(inp *input.Input) (*sx.Pair, bool) {
	if inp.Ch != '{' {
		return nil, false
	}
	inp.Next()
	a := attrMap{}
	if !parseAttributeValues(inp, a) {
		return nil, false
	}
	inp.Next()
	return a.asPairAssoc(), true
}

func parseAttributeValues(inp *input.Input, a attrMap) bool {
	for {
		skipSpaceLine(inp)
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
			if !parseAttributeValue(inp, "", a) {
				return false
			}
		default:
			if !parseNormalAttribute(inp, a) {
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

func skipSpaceLine(inp *input.Input) {
	for {
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

func isNameRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-' || ch == '_'
}

// isBreakSym return true if the object is either a soft or a hard break symbol.
func isBreakSym(obj sx.Object) bool {
	return zsx.SymSoft.IsEqual(obj) || zsx.SymHard.IsEqual(obj)
}
