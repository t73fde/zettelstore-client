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

	"zettelstore.de/client.fossil/api"
	"zettelstore.de/client.fossil/attrs"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/client.fossil/text"
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxhtml"
)

// Evaluator will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Evaluator struct {
	headingOffset int64
	unique        string
	noLinks       bool // true iff output must not include links

	fns     map[sx.Symbol]EvalFn
	minArgs map[sx.Symbol]int
}

// NewEvaluator creates a new Evaluator object.
func NewEvaluator(headingOffset int) *Evaluator {
	ev := &Evaluator{
		headingOffset: int64(headingOffset),

		fns:     make(map[sx.Symbol]EvalFn, 128),
		minArgs: make(map[sx.Symbol]int, 128),
	}
	ev.bindMetadata()
	ev.bindBlocks()
	ev.bindInlines()
	return ev
}

// SetUnique sets a prefix to make several HTML ids unique.
func (tr *Evaluator) SetUnique(s string) { tr.unique = s }

// IsValidName returns true, if name is a valid symbol name.
func (tr *Evaluator) IsValidName(s string) bool { return s != "" }

// EvaluateAttrbute transforms the given attributes into a HTML s-expression.
func (tr *Evaluator) EvaluateAttrbute(a attrs.Attributes) *sx.Pair {
	if len(a) == 0 {
		return nil
	}
	plist := sx.Nil()
	keys := a.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		if key != attrs.DefaultAttribute && tr.IsValidName(key) {
			plist = plist.Cons(sx.Cons(sx.Symbol(key), sx.String(a[key])))
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
		// May extend tr.endnotes

		if env.endnotes[i].noteHx != nil {
			continue
		}

		noteAST := env.endnotes[i].noteAST
		noteHx := sx.Nil()
		curr := noteHx
		for _, inline := range noteAST {
			inl := ev.Eval(inline, env)
			if noteHx == nil {
				noteHx = sx.Cons(inl, nil)
				curr = noteHx
			} else {
				curr = curr.AppendBang(inl)
			}
		}
		if env.err != nil {
			break
		}
		env.endnotes[i].noteHx = noteHx
	}

	return pair, nil
}

// Endnotes returns a SHTML object with all collected endnotes.
func (ev *Evaluator) Endnotes(env *Environment) *sx.Pair {
	if env.err != nil || len(env.endnotes) == 0 {
		return nil
	}

	result := sx.Nil().Cons(SymOL)

	currResult := result.AppendBang(sx.Nil().Cons(sx.Cons(SymAttrClass, sx.String("zs-endnotes"))).Cons(sxhtml.SymAttr))
	for i, fni := range env.endnotes {
		noteNum := strconv.Itoa(i + 1)
		attrs := fni.attrs.Cons(sx.Cons(SymAttrClass, sx.String("zs-endnote"))).
			Cons(sx.Cons(SymAttrValue, sx.String(noteNum))).
			Cons(sx.Cons(SymAttrId, sx.String("fn:"+fni.noteID))).
			Cons(sx.Cons(SymAttrRole, sx.String("doc-endnote"))).
			Cons(sxhtml.SymAttr)

		backref := sx.Nil().Cons(sx.String("\u21a9\ufe0e")).
			Cons(sx.Nil().
				Cons(sx.Cons(SymAttrClass, sx.String("zs-endnote-backref"))).
				Cons(sx.Cons(SymAttrHref, sx.String("#fnref:"+fni.noteID))).
				Cons(sx.Cons(SymAttrRole, sx.String("doc-backlink"))).
				Cons(sxhtml.SymAttr)).
			Cons(SymA)

		li := sx.Nil().Cons(SymLI)
		li.AppendBang(attrs).
			ExtendBang(fni.noteHx).
			AppendBang(sx.String(" ")).AppendBang(backref)
		currResult = currResult.AppendBang(li)
	}
	return result
}

// Environment where sz objects are evaluated to shtml objects
type Environment struct {
	err          error
	langStack    []string
	endnotes     []endnoteInfo
	quoteNesting uint
}
type endnoteInfo struct {
	noteID  string      // link id
	noteAST []sx.Object // Endnote as list of AST inline elements
	attrs   *sx.Pair    // attrs a-list
	noteHx  *sx.Pair    // Endnote as SxHTML
}

