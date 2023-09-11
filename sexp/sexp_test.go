//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexp_test

import (
	"testing"

	"zettelstore.de/client.fossil/sexp"
	"zettelstore.de/sx.fossil"
)

func TestParseObject(t *testing.T) {
	if elems, err := sexp.ParseList(sx.MakeString("a"), "s"); err == nil {
		t.Error("expected an error, but got: ", elems)
	}
	if elems, err := sexp.ParseList(sx.Nil(), ""); err != nil {
		t.Error(err)
	} else if len(elems) != 0 {
		t.Error("Must be empty, but got:", elems)
	}
	if elems, err := sexp.ParseList(sx.Nil(), "b"); err == nil {
		t.Error("expected error, but got: ", elems)
	}

	if elems, err := sexp.ParseList(sx.MakeList(sx.MakeString("a")), "ss"); err == nil {
		t.Error("expected error, but got: ", elems)
	}
	if elems, err := sexp.ParseList(sx.MakeList(sx.MakeString("a")), ""); err == nil {
		t.Error("expected error, but got: ", elems)
	}
	if _, err := sexp.ParseList(sx.MakeList(sx.MakeString("a")), "b"); err != nil {
		t.Error("expected [1], but got error: ", err)
	}
	if elems, err := sexp.ParseList(sx.Cons(sx.Nil(), sx.MakeString("a")), "ps"); err == nil {
		t.Error("expected error, but got: ", elems)
	}

	if elems, err := sexp.ParseList(sx.MakeList(sx.MakeString("a")), "s"); err != nil {
		t.Error(err)
	} else if len(elems) != 1 {
		t.Error("length == 1, but got: ", elems)
	} else {
		_ = elems[0].(sx.String)
	}

}
