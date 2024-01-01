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

import "zettelstore.de/sx.fossil"

// Various constants for Zettel data. Some of them are technically variables.

const (
	// Symbols for Metanodes
	SymBlock  = sx.Symbol("BLOCK")
	SymInline = sx.Symbol("INLINE")
	SymMeta   = sx.Symbol("META")

	// Symbols for Zettel node types.
	SymBLOB            = sx.Symbol("BLOB")
	SymCell            = sx.Symbol("CELL")
	SymCellCenter      = sx.Symbol("CELL-CENTER")
	SymCellLeft        = sx.Symbol("CELL-LEFT")
	SymCellRight       = sx.Symbol("CELL-RIGHT")
	SymCite            = sx.Symbol("CITE")
	SymDescription     = sx.Symbol("DESCRIPTION")
	SymEmbed           = sx.Symbol("EMBED")
	SymEmbedBLOB       = sx.Symbol("EMBED-BLOB")
	SymEndnote         = sx.Symbol("ENDNOTE")
	SymFormatEmph      = sx.Symbol("FORMAT-EMPH")
	SymFormatDelete    = sx.Symbol("FORMAT-DELETE")
	SymFormatInsert    = sx.Symbol("FORMAT-INSERT")
	SymFormatMark      = sx.Symbol("FORMAT-MARK")
	SymFormatQuote     = sx.Symbol("FORMAT-QUOTE")
	SymFormatSpan      = sx.Symbol("FORMAT-SPAN")
	SymFormatSub       = sx.Symbol("FORMAT-SUB")
	SymFormatSuper     = sx.Symbol("FORMAT-SUPER")
	SymFormatStrong    = sx.Symbol("FORMAT-STRONG")
	SymHard            = sx.Symbol("HARD")
	SymHeading         = sx.Symbol("HEADING")
	SymLinkInvalid     = sx.Symbol("LINK-INVALID")
	SymLinkZettel      = sx.Symbol("LINK-ZETTEL")
	SymLinkSelf        = sx.Symbol("LINK-SELF")
	SymLinkFound       = sx.Symbol("LINK-FOUND")
	SymLinkBroken      = sx.Symbol("LINK-BROKEN")
	SymLinkHosted      = sx.Symbol("LINK-HOSTED")
	SymLinkBased       = sx.Symbol("LINK-BASED")
	SymLinkQuery       = sx.Symbol("LINK-QUERY")
	SymLinkExternal    = sx.Symbol("LINK-EXTERNAL")
	SymListOrdered     = sx.Symbol("ORDERED")
	SymListUnordered   = sx.Symbol("UNORDERED")
	SymListQuote       = sx.Symbol("QUOTATION")
	SymLiteralProg     = sx.Symbol("LITERAL-CODE")
	SymLiteralComment  = sx.Symbol("LITERAL-COMMENT")
	SymLiteralHTML     = sx.Symbol("LITERAL-HTML")
	SymLiteralInput    = sx.Symbol("LITERAL-INPUT")
	SymLiteralMath     = sx.Symbol("LITERAL-MATH")
	SymLiteralOutput   = sx.Symbol("LITERAL-OUTPUT")
	SymLiteralZettel   = sx.Symbol("LITERAL-ZETTEL")
	SymMark            = sx.Symbol("MARK")
	SymPara            = sx.Symbol("PARA")
	SymRegionBlock     = sx.Symbol("REGION-BLOCK")
	SymRegionQuote     = sx.Symbol("REGION-QUOTE")
	SymRegionVerse     = sx.Symbol("REGION-VERSE")
	SymSoft            = sx.Symbol("SOFT")
	SymSpace           = sx.Symbol("SPACE")
	SymTable           = sx.Symbol("TABLE")
	SymText            = sx.Symbol("TEXT")
	SymThematic        = sx.Symbol("THEMATIC")
	SymTransclude      = sx.Symbol("TRANSCLUDE")
	SymUnknown         = sx.Symbol("UNKNOWN-NODE")
	SymVerbatimComment = sx.Symbol("VERBATIM-COMMENT")
	SymVerbatimEval    = sx.Symbol("VERBATIM-EVAL")
	SymVerbatimHTML    = sx.Symbol("VERBATIM-HTML")
	SymVerbatimMath    = sx.Symbol("VERBATIM-MATH")
	SymVerbatimProg    = sx.Symbol("VERBATIM-CODE")
	SymVerbatimZettel  = sx.Symbol("VERBATIM-ZETTEL")

	// Constant symbols for reference states.
	SymRefStateInvalid  = sx.Symbol("INVALID")
	SymRefStateZettel   = sx.Symbol("ZETTEL")
	SymRefStateSelf     = sx.Symbol("SELF")
	SymRefStateFound    = sx.Symbol("FOUND")
	SymRefStateBroken   = sx.Symbol("BROKEN")
	SymRefStateHosted   = sx.Symbol("HOSTED")
	SymRefStateBased    = sx.Symbol("BASED")
	SymRefStateQuery    = sx.Symbol("QUERY")
	SymRefStateExternal = sx.Symbol("EXTERNAL")

	// Symbols for metadata types.
	SymTypeCredential   = sx.Symbol("CREDENTIAL")
	SymTypeEmpty        = sx.Symbol("EMPTY-STRING")
	SymTypeID           = sx.Symbol("ZID")
	SymTypeIDSet        = sx.Symbol("ZID-SET")
	SymTypeNumber       = sx.Symbol("NUMBER")
	SymTypeString       = sx.Symbol("STRING")
	SymTypeTagSet       = sx.Symbol("TAG-SET")
	SymTypeTimestamp    = sx.Symbol("TIMESTAMP")
	SymTypeURL          = sx.Symbol("URL")
	SymTypeWord         = sx.Symbol("WORD")
	SymTypeWordSet      = sx.Symbol("WORD-SET")
	SymTypeZettelmarkup = sx.Symbol("ZETTELMARKUP")
)
