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

// Package shtml transforms a s-expr encoded zettel AST into a s-expr representation of HTML.
package shtml

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"t73f.de/r/sx"
	"t73f.de/r/sxwebs/sxhtml"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/attrs"
	"t73f.de/r/zsc/domain/meta"
	"t73f.de/r/zsc/sz"
)

// Evaluator will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Evaluator struct {
	headingOffset int64
	unique        string
	noLinks       bool // true iff output must not include links

	fns     map[string]EvalFn
	minArgs map[string]int
}

// NewEvaluator creates a new Evaluator object.
func NewEvaluator(headingOffset int) *Evaluator {
	ev := &Evaluator{
		headingOffset: int64(headingOffset),

		fns:     make(map[string]EvalFn, 128),
		minArgs: make(map[string]int, 128),
	}
	ev.bindMetadata()
	ev.bindBlocks()
	ev.bindInlines()
	return ev
}

// SetUnique sets a prefix to make several HTML ids unique.
func (ev *Evaluator) SetUnique(s string) { ev.unique = s }

// IsValidName returns true, if name is a valid symbol name.
func isValidName(s string) bool { return s != "" }

// EvaluateAttributes transforms the given attributes into a HTML s-expression.
func EvaluateAttributes(a attrs.Attributes) *sx.Pair {
	if len(a) == 0 {
		return nil
	}
	plist := sx.Nil()
	keys := a.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		if key != attrs.DefaultAttribute && isValidName(key) {
			plist = plist.Cons(sx.Cons(sx.MakeSymbol(key), sx.MakeString(a[key])))
		}
	}
	if plist == nil {
		return nil
	}
	return plist.Cons(sxhtml.SymAttr)
}

// Evaluate a metadata s-expression into a list of HTML s-expressions.
func (ev *Evaluator) Evaluate(lst *sx.Pair, env *Environment) (*sx.Pair, error) {
	result := ev.Eval(lst, env)
	if err := env.err; err != nil {
		return nil, err
	}
	pair, isPair := sx.GetPair(result)
	if !isPair {
		return nil, fmt.Errorf("evaluation does not result in a pair, but %T/%v", result, result)
	}

	for i := 0; i < len(env.endnotes); i++ {
		// May extend tr.endnotes -> do not use for i := range len(...)!!!

		if env.endnotes[i].noteHx != nil {
			continue
		}

		noteHx, _ := ev.EvaluateList(env.endnotes[i].noteAST, env)
		env.endnotes[i].noteHx = noteHx
	}

	return pair, nil
}

// EvaluateList will evaluate all list elements separately and returns them as a sx.Pair list
func (ev *Evaluator) EvaluateList(lst sx.Vector, env *Environment) (*sx.Pair, error) {
	var result sx.ListBuilder
	for _, elem := range lst {
		p := ev.Eval(elem, env)
		result.Add(p)
	}
	if err := env.err; err != nil {
		return nil, err
	}
	return result.List(), nil
}

// Endnotes returns a SHTML object with all collected endnotes.
func Endnotes(env *Environment) *sx.Pair {
	if env.err != nil || len(env.endnotes) == 0 {
		return nil
	}

	var result sx.ListBuilder
	result.AddN(
		SymOL,
		sx.Nil().Cons(sx.Cons(SymAttrClass, sx.MakeString("zs-endnotes"))).Cons(sxhtml.SymAttr),
	)
	for i, fni := range env.endnotes {
		noteNum := strconv.Itoa(i + 1)
		attrs := fni.attrs.Cons(sx.Cons(SymAttrClass, sx.MakeString("zs-endnote"))).
			Cons(sx.Cons(SymAttrValue, sx.MakeString(noteNum))).
			Cons(sx.Cons(SymAttrID, sx.MakeString("fn:"+fni.noteID))).
			Cons(sx.Cons(SymAttrRole, sx.MakeString("doc-endnote"))).
			Cons(sxhtml.SymAttr)

		backref := sx.Nil().Cons(sx.MakeString("\u21a9\ufe0e")).
			Cons(sx.Nil().
				Cons(sx.Cons(SymAttrClass, sx.MakeString("zs-endnote-backref"))).
				Cons(sx.Cons(SymAttrHref, sx.MakeString("#fnref:"+fni.noteID))).
				Cons(sx.Cons(SymAttrRole, sx.MakeString("doc-backlink"))).
				Cons(sxhtml.SymAttr)).
			Cons(SymA)

		var li sx.ListBuilder
		li.AddN(SymLI, attrs)
		li.ExtendBang(fni.noteHx)
		li.AddN(sx.MakeString(" "), backref)
		result.Add(li.List())
	}
	return result.List()
}