// MakeEnvironment builds a new evaluation environment.
func MakeEnvironment(lang string) Environment {
	langStack := make([]string, 1, 16)
	langStack[0] = lang
	return Environment{
		err:          nil,
		langStack:    langStack,
		endnotes:     nil,
		quoteNesting: 0,
	}
}

// GetError returns the last error found.
func (env *Environment) GetError() error { return env.err }

// Reset the environment.
func (env *Environment) Reset() {
	env.langStack = env.langStack[0:1]
	env.endnotes = nil
	env.quoteNesting = 0
}

// PushAttribute adds the current attributes to the environment.
func (env *Environment) pushAttributes(a attrs.Attributes) {
	if value, ok := a.Get("lang"); ok {
		env.langStack = append(env.langStack, value)
	} else {
		env.langStack = append(env.langStack, env.getLanguage())
	}
}

// popAttributes removes the current attributes from the envrionment
func (env *Environment) popAttributes() {
	env.langStack = env.langStack[0 : len(env.langStack)-1]
}

// getLanguage returns the current language
func (env *Environment) getLanguage() string {
	return env.langStack[len(env.langStack)-1]
}

// EvalFn is a function to be called for evaluation.
type EvalFn func([]sx.Object, *Environment) sx.Object

func (ev *Evaluator) bind(sym sx.Symbol, minArgs int, fn EvalFn) {
	ev.fns[sym] = fn
	if minArgs > 0 {
		ev.minArgs[sym] = minArgs
	}
}

// ResolveBinding returns the function bound to the given name.
func (ev *Evaluator) ResolveBinding(sym sx.Symbol) EvalFn {
	if fn, found := ev.fns[sym]; found {
		return fn
	}
	return nil
}

// Rebind overwrites a binding, but leaves the minimum number of arguments intact.
func (ev *Evaluator) Rebind(sym sx.Symbol, fn EvalFn) {
	if _, found := ev.fns[sym]; !found {
		panic(sym)
	}
	ev.fns[sym] = fn
}

