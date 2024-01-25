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

	"zettelstore.de/client.fossil/input"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/client.fossil/zmk"
	"zettelstore.de/sx.fossil"
)

type TestCase struct{ source, want string }
type TestCases []TestCase
type symbolMap map[string]sx.Symbol

func replace(s string, sm symbolMap, tcs TestCases) TestCases {
	var sym string
	if len(sm) > 0 {
		sym = string(sm[s])
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

	for tcn, tc := range tcs {
		t.Run(fmt.Sprintf("TC=%02d,src=%q", tcn, tc.source), func(st *testing.T) {
			st.Helper()
			inp := input.NewInput([]byte(tc.source))
			bns := zmk.ParseBlocks(inp)
			got := bns.String()
			if tc.want != got {
				st.Errorf("\nwant=%q\n got=%q", tc.want, got)
			}
		})
	}
}

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
		{"ab cd", "(BLOCK (PARA (TEXT \"ab\") (SPACE) (TEXT \"cd\")))"},
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
		{"http://a, http://b", "(BLOCK (PARA (TEXT \"http://a,\") (SPACE) (TEXT \"http://b\")))"},
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
		{"[[ ]]", "(BLOCK (PARA (TEXT \"[[\") (SPACE) (TEXT \"]]\")))"},
		{"[[\n]]", "(BLOCK (PARA (TEXT \"[[\") (SOFT) (TEXT \"]]\")))"},
		{"[[ a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\")))"},
		{"[[a ]]", "(BLOCK (PARA (TEXT \"[[a\") (SPACE) (TEXT \"]]\")))"},
		{"[[a\n]]", "(BLOCK (PARA (TEXT \"[[a\") (SOFT) (TEXT \"]]\")))"},
		{"[[a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\")))"},
		{"[[12345678901234]]", "(BLOCK (PARA (LINK-ZETTEL () \"12345678901234\")))"},
		{"[[a]", "(BLOCK (PARA (TEXT \"[[a]\")))"},
		{"[[|a]]", "(BLOCK (PARA (TEXT \"[[|a]]\")))"},
		{"[[b|]]", "(BLOCK (PARA (TEXT \"[[b|]]\")))"},
		{"[[b|a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b\"))))"},
		{"[[b| a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b\"))))"},
		{"[[b%c|a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b%c\"))))"},
		{"[[b%%c|a]]", "(BLOCK (PARA (TEXT \"[[b\") (LITERAL-COMMENT () \"c|a]]\")))"},
		{"[[b|a]", "(BLOCK (PARA (TEXT \"[[b|a]\")))"},
		{"[[b\nc|a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b\") (SOFT) (TEXT \"c\"))))"},
		{"[[b c|a#n]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a#n\" (TEXT \"b\") (SPACE) (TEXT \"c\"))))"},
		{"[[a]]go", "(BLOCK (PARA (LINK-EXTERNAL () \"a\") (TEXT \"go\")))"},
		{"[[b|a]]{go}", "(BLOCK (PARA (LINK-EXTERNAL ((\"go\" . \"\")) \"a\" (TEXT \"b\"))))"},
		{"[[[[a]]|b]]", "(BLOCK (PARA (TEXT \"[[\") (LINK-EXTERNAL () \"a\") (TEXT \"|b]]\")))"},
		{"[[a[b]c|d]]", "(BLOCK (PARA (LINK-EXTERNAL () \"d\" (TEXT \"a[b]c\"))))"},
		{"[[[b]c|d]]", "(BLOCK (PARA (TEXT \"[\") (LINK-EXTERNAL () \"d\" (TEXT \"b]c\"))))"},
		{"[[a[]c|d]]", "(BLOCK (PARA (LINK-EXTERNAL () \"d\" (TEXT \"a[]c\"))))"},
		{"[[a[b]|d]]", "(BLOCK (PARA (LINK-EXTERNAL () \"d\" (TEXT \"a[b]\"))))"},
		{"[[\\|]]", "(BLOCK (PARA (LINK-EXTERNAL () \"\\\\|\")))"},
		{"[[\\||a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"|\"))))"},
		{"[[b\\||a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b|\"))))"},
		{"[[b\\|c|a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b|c\"))))"},
		{"[[\\]]]", "(BLOCK (PARA (LINK-EXTERNAL () \"\\\\]\")))"},
		{"[[\\]|a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"]\"))))"},
		{"[[b\\]|a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"b]\"))))"},
		{"[[\\]\\||a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"a\" (TEXT \"]|\"))))"},
		{"[[http://a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"http://a\")))"},
		{"[[http://a|http://a]]", "(BLOCK (PARA (LINK-EXTERNAL () \"http://a\" (TEXT \"http://a\"))))"},
		{"[[[[a]]]]", "(BLOCK (PARA (TEXT \"[[\") (LINK-EXTERNAL () \"a\") (TEXT \"]]\")))"},
		{"[[query:title]]", "(BLOCK (PARA (LINK-QUERY () \"title\")))"},
		{"[[query:title syntax]]", "(BLOCK (PARA (LINK-QUERY () \"title syntax\")))"},
		{"[[query:title | action]]", "(BLOCK (PARA (LINK-QUERY () \"title | action\")))"},
		{"[[Text|query:title]]", "(BLOCK (PARA (LINK-QUERY () \"title\" (TEXT \"Text\"))))"},
		{"[[Text|query:title syntax]]", "(BLOCK (PARA (LINK-QUERY () \"title syntax\" (TEXT \"Text\"))))"},
		{"[[Text|query:title | action]]", "(BLOCK (PARA (LINK-QUERY () \"title | action\" (TEXT \"Text\"))))"},
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
		{"{{ }}", "(BLOCK (PARA (TEXT \"{{\") (SPACE) (TEXT \"}}\")))"},
		{"{{\n}}", "(BLOCK (PARA (TEXT \"{{\") (SOFT) (TEXT \"}}\")))"},
		{"{{a }}", "(BLOCK (PARA (TEXT \"{{a\") (SPACE) (TEXT \"}}\")))"},
		{"{{a\n}}", "(BLOCK (PARA (TEXT \"{{a\") (SOFT) (TEXT \"}}\")))"},
		{"{{a}}", "(BLOCK (PARA (EMBED () \"a\")))"},
		{"{{12345678901234}}", "(BLOCK (PARA (EMBED () \"12345678901234\")))"},
		{"{{ a}}", "(BLOCK (PARA (EMBED () \"a\")))"},
		{"{{a}", "(BLOCK (PARA (TEXT \"{{a}\")))"},
		{"{{|a}}", "(BLOCK (PARA (TEXT \"{{|a}}\")))"},
		{"{{b|}}", "(BLOCK (PARA (TEXT \"{{b|}}\")))"},
		{"{{b|a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"b\"))))"},
		{"{{b| a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"b\"))))"},
		{"{{b|a}", "(BLOCK (PARA (TEXT \"{{b|a}\")))"},
		{"{{b\nc|a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"b\") (SOFT) (TEXT \"c\"))))"},
		{"{{b c|a#n}}", "(BLOCK (PARA (EMBED () \"a#n\" (TEXT \"b\") (SPACE) (TEXT \"c\"))))"},
		{"{{a}}{go}", "(BLOCK (PARA (EMBED ((\"go\" . \"\")) \"a\")))"},
		{"{{{{a}}|b}}", "(BLOCK (PARA (TEXT \"{{\") (EMBED () \"a\") (TEXT \"|b}}\")))"},
		{"{{\\|}}", "(BLOCK (PARA (EMBED () \"\\\\|\")))"},
		{"{{\\||a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"|\"))))"},
		{"{{b\\||a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"b|\"))))"},
		{"{{b\\|c|a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"b|c\"))))"},
		{"{{\\}}}", "(BLOCK (PARA (EMBED () \"\\\\}\")))"},
		{"{{\\}|a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"}\"))))"},
		{"{{b\\}|a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"b}\"))))"},
		{"{{\\}\\||a}}", "(BLOCK (PARA (EMBED () \"a\" (TEXT \"}|\"))))"},
		{"{{http://a}}", "(BLOCK (PARA (EMBED () \"http://a\")))"},
		{"{{http://a|http://a}}", "(BLOCK (PARA (EMBED () \"http://a\" (TEXT \"http://a\"))))"},
		{"{{{{a}}}}", "(BLOCK (PARA (TEXT \"{{\") (EMBED () \"a\") (TEXT \"}}\")))"},
	})
}

