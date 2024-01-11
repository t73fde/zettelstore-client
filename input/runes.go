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

package input

import "unicode"

// IsSpace returns true if rune is a whitespace.
func IsSpace(ch rune) bool {
	switch ch {
	case ' ', '\t':
		return true
	case '\n', '\r', EOS:
		return false
	}
	return unicode.IsSpace(ch)
}
