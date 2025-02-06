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
	"strings"
	"time"

	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/domain/id"
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

// AsList transforms a value into a list value.
func (val Value) AsList() []string {
	return strings.Fields(string(val))
}

// ToLower maps the value to lowercase runes.
func (val Value) ToLower() Value { return Value(strings.ToLower(string(val))) }

// AsTags returns the value as a sequence of normalized tags.
func (val Value) AsTags() []string {
	tags := val.ToLower().AsList()
	for i, tag := range tags {
		if len(tag) > 1 && tag[0] == '#' {
			tags[i] = tag[1:]
		}
	}
	return tags
}

// CleanTag removes the number character ('#') from a tag value and lowercases it.
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

// Supported syntax values.
const (
	SyntaxCSS      = api.ValueSyntaxCSS
	SyntaxDraw     = api.ValueSyntaxDraw
	SyntaxGif      = api.ValueSyntaxGif
	SyntaxHTML     = api.ValueSyntaxHTML
	SyntaxJPEG     = "jpeg"
	SyntaxJPG      = "jpg"
	SyntaxMarkdown = api.ValueSyntaxMarkdown
	SyntaxMD       = api.ValueSyntaxMD
	SyntaxNone     = api.ValueSyntaxNone
	SyntaxPlain    = "plain"
	SyntaxPNG      = "png"
	SyntaxSVG      = api.ValueSyntaxSVG
	SyntaxSxn      = api.ValueSyntaxSxn
	SyntaxText     = api.ValueSyntaxText
	SyntaxTxt      = "txt"
	SyntaxWebp     = "webp"
	SyntaxZmk      = api.ValueSyntaxZmk

	DefaultSyntax = SyntaxPlain
)

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
	api.ValueVisibilityPublic:  VisibilityPublic,
	api.ValueVisibilityCreator: VisibilityCreator,
	api.ValueVisibilityLogin:   VisibilityLogin,
	api.ValueVisibilityOwner:   VisibilityOwner,
	api.ValueVisibilityExpert:  VisibilityExpert,
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
	api.ValueUserRoleCreator: UserRoleCreator,
	api.ValueUserRoleReader:  UserRoleReader,
	api.ValueUserRoleWriter:  UserRoleWriter,
	api.ValueUserRoleOwner:   UserRoleOwner,
}

// AsUserRole role returns the user role of the given string.
func (val Value) AsUserRole() UserRole {
	if ur, ok := urMap[val]; ok {
		return ur
	}
	return UserRoleUnknown
}
