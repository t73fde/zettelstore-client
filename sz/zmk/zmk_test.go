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
	"zettelstore.de/client.fossil/input"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/client.fossil/sz/zmk"
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

func checkTcs(t *testing.T, isBlock bool, tcs TestCases) {
	t.Helper()

	for tcn, tc := range tcs {
		t.Run(fmt.Sprintf("TC=%02d,src=%q", tcn, tc.source), func(st *testing.T) {
			st.Helper()
			ast := parseInput(tc.source, isBlock)
			got := ast.String()
			if tc.want != got {
				st.Errorf("\nwant=%q\n got=%q", tc.want, got)
			}
		})
	}
}
func parseInput(src string, asBlock bool) sx.Sequence {
	inp := input.NewInput([]byte(src))
	if asBlock {
		bl := zmk.ParseBlocks(inp)
		return bl
	}
	il := zmk.ParseInlines(inp)
	return il
}

func TestEOL(t *testing.T) {
	t.Parallel()
	for _, isBlock := range []bool{true, false} {
		checkTcs(t, isBlock, TestCases{
			{"", "()"},
			{"\n", "()"},
			{"\r", "()"},
			{"\r\n", "()"},
			{"\n\n", "()"},
		})
	}
}

func TestText(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"abcd", "(INLINE (TEXT \"abcd\"))"},
		{"ab cd", "(INLINE (TEXT \"ab\") (SPACE) (TEXT \"cd\"))"},
		{"abcd ", "(INLINE (TEXT \"abcd\"))"},
		{" abcd", "(INLINE (TEXT \"abcd\"))"},
		{"\\", "(INLINE (TEXT \"\\\\\"))"},
		{"\\\n", "()"},
		{"\\\ndef", "(INLINE (HARD) (TEXT \"def\"))"},
		{"\\\r", "()"},
		{"\\\rdef", "(INLINE (HARD) (TEXT \"def\"))"},
		{"\\\r\n", "()"},
		{"\\\r\ndef", "(INLINE (HARD) (TEXT \"def\"))"},
		{"\\a", "(INLINE (TEXT \"a\"))"},
		{"\\aa", "(INLINE (TEXT \"aa\"))"},
		{"a\\a", "(INLINE (TEXT \"aa\"))"},
		{"\\+", "(INLINE (TEXT \"+\"))"},
		{"\\ ", "(INLINE (TEXT \"\u00a0\"))"},
		{"http://a, http://b", "(INLINE (TEXT \"http://a,\") (SPACE) (TEXT \"http://b\"))"},
	})
}

func TestSpace(t *testing.T) {
	t.Parallel()
	for _, isBlock := range []bool{true, false} {
		checkTcs(t, isBlock, TestCases{
			{" ", "()"},
			{"\t", "()"},
			{"  ", "()"},
		})
	}
}

func TestSoftBreak(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"x\ny", "(INLINE (TEXT \"x\") (SOFT) (TEXT \"y\"))"},
		{"z\n", "(INLINE (TEXT \"z\"))"},
		{" \n ", "()"},
		{" \n", "()"},
	})
}

func TestHardBreak(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"x  \ny", "(INLINE (TEXT \"x\") (HARD) (TEXT \"y\"))"},
		{"z  \n", "(INLINE (TEXT \"z\"))"},
		{"   \n ", "()"},
		{"   \n", "()"},
	})
}

