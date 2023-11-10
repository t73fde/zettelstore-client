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
	"log"
	"net/url"
	"strconv"
	"strings"

	"zettelstore.de/client.fossil/api"
	"zettelstore.de/client.fossil/attrs"
	"zettelstore.de/client.fossil/sz"
	"zettelstore.de/sx.fossil"
	"zettelstore.de/sx.fossil/sxhtml"
)

// Transformer will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Transformer struct {
	sf            sx.SymbolFactory
	headingOffset int64
	unique        string
	endnotes      []endnoteInfo
	noLinks       bool // true iff output must not include links
	symAttr       *sx.Symbol
	symClass      *sx.Symbol
	symMeta       *sx.Symbol
	symP          *sx.Symbol
	symA          *sx.Symbol
	symSpan       *sx.Symbol
}

type endnoteInfo struct {
	noteAST *sx.Pair // Endnote as AST
	noteHx  *sx.Pair // Endnote as SxHTML
	attrs   *sx.Pair // attrs a-list
}

// NewTransformer creates a new transformer object.
func NewTransformer(headingOffset int, sf sx.SymbolFactory) *Transformer {
	if sf == nil {
		sf = sx.MakeMappedFactory(128)
	}
	return &Transformer{
		sf:            sf,
		headingOffset: int64(headingOffset),
		symAttr:       sf.MustMake(sxhtml.NameSymAttr),
		symClass:      sf.MustMake("class"),
		symMeta:       sf.MustMake("meta"),
		symP:          sf.MustMake("p"),
		symA:          sf.MustMake("a"),
		symSpan:       sf.MustMake("span"),
	}
}

// SetUnique sets a prefix to make several HTML ids unique.
func (tr *Transformer) SetUnique(s string) { tr.unique = s }

// IsValidName returns true, if name is a valid symbol name.
func (tr *Transformer) IsValidName(s string) bool { return tr.sf.IsValidName(s) }

// Make a new HTML symbol.
func (tr *Transformer) Make(s string) *sx.Symbol { return tr.sf.MustMake(s) }

// TransformAttrbute transforms the given attributes into a HTML s-expression.
func (tr *Transformer) TransformAttrbute(a attrs.Attributes) *sx.Pair {
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

func (tr *Transformer) createEvaluator() evaluator {
	sf := tr.sf
	ev := evaluator{
		tr:          tr,
		sf:          sf,
		err:         nil,
		fns:         make(map[string]evalFn, 128),
		minArgs:     make(map[string]int, 128),
		symNoEscape: sf.MustMake(sxhtml.NameSymNoEscape),
		symAttr:     tr.symAttr,
		symMeta:     tr.symMeta,
		symP:        tr.symP,
		symA:        tr.symA,
		symSpan:     tr.symSpan,
	}
	return ev
}

// TransformMetadate transforms a metadata s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformMetadata(lst *sx.Pair) (sx.Object, error) {
	log.Println("TMET", lst)
	v := tr.createEvaluator()
	v.bindCommon()
	v.bindMetadata()
	v.bindInlines()
	result := v.eval(lst)
	log.Println("RMET", v.err, result)
	return result, v.err
}

// TransformBlock transforms a block AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformBlock(lst *sx.Pair) (sx.Object, error) {
	log.Println("TBLO", lst)
	v := tr.createEvaluator()
	v.bindCommon()
	v.bindBlocks()
	v.bindInlines()
	result := v.eval(lst)
	log.Println("RBLO", v.err, result)
	return result, v.err
}

// TransformInline transforms an inline AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformInline(lst *sx.Pair) (sx.Object, error) {
	log.Println("TINL", lst)
	v := tr.createEvaluator()
	v.bindCommon()
	v.bindInlines()
	result := v.eval(lst)
	log.Println("RINL", v.err, result)
	return result, v.err
}

