//-----------------------------------------------------------------------------
// Copyright (c) 2021-2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package api contains common definitions used for client and server.
package api

// ZettelID contains the identifier of a zettel. It is a string with 14 digits.
type ZettelID string

// InvalidZID is an invalif zettel identifier
const InvalidZID = ""

// IsValid returns true, if the idenfifier contains 14 digits.
func (zid ZettelID) IsValid() bool {
	if len(zid) != 14 {
		return false
	}
	for i := 0; i < 14; i++ {
		ch := zid[i]
		if ch < '0' || '9' < ch {
			return false
		}
	}
	return true
}

// ZettelMeta is a map containg the metadata of a zettel.
type ZettelMeta map[string]string

// ZettelRights is an integer that encode access rights for a zettel.
type ZettelRights uint8

// Values for ZettelRights, can be or-ed
const (
	ZettelCanNone   ZettelRights = 1 << iota
	ZettelCanCreate              // Current user is allowed to create a new zettel
	ZettelCanRead                // Requesting user is allowed to read the zettel
	ZettelCanWrite               // Requesting user is allowed to update the zettel
	ZettelCanRename              // Requesting user is allowed to provide the zettel with a new identifier
	ZettelCanDelete              // Requesting user is allowed to delete the zettel
)

// AuthJSON contains the result of an authentication call.
type AuthJSON struct {
	Token   string `json:"token"`
	Type    string `json:"token_type"`
	Expires int    `json:"expires_in"`
}

// ZidJSON contains the identifier data of a zettel.
type ZidJSON struct {
	ID ZettelID `json:"id"`
}

// MetaJSON contains the metadata of a zettel.
type MetaJSON struct {
	Meta   ZettelMeta   `json:"meta"`
	Rights ZettelRights `json:"rights"`
}

// ZidMetaJSON contains the identifier and the metadata of a zettel.
type ZidMetaJSON struct {
	ID     ZettelID     `json:"id"`
	Meta   ZettelMeta   `json:"meta"`
	Rights ZettelRights `json:"rights"`
}

// ZidMetaRelatedList contains identifier/metadata of a zettel and the same for related zettel
type ZidMetaRelatedList struct {
	ID     ZettelID      `json:"id"`
	Meta   ZettelMeta    `json:"meta"`
	Rights ZettelRights  `json:"rights"`
	List   []ZidMetaJSON `json:"list"`
}

// ZettelLinksJSON store all links / connections from one zettel to other.
type ZettelLinksJSON struct {
	ID     ZettelID `json:"id"`
	Linked struct {
		Outgoing []string `json:"outgoing,omitempty"`
		Local    []string `json:"local,omitempty"`
		External []string `json:"external,omitempty"`
		Meta     []string `json:"meta,omitempty"`
	} `json:"linked"`
	Embedded struct {
		Outgoing []string `json:"outgoing,omitempty"`
		Local    []string `json:"local,omitempty"`
		External []string `json:"external,omitempty"`
	} `json:"embedded,omitempty"`
	Cites []string `json:"cites,omitempty"`
}

// ZettelDataJSON contains all data for a zettel.
type ZettelDataJSON struct {
	Meta     ZettelMeta `json:"meta"`
	Encoding string     `json:"encoding"`
	Content  string     `json:"content"`
}

// ZettelJSON contains all data for a zettel, the identifier, the metadata, and the content.
type ZettelJSON struct {
	ID       ZettelID     `json:"id"`
	Meta     ZettelMeta   `json:"meta"`
	Encoding string       `json:"encoding"`
	Content  string       `json:"content"`
	Rights   ZettelRights `json:"rights"`
}

// ZettelListJSON contains data for a zettel list.
type ZettelListJSON struct {
	Query string        `json:"query"`
	List  []ZidMetaJSON `json:"list"`
}

// TagListJSON specifies the list/map of tags
type TagListJSON struct {
	Tags map[string][]ZettelID `json:"tags"`
}

// RoleListJSON specifies the list of roles.
type RoleListJSON struct {
	Roles []string `json:"role-list"`
}
