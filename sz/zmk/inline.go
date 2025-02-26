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
	"slices"
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/input"
	"t73f.de/r/zsc/sz"
)

func (cp *zmkP) parseInline() *sx.Pair {
	inp := cp.inp
	pos := inp.Pos
	if cp.nestingLevel <= maxNestingLevel {
		cp.nestingLevel++
		defer func() { cp.nestingLevel-- }()

		var in *sx.Pair
		success := false
		switch inp.Ch {
		case input.EOS:
			return nil
		case '\n', '\r':
			return cp.parseSoftBreak()
		case '[':
			switch inp.Next() {
			case '[':
				in, success = cp.parseLink('[', ']')
			case '@':
				in, success = cp.parseCite()
			case '^':
				in, success = cp.parseEndnote()
			case '!':
				in, success = cp.parseMark()
			}
		case '{':
			if inp.Next() == '{' {
				in, success = cp.parseEmbed('{', '}')
			}
		case '%':
			in, success = cp.parseComment()
		case '_', '*', '>', '~', '^', ',', '"', '#', ':':
			in, success = cp.parseFormat()
		case '\'', '`', '=', runeModGrave:
			in, success = cp.parseLiteral()
		case '$':
			in, success = cp.parseLiteralMath()
		case '\\':
			return cp.parseBackslash()
		case '-':
			in, success = cp.parseNdash()
		case '&':
			in, success = cp.parseEntity()
		}
		if success {
			return in
		}
	}
	inp.SetPos(pos)
	return cp.parseText()
}

func (cp *zmkP) parseText() *sx.Pair { return sz.MakeText(cp.parseString()) }

func (cp *zmkP) parseString() string {
	inp := cp.inp
	pos := inp.Pos
	if inp.Ch == '\\' {
		cp.inp.Next()
		return cp.parseBackslashRest()
	}
	for {
		switch inp.Next() {
		// The following case must contain all runes that occur in parseInline!
		// Plus the closing brackets ] and } and ) and the middle |
		case input.EOS, '\n', '\r', '[', ']', '{', '}', '(', ')', '|', '%', '_', '*', '>', '~', '^', ',', '"', '#', ':', '\'', '`', runeModGrave, '$', '=', '\\', '-', '&':
			return string(inp.Src[pos:inp.Pos])
		}
	}
}

func (cp *zmkP) parseBackslash() *sx.Pair {
	inp := cp.inp
	switch inp.Next() {
	case '\n', '\r':
		inp.EatEOL()
		return sz.MakeHard()
	default:
		return sz.MakeText(cp.parseBackslashRest())
	}
}

func (cp *zmkP) parseBackslashRest() string {
	inp := cp.inp
	if input.IsEOLEOS(inp.Ch) {
		return "\\"
	}
	if inp.Ch == ' ' {
		inp.Next()
		return "\u00a0"
	}
	pos := inp.Pos
	inp.Next()
	return string(inp.Src[pos:inp.Pos])
}

func (cp *zmkP) parseSoftBreak() *sx.Pair {
	cp.inp.EatEOL()
	return sz.MakeSoft()
}

func (cp *zmkP) parseLink(openCh, closeCh rune) (*sx.Pair, bool) {
	if refString, text, ok := cp.parseReference(openCh, closeCh); ok {
		attrs := cp.parseInlineAttributes()
		if len(refString) > 0 {
			ref := ParseReference(refString)
			refSym, _ := sx.GetSymbol(ref.Car())
			sym := sz.MapRefStateToLink(refSym)
			return sz.MakeLink(sym, attrs, ref.Tail().Car().(sx.String).GetValue(), text), true
		}
	}
	return nil, false
}
func (cp *zmkP) parseEmbed(openCh, closeCh rune) (*sx.Pair, bool) {
	if refString, text, ok := cp.parseReference(openCh, closeCh); ok {
		attrs := cp.parseInlineAttributes()
		if len(refString) > 0 {
			return sz.MakeEmbed(attrs, ParseReference(refString), "", text), true
		}
	}
	return nil, false
}

func hasQueryPrefix(src []byte) bool {
	return len(src) > len(api.QueryPrefix) && string(src[:len(api.QueryPrefix)]) == api.QueryPrefix
}

