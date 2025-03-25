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

// Package zmk_test provides some tests for the zettelmarkup parser.
package zmk_test

import (
	"fmt"
	"strings"
	"testing"

	"t73f.de/r/sx"
	"t73f.de/r/zsc/sz/zmk"
	"t73f.de/r/zsx"
	"t73f.de/r/zsx/input"
)

type TestCase struct{ source, want string }
type TestCases []TestCase
type symbolMap map[string]*sx.Symbol

func replace(s string, sm symbolMap, tcs TestCases) TestCases {
	var sym string
	if len(sm) > 0 {
		sym = sm[s].GetValue()
	}
	var testCases TestCases
	for _, tc := range tcs {
		source := strings.ReplaceAll(tc.source, "$", s)
		want := tc.want
		if sym != "" {
			want = strings.ReplaceAll(want, "$%", sym)
		}
		want = strings.ReplaceAll(want, "$", s)
		testCases = append(testCases, TestCase{source, want})
	}
	return testCases
}

func checkTcs(t *testing.T, tcs TestCases) {
	t.Helper()

	var parser zmk.Parser
	for tcn, tc := range tcs {
		t.Run(fmt.Sprintf("TC=%02d,src=%q", tcn, tc.source), func(st *testing.T) {
			st.Helper()
			inp := input.NewInput([]byte(tc.source))
			parser.Initialize(inp)
			ast := parser.Parse()
			zsx.Walk(astWalker{}, ast, nil)
			got := ast.String()
			if tc.want != got {
				st.Errorf("\nwant=%q\n got=%q", tc.want, got)
			}
		})
	}
}

type astWalker struct{}

func (astWalker) VisitBefore(node *sx.Pair, env *sx.Pair) (sx.Object, bool) { return sx.Nil(), false }
func (astWalker) VisitAfter(node *sx.Pair, env *sx.Pair) sx.Object          { return node }

func TestEOL(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"", "()"},
		{"\n", "()"},
		{"\r", "()"},
		{"\r\n", "()"},
		{"\n\n", "()"},
	})
}

func TestText(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"abcd", "(BLOCK (PARA (TEXT \"abcd\")))"},
		{"ab cd", "(BLOCK (PARA (TEXT \"ab cd\")))"},
		{"abcd ", "(BLOCK (PARA (TEXT \"abcd\")))"},
		{" abcd", "(BLOCK (PARA (TEXT \"abcd\")))"},
		{"\\", "(BLOCK (PARA (TEXT \"\\\\\")))"},
		{"\\\n", "()"},
		{"\\\ndef", "(BLOCK (PARA (HARD) (TEXT \"def\")))"},
		{"\\\r", "()"},
		{"\\\rdef", "(BLOCK (PARA (HARD) (TEXT \"def\")))"},
		{"\\\r\n", "()"},
		{"\\\r\ndef", "(BLOCK (PARA (HARD) (TEXT \"def\")))"},
		{"\\a", "(BLOCK (PARA (TEXT \"a\")))"},
		{"\\aa", "(BLOCK (PARA (TEXT \"aa\")))"},
		{"a\\a", "(BLOCK (PARA (TEXT \"aa\")))"},
		{"\\+", "(BLOCK (PARA (TEXT \"+\")))"},
		{"\\ ", "(BLOCK (PARA (TEXT \"\u00a0\")))"},
		{"http://a, http://b", "(BLOCK (PARA (TEXT \"http://a, http://b\")))"},
	})
}

func TestSpace(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{" ", "()"},
		{"\t", "()"},
		{"  ", "()"},
	})
}

func TestSoftBreak(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"x\ny", "(BLOCK (PARA (TEXT \"x\") (SOFT) (TEXT \"y\")))"},
		{"z\n", "(BLOCK (PARA (TEXT \"z\")))"},
		{" \n ", "()"},
		{" \n", "()"},
	})
}

func TestHardBreak(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"x  \ny", "(BLOCK (PARA (TEXT \"x\") (HARD) (TEXT \"y\")))"},
		{"z  \n", "(BLOCK (PARA (TEXT \"z\")))"},
		{"   \n ", "()"},
		{"   \n", "()"},
	})
}

func TestLink(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[", "(BLOCK (PARA (TEXT \"[\")))"},
		{"[[", "(BLOCK (PARA (TEXT \"[[\")))"},
		{"[[|", "(BLOCK (PARA (TEXT \"[[|\")))"},
		{"[[]", "(BLOCK (PARA (TEXT \"[[]\")))"},
		{"[[|]", "(BLOCK (PARA (TEXT \"[[|]\")))"},
		{"[[]]", "(BLOCK (PARA (TEXT \"[[]]\")))"},
		{"[[|]]", "(BLOCK (PARA (TEXT \"[[|]]\")))"},
		{"[[ ]]", "(BLOCK (PARA (TEXT \"[[ ]]\")))"},
		{"[[\n]]", "(BLOCK (PARA (TEXT \"[[\") (SOFT) (TEXT \"]]\")))"},
		{"[[ a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\"))))"},
		{"[[a ]]", "(BLOCK (PARA (TEXT \"[[a ]]\")))"},
		{"[[a\n]]", "(BLOCK (PARA (TEXT \"[[a\") (SOFT) (TEXT \"]]\")))"},
		{"[[a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\"))))"},
		{"[[12345678901234]]", "(BLOCK (PARA (LINK () (ZETTEL \"12345678901234\"))))"},
		{"[[a]", "(BLOCK (PARA (TEXT \"[[a]\")))"},
		{"[[|a]]", "(BLOCK (PARA (TEXT \"[[|a]]\")))"},
		{"[[b|]]", "(BLOCK (PARA (TEXT \"[[b|]]\")))"},
		{"[[b|a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b\"))))"},
		{"[[b| a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b\"))))"},
		{"[[b%c|a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b%c\"))))"},
		{"[[b%%c|a]]", "(BLOCK (PARA (TEXT \"[[b\") (LITERAL-COMMENT () \"c|a]]\")))"},
		{"[[b|a]", "(BLOCK (PARA (TEXT \"[[b|a]\")))"},
		{"[[b\nc|a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b\") (SOFT) (TEXT \"c\"))))"},
		{"[[b c|a#n]]", "(BLOCK (PARA (LINK () (HOSTED \"a#n\") (TEXT \"b c\"))))"},
		{"[[a]]go", "(BLOCK (PARA (LINK () (HOSTED \"a\")) (TEXT \"go\")))"},
		{"[[b|a]]{go}", "(BLOCK (PARA (LINK ((\"go\" . \"\")) (HOSTED \"a\") (TEXT \"b\"))))"},
		{"[[[[a]]|b]]", "(BLOCK (PARA (TEXT \"[[\") (LINK () (HOSTED \"a\")) (TEXT \"|b]]\")))"},
		{"[[a[b]c|d]]", "(BLOCK (PARA (LINK () (HOSTED \"d\") (TEXT \"a[b]c\"))))"},
		{"[[[b]c|d]]", "(BLOCK (PARA (TEXT \"[\") (LINK () (HOSTED \"d\") (TEXT \"b]c\"))))"},
		{"[[a[]c|d]]", "(BLOCK (PARA (LINK () (HOSTED \"d\") (TEXT \"a[]c\"))))"},
		{"[[a[b]|d]]", "(BLOCK (PARA (LINK () (HOSTED \"d\") (TEXT \"a[b]\"))))"},
		{"[[\\|]]", "(BLOCK (PARA (LINK () (INVALID \"\\\\|\"))))"},
		{"[[\\||a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"|\"))))"},
		{"[[b\\||a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b|\"))))"},
		{"[[b\\|c|a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b|c\"))))"},
		{"[[\\]]]", "(BLOCK (PARA (LINK () (INVALID \"\\\\]\"))))"},
		{"[[\\]|a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"]\"))))"},
		{"[[b\\]|a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"b]\"))))"},
		{"[[\\]\\||a]]", "(BLOCK (PARA (LINK () (HOSTED \"a\") (TEXT \"]|\"))))"},
		{"[[http://a]]", "(BLOCK (PARA (LINK () (EXTERNAL \"http://a\"))))"},
		{"[[http://a|http://a]]", "(BLOCK (PARA (LINK () (EXTERNAL \"http://a\") (TEXT \"http://a\"))))"},
		{"[[[[a]]]]", "(BLOCK (PARA (TEXT \"[[\") (LINK () (HOSTED \"a\")) (TEXT \"]]\")))"},
		{"[[query:title]]", "(BLOCK (PARA (LINK () (QUERY \"title\"))))"},
		{"[[query:title syntax]]", "(BLOCK (PARA (LINK () (QUERY \"title syntax\"))))"},
		{"[[query:title | action]]", "(BLOCK (PARA (LINK () (QUERY \"title | action\"))))"},
		{"[[Text|query:title]]", "(BLOCK (PARA (LINK () (QUERY \"title\") (TEXT \"Text\"))))"},
		{"[[Text|query:title syntax]]", "(BLOCK (PARA (LINK () (QUERY \"title syntax\") (TEXT \"Text\"))))"},
		{"[[Text|query:title | action]]", "(BLOCK (PARA (LINK () (QUERY \"title | action\") (TEXT \"Text\"))))"},
	})
}

