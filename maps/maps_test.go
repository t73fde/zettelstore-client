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

package maps_test

import (
	"testing"

	"t73f.de/r/zsc/maps"
)

func isSorted(seq []string) bool {
	for i := 1; i < len(seq); i++ {
		if seq[i] < seq[i-1] {
			return false
		}
	}
	return true
}

func TestKeys(t *testing.T) {
	testcases := []struct{ keys []string }{
		{nil}, {[]string{""}},
		{[]string{"z", "y", "a"}},
	}
	for i, tc := range testcases {
		m := make(map[string]struct{})
		for _, k := range tc.keys {
			m[k] = struct{}{}
		}
		got := maps.Keys(m)
		if len(got) != len(tc.keys) {
			t.Errorf("%d: wrong number of keys: exp %d, got %d", i, len(tc.keys), len(got))
		}
		if !isSorted(got) {
			t.Errorf("%d: keys not sorted: %v", i, got)
		}
	}
}