// Endnotes returns a SHTML object with all collected endnotes.
func (tr *Transformer) Endnotes() *sx.Pair {
	if len(tr.endnotes) == 0 {
		return nil
	}
	result := sx.Nil().Cons(tr.Make("ol"))
	currResult := result.AppendBang(sx.Nil().Cons(sx.Cons(tr.symClass, sx.String("zs-endnotes"))).Cons(tr.symAttr))
	for i, fni := range tr.endnotes {
		noteNum := strconv.Itoa(i + 1)
		noteID := tr.unique + noteNum

		attrs := fni.attrs.Cons(sx.Cons(tr.symClass, sx.String("zs-endnote"))).
			Cons(sx.Cons(tr.Make("value"), sx.String(noteNum))).
			Cons(sx.Cons(tr.Make("id"), sx.String("fn:"+noteID))).
			Cons(sx.Cons(tr.Make("role"), sx.String("doc-endnote"))).
			Cons(tr.symAttr)

		backref := sx.Nil().Cons(sx.String("\u21a9\ufe0e")).
			Cons(sx.Nil().
				Cons(sx.Cons(tr.symClass, sx.String("zs-endnote-backref"))).
				Cons(sx.Cons(tr.Make("href"), sx.String("#fnref:"+noteID))).
				Cons(sx.Cons(tr.Make("role"), sx.String("doc-backlink"))).
				Cons(tr.symAttr)).
			Cons(tr.symA)

		li := sx.Nil().Cons(tr.Make("li"))
		li.AppendBang(attrs).
			ExtendBang(fni.noteHx).
			AppendBang(sx.String(" ")).AppendBang(backref)
		currResult = currResult.AppendBang(li)
	}
	tr.endnotes = nil
	return result
}

// evaluator is the environment where the actual transformation takes places.
type evaluator struct {
	tr          *Transformer
	sf          sx.SymbolFactory
	err         error
	fns         map[string]evalFn
	minArgs     map[string]int
	symNoEscape *sx.Symbol
	symAttr     *sx.Symbol
	symMeta     *sx.Symbol
	symP        *sx.Symbol
	symA        *sx.Symbol
	symSpan     *sx.Symbol
}
type evalFn func([]sx.Object) sx.Object

func (ev *evaluator) bind(name string, minArgs int, fn evalFn) {
	ev.fns[name] = fn
	if minArgs > 0 {
		ev.minArgs[name] = minArgs
	}
}

func (ev *evaluator) bindCommon() {
	ev.bind(sx.ListName, 0, ev.evalList)
	ev.bind("quote", 1, func(args []sx.Object) sx.Object { return args[0] })
}