func TestEmbed(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"{", "(BLOCK (PARA (TEXT \"{\")))"},
		{"{{", "(BLOCK (PARA (TEXT \"{{\")))"},
		{"{{|", "(BLOCK (PARA (TEXT \"{{|\")))"},
		{"{{}", "(BLOCK (PARA (TEXT \"{{}\")))"},
		{"{{|}", "(BLOCK (PARA (TEXT \"{{|}\")))"},
		{"{{}}", "(BLOCK (PARA (TEXT \"{{}}\")))"},
		{"{{|}}", "(BLOCK (PARA (TEXT \"{{|}}\")))"},
		{"{{ }}", "(BLOCK (PARA (TEXT \"{{ }}\")))"},
		{"{{\n}}", "(BLOCK (PARA (TEXT \"{{\") (SOFT) (TEXT \"}}\")))"},
		{"{{a }}", "(BLOCK (PARA (TEXT \"{{a }}\")))"},
		{"{{a\n}}", "(BLOCK (PARA (TEXT \"{{a\") (SOFT) (TEXT \"}}\")))"},
		{"{{a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\")))"},
		{"{{12345678901234}}", "(BLOCK (PARA (EMBED () (ZETTEL \"12345678901234\") \"\")))"},
		{"{{ a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\")))"},
		{"{{a}", "(BLOCK (PARA (TEXT \"{{a}\")))"},
		{"{{|a}}", "(BLOCK (PARA (TEXT \"{{|a}}\")))"},
		{"{{b|}}", "(BLOCK (PARA (TEXT \"{{b|}}\")))"},
		{"{{b|a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"b\"))))"},
		{"{{b| a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"b\"))))"},
		{"{{b|a}", "(BLOCK (PARA (TEXT \"{{b|a}\")))"},
		{"{{b\nc|a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"b\") (SOFT) (TEXT \"c\"))))"},
		{"{{b c|a#n}}", "(BLOCK (PARA (EMBED () (HOSTED \"a#n\") \"\" (TEXT \"b c\"))))"},
		{"{{a}}{go}", "(BLOCK (PARA (EMBED ((\"go\" . \"\")) (HOSTED \"a\") \"\")))"},
		{"{{{{a}}|b}}", "(BLOCK (PARA (TEXT \"{{\") (EMBED () (HOSTED \"a\") \"\") (TEXT \"|b}}\")))"},
		{"{{\\|}}", "(BLOCK (PARA (EMBED () (INVALID \"\\\\|\") \"\")))"},
		{"{{\\||a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"|\"))))"},
		{"{{b\\||a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"b|\"))))"},
		{"{{b\\|c|a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"b|c\"))))"},
		{"{{\\}}}", "(BLOCK (PARA (EMBED () (INVALID \"\\\\}\") \"\")))"},
		{"{{\\}|a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"}\"))))"},
		{"{{b\\}|a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"b}\"))))"},
		{"{{\\}\\||a}}", "(BLOCK (PARA (EMBED () (HOSTED \"a\") \"\" (TEXT \"}|\"))))"},
		{"{{http://a}}", "(BLOCK (PARA (EMBED () (EXTERNAL \"http://a\") \"\")))"},
		{"{{http://a|http://a}}", "(BLOCK (PARA (EMBED () (EXTERNAL \"http://a\") \"\" (TEXT \"http://a\"))))"},
		{"{{{{a}}}}", "(BLOCK (PARA (TEXT \"{{\") (EMBED () (HOSTED \"a\") \"\") (TEXT \"}}\")))"},
	})
}

