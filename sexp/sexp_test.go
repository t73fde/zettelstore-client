//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package sexp_test

import (
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/sexp"
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

	if elems, err := sexp.ParseList(sx.Nil(), "r"); err != nil {
		t.Error(err)
	} else if len(elems) != 1 {
		t.Error("length == 1, but got: ", elems)
	} else if !sx.IsNil(elems[0]) {
		t.Error("must be nil, but got:", elems[0])
	}
	if elems, err := sexp.ParseList(sx.MakeList(sx.MakeString("a")), "r"); err != nil {
		t.Error(err)
	} else if len(elems) != 1 {
		t.Error("length == 1, but got: ", elems)
	} else if !elems[0].IsEqual(sx.MakeList(sx.MakeString("a"))) {
		t.Error("must be (\"a\"), but got:", elems[0])
	}
	if elems, err := sexp.ParseList(sx.MakeList(sx.MakeString("a")), "sr"); err != nil {
		t.Error(err)
	} else if len(elems) != 2 {
		t.Error("length == 2, but got: ", elems)
	} else if !elems[0].IsEqual(sx.MakeString("a")) {
		t.Error("0-th must be \"a\", but got:", elems[0])
	} else if !sx.IsNil(elems[1]) {
		t.Error("must be nil, but got:", elems[1])
	}
	if elems, err := sexp.ParseList(sx.MakeList(sx.MakeString("a"), sx.MakeString("b"), sx.MakeString("c")), "sr"); err != nil {
		t.Error(err)
	} else if len(elems) != 2 {
		t.Error("length == 2, but got: ", elems)
	} else if !elems[0].IsEqual(sx.MakeString("a")) {
		t.Error("0-th must be \"a\", but got:", elems[0])
	} else if !elems[1].IsEqual(sx.MakeList(sx.MakeString("b"), sx.MakeString("c"))) {
		t.Error("must be nil, but got:", elems[1])
	}
}
