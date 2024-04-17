// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------
package zmk_test

import (
	"testing"

	"t73f.de/r/zsc/sz/zmk"
)

func TestParseReference(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		link string
		err  bool
		exp  string
	}{
		{"", true, ""},
		{"12345678901234", false, "(ZETTEL \"12345678901234\")"},
		{"123", false, "(EXTERNAL \"123\")"},
		{",://", true, ""},
	}

	for i, tc := range testcases {
		got := zmk.ParseReference(tc.link)
		gotIsValid := zmk.ReferenceIsValid(got)
		if gotIsValid == tc.err {
			t.Errorf(
				"TC=%d, expected parse error of %q: %v, but got %q", i, tc.link, tc.err, got)
		}
		if gotIsValid && got.String() != tc.exp {
			t.Errorf("TC=%d, Reference of %q is %q, but got %q", i, tc.link, tc.exp, got)
		}
	}
}

func TestReferenceIsZettelMaterial(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		link       string
		isZettel   bool
		isExternal bool
		isLocal    bool
	}{
		{"", false, false, false},
		{"00000000000000", false, false, false},
		{"http://zettelstore.de/z/ast", false, true, false},
		{"12345678901234", true, false, false},
		{"12345678901234#local", true, false, false},
		{"http://12345678901234", false, true, false},
		{"http://zettelstore.de/z/12345678901234", false, true, false},
		{"http://zettelstore.de/12345678901234", false, true, false},
		{"/12345678901234", false, false, true},
		{"//12345678901234", false, false, true},
		{"./12345678901234", false, false, true},
		{"../12345678901234", false, false, true},
		{".../12345678901234", false, true, false},
	}

	for i, tc := range testcases {
		ref := zmk.ParseReference(tc.link)
		isZettel := zmk.ReferenceIsZettel(ref)
		if isZettel != tc.isZettel {
			t.Errorf(
				"TC=%d, Reference %q isZettel=%v expected, but got %v",
				i,
				tc.link,
				tc.isZettel,
				isZettel)
		}
		isLocal := zmk.ReferenceIsLocal(ref)
		if isLocal != tc.isLocal {
			t.Errorf(
				"TC=%d, Reference %q isLocal=%v expected, but got %v",
				i,
				tc.link,
				tc.isLocal, isLocal)
		}
		isExternal := zmk.ReferenceIsExternal(ref)
		if isExternal != tc.isExternal {
			t.Errorf(
				"TC=%d, Reference %q isExternal=%v expected, but got %v",
				i,
				tc.link,
				tc.isExternal,
				isExternal)
		}
	}
}
