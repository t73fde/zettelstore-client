//-----------------------------------------------------------------------------
// Copyright (c) 2022-2023 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package html

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/api"
	"zettelstore.de/c/attrs"
	"zettelstore.de/c/sexpr"
	"zettelstore.de/c/text"
)

// EncEnvironment represent the encoding environment.
// It is itself a sxpf.Environment.
//
// Builtins is public, so that HTML encoders based on this one can modify some
// functionality. Builtins should not be updated, but can be used as a parent
// map when creating a new one.
type EncEnvironment struct {
	err            error
	Builtins       *sxpf.SymbolMap
	w              io.Writer
	headingOffset  int
	unique         string
	footnotes      []sfootnodeInfo
	writeFootnotes bool // true iff output should include footnotes and marks
	noLinks        bool // true iff output must not include links
	visibleSpace   bool // true iff space should be "visible" by using EscapeVisible
}
type sfootnodeInfo struct {
	note  *sxpf.Pair
	attrs attrs.Attributes
}

func NewEncEnvironment(w io.Writer, headingOffset int) *EncEnvironment {
	return &EncEnvironment{
		Builtins:       buildBuiltins(),
		w:              w,
		headingOffset:  headingOffset,
		footnotes:      nil,
		writeFootnotes: true,
	}
}

func buildBuiltins() *sxpf.SymbolMap {
	builtins := sxpf.NewSymbolMap(sexpr.Smk, nil)
	for _, b := range defaultEncodingFunctions {
		name := b.sym.GetValue()
		primFunc := b.fn
		builtins.Set(b.sym, sxpf.NewBuiltin(
			name,
			true, b.minArgs, -1,
			func(env sxpf.Environment, args *sxpf.Pair, _ int) (sxpf.Value, error) {
				primFunc(env.(*EncEnvironment), args)
				return nil, nil
			},
		))
	}
	return builtins
}

// SetError marks the environment with an error, if there is not already a marked error.
// The effect is that all future output is not written.
func (env *EncEnvironment) SetError(err error) {
	if env.err == nil {
		env.err = err
	}
}

// GetError returns the first encountered error during encoding.
func (env *EncEnvironment) GetError() error { return env.err }

// ReplaceWriter flushes the previous writer and installs the new one.
func (env *EncEnvironment) ReplaceWriter(w io.Writer) { env.w = w }

// SetUnique sets a string that maked footnote, heading, and mark fragments unique.
func (env *EncEnvironment) SetUnique(s string) {
	if s == "" {
		env.unique = ""
	} else {
		env.unique = ":" + s
	}
}

// IgnoreLinks returns true, if HTML links must not be encoded. This happens if
// the encoded HTML is used in a link itself.
func (env *EncEnvironment) IgnoreLinks() bool { return env.noLinks }

// WriteString encodes a string literally.
func (env *EncEnvironment) Write(b []byte) (l int, err error) {
	if env.err == nil {
		l, env.err = env.w.Write(b)
	}
	return l, env.err
}

// WriteString encodes a string literally.
func (env *EncEnvironment) WriteString(s string) (l int, err error) {
	if env.err == nil {
		l, env.err = io.WriteString(env.w, s)
	}
	return l, env.err
}

// WriteStrings encodes many string literally.
func (env *EncEnvironment) WriteStrings(sl ...string) {
	if env.err == nil {
		for _, s := range sl {
			_, env.err = io.WriteString(env.w, s)
			if env.err != nil {
				return
			}
		}
	}
}

// WriteEscape encodes a string so that it cannot interfere with other HTML code.
func (env *EncEnvironment) WriteEscaped(s string) {
	if env.err == nil {
		_, env.err = Escape(env.w, s)
	}
}

func (env *EncEnvironment) WriteEscapedOrVisible(s string) {
	if env.err == nil {
		if env.visibleSpace {
			_, env.err = EscapeVisible(env.w, s)
		} else {
			_, env.err = Escape(env.w, s)
		}
	}
}

func (env *EncEnvironment) GetSymbol(p *sxpf.Pair) (res *sxpf.Symbol) {
	if env.err != nil {
		return nil
	}
	res, env.err = p.GetSymbol()
	return res
}

