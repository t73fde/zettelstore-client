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

func xTestLink(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[", "(PARA [)"},
		{"[[", "(PARA [[)"},
		{"[[|", "(PARA [[|)"},
		{"[[]", "(PARA [[])"},
		{"[[|]", "(PARA [[|])"},
		{"[[]]", "(PARA [[]])"},
		{"[[|]]", "(PARA [[|]])"},
		{"[[ ]]", "(PARA [[ SP ]])"},
		{"[[\n]]", "(PARA [[ SB ]])"},
		{"[[ a]]", "(PARA (LINK a))"},
		{"[[a ]]", "(PARA [[a SP ]])"},
		{"[[a\n]]", "(PARA [[a SB ]])"},
		{"[[a]]", "(PARA (LINK a))"},
		{"[[12345678901234]]", "(PARA (LINK 12345678901234))"},
		{"[[a]", "(PARA [[a])"},
		{"[[|a]]", "(PARA [[|a]])"},
		{"[[b|]]", "(PARA [[b|]])"},
		{"[[b|a]]", "(PARA (LINK a b))"},
		{"[[b| a]]", "(PARA (LINK a b))"},
		{"[[b%c|a]]", "(PARA (LINK a b%c))"},
		{"[[b%%c|a]]", "(PARA [[b {% c|a]]})"},
		{"[[b|a]", "(PARA [[b|a])"},
		{"[[b\nc|a]]", "(PARA (LINK a b SB c))"},
		{"[[b c|a#n]]", "(PARA (LINK a#n b SP c))"},
		{"[[a]]go", "(PARA (LINK a) go)"},
		{"[[b|a]]{go}", "(PARA (LINK a b)[ATTR go])"},
		{"[[[[a]]|b]]", "(PARA [[ (LINK a) |b]])"},
		{"[[a[b]c|d]]", "(PARA (LINK d a[b]c))"},
		{"[[[b]c|d]]", "(PARA [ (LINK d b]c))"},
		{"[[a[]c|d]]", "(PARA (LINK d a[]c))"},
		{"[[a[b]|d]]", "(PARA (LINK d a[b]))"},
		{"[[\\|]]", "(PARA (LINK %5C%7C))"},
		{"[[\\||a]]", "(PARA (LINK a |))"},
		{"[[b\\||a]]", "(PARA (LINK a b|))"},
		{"[[b\\|c|a]]", "(PARA (LINK a b|c))"},
		{"[[\\]]]", "(PARA (LINK %5C%5D))"},
		{"[[\\]|a]]", "(PARA (LINK a ]))"},
		{"[[b\\]|a]]", "(PARA (LINK a b]))"},
		{"[[\\]\\||a]]", "(PARA (LINK a ]|))"},
		{"[[http://a]]", "(PARA (LINK http://a))"},
		{"[[http://a|http://a]]", "(PARA (LINK http://a http://a))"},
		{"[[[[a]]]]", "(PARA [[ (LINK a) ]])"},
		{"[[query:title]]", "(PARA (LINK query:title))"},
		{"[[query:title syntax]]", "(PARA (LINK query:title syntax))"},
		{"[[query:title | action]]", "(PARA (LINK query:title | action))"},
		{"[[Text|query:title]]", "(PARA (LINK query:title Text))"},
		{"[[Text|query:title syntax]]", "(PARA (LINK query:title syntax Text))"},
		{"[[Text|query:title | action]]", "(PARA (LINK query:title | action Text))"},
	})
}

func xTestCite(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[@", "(PARA [@)"},
		{"[@]", "(PARA [@])"},
		{"[@a]", "(PARA (CITE a))"},
		{"[@ a]", "(PARA [@ SP a])"},
		{"[@a ]", "(PARA (CITE a))"},
		{"[@a\n]", "(PARA (CITE a))"},
		{"[@a\nx]", "(PARA (CITE a SB x))"},
		{"[@a\n\n]", "(PARA [@a)(PARA ])"},
		{"[@a,\n]", "(PARA (CITE a))"},
		{"[@a,n]", "(PARA (CITE a n))"},
		{"[@a| n]", "(PARA (CITE a n))"},
		{"[@a|n ]", "(PARA (CITE a n))"},
		{"[@a,[@b]]", "(PARA (CITE a (CITE b)))"},
		{"[@a]{color=green}", "(PARA (CITE a)[ATTR color=green])"},
	})
}

