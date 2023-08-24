//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package api

import "fmt"

// Predefined Zettel Identifier
const (
	// System zettel
	ZidVersion              = ZettelID("00000000000001")
	ZidHost                 = ZettelID("00000000000002")
	ZidOperatingSystem      = ZettelID("00000000000003")
	ZidLicense              = ZettelID("00000000000004")
	ZidAuthors              = ZettelID("00000000000005")
	ZidDependencies         = ZettelID("00000000000006")
	ZidLog                  = ZettelID("00000000000007")
	ZidBoxManager           = ZettelID("00000000000020")
	ZidMetadataKey          = ZettelID("00000000000090")
	ZidParser               = ZettelID("00000000000092")
	ZidStartupConfiguration = ZettelID("00000000000096")
	ZidConfiguration        = ZettelID("00000000000100")

	// WebUI HTML templates are in the range 10000..19999
	ZidBaseTemplate   = ZettelID("00000000010100")
	ZidLoginTemplate  = ZettelID("00000000010200")
	ZidListTemplate   = ZettelID("00000000010300")
	ZidZettelTemplate = ZettelID("00000000010401")
	ZidInfoTemplate   = ZettelID("00000000010402")
	ZidFormTemplate   = ZettelID("00000000010403")
	ZidRenameTemplate = ZettelID("00000000010404")
	ZidDeleteTemplate = ZettelID("00000000010405")
	ZidErrorTemplate  = ZettelID("00000000010700")

	// WebUI sxn code zettel are in the range 19000..19999
	ZidSxnStart = ZettelID("00000000019000")
	ZidSxnBase  = ZettelID("00000000019990")

	// CSS-related zettel are in the range 20000..29999
	ZidBaseCSS    = ZettelID("00000000020001")
	ZidUserCSS    = ZettelID("00000000025001")
	ZidRoleCSSMap = ZettelID("00000000029000") // Maps roles to CSS zettel, which should be in the range 29001..29999.

	// WebUI JS zettel are in the range 30000..39999

	// WebUI image zettel are in the range 40000..49999
	ZidEmoji = ZettelID("00000000040001")

	// Other sxn code zettel are in the range 50000..59999

	// Range 90000...99999 is reserved for zettel templates
	ZidTOCNewTemplate    = ZettelID("00000000090000")
	ZidTemplateNewZettel = ZettelID("00000000090001")
	ZidTemplateNewUser   = ZettelID("00000000090002")

	ZidDefaultHome = ZettelID("00010000000000")
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
	MetaWordSet      = "WordSet"
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
	KeyPublished    = "published"
	KeyQuery        = "query"
	KeyReadOnly     = "read-only"
	KeySubordinates = "subordinates"
	KeySuccessors   = "successors"
	KeySuperior     = "superior"
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
	MethodMove = "MOVE" // HTTP method for renaming a zettel

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
	QueryKeySeed      = "_seed"
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
