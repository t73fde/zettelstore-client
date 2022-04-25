//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr

// Various constants for Zettel data. Some of them are technically variables.

// Symbols for Zettel node types.
var (
	SymBLOB            = GetSymbol("BLOB")
	SymCell            = GetSymbol("CELL")
	SymCellCenter      = GetSymbol("CELL-CENTER")
	SymCellLeft        = GetSymbol("CELL-LEFT")
	SymCellRight       = GetSymbol("CELL-RIGHT")
	SymCite            = GetSymbol("CITE")
	SymDescription     = GetSymbol("DESCRIPTION")
	SymEmbed           = GetSymbol("EMBED")
	SymEmbedBLOB       = GetSymbol("EMBED-BLOB")
	SymFootnote        = GetSymbol("FOOTNOTE")
	SymFormatEmph      = GetSymbol("FORMAT-EMPH")
	SymFormatDelete    = GetSymbol("FORMAT-DELETE")
	SymFormatInsert    = GetSymbol("FORMAT-INSERT")
	SymFormatQuote     = GetSymbol("FORMAT-QUOTE")
	SymFormatSpan      = GetSymbol("FORMAT-SPAN")
	SymFormatSub       = GetSymbol("FORMAT-SUB")
	SymFormatSuper     = GetSymbol("FORMAT-SUPER")
	SymFormatStrong    = GetSymbol("FORMAT-STRONG")
	SymHard            = GetSymbol("HARD")
	SymHeading         = GetSymbol("HEADING")
	SymLink            = GetSymbol("LINK")
	SymListOrdered     = GetSymbol("ORDERED")
	SymListUnordered   = GetSymbol("UNORDERED")
	SymListQuote       = GetSymbol("QUOTATION")
	SymLiteralProg     = GetSymbol("LITERAL-CODE")
	SymLiteralComment  = GetSymbol("LITERAL-COMMENT")
	SymLiteralHTML     = GetSymbol("LITERAL-HTML")
	SymLiteralInput    = GetSymbol("LITERAL-INPUT")
	SymLiteralMath     = GetSymbol("LITERAL-MATH")
	SymLiteralOutput   = GetSymbol("LITERAL-OUTPUT")
	SymLiteralZettel   = GetSymbol("LITERAL-ZETTEL")
	SymMark            = GetSymbol("MARK")
	SymPara            = GetSymbol("PARA")
	SymRegionSpan      = GetSymbol("REGION-BLOCK")
	SymRegionQuote     = GetSymbol("REGION-QUOTE")
	SymRegionVerse     = GetSymbol("REGION-VERSE")
	SymSoft            = GetSymbol("SOFT")
	SymSpace           = GetSymbol("SPACE")
	SymTable           = GetSymbol("TABLE")
	SymTag             = GetSymbol("TAG")
	SymText            = GetSymbol("TEXT")
	SymThematic        = GetSymbol("THEMATIC")
	SymTransclude      = GetSymbol("TRANSCLUDE")
	SymUnknown         = GetSymbol("UNKNOWN-NODE")
	SymVerbatimComment = GetSymbol("VERBATIM-COMMENT")
	SymVerbatimEval    = GetSymbol("VERBATIM-EVAL")
	SymVerbatimHTML    = GetSymbol("VERBATIM-HTML")
	SymVerbatimMath    = GetSymbol("VERBATIM-MATH")
	SymVerbatimProg    = GetSymbol("VERBATIM-CODE")
	SymVerbatimZettel  = GetSymbol("VERBATIM-ZETTEL")
)

// Constant symbols for reference states.
var (
	SymRefStateInvalid  = GetSymbol("INVALID")
	SymRefStateZettel   = GetSymbol("ZETTEL")
	SymRefStateSelf     = GetSymbol("SELF")
	SymRefStateFound    = GetSymbol("FOUND")
	SymRefStateBroken   = GetSymbol("BROKEN")
	SymRefStateHosted   = GetSymbol("HOSTED")
	SymRefStateBased    = GetSymbol("BASED")
	SymRefStateExternal = GetSymbol("EXTERNAL")
)

// Symbols for metadata types
var (
	SymTypeCredential   = GetSymbol("CREDENTIAL")
	SymTypeEmpty        = GetSymbol("EMPTY-STRING")
	SymTypeID           = GetSymbol("ZID")
	SymTypeIDSet        = GetSymbol("ZID-SET")
	SymTypeNumber       = GetSymbol("NUMBER")
	SymTypeString       = GetSymbol("STRING")
	SymTypeTagSet       = GetSymbol("TAG-SET")
	SymTypeTimestamp    = GetSymbol("TIMESTAMP")
	SymTypeURL          = GetSymbol("URL")
	SymTypeWord         = GetSymbol("WORD")
	SymTypeWordSet      = GetSymbol("WORD-SET")
	SymTypeZettelmarkup = GetSymbol("ZETTELMARKUP")
)
