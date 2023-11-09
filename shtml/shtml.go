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

func (tr *Transformer) createVisitor(lst *sx.Pair) visitor {
	sf := tr.sf
	v := visitor{
		tr:          tr,
		sf:          sf,
		err:         nil,
		fns:         make(map[string]transformFn, 128),
		minArgs:     make(map[string]int, 128),
		symNoEscape: sf.MustMake(sxhtml.NameSymNoEscape),
		symAttr:     tr.symAttr,
		symMeta:     tr.symMeta,
		symP:        tr.symP,
		symA:        tr.symA,
		symSpan:     tr.symSpan,
	}
	return v
}

// TransformMetadate transforms a metadata s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformMetadata(lst *sx.Pair) (sx.Object, error) {
	log.Println("TMET", lst)
	v := tr.createVisitor(lst)
	v.bindCommon()
	v.bindMetadata()
	v.bindInlines()
	result := v.visit(lst)
	log.Println("RMET", v.err, result)
	return result, v.err
}

// TransformBlock transforms a block AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformBlock(lst *sx.Pair) (sx.Object, error) {
	log.Println("TBLO", lst)
	v := tr.createVisitor(lst)
	v.bindCommon()
	v.bindBlocks()
	v.bindInlines()
	result := v.visit(lst)
	log.Println("RBLO", v.err, result)
	return result, v.err
}

// TransformInline transforms an inline AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformInline(lst *sx.Pair) (sx.Object, error) {
	log.Println("TINL", lst)
	v := tr.createVisitor(lst)
	v.bindCommon()
	v.bindInlines()
	result := v.visit(lst)
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

// visitor is the environment where the actual transformation takes places.
type visitor struct {
	tr          *Transformer
	sf          sx.SymbolFactory
	err         error
	fns         map[string]transformFn
	minArgs     map[string]int
	symNoEscape *sx.Symbol
	symAttr     *sx.Symbol
	symMeta     *sx.Symbol
	symP        *sx.Symbol
	symA        *sx.Symbol
	symSpan     *sx.Symbol
}
type transformFn func([]sx.Object) sx.Object

func (v *visitor) bind(name string, minArgs int, fn transformFn) {
	v.fns[name] = fn
	if minArgs > 0 {
		v.minArgs[name] = minArgs
	}
}

func (v *visitor) bindCommon() {
	v.bind(sx.ListName, 0, v.visitList)
	v.bind("quote", 1, func(args []sx.Object) sx.Object { return args[0] })
}