func TestCite(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[@", "(BLOCK (PARA (TEXT \"[@\")))"},
		{"[@]", "(BLOCK (PARA (TEXT \"[@]\")))"},
		{"[@a]", "(BLOCK (PARA (CITE () \"a\")))"},
		{"[@ a]", "(BLOCK (PARA (TEXT \"[@ a]\")))"},
		{"[@a ]", "(BLOCK (PARA (CITE () \"a\")))"},
		{"[@a\n]", "(BLOCK (PARA (CITE () \"a\")))"},
		{"[@a\nx]", "(BLOCK (PARA (CITE () \"a\" (SOFT) (TEXT \"x\"))))"},
		{"[@a\n\n]", "(BLOCK (PARA (TEXT \"[@a\")) (PARA (TEXT \"]\")))"},
		{"[@a,\n]", "(BLOCK (PARA (CITE () \"a\")))"},
		{"[@a,n]", "(BLOCK (PARA (CITE () \"a\" (TEXT \"n\"))))"},
		{"[@a| n]", "(BLOCK (PARA (CITE () \"a\" (TEXT \"n\"))))"},
		{"[@a|n ]", "(BLOCK (PARA (CITE () \"a\" (TEXT \"n\"))))"},
		{"[@a,[@b]]", "(BLOCK (PARA (CITE () \"a\" (CITE () \"b\"))))"},
		{"[@a]{color=green}", "(BLOCK (PARA (CITE ((\"color\" . \"green\")) \"a\")))"},
	})
	checkTcs(t, TestCases{
		{"[@a\n\n]", "(BLOCK (PARA (TEXT \"[@a\")) (PARA (TEXT \"]\")))"},
	})
}

func TestEndnote(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[^", "(BLOCK (PARA (TEXT \"[^\")))"},
		{"[^]", "(BLOCK (PARA (ENDNOTE ())))"},
		{"[^abc]", "(BLOCK (PARA (ENDNOTE () (TEXT \"abc\"))))"},
		{"[^abc ]", "(BLOCK (PARA (ENDNOTE () (TEXT \"abc\"))))"},
		{"[^abc\ndef]", "(BLOCK (PARA (ENDNOTE () (TEXT \"abc\") (SOFT) (TEXT \"def\"))))"},
		{"[^abc\n\ndef]", "(BLOCK (PARA (TEXT \"[^abc\")) (PARA (TEXT \"def]\")))"},
		{"[^abc[^def]]", "(BLOCK (PARA (ENDNOTE () (TEXT \"abc\") (ENDNOTE () (TEXT \"def\")))))"},
		{"[^abc]{-}", "(BLOCK (PARA (ENDNOTE ((\"-\" . \"\")) (TEXT \"abc\"))))"},
	})
	checkTcs(t, TestCases{
		{"[^abc\n\ndef]", "(BLOCK (PARA (TEXT \"[^abc\")) (PARA (TEXT \"def]\")))"},
	})
}

func TestMark(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[!", "(BLOCK (PARA (TEXT \"[!\")))"},
		{"[!\n", "(BLOCK (PARA (TEXT \"[!\")))"},
		{"[!]", "(BLOCK (PARA (MARK \"\" \"\" \"\")))"},
		{"[!][!]", "(BLOCK (PARA (MARK \"\" \"\" \"\") (MARK \"\" \"\" \"\")))"},
		{"[! ]", "(BLOCK (PARA (TEXT \"[! ]\")))"},
		{"[!a]", "(BLOCK (PARA (MARK \"a\" \"\" \"\")))"},
		{"[!a][!a]", "(BLOCK (PARA (MARK \"a\" \"\" \"\") (MARK \"a\" \"\" \"\")))"},
		{"[!a ]", "(BLOCK (PARA (TEXT \"[!a ]\")))"},
		{"[!a_]", "(BLOCK (PARA (MARK \"a_\" \"\" \"\")))"},
		{"[!a_][!a]", "(BLOCK (PARA (MARK \"a_\" \"\" \"\") (MARK \"a\" \"\" \"\")))"},
		{"[!a-b]", "(BLOCK (PARA (MARK \"a-b\" \"\" \"\")))"},
		{"[!a|b]", "(BLOCK (PARA (MARK \"a\" \"\" \"\" (TEXT \"b\"))))"},
		{"[!a|]", "(BLOCK (PARA (MARK \"a\" \"\" \"\")))"},
		{"[!|b]", "(BLOCK (PARA (MARK \"\" \"\" \"\" (TEXT \"b\"))))"},
		{"[!|b ]", "(BLOCK (PARA (MARK \"\" \"\" \"\" (TEXT \"b\"))))"},
		{"[!|b c]", "(BLOCK (PARA (MARK \"\" \"\" \"\" (TEXT \"b c\"))))"},
	})
}

func TestComment(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"%", "(BLOCK (PARA (TEXT \"%\")))"},
		{"%%", "(BLOCK (PARA (LITERAL-COMMENT () \"\")))"},
		{"%\n", "(BLOCK (PARA (TEXT \"%\")))"},
		{"%%\n", "(BLOCK (PARA (LITERAL-COMMENT () \"\")))"},
		{"%%a", "(BLOCK (PARA (LITERAL-COMMENT () \"a\")))"},
		{"%%%a", "(BLOCK (PARA (LITERAL-COMMENT () \"a\")))"},
		{"%% a", "(BLOCK (PARA (LITERAL-COMMENT () \"a\")))"},
		{"%%%  a", "(BLOCK (PARA (LITERAL-COMMENT () \"a\")))"},
		{"%% % a", "(BLOCK (PARA (LITERAL-COMMENT () \"% a\")))"},
		{"%%a", "(BLOCK (PARA (LITERAL-COMMENT () \"a\")))"},
		{"a%%b", "(BLOCK (PARA (TEXT \"a\") (LITERAL-COMMENT () \"b\")))"},
		{"a %%b", "(BLOCK (PARA (TEXT \"a \") (LITERAL-COMMENT () \"b\")))"},
		{" %%b", "(BLOCK (PARA (LITERAL-COMMENT () \"b\")))"},
		{"%%b ", "(BLOCK (PARA (LITERAL-COMMENT () \"b \")))"},
		{"100%", "(BLOCK (PARA (TEXT \"100%\")))"},
		{"%%{=}a", "(BLOCK (PARA (LITERAL-COMMENT ((\"\" . \"\")) \"a\")))"},
	})
}

