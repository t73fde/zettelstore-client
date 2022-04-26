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
	SymBLOB            = NewSymbol("BLOB")
	SymCell            = NewSymbol("CELL")
	SymCellCenter      = NewSymbol("CELL-CENTER")
	SymCellLeft        = NewSymbol("CELL-LEFT")
	SymCellRight       = NewSymbol("CELL-RIGHT")
	SymCite            = NewSymbol("CITE")
	SymDescription     = NewSymbol("DESCRIPTION")
	SymEmbed           = NewSymbol("EMBED")
	SymEmbedBLOB       = NewSymbol("EMBED-BLOB")
	SymFootnote        = NewSymbol("FOOTNOTE")
	SymFormatEmph      = NewSymbol("FORMAT-EMPH")
	SymFormatDelete    = NewSymbol("FORMAT-DELETE")
	SymFormatInsert    = NewSymbol("FORMAT-INSERT")
	SymFormatQuote     = NewSymbol("FORMAT-QUOTE")
	SymFormatSpan      = NewSymbol("FORMAT-SPAN")
	SymFormatSub       = NewSymbol("FORMAT-SUB")
	SymFormatSuper     = NewSymbol("FORMAT-SUPER")
	SymFormatStrong    = NewSymbol("FORMAT-STRONG")
	SymHard            = NewSymbol("HARD")
	SymHeading         = NewSymbol("HEADING")
	SymLink            = NewSymbol("LINK")
	SymListOrdered     = NewSymbol("ORDERED")
	SymListUnordered   = NewSymbol("UNORDERED")
	SymListQuote       = NewSymbol("QUOTATION")
	SymLiteralProg     = NewSymbol("LITERAL-CODE")
	SymLiteralComment  = NewSymbol("LITERAL-COMMENT")
	SymLiteralHTML     = NewSymbol("LITERAL-HTML")
	SymLiteralInput    = NewSymbol("LITERAL-INPUT")
	SymLiteralMath     = NewSymbol("LITERAL-MATH")
	SymLiteralOutput   = NewSymbol("LITERAL-OUTPUT")
	SymLiteralZettel   = NewSymbol("LITERAL-ZETTEL")
	SymMark            = NewSymbol("MARK")
	SymPara            = NewSymbol("PARA")
	SymRegionSpan      = NewSymbol("REGION-BLOCK")
	SymRegionQuote     = NewSymbol("REGION-QUOTE")
	SymRegionVerse     = NewSymbol("REGION-VERSE")
	SymSoft            = NewSymbol("SOFT")
	SymSpace           = NewSymbol("SPACE")
	SymTable           = NewSymbol("TABLE")
	SymTag             = NewSymbol("TAG")
	SymText            = NewSymbol("TEXT")
	SymThematic        = NewSymbol("THEMATIC")
	SymTransclude      = NewSymbol("TRANSCLUDE")
	SymUnknown         = NewSymbol("UNKNOWN-NODE")
	SymVerbatimComment = NewSymbol("VERBATIM-COMMENT")
	SymVerbatimEval    = NewSymbol("VERBATIM-EVAL")
	SymVerbatimHTML    = NewSymbol("VERBATIM-HTML")
	SymVerbatimMath    = NewSymbol("VERBATIM-MATH")
	SymVerbatimProg    = NewSymbol("VERBATIM-CODE")
	SymVerbatimZettel  = NewSymbol("VERBATIM-ZETTEL")
)

// Constant symbols for reference states.
var (
	SymRefStateInvalid  = NewSymbol("INVALID")
	SymRefStateZettel   = NewSymbol("ZETTEL")
	SymRefStateSelf     = NewSymbol("SELF")
	SymRefStateFound    = NewSymbol("FOUND")
	SymRefStateBroken   = NewSymbol("BROKEN")
	SymRefStateHosted   = NewSymbol("HOSTED")
	SymRefStateBased    = NewSymbol("BASED")
	SymRefStateExternal = NewSymbol("EXTERNAL")
)

// Symbols for metadata types
var (
	SymTypeCredential   = NewSymbol("CREDENTIAL")
	SymTypeEmpty        = NewSymbol("EMPTY-STRING")
	SymTypeID           = NewSymbol("ZID")
	SymTypeIDSet        = NewSymbol("ZID-SET")
	SymTypeNumber       = NewSymbol("NUMBER")
	SymTypeString       = NewSymbol("STRING")
	SymTypeTagSet       = NewSymbol("TAG-SET")
	SymTypeTimestamp    = NewSymbol("TIMESTAMP")
	SymTypeURL          = NewSymbol("URL")
	SymTypeWord         = NewSymbol("WORD")
	SymTypeWordSet      = NewSymbol("WORD-SET")
	SymTypeZettelmarkup = NewSymbol("ZETTELMARKUP")
)
