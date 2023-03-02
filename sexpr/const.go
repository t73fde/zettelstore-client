//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr

import "codeberg.org/t73fde/sxpf"

// Various constants for Zettel data. Some of them are technically variables.

const (
	// Symbols for Metanodes
	NameSymBlock  = "BLOCK"
	NameSymInline = "INLINE"
	NameSymList   = "LIST"
	NameSymMeta   = "META"
	NameSymQuote  = "quote"

	// Symbols for Zettel node types.
	NameSymBLOB            = "BLOB"
	NameSymCell            = "CELL"
	NameSymCellCenter      = "CELL-CENTER"
	NameSymCellLeft        = "CELL-LEFT"
	NameSymCellRight       = "CELL-RIGHT"
	NameSymCite            = "CITE"
	NameSymDescription     = "DESCRIPTION"
	NameSymEmbed           = "EMBED"
	NameSymEmbedBLOB       = "EMBED-BLOB"
	NameSymEndnote         = "ENDNOTE"
	NameSymFormatEmph      = "FORMAT-EMPH"
	NameSymFormatDelete    = "FORMAT-DELETE"
	NameSymFormatInsert    = "FORMAT-INSERT"
	NameSymFormatQuote     = "FORMAT-QUOTE"
	NameSymFormatSpan      = "FORMAT-SPAN"
	NameSymFormatSub       = "FORMAT-SUB"
	NameSymFormatSuper     = "FORMAT-SUPER"
	NameSymFormatStrong    = "FORMAT-STRONG"
	NameSymHard            = "HARD"
	NameSymHeading         = "HEADING"
	NameSymLinkInvalid     = "LINK-INVALID"
	NameSymLinkZettel      = "LINK-ZETTEL"
	NameSymLinkSelf        = "LINK-SELF"
	NameSymLinkFound       = "LINK-FOUND"
	NameSymLinkBroken      = "LINK-BROKEN"
	NameSymLinkHosted      = "LINK-HOSTED"
	NameSymLinkBased       = "LINK-BASED"
	NameSymLinkQuery       = "LINK-QUERY"
	NameSymLinkExternal    = "LINK-EXTERNAL"
	NameSymListOrdered     = "ORDERED"
	NameSymListUnordered   = "UNORDERED"
	NameSymListQuote       = "QUOTATION"
	NameSymLiteralProg     = "LITERAL-CODE"
	NameSymLiteralComment  = "LITERAL-COMMENT"
	NameSymLiteralHTML     = "LITERAL-HTML"
	NameSymLiteralInput    = "LITERAL-INPUT"
	NameSymLiteralMath     = "LITERAL-MATH"
	NameSymLiteralOutput   = "LITERAL-OUTPUT"
	NameSymLiteralZettel   = "LITERAL-ZETTEL"
	NameSymMark            = "MARK"
	NameSymPara            = "PARA"
	NameSymRegionBlock     = "REGION-BLOCK"
	NameSymRegionQuote     = "REGION-QUOTE"
	NameSymRegionVerse     = "REGION-VERSE"
	NameSymSoft            = "SOFT"
	NameSymSpace           = "SPACE"
	NameSymTable           = "TABLE"
	NameSymText            = "TEXT"
	NameSymThematic        = "THEMATIC"
	NameSymTransclude      = "TRANSCLUDE"
	NameSymUnknown         = "UNKNOWN-NODE"
	NameSymVerbatimComment = "VERBATIM-COMMENT"
	NameSymVerbatimEval    = "VERBATIM-EVAL"
	NameSymVerbatimHTML    = "VERBATIM-HTML"
	NameSymVerbatimMath    = "VERBATIM-MATH"
	NameSymVerbatimProg    = "VERBATIM-CODE"
	NameSymVerbatimZettel  = "VERBATIM-ZETTEL"

	// Constant symbols for reference states.
	NameSymRefStateInvalid  = "INVALID"
	NameSymRefStateZettel   = "ZETTEL"
	NameSymRefStateSelf     = "SELF"
	NameSymRefStateFound    = "FOUND"
	NameSymRefStateBroken   = "BROKEN"
	NameSymRefStateHosted   = "HOSTED"
	NameSymRefStateBased    = "BASED"
	NameSymRefStateQuery    = "QUERY"
	NameSymRefStateExternal = "EXTERNAL"

	// Symbols for metadata types.
	NameSymTypeCredential   = "CREDENTIAL"
	NameSymTypeEmpty        = "EMPTY-STRING"
	NameSymTypeID           = "ZID"
	NameSymTypeIDSet        = "ZID-SET"
	NameSymTypeNumber       = "NUMBER"
	NameSymTypeString       = "STRING"
	NameSymTypeTagSet       = "TAG-SET"
	NameSymTypeTimestamp    = "TIMESTAMP"
	NameSymTypeURL          = "URL"
	NameSymTypeWord         = "WORD"
	NameSymTypeWordSet      = "WORD-SET"
	NameSymTypeZettelmarkup = "ZETTELMARKUP"
)