func (v *visitor) bindMetadata() {
	v.bind(sz.NameSymMeta, 0, v.visitList)
	metaString := func(args []sx.Object) sx.Object {
		a := make(attrs.Attributes, 2).
			Set("name", v.getSymbol(v.visit(args[0])).Name()).
			Set("content", v.getString(args[1]).String())
		return v.visitMeta(a)
	}
	v.bind(sz.NameSymTypeCredential, 2, metaString)
	v.bind(sz.NameSymTypeEmpty, 2, metaString)
	v.bind(sz.NameSymTypeID, 2, metaString)
	v.bind(sz.NameSymTypeNumber, 2, metaString)
	v.bind(sz.NameSymTypeString, 2, metaString)
	v.bind(sz.NameSymTypeTimestamp, 2, metaString)
	v.bind(sz.NameSymTypeURL, 2, metaString)
	v.bind(sz.NameSymTypeWord, 2, metaString)

	metaSet := func(args []sx.Object) sx.Object {
		var sb strings.Builder
		lst := v.visit(args[1])
		for elem := v.getList(lst); elem != nil; elem = elem.Tail() {
			sb.WriteByte(' ')
			sb.WriteString(v.getString(elem.Car()).String())
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", v.getSymbol(v.visit(args[0])).Name()).
			Set("content", s)
		return v.visitMeta(a)
	}
	v.bind(sz.NameSymTypeIDSet, 2, metaSet)
	v.bind(sz.NameSymTypeTagSet, 2, metaSet)
	v.bind(sz.NameSymTypeWordSet, 2, metaSet)
}

func (v *visitor) visitMeta(a attrs.Attributes) *sx.Pair {
	return sx.Nil().Cons(v.tr.TransformAttrbute(a)).Cons(v.symMeta)
}

func (v *visitor) bindBlocks() {
	v.bind(sz.NameSymBlock, 0, v.visitList)
	v.bind(sz.NameSymPara, 0, func(args []sx.Object) sx.Object {
		return v.visitSlice(args).Cons(v.symP)
	})
	v.bind(sz.NameSymHeading, 5, func(args []sx.Object) sx.Object {
		nLevel := v.getInt64(args[0])
		if nLevel <= 0 {
			v.err = fmt.Errorf("%v is a negative level", nLevel)
			return sx.Nil()
		}
		level := strconv.FormatInt(nLevel+v.tr.headingOffset, 10)

		a := v.getAttributes(args[1])
		if fragment := v.getString(args[3]).String(); fragment != "" {
			a = a.Set("id", v.tr.unique+fragment)
		}

		if result, isPair := sx.GetPair(v.visit(args[4])); isPair && result != nil {
			if len(a) > 0 {
				result = result.Cons(v.visitAttribute(a))
			}
			return result.Cons(v.make("h" + level))
		}
		return sx.MakeList(v.make("h"+level), sx.String("<MISSING TEXT>"))
	})
	v.bind(sz.NameSymThematic, 0, func(args []sx.Object) sx.Object {
		result := sx.Nil()
		if len(args) > 0 {
			if attrList := v.getList(v.visit(args[0])); attrList != nil {
				result = result.Cons(v.visitAttribute(sz.GetAttributes(attrList)))
			}
		}
		return result.Cons(v.make("hr"))
	})

	v.bind(sz.NameSymListOrdered, 0, v.makeListFn("ol"))
	v.bind(sz.NameSymListUnordered, 0, v.makeListFn("ul"))
	v.bind(sz.NameSymDescription, 0, func(args []sx.Object) sx.Object {
		if len(args) == 0 {
			return sx.Nil()
		}
		items := sx.Nil().Cons(v.make("dl"))
		curItem := items
		for pos := 0; pos < len(args); pos++ {
			term := v.getList(v.visit(args[pos]))
			curItem = curItem.AppendBang(term.Cons(v.make("dt")))
			pos++
			if pos >= len(args) {
				break
			}
			ddBlock := v.getList(v.visit(args[pos]))
			if ddBlock == nil {
				continue
			}
			for ddlst := ddBlock; ddlst != nil; ddlst = ddlst.Tail() {
				dditem := v.getList(ddlst.Car())
				curItem = curItem.AppendBang(dditem.Cons(v.make("dd")))
			}
		}
		return items
	})
	v.bind(sz.NameSymListQuote, 0, func(args []sx.Object) sx.Object {
		if args == nil {
			return sx.Nil()
		}
		result := sx.Nil().Cons(v.make("blockquote"))
		currResult := result
		for _, elem := range args {
			if quote, isPair := sx.GetPair(v.visit(elem)); isPair {
				currResult = currResult.AppendBang(quote.Cons(v.symP))
			}
		}
		return result
	})

	v.bind(sz.NameSymTable, 1, func(args []sx.Object) sx.Object {
		thead := sx.Nil()
		if header := v.getList(v.visit(args[0])); !sx.IsNil(header) {
			thead = sx.Nil().Cons(v.transformTableRow(header)).Cons(v.make("thead"))
		}

		tbody := sx.Nil()
		if len(args) > 1 {
			tbody = sx.Nil().Cons(v.make("tbody"))
			curBody := tbody
			for _, row := range args[1:] {
				curBody = curBody.AppendBang(v.transformTableRow(v.getList(v.visit(row))))
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
		return table.Cons(v.make("table"))
	})
	v.bind(sz.NameSymCell, 0, v.makeCellFn(""))
	v.bind(sz.NameSymCellCenter, 0, v.makeCellFn("center"))
	v.bind(sz.NameSymCellLeft, 0, v.makeCellFn("left"))
	v.bind(sz.NameSymCellRight, 0, v.makeCellFn("right"))

	symDiv := v.make("div")
	v.bind(sz.NameSymRegionBlock, 2, v.makeRegionFn(symDiv, true))
	v.bind(sz.NameSymRegionQuote, 2, v.makeRegionFn(v.make("blockquote"), false))
	v.bind(sz.NameSymRegionVerse, 2, v.makeRegionFn(symDiv, false))

	v.bind(sz.NameSymVerbatimComment, 1, func(args []sx.Object) sx.Object {
		if v.getAttributes(args[0]).HasDefault() {
			if len(args) > 1 {
				if s := v.getString(args[1]); s != "" {
					t := sx.String(s.String())
					return sx.Nil().Cons(t).Cons(v.make(sxhtml.NameSymBlockComment))
				}
			}
		}
		return nil
	})
	v.bind(sz.NameSymVerbatimEval, 2, func(args []sx.Object) sx.Object {
		return v.visitVerbatim(v.getAttributes(args[0]).AddClass("zs-eval"), v.getString(args[1]))
	})
	v.bind(sz.NameSymVerbatimHTML, 2, v.visitHTML)
	v.bind(sz.NameSymVerbatimMath, 2, func(args []sx.Object) sx.Object {
		return v.visitVerbatim(v.getAttributes(args[0]).AddClass("zs-math"), v.getString(args[1]))
	})
	v.bind(sz.NameSymVerbatimProg, 2, func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		content := v.getString(args[1])
		if a.HasDefault() {
			content = sx.String(visibleReplacer.Replace(content.String()))
		}
		return v.visitVerbatim(a, content)
	})
	v.bind(sz.NameSymVerbatimZettel, 0, noopFn)
	v.bind(sz.NameSymBLOB, 3, func(args []sx.Object) sx.Object {
		return v.visitBLOB(v.getList(args[0]), v.getString(args[1]), v.getString(args[2]))
	})
	v.bind(sz.NameSymTransclude, 2, func(args []sx.Object) sx.Object {
		ref, isPair := sx.GetPair(v.visit(args[1]))
		if !isPair {
			return sx.Nil()
		}
		refKind := ref.Car()
		if sx.IsNil(refKind) {
			return sx.Nil()
		}
		if refValue := v.getString(ref.Tail().Car()); refValue != "" {
			if refSym, isRefSym := sx.GetSymbol(refKind); isRefSym && refSym.Name() == sz.NameSymRefStateExternal {
				a := v.getAttributes(args[0]).Set("src", refValue.String()).AddClass("external")
				return sx.Nil().Cons(sx.Nil().Cons(v.visitAttribute(a)).Cons(v.make("img"))).Cons(v.symP)
			}
			return sx.MakeList(
				v.make(sxhtml.NameSymInlineComment),
				sx.String("transclude"),
				refKind,
				sx.String("->"),
				refValue,
			)
		}
		return v.visitSlice(args)
	})
}

func (v *visitor) makeListFn(tag string) transformFn {
	sym := v.make(tag)
	symLI := v.make("li")
	return func(args []sx.Object) sx.Object {
		result := sx.Nil().Cons(sym)
		last := result
		for _, elem := range args {
			item := sx.Nil().Cons(symLI)
			if res, isPair := sx.GetPair(v.visit(elem)); isPair {
				item.ExtendBang(res)
			}
			last = last.AppendBang(item)
		}
		return result
	}
}

func (v *visitor) transformTableRow(pairs *sx.Pair) *sx.Pair {
	row := sx.Nil().Cons(v.make("tr"))
	if pairs == nil {
		return nil
	}
	curRow := row
	for pair := pairs; pair != nil; pair = pair.Tail() {
		curRow = curRow.AppendBang(pair.Car())
	}
	return row
}
func (v *visitor) makeCellFn(align string) transformFn {
	return func(args []sx.Object) sx.Object {
		tdata := v.visitSlice(args)
		if align != "" {
			tdata = tdata.Cons(v.visitAttribute(attrs.Attributes{"class": align}))
		}
		return tdata.Cons(v.make("td"))
	}
}

func (v *visitor) makeRegionFn(sym *sx.Symbol, genericToClass bool) transformFn {
	return func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		if genericToClass {
			if val, found := a.Get(""); found {
				a = a.Remove("").AddClass(val)
			}
		}
		result := sx.Nil()
		if len(a) > 0 {
			result = result.Cons(v.visitAttribute(a))
		}
		result = result.Cons(sym)
		currResult := result.LastPair()
		if region, isPair := sx.GetPair(v.visit(args[1])); isPair {
			currResult = currResult.ExtendBang(region)
		}
		if len(args) > 2 {
			if cite, isPair := sx.GetPair(v.visit(args[2])); isPair && cite != nil {
				currResult.AppendBang(cite.Cons(v.make("cite")))
			}
		}
		return result
	}
}

func (v *visitor) visitVerbatim(a attrs.Attributes, s sx.String) sx.Object {
	a = setProgLang(a)
	code := sx.Nil().Cons(s)
	if al := v.visitAttribute(a); al != nil {
		code = code.Cons(al)
	}
	code = code.Cons(v.make("code"))
	return sx.Nil().Cons(code).Cons(v.make("pre"))
}

func (v *visitor) bindInlines() {
	v.bind(sz.NameSymInline, 0, v.visitList)
	v.bind(sz.NameSymText, 1, func(args []sx.Object) sx.Object { return v.getString(args[0]) })
	v.bind(sz.NameSymSpace, 0, func(args []sx.Object) sx.Object {
		if len(args) == 0 {
			return sx.String(" ")
		}
		return v.getString(args[0])
	})
	v.bind(sz.NameSymSoft, 0, func([]sx.Object) sx.Object { return sx.String(" ") })
	brSym := v.make("br")
	v.bind(sz.NameSymHard, 0, func([]sx.Object) sx.Object { return sx.Nil().Cons(brSym) })

	v.bind(sz.NameSymLinkInvalid, 2, func(args []sx.Object) sx.Object {
		// a := te.getAttributes(args)
		var inline *sx.Pair
		if len(args) > 2 {
			inline = v.visitSlice(args[2:])
		}
		if inline == nil {
			inline = sx.Nil().Cons(v.visit(args[1]))
		}
		return inline.Cons(v.symSpan)
	})
	transformHREF := func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		refValue := v.getString(args[1])
		return v.visitLink(a.Set("href", refValue.String()), refValue, args[2:])
	}
	v.bind(sz.NameSymLinkZettel, 2, transformHREF)
	v.bind(sz.NameSymLinkSelf, 2, transformHREF)
	v.bind(sz.NameSymLinkFound, 2, transformHREF)
	v.bind(sz.NameSymLinkBroken, 2, func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		refValue := v.getString(args[1])
		return v.visitLink(a.AddClass("broken"), refValue, args[2:])
	})
	v.bind(sz.NameSymLinkHosted, 2, transformHREF)
	v.bind(sz.NameSymLinkBased, 2, transformHREF)
	v.bind(sz.NameSymLinkQuery, 2, func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		refValue := v.getString(args[1])
		query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue.String())
		return v.visitLink(a.Set("href", query), refValue, args[2:])
	})
	v.bind(sz.NameSymLinkExternal, 2, func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		refValue := v.getString(args[1])
		return v.visitLink(a.Set("href", refValue.String()).AddClass("external"), refValue, args[2:])
	})

	v.bind(sz.NameSymEmbed, 3, func(args []sx.Object) sx.Object {
		ref := v.getList(v.visit(args[1]))
		syntax := v.getString(args[2])
		if syntax == api.ValueSyntaxSVG {
			embedAttr := sx.MakeList(
				v.symAttr,
				sx.Cons(v.make("type"), sx.String("image/svg+xml")),
				sx.Cons(v.make("src"), sx.String("/"+v.getString(ref.Tail()).String()+".svg")),
			)
			return sx.MakeList(
				v.make("figure"),
				sx.MakeList(
					v.make("embed"),
					embedAttr,
				),
			)
		}
		a := v.getAttributes(args[0])
		a = a.Set("src", string(v.getString(ref.Tail().Car())))
		var sb strings.Builder
		v.flattenText(&sb, ref.Tail().Tail().Tail())
		if d := sb.String(); d != "" {
			a = a.Set("alt", d)
		}
		return sx.MakeList(v.make("img"), v.visitAttribute(a))
	})
	v.bind(sz.NameSymEmbedBLOB, 3, noopFn)

	v.bind(sz.NameSymCite, 2, func(args []sx.Object) sx.Object {
		result := sx.Nil()
		if key := v.getString(args[1]); key != "" {
			if len(args) > 2 {
				result = v.visitSlice(args[2:]).Cons(sx.String(", "))
			}
			result = result.Cons(key)
		}
		if a := v.getAttributes(args[0]); len(a) > 0 {
			result = result.Cons(v.visitAttribute(a))
		}
		if result == nil {
			return nil
		}
		return result.Cons(v.symSpan)
	})
	v.bind(sz.NameSymMark, 3, func(args []sx.Object) sx.Object {
		result := v.visitSlice(args[3:])
		if !v.tr.noLinks {
			if fragment := v.getString(args[2]); fragment != "" {
				a := attrs.Attributes{"id": fragment.String() + v.tr.unique}
				return result.Cons(v.visitAttribute(a)).Cons(v.symA)
			}
		}
		return result.Cons(v.symSpan)
	})
	v.bind(sz.NameSymEndnote, 1, func(args []sx.Object) sx.Object {
		attrPlist := sx.Nil()
		if a := v.getAttributes(args[0]); len(a) > 0 {
			if attrs := v.visitAttribute(a); attrs != nil {
				attrPlist = attrs.Tail()
			}
		}

		text, isPair := sx.GetPair(args[1])
		if !isPair {
			return sx.Nil()
		}
		v.tr.endnotes = append(v.tr.endnotes, endnoteInfo{noteAST: text, noteHx: nil, attrs: attrPlist})
		noteNum := strconv.Itoa(len(v.tr.endnotes))
		noteID := v.tr.unique + noteNum
		hrefAttr := sx.Nil().Cons(sx.Cons(v.make("role"), sx.String("doc-noteref"))).
			Cons(sx.Cons(v.make("href"), sx.String("#fn:"+noteID))).
			Cons(sx.Cons(v.tr.symClass, sx.String("zs-noteref"))).
			Cons(v.symAttr)
		href := sx.Nil().Cons(sx.String(noteNum)).Cons(hrefAttr).Cons(v.symA)
		supAttr := sx.Nil().Cons(sx.Cons(v.make("id"), sx.String("fnref:"+noteID))).Cons(v.symAttr)
		return sx.Nil().Cons(href).Cons(supAttr).Cons(v.make("sup"))
	})

	v.bind(sz.NameSymFormatDelete, 1, v.makeFormatFn("del"))
	v.bind(sz.NameSymFormatEmph, 1, v.makeFormatFn("em"))
	v.bind(sz.NameSymFormatInsert, 1, v.makeFormatFn("ins"))
	v.bind(sz.NameSymFormatQuote, 1, func(args []sx.Object) sx.Object {
		const langAttr = "lang"
		a := v.getAttributes(args[0])
		langVal, found := a.Get(langAttr)
		if found {
			a = a.Remove(langAttr)
		}
		if val, found2 := a.Get(""); found2 {
			a = a.Remove("").AddClass(val)
		}
		res := v.visitSlice(args[1:])
		if len(a) > 0 {
			res = res.Cons(v.visitAttribute(a))
		}
		res = res.Cons(v.make("q"))
		if found {
			res = sx.Nil().Cons(res).Cons(v.visitAttribute(attrs.Attributes{}.Set(langAttr, langVal))).Cons(v.symSpan)
		}
		return res
	})
	v.bind(sz.NameSymFormatSpan, 1, v.makeFormatFn("span"))
	v.bind(sz.NameSymFormatStrong, 1, v.makeFormatFn("strong"))
	v.bind(sz.NameSymFormatSub, 1, v.makeFormatFn("sub"))
	v.bind(sz.NameSymFormatSuper, 1, v.makeFormatFn("sup"))

	v.bind(sz.NameSymLiteralComment, 1, func(args []sx.Object) sx.Object {
		if v.getAttributes(args[0]).HasDefault() {
			if len(args) > 1 {
				if s := v.getString(v.visit(args[1])); s != "" {
					return sx.Nil().Cons(s).Cons(v.make(sxhtml.NameSymInlineComment))
				}
			}
		}
		return sx.Nil()
	})
	v.bind(sz.NameSymLiteralHTML, 2, v.visitHTML)
	kbdSym := v.make("kbd")
	v.bind(sz.NameSymLiteralInput, 2, func(args []sx.Object) sx.Object {
		return v.visitLiteral(args, nil, kbdSym)
	})
	codeSym := v.make("code")
	v.bind(sz.NameSymLiteralMath, 2, func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0]).AddClass("zs-math")
		return v.visitLiteral(args, a, codeSym)
	})
	sampSym := v.make("samp")
	v.bind(sz.NameSymLiteralOutput, 2, func(args []sx.Object) sx.Object {
		return v.visitLiteral(args, nil, sampSym)
	})
	v.bind(sz.NameSymLiteralProg, 2, func(args []sx.Object) sx.Object {
		return v.visitLiteral(args, nil, codeSym)
	})

	v.bind(sz.NameSymLiteralZettel, 0, noopFn)
}

