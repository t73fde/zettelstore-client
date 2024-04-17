//-----------------------------------------------------------------------------
// Copyright (c) 2024-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2024-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import (
	"t73f.de/r/sx"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/input"
)

// --- Contains some simple parsers

// ---- Syntax: none

// ParseNoneBlocks parses no block.
func ParseNoneBlocks(*input.Input) *sx.Pair { return nil }

// ParseNoneInlines skips to the end of line and parses no inline.
func ParseNoneInlines(inp *input.Input) *sx.Pair {
	inp.SkipToEOL()
	return nil
}

// ---- Some plain text syntaxes

// ParsePlainBlocks parses the block as plain text with the given syntax.
func ParsePlainBlocks(inp *input.Input, syntax string) *sx.Pair {
	var sym *sx.Symbol
	if syntax == api.ValueSyntaxHTML {
		sym = SymVerbatimHTML
	} else {
		sym = SymVerbatimProg
	}
	return sx.MakeList(
		sym,
		sx.MakeList(sx.Cons(sx.MakeString(""), sx.MakeString(syntax))),
		sx.MakeString(string(inp.ScanLineContent())),
	)
}

// ParsePlainInlines parses the inline as plain text with the given syntax.
func ParsePlainInlines(inp *input.Input, syntax string) *sx.Pair {
	var sym *sx.Symbol
	if syntax == api.ValueSyntaxHTML {
		sym = SymLiteralHTML
	} else {
		sym = SymLiteralProg
	}
	pos := inp.Pos
	inp.SkipToEOL()
	return sx.MakeList(
		sym,
		sx.MakeList(sx.Cons(sx.MakeString(""), sx.MakeString(syntax))),
		sx.MakeString(string(inp.Src[pos:inp.Pos])),
	)
}