func xTestFootnote(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[^", "(PARA [^)"},
		{"[^]", "(PARA (FN))"},
		{"[^abc]", "(PARA (FN abc))"},
		{"[^abc ]", "(PARA (FN abc))"},
		{"[^abc\ndef]", "(PARA (FN abc SB def))"},
		{"[^abc\n\ndef]", "(PARA [^abc)(PARA def])"},
		{"[^abc[^def]]", "(PARA (FN abc (FN def)))"},
		{"[^abc]{-}", "(PARA (FN abc)[ATTR -])"},
	})
}

func xTestEmbed(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"{", "(PARA {)"},
		{"{{", "(PARA {{)"},
		{"{{|", "(PARA {{|)"},
		{"{{}", "(PARA {{})"},
		{"{{|}", "(PARA {{|})"},
		{"{{}}", "(PARA {{}})"},
		{"{{|}}", "(PARA {{|}})"},
		{"{{ }}", "(PARA {{ SP }})"},
		{"{{\n}}", "(PARA {{ SB }})"},
		{"{{a }}", "(PARA {{a SP }})"},
		{"{{a\n}}", "(PARA {{a SB }})"},
		{"{{a}}", "(PARA (EMBED a))"},
		{"{{12345678901234}}", "(PARA (EMBED 12345678901234))"},
		{"{{ a}}", "(PARA (EMBED a))"},
		{"{{a}", "(PARA {{a})"},
		{"{{|a}}", "(PARA {{|a}})"},
		{"{{b|}}", "(PARA {{b|}})"},
		{"{{b|a}}", "(PARA (EMBED a b))"},
		{"{{b| a}}", "(PARA (EMBED a b))"},
		{"{{b|a}", "(PARA {{b|a})"},
		{"{{b\nc|a}}", "(PARA (EMBED a b SB c))"},
		{"{{b c|a#n}}", "(PARA (EMBED a#n b SP c))"},
		{"{{a}}{go}", "(PARA (EMBED a)[ATTR go])"},
		{"{{{{a}}|b}}", "(PARA {{ (EMBED a) |b}})"},
		{"{{\\|}}", "(PARA (EMBED %5C%7C))"},
		{"{{\\||a}}", "(PARA (EMBED a |))"},
		{"{{b\\||a}}", "(PARA (EMBED a b|))"},
		{"{{b\\|c|a}}", "(PARA (EMBED a b|c))"},
		{"{{\\}}}", "(PARA (EMBED %5C%7D))"},
		{"{{\\}|a}}", "(PARA (EMBED a }))"},
		{"{{b\\}|a}}", "(PARA (EMBED a b}))"},
		{"{{\\}\\||a}}", "(PARA (EMBED a }|))"},
		{"{{http://a}}", "(PARA (EMBED http://a))"},
		{"{{http://a|http://a}}", "(PARA (EMBED http://a http://a))"},
		{"{{{{a}}}}", "(PARA {{ (EMBED a) }})"},
	})
}