func (ev *evaluator) bindMetadata() {
	ev.bind(sz.NameSymMeta, 0, ev.evalList)
	evalMetaString := func(args []sx.Object) sx.Object {
		a := make(attrs.Attributes, 2).
			Set("name", ev.getSymbol(ev.eval(args[0])).Name()).
			Set("content", ev.getString(args[1]).String())
		return ev.evalMeta(a)
	}
	ev.bind(sz.NameSymTypeCredential, 2, evalMetaString)
	ev.bind(sz.NameSymTypeEmpty, 2, evalMetaString)
	ev.bind(sz.NameSymTypeID, 2, evalMetaString)
	ev.bind(sz.NameSymTypeNumber, 2, evalMetaString)
	ev.bind(sz.NameSymTypeString, 2, evalMetaString)
	ev.bind(sz.NameSymTypeTimestamp, 2, evalMetaString)
	ev.bind(sz.NameSymTypeURL, 2, evalMetaString)
	ev.bind(sz.NameSymTypeWord, 2, evalMetaString)

	evalMetaSet := func(args []sx.Object) sx.Object {
		var sb strings.Builder
		lst := ev.eval(args[1])
		for elem := ev.getList(lst); elem != nil; elem = elem.Tail() {
			sb.WriteByte(' ')
			sb.WriteString(ev.getString(elem.Car()).String())
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", ev.getSymbol(ev.eval(args[0])).Name()).
			Set("content", s)
		return ev.evalMeta(a)
	}
	ev.bind(sz.NameSymTypeIDSet, 2, evalMetaSet)
	ev.bind(sz.NameSymTypeTagSet, 2, evalMetaSet)
	ev.bind(sz.NameSymTypeWordSet, 2, evalMetaSet)
}

func (ev *evaluator) evalMeta(a attrs.Attributes) *sx.Pair {
	return sx.Nil().Cons(ev.tr.TransformAttrbute(a)).Cons(ev.symMeta)
}

func (ev *evaluator) bindBlocks() {
	ev.bind(sz.NameSymBlock, 0, ev.evalList)
	ev.bind(sz.NameSymPara, 0, func(args []sx.Object) sx.Object {
		return ev.evalSlice(args).Cons(ev.symP)
	})
	ev.bind(sz.NameSymHeading, 5, func(args []sx.Object) sx.Object {
		nLevel := ev.getInt64(args[0])
		if nLevel <= 0 {
			ev.err = fmt.Errorf("%v is a negative level", nLevel)
			return sx.Nil()
		}
		level := strconv.FormatInt(nLevel+ev.tr.headingOffset, 10)

		a := ev.getAttributes(args[1])
		if fragment := ev.getString(args[3]).String(); fragment != "" {
			a = a.Set("id", ev.tr.unique+fragment)
		}

		if result, isPair := sx.GetPair(ev.eval(args[4])); isPair && result != nil {
			if len(a) > 0 {
				result = result.Cons(ev.evalAttribute(a))
			}
			return result.Cons(ev.make("h" + level))
		}
		return sx.MakeList(ev.make("h"+level), sx.String("<MISSING TEXT>"))
	})
	ev.bind(sz.NameSymThematic, 0, func(args []sx.Object) sx.Object {
		result := sx.Nil()
		if len(args) > 0 {
			if attrList := ev.getList(ev.eval(args[0])); attrList != nil {
				result = result.Cons(ev.evalAttribute(sz.GetAttributes(attrList)))
			}
		}
		return result.Cons(ev.make("hr"))
	})

	ev.bind(sz.NameSymListOrdered, 0, ev.makeListFn("ol"))
	ev.bind(sz.NameSymListUnordered, 0, ev.makeListFn("ul"))
	ev.bind(sz.NameSymDescription, 0, func(args []sx.Object) sx.Object {
		if len(args) == 0 {
			return sx.Nil()
		}
		items := sx.Nil().Cons(ev.make("dl"))
		curItem := items
		for pos := 0; pos < len(args); pos++ {
			term := ev.getList(ev.eval(args[pos]))
			curItem = curItem.AppendBang(term.Cons(ev.make("dt")))
			pos++
			if pos >= len(args) {
				break
			}
			ddBlock := ev.getList(ev.eval(args[pos]))
			if ddBlock == nil {
				continue
			}
			for ddlst := ddBlock; ddlst != nil; ddlst = ddlst.Tail() {
				dditem := ev.getList(ddlst.Car())
				curItem = curItem.AppendBang(dditem.Cons(ev.make("dd")))
			}
		}
		return items
	})
	ev.bind(sz.NameSymListQuote, 0, func(args []sx.Object) sx.Object {
		if args == nil {
			return sx.Nil()
		}
		result := sx.Nil().Cons(ev.make("blockquote"))
		currResult := result
		for _, elem := range args {
			if quote, isPair := sx.GetPair(ev.eval(elem)); isPair {
				currResult = currResult.AppendBang(quote.Cons(ev.symP))
			}
		}
		return result
	})

	ev.bind(sz.NameSymTable, 1, func(args []sx.Object) sx.Object {
		thead := sx.Nil()
		if header := ev.getList(ev.eval(args[0])); !sx.IsNil(header) {
			thead = sx.Nil().Cons(ev.transformTableRow(header)).Cons(ev.make("thead"))
		}

		tbody := sx.Nil()
		if len(args) > 1 {
			tbody = sx.Nil().Cons(ev.make("tbody"))
			curBody := tbody
			for _, row := range args[1:] {
				curBody = curBody.AppendBang(ev.transformTableRow(ev.getList(ev.eval(row))))
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
		return table.Cons(ev.make("table"))
	})
	ev.bind(sz.NameSymCell, 0, ev.makeCellFn(""))
	ev.bind(sz.NameSymCellCenter, 0, ev.makeCellFn("center"))
	ev.bind(sz.NameSymCellLeft, 0, ev.makeCellFn("left"))
	ev.bind(sz.NameSymCellRight, 0, ev.makeCellFn("right"))

	symDiv := ev.make("div")
	ev.bind(sz.NameSymRegionBlock, 2, ev.makeRegionFn(symDiv, true))
	ev.bind(sz.NameSymRegionQuote, 2, ev.makeRegionFn(ev.make("blockquote"), false))
	ev.bind(sz.NameSymRegionVerse, 2, ev.makeRegionFn(symDiv, false))

	ev.bind(sz.NameSymVerbatimComment, 1, func(args []sx.Object) sx.Object {
		if ev.getAttributes(args[0]).HasDefault() {
			if len(args) > 1 {
				if s := ev.getString(args[1]); s != "" {
					t := sx.String(s.String())
					return sx.Nil().Cons(t).Cons(ev.make(sxhtml.NameSymBlockComment))
				}
			}
		}
		return nil
	})
	ev.bind(sz.NameSymVerbatimEval, 2, func(args []sx.Object) sx.Object {
		return ev.evalVerbatim(ev.getAttributes(args[0]).AddClass("zs-eval"), ev.getString(args[1]))
	})
	ev.bind(sz.NameSymVerbatimHTML, 2, ev.visitHTML)
	ev.bind(sz.NameSymVerbatimMath, 2, func(args []sx.Object) sx.Object {
		return ev.evalVerbatim(ev.getAttributes(args[0]).AddClass("zs-math"), ev.getString(args[1]))
	})
	ev.bind(sz.NameSymVerbatimProg, 2, func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		content := ev.getString(args[1])
		if a.HasDefault() {
			content = sx.String(visibleReplacer.Replace(content.String()))
		}
		return ev.evalVerbatim(a, content)
	})
	ev.bind(sz.NameSymVerbatimZettel, 0, noopFn)
	ev.bind(sz.NameSymBLOB, 3, func(args []sx.Object) sx.Object {
		return ev.visitBLOB(ev.getList(args[0]), ev.getString(args[1]), ev.getString(args[2]))
	})
	ev.bind(sz.NameSymTransclude, 2, func(args []sx.Object) sx.Object {
		ref, isPair := sx.GetPair(ev.eval(args[1]))
		if !isPair {
			return sx.Nil()
		}
		refKind := ref.Car()
		if sx.IsNil(refKind) {
			return sx.Nil()
		}
		if refValue := ev.getString(ref.Tail().Car()); refValue != "" {
			if refSym, isRefSym := sx.GetSymbol(refKind); isRefSym && refSym.Name() == sz.NameSymRefStateExternal {
				a := ev.getAttributes(args[0]).Set("src", refValue.String()).AddClass("external")
				return sx.Nil().Cons(sx.Nil().Cons(ev.evalAttribute(a)).Cons(ev.make("img"))).Cons(ev.symP)
			}
			return sx.MakeList(
				ev.make(sxhtml.NameSymInlineComment),
				sx.String("transclude"),
				refKind,
				sx.String("->"),
				refValue,
			)
		}
		return ev.evalSlice(args)
	})
}

func (ev *evaluator) makeListFn(tag string) evalFn {
	sym := ev.make(tag)
	symLI := ev.make("li")
	return func(args []sx.Object) sx.Object {
		result := sx.Nil().Cons(sym)
		last := result
		for _, elem := range args {
			item := sx.Nil().Cons(symLI)
			if res, isPair := sx.GetPair(ev.eval(elem)); isPair {
				item.ExtendBang(res)
			}
			last = last.AppendBang(item)
		}
		return result
	}
}

func (ev *evaluator) transformTableRow(pairs *sx.Pair) *sx.Pair {
	row := sx.Nil().Cons(ev.make("tr"))
	if pairs == nil {
		return nil
	}
	curRow := row
	for pair := pairs; pair != nil; pair = pair.Tail() {
		curRow = curRow.AppendBang(pair.Car())
	}
	return row
}
func (ev *evaluator) makeCellFn(align string) evalFn {
	return func(args []sx.Object) sx.Object {
		tdata := ev.evalSlice(args)
		if align != "" {
			tdata = tdata.Cons(ev.evalAttribute(attrs.Attributes{"class": align}))
		}
		return tdata.Cons(ev.make("td"))
	}
}

func (ev *evaluator) makeRegionFn(sym *sx.Symbol, genericToClass bool) evalFn {
	return func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		if genericToClass {
			if val, found := a.Get(""); found {
				a = a.Remove("").AddClass(val)
			}
		}
		result := sx.Nil()
		if len(a) > 0 {
			result = result.Cons(ev.evalAttribute(a))
		}
		result = result.Cons(sym)
		currResult := result.LastPair()
		if region, isPair := sx.GetPair(ev.eval(args[1])); isPair {
			currResult = currResult.ExtendBang(region)
		}
		if len(args) > 2 {
			if cite, isPair := sx.GetPair(ev.eval(args[2])); isPair && cite != nil {
				currResult.AppendBang(cite.Cons(ev.make("cite")))
			}
		}
		return result
	}
}

func (ev *evaluator) evalVerbatim(a attrs.Attributes, s sx.String) sx.Object {
	a = setProgLang(a)
	code := sx.Nil().Cons(s)
	if al := ev.evalAttribute(a); al != nil {
		code = code.Cons(al)
	}
	code = code.Cons(ev.make("code"))
	return sx.Nil().Cons(code).Cons(ev.make("pre"))
}

func (ev *evaluator) bindInlines() {
	ev.bind(sz.NameSymInline, 0, ev.evalList)
	ev.bind(sz.NameSymText, 1, func(args []sx.Object) sx.Object { return ev.getString(args[0]) })
	ev.bind(sz.NameSymSpace, 0, func(args []sx.Object) sx.Object {
		if len(args) == 0 {
			return sx.String(" ")
		}
		return ev.getString(args[0])
	})
	ev.bind(sz.NameSymSoft, 0, func([]sx.Object) sx.Object { return sx.String(" ") })
	symBR := ev.make("br")
	ev.bind(sz.NameSymHard, 0, func([]sx.Object) sx.Object { return sx.Nil().Cons(symBR) })

	ev.bind(sz.NameSymLinkInvalid, 2, func(args []sx.Object) sx.Object {
		// a := ev.getAttributes(args)
		var inline *sx.Pair
		if len(args) > 2 {
			inline = ev.evalSlice(args[2:])
		}
		if inline == nil {
			inline = sx.Nil().Cons(ev.eval(args[1]))
		}
		return inline.Cons(ev.symSpan)
	})
	evalHREF := func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		refValue := ev.getString(args[1])
		return ev.evalLink(a.Set("href", refValue.String()), refValue, args[2:])
	}
	ev.bind(sz.NameSymLinkZettel, 2, evalHREF)
	ev.bind(sz.NameSymLinkSelf, 2, evalHREF)
	ev.bind(sz.NameSymLinkFound, 2, evalHREF)
	ev.bind(sz.NameSymLinkBroken, 2, func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		refValue := ev.getString(args[1])
		return ev.evalLink(a.AddClass("broken"), refValue, args[2:])
	})
	ev.bind(sz.NameSymLinkHosted, 2, evalHREF)
	ev.bind(sz.NameSymLinkBased, 2, evalHREF)
	ev.bind(sz.NameSymLinkQuery, 2, func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		refValue := ev.getString(args[1])
		query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue.String())
		return ev.evalLink(a.Set("href", query), refValue, args[2:])
	})
	ev.bind(sz.NameSymLinkExternal, 2, func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		refValue := ev.getString(args[1])
		return ev.evalLink(a.Set("href", refValue.String()).AddClass("external"), refValue, args[2:])
	})

	ev.bind(sz.NameSymEmbed, 3, func(args []sx.Object) sx.Object {
		ref := ev.getList(ev.eval(args[1]))
		syntax := ev.getString(args[2])
		if syntax == api.ValueSyntaxSVG {
			embedAttr := sx.MakeList(
				ev.symAttr,
				sx.Cons(ev.make("type"), sx.String("image/svg+xml")),
				sx.Cons(ev.make("src"), sx.String("/"+ev.getString(ref.Tail()).String()+".svg")),
			)
			return sx.MakeList(
				ev.make("figure"),
				sx.MakeList(
					ev.make("embed"),
					embedAttr,
				),
			)
		}
		a := ev.getAttributes(args[0])
		a = a.Set("src", string(ev.getString(ref.Tail().Car())))
		var sb strings.Builder
		ev.flattenText(&sb, ref.Tail().Tail().Tail())
		if d := sb.String(); d != "" {
			a = a.Set("alt", d)
		}
		return sx.MakeList(ev.make("img"), ev.evalAttribute(a))
	})
	ev.bind(sz.NameSymEmbedBLOB, 3, noopFn)

	ev.bind(sz.NameSymCite, 2, func(args []sx.Object) sx.Object {
		result := sx.Nil()
		if key := ev.getString(args[1]); key != "" {
			if len(args) > 2 {
				result = ev.evalSlice(args[2:]).Cons(sx.String(", "))
			}
			result = result.Cons(key)
		}
		if a := ev.getAttributes(args[0]); len(a) > 0 {
			result = result.Cons(ev.evalAttribute(a))
		}
		if result == nil {
			return nil
		}
		return result.Cons(ev.symSpan)
	})
	ev.bind(sz.NameSymMark, 3, func(args []sx.Object) sx.Object {
		result := ev.evalSlice(args[3:])
		if !ev.tr.noLinks {
			if fragment := ev.getString(args[2]); fragment != "" {
				a := attrs.Attributes{"id": fragment.String() + ev.tr.unique}
				return result.Cons(ev.evalAttribute(a)).Cons(ev.symA)
			}
		}
		return result.Cons(ev.symSpan)
	})
	ev.bind(sz.NameSymEndnote, 1, func(args []sx.Object) sx.Object {
		attrPlist := sx.Nil()
		if a := ev.getAttributes(args[0]); len(a) > 0 {
			if attrs := ev.evalAttribute(a); attrs != nil {
				attrPlist = attrs.Tail()
			}
		}

		text, isPair := sx.GetPair(args[1])
		if !isPair {
			return sx.Nil()
		}
		ev.tr.endnotes = append(ev.tr.endnotes, endnoteInfo{noteAST: text, noteHx: nil, attrs: attrPlist})
		noteNum := strconv.Itoa(len(ev.tr.endnotes))
		noteID := ev.tr.unique + noteNum
		hrefAttr := sx.Nil().Cons(sx.Cons(ev.make("role"), sx.String("doc-noteref"))).
			Cons(sx.Cons(ev.make("href"), sx.String("#fn:"+noteID))).
			Cons(sx.Cons(ev.tr.symClass, sx.String("zs-noteref"))).
			Cons(ev.symAttr)
		href := sx.Nil().Cons(sx.String(noteNum)).Cons(hrefAttr).Cons(ev.symA)
		supAttr := sx.Nil().Cons(sx.Cons(ev.make("id"), sx.String("fnref:"+noteID))).Cons(ev.symAttr)
		return sx.Nil().Cons(href).Cons(supAttr).Cons(ev.make("sup"))
	})

	ev.bind(sz.NameSymFormatDelete, 1, ev.makeFormatFn("del"))
	ev.bind(sz.NameSymFormatEmph, 1, ev.makeFormatFn("em"))
	ev.bind(sz.NameSymFormatInsert, 1, ev.makeFormatFn("ins"))
	ev.bind(sz.NameSymFormatQuote, 1, func(args []sx.Object) sx.Object {
		const langAttr = "lang"
		a := ev.getAttributes(args[0])
		langVal, found := a.Get(langAttr)
		if found {
			a = a.Remove(langAttr)
		}
		if val, found2 := a.Get(""); found2 {
			a = a.Remove("").AddClass(val)
		}
		res := ev.evalSlice(args[1:])
		if len(a) > 0 {
			res = res.Cons(ev.evalAttribute(a))
		}
		res = res.Cons(ev.make("q"))
		if found {
			res = sx.Nil().Cons(res).Cons(ev.evalAttribute(attrs.Attributes{}.Set(langAttr, langVal))).Cons(ev.symSpan)
		}
		return res
	})
	ev.bind(sz.NameSymFormatSpan, 1, ev.makeFormatFn("span"))
	ev.bind(sz.NameSymFormatStrong, 1, ev.makeFormatFn("strong"))
	ev.bind(sz.NameSymFormatSub, 1, ev.makeFormatFn("sub"))
	ev.bind(sz.NameSymFormatSuper, 1, ev.makeFormatFn("sup"))

	ev.bind(sz.NameSymLiteralComment, 1, func(args []sx.Object) sx.Object {
		if ev.getAttributes(args[0]).HasDefault() {
			if len(args) > 1 {
				if s := ev.getString(ev.eval(args[1])); s != "" {
					return sx.Nil().Cons(s).Cons(ev.make(sxhtml.NameSymInlineComment))
				}
			}
		}
		return sx.Nil()
	})
	ev.bind(sz.NameSymLiteralHTML, 2, ev.visitHTML)
	kbdSym := ev.make("kbd")
	ev.bind(sz.NameSymLiteralInput, 2, func(args []sx.Object) sx.Object {
		return ev.visitLiteral(args, nil, kbdSym)
	})
	codeSym := ev.make("code")
	ev.bind(sz.NameSymLiteralMath, 2, func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0]).AddClass("zs-math")
		return ev.visitLiteral(args, a, codeSym)
	})
	sampSym := ev.make("samp")
	ev.bind(sz.NameSymLiteralOutput, 2, func(args []sx.Object) sx.Object {
		return ev.visitLiteral(args, nil, sampSym)
	})
	ev.bind(sz.NameSymLiteralProg, 2, func(args []sx.Object) sx.Object {
		return ev.visitLiteral(args, nil, codeSym)
	})

	ev.bind(sz.NameSymLiteralZettel, 0, noopFn)
}

