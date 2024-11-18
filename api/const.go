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

// Predefined Zettel Identifier
const (
	// System zettel
	ZidVersion              = ZettelID("00000000000001") // -> 0001
	ZidHost                 = ZettelID("00000000000002") // -> 0002
	ZidOperatingSystem      = ZettelID("00000000000003") // -> 0003
	ZidLicense              = ZettelID("00000000000004") // -> 0004
	ZidAuthors              = ZettelID("00000000000005") // -> 0005
	ZidDependencies         = ZettelID("00000000000006") // -> 0006
	ZidLog                  = ZettelID("00000000000007") // -> 0007
	ZidMemory               = ZettelID("00000000000008") // -> 0008
	ZidSx                   = ZettelID("00000000000009") // -> 0009
	ZidHTTP                 = ZettelID("00000000000010") // -> 000a
	ZidAPI                  = ZettelID("00000000000011") // -> 000b
	ZidWebUI                = ZettelID("00000000000012") // -> 000c
	ZidConsole              = ZettelID("00000000000013") // -> 000d
	ZidBoxManager           = ZettelID("00000000000020") // -> 000e
	ZidZettel               = ZettelID("00000000000021") // -> 000f
	ZidIndex                = ZettelID("00000000000022") // -> 000g
	ZidQuery                = ZettelID("00000000000023") // -> 000h
	ZidMetadataKey          = ZettelID("00000000000090") // -> 000i
	ZidParser               = ZettelID("00000000000092") // -> 000j
	ZidStartupConfiguration = ZettelID("00000000000096") // -> 000k
	ZidConfiguration        = ZettelID("00000000000100") // -> 000l
	ZidDirectory            = ZettelID("00000000000101") // -> 000m
	ZidWarnings             = ZettelID("00000000000102") // -> 000n

	// WebUI HTML templates are in the range 10000..19999
	ZidBaseTemplate   = ZettelID("00000000010100") // -> 000s
	ZidLoginTemplate  = ZettelID("00000000010200") // -> 000t
	ZidListTemplate   = ZettelID("00000000010300") // -> 000u
	ZidZettelTemplate = ZettelID("00000000010401") // -> 000v
	ZidInfoTemplate   = ZettelID("00000000010402") // -> 000w
	ZidFormTemplate   = ZettelID("00000000010403") // -> 000x
	ZidDeleteTemplate = ZettelID("00000000010405") // -> 000y
	ZidErrorTemplate  = ZettelID("00000000010700") // -> 000z

	// WebUI sxn code zettel are in the range 19000..19999
	ZidSxnStart = ZettelID("00000000019000") // -> 000q
	ZidSxnBase  = ZettelID("00000000019990") // -> 000r

	// CSS-related zettel are in the range 20000..29999
	ZidBaseCSS = ZettelID("00000000020001") // -> 0010
	ZidUserCSS = ZettelID("00000000025001") // -> 0011

	// WebUI JS zettel are in the range 30000..39999

	// WebUI image zettel are in the range 40000..49999
	ZidEmoji = ZettelID("00000000040001") // -> 000o

	// Other sxn code zettel are in the range 50000..59999
	ZidSxnPrelude = ZettelID("00000000059900") // -> 000p

	// Predefined Zettelmarkup zettel are in the range 60000..69999
	ZidRoleZettelZettel        = ZettelID("00000000060010") // -> 0012
	ZidRoleConfigurationZettel = ZettelID("00000000060020") // -> 0013
	ZidRoleRoleZettel          = ZettelID("00000000060030") // -> 0014
	ZidRoleTagZettel           = ZettelID("00000000060040") // -> 0015

	// Range 90000...99999 is reserved for zettel templates
	ZidTOCNewTemplate    = ZettelID("00000000090000") // -> 0016
	ZidTemplateNewZettel = ZettelID("00000000090001") // -> 0017
	ZidTemplateNewRole   = ZettelID("00000000090004") // -> 001a
	ZidTemplateNewTag    = ZettelID("00000000090003") // -> 0019
	ZidTemplateNewUser   = ZettelID("00000000090002") // -> 0018

	ZidSession      = ZettelID("00009999999997") // -> 00zx
	ZidAppDirectory = ZettelID("00009999999998") // -> 00zy
	ZidMapping      = ZettelID("00009999999999") // -> 00zz
	ZidDefaultHome  = ZettelID("00010000000000") // -> 0100
)

// LengthZid factors the constant length of a zettel identifier
const LengthZid = len(ZidDefaultHome)

// Values of the metadata key/value type.
const (
	MetaCredential   = "Credential"
	MetaEmpty        = "EString"
	MetaID           = "Identifier"
	MetaIDSet        = "IdentifierSet"
	MetaNumber       = "Number"
	MetaString       = "String"
	MetaTagSet       = "TagSet"
	MetaTimestamp    = "Timestamp"
	MetaURL          = "URL"
	MetaWord         = "Word"
	MetaZettelmarkup = "Zettelmarkup"
)

