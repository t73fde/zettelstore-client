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

import "zettelstore.de/sx.fossil"

// Constant symbols for HTML header tags
const (
	SymBody   = sx.Symbol("body")
	SymHead   = sx.Symbol("head")
	SymHtml   = sx.Symbol("html")
	SymMeta   = sx.Symbol("meta")
	SymScript = sx.Symbol("script")
	SymTitle  = sx.Symbol("title")
)

// Constant symbols for HTML body tags
const (
	SymA          = sx.Symbol("a")
	SymASIDE      = sx.Symbol("aside")
	symBLOCKQUOTE = sx.Symbol("blockquote")
	symBR         = sx.Symbol("br")
	symCITE       = sx.Symbol("cite")
	symCODE       = sx.Symbol("code")
	symDD         = sx.Symbol("dd")
	symDEL        = sx.Symbol("del")
	SymDIV        = sx.Symbol("div")
	symDL         = sx.Symbol("dl")
	symDT         = sx.Symbol("dt")
	symEM         = sx.Symbol("em")
	SymEMBED      = sx.Symbol("embed")
	SymFIGURE     = sx.Symbol("figure")
	SymH1         = sx.Symbol("h1")
	SymH2         = sx.Symbol("h2")
	SymHR         = sx.Symbol("hr")
	SymIMG        = sx.Symbol("img")
	symINS        = sx.Symbol("ins")
	symKBD        = sx.Symbol("kbd")
	SymLI         = sx.Symbol("li")
	symMARK       = sx.Symbol("mark")
	SymOL         = sx.Symbol("ol")
	SymP          = sx.Symbol("p")
	symPRE        = sx.Symbol("pre")
	symSAMP       = sx.Symbol("samp")
	SymSPAN       = sx.Symbol("span")
	SymSTRONG     = sx.Symbol("strong")
	symSUB        = sx.Symbol("sub")
	symSUP        = sx.Symbol("sup")
	symTABLE      = sx.Symbol("table")
	symTBODY      = sx.Symbol("tbody")
	symTHEAD      = sx.Symbol("thead")
	symTD         = sx.Symbol("td")
	symTR         = sx.Symbol("tr")
	SymUL         = sx.Symbol("ul")
)

// Constant symbols for HTML attribute keys
const (
	symAttrAlt    = sx.Symbol("alt")
	SymAttrClass  = sx.Symbol("class")
	SymAttrHref   = sx.Symbol("href")
	SymAttrId     = sx.Symbol("id")
	SymAttrLang   = sx.Symbol("lang")
	SymAttrOpen   = sx.Symbol("open")
	SymAttrRel    = sx.Symbol("rel")
	SymAttrRole   = sx.Symbol("role")
	SymAttrSrc    = sx.Symbol("src")
	SymAttrTarget = sx.Symbol("target")
	SymAttrTitle  = sx.Symbol("title")
	SymAttrType   = sx.Symbol("type")
	SymAttrValue  = sx.Symbol("value")
)