func (ev *evaluator) makeFormatFn(tag string) evalFn {
	sym := ev.make(tag)
	return func(args []sx.Object) sx.Object {
		a := ev.getAttributes(args[0])
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		res := ev.evalSlice(args[1:])
		if len(a) > 0 {
			res = res.Cons(ev.evalAttribute(a))
		}
		return res.Cons(sym)
	}
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func (ev *evaluator) visitLiteral(args []sx.Object, a attrs.Attributes, sym *sx.Symbol) sx.Object {
	if a == nil {
		a = ev.getAttributes(args[0])
	}
	a = setProgLang(a)
	literal := ev.getString(args[1]).String()
	if a.HasDefault() {
		a = a.RemoveDefault()
		literal = visibleReplacer.Replace(literal)
	}
	res := sx.Nil().Cons(sx.String(literal))
	if len(a) > 0 {
		res = res.Cons(ev.evalAttribute(a))
	}
	return res.Cons(sym)
}
func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}

func (ev *evaluator) visitHTML(args []sx.Object) sx.Object {
	if s := ev.getString(ev.eval(args[1])); s != "" && IsSafe(s.String()) {
		return sx.Nil().Cons(s).Cons(ev.symNoEscape)
	}
	return nil
}

func (ev *evaluator) visitBLOB(description *sx.Pair, syntax, data sx.String) sx.Object {
	if data == "" {
		return sx.Nil()
	}
	switch syntax {
	case "":
		return sx.Nil()
	case api.ValueSyntaxSVG:
		return sx.Nil().Cons(sx.Nil().Cons(data).Cons(ev.symNoEscape)).Cons(ev.symP)
	default:
		imgAttr := sx.Nil().Cons(sx.Cons(ev.make("src"), sx.String("data:image/"+syntax.String()+";base64,"+data.String())))
		var sb strings.Builder
		ev.flattenText(&sb, description)
		if d := sb.String(); d != "" {
			imgAttr = imgAttr.Cons(sx.Cons(ev.make("alt"), sx.String(d)))
		}
		return sx.Nil().Cons(sx.Nil().Cons(imgAttr.Cons(ev.symAttr)).Cons(ev.make("img"))).Cons(ev.symP)
	}
}

