//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package html

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/t73fde/sxpf"
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
	err           error
	Builtins      *sxpf.SymbolMap
	w             io.Writer
	headingOffset int
	unique        string
	footnotes     []sfootnodeInfo
	writeFootnote bool // true iff output should include footnotes and marks
	noLinks       bool // true iff output must not include links
	visibleSpace  bool // true iff space should be "visible" by using EscapeVisible
}
type sfootnodeInfo struct {
	note  []sxpf.Value
	attrs attrs.Attributes
}

func NewEncEnvironment(w io.Writer, headingOffset int) *EncEnvironment {
	builtins := sxpf.NewSymbolMap(nil)
	for _, b := range defaultEncodingFunctions {
		name := b.sym.GetValue()
		primFunc := b.fn
		builtins.Set(b.sym, sxpf.NewBuiltin(
			name,
			true, b.minArgs, b.maxArgs,
			func(env sxpf.Environment, args []sxpf.Value) (sxpf.Value, error) {
				primFunc(env.(*EncEnvironment), args)
				return nil, nil
			},
		))
	}
	return &EncEnvironment{
		Builtins:      builtins,
		w:             w,
		headingOffset: headingOffset,
		footnotes:     nil,
		writeFootnote: true,
	}
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
func (env *EncEnvironment) WriteString(s string) {
	if env.err == nil {
		_, env.err = io.WriteString(env.w, s)
	}
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

func (env *EncEnvironment) WriteEscapedLiteral(s string) {
	if env.err == nil {
		if env.visibleSpace {
			_, env.err = EscapeVisible(env.w, s)
		} else {
			_, env.err = EscapeLiteral(env.w, s)
		}
	}
}

func (env *EncEnvironment) MissingArgs(args []sxpf.Value, minArgs int) bool {
	if len(args) < minArgs {
		env.SetError(fmt.Errorf("required args: %d, but got only: %d", minArgs, len(args)))
		return true
	}
	return false
}
func (env *EncEnvironment) GetSymbol(args []sxpf.Value, idx int) (res *sxpf.Symbol) {
	if env.err != nil {
		return nil
	}
	res, env.err = sxpf.GetSymbol(args, idx)
	return res
}
func (env *EncEnvironment) GetString(args []sxpf.Value, idx int) (res string) {
	if env.err != nil {
		return ""
	}
	res, env.err = sxpf.GetString(args, idx)
	return res
}
func (env *EncEnvironment) GetList(args []sxpf.Value, idx int) (res *sxpf.List) {
	if env.err != nil {
		return nil
	}
	res, env.err = sxpf.GetList(args, idx)
	return res
}
func (env *EncEnvironment) GetAttributes(args []sxpf.Value, idx int) attrs.Attributes {
	if env.err != nil {
		return nil
	}
	return sexpr.GetAttributes(env.GetList(args, idx))
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

func (env *EncEnvironment) WriteImage(args []sxpf.Value) {
	a := sexpr.GetAttributes(env.GetList(args, 0))
	ref := env.GetList(args, 1)
	refPair := ref.GetValue()
	a = a.Set("src", env.GetString(refPair, 1))
	if title := args[3:]; len(title) > 0 {
		a = a.Set("title", text.SEncodeInlineString(title))
	}
	env.WriteStartTag("img", a)
}

func (env *EncEnvironment) MakeSymbol(s string) *sxpf.Symbol { return sexpr.Env.MakeSymbol(s) }
func (env *EncEnvironment) LookupForm(sym *sxpf.Symbol) (sxpf.Form, error) {
	return env.Builtins.LookupForm(sym)
}

func (env *EncEnvironment) EvaluateString(val *sxpf.String) (sxpf.Value, error) {
	env.WriteEscaped(val.GetValue())
	return sxpf.Nil(), nil
}

func (env *EncEnvironment) EvaluateSymbol(val *sxpf.Symbol) (sxpf.Value, error) {
	env.WriteEscaped(val.GetValue())
	return sxpf.Nil(), nil
}

func (env *EncEnvironment) EvaluateList(lst *sxpf.List) (sxpf.Value, error) {
	vals := lst.GetValue()
	res, err, done := sxpf.EvaluateCall(env, vals)
	if done {
		return res, err
	}
	result, err := sxpf.EvaluateSlice(env, vals)
	if err != nil {
		return nil, err
	}
	return sxpf.NewList(result...), nil
}

func (env *EncEnvironment) WriteEndnotes() {
	if len(env.footnotes) == 0 {
		return
	}
	env.WriteString("<ol class=\"zs-endnotes\">")
	for i := 0; len(env.footnotes) > 0; i++ {
		fni := env.footnotes[0]
		env.footnotes = env.footnotes[1:]
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
		sxpf.EvaluateSlice(env, fni.note)
		env.WriteStrings(
			` <a class="zs-endnote-backref" href="#fnref:`,
			un,
			"\" role=\"doc-backlink\">&#x21a9;&#xfe0e;</a></li>")
	}
	env.footnotes = nil
	env.WriteString("</ol>")
}

type encodingFunc func(env *EncEnvironment, args []sxpf.Value)

var defaultEncodingFunctions = []struct {
	sym     *sxpf.Symbol
	minArgs int
	maxArgs int
	fn      encodingFunc
}{
	{sexpr.SymPara, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteString("<p>")
		sxpf.EvaluateSlice(env, args)
		env.WriteString("</p>")
	}},
	{sexpr.SymHeading, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if env.MissingArgs(args, 5) {
			return
		}
		nLevel, err := strconv.Atoi(env.GetString(args, 0))
		if err != nil {
			env.SetError(err)
			return
		}
		level := strconv.Itoa(nLevel + env.headingOffset)

		a := env.GetAttributes(args, 1)
		if fragment := env.GetString(args, 3); fragment != "" {
			a = a.Set("id", fragment)
		}

		env.WriteStrings("<h", level)
		env.WriteAttributes(a)
		env.WriteString(">")
		sxpf.EvaluateSlice(env, args[4:])
		env.WriteStrings("</h", level, ">")
	}},
	{sexpr.SymThematic, 0, 1, func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteString("<hr")
		if len(args) > 0 {
			env.WriteAttributes(env.GetAttributes(args, 0))
		}
		env.WriteString(">")
	}},
	{sexpr.SymListUnordered, 0, -1, makeListFn("ul")},
	{sexpr.SymListOrdered, 0, -1, makeListFn("ol")},
	{sexpr.SymListQuote, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteString("<blockquote>")
		if len(args) == 1 {
			sxpf.Evaluate(env, env.GetList(args, 0))
		} else {
			for i := 0; i < len(args); i++ {
				env.WriteString("<p>")
				sxpf.Evaluate(env, env.GetList(args, i))
				env.WriteString("</p>")
			}
		}
		env.WriteString("</blockquote>")
	}},
	{sexpr.SymDescription, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteString("<dl>")
		for i := 0; i < len(args); i += 2 {
			env.WriteString("<dt>")
			sxpf.Evaluate(env, args[i])
			env.WriteString("</dt>")
			i1 := i + 1
			if len(args) <= i1 {
				continue
			}
			ddlist, ok := args[i1].(*sxpf.List)
			if !ok {
				continue
			}
			for _, dditem := range ddlist.GetValue() {
				env.WriteString("<dd>")
				sxpf.Evaluate(env, dditem)
				env.WriteString("</dd>")
			}
		}
		env.WriteString("</dl>")
	}},
	{sexpr.SymTable, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteString("<table>")
		if header := env.GetList(args, 0).GetValue(); len(header) > 0 {
			env.WriteString("<thead>")
			env.writeTableRow(header)
			env.WriteString("</thead>")
		}
		if len(args) > 1 {
			env.WriteString("<tbody>")
			for i := 1; i < len(args); i++ {
				env.writeTableRow(env.GetList(args, i).GetValue())
			}
			env.WriteString("</tbody>")
		}
		env.WriteString("</table>")
	}},
	{sexpr.SymCell, 0, -1, makeCellFn("")},
	{sexpr.SymCellCenter, 0, -1, makeCellFn("center")},
	{sexpr.SymCellLeft, 0, -1, makeCellFn("left")},
	{sexpr.SymCellRight, 0, -1, makeCellFn("right")},
	{sexpr.SymRegionBlock, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a := env.GetAttributes(args, 0)
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		env.writeRegion(args, a, "div")
	}},
	{sexpr.SymRegionQuote, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.writeRegion(args, nil, "blockquote")
	}},
	{sexpr.SymRegionVerse, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.writeRegion(args, nil, "div")
	}},
	{sexpr.SymVerbatimComment, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if env.GetAttributes(args, 0).HasDefault() {
			if s := env.GetString(args, 1); s != "" {
				env.WriteString("<!--\n")
				env.WriteEscaped(s)
				env.WriteString("\n-->")
			}
		}
	}},
	{sexpr.SymVerbatimEval, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a := env.GetAttributes(args, 0).AddClass("zs-eval")
		env.writeVerbatim(args, a)
	}},
	{sexpr.SymVerbatimHTML, 0, -1, execHTML},
	{sexpr.SymVerbatimMath, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a := env.GetAttributes(args, 0).AddClass("zs-math")
		env.writeVerbatim(args, a)
	}},
	{sexpr.SymVerbatimProg, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a := setProgLang(env.GetAttributes(args, 0))
		oldVisible := env.visibleSpace
		if a.HasDefault() {
			a = a.RemoveDefault()
			env.visibleSpace = true
		}
		env.writeVerbatim(args, a)
		env.visibleSpace = oldVisible
	}},
	{sexpr.SymVerbatimZettel, 0, -1, DoNothingFn},
	{sexpr.SymBLOB, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.writeBLOB(env.GetString(args, 0), env.GetString(args, 1), env.GetString(args, 2))
	}},
	{sexpr.SymTransclude, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		ref := env.GetList(args, 0)
		refPair := ref.GetValue()
		refKind := env.GetSymbol(refPair, 0)
		if refKind == nil {
			return
		}
		if refValue := env.GetString(refPair, 1); refValue != "" {
			if sexpr.SymRefStateExternal.Equal(refKind) {
				a := attrs.Attributes{}.Set("src", refValue).AddClass("external")
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
		log.Println("TRAN", args)
	}},
	{sexpr.SymText, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if len(args) > 0 {
			env.WriteEscaped(env.GetString(args, 0))
		}
	}},
	{sexpr.SymSpace, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if len(args) == 0 {
			env.WriteString(" ")
			return
		}
		env.WriteEscaped(env.GetString(args, 0))
	}},
	{sexpr.SymSoft, 0, -1, func(env *EncEnvironment, _ []sxpf.Value) { env.WriteString(" ") }},
	{sexpr.SymHard, 0, -1, func(env *EncEnvironment, _ []sxpf.Value) { env.WriteString("<br>") }},
	{sexpr.SymTag, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if len(args) > 0 {
			env.WriteEscaped(env.GetString(args, 0))
		}
	}},
	{sexpr.SymLink, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if env.noLinks {
			spanList := sxpf.NewList(sexpr.SymFormatSpan)
			spanList.Append(args...)
			sxpf.Evaluate(env, spanList)
			return
		}
		if env.MissingArgs(args, 2) {
			return
		}
		a := env.GetAttributes(args, 0)
		ref := env.GetList(args, 1)
		if ref == nil {
			return
		}
		refPair := ref.GetValue()
		refKind := env.GetSymbol(refPair, 0)
		if refKind == nil {
			return
		}
		refValue := env.GetString(refPair, 1)
		switch {
		case sexpr.SymRefStateExternal.Equal(refKind):
			a = a.Set("href", refValue).AddClass("external")
		case sexpr.SymRefStateZettel.Equal(refKind), sexpr.SymRefStateBased.Equal(refKind), sexpr.SymRefStateHosted.Equal(refKind), sexpr.SymRefStateSelf.Equal(refKind):
			a = a.Set("href", refValue)
		case sexpr.SymRefStateBroken.Equal(refKind):
			a = a.AddClass("broken")
		default:
			log.Println("LINK", sxpf.NewList(args...))
		}
		env.WriteString("<a")
		env.WriteAttributes(a)
		env.WriteString(">")

		if in := args[2:]; len(in) == 0 {
			env.WriteString(refValue)
		} else {
			sxpf.EvaluateSlice(env, in)
		}
		env.WriteString("</a>")
	}},
	{sexpr.SymEmbed, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if syntax := env.GetString(args, 2); syntax == api.ValueSyntaxSVG {
			ref := env.GetList(args, 1)
			refPair := ref.GetValue()
			env.WriteStrings(
				`<figure><embed type="image/svg+xml" src="`, "/", env.GetString(refPair, 1), ".svg", "\" /></figure>")
		} else {
			env.WriteImage(args)
		}
	}},
	{sexpr.SymEmbedBLOB, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a, syntax, data := env.GetAttributes(args, 0), env.GetString(args, 1), env.GetString(args, 2)
		title, _ := a.Get("title")
		env.writeBLOB(title, syntax, data)
	}},
	{sexpr.SymCite, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteStartTag("span", env.GetAttributes(args, 0))
		if key := env.GetString(args, 1); key != "" {
			env.WriteEscaped(key)
			if text := args[2:]; len(text) > 0 {
				env.WriteString(", ")
				sxpf.EvaluateSlice(env, text)
			}
		}
		env.WriteString("</span>")
	}},
	{sexpr.SymMark, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if env.noLinks {
			spanList := sxpf.NewList(sexpr.SymFormatSpan)
			spanList.Append(args...)
			sxpf.Evaluate(env, spanList)
			return
		}
		if fragment := env.GetString(args, 2); fragment != "" {
			env.WriteString(`<a id="`)
			env.WriteString(env.unique)
			env.WriteString(fragment)
			env.WriteString(`">`)
			sxpf.EvaluateSlice(env, args[3:])
			env.WriteString("</a>")
		} else {
			sxpf.EvaluateSlice(env, args[3:])
		}
	}},
	{sexpr.SymFootnote, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if env.writeFootnote {
			a := env.GetAttributes(args, 0)
			env.footnotes = append(env.footnotes, sfootnodeInfo{args[1:], a})
			n := strconv.Itoa(len(env.footnotes))
			un := env.unique + n
			env.WriteStrings(
				`<sup id="fnref:`, un, `"><a class="zs-noteref" href="#fn:`, un,
				`" role="doc-noteref">`, n, `</a></sup>`)
		}
	}},
	{sexpr.SymFormatDelete, 0, -1, makeFormatFn("del")},
	{sexpr.SymFormatEmph, 0, -1, makeFormatFn("em")},
	{sexpr.SymFormatInsert, 0, -1, makeFormatFn("ins")},
	{sexpr.SymFormatQuote, 0, -1, makeFormatFn("q")},
	{sexpr.SymFormatSpan, 0, -1, makeFormatFn("span")},
	{sexpr.SymFormatStrong, 0, -1, makeFormatFn("strong")},
	{sexpr.SymFormatSub, 0, -1, makeFormatFn("sub")},
	{sexpr.SymFormatSuper, 0, -1, makeFormatFn("sup")},
	{sexpr.SymLiteralComment, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		if env.GetAttributes(args, 0).HasDefault() {
			if s := env.GetString(args, 1); s != "" {
				env.WriteString("<!-- ")
				env.WriteEscaped(s)
				env.WriteString("-->")
			}
		}
	}},
	{sexpr.SymLiteralHTML, 0, -1, execHTML},
	{sexpr.SymLiteralInput, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.writeLiteral(args, nil, "kbd")
	}},
	{sexpr.SymLiteralMath, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a := env.GetAttributes(args, 0).AddClass("zs-math")
		env.writeLiteral(args, a, "code")
	}},
	{sexpr.SymLiteralOutput, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		env.writeLiteral(args, nil, "samp")
	}},
	{sexpr.SymLiteralProg, 0, -1, func(env *EncEnvironment, args []sxpf.Value) {
		a := setProgLang(env.GetAttributes(args, 0))
		env.writeLiteral(args, a, "code")
	}},
	{sexpr.SymLiteralZettel, 0, -1, DoNothingFn},
}