func TestLink(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"[", "(INLINE (TEXT \"[\"))"},
		{"[[", "(INLINE (TEXT \"[[\"))"},
		{"[[|", "(INLINE (TEXT \"[[|\"))"},
		{"[[]", "(INLINE (TEXT \"[[]\"))"},
		{"[[|]", "(INLINE (TEXT \"[[|]\"))"},
		{"[[]]", "(INLINE (TEXT \"[[]]\"))"},
		{"[[|]]", "(INLINE (TEXT \"[[|]]\"))"},
		{"[[ ]]", "(INLINE (TEXT \"[[\") (SPACE) (TEXT \"]]\"))"},
		{"[[\n]]", "(INLINE (TEXT \"[[\") (SOFT) (TEXT \"]]\"))"},
		{"[[ a]]", "(INLINE (LINK-EXTERNAL () \"a\"))"},
		{"[[a ]]", "(INLINE (TEXT \"[[a\") (SPACE) (TEXT \"]]\"))"},
		{"[[a\n]]", "(INLINE (TEXT \"[[a\") (SOFT) (TEXT \"]]\"))"},
		{"[[a]]", "(INLINE (LINK-EXTERNAL () \"a\"))"},
		{"[[12345678901234]]", "(INLINE (LINK-ZETTEL () \"12345678901234\"))"},
		{"[[a]", "(INLINE (TEXT \"[[a]\"))"},
		{"[[|a]]", "(INLINE (TEXT \"[[|a]]\"))"},
		{"[[b|]]", "(INLINE (TEXT \"[[b|]]\"))"},
		{"[[b|a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b\")))"},
		{"[[b| a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b\")))"},
		{"[[b%c|a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b%c\")))"},
		{"[[b%%c|a]]", "(INLINE (TEXT \"[[b\") (LITERAL-COMMENT () \"c|a]]\"))"},
		{"[[b|a]", "(INLINE (TEXT \"[[b|a]\"))"},
		{"[[b\nc|a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b\") (SOFT) (TEXT \"c\")))"},
		{"[[b c|a#n]]", "(INLINE (LINK-EXTERNAL () \"a#n\" (TEXT \"b\") (SPACE) (TEXT \"c\")))"},
		{"[[a]]go", "(INLINE (LINK-EXTERNAL () \"a\") (TEXT \"go\"))"},
		{"[[b|a]]{go}", "(INLINE (LINK-EXTERNAL ((\"go\" . \"\")) \"a\" (TEXT \"b\")))"},
		{"[[[[a]]|b]]", "(INLINE (TEXT \"[[\") (LINK-EXTERNAL () \"a\") (TEXT \"|b]]\"))"},
		{"[[a[b]c|d]]", "(INLINE (LINK-EXTERNAL () \"d\" (TEXT \"a[b]c\")))"},
		{"[[[b]c|d]]", "(INLINE (TEXT \"[\") (LINK-EXTERNAL () \"d\" (TEXT \"b]c\")))"},
		{"[[a[]c|d]]", "(INLINE (LINK-EXTERNAL () \"d\" (TEXT \"a[]c\")))"},
		{"[[a[b]|d]]", "(INLINE (LINK-EXTERNAL () \"d\" (TEXT \"a[b]\")))"},
		{"[[\\|]]", "(INLINE (LINK-EXTERNAL () \"\\\\|\"))"},
		{"[[\\||a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"|\")))"},
		{"[[b\\||a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b|\")))"},
		{"[[b\\|c|a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b|c\")))"},
		{"[[\\]]]", "(INLINE (LINK-EXTERNAL () \"\\\\]\"))"},
		{"[[\\]|a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"]\")))"},
		{"[[b\\]|a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"b]\")))"},
		{"[[\\]\\||a]]", "(INLINE (LINK-EXTERNAL () \"a\" (TEXT \"]|\")))"},
		{"[[http://a]]", "(INLINE (LINK-EXTERNAL () \"http://a\"))"},
		{"[[http://a|http://a]]", "(INLINE (LINK-EXTERNAL () \"http://a\" (TEXT \"http://a\")))"},
		{"[[[[a]]]]", "(INLINE (TEXT \"[[\") (LINK-EXTERNAL () \"a\") (TEXT \"]]\"))"},
		{"[[query:title]]", "(INLINE (LINK-QUERY () \"title\"))"},
		{"[[query:title syntax]]", "(INLINE (LINK-QUERY () \"title syntax\"))"},
		{"[[query:title | action]]", "(INLINE (LINK-QUERY () \"title | action\"))"},
		{"[[Text|query:title]]", "(INLINE (LINK-QUERY () \"title\" (TEXT \"Text\")))"},
		{"[[Text|query:title syntax]]", "(INLINE (LINK-QUERY () \"title syntax\" (TEXT \"Text\")))"},
		{"[[Text|query:title | action]]", "(INLINE (LINK-QUERY () \"title | action\" (TEXT \"Text\")))"},
	})
}

func TestEmbed(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"{", "(INLINE (TEXT \"{\"))"},
		{"{{", "(INLINE (TEXT \"{{\"))"},
		{"{{|", "(INLINE (TEXT \"{{|\"))"},
		{"{{}", "(INLINE (TEXT \"{{}\"))"},
		{"{{|}", "(INLINE (TEXT \"{{|}\"))"},
		{"{{}}", "(INLINE (TEXT \"{{}}\"))"},
		{"{{|}}", "(INLINE (TEXT \"{{|}}\"))"},
		{"{{ }}", "(INLINE (TEXT \"{{\") (SPACE) (TEXT \"}}\"))"},
		{"{{\n}}", "(INLINE (TEXT \"{{\") (SOFT) (TEXT \"}}\"))"},
		{"{{a }}", "(INLINE (TEXT \"{{a\") (SPACE) (TEXT \"}}\"))"},
		{"{{a\n}}", "(INLINE (TEXT \"{{a\") (SOFT) (TEXT \"}}\"))"},
		{"{{a}}", "(INLINE (EMBED () \"a\"))"},
		{"{{12345678901234}}", "(INLINE (EMBED () \"12345678901234\"))"},
		{"{{ a}}", "(INLINE (EMBED () \"a\"))"},
		{"{{a}", "(INLINE (TEXT \"{{a}\"))"},
		{"{{|a}}", "(INLINE (TEXT \"{{|a}}\"))"},
		{"{{b|}}", "(INLINE (TEXT \"{{b|}}\"))"},
		{"{{b|a}}", "(INLINE (EMBED () \"a\" (TEXT \"b\")))"},
		{"{{b| a}}", "(INLINE (EMBED () \"a\" (TEXT \"b\")))"},
		{"{{b|a}", "(INLINE (TEXT \"{{b|a}\"))"},
		{"{{b\nc|a}}", "(INLINE (EMBED () \"a\" (TEXT \"b\") (SOFT) (TEXT \"c\")))"},
		{"{{b c|a#n}}", "(INLINE (EMBED () \"a#n\" (TEXT \"b\") (SPACE) (TEXT \"c\")))"},
		{"{{a}}{go}", "(INLINE (EMBED ((\"go\" . \"\")) \"a\"))"},
		{"{{{{a}}|b}}", "(INLINE (TEXT \"{{\") (EMBED () \"a\") (TEXT \"|b}}\"))"},
		{"{{\\|}}", "(INLINE (EMBED () \"\\\\|\"))"},
		{"{{\\||a}}", "(INLINE (EMBED () \"a\" (TEXT \"|\")))"},
		{"{{b\\||a}}", "(INLINE (EMBED () \"a\" (TEXT \"b|\")))"},
		{"{{b\\|c|a}}", "(INLINE (EMBED () \"a\" (TEXT \"b|c\")))"},
		{"{{\\}}}", "(INLINE (EMBED () \"\\\\}\"))"},
		{"{{\\}|a}}", "(INLINE (EMBED () \"a\" (TEXT \"}\")))"},
		{"{{b\\}|a}}", "(INLINE (EMBED () \"a\" (TEXT \"b}\")))"},
		{"{{\\}\\||a}}", "(INLINE (EMBED () \"a\" (TEXT \"}|\")))"},
		{"{{http://a}}", "(INLINE (EMBED () \"http://a\"))"},
		{"{{http://a|http://a}}", "(INLINE (EMBED () \"http://a\" (TEXT \"http://a\")))"},
		{"{{{{a}}}}", "(INLINE (TEXT \"{{\") (EMBED () \"a\") (TEXT \"}}\"))"},
	})
}