func (env *EncEnvironment) GetString(p *sxpf.Pair) (res string) {
	if env.err != nil {
		return ""
	}
	res, env.err = p.GetString()
	return res
}

func (env *EncEnvironment) GetInteger(p *sxpf.Pair) (res int64) {
	if env.err != nil {
		return 0
	}
	res, env.err = p.GetInteger()
	return res
}

func (env *EncEnvironment) GetPair(p *sxpf.Pair) (res *sxpf.Pair) {
	if env.err != nil {
		return nil
	}
	res, env.err = p.GetPair()
	return res
}

func (env *EncEnvironment) GetAttributes(p *sxpf.Pair) attrs.Attributes {
	if env.err != nil {
		return nil
	}
	return sexpr.GetAttributes(env.GetPair(p))
}

func (env *EncEnvironment) WriteAttributes(a attrs.Attributes) {
	if len(a) == 0 {
		return
	}
	for _, key := range a.Keys() {
		if key == "" || key == attrs.DefaultAttribute {
			continue
		}
		val, found := a.Get(key)
		if !found {
			continue
		}
		env.WriteString(" ")
		env.WriteOneAttribute(key, val)
	}
}

func (env *EncEnvironment) WriteOneAttribute(key, val string) {
	if env.err == nil {
		env.WriteString(key)
		if val != "" {
			env.WriteString(`="`)
			_, env.err = AttributeEscape(env.w, val)
			env.WriteString(`"`)
		}
	}
}

func (env *EncEnvironment) WriteStartTag(tag string, a attrs.Attributes) {
	env.WriteStrings("<", tag)
	env.WriteAttributes(a)
	env.WriteString(">")
}

func (env *EncEnvironment) WriteEndTag(tag string) {
	env.WriteStrings("</", tag, ">")
}

func (env *EncEnvironment) WriteImage(args *sxpf.Pair) {
	ref := env.GetPair(args.GetTail())
	env.WriteImageWithSource(args, env.GetString(ref.GetTail()))
}

func (env *EncEnvironment) WriteImageWithSource(args *sxpf.Pair, src string) {
	a := env.GetAttributes(args)
	a = a.Set("src", src)
	if description := args.GetTail().GetTail().GetTail(); !description.IsNil() {
		a = a.Set("alt", text.EvaluateInlineString(description))
	} else {
		a = a.Set("alt", "alternate description missing")
	}
	env.WriteStartTag("img", a)
}

func (env *EncEnvironment) LookupForm(sym *sxpf.Symbol) (sxpf.Form, error) {
	return env.Builtins.LookupForm(sym)
}

func (env *EncEnvironment) EvalOther(val sxpf.Value) (sxpf.Value, error) {
	if strVal, ok := val.(*sxpf.String); ok {
		env.WriteEscaped(strVal.GetValue())
		return nil, nil
	}
	return val, nil
}

func (env *EncEnvironment) EvalSymbol(val *sxpf.Symbol) (sxpf.Value, error) {
	env.WriteEscaped(val.GetValue())
	return nil, nil
}

func (env *EncEnvironment) EvalPair(p *sxpf.Pair) (sxpf.Value, error) {
	return sxpf.EvalCallOrSeq(env, p)
}

func EvaluateInline(baseEnv *EncEnvironment, value sxpf.Value, withFootnotes, noLinks bool) string {
	var buf bytes.Buffer
	env := EncEnvironment{w: &buf, noLinks: noLinks}
	if baseEnv != nil {
		env.Builtins = baseEnv.Builtins
		env.writeFootnotes = withFootnotes && baseEnv.writeFootnotes
		env.footnotes = baseEnv.footnotes
	} else {
		env.Builtins = buildBuiltins()
	}
	_, err := sxpf.Eval(&env, value)
	if err != nil {
		return err.Error()
	}
	if baseEnv != nil {
		baseEnv.footnotes = env.footnotes
	}
	return buf.String()
}