func TestFormat(t *testing.T) {
	symMap := symbolMap{
		"_": zsx.SymFormatEmph,
		"*": zsx.SymFormatStrong,
		">": zsx.SymFormatInsert,
		"~": zsx.SymFormatDelete,
		"^": zsx.SymFormatSuper,
		",": zsx.SymFormatSub,
		"#": zsx.SymFormatMark,
		":": zsx.SymFormatSpan,
	}
	t.Parallel()
	// Not for Insert / '>', because collision with quoted list
	// Not for Quote / '"', because escaped representation.
	for _, ch := range []string{"_", "*", "~", "^", ",", "#", ":"} {
		checkTcs(t, replace(ch, symMap, TestCases{
			{"$", "(BLOCK (PARA (TEXT \"$\")))"},
			{"$$", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$$$$", "(BLOCK (PARA ($% ())))"},
		}))
	}
	// Not for Quote / '"', because escaped representation.
	for _, ch := range []string{"_", "*", ">", "~", "^", ",", "#", ":"} {
		checkTcs(t, replace(ch, symMap, TestCases{
			{"$$a$$", "(BLOCK (PARA ($% () (TEXT \"a\"))))"},
			{"$$a$$$", "(BLOCK (PARA ($% () (TEXT \"a\")) (TEXT \"$\")))"},
			{"$$$a$$", "(BLOCK (PARA ($% () (TEXT \"$a\"))))"},
			{"$$$a$$$", "(BLOCK (PARA ($% () (TEXT \"$a\")) (TEXT \"$\")))"},
			{"$\\$", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$\\$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$$\\$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$$a\\$$", "(BLOCK (PARA (TEXT \"$$a$$\")))"},
			{"$$a$\\$", "(BLOCK (PARA (TEXT \"$$a$$\")))"},
			{"$$a\\$$$", "(BLOCK (PARA ($% () (TEXT \"a$\"))))"},
			{"$$a\na$$", "(BLOCK (PARA ($% () (TEXT \"a\") (SOFT) (TEXT \"a\"))))"},
			{"$$a\n\na$$", "(BLOCK (PARA (TEXT \"$$a\")) (PARA (TEXT \"a$$\")))"},
			{"$$a$${go}", "(BLOCK (PARA ($% ((\"go\" . \"\")) (TEXT \"a\"))))"},
		}))
		checkTcs(t, replace(ch, symMap, TestCases{
			{"$$a\n\na$$", "(BLOCK (PARA (TEXT \"$$a\")) (PARA (TEXT \"a$$\")))"},
		}))
	}
	checkTcs(t, replace(`"`, symbolMap{`"`: zsx.SymFormatQuote}, TestCases{
		{"$", "(BLOCK (PARA (TEXT \"\\\"\")))"},
		{"$$", "(BLOCK (PARA (TEXT \"\\\"\\\"\")))"},
		{"$$$", "(BLOCK (PARA (TEXT \"\\\"\\\"\\\"\")))"},
		{"$$$$", "(BLOCK (PARA ($% ())))"},

		{"$$a$$", "(BLOCK (PARA ($% () (TEXT \"a\"))))"},
		{"$$a$$$", "(BLOCK (PARA ($% () (TEXT \"a\")) (TEXT \"\\\"\")))"},
		{"$$$a$$", "(BLOCK (PARA ($% () (TEXT \"\\\"a\"))))"},
		{"$$$a$$$", "(BLOCK (PARA ($% () (TEXT \"\\\"a\")) (TEXT \"\\\"\")))"},
		{"$\\$", "(BLOCK (PARA (TEXT \"\\\"\\\"\")))"},
		{"$\\$$", "(BLOCK (PARA (TEXT \"\\\"\\\"\\\"\")))"},
		{"$$\\$", "(BLOCK (PARA (TEXT \"\\\"\\\"\\\"\")))"},
		{"$$a\\$$", "(BLOCK (PARA (TEXT \"\\\"\\\"a\\\"\\\"\")))"},
		{"$$a$\\$", "(BLOCK (PARA (TEXT \"\\\"\\\"a\\\"\\\"\")))"},
		{"$$a\\$$$", "(BLOCK (PARA ($% () (TEXT \"a\\\"\"))))"},
		{"$$a\na$$", "(BLOCK (PARA ($% () (TEXT \"a\") (SOFT) (TEXT \"a\"))))"},
		{"$$a\n\na$$", "(BLOCK (PARA (TEXT \"\\\"\\\"a\")) (PARA (TEXT \"a\\\"\\\"\")))"},
		{"$$a$${go}", "(BLOCK (PARA ($% ((\"go\" . \"\")) (TEXT \"a\"))))"},
	}))
	checkTcs(t, TestCases{
		{"__****__", "(BLOCK (PARA (FORMAT-EMPH () (FORMAT-STRONG ()))))"},
		{"__**a**__", "(BLOCK (PARA (FORMAT-EMPH () (FORMAT-STRONG () (TEXT \"a\")))))"},
		{"__**__**", "(BLOCK (PARA (TEXT \"__\") (FORMAT-STRONG () (TEXT \"__\"))))"},
	})
}

func TestLiteral(t *testing.T) {
	symMap := symbolMap{
		"`": zsx.SymLiteralCode,
		"'": zsx.SymLiteralInput,
		"=": zsx.SymLiteralOutput,
	}
	t.Parallel()
	for _, ch := range []string{"`", "'", "="} {
		checkTcs(t, replace(ch, symMap, TestCases{
			{"$", "(BLOCK (PARA (TEXT \"$\")))"},
			{"$$", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$$$$", "(BLOCK (PARA ($% () \"\")))"},
			{"$$a$$", "(BLOCK (PARA ($% () \"a\")))"},
			{"$$a$$$", "(BLOCK (PARA ($% () \"a\") (TEXT \"$\")))"},
			{"$$$a$$", "(BLOCK (PARA ($% () \"$a\")))"},
			{"$$$a$$$", "(BLOCK (PARA ($% () \"$a\") (TEXT \"$\")))"},
			{"$\\$", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$\\$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$$\\$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$$a\\$$", "(BLOCK (PARA (TEXT \"$$a$$\")))"},
			{"$$a$\\$", "(BLOCK (PARA (TEXT \"$$a$$\")))"},
			{"$$a\\$$$", "(BLOCK (PARA ($% () \"a$\")))"},
			{"$$a$${go}", "(BLOCK (PARA ($% ((\"go\" . \"\")) \"a\")))"},
		}))
	}
	checkTcs(t, TestCases{
		{"``<script `` abc", "(BLOCK (PARA (LITERAL-CODE () \"<script \") (TEXT \" abc\")))"},
		{"''````''", "(BLOCK (PARA (LITERAL-INPUT () \"````\")))"},
		{"''``a``''", "(BLOCK (PARA (LITERAL-INPUT () \"``a``\")))"},
		{"''``''``", "(BLOCK (PARA (LITERAL-INPUT () \"``\") (TEXT \"``\")))"},
		{"''\\'''", "(BLOCK (PARA (LITERAL-INPUT () \"'\")))"},
	})
}

func TestLiteralMath(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"$", "(BLOCK (PARA (TEXT \"$\")))"},
		{"$$", "(BLOCK (PARA (TEXT \"$$\")))"},
		{"$$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
		{"$$$$", "(BLOCK (PARA (LITERAL-MATH () \"\")))"},
		{"$$a$$", "(BLOCK (PARA (LITERAL-MATH () \"a\")))"},
		{"$$a$$$", "(BLOCK (PARA (LITERAL-MATH () \"a\") (TEXT \"$\")))"},
		{"$$$a$$", "(BLOCK (PARA (LITERAL-MATH () \"$a\")))"},
		{"$$$a$$$", "(BLOCK (PARA (LITERAL-MATH () \"$a\") (TEXT \"$\")))"},
		{`$\$`, "(BLOCK (PARA (TEXT \"$$\")))"},
		{`$\$$`, "(BLOCK (PARA (TEXT \"$$$\")))"},
		{`$$\$`, "(BLOCK (PARA (TEXT \"$$$\")))"},
		{`$$a\$$`, "(BLOCK (PARA (LITERAL-MATH () \"a\\\\\")))"},
		{`$$a$\$`, "(BLOCK (PARA (TEXT \"$$a$$\")))"},
		{`$$a\$$$`, "(BLOCK (PARA (LITERAL-MATH () \"a\\\\\") (TEXT \"$\")))"},
		{"$$a$${go}", "(BLOCK (PARA (LITERAL-MATH ((\"go\" . \"\")) \"a\")))"},
	})
}

func TestMixFormatCode(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"__abc__\n**def**", "(BLOCK (PARA (FORMAT-EMPH () (TEXT \"abc\")) (SOFT) (FORMAT-STRONG () (TEXT \"def\"))))"},
		{"''abc''\n==def==", "(BLOCK (PARA (LITERAL-INPUT () \"abc\") (SOFT) (LITERAL-OUTPUT () \"def\")))"},
		{"__abc__\n==def==", "(BLOCK (PARA (FORMAT-EMPH () (TEXT \"abc\")) (SOFT) (LITERAL-OUTPUT () \"def\")))"},
		{"__abc__\n``def``", "(BLOCK (PARA (FORMAT-EMPH () (TEXT \"abc\")) (SOFT) (LITERAL-CODE () \"def\")))"},
		{
			"\"\"ghi\"\"\n::abc::\n``def``\n",
			"(BLOCK (PARA (FORMAT-QUOTE () (TEXT \"ghi\")) (SOFT) (FORMAT-SPAN () (TEXT \"abc\")) (SOFT) (LITERAL-CODE () \"def\")))",
		},
	})
}

func TestNDash(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"--", "(BLOCK (PARA (TEXT \"\u2013\")))"},
		{"a--b", "(BLOCK (PARA (TEXT \"a\u2013b\")))"},
	})
}

func TestEntity(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"&", "(BLOCK (PARA (TEXT \"&\")))"},
		{"&;", "(BLOCK (PARA (TEXT \"&;\")))"},
		{"&#;", "(BLOCK (PARA (TEXT \"&#;\")))"},
		{"&#1a;", "(BLOCK (PARA (TEXT \"&#1a;\")))"},
		{"&#x;", "(BLOCK (PARA (TEXT \"&#x;\")))"},
		{"&#x0z;", "(BLOCK (PARA (TEXT \"&#x0z;\")))"},
		{"&1;", "(BLOCK (PARA (TEXT \"&1;\")))"},
		{"&#9;", "(BLOCK (PARA (TEXT \"&#9;\")))"}, // Numeric entities below space are not allowed.
		{"&#x1f;", "(BLOCK (PARA (TEXT \"&#x1f;\")))"},

		// Good cases
		{"&lt;", "(BLOCK (PARA (TEXT \"<\")))"},
		{"&#48;", "(BLOCK (PARA (TEXT \"0\")))"},
		{"&#x4A;", "(BLOCK (PARA (TEXT \"J\")))"},
		{"&#X4a;", "(BLOCK (PARA (TEXT \"J\")))"},
		{"&hellip;", "(BLOCK (PARA (TEXT \"\u2026\")))"},
		{"&nbsp;", "(BLOCK (PARA (TEXT \"\u00a0\")))"},
		{"E: &amp;,&#63;;&#x63;.", "(BLOCK (PARA (TEXT \"E: &,?;c.\")))"},
	})
}

