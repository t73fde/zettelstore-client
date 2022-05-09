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
	Env                = sxpf.NewTrivialEnvironment()
	SymBLOB            = Env.MakeSymbol("BLOB")
	SymCell            = Env.MakeSymbol("CELL")
	SymCellCenter      = Env.MakeSymbol("CELL-CENTER")
	SymCellLeft        = Env.MakeSymbol("CELL-LEFT")
	SymCellRight       = Env.MakeSymbol("CELL-RIGHT")
	SymCite            = Env.MakeSymbol("CITE")
	SymDescription     = Env.MakeSymbol("DESCRIPTION")
	SymEmbed           = Env.MakeSymbol("EMBED")
	SymEmbedBLOB       = Env.MakeSymbol("EMBED-BLOB")
	SymFootnote        = Env.MakeSymbol("FOOTNOTE")
	SymFormatEmph      = Env.MakeSymbol("FORMAT-EMPH")
	SymFormatDelete    = Env.MakeSymbol("FORMAT-DELETE")
	SymFormatInsert    = Env.MakeSymbol("FORMAT-INSERT")
	SymFormatQuote     = Env.MakeSymbol("FORMAT-QUOTE")
	SymFormatSpan      = Env.MakeSymbol("FORMAT-SPAN")
	SymFormatSub       = Env.MakeSymbol("FORMAT-SUB")
	SymFormatSuper     = Env.MakeSymbol("FORMAT-SUPER")
	SymFormatStrong    = Env.MakeSymbol("FORMAT-STRONG")
	SymHard            = Env.MakeSymbol("HARD")
	SymHeading         = Env.MakeSymbol("HEADING")
	SymLink            = Env.MakeSymbol("LINK")
	SymListOrdered     = Env.MakeSymbol("ORDERED")
	SymListUnordered   = Env.MakeSymbol("UNORDERED")
	SymListQuote       = Env.MakeSymbol("QUOTATION")
	SymLiteralProg     = Env.MakeSymbol("LITERAL-CODE")
	SymLiteralComment  = Env.MakeSymbol("LITERAL-COMMENT")
	SymLiteralHTML     = Env.MakeSymbol("LITERAL-HTML")
	SymLiteralInput    = Env.MakeSymbol("LITERAL-INPUT")
	SymLiteralMath     = Env.MakeSymbol("LITERAL-MATH")
	SymLiteralOutput   = Env.MakeSymbol("LITERAL-OUTPUT")
	SymLiteralZettel   = Env.MakeSymbol("LITERAL-ZETTEL")
	SymMark            = Env.MakeSymbol("MARK")
	SymPara            = Env.MakeSymbol("PARA")
	SymRegionBlock     = Env.MakeSymbol("REGION-BLOCK")
	SymRegionQuote     = Env.MakeSymbol("REGION-QUOTE")
	SymRegionVerse     = Env.MakeSymbol("REGION-VERSE")
	SymSoft            = Env.MakeSymbol("SOFT")
	SymSpace           = Env.MakeSymbol("SPACE")
	SymTable           = Env.MakeSymbol("TABLE")
	SymTag             = Env.MakeSymbol("TAG")
	SymText            = Env.MakeSymbol("TEXT")
	SymThematic        = Env.MakeSymbol("THEMATIC")
	SymTransclude      = Env.MakeSymbol("TRANSCLUDE")
	SymUnknown         = Env.MakeSymbol("UNKNOWN-NODE")
	SymVerbatimComment = Env.MakeSymbol("VERBATIM-COMMENT")
	SymVerbatimEval    = Env.MakeSymbol("VERBATIM-EVAL")
	SymVerbatimHTML    = Env.MakeSymbol("VERBATIM-HTML")
	SymVerbatimMath    = Env.MakeSymbol("VERBATIM-MATH")
	SymVerbatimProg    = Env.MakeSymbol("VERBATIM-CODE")
	SymVerbatimZettel  = Env.MakeSymbol("VERBATIM-ZETTEL")
)

// Constant symbols for reference states.
var (
	SymRefStateInvalid  = Env.MakeSymbol("INVALID")
	SymRefStateZettel   = Env.MakeSymbol("ZETTEL")
	SymRefStateSelf     = Env.MakeSymbol("SELF")
	SymRefStateFound    = Env.MakeSymbol("FOUND")
	SymRefStateBroken   = Env.MakeSymbol("BROKEN")
	SymRefStateHosted   = Env.MakeSymbol("HOSTED")
	SymRefStateBased    = Env.MakeSymbol("BASED")
	SymRefStateExternal = Env.MakeSymbol("EXTERNAL")
)

// Symbols for metadata types
var (
	SymTypeCredential   = Env.MakeSymbol("CREDENTIAL")
	SymTypeEmpty        = Env.MakeSymbol("EMPTY-STRING")
	SymTypeID           = Env.MakeSymbol("ZID")
	SymTypeIDSet        = Env.MakeSymbol("ZID-SET")
	SymTypeNumber       = Env.MakeSymbol("NUMBER")
	SymTypeString       = Env.MakeSymbol("STRING")
	SymTypeTagSet       = Env.MakeSymbol("TAG-SET")
	SymTypeTimestamp    = Env.MakeSymbol("TIMESTAMP")
	SymTypeURL          = Env.MakeSymbol("URL")
	SymTypeWord         = Env.MakeSymbol("WORD")
	SymTypeWordSet      = Env.MakeSymbol("WORD-SET")
	SymTypeZettelmarkup = Env.MakeSymbol("ZETTELMARKUP")
)