// Environment where sz objects are evaluated to shtml objects
type Environment struct {
	err          error
	langStack    LangStack
	endnotes     []endnoteInfo
	quoteNesting uint
}
type endnoteInfo struct {
	noteID  string    // link id
	noteAST sx.Vector // Endnote as list of AST inline elements
	attrs   *sx.Pair  // attrs a-list
	noteHx  *sx.Pair  // Endnote as SxHTML
}

// MakeEnvironment builds a new evaluation environment.
func MakeEnvironment(lang string) Environment {
	return Environment{
		err:          nil,
		langStack:    NewLangStack(lang),
		endnotes:     nil,
		quoteNesting: 0,
	}
}

// GetError returns the last error found.
func (env *Environment) GetError() error { return env.err }

// Reset the environment.
func (env *Environment) Reset() {
	env.langStack.Reset()
	env.endnotes = nil
	env.quoteNesting = 0
}

// pushAttribute adds the current attributes to the environment.
func (env *Environment) pushAttributes(a attrs.Attributes) {
	if value, ok := a.Get("lang"); ok {
		env.langStack.Push(value)
	} else {
		env.langStack.Dup()
	}
}

// popAttributes removes the current attributes from the envrionment.
func (env *Environment) popAttributes() { env.langStack.Pop() }

// getLanguage returns the current language.
func (env *Environment) getLanguage() string { return env.langStack.Top() }

func (env *Environment) getQuotes() (string, string, bool) {
	qi := GetQuoteInfo(env.getLanguage())
	leftQ, rightQ := qi.GetQuotes(env.quoteNesting)
	return leftQ, rightQ, qi.GetNBSp()
}

// EvalFn is a function to be called for evaluation.
type EvalFn func(sx.Vector, *Environment) sx.Object

func (ev *Evaluator) bind(sym *sx.Symbol, minArgs int, fn EvalFn) {
	symVal := sym.GetValue()
	ev.fns[symVal] = fn
	if minArgs > 0 {
		ev.minArgs[symVal] = minArgs
	}
}

// ResolveBinding returns the function bound to the given name.
func (ev *Evaluator) ResolveBinding(sym *sx.Symbol) EvalFn {
	if fn, found := ev.fns[sym.GetValue()]; found {
		return fn
	}
	return nil
}

// Rebind overwrites a binding, but leaves the minimum number of arguments intact.
func (ev *Evaluator) Rebind(sym *sx.Symbol, fn EvalFn) {
	symVal := sym.GetValue()
	if _, found := ev.fns[symVal]; !found {
		panic(sym)
	}
	ev.fns[symVal] = fn
}