func TestCite(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"[@", "(INLINE (TEXT \"[@\"))"},
		{"[@]", "(INLINE (TEXT \"[@]\"))"},
		{"[@a]", "(INLINE (CITE () \"a\"))"},
		{"[@ a]", "(INLINE (TEXT \"[@\") (SPACE) (TEXT \"a]\"))"},
		{"[@a ]", "(INLINE (CITE () \"a\"))"},
		{"[@a\n]", "(INLINE (CITE () \"a\"))"},
		{"[@a\nx]", "(INLINE (CITE () \"a\" (SOFT) (TEXT \"x\")))"},
		{"[@a\n\n]", "(INLINE (TEXT \"[@a\") (SOFT) (SOFT) (TEXT \"]\"))"},
		{"[@a,\n]", "(INLINE (CITE () \"a\"))"},
		{"[@a,n]", "(INLINE (CITE () \"a\" (TEXT \"n\")))"},
		{"[@a| n]", "(INLINE (CITE () \"a\" (TEXT \"n\")))"},
		{"[@a|n ]", "(INLINE (CITE () \"a\" (TEXT \"n\")))"},
		{"[@a,[@b]]", "(INLINE (CITE () \"a\" (CITE () \"b\")))"},
		{"[@a]{color=green}", "(INLINE (CITE ((\"color\" . \"green\")) \"a\"))"},
	})
	checkTcs(t, true, TestCases{
		{"[@a\n\n]", "(BLOCK (PARA (TEXT \"[@a\")) (PARA (TEXT \"]\")))"},
	})
}

func TestEndnote(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"[^", "(INLINE (TEXT \"[^\"))"},
		{"[^]", "(INLINE (ENDNOTE ()))"},
		{"[^abc]", "(INLINE (ENDNOTE () (TEXT \"abc\")))"},
		{"[^abc ]", "(INLINE (ENDNOTE () (TEXT \"abc\")))"},
		{"[^abc\ndef]", "(INLINE (ENDNOTE () (TEXT \"abc\") (SOFT) (TEXT \"def\")))"},
		{"[^abc\n\ndef]", "(INLINE (TEXT \"[^abc\") (SOFT) (SOFT) (TEXT \"def]\"))"},
		{"[^abc[^def]]", "(INLINE (ENDNOTE () (TEXT \"abc\") (ENDNOTE () (TEXT \"def\"))))"},
		{"[^abc]{-}", "(INLINE (ENDNOTE ((\"-\" . \"\")) (TEXT \"abc\")))"},
	})
	checkTcs(t, true, TestCases{
		{"[^abc\n\ndef]", "(BLOCK (PARA (TEXT \"[^abc\")) (PARA (TEXT \"def]\")))"},
	})
}

func TestMark(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"[!", "(INLINE (TEXT \"[!\"))"},
		{"[!\n", "(INLINE (TEXT \"[!\"))"},
		{"[!]", "(INLINE (MARK \"\" \"\" \"\"))"},
		{"[!][!]", "(INLINE (MARK \"\" \"\" \"\") (MARK \"\" \"\" \"\"))"},
		{"[! ]", "(INLINE (TEXT \"[!\") (SPACE) (TEXT \"]\"))"},
		{"[!a]", "(INLINE (MARK \"a\" \"\" \"\"))"},
		{"[!a][!a]", "(INLINE (MARK \"a\" \"\" \"\") (MARK \"a\" \"\" \"\"))"},
		{"[!a ]", "(INLINE (TEXT \"[!a\") (SPACE) (TEXT \"]\"))"},
		{"[!a_]", "(INLINE (MARK \"a_\" \"\" \"\"))"},
		{"[!a_][!a]", "(INLINE (MARK \"a_\" \"\" \"\") (MARK \"a\" \"\" \"\"))"},
		{"[!a-b]", "(INLINE (MARK \"a-b\" \"\" \"\"))"},
		{"[!a|b]", "(INLINE (MARK \"a\" \"\" \"\" (TEXT \"b\")))"},
		{"[!a|]", "(INLINE (MARK \"a\" \"\" \"\"))"},
		{"[!|b]", "(INLINE (MARK \"\" \"\" \"\" (TEXT \"b\")))"},
		{"[!|b ]", "(INLINE (MARK \"\" \"\" \"\" (TEXT \"b\")))"},
		{"[!|b c]", "(INLINE (MARK \"\" \"\" \"\" (TEXT \"b\") (SPACE) (TEXT \"c\")))"},
	})
}

func TestComment(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"%", "(INLINE (TEXT \"%\"))"},
		{"%%", "(INLINE (LITERAL-COMMENT () \"\"))"},
		{"%\n", "(INLINE (TEXT \"%\"))"},
		{"%%\n", "(INLINE (LITERAL-COMMENT () \"\"))"},
		{"%%a", "(INLINE (LITERAL-COMMENT () \"a\"))"},
		{"%%%a", "(INLINE (LITERAL-COMMENT () \"a\"))"},
		{"%% a", "(INLINE (LITERAL-COMMENT () \"a\"))"},
		{"%%%  a", "(INLINE (LITERAL-COMMENT () \"a\"))"},
		{"%% % a", "(INLINE (LITERAL-COMMENT () \"% a\"))"},
		{"%%a", "(INLINE (LITERAL-COMMENT () \"a\"))"},
		{"a%%b", "(INLINE (TEXT \"a\") (LITERAL-COMMENT () \"b\"))"},
		{"a %%b", "(INLINE (TEXT \"a\") (SPACE) (LITERAL-COMMENT () \"b\"))"},
		{" %%b", "(INLINE (LITERAL-COMMENT () \"b\"))"},
		{"%%b ", "(INLINE (LITERAL-COMMENT () \"b \"))"},
		{"100%", "(INLINE (TEXT \"100%\"))"},
		{"%%{=}a", "(INLINE (LITERAL-COMMENT ((\"\" . \"\")) \"a\"))"},
	})
}

