//-----------------------------------------------------------------------------
// Copyright (c) 2026-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2026-present Detlef Stern
//-----------------------------------------------------------------------------

package shtml_test

import (
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxreader"
	"t73f.de/r/zsc/shtml"
)

func TestEvaluator(t *testing.T) {
	for _, tc := range evalTestCases {
		t.Run(tc.name, func(t *testing.T) {
			rd := sxreader.MakeReader(strings.NewReader(tc.sz))
			objs, err := rd.ReadAll()
			if err != nil {
				t.Error(err)
				return
			}
			if len(objs) == 0 {
				t.Errorf("no object read: %v", objs)
				return
			}
			if len(objs) > 1 {
				t.Errorf("more than 1 object read: %v", objs)
				return
			}
			lst, isPair := sx.GetPair(objs[0])
			if !isPair {
				t.Errorf("sz is not a list, but %T/%v", objs[0], objs[0])
				return
			}
			ev := shtml.NewEvaluator(3)
			env := shtml.MakeEnvironment("en")
			sxh, err := ev.Evaluate(lst, &env)
			if err != nil {
				t.Errorf("err in Evaluate: %v", err)
				return
			}
			if got := sxh.String(); got != tc.exp {
				t.Errorf("\nexp: %v\ngot: %v", tc.exp, got)
			}
		})
	}
}

type evalTestCase struct {
	name string
	sz   string
	exp  string
}

var evalTestCases = []evalTestCase{
	{name: "simple-block",
		sz:  "(BLOCK)",
		exp: "()",
	},

	{name: "heading-without-id",
		sz:  `(BLOCK (HEADING () 1 (TEXT "???")))`,
		exp: `((h4 "???"))`,
	},

	{name: "mark-without-id",
		sz:  `(BLOCK (PARA (MARK () "!!!" (TEXT "WTF!"))))`,
		exp: `((p (a ((id . "!!!")) "WTF!")))`,
	},

	{name: "nothing",
		sz:  "()",
		exp: "()",
	},
}