func (ev *Evaluator) bindMetadata() {
	ev.bind(sz.SymMeta, 0, ev.evalList)
	evalMetaString := func(args []sx.Object, env *Environment) sx.Object {
		a := make(attrs.Attributes, 2).
			Set("name", string(ev.getSymbol(args[0], env))).
			Set("content", string(getString(args[1], env)))
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

	evalMetaSet := func(args []sx.Object, env *Environment) sx.Object {
		var sb strings.Builder
		for elem := getList(args[1], env); elem != nil; elem = elem.Tail() {
			sb.WriteByte(' ')
			sb.WriteString(string(getString(elem.Car(), env)))
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", string(ev.getSymbol(args[0], env))).
			Set("content", s)
		return ev.EvaluateMeta(a)
	}
	ev.bind(sz.SymTypeIDSet, 2, evalMetaSet)
	ev.bind(sz.SymTypeTagSet, 2, evalMetaSet)
	ev.bind(sz.SymTypeWordSet, 2, evalMetaSet)
	ev.bind(sz.SymTypeZettelmarkup, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := make(attrs.Attributes, 2).
			Set("name", string(ev.getSymbol(args[0], env))).
			Set("content", text.EvaluateInlineString(getList(args[1], env)))
		return ev.EvaluateMeta(a)
	})
}

// EvaluateMeta returns HTML meta object for an attribute.
func (ev *Evaluator) EvaluateMeta(a attrs.Attributes) *sx.Pair {
	return sx.Nil().Cons(ev.EvaluateAttrbute(a)).Cons(SymMeta)
}

func (ev *Evaluator) bindBlocks() {
	ev.bind(sz.SymBlock, 0, ev.evalList)
	ev.bind(sz.SymPara, 0, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalSlice(args, env).Cons(SymP)
	})
	ev.bind(sz.SymHeading, 5, func(args []sx.Object, env *Environment) sx.Object {
		nLevel := getInt64(args[0], env)
		if nLevel <= 0 {
			env.err = fmt.Errorf("%v is a negative level", nLevel)
			return sx.Nil()
		}
		level := strconv.FormatInt(nLevel+ev.headingOffset, 10)
		headingSymbol := sx.Symbol("h" + level)

		a := ev.GetAttributes(args[1], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if fragment := string(getString(args[3], env)); fragment != "" {
			a = a.Set("id", ev.unique+fragment)
		}

		if result, isPair := sx.GetPair(ev.Eval(args[4], env)); isPair && result != nil {
			if len(a) > 0 {
				result = result.Cons(ev.EvaluateAttrbute(a))
			}
			return result.Cons(headingSymbol)
		}
		return sx.MakeList(headingSymbol, sx.String("<MISSING TEXT>"))
	})
	ev.bind(sz.SymThematic, 0, func(args []sx.Object, env *Environment) sx.Object {
		result := sx.Nil()
		if len(args) > 0 {
			if attrList := getList(args[0], env); attrList != nil {
				result = result.Cons(ev.EvaluateAttrbute(sz.GetAttributes(attrList)))
			}
		}
		return result.Cons(SymHR)
	})

	ev.bind(sz.SymListOrdered, 0, ev.makeListFn(SymOL))
	ev.bind(sz.SymListUnordered, 0, ev.makeListFn(SymUL))
	ev.bind(sz.SymDescription, 0, func(args []sx.Object, env *Environment) sx.Object {
		if len(args) == 0 {
			return sx.Nil()
		}
		items := sx.Nil().Cons(symDL)
		curItem := items
		for pos := 0; pos < len(args); pos++ {
			term := getList(ev.Eval(args[pos], env), env)
			curItem = curItem.AppendBang(term.Cons(symDT))
			pos++
			if pos >= len(args) {
				break
			}
			ddBlock := getList(ev.Eval(args[pos], env), env)
			if ddBlock == nil {
				continue
			}
			for ddlst := ddBlock; ddlst != nil; ddlst = ddlst.Tail() {
				dditem := getList(ddlst.Car(), env)
				curItem = curItem.AppendBang(dditem.Cons(symDD))
			}
		}
		return items
	})
	ev.bind(sz.SymListQuote, 0, func(args []sx.Object, env *Environment) sx.Object {
		if args == nil {
			return sx.Nil()
		}
		result := sx.Nil().Cons(symBLOCKQUOTE)
		currResult := result
		for _, elem := range args {
			if quote, isPair := sx.GetPair(ev.Eval(elem, env)); isPair {
				currResult = currResult.AppendBang(quote.Cons(sxhtml.SymListSplice))
			}
		}
		return result
	})

	ev.bind(sz.SymTable, 1, func(args []sx.Object, env *Environment) sx.Object {
		thead := sx.Nil()
		if header := getList(args[0], env); !sx.IsNil(header) {
			thead = sx.Nil().Cons(ev.evalTableRow(header, env)).Cons(symTHEAD)
		}

		tbody := sx.Nil()
		if len(args) > 1 {
			tbody = sx.Nil().Cons(symTBODY)
			curBody := tbody
			for _, row := range args[1:] {
				curBody = curBody.AppendBang(ev.evalTableRow(getList(row, env), env))
			}
		}

		table := sx.Nil()
		if tbody != nil {
			table = table.Cons(tbody)
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

	ev.bind(sz.SymVerbatimComment, 1, func(args []sx.Object, env *Environment) sx.Object {
		if ev.GetAttributes(args[0], env).HasDefault() {
			if len(args) > 1 {
				if s := getString(args[1], env); s != "" {
					return sx.Nil().Cons(s).Cons(sxhtml.SymBlockComment)
				}
			}
		}
		return nil
	})
	ev.bind(sz.SymVerbatimEval, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalVerbatim(ev.GetAttributes(args[0], env).AddClass("zs-eval"), getString(args[1], env))
	})
	ev.bind(sz.SymVerbatimHTML, 2, ev.evalHTML)
	ev.bind(sz.SymVerbatimMath, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalVerbatim(ev.GetAttributes(args[0], env).AddClass("zs-math"), getString(args[1], env))
	})
	ev.bind(sz.SymVerbatimProg, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		content := getString(args[1], env)
		if a.HasDefault() {
			content = sx.String(visibleReplacer.Replace(string(content)))
		}
		return ev.evalVerbatim(a, content)
	})
	ev.bind(sz.SymVerbatimZettel, 0, nilFn)
	ev.bind(sz.SymBLOB, 3, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalBLOB(getList(args[0], env), getString(args[1], env), getString(args[2], env))
	})
	ev.bind(sz.SymTransclude, 2, func(args []sx.Object, env *Environment) sx.Object {
		ref, isPair := sx.GetPair(args[1])
		if !isPair {
			return sx.Nil()
		}
		refKind := ref.Car()
		if sx.IsNil(refKind) {
			return sx.Nil()
		}
		if refValue := getString(ref.Tail().Car(), env); refValue != "" {
			if refSym, isRefSym := sx.GetSymbol(refKind); isRefSym && refSym.IsEqual(sz.SymRefStateExternal) {
				a := ev.GetAttributes(args[0], env).Set("src", string(refValue)).AddClass("external")
				return sx.Nil().Cons(sx.Nil().Cons(ev.EvaluateAttrbute(a)).Cons(SymIMG)).Cons(SymP)
			}
			return sx.MakeList(
				sxhtml.SymInlineComment,
				sx.String("transclude"),
				refKind,
				sx.String("->"),
				refValue,
			)
		}
		return ev.evalSlice(args, env)
	})
}