func (env *EncEnvironment) WriteEndnotes() {
	if len(env.footnotes) == 0 {
		return
	}
	env.WriteString("<ol class=\"zs-endnotes\">")
	for i := 0; i < len(env.footnotes); i++ {
		fni := env.footnotes[i]
		n := strconv.Itoa(i + 1)
		un := env.unique + n
		a := fni.attrs.Clone().AddClass("zs-endnote").Set("value", n)
		if _, found := a.Get("id"); !found {
			a = a.Set("id", "fn:"+un)
		}
		if _, found := a.Get("role"); !found {
			a = a.Set("role", "doc-endnote")
		}
		env.WriteStartTag("li", a)
		sxpf.EvalSequence(env, fni.note) // may add more footnotes
		env.WriteStrings(
			` <a class="zs-endnote-backref" href="#fnref:`,
			un,
			"\" role=\"doc-backlink\">&#x21a9;&#xfe0e;</a></li>")
	}
	env.footnotes = nil
	env.WriteString("</ol>")
}

type encodingFunc func(env *EncEnvironment, args *sxpf.Pair)

var defaultEncodingFunctions = []struct {
	sym     *sxpf.Symbol
	minArgs int
	fn      encodingFunc
}{
	{sexpr.SymPara, 0, func(env *EncEnvironment, args *sxpf.Pair) {
		if !env.isCommentList(args) {
			env.WriteString("<p>")
			sxpf.EvalSequence(env, args)
			env.WriteString("</p>")
		}
	}},
	{sexpr.SymHeading, 5, func(env *EncEnvironment, args *sxpf.Pair) {
		nLevel := env.GetInteger(args)
		if nLevel <= 0 {
			return
		}
		level := strconv.FormatInt(nLevel+int64(env.headingOffset), 10)

		argAttr := args.GetTail()
		a := env.GetAttributes(argAttr)
		argFragment := argAttr.GetTail().GetTail()
		if fragment := env.GetString(argFragment); fragment != "" {
			a = a.Set("id", fragment)
		}

		env.WriteStrings("<h", level)
		env.WriteAttributes(a)
		env.WriteString(">")
		sxpf.EvalSequence(env, argFragment.GetTail())
		env.WriteStrings("</h", level, ">")
	}},
	{sexpr.SymThematic, 0, func(env *EncEnvironment, args *sxpf.Pair) {
		env.WriteString("<hr")
		if !sxpf.IsNil(args) {
			env.WriteAttributes(env.GetAttributes(args))
		}
		env.WriteString(">")
	}},
	{sexpr.SymListUnordered, 0, makeListFn("ul")},
	{sexpr.SymListOrdered, 0, makeListFn("ol")},
	{sexpr.SymListQuote, 0, func(env *EncEnvironment, args *sxpf.Pair) {
		env.WriteString("<blockquote>")
		if !args.IsNil() && args.GetFirst().IsNil() {
			sxpf.Eval(env, env.GetPair(args))
		} else {
			for elem := args; !elem.IsNil(); elem = elem.GetTail() {
				env.WriteString("<p>")
				sxpf.Eval(env, env.GetPair(elem))
				env.WriteString("</p>")
			}
		}
		env.WriteString("</blockquote>")
	}},
	{sexpr.SymDescription, 0, func(env *EncEnvironment, args *sxpf.Pair) {
		env.WriteString("<dl>")
		for elem := args; !elem.IsNil(); elem = elem.GetTail() {
			env.WriteString("<dt>")
			sxpf.Eval(env, elem.GetFirst())
			env.WriteString("</dt>")
			elem = elem.GetTail()
			if elem.IsNil() {
				break
			}
			ddlist, err := elem.GetPair()
			if err != nil {
				continue
			}
			for dditem := ddlist; !dditem.IsNil(); dditem = dditem.GetTail() {
				env.WriteString("<dd>")
				sxpf.Eval(env, dditem.GetFirst())
				env.WriteString("</dd>")
			}
		}
		env.WriteString("</dl>")
	}},
	{sexpr.SymTable, 1, func(env *EncEnvironment, args *sxpf.Pair) {
		env.WriteString("<table>")
		if header := env.GetPair(args); !header.IsNil() {
			env.WriteString("<thead>")
			env.writeTableRow(header)
			env.WriteString("</thead>")
		}
		if argBody := args.GetTail(); !argBody.IsNil() {
			env.WriteString("<tbody>")
			for row := argBody; !row.IsNil(); row = row.GetTail() {
				env.writeTableRow(env.GetPair(row))
			}
			env.WriteString("</tbody>")
		}
		env.WriteString("</table>")
	}},
	{sexpr.SymCell, 0, makeCellFn("")},
	{sexpr.SymCellCenter, 0, makeCellFn("center")},
	{sexpr.SymCellLeft, 0, makeCellFn("left")},
	{sexpr.SymCellRight, 0, makeCellFn("right")},
	{sexpr.SymRegionBlock, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := env.GetAttributes(args)
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		env.writeRegion(args, a, "div")
	}},
	{sexpr.SymRegionQuote, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		env.writeRegion(args, nil, "blockquote")
	}},
	{sexpr.SymRegionVerse, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		env.writeRegion(args, nil, "div")
	}},
	{sexpr.SymVerbatimComment, 1, func(env *EncEnvironment, args *sxpf.Pair) {
		if env.GetAttributes(args).HasDefault() {
			if s := env.GetString(args.GetTail()); s != "" {
				env.WriteString("<!--\n")
				env.WriteEscaped(s)
				env.WriteString("\n-->")
			}
		}
	}},
	{sexpr.SymVerbatimEval, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := env.GetAttributes(args).AddClass("zs-eval")
		env.writeVerbatim(args, a)
	}},
	{sexpr.SymVerbatimHTML, 2, execHTML},
	{sexpr.SymVerbatimMath, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := env.GetAttributes(args).AddClass("zs-math")
		env.writeVerbatim(args, a)
	}},
	{sexpr.SymVerbatimProg, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := setProgLang(env.GetAttributes(args))
		oldVisible := env.visibleSpace
		if a.HasDefault() {
			a = a.RemoveDefault()
			env.visibleSpace = true
		}
		env.writeVerbatim(args, a)
		env.visibleSpace = oldVisible
	}},
	{sexpr.SymVerbatimZettel, 0, DoNothingFn},
	{sexpr.SymBLOB, 3, func(env *EncEnvironment, args *sxpf.Pair) {
		argSyntax := args.GetTail()
		env.writeBLOB(env.GetPair(args), env.GetString(argSyntax), env.GetString(argSyntax.GetTail()))
	}},
	{sexpr.SymTransclude, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := sexpr.GetAttributes(env.GetPair(args))
		ref := env.GetPair(args.GetTail())
		refKind := env.GetSymbol(ref)
		if refKind == nil {
			return
		}
		if refValue := env.GetString(ref.GetTail()); refValue != "" {
			if sexpr.SymRefStateExternal.Equal(refKind) {
				a = a.Set("src", refValue).AddClass("external")
				env.WriteString("<p><img")
				env.WriteAttributes(a)
				env.WriteString("></p>")
				return
			}
			env.WriteStrings("<!-- transclude ", refKind.GetValue(), ": ")
			env.WriteEscaped(refValue)
			env.WriteString(" -->")
			return
		}
		if env.err == nil {
			_, env.err = fmt.Fprintf(env.w, "%v\n", args)
		}
	}},
	{sexpr.SymText, 0, func(env *EncEnvironment, args *sxpf.Pair) {
		if !sxpf.IsNil(args) {
			env.WriteEscaped(env.GetString(args))
		}
	}},
	{sexpr.SymSpace, 0, func(env *EncEnvironment, args *sxpf.Pair) {
		if sxpf.IsNil(args) {
			env.WriteString(" ")
			return
		}
		env.WriteEscaped(env.GetString(args))
	}},
	{sexpr.SymSoft, 0, func(env *EncEnvironment, _ *sxpf.Pair) { env.WriteString(" ") }},
	{sexpr.SymHard, 0, func(env *EncEnvironment, _ *sxpf.Pair) { env.WriteString("<br>") }},
	{sexpr.SymLinkInvalid, 2, func(env *EncEnvironment, args *sxpf.Pair) { WriteAsSpan(env, args) }},
	{sexpr.SymLinkZettel, 2, func(env *EncEnvironment, args *sxpf.Pair) { WriteHRefLink(env, args) }},
	{sexpr.SymLinkSelf, 2, func(env *EncEnvironment, args *sxpf.Pair) { WriteHRefLink(env, args) }},
	{sexpr.SymLinkFound, 2, func(env *EncEnvironment, args *sxpf.Pair) { WriteHRefLink(env, args) }},
	{sexpr.SymLinkBroken, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		if a, refValue, ok := PrepareLink(env, args); ok {
			WriteLink(env, args, a.AddClass("broken"), refValue, "")
		}
	}},
	{sexpr.SymLinkHosted, 2, func(env *EncEnvironment, args *sxpf.Pair) { WriteHRefLink(env, args) }},
	{sexpr.SymLinkBased, 2, func(env *EncEnvironment, args *sxpf.Pair) { WriteHRefLink(env, args) }},
	{sexpr.SymLinkQuery, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		if a, refValue, ok := PrepareLink(env, args); ok {
			query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue)
			WriteLink(env, args, a.Set("href", query), refValue, "")
		}
	}},
	{sexpr.SymLinkExternal, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		if a, refValue, ok := PrepareLink(env, args); ok {
			WriteLink(env, args, a.Set("href", refValue).AddClass("external"), refValue, "")
		}
	}},
	{sexpr.SymEmbed, 3, func(env *EncEnvironment, args *sxpf.Pair) {
		argRef := args.GetTail()
		if syntax := env.GetString(argRef.GetTail()); syntax == api.ValueSyntaxSVG {
			ref := env.GetPair(argRef)
			env.WriteStrings(
				`<figure><embed type="image/svg+xml" src="`, "/", env.GetString(ref.GetTail()), ".svg", "\" /></figure>")
		} else {
			env.WriteImage(args)
		}
	}},
	{sexpr.SymEmbedBLOB, 3, func(env *EncEnvironment, args *sxpf.Pair) {
		argSyntax := args.GetTail()
		a, syntax, data := env.GetAttributes(args), env.GetString(argSyntax), env.GetString(argSyntax.GetTail())
		summary, _ := a.Get(api.KeySummary)
		env.writeBLOB(sxpf.NewPair(sxpf.NewString(summary), sxpf.Nil()), syntax, data)
	}},
	{sexpr.SymCite, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		env.WriteStartTag("span", env.GetAttributes(args))
		argKey := args.GetTail()
		if key := env.GetString(argKey); key != "" {
			env.WriteEscaped(key)
			if text := argKey.GetTail(); !text.IsNil() {
				env.WriteString(", ")
				sxpf.EvalSequence(env, text)
			}
		}
		env.WriteString("</span>")
	}},
	{sexpr.SymMark, 3, func(env *EncEnvironment, args *sxpf.Pair) {
		if env.noLinks {
			sxpf.Eval(env, sxpf.NewPair(sexpr.SymFormatSpan, args))
			return
		}
		argFragment := args.GetTail().GetTail()
		if fragment := env.GetString(argFragment); fragment != "" {
			env.WriteString(`<a id="`)
			env.WriteString(env.unique)
			env.WriteString(fragment)
			env.WriteString(`">`)
			sxpf.EvalSequence(env, argFragment.GetTail())
			env.WriteString("</a>")
		} else {
			sxpf.EvalSequence(env, argFragment.GetTail())
		}
	}},
	{sexpr.SymFootnote, 1, func(env *EncEnvironment, args *sxpf.Pair) {
		if env.writeFootnotes {
			a := env.GetAttributes(args)
			env.footnotes = append(env.footnotes, sfootnodeInfo{args.GetTail(), a})
			n := strconv.Itoa(len(env.footnotes))
			un := env.unique + n
			env.WriteStrings(
				`<sup id="fnref:`, un, `"><a class="zs-noteref" href="#fn:`, un,
				`" role="doc-noteref">`, n, `</a></sup>`)
		}
	}},
	{sexpr.SymFormatDelete, 1, makeFormatFn("del")},
	{sexpr.SymFormatEmph, 1, makeFormatFn("em")},
	{sexpr.SymFormatInsert, 1, makeFormatFn("ins")},
	{sexpr.SymFormatQuote, 1, writeQuote},
	{sexpr.SymFormatSpan, 1, makeFormatFn("span")},
	{sexpr.SymFormatStrong, 1, makeFormatFn("strong")},
	{sexpr.SymFormatSub, 1, makeFormatFn("sub")},
	{sexpr.SymFormatSuper, 1, makeFormatFn("sup")},
	{sexpr.SymLiteralComment, 1, func(env *EncEnvironment, args *sxpf.Pair) {
		if env.GetAttributes(args).HasDefault() {
			if s := env.GetString(args.GetTail()); s != "" {
				env.WriteString("<!-- ")
				env.WriteEscaped(s)
				env.WriteString(" -->")
			}
		}
	}},
	{sexpr.SymLiteralHTML, 2, execHTML},
	{sexpr.SymLiteralInput, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		env.writeLiteral(args, nil, "kbd")
	}},
	{sexpr.SymLiteralMath, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := env.GetAttributes(args).AddClass("zs-math")
		env.writeLiteral(args, a, "code")
	}},
	{sexpr.SymLiteralOutput, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		env.writeLiteral(args, nil, "samp")
	}},
	{sexpr.SymLiteralProg, 2, func(env *EncEnvironment, args *sxpf.Pair) {
		a := setProgLang(env.GetAttributes(args))
		env.writeLiteral(args, a, "code")
	}},
	{sexpr.SymLiteralZettel, 0, DoNothingFn},
}