// DoNothingFn is a function that does nothing.
func DoNothingFn(*EncEnvironment, []sxpf.Value) { /* Should really do nothing */ }

func makeListFn(tag string) encodingFunc {
	return func(env *EncEnvironment, args []sxpf.Value) {
		env.WriteStartTag(tag, nil)
		for _, items := range args {
			env.WriteStartTag("li", nil)
			sxpf.Evaluate(env, items)
			env.WriteEndTag("li")
		}
		env.WriteEndTag(tag)
	}
}

func (env *EncEnvironment) writeTableRow(cells []sxpf.Value) {
	if len(cells) > 0 {
		env.WriteString("<tr>")
		for _, cell := range cells {
			sxpf.Evaluate(env, cell)
		}
		env.WriteString("</tr>")
	}
}
func makeCellFn(align string) encodingFunc {
	return func(env *EncEnvironment, args []sxpf.Value) {
		if align == "" {
			env.WriteString("<td>")
		} else {
			env.WriteStrings(`<td class="`, align, `">`)
		}
		sxpf.EvaluateSlice(env, args)
		env.WriteString("</td>")
	}
}

func (env *EncEnvironment) writeRegion(args []sxpf.Value, a attrs.Attributes, tag string) {
	if a == nil {
		a = env.GetAttributes(args, 0)
	}
	env.WriteStartTag(tag, a)
	sxpf.Evaluate(env, env.GetList(args, 1))
	if cite := env.GetList(args, 2).GetValue(); len(cite) > 0 {
		env.WriteString("<cite>")
		sxpf.EvaluateSlice(env, cite)
		env.WriteString("</cite>")
	}
	env.WriteEndTag(tag)
}