func TestCite(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[@", "(BLOCK (PARA (TEXT \"[@\")))"},
		{"[@]", "(BLOCK (PARA (TEXT \"[@]\")))"},
		{"[@a]", "(BLOCK (PARA (CITE () \"a\")))"},
		{"[@ a]", "(BLOCK (PARA (TEXT \"[@\") (SPACE) (TEXT \"a]\")))"},
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
}

func TestMark(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[!", "(BLOCK (PARA (TEXT \"[!\")))"},
		{"[!\n", "(BLOCK (PARA (TEXT \"[!\")))"},
		{"[!]", "(BLOCK (PARA (MARK \"\" \"\" \"\")))"},
		{"[!][!]", "(BLOCK (PARA (MARK \"\" \"\" \"\") (MARK \"\" \"\" \"\")))"},
		{"[! ]", "(BLOCK (PARA (TEXT \"[!\") (SPACE) (TEXT \"]\")))"},
		{"[!a]", "(BLOCK (PARA (MARK \"a\" \"\" \"\")))"},
		{"[!a][!a]", "(BLOCK (PARA (MARK \"a\" \"\" \"\") (MARK \"a\" \"\" \"\")))"},
		{"[!a ]", "(BLOCK (PARA (TEXT \"[!a\") (SPACE) (TEXT \"]\")))"},
		{"[!a_]", "(BLOCK (PARA (MARK \"a_\" \"\" \"\")))"},
		{"[!a_][!a]", "(BLOCK (PARA (MARK \"a_\" \"\" \"\") (MARK \"a\" \"\" \"\")))"},
		{"[!a-b]", "(BLOCK (PARA (MARK \"a-b\" \"\" \"\")))"},
		{"[!a|b]", "(BLOCK (PARA (MARK \"a\" \"\" \"\" (TEXT \"b\"))))"},
		{"[!a|]", "(BLOCK (PARA (MARK \"a\" \"\" \"\")))"},
		{"[!|b]", "(BLOCK (PARA (MARK \"\" \"\" \"\" (TEXT \"b\"))))"},
		{"[!|b ]", "(BLOCK (PARA (MARK \"\" \"\" \"\" (TEXT \"b\"))))"},
		{"[!|b c]", "(BLOCK (PARA (MARK \"\" \"\" \"\" (TEXT \"b\") (SPACE) (TEXT \"c\"))))"},
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
		{"a %%b", "(BLOCK (PARA (TEXT \"a\") (SPACE) (LITERAL-COMMENT () \"b\")))"},
		{" %%b", "(BLOCK (PARA (LITERAL-COMMENT () \"b\")))"},
		{"%%b ", "(BLOCK (PARA (LITERAL-COMMENT () \"b \")))"},
		{"100%", "(BLOCK (PARA (TEXT \"100%\")))"},
		{"%%{=}a", "(BLOCK (PARA (LITERAL-COMMENT ((\"\" . \"\")) \"a\")))"},
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
	}
	checkTcs(t, replace(`"`, symbolMap{`"`: sz.SymFormatQuote}, TestCases{
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
		"@": sz.SymLiteralZettel,
		"`": sz.SymLiteralProg,
		"'": sz.SymLiteralInput,
		"=": sz.SymLiteralOutput,
	}
	t.Parallel()
	for _, ch := range []string{"@", "`", "'", "="} {
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
		{"''````''", "(BLOCK (PARA (LITERAL-INPUT () \"````\")))"},
		{"''``a``''", "(BLOCK (PARA (LITERAL-INPUT () \"``a``\")))"},
		{"''``''``", "(BLOCK (PARA (LITERAL-INPUT () \"``\") (TEXT \"``\")))"},
		{"''\\'''", "(BLOCK (PARA (LITERAL-INPUT () \"'\")))"},
	})
	checkTcs(t, TestCases{
		{"@@HTML@@{=html}", "(BLOCK (PARA (LITERAL-HTML () \"HTML\")))"},
		{"@@HTML@@{=html lang=en}", "(BLOCK (PARA (LITERAL-HTML ((\"lang\" . \"en\")) \"HTML\")))"},
		{"@@HTML@@{=html,lang=en}", "(BLOCK (PARA (LITERAL-HTML ((\"lang\" . \"en\")) \"HTML\")))"},
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
		{"E: &amp;,&#63;;&#x63;.", "(BLOCK (PARA (TEXT \"E:\") (SPACE) (TEXT \"&,?;c.\")))"},
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
		{"a\n \nb", "(BLOCK (PARA (TEXT \"a\") (SOFT) (HARD) (TEXT \"b\")))"},
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
		{"$$$\n space \n$$$", "(BLOCK (REGION-VERSE () ((PARA (SPACE \"\u00a0\") (TEXT \"space\")))))"},
		{"$$$\n  spaces  \n$$$", "(BLOCK (REGION-VERSE () ((PARA (SPACE \"\u00a0\u00a0\") (TEXT \"spaces\")))))"},
		{"$$$\n  spaces  \n space  \n$$$", "(BLOCK (REGION-VERSE () ((PARA (SPACE \"\u00a0\u00a0\") (TEXT \"spaces\") (SPACE \"\u00a0\u00a0\") (HARD) (SPACE \"\u00a0\") (TEXT \"space\")))))"},
	}))
}

func TestHeading(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
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
	})
}

func xTestList(t *testing.T) {
	t.Parallel()
	// No ">" in the following, because quotation lists may have empty items.
	for _, ch := range []string{"*", "#"} {
		checkTcs(t, replace(ch, nil, TestCases{
			{"$", "(PARA $)"},
			{"$$", "(PARA $$)"},
			{"$$$", "(PARA $$$)"},
			{"$ ", "(PARA $)"},
			{"$$ ", "(PARA $$)"},
			{"$$$ ", "(PARA $$$)"},
		}))
	}
	checkTcs(t, TestCases{
		{"* abc", "(UL {(PARA abc)})"},
		{"** abc", "(UL {(UL {(PARA abc)})})"},
		{"*** abc", "(UL {(UL {(UL {(PARA abc)})})})"},
		{"**** abc", "(UL {(UL {(UL {(UL {(PARA abc)})})})})"},
		{"** abc\n**** def", "(UL {(UL {(PARA abc)(UL {(UL {(PARA def)})})})})"},
		{"* abc\ndef", "(UL {(PARA abc)})(PARA def)"},
		{"* abc\n def", "(UL {(PARA abc)})(PARA def)"},
		{"* abc\n* def", "(UL {(PARA abc)} {(PARA def)})"},
		{"* abc\n  def", "(UL {(PARA abc SB def)})"},
		{"* abc\n   def", "(UL {(PARA abc SB def)})"},
		{"* abc\n\ndef", "(UL {(PARA abc)})(PARA def)"},
		{"* abc\n\n def", "(UL {(PARA abc)})(PARA def)"},
		{"* abc\n\n  def", "(UL {(PARA abc)(PARA def)})"},
		{"* abc\n\n   def", "(UL {(PARA abc)(PARA def)})"},
		{"* abc\n** def", "(UL {(PARA abc)(UL {(PARA def)})})"},
		{"* abc\n** def\n* ghi", "(UL {(PARA abc)(UL {(PARA def)})} {(PARA ghi)})"},
		{"* abc\n\n  def\n* ghi", "(UL {(PARA abc)(PARA def)} {(PARA ghi)})"},
		{"* abc\n** def\n   ghi\n  jkl", "(UL {(PARA abc)(UL {(PARA def SB ghi)})(PARA jkl)})"},

		// A list does not last beyond a region
		{":::\n# abc\n:::\n# def", "(SPAN (OL {(PARA abc)}))(OL {(PARA def)})"},

		// A HRule creates a new list
		{"* abc\n---\n* def", "(UL {(PARA abc)})(HR)(UL {(PARA def)})"},

		// Changing list type adds a new list
		{"* abc\n# def", "(UL {(PARA abc)})(OL {(PARA def)})"},

		// Quotation lists may have empty items
		{">", "(QL {})"},
	})
}

func xTestQuoteList(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"> w1 w2", "(QL {(PARA w1 SP w2)})"},
		{"> w1\n> w2", "(QL {(PARA w1 SB w2)})"},
		{"> w1\n>\n>w2", "(QL {(PARA w1)} {})(PARA >w2)"},
	})
}