func xTestMark(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"[!", "(PARA [!)"},
		{"[!\n", "(PARA [!)"},
		{"[!]", "(PARA (MARK #*))"},
		{"[!][!]", "(PARA (MARK #*) (MARK #*-1))"},
		{"[! ]", "(PARA [! SP ])"},
		{"[!a]", "(PARA (MARK \"a\" #a))"},
		{"[!a][!a]", "(PARA (MARK \"a\" #a) (MARK \"a\" #a-1))"},
		{"[!a ]", "(PARA [!a SP ])"},
		{"[!a_]", "(PARA (MARK \"a_\" #a))"},
		{"[!a_][!a]", "(PARA (MARK \"a_\" #a) (MARK \"a\" #a-1))"},
		{"[!a-b]", "(PARA (MARK \"a-b\" #a-b))"},
		{"[!a|b]", "(PARA (MARK \"a\" #a b))"},
		{"[!a|]", "(PARA (MARK \"a\" #a))"},
		{"[!|b]", "(PARA (MARK #* b))"},
		{"[!|b c]", "(PARA (MARK #* b SP c))"},
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
		{"%%{=}a", "(BLOCK (PARA (LITERAL-COMMENT (@ ((\"\" . \"\"))) \"a\")))"},
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
			{"$$a$${go}", "(BLOCK (PARA ($% (@ ((\"go\" . \"\"))) (TEXT \"a\"))))"},
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
		{"$$a$${go}", "(BLOCK (PARA ($% (@ ((\"go\" . \"\"))) (TEXT \"a\"))))"},
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
			{"$$a$${go}", "(BLOCK (PARA ($% (@ ((\"go\" . \"\"))) \"a\")))"},
		}))
	}
	checkTcs(t, TestCases{
		{"''````''", "(BLOCK (PARA (LITERAL-INPUT () \"````\")))"},
		{"''``a``''", "(BLOCK (PARA (LITERAL-INPUT () \"``a``\")))"},
		{"''``''``", "(BLOCK (PARA (LITERAL-INPUT () \"``\") (TEXT \"``\")))"},
		{"''\\'''", "(BLOCK (PARA (LITERAL-INPUT () \"'\")))"},
	})
}

func xTestLiteralMath(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"$", "(PARA $)"},
		{"$$", "(PARA $$)"},
		{"$$$", "(PARA $$$)"},
		{"$$$$", "(PARA {$})"},
		{"$$a$$", "(PARA {$ a})"},
		{"$$a$$$", "(PARA {$ a} $)"},
		{"$$$a$$", "(PARA {$ $a})"},
		{"$$$a$$$", "(PARA {$ $a} $)"},
		{`$\$`, "(PARA $$)"},
		{`$\$$`, "(PARA $$$)"},
		{`$$\$`, "(PARA $$$)"},
		{`$$a\$$`, `(PARA {$ a\})`},
		{`$$a$\$`, "(PARA $$a$$)"},
		{`$$a\$$$`, `(PARA {$ a\} $)`},
		{"$$a$${go}", "(PARA {$ a}[ATTR go])"},
	})
}

func xTestMixFormatCode(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"__abc__\n**def**", "(PARA {_ abc} SB {* def})"},
		{"''abc''\n==def==", "(PARA {' abc} SB {= def})"},
		{"__abc__\n==def==", "(PARA {_ abc} SB {= def})"},
		{"__abc__\n``def``", "(PARA {_ abc} SB {` def})"},
		{"\"\"ghi\"\"\n::abc::\n``def``\n", "(PARA {\" ghi} SB {: abc} SB {` def})"},
	})
}

func xTestNDash(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"--", "(PARA \u2013)"},
		{"a--b", "(PARA a\u2013b)"},
	})
}

func xTestEntity(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"&", "(PARA &)"},
		{"&;", "(PARA &;)"},
		{"&#;", "(PARA &#;)"},
		{"&#1a;", "(PARA &#1a;)"},
		{"&#x;", "(PARA &#x;)"},
		{"&#x0z;", "(PARA &#x0z;)"},
		{"&1;", "(PARA &1;)"},
		{"&#9;", "(PARA &#9;)"}, // No numeric entities below space are not allowed.
		{"&#x1f;", "(PARA &#x1f;)"},

		// Good cases
		{"&lt;", "(PARA <)"},
		{"&#48;", "(PARA 0)"},
		{"&#x4A;", "(PARA J)"},
		{"&#X4a;", "(PARA J)"},
		{"&hellip;", "(PARA \u2026)"},
		{"&nbsp;", "(PARA \u00a0)"},
		{"E: &amp;,&#63;;&#x63;.", "(PARA E: SP &,?;c.)"},
	})
}

func xTestVerbatimZettel(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"@@@\n@@@", "(ZETTEL)"},
		{"@@@\nabc\n@@@", "(ZETTEL\nabc)"},
		{"@@@@def\nabc\n@@@@", "(ZETTEL\nabc)[ATTR =def]"},
	})
}

