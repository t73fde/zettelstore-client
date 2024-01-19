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

	"zettelstore.de/client.fossil/api"
	"zettelstore.de/client.fossil/input"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxhtml"
)

// parseInlineSlice parses a sequence of Inlines until EOS.
func (cp *zmkP) parseInlineSlice() *sx.Pair {
	var ins []sx.Object
	inp := cp.inp
	for inp.Ch != input.EOS {
		in := cp.parseInline()
		if in == nil {
			break
		}
		ins = append(ins, in)
	}
	return sx.MakeList(ins...).Cons(sz.SymBlock)
}

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
		case ' ', '\t':
			return cp.parseSpace()
		case '[':
			inp.Next()
			switch inp.Ch {
			case '[':
				in, success = cp.parseLink()
			case '@':
				in, success = cp.parseCite()
			case '^':
				in, success = cp.parseFootnote()
			case '!':
				in, success = cp.parseMark()
			}
		case '{':
			inp.Next()
			if inp.Ch == '{' {
				in, success = cp.parseEmbed()
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
		inp.Next()
		switch inp.Ch {
		// The following case must contain all runes that occur in parseInline!
		// Plus the closing brackets ] and } and ) and the middle |
		case input.EOS, '\n', '\r', ' ', '\t', '[', ']', '{', '}', '(', ')', '|', '%', '_', '*', '>', '~', '^', ',', '"', '#', ':', '\'', '@', '`', runeModGrave, '$', '=', '\\', '-', '&':
			return sx.String(string(inp.Src[pos:inp.Pos]))
		}
	}
}

func (cp *zmkP) parseBackslash() *sx.Pair {
	inp := cp.inp
	inp.Next()
	switch inp.Ch {
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
		return sx.String("\\")
	}
	if inp.Ch == ' ' {
		inp.Next()
		return sx.String("\u00a0")
	}
	pos := inp.Pos
	inp.Next()
	return sx.String(string(inp.Src[pos:inp.Pos]))
}

func (cp *zmkP) parseSpace() *sx.Pair {
	inp := cp.inp
	pos := inp.Pos
	for {
		inp.Next()
		switch inp.Ch {
		case ' ', '\t':
		default:
			if cp.inVerse {
				return sx.MakeList(sz.SymSpace, sx.String(string(inp.Src[pos:inp.Pos])))
			}
			return sx.MakeList(sz.SymSpace)
		}
	}
}

func (cp *zmkP) parseSoftBreak() *sx.Pair {
	cp.inp.EatEOL()
	return sx.MakeList(sz.SymSoft)
}

func (cp *zmkP) parseLink() (*sx.Pair /**ast.LinkNode*/, bool) {
	// if ref, is, ok := cp.parseReference('[', ']'); ok {
	// 	attrs := cp.parseInlineAttributes()
	// 	if len(ref) > 0 {
	// 		return &ast.LinkNode{
	// 			Ref:     ast.ParseReference(ref),
	// 			Inlines: is,
	// 			Attrs:   attrs,
	// 		}, true
	// 	}
	// }
	return nil, false
}

func hasQueryPrefix(src []byte) bool {
	return len(src) > len(api.QueryPrefix) && string(src[:len(api.QueryPrefix)]) == api.QueryPrefix
}

// func (cp *zmkP) parseReference(openCh, closeCh rune) (ref string, is ast.InlineSlice, _ bool) {
// 	inp := cp.inp
// 	inp.Next()
// 	cp.skipSpace()
// 	if inp.Ch == openCh {
// 		// Additional opening chars result in a fail
// 		return "", nil, false
// 	}
// 	pos := inp.Pos
// 	if !hasQueryPrefix(inp.Src[pos:]) {
// 		hasSpace, ok := cp.readReferenceToSep(closeCh)
// 		if !ok {
// 			return "", nil, false
// 		}
// 		if inp.Ch == '|' { // First part must be inline text
// 			if pos == inp.Pos { // [[| or {{|
// 				return "", nil, false
// 			}
// 			cp.inp = input.NewInput(inp.Src[pos:inp.Pos])
// 			for {
// 				in := cp.parseInline()
// 				if in == nil {
// 					break
// 				}
// 				is = append(is, in)
// 			}
// 			cp.inp = inp
// 			inp.Next()
// 		} else {
// 			if hasSpace {
// 				return "", nil, false
// 			}
// 			inp.SetPos(pos)
// 		}
// 	}

