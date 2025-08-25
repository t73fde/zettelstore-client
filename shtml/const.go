//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package shtml

import "t73f.de/r/sxwebs/sxhtml"

// Symbols for HTML header tags
var (
	SymBody   = sxhtml.MakeSymbol("body")
	SymHead   = sxhtml.MakeSymbol("head")
	SymHTML   = sxhtml.MakeSymbol("html")
	SymMeta   = sxhtml.MakeSymbol("meta")
	SymScript = sxhtml.MakeSymbol("script")
	SymTitle  = sxhtml.MakeSymbol("title")
)

// Symbols for HTML body tags
var (
	SymA          = sxhtml.MakeSymbol("a")
	SymASIDE      = sxhtml.MakeSymbol("aside")
	symBLOCKQUOTE = sxhtml.MakeSymbol("blockquote")
	symBR         = sxhtml.MakeSymbol("br")
	symCITE       = sxhtml.MakeSymbol("cite")
	symCODE       = sxhtml.MakeSymbol("code")
	symDD         = sxhtml.MakeSymbol("dd")
	symDEL        = sxhtml.MakeSymbol("del")
	SymDIV        = sxhtml.MakeSymbol("div")
	symDL         = sxhtml.MakeSymbol("dl")
	symDT         = sxhtml.MakeSymbol("dt")
	symEM         = sxhtml.MakeSymbol("em")
	SymEMBED      = sxhtml.MakeSymbol("embed")
	SymFIGURE     = sxhtml.MakeSymbol("figure")
	SymH1         = sxhtml.MakeSymbol("h1")
	SymH2         = sxhtml.MakeSymbol("h2")
	SymHR         = sxhtml.MakeSymbol("hr")
	SymIMG        = sxhtml.MakeSymbol("img")
	symINS        = sxhtml.MakeSymbol("ins")
	symKBD        = sxhtml.MakeSymbol("kbd")
	SymLI         = sxhtml.MakeSymbol("li")
	symMARK       = sxhtml.MakeSymbol("mark")
	SymOL         = sxhtml.MakeSymbol("ol")
	SymP          = sxhtml.MakeSymbol("p")
	symPRE        = sxhtml.MakeSymbol("pre")
	symSAMP       = sxhtml.MakeSymbol("samp")
	SymSPAN       = sxhtml.MakeSymbol("span")
	SymSTRONG     = sxhtml.MakeSymbol("strong")
	symSUB        = sxhtml.MakeSymbol("sub")
	symSUP        = sxhtml.MakeSymbol("sup")
	symTABLE      = sxhtml.MakeSymbol("table")
	symTBODY      = sxhtml.MakeSymbol("tbody")
	symTHEAD      = sxhtml.MakeSymbol("thead")
	symTD         = sxhtml.MakeSymbol("td")
	symTH         = sxhtml.MakeSymbol("th")
	symTR         = sxhtml.MakeSymbol("tr")
	SymUL         = sxhtml.MakeSymbol("ul")
)

// Symbols for HTML attribute keys
var (
	SymAttrClass  = sxhtml.MakeSymbol("class")
	SymAttrHref   = sxhtml.MakeSymbol("href")
	SymAttrID     = sxhtml.MakeSymbol("id")
	SymAttrLang   = sxhtml.MakeSymbol("lang")
	SymAttrOpen   = sxhtml.MakeSymbol("open")
	SymAttrRel    = sxhtml.MakeSymbol("rel")
	SymAttrRole   = sxhtml.MakeSymbol("role")
	SymAttrSrc    = sxhtml.MakeSymbol("src")
	SymAttrTarget = sxhtml.MakeSymbol("target")
	SymAttrTitle  = sxhtml.MakeSymbol("title")
	SymAttrType   = sxhtml.MakeSymbol("type")
	SymAttrValue  = sxhtml.MakeSymbol("value")
)