func xTestVerbatimCode(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"```\n```", "(PROG)"},
		{"```\nabc\n```", "(PROG\nabc)"},
		{"```\nabc\n````", "(PROG\nabc)"},
		{"````\nabc\n````", "(PROG\nabc)"},
		{"````\nabc\n```\n````", "(PROG\nabc\n```)"},
		{"````go\nabc\n````", "(PROG\nabc)[ATTR =go]"},
	})
}

func xTestVerbatimEval(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"~~~\n~~~", "(EVAL)"},
		{"~~~\nabc\n~~~", "(EVAL\nabc)"},
		{"~~~\nabc\n~~~~", "(EVAL\nabc)"},
		{"~~~~\nabc\n~~~~", "(EVAL\nabc)"},
		{"~~~~\nabc\n~~~\n~~~~", "(EVAL\nabc\n~~~)"},
		{"~~~~go\nabc\n~~~~", "(EVAL\nabc)[ATTR =go]"},
	})
}

func xTestVerbatimMath(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"$$$\n$$$", "(MATH)"},
		{"$$$\nabc\n$$$", "(MATH\nabc)"},
		{"$$$\nabc\n$$$$", "(MATH\nabc)"},
		{"$$$$\nabc\n$$$$", "(MATH\nabc)"},
		{"$$$$\nabc\n$$$\n$$$$", "(MATH\nabc\n$$$)"},
		{"$$$$go\nabc\n$$$$", "(MATH\nabc)[ATTR =go]"},
	})
}

func xTestVerbatimComment(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"%%%\n%%%", "(COMMENT)"},
		{"%%%\nabc\n%%%", "(COMMENT\nabc)"},
		{"%%%%go\nabc\n%%%%", "(COMMENT\nabc)[ATTR =go]"},
	})
}

func xTestPara(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"a\n\nb", "(PARA a)(PARA b)"},
		{"a\n \nb", "(PARA a)(PARA b)"},
	})
}

func xTestSpanRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{":::\n:::", "(SPAN)"},
		{":::\nabc\n:::", "(SPAN (PARA abc))"},
		{":::\nabc\n::::", "(SPAN (PARA abc))"},
		{"::::\nabc\n::::", "(SPAN (PARA abc))"},
		{"::::\nabc\n:::\ndef\n:::\n::::", "(SPAN (PARA abc)(SPAN (PARA def)))"},
		{":::{go}\n:::", "(SPAN)[ATTR go]"},
		{":::\nabc\n::: def ", "(SPAN (PARA abc) (LINE def))"},
	})
}

func xTestQuoteRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"<<<\n<<<", "(QUOTE)"},
		{"<<<\nabc\n<<<", "(QUOTE (PARA abc))"},
		{"<<<\nabc\n<<<<", "(QUOTE (PARA abc))"},
		{"<<<<\nabc\n<<<<", "(QUOTE (PARA abc))"},
		{"<<<<\nabc\n<<<\ndef\n<<<\n<<<<", "(QUOTE (PARA abc)(QUOTE (PARA def)))"},
		{"<<<go\n<<<", "(QUOTE)[ATTR =go]"},
		{"<<<\nabc\n<<< def ", "(QUOTE (PARA abc) (LINE def))"},
	})
}

func xTestVerseRegion(t *testing.T) {
	t.Parallel()
	checkTcs(t, replace("\"", nil, TestCases{
		{"$$$\n$$$", "(VERSE)"},
		{"$$$\nabc\n$$$", "(VERSE (PARA abc))"},
		{"$$$\nabc\n$$$$", "(VERSE (PARA abc))"},
		{"$$$$\nabc\n$$$$", "(VERSE (PARA abc))"},
		{"$$$\nabc\ndef\n$$$", "(VERSE (PARA abc HB def))"},
		{"$$$$\nabc\n$$$\ndef\n$$$\n$$$$", "(VERSE (PARA abc)(VERSE (PARA def)))"},
		{"$$$go\n$$$", "(VERSE)[ATTR =go]"},
		{"$$$\nabc\n$$$ def ", "(VERSE (PARA abc) (LINE def))"},
	}))
}

