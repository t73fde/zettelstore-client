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

// Predefined zettel identifier.
//
// See [List of predefined zettel].
//
// [List of predefined zettel]: https://zettelstore.de/manual/h/00001005090000
const (
	// System zettel
	ZidVersion              = ZettelID("00000000000001")
	ZidHost                 = ZettelID("00000000000002")
	ZidOperatingSystem      = ZettelID("00000000000003")
	ZidLicense              = ZettelID("00000000000004")
	ZidAuthors              = ZettelID("00000000000005")
	ZidDependencies         = ZettelID("00000000000006")
	ZidLog                  = ZettelID("00000000000007")
	ZidMemory               = ZettelID("00000000000008")
	ZidSx                   = ZettelID("00000000000009")
	ZidHTTP                 = ZettelID("00000000000010")
	ZidAPI                  = ZettelID("00000000000011")
	ZidWebUI                = ZettelID("00000000000012")
	ZidConsole              = ZettelID("00000000000013")
	ZidBoxManager           = ZettelID("00000000000020")
	ZidZettel               = ZettelID("00000000000021")
	ZidIndex                = ZettelID("00000000000022")
	ZidQuery                = ZettelID("00000000000023")
	ZidMetadataKey          = ZettelID("00000000000090")
	ZidParser               = ZettelID("00000000000092")
	ZidStartupConfiguration = ZettelID("00000000000096")
	ZidConfiguration        = ZettelID("00000000000100")
	ZidDirectory            = ZettelID("00000000000101")

	// WebUI HTML templates are in the range 10000..19999
	ZidBaseTemplate   = ZettelID("00000000010100")
	ZidLoginTemplate  = ZettelID("00000000010200")
	ZidListTemplate   = ZettelID("00000000010300")
	ZidZettelTemplate = ZettelID("00000000010401")
	ZidInfoTemplate   = ZettelID("00000000010402")
	ZidFormTemplate   = ZettelID("00000000010403")
	ZidDeleteTemplate = ZettelID("00000000010405")
	ZidErrorTemplate  = ZettelID("00000000010700")

	// WebUI sxn code zettel are in the range 19000..19999
	ZidSxnStart = ZettelID("00000000019000")
	ZidSxnBase  = ZettelID("00000000019990")

	// CSS-related zettel are in the range 20000..29999
	ZidBaseCSS = ZettelID("00000000020001")
	ZidUserCSS = ZettelID("00000000025001")

	// WebUI JS zettel are in the range 30000..39999

	// WebUI image zettel are in the range 40000..49999
	ZidEmoji = ZettelID("00000000040001")

	// Other sxn code zettel are in the range 50000..59999
	ZidSxnPrelude = ZettelID("00000000059900")

	// Predefined Zettelmarkup zettel are in the range 60000..69999
	ZidRoleZettelZettel        = ZettelID("00000000060010")
	ZidRoleConfigurationZettel = ZettelID("00000000060020")
	ZidRoleRoleZettel          = ZettelID("00000000060030")
	ZidRoleTagZettel           = ZettelID("00000000060040")

	// Range 90000...99999 is reserved for zettel templates
	ZidTOCNewTemplate    = ZettelID("00000000090000")
	ZidTemplateNewZettel = ZettelID("00000000090001")
	ZidTemplateNewRole   = ZettelID("00000000090004")
	ZidTemplateNewTag    = ZettelID("00000000090003")
	ZidTemplateNewUser   = ZettelID("00000000090002")

	// Range 00000999999900...00000999999999 are predefined zettel to be searched by content.
	ZidAppDirectory = ZettelID("00000999999999")

	// Default Home Zettel
	ZidDefaultHome = ZettelID("00010000000000")
)

// LengthZid factors the constant length of a zettel identifier
const LengthZid = len(ZidDefaultHome)

// Values of the metadata key/value types.
//
// See [Supported Key Types].
//
// [Supported Key Types]: https://zettelstore.de/manual/h/00001006030000
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

// Predefined / supported metadata keys.
//
// See [Supported Metadata Keys].
//
// [Supported Metadata Keys]: https://zettelstore.de/manual/h/00001006020000
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
	KeySubordinates = "subordinates"
	KeySuccessors   = "successors"
	KeySummary      = "summary"
	KeySuperior     = "superior"
	KeyURL          = "url"
	KeyUselessFiles = "useless-files"
	KeyUserID       = "user-id"
	KeyUserRole     = "user-role"
	KeyVisibility   = "visibility"
)

// Predefined metadata values.
const (
	ValueFalse             = "false"
	ValueTrue              = "true"
	ValueLangEN            = "en"            // Default for "lang"
	ValueRoleConfiguration = "configuration" // A role for internal zettel
	ValueRoleTag           = "tag"           // A role for tag zettel
	ValueRoleRole          = "role"          // A role for role zettel
	ValueRoleZettel        = "zettel"        // A role for zettel
	ValueSyntaxCSS         = "css"           // Syntax: CSS
	ValueSyntaxDraw        = "draw"          // Syntax: Drawing
	ValueSyntaxGif         = "gif"           // Syntax GIF image
	ValueSyntaxHTML        = "html"          // Syntax: HTML
	ValueSyntaxMarkdown    = "markdown"      // Syntax: Markdown / CommonMark
	ValueSyntaxMD          = "md"            // Syntax: Markdown / CommonMark
	ValueSyntaxNone        = "none"          // Syntax: no syntax / content, just metadata
	ValueSyntaxSVG         = "svg"           // Syntax: SVG
	ValueSyntaxSxn         = "sxn"           // Syntax: S-Expression
	ValueSyntaxText        = "text"          // Syntax: plain text
	ValueSyntaxZmk         = "zmk"           // Syntax: Zettelmarkup
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
	BackwardDirective = "BACKWARD" // Backward-only context
	ContextDirective  = "CONTEXT"  // Context directive
	CostDirective     = "COST"     // Maximum cost of a context operation
	ForwardDirective  = "FORWARD"  // Forward-only context
	FullDirective     = "FULL"     // Include tags in context
	IdentDirective    = "IDENT"    // Use only specified zettel
	ItemsDirective    = "ITEMS"    // Select list elements in a zettel
	MaxDirective      = "MAX"      // Maximum number of context results
	LimitDirective    = "LIMIT"    // Maximum number of zettel
	OffsetDirective   = "OFFSET"   // Offset to start returned zettel list
	OrDirective       = "OR"       // Combine several search expression with an "or"
	OrderDirective    = "ORDER"    // Specify metadata keys for the order of returned list
	PhraseDirective   = "PHRASE"   // Only unlinked zettel with given phrase
	PickDirective     = "PICK"     // Pick some random zettel
	RandomDirective   = "RANDOM"   // Order zettel list randomly
	ReverseDirective  = "REVERSE"  // Reverse the order of a zettel list
	UnlinkedDirective = "UNLINKED" // Search for zettel that contain a phase(s) but do not link

	ActionSeparator = "|" // Separates action list of previous elements of query expression

	AtomAction     = "ATOM"     // Return an Atom web feed
	KeysAction     = "KEYS"     // Provide metadata key used
	MinAction      = "MIN"      // Return only those values with a minimum amount of zettel
	MaxAction      = "MAX"      // Return only those values with a maximum amount of zettel
	NumberedAction = "NUMBERED" // Return a numbered list
	RedirectAction = "REDIRECT" // Return the first zettel in list
	ReIndexAction  = "REINDEX"  // Ensure that zettel is/are indexed.
	RSSAction      = "RSS"      // Return a RSS web feed
	TitleAction    = "TITLE"    // Set a title for Atom or RSS web feed

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