func (ev *Evaluator) bindMetadata() {
	ev.bind(sz.SymMeta, 0, ev.evalList)
	evalMetaString := func(args sx.Vector, env *Environment) sx.Object {
		a := make(attrs.Attributes, 2).
			Set("name", getSymbol(args[0], env).GetValue()).
			Set("content", getString(args[1], env).GetValue())
		return ev.EvaluateMeta(a)
	}
	ev.bind(sz.SymTypeCredential, 2, evalMetaString)
	ev.bind(sz.SymTypeEmpty, 2, evalMetaString)
	ev.bind(sz.SymTypeID, 2, evalMetaString)
	ev.bind(sz.SymTypeNumber, 2, evalMetaString)
	ev.bind(sz.SymTypeString, 2, evalMetaString)
	ev.bind(sz.SymTypeTimestamp, 2, evalMetaString)
	ev.bind(sz.SymTypeURL, 2, evalMetaString)
	ev.bind(sz.SymTypeWord, 2, evalMetaString)

	evalMetaSet := func(args sx.Vector, env *Environment) sx.Object {
		var sb strings.Builder
		for obj := range getList(args[1], env).Values() {
			sb.WriteByte(' ')
			sb.WriteString(getString(obj, env).GetValue())
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", getSymbol(args[0], env).GetValue()).
			Set("content", s)
		return ev.EvaluateMeta(a)
	}
	ev.bind(sz.SymTypeIDSet, 2, evalMetaSet)
	ev.bind(sz.SymTypeTagSet, 2, evalMetaSet)
}

// EvaluateMeta returns HTML meta object for an attribute.
func (ev *Evaluator) EvaluateMeta(a attrs.Attributes) *sx.Pair {
	return sx.Nil().Cons(EvaluateAttributes(a)).Cons(SymMeta)
}

func (ev *Evaluator) bindBlocks() {
	ev.bind(sz.SymBlock, 0, ev.evalList)
	ev.bind(sz.SymPara, 0, func(args sx.Vector, env *Environment) sx.Object {
		return ev.evalSlice(args, env).Cons(SymP)
	})
	ev.bind(sz.SymHeading, 5, func(args sx.Vector, env *Environment) sx.Object {
		nLevel := getInt64(args[0], env)
		if nLevel <= 0 {
			env.err = fmt.Errorf("%v is a negative heading level", nLevel)
			return sx.Nil()
		}
		level := strconv.FormatInt(nLevel+ev.headingOffset, 10)
		headingSymbol := sx.MakeSymbol("h" + level)

		a := GetAttributes(args[1], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if fragment := getString(args[3], env).GetValue(); fragment != "" {
			a = a.Set("id", ev.unique+fragment)
		}

		if result, _ := ev.EvaluateList(args[4:], env); result != nil {
			if len(a) > 0 {
				result = result.Cons(EvaluateAttributes(a))
			}
			return result.Cons(headingSymbol)
		}
		return sx.MakeList(headingSymbol, sx.MakeString("<MISSING TEXT>"))
	})
	ev.bind(sz.SymThematic, 0, func(args sx.Vector, env *Environment) sx.Object {
		result := sx.Nil()
		if len(args) > 0 {
			if attrList := getList(args[0], env); attrList != nil {
				result = result.Cons(EvaluateAttributes(sz.GetAttributes(attrList)))
			}
		}
		return result.Cons(SymHR)
	})

	ev.bind(sz.SymListOrdered, 1, ev.makeListFn(SymOL))
	ev.bind(sz.SymListUnordered, 1, ev.makeListFn(SymUL))
	ev.bind(sz.SymListQuote, 1, func(args sx.Vector, env *Environment) sx.Object {
		if len(args) == 1 {
			return sx.Nil()
		}
		var result sx.ListBuilder
		result.Add(symBLOCKQUOTE)
		if attrs := EvaluateAttributes(GetAttributes(args[0], env)); attrs != nil {
			result.Add(attrs)
		}
		for _, elem := range args[1:] {
			if quote, isPair := sx.GetPair(ev.Eval(elem, env)); isPair {
				result.Add(quote.Cons(sxhtml.SymListSplice))
			}
		}
		return result.List()
	})

	ev.bind(sz.SymDescription, 1, func(args sx.Vector, env *Environment) sx.Object {
		if len(args) == 1 {
			return sx.Nil()
		}
		var result sx.ListBuilder
		result.Add(symDL)
		if attrs := EvaluateAttributes(GetAttributes(args[0], env)); attrs != nil {
			result.Add(attrs)
		}
		for pos := 1; pos < len(args); pos++ {
			term := ev.evalDescriptionTerm(getList(args[pos], env), env)
			result.Add(term.Cons(symDT))
			pos++
			if pos >= len(args) {
				break
			}
			ddBlock := getList(ev.Eval(args[pos], env), env)
			if ddBlock == nil {
				continue
			}
			for ddlst := range ddBlock.Values() {
				dditem := getList(ddlst, env)
				result.Add(dditem.Cons(symDD))
			}
		}
		return result.List()
	})

	ev.bind(sz.SymTable, 1, func(args sx.Vector, env *Environment) sx.Object {
		thead := sx.Nil()
		if header := getList(args[0], env); !sx.IsNil(header) {
			thead = sx.Nil().Cons(ev.evalTableRow(symTH, header, env)).Cons(symTHEAD)
		}

		var tbody sx.ListBuilder
		if len(args) > 1 {
			tbody.Add(symTBODY)
			for _, row := range args[1:] {
				tbody.Add(ev.evalTableRow(symTD, getList(row, env), env))
			}
		}

		table := sx.Nil()
		if !tbody.IsEmpty() {
			table = table.Cons(tbody.List())
		}
		if thead != nil {
			table = table.Cons(thead)
		}
		if table == nil {
			return sx.Nil()
		}
		return table.Cons(symTABLE)
	})
	ev.bind(sz.SymCell, 0, ev.makeCellFn(""))
	ev.bind(sz.SymCellCenter, 0, ev.makeCellFn("center"))
	ev.bind(sz.SymCellLeft, 0, ev.makeCellFn("left"))
	ev.bind(sz.SymCellRight, 0, ev.makeCellFn("right"))

	ev.bind(sz.SymRegionBlock, 2, ev.makeRegionFn(SymDIV, true))
	ev.bind(sz.SymRegionQuote, 2, ev.makeRegionFn(symBLOCKQUOTE, false))
	ev.bind(sz.SymRegionVerse, 2, ev.makeRegionFn(SymDIV, false))

	ev.bind(sz.SymVerbatimComment, 1, func(args sx.Vector, env *Environment) sx.Object {
		if GetAttributes(args[0], env).HasDefault() {
			if len(args) > 1 {
				if s := getString(args[1], env); s.GetValue() != "" {
					return sx.Nil().Cons(s).Cons(sxhtml.SymBlockComment)
				}
			}
		}
		return nil
	})
	ev.bind(sz.SymVerbatimEval, 2, func(args sx.Vector, env *Environment) sx.Object {
		return evalVerbatim(GetAttributes(args[0], env).AddClass("zs-eval"), getString(args[1], env))
	})
	ev.bind(sz.SymVerbatimHTML, 2, ev.evalHTML)
	ev.bind(sz.SymVerbatimMath, 2, func(args sx.Vector, env *Environment) sx.Object {
		return evalVerbatim(GetAttributes(args[0], env).AddClass("zs-math"), getString(args[1], env))
	})
	ev.bind(sz.SymVerbatimCode, 2, func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env)
		content := getString(args[1], env)
		if a.HasDefault() {
			content = sx.MakeString(visibleReplacer.Replace(content.GetValue()))
		}
		return evalVerbatim(a, content)
	})
	ev.bind(sz.SymVerbatimZettel, 0, nilFn)
	ev.bind(sz.SymBLOB, 3, func(args sx.Vector, env *Environment) sx.Object {
		return evalBLOB(getList(args[0], env), getString(args[1], env), getString(args[2], env))
	})
	ev.bind(sz.SymTransclude, 2, func(args sx.Vector, env *Environment) sx.Object {
		if refSym, refValue := GetReference(args[1], env); refSym != nil {
			if refSym.IsEqualSymbol(sz.SymRefStateExternal) {
				a := GetAttributes(args[0], env).Set("src", refValue).AddClass("external")
				// TODO: if len(args) > 2, add "alt" attr based on args[2:], as in SymEmbed
				return sx.Nil().Cons(sx.Nil().Cons(EvaluateAttributes(a)).Cons(SymIMG)).Cons(SymP)
			}
			return sx.MakeList(
				sxhtml.SymInlineComment,
				sx.MakeString("transclude"),
				refSym,
				sx.MakeString("->"),
				sx.MakeString(refValue),
			)
		}
		return ev.evalSlice(args, env)
	})
}