func xTestHeading(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"=h", "(PARA =h)"},
		{"= h", "(PARA = SP h)"},
		{"==h", "(PARA ==h)"},
		{"== h", "(PARA == SP h)"},
		{"===h", "(PARA ===h)"},
		{"=== h", "(H1 h #h)"},
		{"===  h", "(H1 h #h)"},
		{"==== h", "(H2 h #h)"},
		{"===== h", "(H3 h #h)"},
		{"====== h", "(H4 h #h)"},
		{"======= h", "(H5 h #h)"},
		{"======== h", "(H5 h #h)"},
		{"=", "(PARA =)"},
		{"=== h=__=a__", "(H1 h= {_ =a} #h-a)"},
		{"=\n", "(PARA =)"},
		{"a=", "(PARA a=)"},
		{" =", "(PARA =)"},
		{"=== h\na", "(H1 h #h)(PARA a)"},
		{"=== h i {-}", "(H1 h SP i #h-i)[ATTR -]"},
		{"=== h {{a}}", "(H1 h SP (EMBED a) #h)"},
		{"=== h{{a}}", "(H1 h (EMBED a) #h)"},
		{"=== {{a}}", "(H1 (EMBED a))"},
		{"=== h {{a}}{-}", "(H1 h SP (EMBED a)[ATTR -] #h)"},
		{"=== h {{a}} {-}", "(H1 h SP (EMBED a) #h)[ATTR -]"},
		{"=== h {-}{{a}}", "(H1 h #h)[ATTR -]"},
		{"=== h{id=abc}", "(H1 h #h)[ATTR id=abc]"},
		{"=== h\n=== h", "(H1 h #h)(H1 h #h-1)"},
	})
}

func xTestHRule(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"-", "(PARA -)"},
		{"---", "(HR)"},
		{"----", "(HR)"},
		{"---A", "(HR)[ATTR =A]"},
		{"---A-", "(HR)[ATTR =A-]"},
		{"-1", "(PARA -1)"},
		{"2-1", "(PARA 2-1)"},
		{"---  {  go  }  ", "(HR)[ATTR go]"},
		{"---  {  .go  }  ", "(HR)[ATTR class=go]"},
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

func xTestTable(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"|", "(TAB (TR))"},
		{"||", "(TAB (TR (TD)))"},
		{"| |", "(TAB (TR (TD)))"},
		{"|a", "(TAB (TR (TD a)))"},
		{"|a|", "(TAB (TR (TD a)))"},
		{"|a| ", "(TAB (TR (TD a)(TD)))"},
		{"|a|b", "(TAB (TR (TD a)(TD b)))"},
		{"|a|b\n|c|d", "(TAB (TR (TD a)(TD b))(TR (TD c)(TD d)))"},
		{"|%", ""},
		{"|a|b\n|%---\n|c|d", "(TAB (TR (TD a)(TD b))(TR (TD c)(TD d)))"},
		{"|a|b\n|c", "(TAB (TR (TD a)(TD b))(TR (TD c)(TD)))"},
	})
}

func xTestTransclude(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"{{{a}}}", "(TRANSCLUDE a)"},
		{"{{{a}}}b", "(TRANSCLUDE a)[ATTR =b]"},
		{"{{{a}}}}", "(TRANSCLUDE a)"},
		{"{{{a\\}}}}", "(TRANSCLUDE a%5C%7D)"},
		{"{{{a\\}}}}b", "(TRANSCLUDE a%5C%7D)[ATTR =b]"},
		{"{{{a}}", "(PARA { (EMBED a))"},
		{"{{{a}}}{go=b}", "(TRANSCLUDE a)[ATTR go=b]"},
	})
}