// Predefined general Metadata keys
const (
	KeyID           = "id"
	KeyTitle        = "title"
	KeyRole         = "role"
	KeyTags         = "tags"
	KeySyntax       = "syntax"
	KeyAuthor       = "author"
	KeyBack         = "back"
	KeyBackward     = "backward"
	KeyBoxNumber    = "box-number"
	KeyCopyright    = "copyright"
	KeyCreated      = "created"
	KeyCredential   = "credential"
	KeyDead         = "dead"
	KeyExpire       = "expire"
	KeyFolge        = "folge"
	KeyFolgeRole    = "folge-role"
	KeyForward      = "forward"
	KeyLang         = "lang"
	KeyLicense      = "license"
	KeyModified     = "modified"
	KeyPrecursor    = "precursor"
	KeyPredecessor  = "predecessor"
	KeyPrequel      = "prequel"
	KeyPublished    = "published"
	KeyQuery        = "query"
	KeyReadOnly     = "read-only"
	KeySequel       = "sequel"
	KeySuccessors   = "successors"
	KeySummary      = "summary"
	KeyURL          = "url"
	KeyUselessFiles = "useless-files"
	KeyUserID       = "user-id"
	KeyUserRole     = "user-role"
	KeyVisibility   = "visibility"
)

// Predefined Metadata values
const (
	ValueFalse             = "false"
	ValueTrue              = "true"
	ValueLangEN            = "en"
	ValueRoleConfiguration = "configuration"
	ValueRoleTag           = "tag"
	ValueRoleRole          = "role"
	ValueRoleZettel        = "zettel"
	ValueSyntaxCSS         = "css"
	ValueSyntaxDraw        = "draw"
	ValueSyntaxGif         = "gif"
	ValueSyntaxHTML        = "html"
	ValueSyntaxMarkdown    = "markdown"
	ValueSyntaxMD          = "md"
	ValueSyntaxNone        = "none"
	ValueSyntaxSVG         = "svg"
	ValueSyntaxSxn         = "sxn"
	ValueSyntaxText        = "text"
	ValueSyntaxZmk         = "zmk"
	ValueUserRoleCreator   = "creator"
	ValueUserRoleOwner     = "owner"
	ValueUserRoleReader    = "reader"
	ValueUserRoleWriter    = "writer"
	ValueVisibilityCreator = "creator"
	ValueVisibilityExpert  = "expert"
	ValueVisibilityLogin   = "login"
	ValueVisibilityOwner   = "owner"
	ValueVisibilityPublic  = "public"
)

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
	EncodingHTML  = "html"
	EncodingMD    = "md"
	EncodingSHTML = "shtml"
	EncodingSz    = "sz"
	EncodingText  = "text"
	EncodingZMK   = "zmk"

	EncodingPlain = "plain"
	EncodingData  = "data"
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

// Supported command values
const (
	CommandAuthenticated = Command("authenticated")
	CommandRefresh       = Command("refresh")
)

// Supported search operator representations
const (
	BackwardDirective = "BACKWARD"
	ContextDirective  = "CONTEXT"
	CostDirective     = "COST"
	ForwardDirective  = "FORWARD"
	FullDirective     = "FULL"
	IdentDirective    = "IDENT"
	ItemsDirective    = "ITEMS"
	MaxDirective      = "MAX"
	LimitDirective    = "LIMIT"
	OffsetDirective   = "OFFSET"
	OrDirective       = "OR"
	OrderDirective    = "ORDER"
	PhraseDirective   = "PHRASE"
	PickDirective     = "PICK"
	RandomDirective   = "RANDOM"
	ReverseDirective  = "REVERSE"
	UnlinkedDirective = "UNLINKED"

	ActionSeparator = "|"

	AtomAction     = "ATOM"
	KeysAction     = "KEYS"
	MinAction      = "MIN"
	MaxAction      = "MAX"
	NumberedAction = "NUMBERED"
	RedirectAction = "REDIRECT"
	ReIndexAction  = "REINDEX"
	RSSAction      = "RSS"
	TitleAction    = "TITLE"

	ExistOperator    = "?"
	ExistNotOperator = "!?"

	SearchOperatorNot        = "!"
	SearchOperatorEqual      = "="
	SearchOperatorNotEqual   = "!="
	SearchOperatorHas        = ":"
	SearchOperatorHasNot     = "!:"
	SearchOperatorPrefix     = "["
	SearchOperatorNoPrefix   = "!["
	SearchOperatorSuffix     = "]"
	SearchOperatorNoSuffix   = "!]"
	SearchOperatorMatch      = "~"
	SearchOperatorNoMatch    = "!~"
	SearchOperatorLess       = "<"
	SearchOperatorNotLess    = "!<"
	SearchOperatorGreater    = ">"
	SearchOperatorNotGreater = "!>"
)

// QueryPrefix is the prefix that denotes a query expression within a reference.
const QueryPrefix = "query:"
