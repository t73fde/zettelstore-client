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
	NameAttribute   = "a"
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
	TypeBlock           = "Block"
	TypeBreakHard       = "Hard"
	TypeBreakSoft       = "Soft"
	TypeBreakThematic   = "Thematic"
	TypeCitation        = "Cite"
	TypeDescription     = "Description"
	TypeEmbed           = "Embed"
	TypeEmbedBLOB       = "EmbedBLOB"
	TypeExcerpt         = "Excerpt"
	TypeFootnote        = "Footnote"
	TypeFormatDelete    = "Delete"
	TypeFormatEmph      = "Emph"
	TypeFormatInsert    = "Insert"
	TypeFormatQuote     = "Quote"
	TypeFormatSpan      = "Span"
	TypeFormatStrong    = "Strong"
	TypeFormatSub       = "Sub"
	TypeFormatSuper     = "Super"
	TypeHeading         = "Heading"
	TypeLink            = "Link"
	TypeListBullet      = "Bullet"
	TypeListOrdered     = "Ordered"
	TypeListQuotation   = "Quotation"
	TypeLiteralCode     = "Code"
	TypeLiteralComment  = "Comment"
	TypeLiteralHTML     = "HTML"
	TypeLiteralInput    = "Input"
	TypeLiteralOutput   = "Output"
	TypeLiteralZettel   = "Zettel"
	TypeMark            = "Mark"
	TypeParagraph       = "Para"
	TypePoem            = "Poem"
	TypeSpace           = "Space"
	TypeTable           = "Table"
	TypeTag             = "Tag"
	TypeText            = "Text"
	TypeTransclude      = "Transclude"
	TypeVerbatimCode    = "CodeBlock"
	TypeVerbatimComment = "CommentBlock"
	TypeVerbatimHTML    = "HTMLBlock"
	TypeVerbatimZettel  = "ZettelBlock"
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

// Values for table cell alignment
const (
	AlignDefault = ""
	AlignLeft    = "<"
	AlignCenter  = ":"
	AlignRight   = ">"
)
