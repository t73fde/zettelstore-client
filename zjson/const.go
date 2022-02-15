//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package zjson

// Values for Zettelmarkup element object names
const (
	NameType        = ""
	NameBLOB        = "j"
	NameBinary      = "o"
	NameBlock       = "b"
	NameDescription = "g"
	NameInline      = "i"
	NameList        = "c"
	NameNumeric     = "n"
	NameString      = "s"
	NameString2     = "q"
	NameString3     = "v"
	NameTable       = "p"
)

// Values to specify the Zettelmarkup element type
const (
	TypeBLOB            = "BLOB"
	TypeBreakHard       = "Hard"
	TypeBreakSoft       = "Soft"
	TypeCitation        = "Cite"
	TypeDescriptionList = "DescriptionList"
	TypeEmbed           = "Embed"
	TypeEmbedBLOB       = "EmbedBLOB"
	TypeFootnote        = "Footnote"
	TypeFormatDelete    = "Delete"
	TypeFormatEmph      = "Emph"
	TypeFormatInsert    = "Insert"
	TypeFormatMonospace = "Mono"
	TypeFormatQuotation = "Quotation"
	TypeFormatQuote     = "Quote"
	TypeFormatSpan      = "Span"
	TypeFormatStrong    = "Strong"
	TypeFormatSub       = "Sub"
	TypeFormatSuper     = "Super"
	TypeHeading         = "Heading"
	TypeLink            = "Link"
	TypeListBullet      = "BulletList"
	TypeListOrdered     = "OrderedList"
	TypeListQuote       = "QuoteList"
	TypeLiteralComment  = "Comment"
	TypeLiteralHTML     = "HTML"
	TypeLiteralKeyb     = "Input"
	TypeLiteralOutput   = "Output"
	TypeLiteralProg     = "Code"
	TypeLiteralZettel   = "Zettel"
	TypeMark            = "Mark"
	TypeParagraph       = "Para"
	TypeQuoteBlock      = "QuoteBlock"
	TypeSpace           = "Space"
	TypeSpanBlock       = "SpanBlock"
	TypeTable           = "Table"
	TypeTag             = "Tag"
	TypeText            = "Text"
	TypeBreakThematic   = "Thematik"
	TypeTransclude      = "Transclude"
	TypeVerbatimCode    = "CodeBlock"
	TypeVerbatimComment = "CommentBlock"
	TypeVerbatimHTML    = "HTMLBlock"
	TypeVerbatimZettel  = "ZettelBlock"
	TypeVerse           = "Verse"
)

// Values to specify the state of a reference
const (
	RefStateBased    = "based"
	RefStateBroken   = "broken"
	RefStateExternal = "external"
	RefStateFound    = "found"
	RefStateHosted   = "local"
	RefStateInvalid  = "invalid"
	RefStateSelf     = "self"
	RefStateZettel   = "zettel"
)
