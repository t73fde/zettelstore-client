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

import "github.com/t73fde/sxpf"

// Various constants for Zettel data. Some of them are technically variables.

// Symbols for Zettel node types.
var (
	SymBLOB            = sxpf.NewSymbol("BLOB")
	SymCell            = sxpf.NewSymbol("CELL")
	SymCellCenter      = sxpf.NewSymbol("CELL-CENTER")
	SymCellLeft        = sxpf.NewSymbol("CELL-LEFT")
	SymCellRight       = sxpf.NewSymbol("CELL-RIGHT")
	SymCite            = sxpf.NewSymbol("CITE")
	SymDescription     = sxpf.NewSymbol("DESCRIPTION")
	SymEmbed           = sxpf.NewSymbol("EMBED")
	SymEmbedBLOB       = sxpf.NewSymbol("EMBED-BLOB")
	SymFootnote        = sxpf.NewSymbol("FOOTNOTE")
	SymFormatEmph      = sxpf.NewSymbol("FORMAT-EMPH")
	SymFormatDelete    = sxpf.NewSymbol("FORMAT-DELETE")
	SymFormatInsert    = sxpf.NewSymbol("FORMAT-INSERT")
	SymFormatQuote     = sxpf.NewSymbol("FORMAT-QUOTE")
	SymFormatSpan      = sxpf.NewSymbol("FORMAT-SPAN")
	SymFormatSub       = sxpf.NewSymbol("FORMAT-SUB")
	SymFormatSuper     = sxpf.NewSymbol("FORMAT-SUPER")
	SymFormatStrong    = sxpf.NewSymbol("FORMAT-STRONG")
	SymHard            = sxpf.NewSymbol("HARD")
	SymHeading         = sxpf.NewSymbol("HEADING")
	SymLink            = sxpf.NewSymbol("LINK")
	SymListOrdered     = sxpf.NewSymbol("ORDERED")
	SymListUnordered   = sxpf.NewSymbol("UNORDERED")
	SymListQuote       = sxpf.NewSymbol("QUOTATION")
	SymLiteralProg     = sxpf.NewSymbol("LITERAL-CODE")
	SymLiteralComment  = sxpf.NewSymbol("LITERAL-COMMENT")
	SymLiteralHTML     = sxpf.NewSymbol("LITERAL-HTML")
	SymLiteralInput    = sxpf.NewSymbol("LITERAL-INPUT")
	SymLiteralMath     = sxpf.NewSymbol("LITERAL-MATH")
	SymLiteralOutput   = sxpf.NewSymbol("LITERAL-OUTPUT")
	SymLiteralZettel   = sxpf.NewSymbol("LITERAL-ZETTEL")
	SymMark            = sxpf.NewSymbol("MARK")
	SymPara            = sxpf.NewSymbol("PARA")
	SymRegionBlock     = sxpf.NewSymbol("REGION-BLOCK")
	SymRegionQuote     = sxpf.NewSymbol("REGION-QUOTE")
	SymRegionVerse     = sxpf.NewSymbol("REGION-VERSE")
	SymSoft            = sxpf.NewSymbol("SOFT")
	SymSpace           = sxpf.NewSymbol("SPACE")
	SymTable           = sxpf.NewSymbol("TABLE")
	SymTag             = sxpf.NewSymbol("TAG")
	SymText            = sxpf.NewSymbol("TEXT")
	SymThematic        = sxpf.NewSymbol("THEMATIC")
	SymTransclude      = sxpf.NewSymbol("TRANSCLUDE")
	SymUnknown         = sxpf.NewSymbol("UNKNOWN-NODE")
	SymVerbatimComment = sxpf.NewSymbol("VERBATIM-COMMENT")
	SymVerbatimEval    = sxpf.NewSymbol("VERBATIM-EVAL")
	SymVerbatimHTML    = sxpf.NewSymbol("VERBATIM-HTML")
	SymVerbatimMath    = sxpf.NewSymbol("VERBATIM-MATH")
	SymVerbatimProg    = sxpf.NewSymbol("VERBATIM-CODE")
	SymVerbatimZettel  = sxpf.NewSymbol("VERBATIM-ZETTEL")
)

// Constant symbols for reference states.
var (
	SymRefStateInvalid  = sxpf.NewSymbol("INVALID")
	SymRefStateZettel   = sxpf.NewSymbol("ZETTEL")
	SymRefStateSelf     = sxpf.NewSymbol("SELF")
	SymRefStateFound    = sxpf.NewSymbol("FOUND")
	SymRefStateBroken   = sxpf.NewSymbol("BROKEN")
	SymRefStateHosted   = sxpf.NewSymbol("HOSTED")
	SymRefStateBased    = sxpf.NewSymbol("BASED")
	SymRefStateExternal = sxpf.NewSymbol("EXTERNAL")
)

// Symbols for metadata types
var (
	SymTypeCredential   = sxpf.NewSymbol("CREDENTIAL")
	SymTypeEmpty        = sxpf.NewSymbol("EMPTY-STRING")
	SymTypeID           = sxpf.NewSymbol("ZID")
	SymTypeIDSet        = sxpf.NewSymbol("ZID-SET")
	SymTypeNumber       = sxpf.NewSymbol("NUMBER")
	SymTypeString       = sxpf.NewSymbol("STRING")
	SymTypeTagSet       = sxpf.NewSymbol("TAG-SET")
	SymTypeTimestamp    = sxpf.NewSymbol("TIMESTAMP")
	SymTypeURL          = sxpf.NewSymbol("URL")
	SymTypeWord         = sxpf.NewSymbol("WORD")
	SymTypeWordSet      = sxpf.NewSymbol("WORD-SET")
	SymTypeZettelmarkup = sxpf.NewSymbol("ZETTELMARKUP")
)