func TestFormat(t *testing.T) {
	symMap := symbolMap{
		"_": sz.SymFormatEmph,
		"*": sz.SymFormatStrong,
		">": sz.SymFormatInsert,
		"~": sz.SymFormatDelete,
		"^": sz.SymFormatSuper,
		",": sz.SymFormatSub,
		"#": sz.SymFormatMark,
		":": sz.SymFormatSpan,
	}
	t.Parallel()
	// Not for Insert / '>', because collision with quoted list
	// Not for Quote / '"', because escaped representation.
	for _, ch := range []string{"_", "*", "~", "^", ",", "#", ":"} {
		checkTcs(t, false, replace(ch, symMap, TestCases{
			{"$", "(INLINE (TEXT \"$\"))"},
			{"$$", "(INLINE (TEXT \"$$\"))"},
			{"$$$", "(INLINE (TEXT \"$$$\"))"},
			{"$$$$", "(INLINE ($% ()))"},
		}))
	}
	// Not for Quote / '"', because escaped representation.
	for _, ch := range []string{"_", "*", ">", "~", "^", ",", "#", ":"} {
		checkTcs(t, false, replace(ch, symMap, TestCases{
			{"$$a$$", "(INLINE ($% () (TEXT \"a\")))"},
			{"$$a$$$", "(INLINE ($% () (TEXT \"a\")) (TEXT \"$\"))"},
			{"$$$a$$", "(INLINE ($% () (TEXT \"$a\")))"},
			{"$$$a$$$", "(INLINE ($% () (TEXT \"$a\")) (TEXT \"$\"))"},
			{"$\\$", "(INLINE (TEXT \"$$\"))"},
			{"$\\$$", "(INLINE (TEXT \"$$$\"))"},
			{"$$\\$", "(INLINE (TEXT \"$$$\"))"},
			{"$$a\\$$", "(INLINE (TEXT \"$$a$$\"))"},
			{"$$a$\\$", "(INLINE (TEXT \"$$a$$\"))"},
			{"$$a\\$$$", "(INLINE ($% () (TEXT \"a$\")))"},
			{"$$a\na$$", "(INLINE ($% () (TEXT \"a\") (SOFT) (TEXT \"a\")))"},
			{"$$a\n\na$$", "(INLINE (TEXT \"$$a\") (SOFT) (SOFT) (TEXT \"a$$\"))"},
			{"$$a$${go}", "(INLINE ($% ((\"go\" . \"\")) (TEXT \"a\")))"},
		}))
		checkTcs(t, true, replace(ch, symMap, TestCases{
			{"$$a\n\na$$", "(BLOCK (PARA (TEXT \"$$a\")) (PARA (TEXT \"a$$\")))"},
		}))
	}
	checkTcs(t, false, replace(`"`, symbolMap{`"`: sz.SymFormatQuote}, TestCases{
		{"$", "(INLINE (TEXT \"\\\"\"))"},
		{"$$", "(INLINE (TEXT \"\\\"\\\"\"))"},
		{"$$$", "(INLINE (TEXT \"\\\"\\\"\\\"\"))"},
		{"$$$$", "(INLINE ($% ()))"},

		{"$$a$$", "(INLINE ($% () (TEXT \"a\")))"},
		{"$$a$$$", "(INLINE ($% () (TEXT \"a\")) (TEXT \"\\\"\"))"},
		{"$$$a$$", "(INLINE ($% () (TEXT \"\\\"a\")))"},
		{"$$$a$$$", "(INLINE ($% () (TEXT \"\\\"a\")) (TEXT \"\\\"\"))"},
		{"$\\$", "(INLINE (TEXT \"\\\"\\\"\"))"},
		{"$\\$$", "(INLINE (TEXT \"\\\"\\\"\\\"\"))"},
		{"$$\\$", "(INLINE (TEXT \"\\\"\\\"\\\"\"))"},
		{"$$a\\$$", "(INLINE (TEXT \"\\\"\\\"a\\\"\\\"\"))"},
		{"$$a$\\$", "(INLINE (TEXT \"\\\"\\\"a\\\"\\\"\"))"},
		{"$$a\\$$$", "(INLINE ($% () (TEXT \"a\\\"\")))"},
		{"$$a\na$$", "(INLINE ($% () (TEXT \"a\") (SOFT) (TEXT \"a\")))"},
		{"$$a\n\na$$", "(INLINE (TEXT \"\\\"\\\"a\") (SOFT) (SOFT) (TEXT \"a\\\"\\\"\"))"},
		{"$$a$${go}", "(INLINE ($% ((\"go\" . \"\")) (TEXT \"a\")))"},
	}))
	checkTcs(t, false, TestCases{
		{"__****__", "(INLINE (FORMAT-EMPH () (FORMAT-STRONG ())))"},
		{"__**a**__", "(INLINE (FORMAT-EMPH () (FORMAT-STRONG () (TEXT \"a\"))))"},
		{"__**__**", "(INLINE (TEXT \"__\") (FORMAT-STRONG () (TEXT \"__\")))"},
	})
}