func (ev *Evaluator) makeListFn(sym *sx.Symbol) EvalFn {
	return func(args sx.Vector, env *Environment) sx.Object {
		var result sx.ListBuilder
		result.Add(sym)
		if attrs := EvaluateAttributes(GetAttributes(args[0], env)); attrs != nil {
			result.Add(attrs)
		}
		if len(args) > 1 {
			for _, elem := range args[1:] {
				item := sx.Nil().Cons(SymLI)
				if res, isPair := sx.GetPair(ev.Eval(elem, env)); isPair {
					item.ExtendBang(res)
				}
				result.Add(item)
			}
		}
		return result.List()
	}
}

func (ev *Evaluator) evalDescriptionTerm(term *sx.Pair, env *Environment) *sx.Pair {
	var result sx.ListBuilder
	for obj := range term.Values() {
		elem := ev.Eval(obj, env)
		result.Add(elem)
	}
	return result.List()
}

func (ev *Evaluator) evalTableRow(sym *sx.Symbol, pairs *sx.Pair, env *Environment) *sx.Pair {
	if pairs == nil {
		return nil
	}
	var row sx.ListBuilder
	row.Add(symTR)
	for obj := range pairs.Values() {
		row.Add(sx.Cons(sym, ev.Eval(obj, env)))
	}
	return row.List()
}
func (ev *Evaluator) makeCellFn(align string) EvalFn {
	return func(args sx.Vector, env *Environment) sx.Object {
		tdata := ev.evalSlice(args, env)
		if align != "" {
			tdata = tdata.Cons(EvaluateAttributes(attrs.Attributes{"class": align}))
		}
		return tdata
	}
}

