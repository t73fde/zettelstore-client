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

// Package id provides zettel specific types, constants, and functions about
// zettel identifier.
package id

import (
	"strconv"
	"time"
)

// Zid is the internal identifier of a zettel. Typically, it is a time stamp
// of the form "YYYYMMDDHHmmSS" converted to an unsigned integer.
type Zid uint64

// LengthZid factors the constant length of a zettel identifier
const LengthZid = 14

// Some important ZettelIDs.
const (
	Invalid = Zid(0) // Invalid is a Zid that will never be valid

	maxZid = 99999999999999
)

// Predefined zettel identifier.
//
// See [List of predefined zettel].
//
// [List of predefined zettel]: https://zettelstore.de/manual/h/00001005090000
const (
	// System zettel
	ZidVersion              = Zid(1)
	ZidHost                 = Zid(2)
	ZidOperatingSystem      = Zid(3)
	ZidLicense              = Zid(4)
	ZidAuthors              = Zid(5)
	ZidDependencies         = Zid(6)
	ZidModules              = Zid(7)
	ZidLog                  = Zid(9)
	ZidMemory               = Zid(10)
	ZidSx                   = Zid(11)
	ZidHTTP                 = Zid(12)
	ZidAPI                  = Zid(13)
	ZidWebUI                = Zid(14)
	ZidConsole              = Zid(15)
	ZidBoxManager           = Zid(20)
	ZidZettel               = Zid(21)
	ZidIndex                = Zid(22)
	ZidQuery                = Zid(23)
	ZidMetadataKey          = Zid(90)
	ZidParser               = Zid(92)
	ZidStartupConfiguration = Zid(96)
	ZidConfiguration        = Zid(100)
	ZidDirectory            = Zid(101)

	// WebUI HTML templates are in the range 10000..19999
	ZidBaseTemplate   = Zid(10100)
	ZidLoginTemplate  = Zid(10200)
	ZidListTemplate   = Zid(10300)
	ZidZettelTemplate = Zid(10401)
	ZidInfoTemplate   = Zid(10402)
	ZidFormTemplate   = Zid(10403)
	ZidDeleteTemplate = Zid(10405)
	ZidErrorTemplate  = Zid(10700)

	// WebUI sxn code zettel are in the range 19000..19999
	ZidSxnStart = Zid(19000)
	ZidSxnBase  = Zid(19990)

	// CSS-related zettel are in the range 20000..29999
	ZidBaseCSS = Zid(20001)
	ZidUserCSS = Zid(25001)

	// WebUI JS zettel are in the range 30000..39999

	// WebUI image zettel are in the range 40000..49999
	ZidEmoji = Zid(40001)

	// Other sxn code zettel are in the range 50000..59999
	ZidSxnPrelude = Zid(59900)

	// Predefined Zettelmarkup zettel are in the range 60000..69999
	ZidRoleZettelZettel        = Zid(60010)
	ZidRoleConfigurationZettel = Zid(60020)
	ZidRoleRoleZettel          = Zid(60030)
	ZidRoleTagZettel           = Zid(60040)

	// Range 80000...89999 is reserved for web ui menus
	ZidTOCListsMenu = Zid(80001) // "Lists" menu

	// Range 90000...99999 is reserved for zettel templates
	ZidTOCNewTemplate    = Zid(90000)
	ZidTemplateNewZettel = Zid(90001)
	ZidTemplateNewRole   = Zid(90004)
	ZidTemplateNewTag    = Zid(90003)
	ZidTemplateNewUser   = Zid(90002)

	// Range 00000999999900...00000999999999 are predefined zettel to be searched by content.
	ZidAppDirectory = Zid(999999999)

	// Default Home Zettel
	ZidDefaultHome = Zid(10000000000)
)

// ParseUint interprets a string as a possible zettel identifier
// and returns its integer value.
func ParseUint(s string) (uint64, error) {
	res, err := strconv.ParseUint(s, 10, 47)
	if err != nil {
		return 0, err
	}
	if res == 0 || res > maxZid {
		return res, strconv.ErrRange
	}
	return res, nil
}

// Parse interprets a string as a zettel identification and
// returns its value.
func Parse(s string) (Zid, error) {
	if len(s) != LengthZid {
		return Invalid, strconv.ErrSyntax
	}
	res, err := ParseUint(s)
	if err != nil {
		return Invalid, err
	}
	return Zid(res), nil
}

// MustParse tries to interpret a string as a zettel identifier and returns
// its value or panics otherwise.
func MustParse(s string) Zid {
	zid, err := Parse(string(s))
	if err == nil {
		return zid
	}
	panic(err)
}

// String converts the zettel identification to a string of 14 digits.
// Only defined for valid ids.
func (zid Zid) String() string {
	var result [LengthZid]byte
	zid.toByteArray(&result)
	return string(result[:])
}

// Bytes converts the zettel identification to a byte slice of 14 digits.
// Only defined for valid ids.
func (zid Zid) Bytes() []byte {
	var result [LengthZid]byte
	zid.toByteArray(&result)
	return result[:]
}

// toByteArray converts the Zid into a fixed byte array, usable for printing.
//
// Based on idea by Daniel Lemire: "Converting integers to fix-digit representations quickly"
// https://lemire.me/blog/2021/11/18/converting-integers-to-fix-digit-representations-quickly/
func (zid Zid) toByteArray(result *[LengthZid]byte) {
	date := uint64(zid) / 1000000
	fullyear := date / 10000
	century, year := fullyear/100, fullyear%100
	monthday := date % 10000
	month, day := monthday/100, monthday%100
	time := uint64(zid) % 1000000
	hmtime, second := time/100, time%100
	hour, minute := hmtime/100, hmtime%100

	result[0] = byte(century/10) + '0'
	result[1] = byte(century%10) + '0'
	result[2] = byte(year/10) + '0'
	result[3] = byte(year%10) + '0'
	result[4] = byte(month/10) + '0'
	result[5] = byte(month%10) + '0'
	result[6] = byte(day/10) + '0'
	result[7] = byte(day%10) + '0'
	result[8] = byte(hour/10) + '0'
	result[9] = byte(hour%10) + '0'
	result[10] = byte(minute/10) + '0'
	result[11] = byte(minute%10) + '0'
	result[12] = byte(second/10) + '0'
	result[13] = byte(second%10) + '0'
}

// IsValid determines if zettel id is a valid one, e.g. consists of max. 14 digits.
func (zid Zid) IsValid() bool { return 0 < zid && zid <= maxZid }

// TimestampLayout to transform a date into a Zid and into other internal dates.
const TimestampLayout = "20060102150405"

// New returns a new zettel id based on the current time.
func New(withSeconds bool) Zid {
	now := time.Now().Local()
	var s string
	if withSeconds {
		s = now.Format(TimestampLayout)
	} else {
		s = now.Format("20060102150400")
	}
	res, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return res
}
