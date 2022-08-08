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

import "codeberg.org/t73fde/sxpf"

// Various constants for Zettel data. Some of them are technically variables.

// Symbols for Zettel node types.
var (
	Smk                = sxpf.NewTrivialSymbolMaker()
	SymBLOB            = Smk.MakeSymbol("BLOB")
	SymCell            = Smk.MakeSymbol("CELL")
	SymCellCenter      = Smk.MakeSymbol("CELL-CENTER")
	SymCellLeft        = Smk.MakeSymbol("CELL-LEFT")
	SymCellRight       = Smk.MakeSymbol("CELL-RIGHT")
	SymCite            = Smk.MakeSymbol("CITE")
	SymDescription     = Smk.MakeSymbol("DESCRIPTION")
	SymEmbed           = Smk.MakeSymbol("EMBED")
	SymEmbedBLOB       = Smk.MakeSymbol("EMBED-BLOB")
	SymFootnote        = Smk.MakeSymbol("FOOTNOTE")
	SymFormatEmph      = Smk.MakeSymbol("FORMAT-EMPH")
	SymFormatDelete    = Smk.MakeSymbol("FORMAT-DELETE")
	SymFormatInsert    = Smk.MakeSymbol("FORMAT-INSERT")
	SymFormatQuote     = Smk.MakeSymbol("FORMAT-QUOTE")
	SymFormatSpan      = Smk.MakeSymbol("FORMAT-SPAN")
	SymFormatSub       = Smk.MakeSymbol("FORMAT-SUB")
	SymFormatSuper     = Smk.MakeSymbol("FORMAT-SUPER")
	SymFormatStrong    = Smk.MakeSymbol("FORMAT-STRONG")
	SymHard            = Smk.MakeSymbol("HARD")
	SymHeading         = Smk.MakeSymbol("HEADING")
	SymLinkInvalid     = Smk.MakeSymbol("LINK-INVALID")
	SymLinkZettel      = Smk.MakeSymbol("LINK-ZETTEL")
	SymLinkSelf        = Smk.MakeSymbol("LINK-SELF")
	SymLinkFound       = Smk.MakeSymbol("LINK-FOUND")
	SymLinkBroken      = Smk.MakeSymbol("LINK-BROKEN")
	SymLinkHosted      = Smk.MakeSymbol("LINK-HOSTED")
	SymLinkBased       = Smk.MakeSymbol("LINK-BASED")
	SymLinkSearch      = Smk.MakeSymbol("LINK-SEARCH")
	SymLinkExternal    = Smk.MakeSymbol("LINK-EXTERNAL")
	SymListOrdered     = Smk.MakeSymbol("ORDERED")
	SymListUnordered   = Smk.MakeSymbol("UNORDERED")
	SymListQuote       = Smk.MakeSymbol("QUOTATION")
	SymLiteralProg     = Smk.MakeSymbol("LITERAL-CODE")
	SymLiteralComment  = Smk.MakeSymbol("LITERAL-COMMENT")
	SymLiteralHTML     = Smk.MakeSymbol("LITERAL-HTML")
	SymLiteralInput    = Smk.MakeSymbol("LITERAL-INPUT")
	SymLiteralMath     = Smk.MakeSymbol("LITERAL-MATH")
	SymLiteralOutput   = Smk.MakeSymbol("LITERAL-OUTPUT")
	SymLiteralZettel   = Smk.MakeSymbol("LITERAL-ZETTEL")
	SymMark            = Smk.MakeSymbol("MARK")
	SymPara            = Smk.MakeSymbol("PARA")
	SymRegionBlock     = Smk.MakeSymbol("REGION-BLOCK")
	SymRegionQuote     = Smk.MakeSymbol("REGION-QUOTE")
	SymRegionVerse     = Smk.MakeSymbol("REGION-VERSE")
	SymSoft            = Smk.MakeSymbol("SOFT")
	SymSpace           = Smk.MakeSymbol("SPACE")
	SymTable           = Smk.MakeSymbol("TABLE")
	SymTag             = Smk.MakeSymbol("TAG")
	SymText            = Smk.MakeSymbol("TEXT")
	SymThematic        = Smk.MakeSymbol("THEMATIC")
	SymTransclude      = Smk.MakeSymbol("TRANSCLUDE")
	SymUnknown         = Smk.MakeSymbol("UNKNOWN-NODE")
	SymVerbatimComment = Smk.MakeSymbol("VERBATIM-COMMENT")
	SymVerbatimEval    = Smk.MakeSymbol("VERBATIM-EVAL")
	SymVerbatimHTML    = Smk.MakeSymbol("VERBATIM-HTML")
	SymVerbatimMath    = Smk.MakeSymbol("VERBATIM-MATH")
	SymVerbatimProg    = Smk.MakeSymbol("VERBATIM-CODE")
	SymVerbatimZettel  = Smk.MakeSymbol("VERBATIM-ZETTEL")
)

// Constant symbols for reference states.
var (
	SymRefStateInvalid  = Smk.MakeSymbol("INVALID")
	SymRefStateZettel   = Smk.MakeSymbol("ZETTEL")
	SymRefStateSelf     = Smk.MakeSymbol("SELF")
	SymRefStateFound    = Smk.MakeSymbol("FOUND")
	SymRefStateBroken   = Smk.MakeSymbol("BROKEN")
	SymRefStateHosted   = Smk.MakeSymbol("HOSTED")
	SymRefStateBased    = Smk.MakeSymbol("BASED")
	SymRefStateSearch   = Smk.MakeSymbol("SEARCH")
	SymRefStateExternal = Smk.MakeSymbol("EXTERNAL")
)

// Symbols for metadata types
var (
	SymTypeCredential   = Smk.MakeSymbol("CREDENTIAL")
	SymTypeEmpty        = Smk.MakeSymbol("EMPTY-STRING")
	SymTypeID           = Smk.MakeSymbol("ZID")
	SymTypeIDSet        = Smk.MakeSymbol("ZID-SET")
	SymTypeNumber       = Smk.MakeSymbol("NUMBER")
	SymTypeString       = Smk.MakeSymbol("STRING")
	SymTypeTagSet       = Smk.MakeSymbol("TAG-SET")
	SymTypeTimestamp    = Smk.MakeSymbol("TIMESTAMP")
	SymTypeURL          = Smk.MakeSymbol("URL")
	SymTypeWord         = Smk.MakeSymbol("WORD")
	SymTypeWordSet      = Smk.MakeSymbol("WORD-SET")
	SymTypeZettelmarkup = Smk.MakeSymbol("ZETTELMARKUP")
)
