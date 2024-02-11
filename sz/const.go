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
	NameBlock  = "BLOCK"
	NameInline = "INLINE"
	NameMeta   = "META"

	// Symbols for Zettel noMakede types.
	NameBLOB            = "BLOB"
	NameCell            = "CELL"
	NameCellCenter      = "CELL-CENTER"
	NameCellLeft        = "CELL-LEFT"
	NameCellRight       = "CELL-RIGHT"
	NameCite            = "CITE"
	NameDescription     = "DESCRIPTION"
	NameEmbed           = "EMBED"
	NameEmbedBLOB       = "EMBED-BLOB"
	NameEndnote         = "ENDNOTE"
	NameFormatEmph      = "FORMAT-EMPH"
	NameFormatDelete    = "FORMAT-DELETE"
	NameFormatInsert    = "FORMAT-INSERT"
	NameFormatMark      = "FORMAT-MARK"
	NameFormatQuote     = "FORMAT-QUOTE"
	NameFormatSpan      = "FORMAT-SPAN"
	NameFormatSub       = "FORMAT-SUB"
	NameFormatSuper     = "FORMAT-SUPER"
	NameFormatStrong    = "FORMAT-STRONG"
	NameHard            = "HARD"
	NameHeading         = "HEADING"
	NameLinkInvalid     = "LINK-INVALID"
	NameLinkZettel      = "LINK-ZETTEL"
	NameLinkSelf        = "LINK-SELF"
	NameLinkFound       = "LINK-FOUND"
	NameLinkBroken      = "LINK-BROKEN"
	NameLinkHosted      = "LINK-HOSTED"
	NameLinkBased       = "LINK-BASED"
	NameLinkQuery       = "LINK-QUERY"
	NameLinkExternal    = "LINK-EXTERNAL"
	NameListOrdered     = "ORDERED"
	NameListUnordered   = "UNORDERED"
	NameListQuote       = "QUOTATION"
	NameLiteralProg     = "LITERAL-CODE"
	NameLiteralComment  = "LITERAL-COMMENT"
	NameLiteralHTML     = "LITERAL-HTML"
	NameLiteralInput    = "LITERAL-INPUT"
	NameLiteralMath     = "LITERAL-MATH"
	NameLiteralOutput   = "LITERAL-OUTPUT"
	NameLiteralZettel   = "LITERAL-ZETTEL"
	NameMark            = "MARK"
	NamePara            = "PARA"
	NameRegionBlock     = "REGION-BLOCK"
	NameRegionQuote     = "REGION-QUOTE"
	NameRegionVerse     = "REGION-VERSE"
	NameSoft            = "SOFT"
	NameSpace           = "SPACE"
	NameTable           = "TABLE"
	NameText            = "TEXT"
	NameThematic        = "THEMATIC"
	NameTransclude      = "TRANSCLUDE"
	NameUnknown         = "UNKNOWN-NODE"
	NameVerbatimComment = "VERBATIM-COMMENT"
	NameVerbatimEval    = "VERBATIM-EVAL"
	NameVerbatimHTML    = "VERBATIM-HTML"
	NameVerbatimMath    = "VERBATIM-MATH"
	NameVerbatimProg    = "VERBATIM-CODE"
	NameVerbatimZettel  = "VERBATIM-ZETTEL"

	// Constant symbols for reference states.
	NameRefStateInvalid  = "INVALID"
	NameRefStateZettel   = "ZETTEL"
	NameRefStateSelf     = "SELF"
	NameRefStateFound    = "FOUND"
	NameRefStateBroken   = "BROKEN"
	NameRefStateHosted   = "HOSTED"
	NameRefStateBased    = "BASED"
	NameRefStateQuery    = "QUERY"
	NameRefStateExternal = "EXTERNAL"

	// Symbols for metadata types.
	NameTypeCredential   = "CREDENTIAL"
	NameTypeEmpty        = "EMPTY-STRING"
	NameTypeID           = "ZID"
	NameTypeIDSet        = "ZID-SET"
	NameTypeNumber       = "NUMBER"
	NameTypeString       = "STRING"
	NameTypeTagSet       = "TAG-SET"
	NameTypeTimestamp    = "TIMESTAMP"
	NameTypeURL          = "URL"
	NameTypeWord         = "WORD"
	NameTypeWordSet      = "WORD-SET"
	NameTypeZettelmarkup = "ZETTELMARKUP"
)