func (ev *Evaluator) makeListFn(sym sx.Symbol) EvalFn {
	return func(args []sx.Object, env *Environment) sx.Object {
		result := sx.Nil().Cons(sym)
		last := result
		for _, elem := range args {
			item := sx.Nil().Cons(SymLI)
			if res, isPair := sx.GetPair(ev.Eval(elem, env)); isPair {
				item.ExtendBang(res)
			}
			last = last.AppendBang(item)
		}
		return result
	}
}

func (ev *Evaluator) evalTableRow(pairs *sx.Pair, env *Environment) *sx.Pair {
	row := sx.Nil().Cons(symTR)
	if pairs == nil {
		return nil
	}
	curRow := row
	for pair := pairs; pair != nil; pair = pair.Tail() {
		curRow = curRow.AppendBang(ev.Eval(pair.Car(), env))
	}
	return row
}
func (ev *Evaluator) makeCellFn(align string) EvalFn {
	return func(args []sx.Object, env *Environment) sx.Object {
		tdata := ev.evalSlice(args, env)
		if align != "" {
			tdata = tdata.Cons(ev.EvaluateAttrbute(attrs.Attributes{"class": align}))
		}
		return tdata.Cons(symTD)
	}
}

func (ev *Evaluator) makeRegionFn(sym sx.Symbol, genericToClass bool) EvalFn {
	return func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if genericToClass {
			if val, found := a.Get(""); found {
				a = a.Remove("").AddClass(val)
			}
		}
		result := sx.Nil()
		if len(a) > 0 {
			result = result.Cons(ev.EvaluateAttrbute(a))
		}
		result = result.Cons(sym)
		currResult := result.LastPair()
		if region, isPair := sx.GetPair(ev.Eval(args[1], env)); isPair {
			currResult = currResult.ExtendBang(region)
		}
		if len(args) > 2 {
			if cite, isPair := sx.GetPair(ev.Eval(args[2], env)); isPair && cite != nil {
				currResult.AppendBang(cite.Cons(symCITE))
			}
		}
		return result
	}
}

func (ev *Evaluator) evalVerbatim(a attrs.Attributes, s sx.String) sx.Object {
	a = setProgLang(a)
	code := sx.Nil().Cons(s)
	if al := ev.EvaluateAttrbute(a); al != nil {
		code = code.Cons(al)
	}
	code = code.Cons(symCODE)
	return sx.Nil().Cons(code).Cons(symPRE)
}