// ZettelSymbols collect all symbols needed to represent zettel data.
type ZettelSymbols struct {
	// Symbols for Metanodes
	SymBlock  *sxpf.Symbol
	SymInline *sxpf.Symbol
	SymList   *sxpf.Symbol
	SymMeta   *sxpf.Symbol
	SymQuote  *sxpf.Symbol

	// Symbols for Zettel node types.
	SymBLOB            *sxpf.Symbol
	SymCell            *sxpf.Symbol
	SymCellCenter      *sxpf.Symbol
	SymCellLeft        *sxpf.Symbol
	SymCellRight       *sxpf.Symbol
	SymCite            *sxpf.Symbol
	SymDescription     *sxpf.Symbol
	SymEmbed           *sxpf.Symbol
	SymEmbedBLOB       *sxpf.Symbol
	SymEndnote         *sxpf.Symbol
	SymFormatEmph      *sxpf.Symbol
	SymFormatDelete    *sxpf.Symbol
	SymFormatInsert    *sxpf.Symbol
	SymFormatQuote     *sxpf.Symbol
	SymFormatSpan      *sxpf.Symbol
	SymFormatSub       *sxpf.Symbol
	SymFormatSuper     *sxpf.Symbol
	SymFormatStrong    *sxpf.Symbol
	SymHard            *sxpf.Symbol
	SymHeading         *sxpf.Symbol
	SymLinkInvalid     *sxpf.Symbol
	SymLinkZettel      *sxpf.Symbol
	SymLinkSelf        *sxpf.Symbol
	SymLinkFound       *sxpf.Symbol
	SymLinkBroken      *sxpf.Symbol
	SymLinkHosted      *sxpf.Symbol
	SymLinkBased       *sxpf.Symbol
	SymLinkQuery       *sxpf.Symbol
	SymLinkExternal    *sxpf.Symbol
	SymListOrdered     *sxpf.Symbol
	SymListUnordered   *sxpf.Symbol
	SymListQuote       *sxpf.Symbol
	SymLiteralProg     *sxpf.Symbol
	SymLiteralComment  *sxpf.Symbol
	SymLiteralHTML     *sxpf.Symbol
	SymLiteralInput    *sxpf.Symbol
	SymLiteralMath     *sxpf.Symbol
	SymLiteralOutput   *sxpf.Symbol
	SymLiteralZettel   *sxpf.Symbol
	SymMark            *sxpf.Symbol
	SymPara            *sxpf.Symbol
	SymRegionBlock     *sxpf.Symbol
	SymRegionQuote     *sxpf.Symbol
	SymRegionVerse     *sxpf.Symbol
	SymSoft            *sxpf.Symbol
	SymSpace           *sxpf.Symbol
	SymTable           *sxpf.Symbol
	SymText            *sxpf.Symbol
	SymThematic        *sxpf.Symbol
	SymTransclude      *sxpf.Symbol
	SymUnknown         *sxpf.Symbol
	SymVerbatimComment *sxpf.Symbol
	SymVerbatimEval    *sxpf.Symbol
	SymVerbatimHTML    *sxpf.Symbol
	SymVerbatimMath    *sxpf.Symbol
	SymVerbatimProg    *sxpf.Symbol
	SymVerbatimZettel  *sxpf.Symbol

	// Constant symbols for reference states.

	SymRefStateInvalid  *sxpf.Symbol
	SymRefStateZettel   *sxpf.Symbol
	SymRefStateSelf     *sxpf.Symbol
	SymRefStateFound    *sxpf.Symbol
	SymRefStateBroken   *sxpf.Symbol
	SymRefStateHosted   *sxpf.Symbol
	SymRefStateBased    *sxpf.Symbol
	SymRefStateQuery    *sxpf.Symbol
	SymRefStateExternal *sxpf.Symbol

	// Symbols for metadata types

	SymTypeCredential   *sxpf.Symbol
	SymTypeEmpty        *sxpf.Symbol
	SymTypeID           *sxpf.Symbol
	SymTypeIDSet        *sxpf.Symbol
	SymTypeNumber       *sxpf.Symbol
	SymTypeString       *sxpf.Symbol
	SymTypeTagSet       *sxpf.Symbol
	SymTypeTimestamp    *sxpf.Symbol
	SymTypeURL          *sxpf.Symbol
	SymTypeWord         *sxpf.Symbol
	SymTypeWordSet      *sxpf.Symbol
	SymTypeZettelmarkup *sxpf.Symbol
}

