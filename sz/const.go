//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2022-present Detlef Stern
//-----------------------------------------------------------------------------

package sz

import "t73f.de/r/sx"

// Various constants for Zettel data. They are technically variables.
// These are only Zettelstore-specific symbols. The more general symbols are
// defined in t73f.de/r/zsx
var (
	// Symbols for Metanodes
	SymMeta = sx.MakeSymbol("META")

	// Constant symbols for reference states.
	SymRefStateZettel = sx.MakeSymbol("ZETTEL")
	SymRefStateFound  = sx.MakeSymbol("FOUND")
	SymRefStateBroken = sx.MakeSymbol("BROKEN")
	SymRefStateBased  = sx.MakeSymbol("BASED")
	SymRefStateQuery  = sx.MakeSymbol("QUERY")

	// Symbols for metadata types.
	SymTypeCredential = sx.MakeSymbol("CREDENTIAL")
	SymTypeEmpty      = sx.MakeSymbol("EMPTY-STRING")
	SymTypeID         = sx.MakeSymbol("ZID")
	SymTypeIDSet      = sx.MakeSymbol("ZID-SET")
	SymTypeNumber     = sx.MakeSymbol("NUMBER")
	SymTypeString     = sx.MakeSymbol("STRING")
	SymTypeTagSet     = sx.MakeSymbol("TAG-SET")
	SymTypeTimestamp  = sx.MakeSymbol("TIMESTAMP")
	SymTypeURL        = sx.MakeSymbol("URL")
	SymTypeWord       = sx.MakeSymbol("WORD")
)
