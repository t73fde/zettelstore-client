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

package api

import (
	"t73f.de/r/webs/urlbuilder"
	"t73f.de/r/zsc/domain/id"
)

// URLBuilder should be used to create zettelstore URLs.
type URLBuilder struct {
	base   urlbuilder.URLBuilder
	prefix string
}

// NewURLBuilder creates a new URL builder with the given prefix and key.
func NewURLBuilder(prefix string, key byte) *URLBuilder {
	for len(prefix) > 0 && prefix[len(prefix)-1] == '/' {
		prefix = prefix[0 : len(prefix)-1]
	}
	result := URLBuilder{prefix: prefix}
	if key != '/' {
		result.base.AddPath(string([]byte{key}))
	}
	return &result
}

// Clone an URLBuilder.
func (ub *URLBuilder) Clone() *URLBuilder {
	cpy := new(URLBuilder)
	ub.base.Copy(&cpy.base)
	cpy.prefix = ub.prefix
	return cpy
}

// SetZid sets the zettel identifier.
func (ub *URLBuilder) SetZid(zid id.Zid) *URLBuilder {
	ub.base.AddPath(zid.String())
	return ub
}

// AppendPath adds a new path element.
func (ub *URLBuilder) AppendPath(p string) *URLBuilder {
	ub.base.AddPath(p)
	return ub
}

// AppendKVQuery adds a new key/value query parameter.
func (ub *URLBuilder) AppendKVQuery(key, value string) *URLBuilder {
	ub.base.AddQuery(key, value)
	return ub
}

// AppendQuery adds a new query.
//
// Basically the same as [URLBuilder.AppendKVQuery]([api.QueryKeyQuery], value)
func (ub *URLBuilder) AppendQuery(value string) *URLBuilder {
	if value != "" {
		ub.base.AddQuery(QueryKeyQuery, value)
	}
	return ub
}

// ClearQuery removes all query parameters.
func (ub *URLBuilder) ClearQuery() *URLBuilder {
	ub.base.RemoveQueries()
	return ub
}

// SetFragment sets the fragment.
func (ub *URLBuilder) SetFragment(s string) *URLBuilder {
	ub.base.SetFragment(s)
	return ub
}

// String produces a string value.
func (ub *URLBuilder) String() string {
	return ub.prefix + ub.base.String()
}