func (v *visitor) makeFormatFn(tag string) transformFn {
	sym := v.make(tag)
	return func(args []sx.Object) sx.Object {
		a := v.getAttributes(args[0])
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		res := v.visitSlice(args[1:])
		if len(a) > 0 {
			res = res.Cons(v.visitAttribute(a))
		}
		return res.Cons(sym)
	}
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func (te *visitor) visitLiteral(args []sx.Object, a attrs.Attributes, sym *sx.Symbol) sx.Object {
	if a == nil {
		a = te.getAttributes(args[0])
	}
	a = setProgLang(a)
	literal := te.getString(args[1]).String()
	if a.HasDefault() {
		a = a.RemoveDefault()
		literal = visibleReplacer.Replace(literal)
	}
	res := sx.Nil().Cons(sx.String(literal))
	if len(a) > 0 {
		res = res.Cons(te.visitAttribute(a))
	}
	return res.Cons(sym)
}
func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}

func (v *visitor) visitHTML(args []sx.Object) sx.Object {
	if s := v.getString(v.visit(args[1])); s != "" && IsSafe(s.String()) {
		return sx.Nil().Cons(s).Cons(v.symNoEscape)
	}
	return nil
}

func (te *visitor) visitBLOB(description *sx.Pair, syntax, data sx.String) sx.Object {
	if data == "" {
		return sx.Nil()
	}
	switch syntax {
	case "":
		return sx.Nil()
	case api.ValueSyntaxSVG:
		return sx.Nil().Cons(sx.Nil().Cons(data).Cons(te.symNoEscape)).Cons(te.symP)
	default:
		imgAttr := sx.Nil().Cons(sx.Cons(te.make("src"), sx.String("data:image/"+syntax.String()+";base64,"+data.String())))
		var sb strings.Builder
		te.flattenText(&sb, description)
		if d := sb.String(); d != "" {
			imgAttr = imgAttr.Cons(sx.Cons(te.make("alt"), sx.String(d)))
		}
		return sx.Nil().Cons(sx.Nil().Cons(imgAttr.Cons(te.symAttr)).Cons(te.make("img"))).Cons(te.symP)
	}
}

