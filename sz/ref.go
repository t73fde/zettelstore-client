//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import (
	"net/url"
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsx"
)

// ScanReference scans a string and returns a reference.
//
// This function is very specific for Zettelstore.
func ScanReference(s string) *sx.Pair {
	if len(s) == id.LengthZid {
		if _, err := id.Parse(s); err == nil {
			return zsx.MakeReference(SymRefStateZettel, s)
		}
		if s == "00000000000000" {
			return zsx.MakeReference(zsx.SymRefStateInvalid, s)
		}
	} else if len(s) > id.LengthZid && s[id.LengthZid] == '#' {
		zidPart := s[:id.LengthZid]
		if _, err := id.Parse(zidPart); err == nil {
			if u, err2 := url.Parse(s); err2 != nil || u.String() != s {
				return zsx.MakeReference(zsx.SymRefStateInvalid, s)
			}
			return zsx.MakeReference(SymRefStateZettel, s)
		}
		if zidPart == "00000000000000" {
			return zsx.MakeReference(zsx.SymRefStateInvalid, s)
		}
	}
	if strings.HasPrefix(s, api.QueryPrefix) {
		return zsx.MakeReference(SymRefStateQuery, s[len(api.QueryPrefix):])
	}
	if strings.HasPrefix(s, "//") {
		if u, err := url.Parse(s[1:]); err == nil {
			if u.Scheme == "" && u.Opaque == "" && u.Host == "" && u.User == nil {
				if u.String() == s[1:] {
					return zsx.MakeReference(SymRefStateBased, s[1:])
				}
				return zsx.MakeReference(zsx.SymRefStateInvalid, s)
			}
		}
	}

	if s == "" {
		return zsx.MakeReference(zsx.SymRefStateInvalid, s)
	}
	u, err := url.Parse(s)
	if err != nil || u.String() != s {
		return zsx.MakeReference(zsx.SymRefStateInvalid, s)
	}
	sym := zsx.SymRefStateExternal
	if u.Scheme == "" && u.Opaque == "" && u.Host == "" && u.User == nil {
		if s[0] == '#' {
			sym = zsx.SymRefStateSelf
		} else {
			sym = zsx.SymRefStateHosted
		}
	}
	return zsx.MakeReference(sym, s)
}