func (ev *Evaluator) bindInlines() {
	ev.bind(sz.SymInline, 0, ev.evalList)
	ev.bind(sz.SymText, 1, func(args []sx.Object, env *Environment) sx.Object { return getString(args[0], env) })
	ev.bind(sz.SymSpace, 0, func(args []sx.Object, env *Environment) sx.Object {
		if len(args) == 0 {
			return sx.String(" ")
		}
		return getString(args[0], env)
	})
	ev.bind(sz.SymSoft, 0, func([]sx.Object, *Environment) sx.Object { return sx.String(" ") })
	ev.bind(sz.SymHard, 0, func([]sx.Object, *Environment) sx.Object { return sx.Nil().Cons(symBR) })

	ev.bind(sz.SymLinkInvalid, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		var inline *sx.Pair
		if len(args) > 2 {
			inline = ev.evalSlice(args[2:], env)
		}
		if inline == nil {
			inline = sx.Nil().Cons(ev.Eval(args[1], env))
		}
		return inline.Cons(SymSPAN)
	})
	evalHREF := func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		return ev.evalLink(a.Set("href", string(refValue)), refValue, args[2:], env)
	}
	ev.bind(sz.SymLinkZettel, 2, evalHREF)
	ev.bind(sz.SymLinkSelf, 2, evalHREF)
	ev.bind(sz.SymLinkFound, 2, evalHREF)
	ev.bind(sz.SymLinkBroken, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		return ev.evalLink(a.AddClass("broken"), refValue, args[2:], env)
	})
	ev.bind(sz.SymLinkHosted, 2, evalHREF)
	ev.bind(sz.SymLinkBased, 2, evalHREF)
	ev.bind(sz.SymLinkQuery, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(string(refValue))
		return ev.evalLink(a.Set("href", query), refValue, args[2:], env)
	})
	ev.bind(sz.SymLinkExternal, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		return ev.evalLink(a.Set("href", string(refValue)).AddClass("external"), refValue, args[2:], env)
	})

	ev.bind(sz.SymEmbed, 3, func(args []sx.Object, env *Environment) sx.Object {
		ref := getList(args[1], env)
		syntax := getString(args[2], env)
		if syntax == api.ValueSyntaxSVG {
			embedAttr := sx.MakeList(
				sxhtml.SymAttr,
				sx.Cons(SymAttrType, sx.String("image/svg+xml")),
				sx.Cons(SymAttrSrc, sx.String("/"+string(getString(ref.Tail(), env))+".svg")),
			)
			return sx.MakeList(
				SymFIGURE,
				sx.MakeList(
					SymEMBED,
					embedAttr,
				),
			)
		}
		a := ev.GetAttributes(args[0], env)
		a = a.Set("src", string(getString(ref.Tail().Car(), env)))
		if len(args) > 3 {
			var sb strings.Builder
			flattenText(&sb, sx.MakeList(args[3:]...))
			if d := sb.String(); d != "" {
				a = a.Set("alt", d)
			}
		}
		return sx.MakeList(SymIMG, ev.EvaluateAttrbute(a))
	})
	ev.bind(sz.SymEmbedBLOB, 3, func(args []sx.Object, env *Environment) sx.Object {
		a, syntax, data := ev.GetAttributes(args[0], env), getString(args[1], env), getString(args[2], env)
		summary, hasSummary := a.Get(api.KeySummary)
		if !hasSummary {
			summary = ""
		}
		return ev.evalBLOB(
			sx.MakeList(sxhtml.SymListSplice, sx.String(summary)),
			syntax,
			data,
		)
	})

	ev.bind(sz.SymCite, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		result := sx.Nil()
		if key := getString(args[1], env); key != "" {
			if len(args) > 2 {
				result = ev.evalSlice(args[2:], env).Cons(sx.String(", "))
			}
			result = result.Cons(key)
		}
		if len(a) > 0 {
			result = result.Cons(ev.EvaluateAttrbute(a))
		}
		if result == nil {
			return nil
		}
		return result.Cons(SymSPAN)
	})
	ev.bind(sz.SymMark, 3, func(args []sx.Object, env *Environment) sx.Object {
		result := ev.evalSlice(args[3:], env)
		if !ev.noLinks {
			if fragment := getString(args[2], env); fragment != "" {
				a := attrs.Attributes{"id": string(fragment) + ev.unique}
				return result.Cons(ev.EvaluateAttrbute(a)).Cons(SymA)
			}
		}
		return result.Cons(SymSPAN)
	})
	ev.bind(sz.SymEndnote, 1, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		attrPlist := sx.Nil()
		if len(a) > 0 {
			if attrs := ev.EvaluateAttrbute(a); attrs != nil {
				attrPlist = attrs.Tail()
			}
		}

		noteNum := strconv.Itoa(len(env.endnotes) + 1)
		noteID := ev.unique + noteNum
		env.endnotes = append(env.endnotes, endnoteInfo{
			noteID: noteID, noteAST: args[1:], noteHx: nil, attrs: attrPlist})
		hrefAttr := sx.Nil().Cons(sx.Cons(SymAttrRole, sx.String("doc-noteref"))).
			Cons(sx.Cons(SymAttrHref, sx.String("#fn:"+noteID))).
			Cons(sx.Cons(SymAttrClass, sx.String("zs-noteref"))).
			Cons(sxhtml.SymAttr)
		href := sx.Nil().Cons(sx.String(noteNum)).Cons(hrefAttr).Cons(SymA)
		supAttr := sx.Nil().Cons(sx.Cons(SymAttrId, sx.String("fnref:"+noteID))).Cons(sxhtml.SymAttr)
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

	ev.bind(sz.SymLiteralComment, 1, func(args []sx.Object, env *Environment) sx.Object {
		if ev.GetAttributes(args[0], env).HasDefault() {
			if len(args) > 1 {
				if s := getString(ev.Eval(args[1], env), env); s != "" {
					return sx.Nil().Cons(s).Cons(sxhtml.SymInlineComment)
				}
			}
		}
		return sx.Nil()
	})
	ev.bind(sz.SymLiteralHTML, 2, ev.evalHTML)
	ev.bind(sz.SymLiteralInput, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalLiteral(args, nil, symKBD, env)
	})
	ev.bind(sz.SymLiteralMath, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env).AddClass("zs-math")
		return ev.evalLiteral(args, a, symCODE, env)
	})
	ev.bind(sz.SymLiteralOutput, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalLiteral(args, nil, symSAMP, env)
	})
	ev.bind(sz.SymLiteralProg, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalLiteral(args, nil, symCODE, env)
	})

	ev.bind(sz.SymLiteralZettel, 0, nilFn)
}