func (v *visitor) flattenText(sb *strings.Builder, lst *sx.Pair) {
	for elem := lst; elem != nil; elem = elem.Tail() {
		switch obj := elem.Car().(type) {
		case sx.String:
			sb.WriteString(obj.String())
		case *sx.Pair:
			v.flattenText(sb, obj)
		}
	}
}

func (v *visitor) visitList(args []sx.Object) sx.Object { return v.visitSlice(args) }
func noopFn([]sx.Object) sx.Object                      { return sx.Nil() }

func (v *visitor) visit(obj sx.Object) sx.Object {
	if v.err != nil {
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
		v.err = fmt.Errorf("symbol expected, but got %T/%v", lst.Car(), lst.Car())
		return sx.Nil()
	}
	name := sym.Name()
	fn, found := v.fns[name]
	if !found {
		v.err = fmt.Errorf("symbol %q not bound", name)
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
	if minArgs, hasMinArgs := v.minArgs[name]; hasMinArgs {
		if minArgs > len(args) {
			v.err = fmt.Errorf("%v needs at least %d arguments, but got only %d", name, minArgs, len(args))
			return sx.Nil()
		}
	}
	log.Println("EXEC", sym, args)
	result := fn(args)
	if v.err != nil {
		log.Println("EERR", v.err)
		return sx.Nil()
	}
	log.Println("REXE", result)
	return result
}

func (v *visitor) visitSlice(args []sx.Object) *sx.Pair {
	result := sx.Cons(sx.Nil(), sx.Nil())
	curr := result
	for _, arg := range args {
		elem := v.visit(arg)
		if v.err != nil {
			return nil
		}
		curr = curr.AppendBang(elem)
	}
	return result.Tail()
}

func (v *visitor) visitLink(a attrs.Attributes, refValue sx.String, inline []sx.Object) sx.Object {
	result := v.visitSlice(inline)
	if len(inline) == 0 {
		result = sx.Nil().Cons(refValue)
	}
	if v.tr.noLinks {
		return result.Cons(v.symSpan)
	}
	return result.Cons(v.visitAttribute(a)).Cons(v.symA)
}

func (v *visitor) visitAttribute(a attrs.Attributes) *sx.Pair { return v.tr.TransformAttrbute(a) }

func (v *visitor) getSymbol(val sx.Object) *sx.Symbol {
	if v.err == nil {
		if sym, ok := sx.GetSymbol(val); ok {
			return sym
		}
		v.err = fmt.Errorf("%v/%T is not a symbol", val, val)
	}
	return v.make("???")
}
func (v *visitor) getString(val sx.Object) sx.String {
	if v.err != nil {
		return ""
	}
	if s, ok := sx.GetString(val); ok {
		return s
	}
	v.err = fmt.Errorf("%v/%T is not a string", val, val)
	return ""
}
func (v *visitor) getList(val sx.Object) *sx.Pair {
	if v.err == nil {
		if res, isPair := sx.GetPair(val); isPair {
			return res
		}
		v.err = fmt.Errorf("%v/%T is not a list", val, val)
	}
	return nil
}
func (v *visitor) getInt64(val sx.Object) int64 {
	if v.err != nil {
		return -1017
	}
	if num, ok := sx.GetNumber(val); ok {
		return int64(num.(sx.Int64))
	}
	v.err = fmt.Errorf("%v/%T is not a number", val, val)
	return -1017
}
func (v *visitor) getAttributes(arg sx.Object) attrs.Attributes {
	return sz.GetAttributes(v.getList(v.visit(arg)))
}

func (v *visitor) make(name string) *sx.Symbol { return v.sf.MustMake(name) }

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