func TestLiteral(t *testing.T) {
	symMap := symbolMap{
		"@": sz.SymLiteralZettel,
		"`": sz.SymLiteralProg,
		"'": sz.SymLiteralInput,
		"=": sz.SymLiteralOutput,
	}
	t.Parallel()
	for _, ch := range []string{"@", "`", "'", "="} {
		checkTcs(t, false, replace(ch, symMap, TestCases{
			{"$", "(INLINE (TEXT \"$\"))"},
			{"$$", "(INLINE (TEXT \"$$\"))"},
			{"$$$", "(INLINE (TEXT \"$$$\"))"},
			{"$$$$", "(INLINE ($% () \"\"))"},
			{"$$a$$", "(INLINE ($% () \"a\"))"},
			{"$$a$$$", "(INLINE ($% () \"a\") (TEXT \"$\"))"},
			{"$$$a$$", "(INLINE ($% () \"$a\"))"},
			{"$$$a$$$", "(INLINE ($% () \"$a\") (TEXT \"$\"))"},
			{"$\\$", "(INLINE (TEXT \"$$\"))"},
			{"$\\$$", "(INLINE (TEXT \"$$$\"))"},
			{"$$\\$", "(INLINE (TEXT \"$$$\"))"},
			{"$$a\\$$", "(INLINE (TEXT \"$$a$$\"))"},
			{"$$a$\\$", "(INLINE (TEXT \"$$a$$\"))"},
			{"$$a\\$$$", "(INLINE ($% () \"a$\"))"},
			{"$$a$${go}", "(INLINE ($% ((\"go\" . \"\")) \"a\"))"},
		}))
	}
	checkTcs(t, false, TestCases{
		{"''````''", "(INLINE (LITERAL-INPUT () \"````\"))"},
		{"''``a``''", "(INLINE (LITERAL-INPUT () \"``a``\"))"},
		{"''``''``", "(INLINE (LITERAL-INPUT () \"``\") (TEXT \"``\"))"},
		{"''\\'''", "(INLINE (LITERAL-INPUT () \"'\"))"},
	})
	checkTcs(t, false, TestCases{
		{"@@HTML@@{=html}", "(INLINE (LITERAL-HTML () \"HTML\"))"},
		{"@@HTML@@{=html lang=en}", "(INLINE (LITERAL-HTML ((\"lang\" . \"en\")) \"HTML\"))"},
		{"@@HTML@@{=html,lang=en}", "(INLINE (LITERAL-HTML ((\"lang\" . \"en\")) \"HTML\"))"},
	})
}

func TestLiteralMath(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"$", "(INLINE (TEXT \"$\"))"},
		{"$$", "(INLINE (TEXT \"$$\"))"},
		{"$$$", "(INLINE (TEXT \"$$$\"))"},
		{"$$$$", "(INLINE (LITERAL-MATH () \"\"))"},
		{"$$a$$", "(INLINE (LITERAL-MATH () \"a\"))"},
		{"$$a$$$", "(INLINE (LITERAL-MATH () \"a\") (TEXT \"$\"))"},
		{"$$$a$$", "(INLINE (LITERAL-MATH () \"$a\"))"},
		{"$$$a$$$", "(INLINE (LITERAL-MATH () \"$a\") (TEXT \"$\"))"},
		{`$\$`, "(INLINE (TEXT \"$$\"))"},
		{`$\$$`, "(INLINE (TEXT \"$$$\"))"},
		{`$$\$`, "(INLINE (TEXT \"$$$\"))"},
		{`$$a\$$`, "(INLINE (LITERAL-MATH () \"a\\\\\"))"},
		{`$$a$\$`, "(INLINE (TEXT \"$$a$$\"))"},
		{`$$a\$$$`, "(INLINE (LITERAL-MATH () \"a\\\\\") (TEXT \"$\"))"},
		{"$$a$${go}", "(INLINE (LITERAL-MATH ((\"go\" . \"\")) \"a\"))"},
	})
}

func TestMixFormatCode(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"__abc__\n**def**", "(INLINE (FORMAT-EMPH () (TEXT \"abc\")) (SOFT) (FORMAT-STRONG () (TEXT \"def\")))"},
		{"''abc''\n==def==", "(INLINE (LITERAL-INPUT () \"abc\") (SOFT) (LITERAL-OUTPUT () \"def\"))"},
		{"__abc__\n==def==", "(INLINE (FORMAT-EMPH () (TEXT \"abc\")) (SOFT) (LITERAL-OUTPUT () \"def\"))"},
		{"__abc__\n``def``", "(INLINE (FORMAT-EMPH () (TEXT \"abc\")) (SOFT) (LITERAL-CODE () \"def\"))"},
		{
			"\"\"ghi\"\"\n::abc::\n``def``\n",
			"(INLINE (FORMAT-QUOTE () (TEXT \"ghi\")) (SOFT) (FORMAT-SPAN () (TEXT \"abc\")) (SOFT) (LITERAL-CODE () \"def\"))",
		},
	})
}

func TestNDash(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"--", "(INLINE (TEXT \"\u2013\"))"},
		{"a--b", "(INLINE (TEXT \"a\u2013b\"))"},
	})
}

