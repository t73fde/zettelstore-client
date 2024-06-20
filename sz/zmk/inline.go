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
				in, success = cp.parseLinkEmbed('[', ']', true)
			case '@':
				in, success = cp.parseCite()
			case '^':
				in, success = cp.parseEndnote()
			case '!':
				in, success = cp.parseMark()
			}
		case '{':
			if inp.Next() == '{' {
				in, success = cp.parseLinkEmbed('{', '}', false)
			}
		case '%':
			in, success = cp.parseComment()
		case '_', '*', '>', '~', '^', ',', '"', '#', ':':
			in, success = cp.parseFormat()
		case '@', '\'', '`', '=', runeModGrave:
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

func (cp *zmkP) parseText() *sx.Pair {
	return sx.MakeList(sz.SymText, cp.parseString())
}

func (cp *zmkP) parseString() sx.String {
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
		case input.EOS, '\n', '\r', '[', ']', '{', '}', '(', ')', '|', '%', '_', '*', '>', '~', '^', ',', '"', '#', ':', '\'', '@', '`', runeModGrave, '$', '=', '\\', '-', '&':
			return sx.MakeString(string(inp.Src[pos:inp.Pos]))
		}
	}
}

func (cp *zmkP) parseBackslash() *sx.Pair {
	inp := cp.inp
	switch inp.Next() {
	case '\n', '\r':
		inp.EatEOL()
		return sx.MakeList(sz.SymHard)
	default:
		return sx.MakeList(sz.SymText, cp.parseBackslashRest())
	}
}

func (cp *zmkP) parseBackslashRest() sx.String {
	inp := cp.inp
	if input.IsEOLEOS(inp.Ch) {
		return sx.MakeString("\\")
	}
	if inp.Ch == ' ' {
		inp.Next()
		return sx.MakeString("\u00a0")
	}
	pos := inp.Pos
	inp.Next()
	return sx.MakeString(string(inp.Src[pos:inp.Pos]))
}

func (cp *zmkP) parseSoftBreak() *sx.Pair {
	cp.inp.EatEOL()
	return sx.MakeList(sz.SymSoft)
}

func (cp *zmkP) parseLinkEmbed(openCh, closeCh rune, forLink bool) (*sx.Pair, bool) {
	if refString, text, ok := cp.parseReference(openCh, closeCh); ok {
		attrs := cp.parseInlineAttributes()
		if len(refString) > 0 {
			ref := ParseReference(refString)
			refSym, _ := sx.GetSymbol(ref.Car())
			sym := sz.MapRefStateToLinkEmbed(refSym, forLink)
			ln := text.
				Cons(ref.Tail().Car()). // reference value
				Cons(attrs).
				Cons(sym)
			return ln, true
		}
	}
	return nil, false
}

func hasQueryPrefix(src []byte) bool {
	return len(src) > len(api.QueryPrefix) && string(src[:len(api.QueryPrefix)]) == api.QueryPrefix
}

func (cp *zmkP) parseReference(openCh, closeCh rune) (ref string, text *sx.Pair, _ bool) {
	inp := cp.inp
	inp.Next()
	cp.skipSpace()
	if inp.Ch == openCh {
		// Additional opening chars result in a fail
		return "", nil, false
	}
	var is sx.Vector
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
				is = append(is, in)
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

	cp.skipSpace()
	pos = inp.Pos
	if !cp.readReferenceToClose(closeCh) {
		return "", nil, false
	}
	ref = strings.TrimSpace(string(inp.Src[pos:inp.Pos]))
	if inp.Next() != closeCh {
		return "", nil, false
	}
	inp.Next()
	if len(is) == 0 {
		return ref, nil, true
	}
	return ref, sx.MakeList(is...), true
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
	cn := ins.Cons(sx.MakeString(string(inp.Src[pos:posL]))).Cons(attrs).Cons(sz.SymCite)
	return cn, true
}

func (cp *zmkP) parseEndnote() (*sx.Pair, bool) {
	cp.inp.Next()
	ins, ok := cp.parseLinkLikeRest()
	if !ok {
		return nil, false
	}
	attrs := cp.parseInlineAttributes()
	return ins.Cons(attrs).Cons(sz.SymEndnote), true
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
	mark := inp.Src[pos:inp.Pos]
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
	mn := ins.
		Cons(sx.MakeString("")). // Fragment
		Cons(sx.MakeString("")). // Slug
		Cons(sx.MakeString(string(mark))).
		Cons(sz.SymMark)
	return mn, true
	// Problematisch ist, dass hier noch nicht mn.Fragment und mn.Slug gesetzt werden.
	// Evtl. muss es ein PreMark-Symbol geben
}

