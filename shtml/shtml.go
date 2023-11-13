//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
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
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxhtml"
)

// Evaluator will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Evaluator struct {
	sf            sx.SymbolFactory
	headingOffset int64
	unique        string
	noLinks       bool // true iff output must not include links

	symAttr     *sx.Symbol
	symList     *sx.Symbol
	symNoEscape *sx.Symbol
	symClass    *sx.Symbol
	symMeta     *sx.Symbol
	symP        *sx.Symbol
	symLI       *sx.Symbol
	symA        *sx.Symbol
	symHREF     *sx.Symbol
	symSpan     *sx.Symbol

	fns     map[string]EvalFn
	minArgs map[string]int
}

// NewEvaluator creates a new Evaluator object.
func NewEvaluator(headingOffset int, sf sx.SymbolFactory) *Evaluator {
	if sf == nil {
		sf = sx.MakeMappedFactory(128)
	}
	ev := &Evaluator{
		sf:            sf,
		headingOffset: int64(headingOffset),
		symAttr:       sf.MustMake(sxhtml.NameSymAttr),
		symList:       sf.MustMake(sxhtml.NameSymList),
		symNoEscape:   sf.MustMake(sxhtml.NameSymNoEscape),
		symClass:      sf.MustMake("class"),
		symMeta:       sf.MustMake("meta"),
		symP:          sf.MustMake("p"),
		symLI:         sf.MustMake("li"),
		symA:          sf.MustMake("a"),
		symHREF:       sf.MustMake("href"),
		symSpan:       sf.MustMake("span"),

		fns:     make(map[string]EvalFn, 128),
		minArgs: make(map[string]int, 128),
	}
	ev.bindCommon()
	ev.bindMetadata()
	ev.bindBlocks()
	ev.bindInlines()
	return ev
}

// SetUnique sets a prefix to make several HTML ids unique.
func (tr *Evaluator) SetUnique(s string) { tr.unique = s }

// SymbolFactory returns the symbol factory of this evaluator.
func (ev *Evaluator) SymbolFactory() sx.SymbolFactory { return ev.sf }

// IsValidName returns true, if name is a valid symbol name.
func (tr *Evaluator) IsValidName(s string) bool { return tr.sf.IsValidName(s) }

// Make a new HTML symbol.
func (tr *Evaluator) Make(s string) *sx.Symbol { return tr.sf.MustMake(s) }

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
			plist = plist.Cons(sx.Cons(tr.Make(key), sx.String(a[key])))
		}
	}
	if plist == nil {
		return nil
	}
	return plist.Cons(tr.symAttr)
}

