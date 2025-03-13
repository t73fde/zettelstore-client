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
)

// MakeReference builds a reference node.
func MakeReference(sym *sx.Symbol, val string) *sx.Pair {
	return sx.Cons(sym, sx.Cons(sx.MakeString(val), sx.Nil()))
}

// GetReference returns the reference symbol and value.
func GetReference(ref *sx.Pair) (*sx.Symbol, string) {
	if ref != nil {
		if sym, isSymbol := sx.GetSymbol(ref.Car()); isSymbol {
			val, isString := sx.GetString(ref.Cdr())
			if !isString {
				val, isString = sx.GetString(ref.Tail().Car())
			}
			if isString {
				return sym, val.GetValue()
			}
		}
	}
	return nil, ""
}

// ScanReference scans a string and returns a reference.
//
// This function is very specific for Zettelstore.
func ScanReference(s string) *sx.Pair {
	if invalidReference(s) {
		return MakeReference(SymRefStateInvalid, s)
	}
	if strings.HasPrefix(s, api.QueryPrefix) {
		return MakeReference(SymRefStateQuery, s[len(api.QueryPrefix):])
	}
	if state, ok := localState(s); ok {
		if state.IsEqualSymbol(SymRefStateBased) {
			s = s[1:]
		}
		_, err := url.Parse(s)
		if err == nil {
			return MakeReference(state, s)
		}
	}
	u, err := url.Parse(s)
	if err != nil {
		return MakeReference(SymRefStateInvalid, s)
	}
	if !externalURL(u) {
		if _, err = id.Parse(u.Path); err == nil {
			return MakeReference(SymRefStateZettel, s)
		}
		if u.Path == "" && u.Fragment != "" {
			return MakeReference(SymRefStateSelf, s)
		}
	}
	return MakeReference(SymRefStateExternal, s)
}

func invalidReference(s string) bool { return s == "" || s == "00000000000000" }

func externalURL(u *url.URL) bool {
	return u.Scheme != "" || u.Opaque != "" || u.Host != "" || u.User != nil
}

func localState(path string) (*sx.Symbol, bool) {
	if len(path) > 0 && path[0] == '/' {
		if len(path) > 1 && path[1] == '/' {
			return SymRefStateBased, true
		}
		return SymRefStateHosted, true
	}
	if len(path) > 1 && path[0] == '.' {
		if len(path) > 2 && path[1] == '.' && path[2] == '/' {
			return SymRefStateHosted, true
		}
		return SymRefStateHosted, path[1] == '/'
	}
	return SymRefStateInvalid, false
}