func (cp *zmkP) parseLinkLikeRest() (*sx.Pair, bool) {
	cp.skipSpace()
	var ins sx.Vector
	inp := cp.inp
	for inp.Ch != ']' {
		in := cp.parseInline()
		if in == nil {
			return nil, false
		}
		ins = append(ins, in)
		if input.IsEOLEOS(inp.Ch) && sz.IsBreakSym(in.Car()) {
			return nil, false
		}
	}
	inp.Next()
	if len(ins) == 0 {
		return nil, true
	}
	return sx.MakeList(ins...), true
}

func (cp *zmkP) parseComment() (res *sx.Pair, success bool) {
	inp := cp.inp
	if inp.Next() != '%' {
		return nil, false
	}
	for inp.Ch == '%' {
		inp.Next()
	}
	attrs := cp.parseInlineAttributes()
	cp.skipSpace()
	pos := inp.Pos
	for {
		if input.IsEOLEOS(inp.Ch) {
			return sx.MakeList(
				sz.SymLiteralComment,
				attrs,
				sx.MakeString(string(inp.Src[pos:inp.Pos])),
			), true
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

func (cp *zmkP) parseFormat() (res *sx.Pair, success bool) {
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
	var inlines sx.Vector
	for {
		if inp.Ch == input.EOS {
			return nil, false
		}
		if inp.Ch == fch {
			if inp.Next() == fch {
				inp.Next()
				attrs := cp.parseInlineAttributes()
				fn := sx.MakeList(inlines...).Cons(attrs).Cons(symFormat)
				return fn, true
			}
			inlines = append(inlines, sx.MakeList(sz.SymText, sx.MakeString(string(fch))))
		} else if in := cp.parseInline(); in != nil {
			if input.IsEOLEOS(inp.Ch) && sz.IsBreakSym(in.Car()) {
				return nil, false
			}
			inlines = append(inlines, in)
		}
	}
}

var mapRuneLiteral = map[rune]*sx.Symbol{
	'@':          sz.SymLiteralZettel,
	'`':          sz.SymLiteralProg,
	runeModGrave: sz.SymLiteralProg,
	'\'':         sz.SymLiteralInput,
	'=':          sz.SymLiteralOutput,
	// No '$': sz.SymLiteralMath, because pairing literal math is a little different
}

func (cp *zmkP) parseLiteral() (res *sx.Pair, success bool) {
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
				return createLiteralNode(symLiteral, cp.parseInlineAttributes(), sb.String()), true
			}
			sb.WriteRune(fch)
			inp.Next()
		} else {
			s := cp.parseString()
			sb.WriteString(s.GetValue())
		}
	}
}

func createLiteralNode(sym *sx.Symbol, attrs *sx.Pair, content string) *sx.Pair {
	if sym.IsEqual(sz.SymLiteralZettel) {
		if p := attrs.Assoc(sx.MakeString("")); p != nil {
			if val, isString := sx.GetString(p.Cdr()); isString && val.GetValue() == api.ValueSyntaxHTML {
				sym = sz.SymLiteralHTML
				attrs = attrs.RemoveAssoc(sx.MakeString(""))
			}
		}
	}
	return sx.MakeList(sym, attrs, sx.MakeString(content))
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
			content := append([]byte{}, inp.Src[pos:inp.Pos]...)
			inp.Next()
			inp.Next()
			fn := sx.MakeList(sz.SymLiteralMath, cp.parseInlineAttributes(), sx.MakeString(string(content)))
			return fn, true
		}
		inp.Next()
	}
}

func (cp *zmkP) parseNdash() (res *sx.Pair, success bool) {
	inp := cp.inp
	if inp.Peek() != inp.Ch {
		return nil, false
	}
	inp.Next()
	inp.Next()
	return sx.MakeList(sz.SymText, sx.MakeString("\u2013")), true
}

func (cp *zmkP) parseEntity() (res *sx.Pair, success bool) {
	if text, ok := cp.inp.ScanEntity(); ok {
		return sx.MakeList(sz.SymText, sx.MakeString(text)), true
	}
	return nil, false
}