// Evaluate a metadata s-expression into a list of HTML s-expressions.
func (ev *Evaluator) Evaluate(lst *sx.Pair, env *Environment) (*sx.Pair, error) {
	result := ev.eval(lst, env)
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

		objHx := ev.eval(env.endnotes[i].noteAST, env)
		if env.err != nil {
			break
		}
		noteHx, isHx := sx.GetPair(objHx)
		if !isHx {
			return nil, fmt.Errorf("endnote evaluation does not result in pair, but %T/%v", objHx, objHx)
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

	result := sx.Nil().Cons(ev.Make("ol"))
	symValue, symId, symRole := ev.Make("value"), ev.Make("id"), ev.Make("role")

	currResult := result.AppendBang(sx.Nil().Cons(sx.Cons(ev.symClass, sx.String("zs-endnotes"))).Cons(ev.symAttr))
	for i, fni := range env.endnotes {
		noteNum := strconv.Itoa(i + 1)
		attrs := fni.attrs.Cons(sx.Cons(ev.symClass, sx.String("zs-endnote"))).
			Cons(sx.Cons(symValue, sx.String(noteNum))).
			Cons(sx.Cons(symId, sx.String("fn:"+fni.noteID))).
			Cons(sx.Cons(symRole, sx.String("doc-endnote"))).
			Cons(ev.symAttr)

		backref := sx.Nil().Cons(sx.String("\u21a9\ufe0e")).
			Cons(sx.Nil().
				Cons(sx.Cons(ev.symClass, sx.String("zs-endnote-backref"))).
				Cons(sx.Cons(ev.symHREF, sx.String("#fnref:"+fni.noteID))).
				Cons(sx.Cons(symRole, sx.String("doc-backlink"))).
				Cons(ev.symAttr)).
			Cons(ev.symA)

		li := sx.Nil().Cons(ev.symLI)
		li.AppendBang(attrs).
			ExtendBang(fni.noteHx).
			AppendBang(sx.String(" ")).AppendBang(backref)
		currResult = currResult.AppendBang(li)
	}
	return result
}

// Environment where sz objects are evaluated to shtml objects
type Environment struct {
	err             error
	langStack       []string
	endnotes        []endnoteInfo
	secondaryQuotes bool
}
type endnoteInfo struct {
	noteID  string    // link id
	noteAST sx.Object // Endnote as AST
	attrs   *sx.Pair  // attrs a-list
	noteHx  *sx.Pair  // Endnote as SxHTML
}

// MakeEnvironment builds a new evaluation environment.
func MakeEnvironment(lang string) Environment {
	langStack := make([]string, 1, 16)
	langStack[0] = lang
	return Environment{
		err:             nil,
		langStack:       langStack,
		endnotes:        nil,
		secondaryQuotes: false,
	}
}

// GetError returns the last error found.
func (env *Environment) GetError() error { return env.err }

// Reset the environment.
func (env *Environment) Reset() {
	env.langStack = env.langStack[0:1]
	env.endnotes = nil
	env.secondaryQuotes = false
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

func (ev *Evaluator) bind(name string, minArgs int, fn EvalFn) {
	ev.fns[name] = fn
	if minArgs > 0 {
		ev.minArgs[name] = minArgs
	}
}

// ResolveBinding returns the function bound to the given name.
func (ev *Evaluator) ResolveBinding(name string) EvalFn {
	if fn, found := ev.fns[name]; found {
		return fn
	}
	return nil
}

// Rebind overwrites a binding, but leaves the minimum number of arguments intact.
func (ev *Evaluator) Rebind(name string, fn EvalFn) {
	if _, found := ev.fns[name]; !found {
		panic(name)
	}
	ev.fns[name] = fn
}

func (ev *Evaluator) bindCommon() {
	ev.bind(sx.ListName, 0, ev.evalList)
	ev.bind("quote", 1, func(args []sx.Object, _ *Environment) sx.Object { return args[0] })
}

func (ev *Evaluator) bindMetadata() {
	ev.bind(sz.NameSymMeta, 0, ev.evalList)
	evalMetaString := func(args []sx.Object, env *Environment) sx.Object {
		a := make(attrs.Attributes, 2).
			Set("name", ev.getSymbol(ev.eval(args[0], env), env).Name()).
			Set("content", getString(args[1], env).String())
		return ev.EvaluateMeta(a)
	}
	ev.bind(sz.NameSymTypeCredential, 2, evalMetaString)
	ev.bind(sz.NameSymTypeEmpty, 2, evalMetaString)
	ev.bind(sz.NameSymTypeID, 2, evalMetaString)
	ev.bind(sz.NameSymTypeNumber, 2, evalMetaString)
	ev.bind(sz.NameSymTypeString, 2, evalMetaString)
	ev.bind(sz.NameSymTypeTimestamp, 2, evalMetaString)
	ev.bind(sz.NameSymTypeURL, 2, evalMetaString)
	ev.bind(sz.NameSymTypeWord, 2, evalMetaString)

	evalMetaSet := func(args []sx.Object, env *Environment) sx.Object {
		var sb strings.Builder
		lst := ev.eval(args[1], env)
		for elem := getList(lst, env); elem != nil; elem = elem.Tail() {
			sb.WriteByte(' ')
			sb.WriteString(getString(elem.Car(), env).String())
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", ev.getSymbol(ev.eval(args[0], env), env).Name()).
			Set("content", s)
		return ev.EvaluateMeta(a)
	}
	ev.bind(sz.NameSymTypeIDSet, 2, evalMetaSet)
	ev.bind(sz.NameSymTypeTagSet, 2, evalMetaSet)
	ev.bind(sz.NameSymTypeWordSet, 2, evalMetaSet)
}

// EvaluateMeta returns HTML meta object for an attribute.
func (ev *Evaluator) EvaluateMeta(a attrs.Attributes) *sx.Pair {
	return sx.Nil().Cons(ev.EvaluateAttrbute(a)).Cons(ev.symMeta)
}

func (ev *Evaluator) bindBlocks() {
	ev.bind(sz.NameSymBlock, 0, ev.evalList)
	ev.bind(sz.NameSymPara, 0, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalSlice(args, env).Cons(ev.symP)
	})
	ev.bind(sz.NameSymHeading, 5, func(args []sx.Object, env *Environment) sx.Object {
		nLevel := getInt64(args[0], env)
		if nLevel <= 0 {
			env.err = fmt.Errorf("%v is a negative level", nLevel)
			return sx.Nil()
		}
		level := strconv.FormatInt(nLevel+ev.headingOffset, 10)

		a := ev.getAttributes(args[1], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		if fragment := getString(args[3], env).String(); fragment != "" {
			a = a.Set("id", ev.unique+fragment)
		}

		if result, isPair := sx.GetPair(ev.eval(args[4], env)); isPair && result != nil {
			if len(a) > 0 {
				result = result.Cons(ev.EvaluateAttrbute(a))
			}
			return result.Cons(ev.Make("h" + level))
		}
		return sx.MakeList(ev.Make("h"+level), sx.String("<MISSING TEXT>"))
	})
	ev.bind(sz.NameSymThematic, 0, func(args []sx.Object, env *Environment) sx.Object {
		result := sx.Nil()
		if len(args) > 0 {
			if attrList := getList(ev.eval(args[0], env), env); attrList != nil {
				result = result.Cons(ev.EvaluateAttrbute(sz.GetAttributes(attrList)))
			}
		}
		return result.Cons(ev.Make("hr"))
	})

	ev.bind(sz.NameSymListOrdered, 0, ev.makeListFn("ol"))
	ev.bind(sz.NameSymListUnordered, 0, ev.makeListFn("ul"))
	ev.bind(sz.NameSymDescription, 0, func(args []sx.Object, env *Environment) sx.Object {
		if len(args) == 0 {
			return sx.Nil()
		}
		items := sx.Nil().Cons(ev.Make("dl"))
		curItem := items
		for pos := 0; pos < len(args); pos++ {
			term := getList(ev.eval(args[pos], env), env)
			curItem = curItem.AppendBang(term.Cons(ev.Make("dt")))
			pos++
			if pos >= len(args) {
				break
			}
			ddBlock := getList(ev.eval(args[pos], env), env)
			if ddBlock == nil {
				continue
			}
			for ddlst := ddBlock; ddlst != nil; ddlst = ddlst.Tail() {
				dditem := getList(ddlst.Car(), env)
				curItem = curItem.AppendBang(dditem.Cons(ev.Make("dd")))
			}
		}
		return items
	})
	ev.bind(sz.NameSymListQuote, 0, func(args []sx.Object, env *Environment) sx.Object {
		if args == nil {
			return sx.Nil()
		}
		result := sx.Nil().Cons(ev.Make("blockquote"))
		currResult := result
		for _, elem := range args {
			if quote, isPair := sx.GetPair(ev.eval(elem, env)); isPair {
				currResult = currResult.AppendBang(quote.Cons(ev.symP))
			}
		}
		return result
	})

	ev.bind(sz.NameSymTable, 1, func(args []sx.Object, env *Environment) sx.Object {
		thead := sx.Nil()
		if header := getList(ev.eval(args[0], env), env); !sx.IsNil(header) {
			thead = sx.Nil().Cons(ev.evalTableRow(header)).Cons(ev.Make("thead"))
		}

		tbody := sx.Nil()
		if len(args) > 1 {
			tbody = sx.Nil().Cons(ev.Make("tbody"))
			curBody := tbody
			for _, row := range args[1:] {
				curBody = curBody.AppendBang(ev.evalTableRow(getList(ev.eval(row, env), env)))
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
		return table.Cons(ev.Make("table"))
	})
	ev.bind(sz.NameSymCell, 0, ev.makeCellFn(""))
	ev.bind(sz.NameSymCellCenter, 0, ev.makeCellFn("center"))
	ev.bind(sz.NameSymCellLeft, 0, ev.makeCellFn("left"))
	ev.bind(sz.NameSymCellRight, 0, ev.makeCellFn("right"))

	symDiv := ev.Make("div")
	ev.bind(sz.NameSymRegionBlock, 2, ev.makeRegionFn(symDiv, true))
	ev.bind(sz.NameSymRegionQuote, 2, ev.makeRegionFn(ev.Make("blockquote"), false))
	ev.bind(sz.NameSymRegionVerse, 2, ev.makeRegionFn(symDiv, false))

	ev.bind(sz.NameSymVerbatimComment, 1, func(args []sx.Object, env *Environment) sx.Object {
		if ev.getAttributes(args[0], env).HasDefault() {
			if len(args) > 1 {
				if s := getString(args[1], env); s != "" {
					t := sx.String(s.String())
					return sx.Nil().Cons(t).Cons(ev.Make(sxhtml.NameSymBlockComment))
				}
			}
		}
		return nil
	})
	ev.bind(sz.NameSymVerbatimEval, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalVerbatim(ev.getAttributes(args[0], env).AddClass("zs-eval"), getString(args[1], env))
	})
	ev.bind(sz.NameSymVerbatimHTML, 2, ev.evalHTML)
	ev.bind(sz.NameSymVerbatimMath, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalVerbatim(ev.getAttributes(args[0], env).AddClass("zs-math"), getString(args[1], env))
	})
	ev.bind(sz.NameSymVerbatimProg, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		content := getString(args[1], env)
		if a.HasDefault() {
			content = sx.String(visibleReplacer.Replace(content.String()))
		}
		return ev.evalVerbatim(a, content)
	})
	ev.bind(sz.NameSymVerbatimZettel, 0, nilFn)
	ev.bind(sz.NameSymBLOB, 3, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalBLOB(getList(args[0], env), getString(args[1], env), getString(args[2], env))
	})
	ev.bind(sz.NameSymTransclude, 2, func(args []sx.Object, env *Environment) sx.Object {
		ref, isPair := sx.GetPair(ev.eval(args[1], env))
		if !isPair {
			return sx.Nil()
		}
		refKind := ref.Car()
		if sx.IsNil(refKind) {
			return sx.Nil()
		}
		if refValue := getString(ref.Tail().Car(), env); refValue != "" {
			if refSym, isRefSym := sx.GetSymbol(refKind); isRefSym && refSym.Name() == sz.NameSymRefStateExternal {
				a := ev.getAttributes(args[0], env).Set("src", refValue.String()).AddClass("external")
				return sx.Nil().Cons(sx.Nil().Cons(ev.EvaluateAttrbute(a)).Cons(ev.Make("img"))).Cons(ev.symP)
			}
			return sx.MakeList(
				ev.Make(sxhtml.NameSymInlineComment),
				sx.String("transclude"),
				refKind,
				sx.String("->"),
				refValue,
			)
		}
		return ev.evalSlice(args, env)
	})
}

func (ev *Evaluator) makeListFn(tag string) EvalFn {
	sym := ev.Make(tag)
	return func(args []sx.Object, env *Environment) sx.Object {
		result := sx.Nil().Cons(sym)
		last := result
		for _, elem := range args {
			item := sx.Nil().Cons(ev.symLI)
			if res, isPair := sx.GetPair(ev.eval(elem, env)); isPair {
				item.ExtendBang(res)
			}
			last = last.AppendBang(item)
		}
		return result
	}
}

func (ev *Evaluator) evalTableRow(pairs *sx.Pair) *sx.Pair {
	row := sx.Nil().Cons(ev.Make("tr"))
	if pairs == nil {
		return nil
	}
	curRow := row
	for pair := pairs; pair != nil; pair = pair.Tail() {
		curRow = curRow.AppendBang(pair.Car())
	}
	return row
}
func (ev *Evaluator) makeCellFn(align string) EvalFn {
	return func(args []sx.Object, env *Environment) sx.Object {
		tdata := ev.evalSlice(args, env)
		if align != "" {
			tdata = tdata.Cons(ev.EvaluateAttrbute(attrs.Attributes{"class": align}))
		}
		return tdata.Cons(ev.Make("td"))
	}
}

func (ev *Evaluator) makeRegionFn(sym *sx.Symbol, genericToClass bool) EvalFn {
	return func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
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
		if region, isPair := sx.GetPair(ev.eval(args[1], env)); isPair {
			currResult = currResult.ExtendBang(region)
		}
		if len(args) > 2 {
			if cite, isPair := sx.GetPair(ev.eval(args[2], env)); isPair && cite != nil {
				currResult.AppendBang(cite.Cons(ev.Make("cite")))
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
	code = code.Cons(ev.Make("code"))
	return sx.Nil().Cons(code).Cons(ev.Make("pre"))
}

func (ev *Evaluator) bindInlines() {
	ev.bind(sz.NameSymInline, 0, ev.evalList)
	ev.bind(sz.NameSymText, 1, func(args []sx.Object, env *Environment) sx.Object { return getString(args[0], env) })
	ev.bind(sz.NameSymSpace, 0, func(args []sx.Object, env *Environment) sx.Object {
		if len(args) == 0 {
			return sx.String(" ")
		}
		return getString(args[0], env)
	})
	ev.bind(sz.NameSymSoft, 0, func([]sx.Object, *Environment) sx.Object { return sx.String(" ") })
	symBR := ev.Make("br")
	ev.bind(sz.NameSymHard, 0, func([]sx.Object, *Environment) sx.Object { return sx.Nil().Cons(symBR) })

	ev.bind(sz.NameSymLinkInvalid, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		var inline *sx.Pair
		if len(args) > 2 {
			inline = ev.evalSlice(args[2:], env)
		}
		if inline == nil {
			inline = sx.Nil().Cons(ev.eval(args[1], env))
		}
		return inline.Cons(ev.symSpan)
	})
	evalHREF := func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		return ev.evalLink(a.Set("href", refValue.String()), refValue, args[2:], env)
	}
	ev.bind(sz.NameSymLinkZettel, 2, evalHREF)
	ev.bind(sz.NameSymLinkSelf, 2, evalHREF)
	ev.bind(sz.NameSymLinkFound, 2, evalHREF)
	ev.bind(sz.NameSymLinkBroken, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		return ev.evalLink(a.AddClass("broken"), refValue, args[2:], env)
	})
	ev.bind(sz.NameSymLinkHosted, 2, evalHREF)
	ev.bind(sz.NameSymLinkBased, 2, evalHREF)
	ev.bind(sz.NameSymLinkQuery, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue.String())
		return ev.evalLink(a.Set("href", query), refValue, args[2:], env)
	})
	ev.bind(sz.NameSymLinkExternal, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		refValue := getString(args[1], env)
		return ev.evalLink(a.Set("href", refValue.String()).AddClass("external"), refValue, args[2:], env)
	})

	ev.bind(sz.NameSymEmbed, 3, func(args []sx.Object, env *Environment) sx.Object {
		ref := getList(ev.eval(args[1], env), env)
		syntax := getString(args[2], env)
		if syntax == api.ValueSyntaxSVG {
			embedAttr := sx.MakeList(
				ev.symAttr,
				sx.Cons(ev.Make("type"), sx.String("image/svg+xml")),
				sx.Cons(ev.Make("src"), sx.String("/"+getString(ref.Tail(), env).String()+".svg")),
			)
			return sx.MakeList(
				ev.Make("figure"),
				sx.MakeList(
					ev.Make("embed"),
					embedAttr,
				),
			)
		}
		a := ev.getAttributes(args[0], env)
		a = a.Set("src", getString(ref.Tail().Car(), env).String())
		if len(args) > 3 {
			var sb strings.Builder
			flattenText(&sb, sx.MakeList(args[3:]...))
			if d := sb.String(); d != "" {
				a = a.Set("alt", d)
			}
		}
		return sx.MakeList(ev.Make("img"), ev.EvaluateAttrbute(a))
	})
	ev.bind(sz.NameSymEmbedBLOB, 3, func(args []sx.Object, env *Environment) sx.Object {
		a, syntax, data := ev.getAttributes(args[0], env), getString(args[1], env), getString(args[2], env)
		summary, hasSummary := a.Get(api.KeySummary)
		if !hasSummary {
			summary = ""
		}
		return ev.evalBLOB(
			sx.MakeList(ev.Make(sx.ListName), sx.String(summary)),
			syntax,
			data,
		)
	})

	ev.bind(sz.NameSymCite, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
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
		return result.Cons(ev.symSpan)
	})
	ev.bind(sz.NameSymMark, 3, func(args []sx.Object, env *Environment) sx.Object {
		result := ev.evalSlice(args[3:], env)
		if !ev.noLinks {
			if fragment := getString(args[2], env); fragment != "" {
				a := attrs.Attributes{"id": fragment.String() + ev.unique}
				return result.Cons(ev.EvaluateAttrbute(a)).Cons(ev.symA)
			}
		}
		return result.Cons(ev.symSpan)
	})
	ev.bind(sz.NameSymEndnote, 1, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
		env.pushAttributes(a)
		defer env.popAttributes()
		attrPlist := sx.Nil()
		if len(a) > 0 {
			if attrs := ev.EvaluateAttrbute(a); attrs != nil {
				attrPlist = attrs.Tail()
			}
		}

		text, isPair := sx.GetPair(args[1])
		if !isPair {
			return sx.Nil()
		}
		noteNum := strconv.Itoa(len(env.endnotes) + 1)
		noteID := ev.unique + noteNum
		env.endnotes = append(env.endnotes, endnoteInfo{
			noteID: noteID, noteAST: ev.eval(text, env), noteHx: nil, attrs: attrPlist})
		hrefAttr := sx.Nil().Cons(sx.Cons(ev.Make("role"), sx.String("doc-noteref"))).
			Cons(sx.Cons(ev.symHREF, sx.String("#fn:"+noteID))).
			Cons(sx.Cons(ev.symClass, sx.String("zs-noteref"))).
			Cons(ev.symAttr)
		href := sx.Nil().Cons(sx.String(noteNum)).Cons(hrefAttr).Cons(ev.symA)
		supAttr := sx.Nil().Cons(sx.Cons(ev.Make("id"), sx.String("fnref:"+noteID))).Cons(ev.symAttr)
		return sx.Nil().Cons(href).Cons(supAttr).Cons(ev.Make("sup"))
	})

	ev.bind(sz.NameSymFormatDelete, 1, ev.makeFormatFn("del"))
	ev.bind(sz.NameSymFormatEmph, 1, ev.makeFormatFn("em"))
	ev.bind(sz.NameSymFormatInsert, 1, ev.makeFormatFn("ins"))
	ev.bind(sz.NameSymFormatQuote, 1, ev.evalQuote)
	ev.bind(sz.NameSymFormatSpan, 1, ev.makeFormatFn("span"))
	ev.bind(sz.NameSymFormatStrong, 1, ev.makeFormatFn("strong"))
	ev.bind(sz.NameSymFormatSub, 1, ev.makeFormatFn("sub"))
	ev.bind(sz.NameSymFormatSuper, 1, ev.makeFormatFn("sup"))

	ev.bind(sz.NameSymLiteralComment, 1, func(args []sx.Object, env *Environment) sx.Object {
		if ev.getAttributes(args[0], env).HasDefault() {
			if len(args) > 1 {
				if s := getString(ev.eval(args[1], env), env); s != "" {
					return sx.Nil().Cons(s).Cons(ev.Make(sxhtml.NameSymInlineComment))
				}
			}
		}
		return sx.Nil()
	})
	ev.bind(sz.NameSymLiteralHTML, 2, ev.evalHTML)
	kbdSym := ev.Make("kbd")
	ev.bind(sz.NameSymLiteralInput, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalLiteral(args, nil, kbdSym, env)
	})
	codeSym := ev.Make("code")
	ev.bind(sz.NameSymLiteralMath, 2, func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env).AddClass("zs-math")
		return ev.evalLiteral(args, a, codeSym, env)
	})
	sampSym := ev.Make("samp")
	ev.bind(sz.NameSymLiteralOutput, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalLiteral(args, nil, sampSym, env)
	})
	ev.bind(sz.NameSymLiteralProg, 2, func(args []sx.Object, env *Environment) sx.Object {
		return ev.evalLiteral(args, nil, codeSym, env)
	})

	ev.bind(sz.NameSymLiteralZettel, 0, nilFn)
}