func TestEntity(t *testing.T) {
	t.Parallel()
	checkTcs(t, false, TestCases{
		{"&", "(INLINE (TEXT \"&\"))"},
		{"&;", "(INLINE (TEXT \"&;\"))"},
		{"&#;", "(INLINE (TEXT \"&#;\"))"},
		{"&#1a;", "(INLINE (TEXT \"&#1a;\"))"},
		{"&#x;", "(INLINE (TEXT \"&#x;\"))"},
		{"&#x0z;", "(INLINE (TEXT \"&#x0z;\"))"},
		{"&1;", "(INLINE (TEXT \"&1;\"))"},
		{"&#9;", "(INLINE (TEXT \"&#9;\"))"}, // Numeric entities below space are not allowed.
		{"&#x1f;", "(INLINE (TEXT \"&#x1f;\"))"},

		// Good cases
		{"&lt;", "(INLINE (TEXT \"<\"))"},
		{"&#48;", "(INLINE (TEXT \"0\"))"},
		{"&#x4A;", "(INLINE (TEXT \"J\"))"},
		{"&#X4a;", "(INLINE (TEXT \"J\"))"},
		{"&hellip;", "(INLINE (TEXT \"\u2026\"))"},
		{"&nbsp;", "(INLINE (TEXT \"\u00a0\"))"},
		{"E: &amp;,&#63;;&#x63;.", "(INLINE (TEXT \"E:\") (SPACE) (TEXT \"&,?;c.\"))"},
	})
}

func TestVerbatimZettel(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"@@@\n@@@", "()"},
		{"@@@\nabc\n@@@", "(BLOCK (VERBATIM-ZETTEL () \"abc\"))"},
		{"@@@@def\nabc\n@@@@", "(BLOCK (VERBATIM-ZETTEL ((\"\" . \"def\")) \"abc\"))"},
	})
}

func TestVerbatimCode(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, TestCases{
		{"%%%\n%%%", "()"},
		{"%%%\nabc\n%%%", "(BLOCK (VERBATIM-COMMENT () \"abc\"))"},
		{"%%%%go\nabc\n%%%%", "(BLOCK (VERBATIM-COMMENT ((\"\" . \"go\")) \"abc\"))"},
	})
}

func TestPara(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"a\n\nb", "(BLOCK (PARA (TEXT \"a\")) (PARA (TEXT \"b\")))"},
		{"a\n \nb", "(BLOCK (PARA (TEXT \"a\") (SOFT) (HARD) (TEXT \"b\")))"},
	})
}

func TestSpanRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, replace("\"", nil, TestCases{
		{"$$$\n$$$", "()"},
		{"$$$\nabc\n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")))))"},
		{"$$$\nabc\n$$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")))))"},
		{"$$$$\nabc\n$$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")))))"},
		{"$$$\nabc\ndef\n$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\") (HARD) (TEXT \"def\")))))"},
		{"$$$$\nabc\n$$$\ndef\n$$$\n$$$$", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\")) (REGION-VERSE () ((PARA (TEXT \"def\")))))))"},
		{"$$$go\n$$$x", "(BLOCK (REGION-VERSE ((\"\" . \"go\")) () (TEXT \"x\")))"},
		{"$$$\nabc\n$$$ def ", "(BLOCK (REGION-VERSE () ((PARA (TEXT \"abc\"))) (TEXT \"def\")))"},
		{"$$$\n space \n$$$", "(BLOCK (REGION-VERSE () ((PARA (SPACE \"\u00a0\") (TEXT \"space\")))))"},
		{"$$$\n  spaces  \n$$$", "(BLOCK (REGION-VERSE () ((PARA (SPACE \"\u00a0\u00a0\") (TEXT \"spaces\")))))"},
		{"$$$\n  spaces  \n space  \n$$$", "(BLOCK (REGION-VERSE () ((PARA (SPACE \"\u00a0\u00a0\") (TEXT \"spaces\") (SPACE \"\u00a0\u00a0\") (HARD) (SPACE \"\u00a0\") (TEXT \"space\")))))"},
	}))
}

func TestHeading(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"=h", "(BLOCK (PARA (TEXT \"=h\")))"},
		{"= h", "(BLOCK (PARA (TEXT \"=\") (SPACE) (TEXT \"h\")))"},
		{"==h", "(BLOCK (PARA (TEXT \"==h\")))"},
		{"== h", "(BLOCK (PARA (TEXT \"==\") (SPACE) (TEXT \"h\")))"},
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
		{"=== h i {-}", "(BLOCK (HEADING 1 ((\"-\" . \"\")) \"\" \"\" (TEXT \"h\") (SPACE) (TEXT \"i\")))"},
		{"=== h {{a}}", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\") (SPACE) (EMBED () \"a\")))"},
		{"=== h{{a}}", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\") (EMBED () \"a\")))"},
		{"=== {{a}}", "(BLOCK (HEADING 1 () \"\" \"\" (EMBED () \"a\")))"},
		{"=== h {{a}}{-}", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\") (SPACE) (EMBED ((\"-\" . \"\")) \"a\")))"},
		{"=== h {{a}} {-}", "(BLOCK (HEADING 1 ((\"-\" . \"\")) \"\" \"\" (TEXT \"h\") (SPACE) (EMBED () \"a\")))"},
		{"=== h {-}{{a}}", "(BLOCK (HEADING 1 ((\"-\" . \"\")) \"\" \"\" (TEXT \"h\")))"},
		{"=== h{id=abc}", "(BLOCK (HEADING 1 ((\"id\" . \"abc\")) \"\" \"\" (TEXT \"h\")))"},
		{"=== h\n=== h", "(BLOCK (HEADING 1 () \"\" \"\" (TEXT \"h\")) (HEADING 1 () \"\" \"\" (TEXT \"h\")))"},
	})
}

func TestHRule(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"-", "(BLOCK (PARA (TEXT \"-\")))"},
		{"---", "(BLOCK (THEMATIC ()))"},
		{"----", "(BLOCK (THEMATIC ()))"},
		{"---A", "(BLOCK (THEMATIC ((\"\" . \"A\"))))"},
		{"---A-", "(BLOCK (THEMATIC ((\"\" . \"A-\"))))"},
		{"-1", "(BLOCK (PARA (TEXT \"-1\")))"},
		{"2-1", "(BLOCK (PARA (TEXT \"2-1\")))"},
		{"---  {  go  }  ", "(BLOCK (THEMATIC ((\"go\" . \"\"))))"},
		{"---  {  .go  }  ", "(BLOCK (THEMATIC ((\"class\" . \"go\"))))"},
	})
}