func xTestEnumAfterPara(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"abc\n* def", "(PARA abc)(UL {(PARA def)})"},
		{"abc\n*def", "(PARA abc SB *def)"},
	})
}

func xTestDefinition(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{";", "(PARA ;)"},
		{"; ", "(PARA ;)"},
		{"; abc", "(DL (DT abc))"},
		{"; abc\ndef", "(DL (DT abc))(PARA def)"},
		{"; abc\n def", "(DL (DT abc))(PARA def)"},
		{"; abc\n  def", "(DL (DT abc SB def))"},
		{":", "(PARA :)"},
		{": ", "(PARA :)"},
		{": abc", "(PARA : SP abc)"},
		{"; abc\n: def", "(DL (DT abc) (DD (PARA def)))"},
		{"; abc\n: def\nghi", "(DL (DT abc) (DD (PARA def)))(PARA ghi)"},
		{"; abc\n: def\n ghi", "(DL (DT abc) (DD (PARA def)))(PARA ghi)"},
		{"; abc\n: def\n  ghi", "(DL (DT abc) (DD (PARA def SB ghi)))"},
		{"; abc\n: def\n\n  ghi", "(DL (DT abc) (DD (PARA def)(PARA ghi)))"},
		{"; abc\n:", "(DL (DT abc))(PARA :)"},
		{"; abc\n: def\n: ghi", "(DL (DT abc) (DD (PARA def)) (DD (PARA ghi)))"},
		{"; abc\n: def\n; ghi\n: jkl", "(DL (DT abc) (DD (PARA def)) (DT ghi) (DD (PARA jkl)))"},
	})
}

func TestTable(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
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
	checkTcs(t, TestCases{
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
		{"::a::{py=\"2 3}", "(BLOCK (PARA (FORMAT-SPAN () (TEXT \"a\")) (TEXT \"{py=\\\"2\") (SPACE) (TEXT \"3}\")))"},
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