func (ev *evaluator) flattenText(sb *strings.Builder, lst *sx.Pair) {
	for elem := lst; elem != nil; elem = elem.Tail() {
		switch obj := elem.Car().(type) {
		case sx.String:
			sb.WriteString(obj.String())
		case *sx.Pair:
			ev.flattenText(sb, obj)
		}
	}
}

func (ev *evaluator) evalList(args []sx.Object) sx.Object { return ev.evalSlice(args) }
func noopFn([]sx.Object) sx.Object                        { return sx.Nil() }

func (ev *evaluator) eval(obj sx.Object) sx.Object {
	if ev.err != nil {
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
		ev.err = fmt.Errorf("symbol expected, but got %T/%v", lst.Car(), lst.Car())
		return sx.Nil()
	}
	name := sym.Name()
	fn, found := ev.fns[name]
	if !found {
		ev.err = fmt.Errorf("symbol %q not bound", name)
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
			ev.err = fmt.Errorf("%v needs at least %d arguments, but got only %d", name, minArgs, len(args))
			return sx.Nil()
		}
	}
	log.Println("EXEC", sym, args)
	result := fn(args)
	if ev.err != nil {
		log.Println("EERR", ev.err)
		return sx.Nil()
	}
	log.Println("REXE", result)
	return result
}