func xTestBlockAttr(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{":::go\n:::", "(SPAN)[ATTR =go]"},
		{":::go=\n:::", "(SPAN)[ATTR =go]"},
		{":::{}\n:::", "(SPAN)"},
		{":::{ }\n:::", "(SPAN)"},
		{":::{.go}\n:::", "(SPAN)[ATTR class=go]"},
		{":::{=go}\n:::", "(SPAN)[ATTR =go]"},
		{":::{go}\n:::", "(SPAN)[ATTR go]"},
		{":::{go=py}\n:::", "(SPAN)[ATTR go=py]"},
		{":::{.go=py}\n:::", "(SPAN)"},
		{":::{go=}\n:::", "(SPAN)[ATTR go]"},
		{":::{.go=}\n:::", "(SPAN)"},
		{":::{go py}\n:::", "(SPAN)[ATTR go py]"},
		{":::{go\npy}\n:::", "(SPAN)[ATTR go py]"},
		{":::{.go py}\n:::", "(SPAN)[ATTR class=go py]"},
		{":::{go .py}\n:::", "(SPAN)[ATTR class=py go]"},
		{":::{.go py=3}\n:::", "(SPAN)[ATTR class=go py=3]"},
		{":::  {  go  }  \n:::", "(SPAN)[ATTR go]"},
		{":::  {  .go  }  \n:::", "(SPAN)[ATTR class=go]"},
	})
	checkTcs(t, replace("\"", nil, TestCases{
		{":::{py=3}\n:::", "(SPAN)[ATTR py=3]"},
		{":::{py=$2 3$}\n:::", "(SPAN)[ATTR py=$2 3$]"},
		{":::{py=$2\\$3$}\n:::", "(SPAN)[ATTR py=2$3]"},
		{":::{py=2$3}\n:::", "(SPAN)[ATTR py=2$3]"},
		{":::{py=$2\n3$}\n:::", "(SPAN)[ATTR py=$2\n3$]"},
		{":::{py=$2 3}\n:::", "(SPAN)"},
		{":::{py=2 py=3}\n:::", "(SPAN)[ATTR py=$2 3$]"},
		{":::{.go .py}\n:::", "(SPAN)[ATTR class=$go py$]"},
		{":::{go go}\n:::", "(SPAN)[ATTR go]"},
		{":::{=py =go}\n:::", "(SPAN)[ATTR =go]"},
	}))
}

func xTestInlineAttr(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"::a::{}", "(PARA {: a})"},
		{"::a::{ }", "(PARA {: a})"},
		{"::a::{.go}", "(PARA {: a}[ATTR class=go])"},
		{"::a::{=go}", "(PARA {: a}[ATTR =go])"},
		{"::a::{go}", "(PARA {: a}[ATTR go])"},
		{"::a::{go=py}", "(PARA {: a}[ATTR go=py])"},
		{"::a::{.go=py}", "(PARA {: a} {.go=py})"},
		{"::a::{go=}", "(PARA {: a}[ATTR go])"},
		{"::a::{.go=}", "(PARA {: a} {.go=})"},
		{"::a::{go py}", "(PARA {: a}[ATTR go py])"},
		{"::a::{go\npy}", "(PARA {: a}[ATTR go py])"},
		{"::a::{.go py}", "(PARA {: a}[ATTR class=go py])"},
		{"::a::{go .py}", "(PARA {: a}[ATTR class=py go])"},
		{"::a::{  \n go \n .py\n  \n}", "(PARA {: a}[ATTR class=py go])"},
		{"::a::{  \n go \n .py\n\n}", "(PARA {: a}[ATTR class=py go])"},
		{"::a::{\ngo\n}", "(PARA {: a}[ATTR go])"},
	})
	checkTcs(t, replace("\"", nil, TestCases{
		{"::a::{py=3}", "(PARA {: a}[ATTR py=3])"},
		{"::a::{py=$2 3$}", "(PARA {: a}[ATTR py=$2 3$])"},
		{"::a::{py=$2\\$3$}", "(PARA {: a}[ATTR py=2$3])"},
		{"::a::{py=2$3}", "(PARA {: a}[ATTR py=2$3])"},
		{"::a::{py=$2\n3$}", "(PARA {: a}[ATTR py=$2\n3$])"},
		{"::a::{py=$2 3}", "(PARA {: a} {py=$2 SP 3})"},

		{"::a::{py=2 py=3}", "(PARA {: a}[ATTR py=$2 3$])"},
		{"::a::{.go .py}", "(PARA {: a}[ATTR class=$go py$])"},
	}))
}

func TestTemp(t *testing.T) {
	t.Parallel()
	checkTcs(t, TestCases{
		{"", "()"},
	})
}