func TestVerbatimZettel(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"@@@\n@@@", "()"},
		{"@@@\nabc\n@@@", "(BLOCK (VERBATIM-ZETTEL () \"abc\"))"},
		{"@@@@def\nabc\n@@@@", "(BLOCK (VERBATIM-ZETTEL ((\"\" . \"def\")) \"abc\"))"},
	})
}

func TestVerbatimCode(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"```\n```", "()"},
		{"```\nabc\n```", "(BLOCK (VERBATIM-CODE () \"abc\"))"},
		{"```\nabc\n````", "(BLOCK (VERBATIM-CODE () \"abc\"))"},
		{"````\nabc\n````", "(BLOCK (VERBATIM-CODE () \"abc\"))"},
		{"````\nabc\n```\n````", "(BLOCK (VERBATIM-CODE () \"abc\\n```\"))"},
		{"````go\nabc\n````", "(BLOCK (VERBATIM-CODE ((\"\" . \"go\")) \"abc\"))"},
	})
}

func TestVerbatimEval(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"~~~\n~~~", "()"},
		{"~~~\nabc\n~~~", "(BLOCK (VERBATIM-EVAL () \"abc\"))"},
		{"~~~\nabc\n~~~~", "(BLOCK (VERBATIM-EVAL () \"abc\"))"},
		{"~~~~\nabc\n~~~~", "(BLOCK (VERBATIM-EVAL () \"abc\"))"},
		{"~~~~\nabc\n~~~\n~~~~", "(BLOCK (VERBATIM-EVAL () \"abc\\n~~~\"))"},
		{"~~~~go\nabc\n~~~~", "(BLOCK (VERBATIM-EVAL ((\"\" . \"go\")) \"abc\"))"},
	})
}

func TestVerbatimMath(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"$$$\n$$$", "()"},
		{"$$$\nabc\n$$$", "(BLOCK (VERBATIM-MATH () \"abc\"))"},
		{"$$$\nabc\n$$$$", "(BLOCK (VERBATIM-MATH () \"abc\"))"},
		{"$$$$\nabc\n$$$$", "(BLOCK (VERBATIM-MATH () \"abc\"))"},
		{"$$$$\nabc\n$$$\n$$$$", "(BLOCK (VERBATIM-MATH () \"abc\\n$$$\"))"},
		{"$$$$go\nabc\n$$$$", "(BLOCK (VERBATIM-MATH ((\"\" . \"go\")) \"abc\"))"},
	})
}

func TestVerbatimComment(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"%%%\n%%%", "()"},
		{"%%%\nabc\n%%%", "(BLOCK (VERBATIM-COMMENT () \"abc\"))"},
		{"%%%%go\nabc\n%%%%", "(BLOCK (VERBATIM-COMMENT ((\"\" . \"go\")) \"abc\"))"},
	})
}

func TestPara(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"a\n\nb", "(BLOCK (PARA (TEXT \"a\")) (PARA (TEXT \"b\")))"},
		{"a\n \nb", "(BLOCK (PARA (TEXT \"a\")) (PARA (TEXT \"b\")))"},
	})
}

func TestSpanRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{":::\n:::", "()"},
		{":::\nabc\n:::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"abc\")))))"},
		{":::\nabc\n::::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"abc\")))))"},
		{"::::\nabc\n::::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"abc\")))))"},
		{"::::\nabc\n:::\ndef\n:::\n::::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"abc\")) (REGION-BLOCK () ((PARA (TEXT \"def\")))))))"},
		{":::{go}\n:::a", "(BLOCK (REGION-BLOCK ((\"go\" . \"\")) () (TEXT \"a\")))"},
		{":::\nabc\n::: def ", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"abc\"))) (TEXT \"def\")))"},
	})
}

func TestQuoteRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"<<<\n<<<", "()"},
		{"<<<\nabc\n<<<", "(BLOCK (REGION-QUOTE () ((PARA (TEXT \"abc\")))))"},
		{"<<<\nabc\n<<<<", "(BLOCK (REGION-QUOTE () ((PARA (TEXT \"abc\")))))"},
		{"<<<<\nabc\n<<<<", "(BLOCK (REGION-QUOTE () ((PARA (TEXT \"abc\")))))"},
		{"<<<<\nabc\n<<<\ndef\n<<<\n<<<<", "(BLOCK (REGION-QUOTE () ((PARA (TEXT \"abc\")) (REGION-QUOTE () ((PARA (TEXT \"def\")))))))"},
		{"<<<go\n<<< a", "(BLOCK (REGION-QUOTE ((\"\" . \"go\")) () (TEXT \"a\")))"},
		{"<<<\nabc\n<<< def ", "(BLOCK (REGION-QUOTE () ((PARA (TEXT \"abc\"))) (TEXT \"def\")))"},
	})
}

func TestVerseRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, replace("\"", nil, TestCases{
		{"$$$\n$$$", "()"},
		{"$$$\nabc\n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")))))"},
		{"$$$\nabc\n$$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")))))"},
		{"$$$$\nabc\n$$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")))))"},
		{"$$$\nabc\ndef\n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\") (HARD) (TEXT \"def\")))))"},
		{"$$$$\nabc\n$$$\ndef\n$$$\n$$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")) (REGION-VERSE () ((PARA (TEXT \"def\")))))))"},
		{"$$$go\n$$$x", "(BLOCK (REGION-VERSE ((\"\" . \"go\")) () (TEXT \"x\")))"},
		{"$$$\nabc\n$$$ def ", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\"))) (TEXT \"def\")))"},
		{"$$$\n space \n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"\u00a0space\u00a0\")))))"},
		{"$$$\n  spaces  \n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"\u00a0\u00a0spaces\u00a0\u00a0\")))))"},
		{"$$$\n  spaces  \n space  \n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"\u00a0\u00a0spaces\u00a0\u00a0\") (HARD) (TEXT \"\u00a0space\u00a0\u00a0\")))))"},
		{"$$$\n space space \n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"\u00a0space\u00a0space\u00a0\")))))"},
	}))
}

func TestHeading(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"=h", "(BLOCK (PARA (TEXT \"=h\")))"},
		{"= h", "(BLOCK (PARA (TEXT \"= h\")))"},
		{"==h", "(BLOCK (PARA (TEXT \"==h\")))"},
		{"== h", "(BLOCK (PARA (TEXT \"== h\")))"},
		{"===h", "(BLOCK (PARA (TEXT \"===h\")))"},
		{"===", "(BLOCK (PARA (TEXT \"===\")))"},
		{"=== h", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\")))"},
		{"===  h", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\")))"},
		{"==== h", "(BLOCK (HEADING 2 () \"\" \"\" (TEXT \"h\")))"},
		{"===== h", "(BLOCK (HEADING 3 () \"\" \"\" (TEXT \"h\")))"},
		{"====== h", "(BLOCK (HEADING 4 () \"\" \"\" (TEXT \"h\")))"},
		{"======= h", "(BLOCK (HEADING 5 () \"\" \"\" (TEXT \"h\")))"},
		{"======== h", "(BLOCK (HEADING 5 () \"\" \"\" (TEXT \"h\")))"},
		{"=", "(BLOCK (PARA (TEXT \"=\")))"},
		{"=== h=__=a__", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h=\") (FORMAT-EMPH () (TEXT \"=a\"))))"},
		{"=\n", "(BLOCK (PARA (TEXT \"=\")))"},
		{"a=", "(BLOCK (PARA (TEXT \"a=\")))"},
		{" =", "(BLOCK (PARA (TEXT \"=\")))"},
		{"=== h\na", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\")) (PARA (TEXT \"a\")))"},
		{"=== h i {-}", "(BLOCK (HEADING 1 ((\"-\" . \"\")) \"\" \"\" (TEXT \"h i\")))"},
		{"=== h {{a}}", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h \") (EMBED () (HOSTED \"a\") \"\")))"},
		{"=== h{{a}}", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\") (EMBED () (HOSTED \"a\") \"\")))"},
		{"=== {{a}}", "(BLOCK (HEADING 1 () \"\" \"\" (EMBED () (HOSTED \"a\") \"\")))"},
		{"=== h {{a}}{-}", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h \") (EMBED ((\"-\" . \"\")) (HOSTED \"a\") \"\")))"},
		{"=== h {{a}} {-}", "(BLOCK (HEADING 1 ((\"-\" . \"\")) \"\" \"\" (TEXT \"h \") (EMBED () (HOSTED \"a\") \"\")))"},
		{"=== h {-}{{a}}", "(BLOCK (HEADING 1 ((\"-\" . \"\")) \"\" \"\" (TEXT \"h\")))"},
		{"=== h{id=abc}", "(BLOCK (HEADING 1 ((\"id\" . \"abc\")) \"\" \"\" (TEXT \"h\")))"},
		{"=== h\n=== h", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\")) (HEADING 1 () \"\" \"\" (TEXT \"h\")))"},
	})
}

func TestHRule(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"-", "(BLOCK (PARA (TEXT \"-\")))"},
		{"---", "(BLOCK (THEMATIC ()))"},
		{"----", "(BLOCK (THEMATIC ()))"},
		{"---A", "(BLOCK (THEMATIC ((\"\" . \"A\"))))"},
		{"---A-", "(BLOCK (THEMATIC ((\"\" . \"A-\"))))"},
		{"-1", "(BLOCK (PARA (TEXT \"-1\")))"},
		{"2-1", "(BLOCK (PARA (TEXT \"2-1\")))"},
		{"---  {  go  }  ", "(BLOCK (THEMATIC ((\"go\" . \"\"))))"},
		{"---  {  .go  }  ", "(BLOCK (THEMATIC ((\"class\" . \"go\"))))"},
		{`---{lang=zmk}`, "(BLOCK (THEMATIC ((\"lang\" . \"zmk\"))))"},
		{`---{lang="zmk"}`, "(BLOCK (THEMATIC ((\"lang\" . \"zmk\"))))"},
	})
}

