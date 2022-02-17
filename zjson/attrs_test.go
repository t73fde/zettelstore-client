//-----------------------------------------------------------------------------
// Copyright (c) 2020-2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package zjson_test

import (
	"testing"

	"zettelstore.de/c/zjson"
)

func TestHasDefault(t *testing.T) {
	t.Parallel()
	attr := zjson.Attributes{}
	if attr.HasDefault() {
		t.Error("Should not have default attr")
	}
	attr = zjson.Attributes(map[string]string{"-": "value"})
	if !attr.HasDefault() {
		t.Error("Should have default attr")
	}
}

func TestAttrClone(t *testing.T) {
	t.Parallel()
	orig := zjson.Attributes{}
	clone := orig.Clone()
	if !clone.IsEmpty() {
		t.Error("Attrs must be empty")
	}

	orig = zjson.Attributes(map[string]string{"": "0", "-": "1", "a": "b"})
	clone = orig.Clone()
	if clone[""] != "0" || clone["-"] != "1" || clone["a"] != "b" || len(clone) != len(orig) {
		t.Error("Wrong cloned map")
	}
	clone["a"] = "c"
	if orig["a"] != "b" {
		t.Error("Aliased map")
	}
}