// DoNothingFn is a function that does nothing.
func DoNothingFn(*EncEnvironment, *sxpf.Pair) { /* Should really do nothing */ }

func (env *EncEnvironment) isCommentList(seq *sxpf.Pair) bool {
	if seq.IsEmpty() {
		return false
	}
	elem := seq
	for {
		item, err := elem.GetPair()
		if err != nil {
			return false
		}
		if sym := item.GetFirst(); sym == sexpr.SymLiteralComment {
			args := item.GetTail()
			if args == nil {
				return true
			}
			attr := env.GetAttributes(args)
			if _, found := attr[attrs.DefaultAttribute]; found {
				return false
			}
		} else if sym != sexpr.SymSoft {
			return false
		}

		nVal := elem.GetSecond()
		if sxpf.IsNil(nVal) {
			return true
		}
		next, ok := nVal.(*sxpf.Pair)
		if !ok {
			return false
		}
		elem = next
	}
}

func makeListFn(tag string) encodingFunc {
	return func(env *EncEnvironment, args *sxpf.Pair) {
		env.WriteStartTag(tag, nil)
		for elem := args; !elem.IsNil(); elem = elem.GetTail() {
			env.WriteStartTag("li", nil)
			sxpf.Eval(env, elem.GetFirst())
			env.WriteEndTag("li")
		}
		env.WriteEndTag(tag)
	}
}