func (cp *zmkP) parseReference(openCh, closeCh rune) (string, *sx.Pair, bool) {
	inp := cp.inp
	inp.Next()
	inp.SkipSpace()
	if inp.Ch == openCh {
		// Additional opening chars result in a fail
		return "", nil, false
	}
	var lb sx.ListBuilder
	pos := inp.Pos
	if !hasQueryPrefix(inp.Src[pos:]) {
		hasSpace, ok := cp.readReferenceToSep(closeCh)
		if !ok {
			return "", nil, false
		}
		if inp.Ch == '|' { // First part must be inline text
			if pos == inp.Pos { // [[| or {{|
				return "", nil, false
			}
			cp.inp = input.NewInput(inp.Src[pos:inp.Pos])
			for {
				in := cp.parseInline()
				if in == nil {
					break
				}
				lb.Add(in)
			}
			cp.inp = inp
			inp.Next()
		} else {
			if hasSpace {
				return "", nil, false
			}
			inp.SetPos(pos)
		}
	}

	inp.SkipSpace()
	pos = inp.Pos
	if !cp.readReferenceToClose(closeCh) {
		return "", nil, false
	}
	ref := strings.TrimSpace(string(inp.Src[pos:inp.Pos]))
	if inp.Next() != closeCh {
		return "", nil, false
	}
	inp.Next()
	return ref, lb.List(), true
}

func (cp *zmkP) readReferenceToSep(closeCh rune) (bool, bool) {
	hasSpace := false
	inp := cp.inp
	for {
		switch inp.Ch {
		case input.EOS:
			return false, false
		case '\n', '\r', ' ':
			hasSpace = true
		case '|':
			return hasSpace, true
		case '\\':
			switch inp.Next() {
			case input.EOS:
				return false, false
			case '\n', '\r':
				hasSpace = true
			}
		case '%':
			if inp.Next() == '%' {
				inp.SkipToEOL()
			}
			continue
		case closeCh:
			if inp.Next() == closeCh {
				return hasSpace, true
			}
			continue
		}
		inp.Next()
	}
}

func (cp *zmkP) readReferenceToClose(closeCh rune) bool {
	inp := cp.inp
	pos := inp.Pos
	for {
		switch inp.Ch {
		case input.EOS:
			return false
		case '\t', '\r', '\n', ' ':
			if !hasQueryPrefix(inp.Src[pos:]) {
				return false
			}
		case '\\':
			switch inp.Next() {
			case input.EOS, '\n', '\r':
				return false
			}
		case closeCh:
			return true
		}
		inp.Next()
	}
}

func (cp *zmkP) parseCite() (*sx.Pair, bool) {
	inp := cp.inp
	switch inp.Next() {
	case ' ', ',', '|', ']', '\n', '\r':
		return nil, false
	}
	pos := inp.Pos
loop:
	for {
		switch inp.Ch {
		case input.EOS:
			return nil, false
		case ' ', ',', '|', ']', '\n', '\r':
			break loop
		}
		inp.Next()
	}
	posL := inp.Pos
	switch inp.Ch {
	case ' ', ',', '|':
		inp.Next()
	}
	ins, ok := cp.parseLinkLikeRest()
	if !ok {
		return nil, false
	}
	attrs := cp.parseInlineAttributes()
	return sz.MakeCite(attrs, string(inp.Src[pos:posL]), ins), true
}

func (cp *zmkP) parseEndnote() (*sx.Pair, bool) {
	cp.inp.Next()
	ins, ok := cp.parseLinkLikeRest()
	if !ok {
		return nil, false
	}
	attrs := cp.parseInlineAttributes()
	return sz.MakeEndnote(attrs, ins), true
}

func (cp *zmkP) parseMark() (*sx.Pair, bool) {
	inp := cp.inp
	inp.Next()
	pos := inp.Pos
	for inp.Ch != '|' && inp.Ch != ']' {
		if !isNameRune(inp.Ch) {
			return nil, false
		}
		inp.Next()
	}
	mark := string(inp.Src[pos:inp.Pos])
	var ins *sx.Pair
	if inp.Ch == '|' {
		inp.Next()
		var ok bool
		ins, ok = cp.parseLinkLikeRest()
		if !ok {
			return nil, false
		}
	} else {
		inp.Next()
	}
	return sz.MakeMark(mark, "", "", ins), true
	// Problematisch ist, dass hier noch nicht mn.Fragment und mn.Slug gesetzt werden.
	// Evtl. muss es ein PreMark-Symbol geben
}

func (cp *zmkP) parseLinkLikeRest() (*sx.Pair, bool) {
	var ins sx.ListBuilder
	inp := cp.inp
	inp.SkipSpace()
	for inp.Ch != ']' {
		in := cp.parseInline()
		if in == nil {
			return nil, false
		}
		ins.Add(in)
		if input.IsEOLEOS(inp.Ch) && sz.IsBreakSym(in.Car()) {
			return nil, false
		}
	}
	inp.Next()
	return ins.List(), true
}