// 	cp.skipSpace()
// 	pos = inp.Pos
// 	if !cp.readReferenceToClose(closeCh) {
// 		return "", nil, false
// 	}
// 	ref = strings.TrimSpace(string(inp.Src[pos:inp.Pos]))
// 	inp.Next()
// 	if inp.Ch != closeCh {
// 		return "", nil, false
// 	}
// 	inp.Next()
// 	if len(is) == 0 {
// 		return ref, nil, true
// 	}
// 	return ref, is, true
// }

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
			inp.Next()
			switch inp.Ch {
			case input.EOS:
				return false, false
			case '\n', '\r':
				hasSpace = true
			}
		case '%':
			inp.Next()
			if inp.Ch == '%' {
				inp.SkipToEOL()
			}
			continue
		case closeCh:
			inp.Next()
			if inp.Ch == closeCh {
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
			inp.Next()
			switch inp.Ch {
			case input.EOS, '\n', '\r':
				return false
			}
		case closeCh:
			return true
		}
		inp.Next()
	}
}

func (cp *zmkP) parseCite() (*sx.Pair /**ast.CiteNode*/, bool) {
	// 	inp := cp.inp
	// 	inp.Next()
	// 	switch inp.Ch {
	// 	case ' ', ',', '|', ']', '\n', '\r':
	// 		return nil, false
	// 	}
	// 	pos := inp.Pos
	// loop:
	// 	for {
	// 		switch inp.Ch {
	// 		case input.EOS:
	// 			return nil, false
	// 		case ' ', ',', '|', ']', '\n', '\r':
	// 			break loop
	// 		}
	// 		inp.Next()
	// 	}
	// 	posL := inp.Pos
	// 	switch inp.Ch {
	// 	case ' ', ',', '|':
	// 		inp.Next()
	// 	}
	// 	ins, ok := cp.parseLinkLikeRest()
	// 	if !ok {
	// 		return nil, false
	// 	}
	// 	attrs := cp.parseInlineAttributes()
	return nil, false //&ast.CiteNode{Key: string(inp.Src[pos:posL]), Inlines: ins, Attrs: attrs}, true
}

func (cp *zmkP) parseFootnote() (*sx.Pair /**ast.FootnoteNode*/, bool) {
	// cp.inp.Next()
	// ins, ok := cp.parseLinkLikeRest()
	// if !ok {
	// 	return nil, false
	// }
	// attrs := cp.parseInlineAttributes()
	// return &ast.FootnoteNode{Inlines: ins, Attrs: attrs}, true
	return nil, false
}

func (cp *zmkP) parseLinkLikeRest() (*sx.Pair /*ast.InlineSlice*/, bool) {
	// cp.skipSpace()
	// ins := ast.InlineSlice{}
	// inp := cp.inp
	// for inp.Ch != ']' {
	// 	in := cp.parseInline()
	// 	if in == nil {
	// 		return nil, false
	// 	}
	// 	ins = append(ins, in)
	// 	if _, ok := in.(*ast.BreakNode); ok && input.IsEOLEOS(inp.Ch) {
	// 		return nil, false
	// 	}
	// }
	// inp.Next()
	// if len(ins) == 0 {
	// 	return nil, true
	// }
	// return ins, true
	return nil, false
}

func (cp *zmkP) parseEmbed() (*sx.Pair /*ast.InlineNode*/, bool) {
	// if ref, ins, ok := cp.parseReference('{', '}'); ok {
	// 	attrs := cp.parseInlineAttributes()
	// 	if len(ref) > 0 {
	// 		r := ast.ParseReference(ref)
	// 		return &ast.EmbedRefNode{
	// 			Ref:     r,
	// 			Inlines: ins,
	// 			Attrs:   attrs,
	// 		}, true
	// 	}
	// }
	return nil, false
}