func (ev *Evaluator) makeRegionFn(sym *sx.Symbol, genericToClass bool) EvalFn {
	return func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if genericToClass {
			if val, found := a.Get(""); found {
				a = a.Remove("").AddClass(val)
			}
		}
		var result sx.ListBuilder
		result.Add(sym)
		if len(a) > 0 {
			result.Add(EvaluateAttributes(a))
		}
		if region, isPair := sx.GetPair(args[1]); isPair {
			if evalRegion := ev.EvalPairList(region, env); evalRegion != nil {
				result.ExtendBang(evalRegion)
			}
		}
		if len(args) > 2 {
			if cite, _ := ev.EvaluateList(args[2:], env); cite != nil {
				result.Add(cite.Cons(symCITE))
			}
		}
		return result.List()
	}
}

func evalVerbatim(a attrs.Attributes, s sx.String) sx.Object {
	a = setProgLang(a)
	code := sx.Nil().Cons(s)
	if al := EvaluateAttributes(a); al != nil {
		code = code.Cons(al)
	}
	code = code.Cons(symCODE)
	return sx.Nil().Cons(code).Cons(symPRE)
}

func (ev *Evaluator) bindInlines() {
	ev.bind(sz.SymInline, 0, ev.evalList)
	ev.bind(sz.SymText, 1, func(args sx.Vector, env *Environment) sx.Object { return getString(args[0], env) })
	ev.bind(sz.SymSoft, 0, func(sx.Vector, *Environment) sx.Object { return sx.MakeString(" ") })
	ev.bind(sz.SymHard, 0, func(sx.Vector, *Environment) sx.Object { return sx.Nil().Cons(symBR) })

	ev.bind(sz.SymLink, 2, func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refSym, refValue := GetReference(args[1], env)
		switch refSym {
		case sz.SymRefStateZettel, sz.SymRefStateSelf, sz.SymRefStateFound, sz.SymRefStateHosted, sz.SymRefStateBased:
			return ev.evalLink(a.Set("href", refValue), refValue, args[2:], env)

		case sz.SymRefStateExternal:
			return ev.evalLink(a.Set("href", refValue).Add("rel", "external"), refValue, args[2:], env)

		case sz.SymRefStateQuery:
			query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue)
			return ev.evalLink(a.Set("href", query), refValue, args[2:], env)

		case sz.SymRefStateBroken:
			return ev.evalLink(a.AddClass("broken"), refValue, args[2:], env)
		}

		// sz.SymRefStateInvalid or unknown
		var inline *sx.Pair
		if len(args) > 2 {
			inline = ev.evalSlice(args[2:], env)
		}
		if inline == nil {
			inline = sx.Nil().Cons(sx.MakeString(refValue))
		}
		return inline.Cons(SymSPAN)
	})

	ev.bind(sz.SymEmbed, 3, func(args sx.Vector, env *Environment) sx.Object {
		_, refValue := GetReference(args[1], env)
		a := GetAttributes(args[0], env).Set("src", refValue)
		if len(args) > 3 {
			var sb strings.Builder
			flattenText(&sb, sx.MakeList(args[3:]...))
			if d := sb.String(); d != "" {
				a = a.Set("alt", d)
			}
		}
		return sx.MakeList(SymIMG, EvaluateAttributes(a))
	})
	ev.bind(sz.SymEmbedBLOB, 3, func(args sx.Vector, env *Environment) sx.Object {
		a, syntax, data := GetAttributes(args[0], env), getString(args[1], env), getString(args[2], env)
		summary, hasSummary := a.Get(meta.KeySummary)
		if !hasSummary {
			summary = ""
		}
		return evalBLOB(
			sx.MakeList(sxhtml.SymListSplice, sx.MakeString(summary)),
			syntax,
			data,
		)
	})

	ev.bind(sz.SymCite, 2, func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		result := sx.Nil()
		if key := getString(args[1], env); key.GetValue() != "" {
			if len(args) > 2 {
				result = ev.evalSlice(args[2:], env).Cons(sx.MakeString(", "))
			}
			result = result.Cons(key)
		}
		if len(a) > 0 {
			result = result.Cons(EvaluateAttributes(a))
		}
		if result == nil {
			return nil
		}
		return result.Cons(SymSPAN)
	})
	ev.bind(sz.SymMark, 3, func(args sx.Vector, env *Environment) sx.Object {
		result := ev.evalSlice(args[3:], env)
		if !ev.noLinks {
			if fragment := getString(args[2], env).GetValue(); fragment != "" {
				a := attrs.Attributes{"id": fragment + ev.unique}
				return result.Cons(EvaluateAttributes(a)).Cons(SymA)
			}
		}
		return result.Cons(SymSPAN)
	})
	ev.bind(sz.SymEndnote, 1, func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		attrPlist := sx.Nil()
		if len(a) > 0 {
			if attrs := EvaluateAttributes(a); attrs != nil {
				attrPlist = attrs.Tail()
			}
		}

		noteNum := strconv.Itoa(len(env.endnotes) + 1)
		noteID := ev.unique + noteNum
		env.endnotes = append(env.endnotes, endnoteInfo{
			noteID: noteID, noteAST: args[1:], noteHx: nil, attrs: attrPlist})
		hrefAttr := sx.Nil().Cons(sx.Cons(SymAttrRole, sx.MakeString("doc-noteref"))).
			Cons(sx.Cons(SymAttrHref, sx.MakeString("#fn:"+noteID))).
			Cons(sx.Cons(SymAttrClass, sx.MakeString("zs-noteref"))).
			Cons(sxhtml.SymAttr)
		href := sx.Nil().Cons(sx.MakeString(noteNum)).Cons(hrefAttr).Cons(SymA)
		supAttr := sx.Nil().Cons(sx.Cons(SymAttrID, sx.MakeString("fnref:"+noteID))).Cons(sxhtml.SymAttr)
		return sx.Nil().Cons(href).Cons(supAttr).Cons(symSUP)
	})

	ev.bind(sz.SymFormatDelete, 1, ev.makeFormatFn(symDEL))
	ev.bind(sz.SymFormatEmph, 1, ev.makeFormatFn(symEM))
	ev.bind(sz.SymFormatInsert, 1, ev.makeFormatFn(symINS))
	ev.bind(sz.SymFormatMark, 1, ev.makeFormatFn(symMARK))
	ev.bind(sz.SymFormatQuote, 1, ev.evalQuote)
	ev.bind(sz.SymFormatSpan, 1, ev.makeFormatFn(SymSPAN))
	ev.bind(sz.SymFormatStrong, 1, ev.makeFormatFn(SymSTRONG))
	ev.bind(sz.SymFormatSub, 1, ev.makeFormatFn(symSUB))
	ev.bind(sz.SymFormatSuper, 1, ev.makeFormatFn(symSUP))

	ev.bind(sz.SymLiteralComment, 1, func(args sx.Vector, env *Environment) sx.Object {
		if GetAttributes(args[0], env).HasDefault() {
			if len(args) > 1 {
				if s := getString(ev.Eval(args[1], env), env); s.GetValue() != "" {
					return sx.Nil().Cons(s).Cons(sxhtml.SymInlineComment)
				}
			}
		}
		return sx.Nil()
	})
	ev.bind(sz.SymLiteralInput, 2, func(args sx.Vector, env *Environment) sx.Object {
		return evalLiteral(args, nil, symKBD, env)
	})
	ev.bind(sz.SymLiteralMath, 2, func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env).AddClass("zs-math")
		return evalLiteral(args, a, symCODE, env)
	})
	ev.bind(sz.SymLiteralOutput, 2, func(args sx.Vector, env *Environment) sx.Object {
		return evalLiteral(args, nil, symSAMP, env)
	})
	ev.bind(sz.SymLiteralCode, 2, func(args sx.Vector, env *Environment) sx.Object {
		return evalLiteral(args, nil, symCODE, env)
	})
}

