// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------

package zmk

import (
	"net/url"
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/sz"
)

// ParseReference parses a string and returns a reference.
func ParseReference(s string) *sx.Pair {
	if invalidReference(s) {
		return makePairRef(sz.SymRefStateInvalid, s)
	}
	if strings.HasPrefix(s, api.QueryPrefix) {
		return makePairRef(sz.SymRefStateQuery, s[len(api.QueryPrefix):])
	}
	if state, ok := localState(s); ok {
		if state.IsEqualSymbol(sz.SymRefStateBased) {
			s = s[1:]
		}
		_, err := url.Parse(s)
		if err == nil {
			return makePairRef(state, s)
		}
	}
	u, err := url.Parse(s)
	if err != nil {
		return makePairRef(sz.SymRefStateInvalid, s)
	}
	if !externalURL(u) {
		if _, err := id.Parse(u.Path); err == nil {
			return makePairRef(sz.SymRefStateZettel, s)
		}
		if u.Path == "" && u.Fragment != "" {
			return makePairRef(sz.SymRefStateSelf, s)
		}
	}
	return makePairRef(sz.SymRefStateExternal, s)
}
func makePairRef(sym *sx.Symbol, val string) *sx.Pair {
	return sx.MakeList(sym, sx.MakeString(val))
}

func invalidReference(s string) bool { return s == "" || s == "00000000000000" }

func externalURL(u *url.URL) bool {
	return u.Scheme != "" || u.Opaque != "" || u.Host != "" || u.User != nil
}

func localState(path string) (*sx.Symbol, bool) {
	if len(path) > 0 && path[0] == '/' {
		if len(path) > 1 && path[1] == '/' {
			return sz.SymRefStateBased, true
		}
		return sz.SymRefStateHosted, true
	}
	if len(path) > 1 && path[0] == '.' {
		if len(path) > 2 && path[1] == '.' && path[2] == '/' {
			return sz.SymRefStateHosted, true
		}
		return sz.SymRefStateHosted, path[1] == '/'
	}
	return sz.SymRefStateInvalid, false
}

// ReferenceIsValid returns true if reference is valid
func ReferenceIsValid(ref *sx.Pair) bool {
	return !ref.Car().IsEqual(sz.SymRefStateInvalid)
}

// ReferenceIsZettel returns true if it is a reference to a local zettel.
func ReferenceIsZettel(ref *sx.Pair) bool {
	state := ref.Car()
	return state.IsEqual(sz.SymRefStateZettel) ||
		state.IsEqual(sz.SymRefStateSelf) ||
		state.IsEqual(sz.SymRefStateFound) ||
		state.IsEqual(sz.SymRefStateBroken)
}

// ReferenceIsLocal returns true if reference is local
func ReferenceIsLocal(ref *sx.Pair) bool {
	state := ref.Car()
	return state.IsEqual(sz.SymRefStateHosted) ||
		state.IsEqual(sz.SymRefStateBased)
}

// ReferenceIsExternal returns true if it is a reference to external material.
func ReferenceIsExternal(ref *sx.Pair) bool {
	return ref.Car().IsEqual(sz.SymRefStateExternal)
}