var (
	// Symbols for Metanodes
	SymBlock  = sx.MakeSymbol(NameBlock)
	SymInline = sx.MakeSymbol(NameInline)
	SymMeta   = sx.MakeSymbol(NameMeta)

	// Symbols for Zettel noMakede types.
	SymBLOB            = sx.MakeSymbol(NameBLOB)
	SymCell            = sx.MakeSymbol(NameCell)
	SymCellCenter      = sx.MakeSymbol(NameCellCenter)
	SymCellLeft        = sx.MakeSymbol(NameCellLeft)
	SymCellRight       = sx.MakeSymbol(NameCellRight)
	SymCite            = sx.MakeSymbol(NameCite)
	SymDescription     = sx.MakeSymbol(NameDescription)
	SymEmbed           = sx.MakeSymbol(NameEmbed)
	SymEmbedBLOB       = sx.MakeSymbol(NameEmbedBLOB)
	SymEndnote         = sx.MakeSymbol(NameEndnote)
	SymFormatEmph      = sx.MakeSymbol(NameFormatEmph)
	SymFormatDelete    = sx.MakeSymbol(NameFormatDelete)
	SymFormatInsert    = sx.MakeSymbol(NameFormatInsert)
	SymFormatMark      = sx.MakeSymbol(NameFormatMark)
	SymFormatQuote     = sx.MakeSymbol(NameFormatQuote)
	SymFormatSpan      = sx.MakeSymbol(NameFormatSpan)
	SymFormatSub       = sx.MakeSymbol(NameFormatSub)
	SymFormatSuper     = sx.MakeSymbol(NameFormatSuper)
	SymFormatStrong    = sx.MakeSymbol(NameFormatStrong)
	SymHard            = sx.MakeSymbol(NameHard)
	SymHeading         = sx.MakeSymbol(NameHeading)
	SymLinkInvalid     = sx.MakeSymbol(NameLinkInvalid)
	SymLinkZettel      = sx.MakeSymbol(NameLinkZettel)
	SymLinkSelf        = sx.MakeSymbol(NameLinkSelf)
	SymLinkFound       = sx.MakeSymbol(NameLinkFound)
	SymLinkBroken      = sx.MakeSymbol(NameLinkBroken)
	SymLinkHosted      = sx.MakeSymbol(NameLinkHosted)
	SymLinkBased       = sx.MakeSymbol(NameLinkBased)
	SymLinkQuery       = sx.MakeSymbol(NameLinkQuery)
	SymLinkExternal    = sx.MakeSymbol(NameLinkExternal)
	SymListOrdered     = sx.MakeSymbol(NameListOrdered)
	SymListUnordered   = sx.MakeSymbol(NameListUnordered)
	SymListQuote       = sx.MakeSymbol(NameListQuote)
	SymLiteralProg     = sx.MakeSymbol(NameLiteralProg)
	SymLiteralComment  = sx.MakeSymbol(NameLiteralComment)
	SymLiteralHTML     = sx.MakeSymbol(NameLiteralHTML)
	SymLiteralInput    = sx.MakeSymbol(NameLiteralInput)
	SymLiteralMath     = sx.MakeSymbol(NameLiteralMath)
	SymLiteralOutput   = sx.MakeSymbol(NameLiteralOutput)
	SymLiteralZettel   = sx.MakeSymbol(NameLiteralZettel)
	SymMark            = sx.MakeSymbol(NameMark)
	SymPara            = sx.MakeSymbol(NamePara)
	SymRegionBlock     = sx.MakeSymbol(NameRegionBlock)
	SymRegionQuote     = sx.MakeSymbol(NameRegionQuote)
	SymRegionVerse     = sx.MakeSymbol(NameRegionVerse)
	SymSoft            = sx.MakeSymbol(NameSoft)
	SymSpace           = sx.MakeSymbol(NameSpace)
	SymTable           = sx.MakeSymbol(NameTable)
	SymText            = sx.MakeSymbol(NameText)
	SymThematic        = sx.MakeSymbol(NameThematic)
	SymTransclude      = sx.MakeSymbol(NameTransclude)
	SymUnknown         = sx.MakeSymbol(NameUnknown)
	SymVerbatimComment = sx.MakeSymbol(NameVerbatimComment)
	SymVerbatimEval    = sx.MakeSymbol(NameVerbatimEval)
	SymVerbatimHTML    = sx.MakeSymbol(NameVerbatimHTML)
	SymVerbatimMath    = sx.MakeSymbol(NameVerbatimMath)
	SymVerbatimProg    = sx.MakeSymbol(NameVerbatimProg)
	SymVerbatimZettel  = sx.MakeSymbol(NameVerbatimZettel)

	// Constant symbols for reference states.
	SymRefStateInvalid  = sx.MakeSymbol(NameRefStateInvalid)
	SymRefStateZettel   = sx.MakeSymbol(NameRefStateZettel)
	SymRefStateSelf     = sx.MakeSymbol(NameRefStateSelf)
	SymRefStateFound    = sx.MakeSymbol(NameRefStateFound)
	SymRefStateBroken   = sx.MakeSymbol(NameRefStateBroken)
	SymRefStateHosted   = sx.MakeSymbol(NameRefStateHosted)
	SymRefStateBased    = sx.MakeSymbol(NameRefStateBased)
	SymRefStateQuery    = sx.MakeSymbol(NameRefStateQuery)
	SymRefStateExternal = sx.MakeSymbol(NameRefStateExternal)

	// Symbols for metadata types.
	SymTypeCredential   = sx.MakeSymbol(NameTypeCredential)
	SymTypeEmpty        = sx.MakeSymbol(NameTypeEmpty)
	SymTypeID           = sx.MakeSymbol(NameTypeID)
	SymTypeIDSet        = sx.MakeSymbol(NameTypeIDSet)
	SymTypeNumber       = sx.MakeSymbol(NameTypeNumber)
	SymTypeString       = sx.MakeSymbol(NameTypeString)
	SymTypeTagSet       = sx.MakeSymbol(NameTypeTagSet)
	SymTypeTimestamp    = sx.MakeSymbol(NameTypeTimestamp)
	SymTypeURL          = sx.MakeSymbol(NameTypeURL)
	SymTypeWord         = sx.MakeSymbol(NameTypeWord)
	SymTypeWordSet      = sx.MakeSymbol(NameTypeWordSet)
	SymTypeZettelmarkup = sx.MakeSymbol(NameTypeZettelmarkup)
)