func (ev *Evaluator) makeFormatFn(sym *sx.Symbol) EvalFn {
	return func(args sx.Vector, env *Environment) sx.Object {
		a := GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if val, hasClass := a.Get(""); hasClass {
			a = a.Remove("").AddClass(val)
		}
		res := ev.evalSlice(args[1:], env)
		if len(a) > 0 {
			res = res.Cons(EvaluateAttributes(a))
		}
		return res.Cons(sym)
	}
}

func (ev *Evaluator) evalQuote(args sx.Vector, env *Environment) sx.Object {
	a := GetAttributes(args[0], env)
	env.pushAttributes(a)
	defer env.popAttributes()

	if val, hasClass := a.Get(""); hasClass {
		a = a.Remove("").AddClass(val)
	}
	leftQ, rightQ, withNbsp := env.getQuotes()

	env.quoteNesting++
	res := ev.evalSlice(args[1:], env)
	env.quoteNesting--

	lastPair := res.LastPair()
	if lastPair.IsNil() {
		res = sx.Cons(sx.MakeList(sxhtml.SymNoEscape, sx.MakeString(leftQ), sx.MakeString(rightQ)), sx.Nil())
	} else {
		if withNbsp {
			lastPair.AppendBang(sx.MakeList(sxhtml.SymNoEscape, sx.MakeString("&nbsp;"), sx.MakeString(rightQ)))
			res = res.Cons(sx.MakeList(sxhtml.SymNoEscape, sx.MakeString(leftQ), sx.MakeString("&nbsp;")))
		} else {
			lastPair.AppendBang(sx.MakeList(sxhtml.SymNoEscape, sx.MakeString(rightQ)))
			res = res.Cons(sx.MakeList(sxhtml.SymNoEscape, sx.MakeString(leftQ)))
		}
	}
	if len(a) > 0 {
		res = res.Cons(EvaluateAttributes(a))
		return res.Cons(SymSPAN)
	}
	return res.Cons(sxhtml.SymListSplice)
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func evalLiteral(args sx.Vector, a attrs.Attributes, sym *sx.Symbol, env *Environment) sx.Object {
	if a == nil {
		a = GetAttributes(args[0], env)
	}
	a = setProgLang(a)
	literal := getString(args[1], env).GetValue()
	if a.HasDefault() {
		a = a.RemoveDefault()
		literal = visibleReplacer.Replace(literal)
	}
	res := sx.Nil().Cons(sx.MakeString(literal))
	if len(a) > 0 {
		res = res.Cons(EvaluateAttributes(a))
	}
	return res.Cons(sym)
}
func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}