func (env *EncEnvironment) writeVerbatim(args []sxpf.Value, a attrs.Attributes) {
	env.WriteString("<pre>")
	env.WriteStartTag("code", a)
	env.WriteEscapedLiteral(env.GetString(args, 1))
	env.WriteString("</code></pre>")
}

func execHTML(env *EncEnvironment, args []sxpf.Value) {
	if s := env.GetString(args, 1); s != "" && IsSafe(s) {
		env.WriteString(s)
	}
}

func (env *EncEnvironment) writeBLOB(title, syntax, data string) {
	if data == "" {
		return
	}
	switch syntax {
	case "":
	case api.ValueSyntaxSVG:
		// TODO: add  title as description
		env.WriteStrings("<p>", data, "</p>")
	default:
		env.WriteStrings(`<p><img src="data:image/`, syntax, ";base64,", data)
		if title != "" {
			env.WriteString(`" `)
			env.WriteOneAttribute("title", title)
		}
		env.WriteString(`"></p>`)
	}
}

func makeFormatFn(tag string) encodingFunc {
	return func(env *EncEnvironment, args []sxpf.Value) {
		if env.MissingArgs(args, 1) {
			return
		}
		a := env.GetAttributes(args, 0)
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		env.WriteStartTag(tag, a)
		sxpf.EvaluateSlice(env, args[1:])
		env.WriteEndTag(tag)
	}
}

func (env *EncEnvironment) writeLiteral(args []sxpf.Value, a attrs.Attributes, tag string) {
	if env.MissingArgs(args, 2) {
		return
	}

	if a == nil {
		a = env.GetAttributes(args, 0)
	}
	oldVisible := env.visibleSpace
	if a.HasDefault() {
		env.visibleSpace = true
		a = a.RemoveDefault()
	}
	env.WriteStartTag(tag, a)
	env.visibleSpace = oldVisible
	env.WriteString(env.GetString(args, 1))
	env.WriteEndTag(tag)
}

func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}