func TestList(t *testing.T) {
	t.Parallel()
	// No ">" in the following, because quotation lists may have empty items.
	for _, ch := range []string{"*", "#"} {
		checkTcs(t, true, replace(ch, nil, TestCases{
			{"$", "(BLOCK (PARA (TEXT \"$\")))"},
			{"$$", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$$$", "(BLOCK (PARA (TEXT \"$$$\")))"},
			{"$ ", "(BLOCK (PARA (TEXT \"$\")))"},
			{"$$ ", "(BLOCK (PARA (TEXT \"$$\")))"},
			{"$$$ ", "(BLOCK (PARA (TEXT \"$$$\")))"},
		}))
	}
	checkTcs(t, true, TestCases{
		{"* abc", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))))"},
		{"** abc", "(BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))))))"},
		{"*** abc", "(BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))))))))"},
		{"**** abc", "(BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))))))))))"},
		{"** abc\n**** def", "(BLOCK (UNORDERED (BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (UNORDERED (BLOCK (UNORDERED (BLOCK (PARA (TEXT \"def\")))))))))))"},
		{"* abc\ndef", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n* def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\"))) (BLOCK (PARA (TEXT \"def\")))))"},
		{"* abc\n  def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\") (SOFT) (TEXT \"def\")))))"},
		{"* abc\n   def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\") (SOFT) (TEXT \"def\")))))"},
		{"* abc\n\ndef", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n\n def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))) (PARA (TEXT \"def\")))"},
		{"* abc\n\n  def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (PARA (TEXT \"def\")))))"},
		{"* abc\n\n   def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (PARA (TEXT \"def\")))))"},
		{"* abc\n** def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (UNORDERED (BLOCK (PARA (TEXT \"def\")))))))"},
		{"* abc\n** def\n* ghi", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (UNORDERED (BLOCK (PARA (TEXT \"def\"))))) (BLOCK (PARA (TEXT \"ghi\")))))"},
		{"* abc\n\n  def\n* ghi", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (PARA (TEXT \"def\"))) (BLOCK (PARA (TEXT \"ghi\")))))"},
		{"* abc\n** def\n   ghi\n  jkl", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")) (UNORDERED (BLOCK (PARA (TEXT \"def\") (SOFT) (TEXT \"ghi\")))) (PARA (TEXT \"jkl\")))))"},

		// A list does not last beyond a region
		{":::\n# abc\n:::\n# def", "(BLOCK (REGION-BLOCK () ((ORDERED (BLOCK (PARA (TEXT \"abc\")))))) (ORDERED (BLOCK (PARA (TEXT \"def\")))))"},

		// A HRule creates a new list
		{"* abc\n---\n* def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))) (THEMATIC ()) (UNORDERED (BLOCK (PARA (TEXT \"def\")))))"},

		// Changing list type adds a new list
		{"* abc\n# def", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))) (ORDERED (BLOCK (PARA (TEXT \"def\")))))"},

		// Quotation lists may have empty items
		{">", "(BLOCK (QUOTATION (BLOCK)))"},

		// Empty continuation
		{"* abc\n  ", "(BLOCK (UNORDERED (BLOCK (PARA (TEXT \"abc\")))))"},
	})
}

func TestQuoteList(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"> w1 w2", "(BLOCK (QUOTATION (BLOCK (PARA (TEXT \"w1\") (SPACE) (TEXT \"w2\")))))"},
		{"> w1\n> w2", "(BLOCK (QUOTATION (BLOCK (PARA (TEXT \"w1\") (SOFT) (TEXT \"w2\")))))"},
		{"> w1\n>w2", "(BLOCK (QUOTATION (BLOCK (PARA (TEXT \"w1\")))) (PARA (TEXT \">w2\")))"},
		{"> w1\n>\n>w2", "(BLOCK (QUOTATION (BLOCK (PARA (TEXT \"w1\"))) (BLOCK)) (PARA (TEXT \">w2\")))"},
		{"> w1\n> \n> w2", "(BLOCK (QUOTATION (BLOCK (PARA (TEXT \"w1\"))) (BLOCK) (BLOCK (PARA (TEXT \"w2\")))))"},
	})
}

func TestEnumAfterPara(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"abc\n* def", "(BLOCK (PARA (TEXT \"abc\")) (UNORDERED (BLOCK (PARA (TEXT \"def\")))))"},
		{"abc\n*def", "(BLOCK (PARA (TEXT \"abc\") (SOFT) (TEXT \"*def\")))"},
	})
}