func (ev *Evaluator) evalHTML(args sx.Vector, env *Environment) sx.Object {
	if s := getString(ev.Eval(args[1], env), env); s.GetValue() != "" && IsSafe(s.GetValue()) {
		return sx.Nil().Cons(s).Cons(sxhtml.SymNoEscape)
	}
	return nil
}

func evalBLOB(description *sx.Pair, syntax, data sx.String) sx.Object {
	if data.GetValue() == "" {
		return sx.Nil()
	}
	switch syntax.GetValue() {
	case "":
		return sx.Nil()
	case meta.ValueSyntaxSVG:
		return sx.Nil().Cons(sx.Nil().Cons(data).Cons(sxhtml.SymNoEscape)).Cons(SymP)
	default:
		imgAttr := sx.Nil().Cons(sx.Cons(SymAttrSrc, sx.MakeString("data:image/"+syntax.GetValue()+";base64,"+data.GetValue())))
		var sb strings.Builder
		flattenText(&sb, description)
		if d := sb.String(); d != "" {
			imgAttr = imgAttr.Cons(sx.Cons(symAttrAlt, sx.MakeString(d)))
		}
		return sx.Nil().Cons(sx.Nil().Cons(imgAttr.Cons(sxhtml.SymAttr)).Cons(SymIMG)).Cons(SymP)
	}
}

func flattenText(sb *strings.Builder, lst *sx.Pair) {
	for elem := range lst.Values() {
		switch obj := elem.(type) {
		case sx.String:
			sb.WriteString(obj.GetValue())
		case *sx.Pair:
			flattenText(sb, obj)
		}
	}
}

func (ev *Evaluator) evalList(args sx.Vector, env *Environment) sx.Object {
	return ev.evalSlice(args, env)
}
func nilFn(sx.Vector, *Environment) sx.Object { return sx.Nil() }