func (env *EncEnvironment) writeTableRow(cells *sxpf.Pair) {
	if !cells.IsNil() {
		env.WriteString("<tr>")
		for cell := cells; !cell.IsNil(); cell = cell.GetTail() {
			sxpf.Eval(env, cell.GetFirst())
		}
		env.WriteString("</tr>")
	}
}
func makeCellFn(align string) encodingFunc {
	return func(env *EncEnvironment, args *sxpf.Pair) {
		if align == "" {
			env.WriteString("<td>")
		} else {
			env.WriteStrings(`<td class="`, align, `">`)
		}
		sxpf.EvalSequence(env, args)
		env.WriteString("</td>")
	}
}

func (env *EncEnvironment) writeRegion(args *sxpf.Pair, a attrs.Attributes, tag string) {
	if a == nil {
		a = env.GetAttributes(args)
	}
	env.WriteStartTag(tag, a)
	sxpf.Eval(env, env.GetPair(args.GetTail()))
	if cite := env.GetPair(args.GetTail().GetTail()); !cite.IsNil() {
		env.WriteString("<cite>")
		sxpf.EvalSequence(env, cite)
		env.WriteString("</cite>")
	}
	env.WriteEndTag(tag)
}

func (env *EncEnvironment) writeVerbatim(args *sxpf.Pair, a attrs.Attributes) {
	env.WriteString("<pre>")
	env.WriteStartTag("code", a)
	env.WriteEscapedOrVisible(env.GetString(args.GetTail()))
	env.WriteString("</code></pre>")
}