func (cp *zmkP) parseMark() (*sx.Pair /**ast.MarkNode*/, bool) {
	// inp := cp.inp
	// inp.Next()
	// pos := inp.Pos
	// for inp.Ch != '|' && inp.Ch != ']' {
	// 	if !isNameRune(inp.Ch) {
	// 		return nil, false
	// 	}
	// 	inp.Next()
	// }
	// mark := inp.Src[pos:inp.Pos]
	// ins := ast.InlineSlice{}
	// if inp.Ch == '|' {
	// 	inp.Next()
	// 	var ok bool
	// 	ins, ok = cp.parseLinkLikeRest()
	// 	if !ok {
	// 		return nil, false
	// 	}
	// } else {
	// 	inp.Next()
	// }
	// mn := &ast.MarkNode{Mark: string(mark), Inlines: ins}
	// return mn, true
	return nil, false
	// Problematisch ist, dass hier noch nicht mn.Fragment und mn.Slug gesetzt werden.
	// Evtl. muss es ein PreMark-Symbol geben
}

func (cp *zmkP) parseComment() (res *sx.Pair, success bool) {
	inp := cp.inp
	inp.Next()
	if inp.Ch != '%' {
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
				sx.String(inp.Src[pos:inp.Pos]),
			), true
		}
		inp.Next()
	}
}

var mapRuneFormat = map[rune]sx.Symbol{
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
	inp.Next() // read 2nd formatting character
	if inp.Ch != fch {
		return nil, false
	}
	inp.Next()
	var inlines []sx.Object
	for {
		if inp.Ch == input.EOS {
			return nil, false
		}
		if inp.Ch == fch {
			inp.Next()
			if inp.Ch == fch {
				inp.Next()
				attrs := cp.parseInlineAttributes()
				fn := sx.MakeList(inlines...).Cons(attrs).Cons(symFormat)
				return fn, true
			}
			inlines = append(inlines, sx.MakeList(sz.SymText, sx.String(fch)))
		} else if in := cp.parseInline(); in != nil {
			if sym := in.Car(); (sym.IsEqual(sz.SymSoft) || sym.IsEqual(sz.SymHard)) && input.IsEOLEOS(inp.Ch) {
				return nil, false
			}
			inlines = append(inlines, in)
		}
	}
}

var mapRuneLiteral = map[rune]sx.Symbol{
	'@':          sz.SymLiteralZettel,
	'`':          sz.SymLiteralProg,
	runeModGrave: sz.SymLiteralProg,
	'\'':         sz.SymLiteralInput,
	'=':          sz.SymLiteralOutput,
	// No '$': ast.LiteralMath, because paring literal math is a little different
}

func (cp *zmkP) parseLiteral() (res *sx.Pair, success bool) {
	inp := cp.inp
	fch := inp.Ch
	symLiteral, ok := mapRuneLiteral[fch]
	if !ok {
		panic(fmt.Sprintf("%q is not a formatting char", fch))
	}
	inp.Next() // read 2nd formatting character
	if inp.Ch != fch {
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
			sb.WriteString(string(s))
		}
	}
}

func createLiteralNode(sym sx.Symbol, attrs *sx.Pair, content string) *sx.Pair {
	if sym == sz.SymLiteralZettel {
		assoc := attrs.Tail().Head()
		if p := assoc.Assoc(sx.String("")); p != nil {
			if val, isString := sx.GetString(p.Cdr()); isString && val == api.ValueSyntaxHTML {
				sym = sz.SymLiteralHTML
				// remove "" from attrs
				if assoc = assoc.RemoveAssoc(sx.String("")); assoc == nil {
					attrs = nil
				} else {
					attrs = sx.MakeList(sxhtml.SymAttr, assoc)
				}
			}
		}
	}
	return sx.MakeList(sym, attrs, sx.String(content))
}

func (cp *zmkP) parseLiteralMath() (res *sx.Pair /*ast.InlineNode*/, success bool) {
	inp := cp.inp
	inp.Next() // read 2nd formatting character
	if inp.Ch != '$' {
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
			fn := sx.MakeList(sz.SymLiteralMath, cp.parseInlineAttributes(), sx.String(content))
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
	return sx.MakeList(sz.SymText, sx.String("\u2013")), true
}

func (cp *zmkP) parseEntity() (res *sx.Pair, success bool) {
	if text, ok := cp.inp.ScanEntity(); ok {
		return sx.MakeList(sz.SymText, sx.String(text)), true
	}
	return nil, false
}
