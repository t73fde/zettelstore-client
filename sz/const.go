//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import "t73f.de/r/sx"

// Various constants for Zettel data. They are technically variables.
var (
	// Symbols for Metanodes
	SymBlock  = sx.MakeSymbol("BLOCK")
	SymInline = sx.MakeSymbol("INLINE")
	SymMeta   = sx.MakeSymbol("META")

	// Symbols for Zettel noMakede types.
	SymBLOB            = sx.MakeSymbol("BLOB")
	SymCell            = sx.MakeSymbol("CELL")
	SymCellCenter      = sx.MakeSymbol("CELL-CENTER")
	SymCellLeft        = sx.MakeSymbol("CELL-LEFT")
	SymCellRight       = sx.MakeSymbol("CELL-RIGHT")
	SymCite            = sx.MakeSymbol("CITE")
	SymDescription     = sx.MakeSymbol("DESCRIPTION")
	SymEmbed           = sx.MakeSymbol("EMBED")
	SymEmbedBLOB       = sx.MakeSymbol("EMBED-BLOB")
	SymEndnote         = sx.MakeSymbol("ENDNOTE")
	SymFormatEmph      = sx.MakeSymbol("FORMAT-EMPH")
	SymFormatDelete    = sx.MakeSymbol("FORMAT-DELETE")
	SymFormatInsert    = sx.MakeSymbol("FORMAT-INSERT")
	SymFormatMark      = sx.MakeSymbol("FORMAT-MARK")
	SymFormatQuote     = sx.MakeSymbol("FORMAT-QUOTE")
	SymFormatSpan      = sx.MakeSymbol("FORMAT-SPAN")
	SymFormatSub       = sx.MakeSymbol("FORMAT-SUB")
	SymFormatSuper     = sx.MakeSymbol("FORMAT-SUPER")
	SymFormatStrong    = sx.MakeSymbol("FORMAT-STRONG")
	SymHard            = sx.MakeSymbol("HARD")
	SymHeading         = sx.MakeSymbol("HEADING")
	SymLinkInvalid     = sx.MakeSymbol("LINK-INVALID")
	SymLinkZettel      = sx.MakeSymbol("LINK-ZETTEL")
	SymLinkSelf        = sx.MakeSymbol("LINK-SELF")
	SymLinkFound       = sx.MakeSymbol("LINK-FOUND")
	SymLinkBroken      = sx.MakeSymbol("LINK-BROKEN")
	SymLinkHosted      = sx.MakeSymbol("LINK-HOSTED")
	SymLinkBased       = sx.MakeSymbol("LINK-BASED")
	SymLinkQuery       = sx.MakeSymbol("LINK-QUERY")
	SymLinkExternal    = sx.MakeSymbol("LINK-EXTERNAL")
	SymListOrdered     = sx.MakeSymbol("ORDERED")
	SymListUnordered   = sx.MakeSymbol("UNORDERED")
	SymListQuote       = sx.MakeSymbol("QUOTATION")
	SymLiteralProg     = sx.MakeSymbol("LITERAL-CODE")
	SymLiteralComment  = sx.MakeSymbol("LITERAL-COMMENT")
	SymLiteralHTML     = sx.MakeSymbol("LITERAL-HTML")
	SymLiteralInput    = sx.MakeSymbol("LITERAL-INPUT")
	SymLiteralMath     = sx.MakeSymbol("LITERAL-MATH")
	SymLiteralOutput   = sx.MakeSymbol("LITERAL-OUTPUT")
	SymLiteralZettel   = sx.MakeSymbol("LITERAL-ZETTEL")
	SymMark            = sx.MakeSymbol("MARK")
	SymPara            = sx.MakeSymbol("PARA")
	SymRegionBlock     = sx.MakeSymbol("REGION-BLOCK")
	SymRegionQuote     = sx.MakeSymbol("REGION-QUOTE")
	SymRegionVerse     = sx.MakeSymbol("REGION-VERSE")
	SymSoft            = sx.MakeSymbol("SOFT")
	SymTable           = sx.MakeSymbol("TABLE")
	SymText            = sx.MakeSymbol("TEXT")
	SymThematic        = sx.MakeSymbol("THEMATIC")
	SymTransclude      = sx.MakeSymbol("TRANSCLUDE")
	SymUnknown         = sx.MakeSymbol("UNKNOWN-NODE")
	SymVerbatimComment = sx.MakeSymbol("VERBATIM-COMMENT")
	SymVerbatimEval    = sx.MakeSymbol("VERBATIM-EVAL")
	SymVerbatimHTML    = sx.MakeSymbol("VERBATIM-HTML")
	SymVerbatimMath    = sx.MakeSymbol("VERBATIM-MATH")
	SymVerbatimProg    = sx.MakeSymbol("VERBATIM-CODE")
	SymVerbatimZettel  = sx.MakeSymbol("VERBATIM-ZETTEL")

	// Constant symbols for reference states.
	SymRefStateInvalid  = sx.MakeSymbol("INVALID")
	SymRefStateZettel   = sx.MakeSymbol("ZETTEL")
	SymRefStateSelf     = sx.MakeSymbol("SELF")
	SymRefStateFound    = sx.MakeSymbol("FOUND")
	SymRefStateBroken   = sx.MakeSymbol("BROKEN")
	SymRefStateHosted   = sx.MakeSymbol("HOSTED")
	SymRefStateBased    = sx.MakeSymbol("BASED")
	SymRefStateQuery    = sx.MakeSymbol("QUERY")
	SymRefStateExternal = sx.MakeSymbol("EXTERNAL")

	// Symbols for metadata types.
	SymTypeCredential   = sx.MakeSymbol("CREDENTIAL")
	SymTypeEmpty        = sx.MakeSymbol("EMPTY-STRING")
	SymTypeID           = sx.MakeSymbol("ZID")
	SymTypeIDSet        = sx.MakeSymbol("ZID-SET")
	SymTypeNumber       = sx.MakeSymbol("NUMBER")
	SymTypeString       = sx.MakeSymbol("STRING")
	SymTypeTagSet       = sx.MakeSymbol("TAG-SET")
	SymTypeTimestamp    = sx.MakeSymbol("TIMESTAMP")
	SymTypeURL          = sx.MakeSymbol("URL")
	SymTypeWord         = sx.MakeSymbol("WORD")
	SymTypeZettelmarkup = sx.MakeSymbol("ZETTELMARKUP")
)