func execHTML(env *EncEnvironment, args *sxpf.Pair) {
	if s := env.GetString(args.GetTail()); s != "" && IsSafe(s) {
		env.WriteString(s)
	}
}

func (env *EncEnvironment) writeBLOB(description *sxpf.Pair, syntax, data string) {
	if data == "" {
		return
	}
	switch syntax {
	case "":
	case api.ValueSyntaxSVG:
		// TODO: add description
		env.WriteStrings("<p>", data, "</p>")
	default:
		env.WriteStrings(`<p><img src="data:image/`, syntax, ";base64,", data, `" `)
		if description.IsEmpty() {
			env.WriteOneAttribute("alt", "alternate description missing")
		} else {
			env.WriteOneAttribute("alt", text.EvaluateInlineString(description))
		}
		env.WriteString(`></p>`)
	}
}

func PrepareLink(env *EncEnvironment, args *sxpf.Pair) (attrs.Attributes, string, bool) {
	if env.noLinks {
		WriteAsSpan(env, args)
		return nil, "", false
	}
	return env.GetAttributes(args), env.GetString(args.GetTail()), true
}

func WriteAsSpan(env *EncEnvironment, args *sxpf.Pair) {
	if args.Length() > 2 {
		sxpf.Eval(env, sxpf.NewPair(sexpr.SymFormatSpan, sxpf.NewPair(args.GetFirst(), args.GetTail().GetTail())))
	}
}

