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
	NameSymAttr   = "ATTR"
	NameSymBlock  = "BLOCK"
	NameSymInline = "INLINE"

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
	NameSymFootnote        = "FOOTNOTE"
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
	NameSymRow             = "ROW"
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
	SymAttr   *sxpf.Symbol
	SymBlock  *sxpf.Symbol
	SymInline *sxpf.Symbol

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
	SymFootnote        *sxpf.Symbol
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
	SymRow             *sxpf.Symbol
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
	zs.SymAttr = sf.Make(NameSymAttr)
	zs.SymBlock = sf.Make(NameSymBlock)
	zs.SymInline = sf.Make(NameSymInline)

	// Symbols for Zettel node types.
	zs.SymBLOB = sf.Make(NameSymBLOB)
	zs.SymCell = sf.Make(NameSymCell)
	zs.SymCellCenter = sf.Make(NameSymCellCenter)
	zs.SymCellLeft = sf.Make(NameSymCellLeft)
	zs.SymCellRight = sf.Make(NameSymCellRight)
	zs.SymCite = sf.Make(NameSymCite)
	zs.SymDescription = sf.Make(NameSymDescription)
	zs.SymEmbed = sf.Make(NameSymEmbed)
	zs.SymEmbedBLOB = sf.Make(NameSymEmbedBLOB)
	zs.SymFootnote = sf.Make(NameSymFootnote)
	zs.SymFormatEmph = sf.Make(NameSymFormatEmph)
	zs.SymFormatDelete = sf.Make(NameSymFormatDelete)
	zs.SymFormatInsert = sf.Make(NameSymFormatInsert)
	zs.SymFormatQuote = sf.Make(NameSymFormatQuote)
	zs.SymFormatSpan = sf.Make(NameSymFormatSpan)
	zs.SymFormatSub = sf.Make(NameSymFormatSub)
	zs.SymFormatSuper = sf.Make(NameSymFormatSuper)
	zs.SymFormatStrong = sf.Make(NameSymFormatStrong)
	zs.SymHard = sf.Make(NameSymHard)
	zs.SymHeading = sf.Make(NameSymHeading)
	zs.SymLinkInvalid = sf.Make(NameSymLinkInvalid)
	zs.SymLinkZettel = sf.Make(NameSymLinkZettel)
	zs.SymLinkSelf = sf.Make(NameSymLinkSelf)
	zs.SymLinkFound = sf.Make(NameSymLinkFound)
	zs.SymLinkBroken = sf.Make(NameSymLinkBroken)
	zs.SymLinkHosted = sf.Make(NameSymLinkHosted)
	zs.SymLinkBased = sf.Make(NameSymLinkBased)
	zs.SymLinkQuery = sf.Make(NameSymLinkQuery)
	zs.SymLinkExternal = sf.Make(NameSymLinkExternal)
	zs.SymListOrdered = sf.Make(NameSymListOrdered)
	zs.SymListUnordered = sf.Make(NameSymListUnordered)
	zs.SymListQuote = sf.Make(NameSymListQuote)
	zs.SymLiteralProg = sf.Make(NameSymLiteralProg)
	zs.SymLiteralComment = sf.Make(NameSymLiteralComment)
	zs.SymLiteralHTML = sf.Make(NameSymLiteralHTML)
	zs.SymLiteralInput = sf.Make(NameSymLiteralInput)
	zs.SymLiteralMath = sf.Make(NameSymLiteralMath)
	zs.SymLiteralOutput = sf.Make(NameSymLiteralOutput)
	zs.SymLiteralZettel = sf.Make(NameSymLiteralZettel)
	zs.SymMark = sf.Make(NameSymMark)
	zs.SymPara = sf.Make(NameSymPara)
	zs.SymRegionBlock = sf.Make(NameSymRegionBlock)
	zs.SymRegionQuote = sf.Make(NameSymRegionQuote)
	zs.SymRegionVerse = sf.Make(NameSymRegionVerse)
	zs.SymRow = sf.Make(NameSymRow)
	zs.SymSoft = sf.Make(NameSymSoft)
	zs.SymSpace = sf.Make(NameSymSpace)
	zs.SymTable = sf.Make(NameSymTable)
	zs.SymText = sf.Make(NameSymText)
	zs.SymThematic = sf.Make(NameSymThematic)
	zs.SymTransclude = sf.Make(NameSymTransclude)
	zs.SymUnknown = sf.Make(NameSymUnknown)
	zs.SymVerbatimComment = sf.Make(NameSymVerbatimComment)
	zs.SymVerbatimEval = sf.Make(NameSymVerbatimEval)
	zs.SymVerbatimHTML = sf.Make(NameSymVerbatimHTML)
	zs.SymVerbatimMath = sf.Make(NameSymVerbatimMath)
	zs.SymVerbatimProg = sf.Make(NameSymVerbatimProg)
	zs.SymVerbatimZettel = sf.Make(NameSymVerbatimZettel)

	// Constant symbols for reference states.
	zs.SymRefStateInvalid = sf.Make(NameSymRefStateInvalid)
	zs.SymRefStateZettel = sf.Make(NameSymRefStateZettel)
	zs.SymRefStateSelf = sf.Make(NameSymRefStateSelf)
	zs.SymRefStateFound = sf.Make(NameSymRefStateFound)
	zs.SymRefStateBroken = sf.Make(NameSymRefStateBroken)
	zs.SymRefStateHosted = sf.Make(NameSymRefStateHosted)
	zs.SymRefStateBased = sf.Make(NameSymRefStateBased)
	zs.SymRefStateQuery = sf.Make(NameSymRefStateQuery)
	zs.SymRefStateExternal = sf.Make(NameSymRefStateExternal)

	// Symbols for metadata types.
	zs.SymTypeCredential = sf.Make(NameSymTypeCredential)
	zs.SymTypeEmpty = sf.Make(NameSymTypeEmpty)
	zs.SymTypeID = sf.Make(NameSymTypeID)
	zs.SymTypeIDSet = sf.Make(NameSymTypeIDSet)
	zs.SymTypeNumber = sf.Make(NameSymTypeNumber)
	zs.SymTypeString = sf.Make(NameSymTypeString)
	zs.SymTypeTagSet = sf.Make(NameSymTypeTagSet)
	zs.SymTypeTimestamp = sf.Make(NameSymTypeTimestamp)
	zs.SymTypeURL = sf.Make(NameSymTypeURL)
	zs.SymTypeWord = sf.Make(NameSymTypeWord)
	zs.SymTypeWordSet = sf.Make(NameSymTypeWordSet)
	zs.SymTypeZettelmarkup = sf.Make(NameSymTypeZettelmarkup)
}