func (ev *Evaluator) makeFormatFn(tag string) EvalFn {
	sym := ev.Make(tag)
	return func(args []sx.Object, env *Environment) sx.Object {
		a := ev.getAttributes(args[0], env)
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
	if env.secondaryQuotes {
		return data.secLeft, data.secRight
	}
	return data.primLeft, data.primRight
}

func (ev *Evaluator) evalQuote(args []sx.Object, env *Environment) sx.Object {
	a := ev.getAttributes(args[0], env)
	env.pushAttributes(a)
	defer env.popAttributes()

	if val, hasClass := a.Get(""); hasClass {
		a = a.Remove("").AddClass(val)
	}
	quotes := getQuoteData(env.getLanguage())
	leftQ, rightQ := getQuotes(&quotes, env)
	env.secondaryQuotes = !env.secondaryQuotes

	res := ev.evalSlice(args[1:], env)
	lastPair := res.LastPair()
	if lastPair.IsNil() {
		res = sx.Cons(sx.MakeList(ev.symNoEscape, sx.String(leftQ), sx.String(rightQ)), sx.Nil())
	} else {
		if quotes.nbsp {
			lastPair.AppendBang(sx.MakeList(ev.symNoEscape, sx.String("&nbsp;"), sx.String(rightQ)))
			res = res.Cons(sx.MakeList(ev.symNoEscape, sx.String(leftQ), sx.String("&nbsp;")))
		} else {
			lastPair.AppendBang(sx.MakeList(ev.symNoEscape, sx.String(rightQ)))
			res = res.Cons(sx.MakeList(ev.symNoEscape, sx.String(leftQ)))
		}
	}
	if len(a) > 0 {
		res = res.Cons(ev.EvaluateAttrbute(a))
		return res.Cons(ev.symSpan)
	}
	return res.Cons(ev.symList)
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func (ev *Evaluator) evalLiteral(args []sx.Object, a attrs.Attributes, sym *sx.Symbol, env *Environment) sx.Object {
	if a == nil {
		a = ev.getAttributes(args[0], env)
	}
	a = setProgLang(a)
	literal := getString(args[1], env).String()
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
	if s := getString(ev.eval(args[1], env), env); s != "" && IsSafe(s.String()) {
		return sx.Nil().Cons(s).Cons(ev.symNoEscape)
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
		return sx.Nil().Cons(sx.Nil().Cons(data).Cons(ev.symNoEscape)).Cons(ev.symP)
	default:
		imgAttr := sx.Nil().Cons(sx.Cons(ev.Make("src"), sx.String("data:image/"+syntax.String()+";base64,"+data.String())))
		var sb strings.Builder
		flattenText(&sb, description)
		if d := sb.String(); d != "" {
			imgAttr = imgAttr.Cons(sx.Cons(ev.Make("alt"), sx.String(d)))
		}
		return sx.Nil().Cons(sx.Nil().Cons(imgAttr.Cons(ev.symAttr)).Cons(ev.Make("img"))).Cons(ev.symP)
	}
}

func flattenText(sb *strings.Builder, lst *sx.Pair) {
	for elem := lst; elem != nil; elem = elem.Tail() {
		switch obj := elem.Car().(type) {
		case sx.String:
			sb.WriteString(obj.String())
		case *sx.Symbol:
			if obj.Name() == sz.NameSymSpace {
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

func (ev *Evaluator) eval(obj sx.Object, env *Environment) sx.Object {
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
	name := sym.Name()
	fn, found := ev.fns[name]
	if !found {
		env.err = fmt.Errorf("symbol %q not bound", name)
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
	if minArgs, hasMinArgs := ev.minArgs[name]; hasMinArgs {
		if minArgs > len(args) {
			env.err = fmt.Errorf("%v needs at least %d arguments, but got only %d", name, minArgs, len(args))
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
		elem := ev.eval(arg, env)
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
		return result.Cons(ev.symSpan)
	}
	return result.Cons(ev.EvaluateAttrbute(a)).Cons(ev.symA)
}

func (ev *Evaluator) getSymbol(val sx.Object, env *Environment) *sx.Symbol {
	if env.err == nil {
		if sym, ok := sx.GetSymbol(val); ok {
			return sym
		}
		env.err = fmt.Errorf("%v/%T is not a symbol", val, val)
	}
	return ev.Make("???")
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
func (ev *Evaluator) getAttributes(arg sx.Object, env *Environment) attrs.Attributes {
	return sz.GetAttributes(getList(ev.eval(arg, env), env))
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