func (ev *evaluator) evalSlice(args []sx.Object) *sx.Pair {
	result := sx.Cons(sx.Nil(), sx.Nil())
	curr := result
	for _, arg := range args {
		elem := ev.eval(arg)
		if ev.err != nil {
			return nil
		}
		curr = curr.AppendBang(elem)
	}
	return result.Tail()
}

func (ev *evaluator) evalLink(a attrs.Attributes, refValue sx.String, inline []sx.Object) sx.Object {
	result := ev.evalSlice(inline)
	if len(inline) == 0 {
		result = sx.Nil().Cons(refValue)
	}
	if ev.tr.noLinks {
		return result.Cons(ev.symSpan)
	}
	return result.Cons(ev.evalAttribute(a)).Cons(ev.symA)
}

func (ev *evaluator) evalAttribute(a attrs.Attributes) *sx.Pair { return ev.tr.TransformAttrbute(a) }

func (ev *evaluator) getSymbol(val sx.Object) *sx.Symbol {
	if ev.err == nil {
		if sym, ok := sx.GetSymbol(val); ok {
			return sym
		}
		ev.err = fmt.Errorf("%v/%T is not a symbol", val, val)
	}
	return ev.make("???")
}
func (ev *evaluator) getString(val sx.Object) sx.String {
	if ev.err != nil {
		return ""
	}
	if s, ok := sx.GetString(val); ok {
		return s
	}
	ev.err = fmt.Errorf("%v/%T is not a string", val, val)
	return ""
}
func (ev *evaluator) getList(val sx.Object) *sx.Pair {
	if ev.err == nil {
		if res, isPair := sx.GetPair(val); isPair {
			return res
		}
		ev.err = fmt.Errorf("%v/%T is not a list", val, val)
	}
	return nil
}
func (ev *evaluator) getInt64(val sx.Object) int64 {
	if ev.err != nil {
		return -1017
	}
	if num, ok := sx.GetNumber(val); ok {
		return int64(num.(sx.Int64))
	}
	ev.err = fmt.Errorf("%v/%T is not a number", val, val)
	return -1017
}
func (ev *evaluator) getAttributes(arg sx.Object) attrs.Attributes {
	return sz.GetAttributes(ev.getList(ev.eval(arg)))
}

func (ev *evaluator) make(name string) *sx.Symbol { return ev.sf.MustMake(name) }

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