func (ev *Evaluator) makeFormatFn(sym sx.Symbol) EvalFn {
	return func(args []sx.Object, env *Environment) sx.Object {
		a := ev.GetAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if val, hasClass := a.Get(""); hasClass {
			a = a.Remove("").AddClass(val)
		}
		res := ev.evalSlice(args[1:], env)
		if len(a) > 0 {
			res = res.Cons(ev.EvaluateAttrbute(a))
		}
		return res.Cons(sym)
	}
}

type quoteData struct {
	primLeft, primRight string
	secLeft, secRight   string
	nbsp                bool
}

var langQuotes = map[string]quoteData{
	"":              {"&quot;", "&quot;", "&quot;", "&quot;", false},
	api.ValueLangEN: {"&ldquo;", "&rdquo;", "&lsquo;", "&rsquo;", false},
	"de":            {"&bdquo;", "&ldquo;", "&sbquo;", "&lsquo;", false},
	"fr":            {"&laquo;", "&raquo;", "&lsaquo;", "&rsaquo;", true},
}

func getQuoteData(lang string) quoteData {
	langFields := strings.FieldsFunc(lang, func(r rune) bool { return r == '-' || r == '_' })
	for len(langFields) > 0 {
		langSup := strings.Join(langFields, "-")
		quotes, ok := langQuotes[langSup]
		if ok {
			return quotes
		}
		langFields = langFields[0 : len(langFields)-1]
	}
	return langQuotes[""]
}

func getQuotes(data *quoteData, env *Environment) (string, string) {
	if env.quoteNesting%2 == 0 {
		return data.primLeft, data.primRight
	}
	return data.secLeft, data.secRight
}

func (ev *Evaluator) evalQuote(args []sx.Object, env *Environment) sx.Object {
	a := ev.GetAttributes(args[0], env)
	env.pushAttributes(a)
	defer env.popAttributes()

	if val, hasClass := a.Get(""); hasClass {
		a = a.Remove("").AddClass(val)
	}
	quotes := getQuoteData(env.getLanguage())
	leftQ, rightQ := getQuotes(&quotes, env)

	env.quoteNesting++
	res := ev.evalSlice(args[1:], env)
	env.quoteNesting--

	lastPair := res.LastPair()
	if lastPair.IsNil() {
		res = sx.Cons(sx.MakeList(sxhtml.SymNoEscape, sx.String(leftQ), sx.String(rightQ)), sx.Nil())
	} else {
		if quotes.nbsp {
			lastPair.AppendBang(sx.MakeList(sxhtml.SymNoEscape, sx.String("&nbsp;"), sx.String(rightQ)))
			res = res.Cons(sx.MakeList(sxhtml.SymNoEscape, sx.String(leftQ), sx.String("&nbsp;")))
		} else {
			lastPair.AppendBang(sx.MakeList(sxhtml.SymNoEscape, sx.String(rightQ)))
			res = res.Cons(sx.MakeList(sxhtml.SymNoEscape, sx.String(leftQ)))
		}
	}
	if len(a) > 0 {
		res = res.Cons(ev.EvaluateAttrbute(a))
		return res.Cons(SymSPAN)
	}
	return res.Cons(sxhtml.SymListSplice)
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func (ev *Evaluator) evalLiteral(args []sx.Object, a attrs.Attributes, sym sx.Symbol, env *Environment) sx.Object {
	if a == nil {
		a = ev.GetAttributes(args[0], env)
	}
	a = setProgLang(a)
	literal := string(getString(args[1], env))
	if a.HasDefault() {
		a = a.RemoveDefault()
		literal = visibleReplacer.Replace(literal)
	}
	res := sx.Nil().Cons(sx.String(literal))
	if len(a) > 0 {
		res = res.Cons(ev.EvaluateAttrbute(a))
	}
	return res.Cons(sym)
}
func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}