func (cp *zmkP) parseComment() (*sx.Pair, bool) {
	inp := cp.inp
	if inp.Next() != '%' {
		return nil, false
	}
	for inp.Ch == '%' {
		inp.Next()
	}
	attrs := cp.parseInlineAttributes()
	inp.SkipSpace()
	pos := inp.Pos
	for {
		if input.IsEOLEOS(inp.Ch) {
			return sz.MakeLiteral(sz.SymLiteralComment, attrs, string(inp.Src[pos:inp.Pos])), true
		}
		inp.Next()
	}
}

var mapRuneFormat = map[rune]*sx.Symbol{
	'_': sz.SymFormatEmph,
	'*': sz.SymFormatStrong,
	'>': sz.SymFormatInsert,
	'~': sz.SymFormatDelete,
	'^': sz.SymFormatSuper,
	',': sz.SymFormatSub,
	'"': sz.SymFormatQuote,
	'#': sz.SymFormatMark,
	':': sz.SymFormatSpan,
}

func (cp *zmkP) parseFormat() (*sx.Pair, bool) {
	inp := cp.inp
	fch := inp.Ch
	symFormat, ok := mapRuneFormat[fch]
	if !ok {
		panic(fmt.Sprintf("%q is not a formatting char", fch))
	}
	// read 2nd formatting character
	if inp.Next() != fch {
		return nil, false
	}
	inp.Next()
	var inlines sx.ListBuilder
	for {
		if inp.Ch == input.EOS {
			return nil, false
		}
		if inp.Ch == fch {
			if inp.Next() == fch {
				inp.Next()
				attrs := cp.parseInlineAttributes()
				return sz.MakeFormat(symFormat, attrs, inlines.List()), true
			}
			inlines.Add(sz.MakeText(string(fch)))
		} else if in := cp.parseInline(); in != nil {
			if input.IsEOLEOS(inp.Ch) && sz.IsBreakSym(in.Car()) {
				return nil, false
			}
			inlines.Add(in)
		}
	}
}

var mapRuneLiteral = map[rune]*sx.Symbol{
	'`':          sz.SymLiteralCode,
	runeModGrave: sz.SymLiteralCode,
	'\'':         sz.SymLiteralInput,
	'=':          sz.SymLiteralOutput,
	// No '$': sz.SymLiteralMath, because pairing literal math is a little different
}

func (cp *zmkP) parseLiteral() (*sx.Pair, bool) {
	inp := cp.inp
	fch := inp.Ch
	symLiteral, ok := mapRuneLiteral[fch]
	if !ok {
		panic(fmt.Sprintf("%q is not a formatting char", fch))
	}
	// read 2nd formatting character
	if inp.Next() != fch {
		return nil, false
	}
	inp.Next()
	var sb strings.Builder
	for {
		if inp.Ch == input.EOS {
			return nil, false
		}
		if inp.Ch == fch {
			if inp.Peek() == fch {
				inp.Next()
				inp.Next()
				return sz.MakeLiteral(symLiteral, cp.parseInlineAttributes(), sb.String()), true
			}
			sb.WriteRune(fch)
			inp.Next()
		} else {
			s := cp.parseString()
			sb.WriteString(s)
		}
	}
}

func (cp *zmkP) parseLiteralMath() (res *sx.Pair, success bool) {
	inp := cp.inp
	// read 2nd formatting character
	if inp.Next() != '$' {
		return nil, false
	}
	inp.Next()
	pos := inp.Pos
	for {
		if inp.Ch == input.EOS {
			return nil, false
		}
		if inp.Ch == '$' && inp.Peek() == '$' {
			content := slices.Clone(inp.Src[pos:inp.Pos])
			inp.Next()
			inp.Next()
			return sz.MakeLiteral(sz.SymLiteralMath, cp.parseInlineAttributes(), string(content)), true
		}
		inp.Next()
	}
}

func (cp *zmkP) parseNdash() (*sx.Pair, bool) {
	inp := cp.inp
	if inp.Peek() != inp.Ch {
		return nil, false
	}
	inp.Next()
	inp.Next()
	return sz.MakeText("\u2013"), true
}

func (cp *zmkP) parseEntity() (res *sx.Pair, success bool) {
	if text, ok := cp.inp.ScanEntity(); ok {
		return sz.MakeText(text), true
	}
	return nil, false
}
