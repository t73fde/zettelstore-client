//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sx_test

import (
	"testing"

	"zettelstore.de/client.fossil/sx"
	"zettelstore.de/sx.fossil/sxpf"
)

func TestParseObject(t *testing.T) {
	if elems, err := sx.ParseObject(sxpf.MakeString("a"), "s"); err == nil {
		t.Error("expected an error, but got: ", elems)
	}
	if elems, err := sx.ParseObject(sxpf.Nil(), ""); err != nil {
		t.Error(err)
	} else if len(elems) != 0 {
		t.Error("Must be empty, but got:", elems)
	}
	if elems, err := sx.ParseObject(sxpf.Nil(), "b"); err == nil {
		t.Error("expected error, but got: ", elems)
	}

	if elems, err := sx.ParseObject(sxpf.MakeList(sxpf.MakeString("a")), "ss"); err == nil {
		t.Error("expected error, but got: ", elems)
	}
	if elems, err := sx.ParseObject(sxpf.MakeList(sxpf.MakeString("a")), ""); err == nil {
		t.Error("expected error, but got: ", elems)
	}
	if elems, err := sx.ParseObject(sxpf.MakeList(sxpf.MakeString("a")), "b"); err == nil {
		t.Error("expected error, but got: ", elems)
	}
	if elems, err := sx.ParseObject(sxpf.Cons(sxpf.Nil(), sxpf.MakeString("a")), "ps"); err == nil {
		t.Error("expected error, but got: ", elems)
	}

	if elems, err := sx.ParseObject(sxpf.MakeList(sxpf.MakeString("a")), "s"); err != nil {
		t.Error(err)
	} else if len(elems) != 1 {
		t.Error("length == 1, but got: ", elems)
	} else {
		_ = elems[0].(sxpf.String)
	}

}