func (ev *Evaluator) evalHTML(args []sx.Object, env *Environment) sx.Object {
	if s := getString(ev.Eval(args[1], env), env); s != "" && IsSafe(string(s)) {
		return sx.Nil().Cons(s).Cons(sxhtml.SymNoEscape)
	}
	return nil
}

func (ev *Evaluator) evalBLOB(description *sx.Pair, syntax, data sx.String) sx.Object {
	if data == "" {
		return sx.Nil()
	}
	switch syntax {
	case "":
		return sx.Nil()
	case api.ValueSyntaxSVG:
		return sx.Nil().Cons(sx.Nil().Cons(data).Cons(sxhtml.SymNoEscape)).Cons(SymP)
	default:
		imgAttr := sx.Nil().Cons(sx.Cons(SymAttrSrc, sx.String("data:image/"+string(syntax)+";base64,"+string(data))))
		var sb strings.Builder
		flattenText(&sb, description)
		if d := sb.String(); d != "" {
			imgAttr = imgAttr.Cons(sx.Cons(symAttrAlt, sx.String(d)))
		}
		return sx.Nil().Cons(sx.Nil().Cons(imgAttr.Cons(sxhtml.SymAttr)).Cons(SymIMG)).Cons(SymP)
	}
}

func flattenText(sb *strings.Builder, lst *sx.Pair) {
	for elem := lst; elem != nil; elem = elem.Tail() {
		switch obj := elem.Car().(type) {
		case sx.String:
			sb.WriteString(string(obj))
		case sx.Symbol:
			if obj.IsEqual(sz.SymSpace) {
				sb.WriteByte(' ')
				break
			}
		case *sx.Pair:
			flattenText(sb, obj)
		}
	}
}

func (ev *Evaluator) evalList(args []sx.Object, env *Environment) sx.Object {
	return ev.evalSlice(args, env)
}
func nilFn([]sx.Object, *Environment) sx.Object { return sx.Nil() }

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
	fn, found := ev.fns[sym]
	if !found {
		env.err = fmt.Errorf("symbol %q not bound", sym)
		return sx.Nil()
	}
	var args []sx.Object
	for cdr := lst.Cdr(); !sx.IsNil(cdr); {
		pair, isPair := sx.GetPair(cdr)
		if !isPair {
			break
		}
		args = append(args, pair.Car())
		cdr = pair.Cdr()
	}
	if minArgs, hasMinArgs := ev.minArgs[sym]; hasMinArgs {
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

func (ev *Evaluator) evalSlice(args []sx.Object, env *Environment) *sx.Pair {
	result := sx.Cons(sx.Nil(), sx.Nil())
	curr := result
	for _, arg := range args {
		elem := ev.Eval(arg, env)
		if env.err != nil {
			return nil
		}
		curr = curr.AppendBang(elem)
	}
	return result.Tail()
}

func (ev *Evaluator) evalLink(a attrs.Attributes, refValue sx.String, inline []sx.Object, env *Environment) sx.Object {
	result := ev.evalSlice(inline, env)
	if len(inline) == 0 {
		result = sx.Nil().Cons(refValue)
	}
	if ev.noLinks {
		return result.Cons(SymSPAN)
	}
	return result.Cons(ev.EvaluateAttrbute(a)).Cons(SymA)
}

func (ev *Evaluator) getSymbol(obj sx.Object, env *Environment) sx.Symbol {
	if env.err == nil {
		if sym, ok := sx.GetSymbol(obj); ok {
			return sym
		}
		env.err = fmt.Errorf("%v/%T is not a symbol", obj, obj)
	}
	return sx.Symbol("???")
}
func getString(val sx.Object, env *Environment) sx.String {
	if env.err != nil {
		return ""
	}
	if s, ok := sx.GetString(val); ok {
		return s
	}
	env.err = fmt.Errorf("%v/%T is not a string", val, val)
	return ""
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
func (ev *Evaluator) GetAttributes(arg sx.Object, env *Environment) attrs.Attributes {
	return sz.GetAttributes(getList(arg, env))
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
