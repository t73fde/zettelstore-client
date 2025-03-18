//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore Client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package meta

import (
	"fmt"
	"iter"
	"slices"
	"strings"
	"time"

	zeroiter "t73f.de/r/zero/iter"
	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsx/input"
)

// Value ist a single metadata value.
type Value string

// AsBool returns the value interpreted as a bool.
func (val Value) AsBool() bool {
	if len(val) > 0 {
		switch val[0] {
		case '0', 'f', 'F', 'n', 'N':
			return false
		}
	}
	return true
}

// AsTime returns the time value of the given value.
func (val Value) AsTime() (time.Time, bool) {
	if t, err := time.Parse(id.TimestampLayout, ExpandTimestamp(val)); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// ExpandTimestamp makes a short-form timestamp larger.
func ExpandTimestamp(val Value) string {
	switch l := len(val); l {
	case 4: // YYYY
		return string(val) + "0101000000"
	case 6: // YYYYMM
		return string(val) + "01000000"
	case 8, 10, 12: // YYYYMMDD, YYYYMMDDhh, YYYYMMDDhhmm
		return string(val) + "000000"[:14-l]
	case 14: // YYYYMMDDhhmmss
		return string(val)
	default:
		if l > 14 {
			return string(val[:14])
		}
		return string(val)
	}
}

// Fields iterates over the value as a list/set of strings.
func (val Value) Fields() iter.Seq[string] {
	return strings.FieldsSeq(string(val))
}

// Elems iterates over the value as a list/set of values.
func (val Value) Elems() iter.Seq[Value] {
	return zeroiter.MapSeq(val.Fields(), func(s string) Value { return Value(s) })
}

// AsSlice transforms a value into a slice of strings.
func (val Value) AsSlice() []string {
	return strings.Fields(string(val))
}

// ToLower maps the value to lowercase runes.
func (val Value) ToLower() Value { return Value(strings.ToLower(string(val))) }

// TrimSpace removes all leading and remaining space from value
func (val Value) TrimSpace() Value {
	return Value(strings.TrimFunc(string(val), input.IsSpace))
}

// AsTags returns the value as a sequence of normalized tags.
func (val Value) AsTags() []string {
	return slices.Collect(zeroiter.MapSeq(
		val.Fields(),
		func(e string) string { return string(Value(e).ToLower().CleanTag()) }))
}

// CleanTag removes the number character ('#') from a tag value.
func (val Value) CleanTag() Value {
	if len(val) > 1 && val[0] == '#' {
		return val[1:]
	}
	return val
}

// NormalizeTag adds a missing prefix "#" to the tag
func (val Value) NormalizeTag() Value {
	if len(val) > 0 && val[0] == '#' {
		return val
	}
	return "#" + val
}

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
	ValueSyntaxGif         = "gif"           // Syntax: GIF image
	ValueSyntaxHTML        = "html"          // Syntax: HTML
	ValueSyntaxJPEG        = "jpeg"          // Syntax: JPEG image
	ValueSyntaxJPG         = "jpg"           // Syntax: PEG image
	ValueSyntaxMarkdown    = "markdown"      // Syntax: Markdown / CommonMark
	ValueSyntaxMD          = "md"            // Syntax: Markdown / CommonMark
	ValueSyntaxNone        = "none"          // Syntax: no syntax / content, just metadata
	ValueSyntaxPlain       = "plain"         // Syntax: plain text
	ValueSyntaxPNG         = "png"           // Syntax: PNG image
	ValueSyntaxSVG         = "svg"           // Syntax: SVG
	ValueSyntaxSxn         = "sxn"           // Syntax: S-Expression
	ValueSyntaxText        = "text"          // Syntax: plain text
	ValueSyntaxTxt         = "txt"           // Syntax: plain text
	ValueSyntaxWebp        = "webp"          // Syntax: WEBP image
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

// DefaultSyntax is the default value for metadata 'syntax'.
const DefaultSyntax = ValueSyntaxPlain

// Visibility enumerates the variations of the 'visibility' meta key.
type Visibility int

// Supported values for visibility.
const (
	_ Visibility = iota
	VisibilityUnknown
	VisibilityPublic
	VisibilityCreator
	VisibilityLogin
	VisibilityOwner
	VisibilityExpert
)

var visMap = map[Value]Visibility{
	ValueVisibilityPublic:  VisibilityPublic,
	ValueVisibilityCreator: VisibilityCreator,
	ValueVisibilityLogin:   VisibilityLogin,
	ValueVisibilityOwner:   VisibilityOwner,
	ValueVisibilityExpert:  VisibilityExpert,
}
var revVisMap = map[Visibility]Value{}

func init() {
	for k, v := range visMap {
		revVisMap[v] = k
	}
}

// AsVisibility returns the visibility value of the given value string
func (val Value) AsVisibility() Visibility {
	if vis, ok := visMap[val]; ok {
		return vis
	}
	return VisibilityUnknown
}

func (v Visibility) String() string {
	if s, ok := revVisMap[v]; ok {
		return string(s)
	}
	return fmt.Sprintf("Unknown (%d)", v)
}

// UserRole enumerates the supported values of meta key 'user-role'.
type UserRole int

// Supported values for user roles.
const (
	_ UserRole = iota
	UserRoleUnknown
	UserRoleCreator
	UserRoleReader
	UserRoleWriter
	UserRoleOwner
)

var urMap = map[Value]UserRole{
	ValueUserRoleCreator: UserRoleCreator,
	ValueUserRoleReader:  UserRoleReader,
	ValueUserRoleWriter:  UserRoleWriter,
	ValueUserRoleOwner:   UserRoleOwner,
}

// AsUserRole role returns the user role of the given string.
func (val Value) AsUserRole() UserRole {
	if ur, ok := urMap[val]; ok {
		return ur
	}
	return UserRoleUnknown
}
