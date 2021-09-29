//-----------------------------------------------------------------------------
// Copyright (c) 2021 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package api contains common definition used for client and server.
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
	ZidBoxManager           = ZettelID("00000000000020")
	ZidMetadataKey          = ZettelID("00000000000090")
	ZidStartupConfiguration = ZettelID("00000000000096")
	ZidConfiguration        = ZettelID("00000000000100")

	// WebUI HTML templates are in the range 10000..19999
	ZidBaseTemplate    = ZettelID("00000000010100")
	ZidLoginTemplate   = ZettelID("00000000010200")
	ZidListTemplate    = ZettelID("00000000010300")
	ZidZettelTemplate  = ZettelID("00000000010401")
	ZidInfoTemplate    = ZettelID("00000000010402")
	ZidFormTemplate    = ZettelID("00000000010403")
	ZidRenameTemplate  = ZettelID("00000000010404")
	ZidDeleteTemplate  = ZettelID("00000000010405")
	ZidContextTemplate = ZettelID("00000000010406")
	ZidRolesTemplate   = ZettelID("00000000010500")
	ZidTagsTemplate    = ZettelID("00000000010600")
	ZidErrorTemplate   = ZettelID("00000000010700")

	// WebUI CSS zettel are in the range 20000..29999
	ZidBaseCSS = ZettelID("00000000020001")
	ZidUserCSS = ZettelID("00000000025001")

	// WebUI JS zettel are in the range 30000..39999

	// WebUI image zettel are in the range 40000..49999
	ZidEmoji = ZettelID("00000000040001")

	// Range 90000...99999 is reserved for zettel templates
	ZidTOCNewTemplate    = ZettelID("00000000090000")
	ZidTemplateNewZettel = ZettelID("00000000090001")
	ZidTemplateNewUser   = ZettelID("00000000090002")

	ZidDefaultHome = ZettelID("00010000000000")
)

// Predefined Metadata keys
const (
	KeyRole  = "role"
	KeyTitle = "title"
)

// Predefined Metadata values
const (
	ValueRoleConfiguration = "configuration"
)

// Additional HTTP constants used.
const (
	MethodMove = "MOVE" // HTTP method for renaming a zettel

	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"
	HeaderDestination = "Destination"
	HeaderLocation    = "Location"
)

// Values for HTTP query parameter.
const (
	QueryKeyDepth    = "depth"
	QueryKeyDir      = "dir"
	QueryKeyEncoding = "_enc"
	QueryKeyLimit    = "_limit"
	QueryKeyNegate   = "_negate"
	QueryKeyOffset   = "_offset"
	QueryKeyOrder    = "_order"
	QueryKeyPart     = "_part"
	QueryKeySearch   = "_s"
	QueryKeySort     = "_sort"
)

// Supported dir values.
const (
	DirBackward = "backward"
	DirForward  = "forward"
)

// Supported encoding values.
const (
	EncodingDJSON  = "djson"
	EncodingHTML   = "html"
	EncodingNative = "native"
	EncodingText   = "text"
	EncodingZMK    = "zmk"
)

var mapEncodingEnum = map[string]EncodingEnum{
	EncodingDJSON:  EncoderDJSON,
	EncodingHTML:   EncoderHTML,
	EncodingNative: EncoderNative,
	EncodingText:   EncoderText,
	EncodingZMK:    EncoderZmk,
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
	EncoderDJSON
	EncoderHTML
	EncoderNative
	EncoderText
	EncoderZmk
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