func TestList(t *testing.T) {
	t.Parallel()
	// No ">" in the following, because quotation lists may have empty items.
	for _, ch := range []string{"*", "#"} {
		checkTcs(t, replace(ch, nil, TestCases{
			{"$", "(BLOCK (PARA (TEXT \"$\")))"},
			{"$$", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$ ", "(BLOCK (PARA (TEXT \"$\")))"},
			{"$$ ", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$$$ ", "(BLOCK (PARA (TEXT \"$$$\")))"},
		}))
	}
	checkTcs(t, TestCases{
		{"* abc", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))))"},
		{"** abc", "(BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))))))"},
		{"*** abc", "(BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))))))))"},
		{"**** abc", "(BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))))))))))"},
		{"** abc\n**** def", "(BLOCK (UNORDERED () (BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (UNORDERED () (BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"def\")))))))))))"},
		{"* abc\ndef", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n* def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\"))) (BLOCK (PARA (TEXT \"def\")))))"},
		{"* abc\n  def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\") (SOFT) (TEXT \"def\")))))"},
		{"* abc\n   def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\") (SOFT) (TEXT \"def\")))))"},
		{"* abc\n\ndef", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n\n def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n\n  def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (PARA (TEXT \"def\")))))"},
		{"* abc\n\n   def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (PARA (TEXT \"def\")))))"},
		{"* abc\n** def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (UNORDERED () (BLOCK (PARA (TEXT \"def\")))))))"},
		{"* abc\n** def\n* ghi", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (UNORDERED () (BLOCK (PARA (TEXT \"def\"))))) (BLOCK (PARA (TEXT \"ghi\")))))"},
		{"* abc\n\n  def\n* ghi", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (PARA (TEXT \"def\"))) (BLOCK (PARA (TEXT \"ghi\")))))"},
		{"* abc\n** def\n   ghi\n  jkl", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")) (UNORDERED () (BLOCK (PARA (TEXT \"def\") (SOFT) (TEXT \"ghi\")))) (PARA (TEXT \"jkl\")))))"},

		// A list does not last beyond a region
		{":::\n# abc\n:::\n# def", "(BLOCK (REGION-BLOCK () ((ORDERED () (BLOCK (PARA (TEXT \"abc\")))))) (ORDERED () (BLOCK (PARA (TEXT \"def\")))))"},

		// A HRule creates a new list
		{"* abc\n---\n* def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))) (THEMATIC ()) (UNORDERED () (BLOCK (PARA (TEXT \"def\")))))"},

		// Changing list type adds a new list
		{"* abc\n# def", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))) (ORDERED () (BLOCK (PARA (TEXT \"def\")))))"},

		// Quotation lists may have empty items
		{">", "(BLOCK (QUOTATION () (BLOCK)))"},

		// Empty continuation
		{"* abc\n  ", "(BLOCK (UNORDERED () (BLOCK (PARA (TEXT \"abc\")))))"},
	})
}

func TestQuoteList(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"> w1 w2", "(BLOCK (QUOTATION () (BLOCK (PARA (TEXT \"w1 w2\")))))"},
		{"> w1\n> w2", "(BLOCK (QUOTATION () (BLOCK (PARA (TEXT \"w1\") (SOFT) (TEXT \"w2\")))))"},
		{"> w1\n>w2", "(BLOCK (QUOTATION () (BLOCK (PARA (TEXT \"w1\")))) (PARA (TEXT \">w2\")))"},
		{"> w1\n>\n>w2", "(BLOCK (QUOTATION () (BLOCK (PARA (TEXT \"w1\"))) (BLOCK)) (PARA (TEXT \">w2\")))"},
		{"> w1\n> \n> w2", "(BLOCK (QUOTATION () (BLOCK (PARA (TEXT \"w1\"))) (BLOCK) (BLOCK (PARA (TEXT \"w2\")))))"},
	})
}

func TestEnumAfterPara(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"abc\n* def", "(BLOCK (PARA (TEXT \"abc\")) (UNORDERED () (BLOCK (PARA (TEXT \"def\")))))"},
		{"abc\n*def", "(BLOCK (PARA (TEXT \"abc\") (SOFT) (TEXT \"*def\")))"},
	})
}

func TestDefinition(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{";", "(BLOCK (PARA (TEXT \";\")))"},
		{"; ", "(BLOCK (PARA (TEXT \";\")))"},
		{"; abc", "(BLOCK (DESCRIPTION () ((TEXT \"abc\"))))"},
		{"; abc\ndef", "(BLOCK (DESCRIPTION () ((TEXT \"abc\"))) (PARA (TEXT \"def\")))"},
		{"; abc\n def", "(BLOCK (DESCRIPTION () ((TEXT \"abc\"))) (PARA (TEXT \"def\")))"},
		{"; abc\n  def", "(BLOCK (DESCRIPTION () ((TEXT \"abc\") (SOFT) (TEXT \"def\"))))"},
		{"; abc\n  def\n  ghi", "(BLOCK (DESCRIPTION () ((TEXT \"abc\") (SOFT) (TEXT \"def\") (SOFT) (TEXT \"ghi\"))))"},
		{":", "(BLOCK (PARA (TEXT \":\")))"},
		{": ", "(BLOCK (PARA (TEXT \":\")))"},
		{": abc", "(BLOCK (PARA (TEXT \": abc\")))"},
		{"; abc\n: def", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))))"},
		{"; abc\n: def\nghi", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))) (PARA (TEXT \"ghi\")))"},
		{"; abc\n: def\n ghi", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))) (PARA (TEXT \"ghi\")))"},
		{"; abc\n: def\n  ghi", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\") (SOFT) (TEXT \"ghi\"))))))"},
		{"; abc\n: def\n\n  ghi", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\")) (PARA (TEXT \"ghi\"))))))"},
		{"; abc\n: def\n\n  ghi\n\n  jkl", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\")) (PARA (TEXT \"ghi\")) (PARA (TEXT \"jkl\"))))))"},
		{"; abc\n:", "(BLOCK (DESCRIPTION () ((TEXT \"abc\"))) (PARA (TEXT \":\")))"},
		{"; abc\n: def\n: ghi", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))) (BLOCK (PARA (TEXT \"ghi\"))))))"},
		{"; abc\n: def\n; ghi\n: jkl", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\")))) ((TEXT \"ghi\")) (BLOCK (BLOCK (PARA (TEXT \"jkl\"))))))"},

		// Empty description
		{"; abc\n: ", "(BLOCK (DESCRIPTION () ((TEXT \"abc\"))) (PARA (TEXT \":\")))"},
		// Empty continuation of definition
		{"; abc\n: def\n  ", "(BLOCK (DESCRIPTION () ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))))"},
	})
}

func TestTable(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"|", "()"},
		{"||", "(BLOCK (TABLE () ((CELL ()))))"},
		{"| |", "(BLOCK (TABLE () ((CELL ()))))"},
		{"|a", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")))))"},
		{"|a|", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")))))"},
		{"|a| ", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")) (CELL ()))))"},
		{"|a|b", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")) (CELL () (TEXT \"b\")))))"},
		{"|a\n|b", "(BLOCK (TABLE () ((CELL () (TEXT \"a\"))) ((CELL () (TEXT \"b\")))))"},
		{"|a|b\n|c|d", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")) (CELL () (TEXT \"b\"))) ((CELL () (TEXT \"c\")) (CELL () (TEXT \"d\")))))"},
		{"|%", "()"},
		{"|=a", "(BLOCK (TABLE ((CELL () (TEXT \"a\")))))"},
		{"|=a\n|b", "(BLOCK (TABLE ((CELL () (TEXT \"a\"))) ((CELL () (TEXT \"b\")))))"},
		{"|a|b\n|%---\n|c|d", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")) (CELL () (TEXT \"b\"))) ((CELL () (TEXT \"c\")) (CELL () (TEXT \"d\")))))"},
		{"|a|b\n|c", "(BLOCK (TABLE () ((CELL () (TEXT \"a\")) (CELL () (TEXT \"b\"))) ((CELL () (TEXT \"c\")) (CELL ()))))"},
		{"|=<a>\n|b|c", "(BLOCK (TABLE ((CELL ((align . \"left\")) (TEXT \"a\")) (CELL ())) ((CELL ((align . \"right\")) (TEXT \"b\")) (CELL () (TEXT \"c\")))))"},
		{"|=<a|=b>\n||", "(BLOCK (TABLE ((CELL ((align . \"left\")) (TEXT \"a\")) (CELL ((align . \"right\")) (TEXT \"b\"))) ((CELL ()) (CELL ((align . \"right\"))))))"},
	})
}