func (zs *ZettelSymbols) InitializeZettelSymbols(sf sxpf.SymbolFactory) {
	// Symbols for Metanodes
	zs.SymBlock = sf.MustMake(NameSymBlock)
	zs.SymInline = sf.MustMake(NameSymInline)
	zs.SymList = sf.MustMake(NameSymList)
	zs.SymMeta = sf.MustMake(NameSymMeta)
	zs.SymQuote = sf.MustMake(NameSymQuote)

	// Symbols for Zettel node types.
	zs.SymBLOB = sf.MustMake(NameSymBLOB)
	zs.SymCell = sf.MustMake(NameSymCell)
	zs.SymCellCenter = sf.MustMake(NameSymCellCenter)
	zs.SymCellLeft = sf.MustMake(NameSymCellLeft)
	zs.SymCellRight = sf.MustMake(NameSymCellRight)
	zs.SymCite = sf.MustMake(NameSymCite)
	zs.SymDescription = sf.MustMake(NameSymDescription)
	zs.SymEmbed = sf.MustMake(NameSymEmbed)
	zs.SymEmbedBLOB = sf.MustMake(NameSymEmbedBLOB)
	zs.SymEndnote = sf.MustMake(NameSymEndnote)
	zs.SymFormatEmph = sf.MustMake(NameSymFormatEmph)
	zs.SymFormatDelete = sf.MustMake(NameSymFormatDelete)
	zs.SymFormatInsert = sf.MustMake(NameSymFormatInsert)
	zs.SymFormatQuote = sf.MustMake(NameSymFormatQuote)
	zs.SymFormatSpan = sf.MustMake(NameSymFormatSpan)
	zs.SymFormatSub = sf.MustMake(NameSymFormatSub)
	zs.SymFormatSuper = sf.MustMake(NameSymFormatSuper)
	zs.SymFormatStrong = sf.MustMake(NameSymFormatStrong)
	zs.SymHard = sf.MustMake(NameSymHard)
	zs.SymHeading = sf.MustMake(NameSymHeading)
	zs.SymLinkInvalid = sf.MustMake(NameSymLinkInvalid)
	zs.SymLinkZettel = sf.MustMake(NameSymLinkZettel)
	zs.SymLinkSelf = sf.MustMake(NameSymLinkSelf)
	zs.SymLinkFound = sf.MustMake(NameSymLinkFound)
	zs.SymLinkBroken = sf.MustMake(NameSymLinkBroken)
	zs.SymLinkHosted = sf.MustMake(NameSymLinkHosted)
	zs.SymLinkBased = sf.MustMake(NameSymLinkBased)
	zs.SymLinkQuery = sf.MustMake(NameSymLinkQuery)
	zs.SymLinkExternal = sf.MustMake(NameSymLinkExternal)
	zs.SymListOrdered = sf.MustMake(NameSymListOrdered)
	zs.SymListUnordered = sf.MustMake(NameSymListUnordered)
	zs.SymListQuote = sf.MustMake(NameSymListQuote)
	zs.SymLiteralProg = sf.MustMake(NameSymLiteralProg)
	zs.SymLiteralComment = sf.MustMake(NameSymLiteralComment)
	zs.SymLiteralHTML = sf.MustMake(NameSymLiteralHTML)
	zs.SymLiteralInput = sf.MustMake(NameSymLiteralInput)
	zs.SymLiteralMath = sf.MustMake(NameSymLiteralMath)
	zs.SymLiteralOutput = sf.MustMake(NameSymLiteralOutput)
	zs.SymLiteralZettel = sf.MustMake(NameSymLiteralZettel)
	zs.SymMark = sf.MustMake(NameSymMark)
	zs.SymPara = sf.MustMake(NameSymPara)
	zs.SymRegionBlock = sf.MustMake(NameSymRegionBlock)
	zs.SymRegionQuote = sf.MustMake(NameSymRegionQuote)
	zs.SymRegionVerse = sf.MustMake(NameSymRegionVerse)
	zs.SymSoft = sf.MustMake(NameSymSoft)
	zs.SymSpace = sf.MustMake(NameSymSpace)
	zs.SymTable = sf.MustMake(NameSymTable)
	zs.SymText = sf.MustMake(NameSymText)
	zs.SymThematic = sf.MustMake(NameSymThematic)
	zs.SymTransclude = sf.MustMake(NameSymTransclude)
	zs.SymUnknown = sf.MustMake(NameSymUnknown)
	zs.SymVerbatimComment = sf.MustMake(NameSymVerbatimComment)
	zs.SymVerbatimEval = sf.MustMake(NameSymVerbatimEval)
	zs.SymVerbatimHTML = sf.MustMake(NameSymVerbatimHTML)
	zs.SymVerbatimMath = sf.MustMake(NameSymVerbatimMath)
	zs.SymVerbatimProg = sf.MustMake(NameSymVerbatimProg)
	zs.SymVerbatimZettel = sf.MustMake(NameSymVerbatimZettel)

	// Constant symbols for reference states.
	zs.SymRefStateInvalid = sf.MustMake(NameSymRefStateInvalid)
	zs.SymRefStateZettel = sf.MustMake(NameSymRefStateZettel)
	zs.SymRefStateSelf = sf.MustMake(NameSymRefStateSelf)
	zs.SymRefStateFound = sf.MustMake(NameSymRefStateFound)
	zs.SymRefStateBroken = sf.MustMake(NameSymRefStateBroken)
	zs.SymRefStateHosted = sf.MustMake(NameSymRefStateHosted)
	zs.SymRefStateBased = sf.MustMake(NameSymRefStateBased)
	zs.SymRefStateQuery = sf.MustMake(NameSymRefStateQuery)
	zs.SymRefStateExternal = sf.MustMake(NameSymRefStateExternal)

	// Symbols for metadata types.
	zs.SymTypeCredential = sf.MustMake(NameSymTypeCredential)
	zs.SymTypeEmpty = sf.MustMake(NameSymTypeEmpty)
	zs.SymTypeID = sf.MustMake(NameSymTypeID)
	zs.SymTypeIDSet = sf.MustMake(NameSymTypeIDSet)
	zs.SymTypeNumber = sf.MustMake(NameSymTypeNumber)
	zs.SymTypeString = sf.MustMake(NameSymTypeString)
	zs.SymTypeTagSet = sf.MustMake(NameSymTypeTagSet)
	zs.SymTypeTimestamp = sf.MustMake(NameSymTypeTimestamp)
	zs.SymTypeURL = sf.MustMake(NameSymTypeURL)
	zs.SymTypeWord = sf.MustMake(NameSymTypeWord)
	zs.SymTypeWordSet = sf.MustMake(NameSymTypeWordSet)
	zs.SymTypeZettelmarkup = sf.MustMake(NameSymTypeZettelmarkup)
}
