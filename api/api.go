//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2021-present Detlef Stern
//-----------------------------------------------------------------------------

// Package api contains common definitions used for client and server.
package api

import "t73f.de/r/zsc/domain/id"

// ZettelMeta is a map containg the normalized metadata of a zettel.
type ZettelMeta map[string]string

// ZettelRights is an integer that encode access rights for a zettel.
type ZettelRights uint8

// Values for ZettelRights, can be or-ed
const (
	ZettelCanNone   ZettelRights = 1 << iota
	ZettelCanCreate              // Current user is allowed to create a new zettel
	ZettelCanRead                // Requesting user is allowed to read the zettel
	ZettelCanWrite               // Requesting user is allowed to update the zettel
	placeholdergo1               // Was assigned to rename right, which is now removed
	ZettelCanDelete              // Requesting user is allowed to delete the zettel
	ZettelMaxRight               // Sentinel value
)

// MetaRights contains the metadata of a zettel, and its rights.
type MetaRights struct {
	Meta   ZettelMeta
	Rights ZettelRights
}

// ZidMetaRights contains the identifier, the metadata of a zettel, and its rights.
type ZidMetaRights struct {
	ID     id.Zid
	Meta   ZettelMeta
	Rights ZettelRights
}

// ZettelData contains all data for a zettel.
//
//   - Meta is a map containing the metadata of the zettel.
//   - Rights is an integer specifying the access rights.
//   - Encoding is a string specifying the encoding of the zettel content.
//   - Content is the zettel content itself.
type ZettelData struct {
	Meta     ZettelMeta
	Rights   ZettelRights
	Encoding string
	Content  string // raw, uninterpreted zettel content
}

// Aggregate maps metadata keys to list of zettel identifier.
type Aggregate map[string][]id.Zid