func TestTransclude(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"{{{a}}}", "(BLOCK (TRANSCLUDE () (HOSTED \"a\")))"},
		{"{{{a}}}b", "(BLOCK (TRANSCLUDE ((\"\" . \"b\")) (HOSTED \"a\")))"},
		{"{{{a}}}}", "(BLOCK (TRANSCLUDE () (HOSTED \"a\")))"},
		{"{{{a\\}}}}", "(BLOCK (TRANSCLUDE () (INVALID \"a\\\\}\")))"},
		{"{{{a\\}}}}b", "(BLOCK (TRANSCLUDE ((\"\" . \"b\")) (INVALID \"a\\\\}\")))"},
		{"{{{a}}", "(BLOCK (PARA (TEXT \"{\") (EMBED () (HOSTED \"a\") \"\")))"},
		{"{{{a}}}{go=b}", "(BLOCK (TRANSCLUDE ((\"go\" . \"b\")) (HOSTED \"a\")))"},
	})
}

func TestBlockAttr(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{":::go\na\n:::", "(BLOCK (REGION-BLOCK ((\"\" . \"go\")) ((PARA (TEXT \"a\")))))"},
		{":::go=\na\n:::", "(BLOCK (REGION-BLOCK ((\"\" . \"go\")) ((PARA (TEXT \"a\")))))"},
		{":::{}\na\n:::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"a\")))))"},
		{":::{ }\na\n:::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"a\")))))"},
		{":::{.go}\na\n:::", "(BLOCK (REGION-BLOCK ((\"class\" . \"go\")) ((PARA (TEXT \"a\")))))"},
		{":::{=go}\na\n:::", "(BLOCK (REGION-BLOCK ((\"\" . \"go\")) ((PARA (TEXT \"a\")))))"},
		{":::{go}\na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{go=py}\na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"py\")) ((PARA (TEXT \"a\")))))"},
		{":::{.go=py}\na\n:::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"a\")))))"},
		{":::{go=}\na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{.go=}\na\n:::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"a\")))))"},
		{":::{go py}\na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"\") (\"py\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{go\npy}\na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"\") (\"py\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{.go py}\na\n:::", "(BLOCK (REGION-BLOCK ((\"class\" . \"go\") (\"py\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{go .py}\na\n:::", "(BLOCK (REGION-BLOCK ((\"class\" . \"py\") (\"go\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{.go py=3}\na\n:::", "(BLOCK (REGION-BLOCK ((\"class\" . \"go\") (\"py\" . \"3\")) ((PARA (TEXT \"a\")))))"},
		{":::  {  go  }  \na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::  {  .go  }  \na\n:::", "(BLOCK (REGION-BLOCK ((\"class\" . \"go\")) ((PARA (TEXT \"a\")))))"},
	})
	checkTcs(t, replace("\"", nil, TestCases{
		{":::{py=3}\na\n:::", "(BLOCK (REGION-BLOCK ((\"py\" . \"3\")) ((PARA (TEXT \"a\")))))"},
		{":::{py=$2 3$}\na\n:::", "(BLOCK (REGION-BLOCK ((\"py\" . \"2 3\")) ((PARA (TEXT \"a\")))))"},
		{":::{py=$2\\$3$}\na\n:::", "(BLOCK (REGION-BLOCK ((\"py\" . \"2\\\"3\")) ((PARA (TEXT \"a\")))))"},
		{":::{py=2$3}\na\n:::", "(BLOCK (REGION-BLOCK ((\"py\" . \"2\\\"3\")) ((PARA (TEXT \"a\")))))"},
		{":::{py=$2\n3$}\na\n:::", "(BLOCK (REGION-BLOCK ((\"py\" . \"2\\n3\")) ((PARA (TEXT \"a\")))))"},
		{":::{py=$2 3}\na\n:::", "(BLOCK (REGION-BLOCK () ((PARA (TEXT \"a\")))))"},
		{":::{py=2 py=3}\na\n:::", "(BLOCK (REGION-BLOCK ((\"py\" . \"2 3\")) ((PARA (TEXT \"a\")))))"},
		{":::{.go .py}\na\n:::", "(BLOCK (REGION-BLOCK ((\"class\" . \"go py\")) ((PARA (TEXT \"a\")))))"},
		{":::{go go}\na\n:::", "(BLOCK (REGION-BLOCK ((\"go\" . \"\")) ((PARA (TEXT \"a\")))))"},
		{":::{=py =go}\na\n:::", "(BLOCK (REGION-BLOCK ((\"\" . \"go\")) ((PARA (TEXT \"a\")))))"},
	}))
}

func TestInlineAttr(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"::a::{}", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\"))))"},
		{"::a::{ }", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\"))))"},
		{"::a::{.go}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"go\")) (TEXT \"a\"))))"},
		{"::a::{=go}", "(BLOCK (PARA (FORMAT-SPAN ((\"\" . \"go\")) (TEXT \"a\"))))"},
		{"::a::{go}", "(BLOCK (PARA (FORMAT-SPAN ((\"go\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{go=py}", "(BLOCK (PARA (FORMAT-SPAN ((\"go\" . \"py\")) (TEXT \"a\"))))"},
		{"::a::{.go=py}", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\")) (TEXT \"{.go=py}\")))"},
		{"::a::{go=}", "(BLOCK (PARA (FORMAT-SPAN ((\"go\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{.go=}", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\")) (TEXT \"{.go=}\")))"},
		{"::a::{go py}", "(BLOCK (PARA (FORMAT-SPAN ((\"go\" . \"\") (\"py\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{go\npy}", "(BLOCK (PARA (FORMAT-SPAN ((\"go\" . \"\") (\"py\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{.go py}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"go\") (\"py\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{go .py}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"py\") (\"go\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{  \n go \n .py\n  \n}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"py\") (\"go\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{  \n go \n .py\n\n}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"py\") (\"go\" . \"\")) (TEXT \"a\"))))"},
		{"::a::{\ngo\n}", "(BLOCK (PARA (FORMAT-SPAN ((\"go\" . \"\")) (TEXT \"a\"))))"},
	})
	checkTcs(t, TestCases{
		{"::a::{py=3}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2 3\"}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2 3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2\\\"3\"}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2\\\"3\")) (TEXT \"a\"))))"},
		{"::a::{py=2\"3}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2\\\"3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2\n3\"}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2\\n3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2 3}", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\")) (TEXT \"{py=\\\"2 3}\")))"},
	})
	checkTcs(t, TestCases{
		{"::a::{py=2 py=3}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2 3\")) (TEXT \"a\"))))"},
		{"::a::{.go .py}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"go py\")) (TEXT \"a\"))))"},
	})
}

func TestTemp(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"", "()"},
	})
}