// Eval evaluates an object in an environment.
func (ev *Evaluator) Eval(obj sx.Object, env *Environment) sx.Object {
	if env.err != nil {
		return sx.Nil()
	}
	if sx.IsNil(obj) {
		return obj
	}
	lst, isLst := sx.GetPair(obj)
	if !isLst {
		return obj
	}
	sym, found := sx.GetSymbol(lst.Car())
	if !found {
		env.err = fmt.Errorf("symbol expected, but got %T/%v", lst.Car(), lst.Car())
		return sx.Nil()
	}
	symVal := sym.GetValue()
	fn, found := ev.fns[symVal]
	if !found {
		env.err = fmt.Errorf("symbol %q not bound", sym)
		return sx.Nil()
	}
	var args sx.Vector
	for cdr := lst.Cdr(); !sx.IsNil(cdr); {
		pair, isPair := sx.GetPair(cdr)
		if !isPair {
			break
		}
		args = append(args, pair.Car())
		cdr = pair.Cdr()
	}
	if minArgs, hasMinArgs := ev.minArgs[symVal]; hasMinArgs {
		if minArgs > len(args) {
			env.err = fmt.Errorf("%v needs at least %d arguments, but got only %d", sym, minArgs, len(args))
			return sx.Nil()
		}
	}
	result := fn(args, env)
	if env.err != nil {
		return sx.Nil()
	}
	return result
}

func (ev *Evaluator) evalSlice(args sx.Vector, env *Environment) *sx.Pair {
	var result sx.ListBuilder
	for _, arg := range args {
		elem := ev.Eval(arg, env)
		result.Add(elem)
	}
	if env.err == nil {
		return result.List()
	}
	return nil
}

// EvalPairList evaluates a list of lists.
func (ev *Evaluator) EvalPairList(pair *sx.Pair, env *Environment) *sx.Pair {
	var result sx.ListBuilder
	for obj := range pair.Values() {
		elem := ev.Eval(obj, env)
		result.Add(elem)
	}
	if env.err == nil {
		return result.List()
	}
	return nil
}

func (ev *Evaluator) evalLink(a attrs.Attributes, refValue string, inline sx.Vector, env *Environment) sx.Object {
	result := ev.evalSlice(inline, env)
	if len(inline) == 0 {
		result = sx.Nil().Cons(sx.MakeString(refValue))
	}
	if ev.noLinks {
		return result.Cons(SymSPAN)
	}
	return result.Cons(EvaluateAttributes(a)).Cons(SymA)
}

func getSymbol(obj sx.Object, env *Environment) *sx.Symbol {
	if env.err == nil {
		if sym, ok := sx.GetSymbol(obj); ok {
			return sym
		}
		env.err = fmt.Errorf("%v/%T is not a symbol", obj, obj)
	}
	return sx.MakeSymbol("???")
}
func getString(val sx.Object, env *Environment) sx.String {
	if env.err == nil {
		if s, ok := sx.GetString(val); ok {
			return s
		}
		env.err = fmt.Errorf("%v/%T is not a string", val, val)
	}
	return sx.MakeString("")
}
func getList(val sx.Object, env *Environment) *sx.Pair {
	if env.err == nil {
		if res, isPair := sx.GetPair(val); isPair {
			return res
		}
		env.err = fmt.Errorf("%v/%T is not a list", val, val)
	}
	return nil
}
func getInt64(val sx.Object, env *Environment) int64 {
	if env.err != nil {
		return -1017
	}
	if num, ok := sx.GetNumber(val); ok {
		return int64(num.(sx.Int64))
	}
	env.err = fmt.Errorf("%v/%T is not a number", val, val)
	return -1017
}

// GetAttributes evaluates the given arg in the given environment and returns
// the contained attributes.
func GetAttributes(arg sx.Object, env *Environment) attrs.Attributes {
	return sz.GetAttributes(getList(arg, env))
}

// GetReference returns the reference symbol and the reference value of a reference pair.
func GetReference(val sx.Object, env *Environment) (*sx.Symbol, string) {
	if env.err == nil {
		if p := getList(val, env); env.err == nil {
			sym, val := sz.GetReference(p)
			if sym != nil {
				return sym, val
			}
			env.err = fmt.Errorf("%v/%T is not a reference", val, val)
		}
	}
	return nil, ""
}

var unsafeSnippets = []string{
	"<script", "</script",
	"<iframe", "</iframe",
}

// IsSafe returns true if the given string does not contain unsafe HTML elements.
func IsSafe(s string) bool {
	lower := strings.ToLower(s)
	for _, snippet := range unsafeSnippets {
		if strings.Contains(lower, snippet) {
			return false
		}
	}
	return true
}
