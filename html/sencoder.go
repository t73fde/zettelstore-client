//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package html

import "io"

// EncEnvironment represent the encoding environment.
type EncEnvironment struct{}

func NewEncEnvironment(io.Writer, int) *EncEnvironment {
	return &EncEnvironment{}
}

// GetError returns the first encountered error during encoding.
func (env *EncEnvironment) GetError() error { return nil }

func (env *EncEnvironment) WriteEndnotes() {}
