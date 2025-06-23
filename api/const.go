//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2021-present Detlef Stern
//-----------------------------------------------------------------------------

package api

import "fmt"

// Additional HTTP constants.
const (
	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"
	HeaderDestination = "Destination"
	HeaderLocation    = "Location"
)

// Values for HTTP query parameter.
const (
	QueryKeyCommand   = "cmd"
	QueryKeyEncoding  = "enc"
	QueryKeyParseOnly = "parseonly"
	QueryKeyPart      = "part"
	QueryKeyPhrase    = "phrase"
	QueryKeyQuery     = "q"
	QueryKeyRole      = "role"
	QueryKeySeed      = "_seed"
	QueryKeyTag       = "tag"
)

// Supported encoding values.
const (
	EncodingHTML  = "html"  // Plain HTML
	EncodingMD    = "md"    // Markdown
	EncodingSHTML = "shtml" // SxHTML
	EncodingSz    = "sz"    // Structure of zettel, encoded a an S-expression
	EncodingText  = "text"  // plain text content
	EncodingZMK   = "zmk"   // Zettelmarkup

	EncodingPlain = "plain" // Plain zettel, no processing
	EncodingData  = "data"  // Plain zettel, metadata as S-Expression
)

var mapEncodingEnum = map[string]EncodingEnum{
	EncodingHTML:  EncoderHTML,
	EncodingMD:    EncoderMD,
	EncodingSHTML: EncoderSHTML,
	EncodingSz:    EncoderSz,
	EncodingText:  EncoderText,
	EncodingZMK:   EncoderZmk,

	EncodingPlain: EncoderPlain,
	EncodingData:  EncoderData,
}
var mapEnumEncoding = map[EncodingEnum]string{}

func init() {
	for k, v := range mapEncodingEnum {
		mapEnumEncoding[v] = k
	}
}

// Encoder returns the internal encoder code for the given encoding string.
func Encoder(encoding string) EncodingEnum {
	if e, ok := mapEncodingEnum[encoding]; ok {
		return e
	}
	return EncoderUnknown
}

// EncodingEnum lists all valid encoder keys.
type EncodingEnum uint8

// Values for EncoderEnum
const (
	EncoderUnknown EncodingEnum = iota
	EncoderHTML
	EncoderMD
	EncoderSHTML
	EncoderSz
	EncoderText
	EncoderZmk

	EncoderPlain
	EncoderData
)

// String representation of an encoder key.
func (e EncodingEnum) String() string {
	if f, ok := mapEnumEncoding[e]; ok {
		return f
	}
	return fmt.Sprintf("*Unknown*(%d)", e)
}

// Supported part values.
const (
	PartMeta    = "meta"
	PartContent = "content"
	PartZettel  = "zettel"
)

// Command to be executed atthe Zettelstore
type Command string

// Supported command values.
const (
	CommandAuthenticated = Command("authenticated")
	CommandRefresh       = Command("refresh")
)

// Supported search operator representations.
const (
	BackwardDirective = "BACKWARD" // Backward-only context / thread
	ContextDirective  = "CONTEXT"  // Context directive
	CostDirective     = "COST"     // Maximum cost of a context operation
	FolgeDirective    = "FOLGE"    // Folge thread
	ForwardDirective  = "FORWARD"  // Forward-only context / thread
	FullDirective     = "FULL"     // Include tags in context
	IdentDirective    = "IDENT"    // Use only specified zettel
	ItemsDirective    = "ITEMS"    // Select list elements in a zettel
	MaxDirective      = "MAX"      // Maximum number of context / thread results
	MinDirective      = "MIN"      // Minimum number of context results
	LimitDirective    = "LIMIT"    // Maximum number of zettel
	OffsetDirective   = "OFFSET"   // Offset to start returned zettel list
	OrDirective       = "OR"       // Combine several search expression with an "or"
	OrderDirective    = "ORDER"    // Specify metadata keys for the order of returned list
	PhraseDirective   = "PHRASE"   // Only unlinked zettel with given phrase
	PickDirective     = "PICK"     // Pick some random zettel
	RandomDirective   = "RANDOM"   // Order zettel list randomly
	ReverseDirective  = "REVERSE"  // Reverse the order of a zettel list
	SequelDirective   = "SEQUEL"   // Sequel / branching thread
	ThreadDirective   = "THREAD"   // Both folge and Sequel thread
	UnlinkedDirective = "UNLINKED" // Search for zettel that contain a phase(s) but do not link

	ActionSeparator = "|" // Separates action list of previous elements of query expression

	KeysAction     = "KEYS"     // Provide metadata key used
	MinAction      = "MIN"      // Return only those values with a minimum amount of zettel
	MaxAction      = "MAX"      // Return only those values with a maximum amount of zettel
	NumberedAction = "NUMBERED" // Return a numbered list
	RedirectAction = "REDIRECT" // Return the first zettel in list
	ReIndexAction  = "REINDEX"  // Ensure that zettel is/are indexed.

	ExistOperator    = "?"  // Does zettel have metadata with given key?
	ExistNotOperator = "!?" // True id zettel does not have metadata with given key.

	SearchOperatorNot        = "!"
	SearchOperatorEqual      = "="  // True if values are equal
	SearchOperatorNotEqual   = "!=" // False if values are equal
	SearchOperatorHas        = ":"  // True if values are equal/included
	SearchOperatorHasNot     = "!:" // False if values are equal/included
	SearchOperatorPrefix     = "["  // True if value is prefix of the other
	SearchOperatorNoPrefix   = "![" // False if value is prefix of the other
	SearchOperatorSuffix     = "]"  // True if value is suffix of other
	SearchOperatorNoSuffix   = "!]" // False if value is suffix of other
	SearchOperatorMatch      = "~"  // True if value is included in other
	SearchOperatorNoMatch    = "!~" // False if value is included in other
	SearchOperatorLess       = "<"  // True if value is smaller than other
	SearchOperatorNotLess    = "!<" // False if value is smaller than other
	SearchOperatorGreater    = ">"  // True if value is greater than other
	SearchOperatorNotGreater = "!>" // False if value is greater than other
)

// QueryPrefix is the prefix that denotes a query expression within a reference.
const QueryPrefix = "query:"
