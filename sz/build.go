//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import "t73f.de/r/sx"

// MakeBlock builds a block node.
func MakeBlock(blocks ...*sx.Pair) *sx.Pair {
	var lb sx.ListBuilder
	lb.Add(SymBlock)
	for _, block := range blocks {
		lb.Add(block)
	}
	return lb.List()
}

// MakeBlockList builds a block node from a list of blocks.
func MakeBlockList(blocks *sx.Pair) *sx.Pair { return blocks.Cons(SymBlock) }

// MakeInlineList builds an inline node from a list of inlines.
func MakeInlineList(inlines *sx.Pair) *sx.Pair { return inlines.Cons(SymInline) }

// MakePara builds a paragraph node.
func MakePara(inlines *sx.Pair) *sx.Pair { return inlines.Cons(SymPara) }

// MakeVerbatim builds a node for verbatim text.
func MakeVerbatim(sym *sx.Symbol, attrs *sx.Pair, content string) *sx.Pair {
	return sx.MakeList(sym, attrs, sx.MakeString(content))
}

// MakeRegion builds a region node.
func MakeRegion(sym *sx.Symbol, attrs *sx.Pair, blocks *sx.Pair, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(blocks).Cons(attrs).Cons(sym)
}

// MakeHeading builds a heading node.
func MakeHeading(level int, attrs, text *sx.Pair, slug, fragment string) *sx.Pair {
	return text.
		Cons(sx.MakeString(fragment)).
		Cons(sx.MakeString(slug)).
		Cons(attrs).
		Cons(sx.Int64(level)).
		Cons(SymHeading)
}

// MakeThematic builds a node to implement a thematic break.
func MakeThematic(attrs *sx.Pair) *sx.Pair {
	return sx.Cons(SymThematic, sx.Cons(attrs, sx.Nil()))
}

// MakeCell builds a table cell node.
func MakeCell(sym *sx.Symbol, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(sym)
}

// MakeTransclusion builds a transclusion node.
func MakeTransclusion(attrs *sx.Pair, ref *sx.Pair) *sx.Pair {
	var lb sx.ListBuilder
	lb.AddN(SymTransclude, attrs, ref)
	return lb.List()
}

// MakeBLOB builds a block BLOB node.
func MakeBLOB(description *sx.Pair, syntax, content string) *sx.Pair {
	return sx.Cons(SymBLOB,
		sx.Cons(description,
			sx.Cons(sx.MakeString(syntax),
				sx.Cons(sx.MakeString(content), sx.Nil()))))
}

// MakeText builds a text node.
func MakeText(text string) *sx.Pair {
	return sx.Cons(SymText, sx.Cons(sx.MakeString(text), sx.Nil()))
}

// MakeSoft builds a node for a soft line break.
func MakeSoft() *sx.Pair { return sx.Cons(SymSoft, sx.Nil()) }

// MakeHard builds a node for a hard line break.
func MakeHard() *sx.Pair { return sx.Cons(SymHard, sx.Nil()) }

// MakeLink builds a link node.
func MakeLink(sym *sx.Symbol, attrs *sx.Pair, ref string, text *sx.Pair) *sx.Pair {
	return text.Cons(sx.MakeString(ref)).Cons(attrs).Cons(sym)
}

// MakeEmbed builds a embed node.
func MakeEmbed(attrs *sx.Pair, ref sx.Object, syntax string, text *sx.Pair) *sx.Pair {
	return text.Cons(sx.MakeString(syntax)).Cons(ref).Cons(attrs).Cons(SymEmbed)
}

// MakeEmbedBLOB builds an embedded inline BLOB node.
func MakeEmbedBLOB(attrs *sx.Pair, syntax string, content string, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(sx.MakeString(content)).Cons(sx.MakeString(syntax)).Cons(attrs).Cons(SymEmbedBLOB)
}

// MakeCite builds a node that specifies a citation.
func MakeCite(attrs *sx.Pair, text string, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(sx.MakeString(text)).Cons(attrs).Cons(SymCite)
}

// MakeEndnote builds an endnote node.
func MakeEndnote(attrs, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(attrs).Cons(SymEndnote)
}

// MakeMark builds a mark note.
func MakeMark(mark string, slug, fragment string, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(sx.MakeString(fragment)).Cons(sx.MakeString(slug)).Cons(sx.MakeString(mark)).Cons(SymMark)
}

// MakeFormat builds an inline formatting node.
func MakeFormat(sym *sx.Symbol, attrs, inlines *sx.Pair) *sx.Pair {
	return inlines.Cons(attrs).Cons(sym)
}

// MakeLiteral builds a inline node with literal text.
func MakeLiteral(sym *sx.Symbol, attrs *sx.Pair, text string) *sx.Pair {
	return sx.Cons(sym, sx.Cons(attrs, sx.Cons(sx.MakeString(text), sx.Nil())))
}