func WriteLink(env *EncEnvironment, args *sxpf.Pair, a attrs.Attributes, refValue, suffix string) {
	env.WriteString("<a")
	env.WriteAttributes(a)
	env.WriteString(">")

	if args.Length() > 2 {
		sxpf.EvalSequence(env, args.GetTail().GetTail())
	} else {
		env.WriteString(refValue)
	}
	env.WriteStrings("</a>", suffix)
}

func WriteHRefLink(env *EncEnvironment, args *sxpf.Pair) {
	if a, refValue, ok := PrepareLink(env, args); ok {
		WriteLink(env, args, a.Set("href", refValue), refValue, "")
	}
}

func makeFormatFn(tag string) encodingFunc {
	return func(env *EncEnvironment, args *sxpf.Pair) {
		a := env.GetAttributes(args)
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		env.WriteStartTag(tag, a)
		sxpf.EvalSequence(env, args.GetTail())
		env.WriteEndTag(tag)
	}
}

func writeQuote(env *EncEnvironment, args *sxpf.Pair) {
	const langAttr = "lang"
	a := env.GetAttributes(args)
	lang, hasLang := a.Get(langAttr)
	if hasLang {
		a = a.Remove(langAttr)
		env.WriteStartTag("span", attrs.Attributes{}.Set(langAttr, lang))
	}
	env.WriteStartTag("q", a)
	sxpf.EvalSequence(env, args.GetTail())
	env.WriteEndTag("q")
	if hasLang {
		env.WriteEndTag("span")
	}
}

func (env *EncEnvironment) writeLiteral(args *sxpf.Pair, a attrs.Attributes, tag string) {
	if a == nil {
		a = env.GetAttributes(args)
	}
	oldVisible := env.visibleSpace
	if a.HasDefault() {
		env.visibleSpace = true
		a = a.RemoveDefault()
	}
	env.WriteStartTag(tag, a)
	env.WriteEscapedOrVisible(env.GetString(args.GetTail()))
	env.visibleSpace = oldVisible
	env.WriteEndTag(tag)
}

func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}