func TestDefinition(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{";", "(BLOCK (PARA (TEXT \";\")))"},
		{"; ", "(BLOCK (PARA (TEXT \";\")))"},
		{"; abc", "(BLOCK (DESCRIPTION ((TEXT \"abc\"))))"},
		{"; abc\ndef", "(BLOCK (DESCRIPTION ((TEXT \"abc\"))) (PARA (TEXT \"def\")))"},
		{"; abc\n def", "(BLOCK (DESCRIPTION ((TEXT \"abc\"))) (PARA (TEXT \"def\")))"},
		{"; abc\n  def", "(BLOCK (DESCRIPTION ((TEXT \"abc\") (SOFT) (TEXT \"def\"))))"},
		{"; abc\n  def\n  ghi", "(BLOCK (DESCRIPTION ((TEXT \"abc\") (SOFT) (TEXT \"def\") (SOFT) (TEXT \"ghi\"))))"},
		{":", "(BLOCK (PARA (TEXT \":\")))"},
		{": ", "(BLOCK (PARA (TEXT \":\")))"},
		{": abc", "(BLOCK (PARA (TEXT \":\") (SPACE) (TEXT \"abc\")))"},
		{"; abc\n: def", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))))"},
		{"; abc\n: def\nghi", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))) (PARA (TEXT \"ghi\")))"},
		{"; abc\n: def\n ghi", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))) (PARA (TEXT \"ghi\")))"},
		{"; abc\n: def\n  ghi", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\") (SOFT) (TEXT \"ghi\"))))))"},
		{"; abc\n: def\n\n  ghi", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))) (PARA (TEXT \"ghi\")))))"},
		{"; abc\n:", "(BLOCK (DESCRIPTION ((TEXT \"abc\"))) (PARA (TEXT \":\")))"},
		{"; abc\n: def\n: ghi", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))) (BLOCK (PARA (TEXT \"ghi\"))))))"},
		{"; abc\n: def\n; ghi\n: jkl", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\")))) ((TEXT \"ghi\")) (BLOCK (BLOCK (PARA (TEXT \"jkl\"))))))"},

		// Empty description
		{"; abc\n: ", "(BLOCK (DESCRIPTION ((TEXT \"abc\"))) (PARA (TEXT \":\")))"},
		// Empty continuation of definition
		{"; abc\n: def\n  ", "(BLOCK (DESCRIPTION ((TEXT \"abc\")) (BLOCK (BLOCK (PARA (TEXT \"def\"))))))"},
	})
}

func TestTable(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"|", "()"},
		{"||", "(BLOCK (TABLE () ((CELL))))"},
		{"| |", "(BLOCK (TABLE () ((CELL))))"},
		{"|a", "(BLOCK (TABLE () ((CELL (TEXT \"a\")))))"},
		{"|a|", "(BLOCK (TABLE () ((CELL (TEXT \"a\")))))"},
		{"|a| ", "(BLOCK (TABLE () ((CELL (TEXT \"a\")) (CELL))))"},
		{"|a|b", "(BLOCK (TABLE () ((CELL (TEXT \"a\")) (CELL (TEXT \"b\")))))"},
		{"|a\n|b", "(BLOCK (TABLE () ((CELL (TEXT \"a\"))) ((CELL (TEXT \"b\")))))"},
		{"|a|b\n|c|d", "(BLOCK (TABLE () ((CELL (TEXT \"a\")) (CELL (TEXT \"b\"))) ((CELL (TEXT \"c\")) (CELL (TEXT \"d\")))))"},
		{"|%", "()"},
		{"|=a", "(BLOCK (TABLE ((CELL (TEXT \"a\")))))"},
		{"|=a\n|b", "(BLOCK (TABLE ((CELL (TEXT \"a\"))) ((CELL (TEXT \"b\")))))"},
		{"|a|b\n|%---\n|c|d", "(BLOCK (TABLE () ((CELL (TEXT \"a\")) (CELL (TEXT \"b\"))) ((CELL (TEXT \"c\")) (CELL (TEXT \"d\")))))"},
		{"|a|b\n|c", "(BLOCK (TABLE () ((CELL (TEXT \"a\")) (CELL (TEXT \"b\"))) ((CELL (TEXT \"c\")) (CELL))))"},
		{"|=<a>\n|b|c", "(BLOCK (TABLE ((CELL-LEFT (TEXT \"a\")) (CELL)) ((CELL-RIGHT (TEXT \"b\")) (CELL (TEXT \"c\")))))"},
		{"|=<a|=b>\n||", "(BLOCK (TABLE ((CELL-LEFT (TEXT \"a\")) (CELL-RIGHT (TEXT \"b\"))) ((CELL) (CELL-RIGHT))))"},
	})
}

func TestTransclude(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"{{{a}}}", "(BLOCK (TRANSCLUDE () (EXTERNAL \"a\")))"},
		{"{{{a}}}b", "(BLOCK (TRANSCLUDE ((\"\" . \"b\")) (EXTERNAL \"a\")))"},
		{"{{{a}}}}", "(BLOCK (TRANSCLUDE () (EXTERNAL \"a\")))"},
		{"{{{a\\}}}}", "(BLOCK (TRANSCLUDE () (EXTERNAL \"a\\\\}\")))"},
		{"{{{a\\}}}}b", "(BLOCK (TRANSCLUDE ((\"\" . \"b\")) (EXTERNAL \"a\\\\}\")))"},
		{"{{{a}}", "(BLOCK (PARA (TEXT \"{\") (EMBED () \"a\")))"},
		{"{{{a}}}{go=b}", "(BLOCK (TRANSCLUDE ((\"go\" . \"b\")) (EXTERNAL \"a\")))"},
	})
}

func TestBlockAttr(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, replace("\"", nil, TestCases{
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
	checkTcs(t, true, TestCases{
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
	checkTcs(t, true, TestCases{
		{"::a::{py=3}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2 3\"}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2 3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2\\\"3\"}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2\\\"3\")) (TEXT \"a\"))))"},
		{"::a::{py=2\"3}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2\\\"3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2\n3\"}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2\\n3\")) (TEXT \"a\"))))"},
		{"::a::{py=\"2 3}", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\")) (TEXT \"{py=\\\"2\") (SPACE) (TEXT \"3}\")))"},
	})
	checkTcs(t, true, TestCases{
		{"::a::{py=2 py=3}", "(BLOCK (PARA (FORMAT-SPAN ((\"py\" . \"2 3\")) (TEXT \"a\"))))"},
		{"::a::{.go .py}", "(BLOCK (PARA (FORMAT-SPAN ((\"class\" . \"go py\")) (TEXT \"a\"))))"},
	})
}

func TestTemp(t *testing.T) {
	t.Parallel()
	checkTcs(t, true, TestCases{
		{"", "()"},
	})
}
