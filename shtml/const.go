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

import "t73f.de/r/sx"

// Symbols for HTML header tags
var (
	SymBody   = sx.MakeSymbol("body")
	SymHead   = sx.MakeSymbol("head")
	SymHTML   = sx.MakeSymbol("html")
	SymMeta   = sx.MakeSymbol("meta")
	SymScript = sx.MakeSymbol("script")
	SymTitle  = sx.MakeSymbol("title")
)

// Symbols for HTML body tags
var (
	SymA          = sx.MakeSymbol("a")
	SymASIDE      = sx.MakeSymbol("aside")
	symBLOCKQUOTE = sx.MakeSymbol("blockquote")
	symBR         = sx.MakeSymbol("br")
	symCITE       = sx.MakeSymbol("cite")
	symCODE       = sx.MakeSymbol("code")
	symDD         = sx.MakeSymbol("dd")
	symDEL        = sx.MakeSymbol("del")
	SymDIV        = sx.MakeSymbol("div")
	symDL         = sx.MakeSymbol("dl")
	symDT         = sx.MakeSymbol("dt")
	symEM         = sx.MakeSymbol("em")
	SymEMBED      = sx.MakeSymbol("embed")
	SymFIGURE     = sx.MakeSymbol("figure")
	SymH1         = sx.MakeSymbol("h1")
	SymH2         = sx.MakeSymbol("h2")
	SymHR         = sx.MakeSymbol("hr")
	SymIMG        = sx.MakeSymbol("img")
	symINS        = sx.MakeSymbol("ins")
	symKBD        = sx.MakeSymbol("kbd")
	SymLI         = sx.MakeSymbol("li")
	symMARK       = sx.MakeSymbol("mark")
	SymOL         = sx.MakeSymbol("ol")
	SymP          = sx.MakeSymbol("p")
	symPRE        = sx.MakeSymbol("pre")
	symSAMP       = sx.MakeSymbol("samp")
	SymSPAN       = sx.MakeSymbol("span")
	SymSTRONG     = sx.MakeSymbol("strong")
	symSUB        = sx.MakeSymbol("sub")
	symSUP        = sx.MakeSymbol("sup")
	symTABLE      = sx.MakeSymbol("table")
	symTBODY      = sx.MakeSymbol("tbody")
	symTHEAD      = sx.MakeSymbol("thead")
	symTD         = sx.MakeSymbol("td")
	symTH         = sx.MakeSymbol("th")
	symTR         = sx.MakeSymbol("tr")
	SymUL         = sx.MakeSymbol("ul")
)

// Symbols for HTML attribute keys
var (
	symAttrAlt    = sx.MakeSymbol("alt")
	SymAttrClass  = sx.MakeSymbol("class")
	SymAttrHref   = sx.MakeSymbol("href")
	SymAttrID     = sx.MakeSymbol("id")
	SymAttrLang   = sx.MakeSymbol("lang")
	SymAttrOpen   = sx.MakeSymbol("open")
	SymAttrRel    = sx.MakeSymbol("rel")
	SymAttrRole   = sx.MakeSymbol("role")
	SymAttrSrc    = sx.MakeSymbol("src")
	SymAttrTarget = sx.MakeSymbol("target")
	SymAttrTitle  = sx.MakeSymbol("title")
	SymAttrType   = sx.MakeSymbol("type")
	SymAttrValue  = sx.MakeSymbol("value")
)
