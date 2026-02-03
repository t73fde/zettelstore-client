//-----------------------------------------------------------------------------
// Copyright (c) 2026-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2026-present Detlef Stern
//-----------------------------------------------------------------------------

package zmk

import "t73f.de/r/zero/set"

// EscapeZmkSyntax returns a string, where all syntax elements of Zettelmarkup
// are escaped with a backslash.
func EscapeZmkSyntax(s string) string {
	if !containsZmkSyntax(s) {
		return s
	}
	result := make([]rune, 0, len(s)*2)
	var lastCh rune
	runLength := 1
	for _, ch := range s {
		if ch == lastCh {
			runLength++
		} else {
			runLength = 1
		}
		if zmkSyntaxChars.Contains(ch) && runLength%2 == 0 {
			result = append(result, '\\')
		}
		result = append(result, ch)
		lastCh = ch
	}
	return string(result)
}

func containsZmkSyntax(s string) bool {
	for _, ch := range s {
		if zmkSyntaxChars.Contains(ch) {
			return true
		}
	}
	return false
}

var zmkSyntaxChars = set.New(
	'"', '#', '%', '&', '\'', '*', ',', '-', ':', ';', '<', '=', '>', '@',
	'[', '\\', ']', '^', '_', '`', '{', '|', '}', '~',
)
