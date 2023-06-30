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

	"codeberg.org/t73fde/sxhtml"
	"codeberg.org/t73fde/sxpf"
	"codeberg.org/t73fde/sxpf/builtins/quote"
	"codeberg.org/t73fde/sxpf/eval"
	"zettelstore.de/c/api"
	"zettelstore.de/c/attrs"
	"zettelstore.de/c/sz"
	"zettelstore.de/c/text"
)

// Transformer will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Transformer struct {
	sf            sxpf.SymbolFactory
	rebinder      RebindProc
	headingOffset int64
	unique        string
	endnotes      []endnoteInfo
	noLinks       bool // true iff output must not include links
	symAttr       *sxpf.Symbol
	symClass      *sxpf.Symbol
	symMeta       *sxpf.Symbol
	symA          *sxpf.Symbol
	symSpan       *sxpf.Symbol
}

type endnoteInfo struct {
	noteAST *sxpf.Cell // Endnote as AST
	noteHx  *sxpf.Cell // Endnote as SxHTML
	attrs   *sxpf.Cell // attrs a-list
}

// NewTransformer creates a new transformer object.
func NewTransformer(headingOffset int, sf sxpf.SymbolFactory) *Transformer {
	if sf == nil {
		sf = sxpf.MakeMappedFactory()
	}
	return &Transformer{
		sf:            sf,
		rebinder:      nil,
		headingOffset: int64(headingOffset),
		symAttr:       sf.MustMake(sxhtml.NameSymAttr),
		symClass:      sf.MustMake("class"),
		symMeta:       sf.MustMake("meta"),
		symA:          sf.MustMake("a"),
		symSpan:       sf.MustMake("span"),
	}
}

// SymbolFactory returns the symbol factory to create HTML symbols.
func (tr *Transformer) SymbolFactory() sxpf.SymbolFactory { return tr.sf }

// SetUnique sets a prefix to make several HTML ids unique.
func (tr *Transformer) SetUnique(s string) { tr.unique = s }

// IsValidName returns true, if name is a valid symbol name.
func (tr *Transformer) IsValidName(s string) bool { return tr.sf.IsValidName(s) }

// Make a new HTML symbol.
func (tr *Transformer) Make(s string) *sxpf.Symbol { return tr.sf.MustMake(s) }

// RebindProc is a procedure which is called every time before a tranformation takes place.
type RebindProc func(*TransformEnv)

// SetRebinder sets the rebinder procedure.
func (tr *Transformer) SetRebinder(rb RebindProc) { tr.rebinder = rb }

// TransformAttrbute transforms the given attributes into a HTML s-expression.
func (tr *Transformer) TransformAttrbute(a attrs.Attributes) *sxpf.Cell {
	if len(a) == 0 {
		return nil
	}
	plist := sxpf.Nil()
	keys := a.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		if key != attrs.DefaultAttribute && tr.IsValidName(key) {
			plist = plist.Cons(sxpf.Cons(tr.Make(key), sxpf.MakeString(a[key])))
		}
	}
	if plist == nil {
		return nil
	}
	return plist.Cons(tr.symAttr)
}

// TransformMeta creates a HTML meta s-expression
func (tr *Transformer) TransformMeta(a attrs.Attributes) *sxpf.Cell {
	return sxpf.Nil().Cons(tr.TransformAttrbute(a)).Cons(tr.symMeta)
}

// Transform an AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) Transform(lst *sxpf.Cell) (*sxpf.Cell, error) {
	astSF := sxpf.FindSymbolFactory(lst)
	if astSF != nil {
		if astSF == tr.sf {
			panic("Invalid AST SymbolFactory")
		}
	} else {
		astSF = sxpf.MakeMappedFactory()
	}
	astEnv := sxpf.MakeRootEnvironment()
	engine := eval.MakeEngine(astSF, astEnv)
	quote.InstallQuoteSyntax(astEnv, astSF.MustMake(sz.NameSymQuote))
	te := TransformEnv{
		tr:      tr,
		astSF:   astSF,
		astEnv:  astEnv,
		err:     nil,
		textEnc: text.NewEncoder(astSF),
	}
	te.initialize()
	if rb := tr.rebinder; rb != nil {
		rb(&te)
	}

	val, err := engine.Eval(te.astEnv, lst)
	if err != nil {
		return nil, err
	}
	res, ok := val.(*sxpf.Cell)
	if !ok {
		panic("Result is not a list")
	}
	for i := 0; i < len(tr.endnotes); i++ {
		// May extend tr.endnotes
		val, err = engine.Eval(te.astEnv, tr.endnotes[i].noteAST)
		if err != nil {
			return res, err
		}
		en, ok2 := val.(*sxpf.Cell)
		if !ok2 {
			panic("Endnote is not a list")
		}
		tr.endnotes[i].noteHx = en
	}
	return res, err

}

// Endnotes returns a SHTML object with all collected endnotes.
func (tr *Transformer) Endnotes() *sxpf.Cell {
	if len(tr.endnotes) == 0 {
		return nil
	}
	result := sxpf.Nil().Cons(tr.Make("ol"))
	currResult := result.AppendBang(sxpf.Nil().Cons(sxpf.Cons(tr.symClass, sxpf.MakeString("zs-endnotes"))).Cons(tr.symAttr))
	for i, fni := range tr.endnotes {
		noteNum := strconv.Itoa(i + 1)
		noteID := tr.unique + noteNum

		attrs := fni.attrs.Cons(sxpf.Cons(tr.symClass, sxpf.MakeString("zs-endnote"))).
			Cons(sxpf.Cons(tr.Make("value"), sxpf.MakeString(noteNum))).
			Cons(sxpf.Cons(tr.Make("id"), sxpf.MakeString("fn:"+noteID))).
			Cons(sxpf.Cons(tr.Make("role"), sxpf.MakeString("doc-endnote"))).
			Cons(tr.symAttr)

		backref := sxpf.Nil().Cons(sxpf.MakeString("\u21a9\ufe0e")).
			Cons(sxpf.Nil().
				Cons(sxpf.Cons(tr.symClass, sxpf.MakeString("zs-endnote-backref"))).
				Cons(sxpf.Cons(tr.Make("href"), sxpf.MakeString("#fnref:"+noteID))).
				Cons(sxpf.Cons(tr.Make("role"), sxpf.MakeString("doc-backlink"))).
				Cons(tr.symAttr)).
			Cons(tr.symA)

		li := sxpf.Nil().Cons(tr.Make("li"))
		li.AppendBang(attrs).
			ExtendBang(fni.noteHx).
			AppendBang(sxpf.MakeString(" ")).AppendBang(backref)
		currResult = currResult.AppendBang(li)
	}
	tr.endnotes = nil
	return result
}

// TransformEnv is the environment where the actual transformation takes places.
type TransformEnv struct {
	tr          *Transformer
	astSF       sxpf.SymbolFactory
	astEnv      sxpf.Environment
	err         error
	textEnc     *text.Encoder
	symNoEscape *sxpf.Symbol
	symAttr     *sxpf.Symbol
	symA        *sxpf.Symbol
	symSpan     *sxpf.Symbol
	symP        *sxpf.Symbol
}

func (te *TransformEnv) initialize() {
	te.symNoEscape = te.Make(sxhtml.NameSymNoEscape)
	te.symAttr = te.tr.symAttr
	te.symA = te.tr.symA
	te.symSpan = te.tr.symSpan
	te.symP = te.Make("p")

	te.bind(sz.NameSymList, 0, listArgs)
	te.bindMetadata()
	te.bindBlocks()
	te.bindInlines()
}

func listArgs(args []sxpf.Object) sxpf.Object { return sxpf.MakeList(args...) }

func (te *TransformEnv) bindMetadata() {
	te.bind(sz.NameSymMeta, 0, listArgs)
	te.bind(sz.NameSymTypeZettelmarkup, 2, func(args []sxpf.Object) sxpf.Object {
		a := make(attrs.Attributes, 2).
			Set("name", te.getSymbol(args[0]).String()).
			Set("content", te.textEnc.Encode(te.getList(args[1])))
		return te.transformMeta(a)
	})
	metaString := func(args []sxpf.Object) sxpf.Object {
		a := make(attrs.Attributes, 2).
			Set("name", te.getSymbol(args[0]).Name()).
			Set("content", te.getString(args[1]).String())
		return te.transformMeta(a)
	}
	te.bind(sz.NameSymTypeCredential, 2, metaString)
	te.bind(sz.NameSymTypeEmpty, 2, metaString)
	te.bind(sz.NameSymTypeID, 2, metaString)
	te.bind(sz.NameSymTypeNumber, 2, metaString)
	te.bind(sz.NameSymTypeString, 2, metaString)
	te.bind(sz.NameSymTypeTimestamp, 2, metaString)
	te.bind(sz.NameSymTypeURL, 2, metaString)
	te.bind(sz.NameSymTypeWord, 2, metaString)
	metaSet := func(args []sxpf.Object) sxpf.Object {
		var sb strings.Builder
		for elem := te.getList(args[1]); elem != nil; elem = elem.Tail() {
			sb.WriteByte(' ')
			sb.WriteString(te.getString(elem.Car()).String())
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", te.getSymbol(args[0]).Name()).
			Set("content", s)
		return te.transformMeta(a)
	}
	te.bind(sz.NameSymTypeIDSet, 2, metaSet)
	te.bind(sz.NameSymTypeTagSet, 2, metaSet)
	te.bind(sz.NameSymTypeWordSet, 2, metaSet)
}

func (te *TransformEnv) bindBlocks() {
	te.bind(sz.NameSymBlock, 0, listArgs)
	te.bind(sz.NameSymPara, 0, func(args []sxpf.Object) sxpf.Object {
		// for ; args != nil; args = args.Tail() {
		// 	lst, ok := sxpf.GetList(args.Car())
		// 	if !ok || lst != nil {
		// 		break
		// 	}
		// }
		return sxpf.MakeList(args...).Cons(te.symP)
	})
	te.bind(sz.NameSymHeading, 5, func(args []sxpf.Object) sxpf.Object {
		nLevel := te.getInt64(args[0])
		if nLevel <= 0 {
			te.err = fmt.Errorf("%v is a negative level", nLevel)
			return sxpf.Nil()
		}
		level := strconv.FormatInt(nLevel+te.tr.headingOffset, 10)

		a := te.getAttributes(args[1])
		if fragment := te.getString(args[3]).String(); fragment != "" {
			a = a.Set("id", te.tr.unique+fragment)
		}

		if result, isCell := sxpf.GetCell(args[4]); isCell && result != nil {
			if len(a) > 0 {
				result = result.Cons(te.transformAttribute(a))
			}
			return result.Cons(te.Make("h" + level))
		}
		return sxpf.MakeList(te.Make("h"+level), sxpf.MakeString("<MISSING TEXT>"))
	})
	te.bind(sz.NameSymThematic, 0, func(args []sxpf.Object) sxpf.Object {
		result := sxpf.Nil()
		if len(args) > 0 {
			if attrList := te.getList(args[0]); attrList != nil {
				result = result.Cons(te.transformAttribute(sz.GetAttributes(attrList)))
			}
		}
		return result.Cons(te.Make("hr"))
	})
	te.bind(sz.NameSymListOrdered, 0, te.makeListFn("ol"))
	te.bind(sz.NameSymListUnordered, 0, te.makeListFn("ul"))
	te.bind(sz.NameSymDescription, 0, func(args []sxpf.Object) sxpf.Object {
		if len(args) == 0 {
			return sxpf.Nil()
		}
		items := sxpf.Nil().Cons(te.Make("dl"))
		curItem := items
		for pos := 0; pos < len(args); pos++ {
			term := te.getList(args[pos])
			curItem = curItem.AppendBang(term.Cons(te.Make("dt")))
			pos++
			if pos >= len(args) {
				break
			}
			ddBlock := te.getList(args[pos])
			if ddBlock == nil {
				break
			}
			for ddlst := ddBlock; ddlst != nil; ddlst = ddlst.Tail() {
				dditem := te.getList(ddlst.Car())
				curItem = curItem.AppendBang(dditem.Cons(te.Make("dd")))
			}
		}
		return items
	})

	te.bind(sz.NameSymListQuote, 0, func(args []sxpf.Object) sxpf.Object {
		if args == nil {
			return sxpf.Nil()
		}
		result := sxpf.Nil().Cons(te.Make("blockquote"))
		currResult := result
		for _, elem := range args {
			if quote, isCell := sxpf.GetCell(elem); isCell {
				currResult = currResult.AppendBang(quote.Cons(te.symP))
			}
		}
		return result
	})

	te.bind(sz.NameSymTable, 1, func(args []sxpf.Object) sxpf.Object {
		thead := sxpf.Nil()
		if header := te.getList(args[0]); !sxpf.IsNil(header) {
			thead = sxpf.Nil().Cons(te.transformTableRow(header)).Cons(te.Make("thead"))
		}

		tbody := sxpf.Nil()
		if len(args) > 1 {
			tbody = sxpf.Nil().Cons(te.Make("tbody"))
			curBody := tbody
			for _, row := range args[1:] {
				curBody = curBody.AppendBang(te.transformTableRow(te.getList(row)))
			}
		}

		table := sxpf.Nil()
		if tbody != nil {
			table = table.Cons(tbody)
		}
		if thead != nil {
			table = table.Cons(thead)
		}
		if table == nil {
			return sxpf.Nil()
		}
		return table.Cons(te.Make("table"))
	})
	te.bind(sz.NameSymCell, 0, te.makeCellFn(""))
	te.bind(sz.NameSymCellCenter, 0, te.makeCellFn("center"))
	te.bind(sz.NameSymCellLeft, 0, te.makeCellFn("left"))
	te.bind(sz.NameSymCellRight, 0, te.makeCellFn("right"))

	te.bind(sz.NameSymRegionBlock, 2, te.makeRegionFn(te.Make("div"), true))
	te.bind(sz.NameSymRegionQuote, 2, te.makeRegionFn(te.Make("blockquote"), false))
	te.bind(sz.NameSymRegionVerse, 2, te.makeRegionFn(te.Make("div"), false))

	te.bind(sz.NameSymVerbatimComment, 1, func(args []sxpf.Object) sxpf.Object {
		if te.getAttributes(args[0]).HasDefault() {
			if len(args) > 1 {
				if s := te.getString(args[1]); s != "" {
					t := sxpf.MakeString(s.String())
					return sxpf.Nil().Cons(t).Cons(te.Make(sxhtml.NameSymBlockComment))
				}
			}
		}
		return nil
	})

	te.bind(sz.NameSymVerbatimEval, 2, func(args []sxpf.Object) sxpf.Object {
		return te.transformVerbatim(te.getAttributes(args[0]).AddClass("zs-eval"), te.getString(args[1]))
	})
	te.bind(sz.NameSymVerbatimHTML, 2, te.transformHTML)
	te.bind(sz.NameSymVerbatimMath, 2, func(args []sxpf.Object) sxpf.Object {
		return te.transformVerbatim(te.getAttributes(args[0]).AddClass("zs-math"), te.getString(args[1]))
	})
	te.bind(sz.NameSymVerbatimProg, 2, func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		content := te.getString(args[1])
		if a.HasDefault() {
			content = sxpf.MakeString(visibleReplacer.Replace(content.String()))
		}
		return te.transformVerbatim(a, content)
	})
	te.bind(sz.NameSymVerbatimZettel, 0, func([]sxpf.Object) sxpf.Object { return sxpf.Nil() })

	te.bind(sz.NameSymBLOB, 3, func(args []sxpf.Object) sxpf.Object {
		return te.transformBLOB(te.getList(args[0]), te.getString(args[1]), te.getString(args[2]))
	})

	te.bind(sz.NameSymTransclude, 2, func(args []sxpf.Object) sxpf.Object {
		ref, isCell := sxpf.GetCell(args[1])
		if !isCell {
			return sxpf.Nil()
		}
		refKind := ref.Car()
		if sxpf.IsNil(refKind) {
			return sxpf.Nil()
		}
		if refValue := te.getString(ref.Tail().Car()); refValue != "" {
			if te.astSF.MustMake(sz.NameSymRefStateExternal).IsEqual(refKind) {
				a := te.getAttributes(args[0]).Set("src", refValue.String()).AddClass("external")
				return sxpf.Nil().Cons(sxpf.Nil().Cons(te.transformAttribute(a)).Cons(te.Make("img"))).Cons(te.symP)
			}
			return sxpf.MakeList(
				te.Make(sxhtml.NameSymInlineComment),
				sxpf.MakeString("transclude"),
				refKind,
				sxpf.MakeString("->"),
				refValue,
			)
		}
		return sxpf.MakeList(args...)
	})
}

func (te *TransformEnv) makeListFn(tag string) transformFn {
	sym := te.Make(tag)
	return func(args []sxpf.Object) sxpf.Object {
		result := sxpf.Nil().Cons(sym)
		last := result
		for _, elem := range args {
			item := sxpf.Nil().Cons(te.Make("li"))
			if res, isCell := sxpf.GetCell(elem); isCell {
				item.ExtendBang(res)
			}
			last = last.AppendBang(item)
		}
		return result
	}
}
func (te *TransformEnv) transformTableRow(cells *sxpf.Cell) *sxpf.Cell {
	row := sxpf.Nil().Cons(te.Make("tr"))
	if cells == nil {
		return nil
	}
	curRow := row
	for cell := cells; cell != nil; cell = cell.Tail() {
		curRow = curRow.AppendBang(cell.Car())
	}
	return row
}

func (te *TransformEnv) makeCellFn(align string) transformFn {
	return func(args []sxpf.Object) sxpf.Object {
		tdata := sxpf.MakeList(args...)
		if align != "" {
			tdata = tdata.Cons(te.transformAttribute(attrs.Attributes{"class": align}))
		}
		return tdata.Cons(te.Make("td"))
	}
}

func (te *TransformEnv) makeRegionFn(sym *sxpf.Symbol, genericToClass bool) transformFn {
	return func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		if genericToClass {
			if val, found := a.Get(""); found {
				a = a.Remove("").AddClass(val)
			}
		}
		result := sxpf.Nil()
		if len(a) > 0 {
			result = result.Cons(te.transformAttribute(a))
		}
		result = result.Cons(sym)
		currResult := result.LastPair()
		if region, isCell := sxpf.GetCell(args[1]); isCell {
			currResult = currResult.ExtendBang(region)
		}
		if len(args) > 2 {
			if cite, isCell := sxpf.GetCell(args[2]); isCell && cite != nil {
				currResult.AppendBang(cite.Cons(te.Make("cite")))
			}
		}
		return result
	}
}

func (te *TransformEnv) transformVerbatim(a attrs.Attributes, s sxpf.String) sxpf.Object {
	a = setProgLang(a)
	code := sxpf.Nil().Cons(s)
	if al := te.transformAttribute(a); al != nil {
		code = code.Cons(al)
	}
	code = code.Cons(te.Make("code"))
	return sxpf.Nil().Cons(code).Cons(te.Make("pre"))
}

func (te *TransformEnv) bindInlines() {
	te.bind(sz.NameSymInline, 0, listArgs)
	te.bind(sz.NameSymText, 1, func(args []sxpf.Object) sxpf.Object { return te.getString(args[0]) })
	te.bind(sz.NameSymSpace, 0, func(args []sxpf.Object) sxpf.Object {
		if len(args) == 0 {
			return sxpf.MakeString(" ")
		}
		return te.getString(args[0])
	})
	te.bind(sz.NameSymSoft, 0, func([]sxpf.Object) sxpf.Object { return sxpf.MakeString(" ") })
	brSym := te.Make("br")
	te.bind(sz.NameSymHard, 0, func([]sxpf.Object) sxpf.Object { return sxpf.Nil().Cons(brSym) })

	te.bind(sz.NameSymLinkInvalid, 2, func(args []sxpf.Object) sxpf.Object {
		// a := te.getAttributes(args)
		var inline *sxpf.Cell
		if len(args) > 2 {
			inline = sxpf.MakeList(args[2:]...)
		}
		if inline == nil {
			inline = sxpf.Nil().Cons(args[1])
		}
		return inline.Cons(te.symSpan)
	})
	transformHREF := func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		refValue := te.getString(args[1])
		return te.transformLink(a.Set("href", refValue.String()), refValue, args[2:])
	}
	te.bind(sz.NameSymLinkZettel, 2, transformHREF)
	te.bind(sz.NameSymLinkSelf, 2, transformHREF)
	te.bind(sz.NameSymLinkFound, 2, transformHREF)
	te.bind(sz.NameSymLinkBroken, 2, func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		refValue := te.getString(args[1])
		return te.transformLink(a.AddClass("broken"), refValue, args[2:])
	})
	te.bind(sz.NameSymLinkHosted, 2, transformHREF)
	te.bind(sz.NameSymLinkBased, 2, transformHREF)
	te.bind(sz.NameSymLinkQuery, 2, func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		refValue := te.getString(args[1])
		query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue.String())
		return te.transformLink(a.Set("href", query), refValue, args[2:])
	})
	te.bind(sz.NameSymLinkExternal, 2, func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		refValue := te.getString(args[1])
		return te.transformLink(a.Set("href", refValue.String()).AddClass("external"), refValue, args[2:])
	})

	te.bind(sz.NameSymEmbed, 3, func(args []sxpf.Object) sxpf.Object {
		ref := te.getList(args[1])
		syntax := te.getString(args[2])
		if syntax == api.ValueSyntaxSVG {
			embedAttr := sxpf.MakeList(
				te.symAttr,
				sxpf.Cons(te.Make("type"), sxpf.MakeString("image/svg+xml")),
				sxpf.Cons(te.Make("src"), sxpf.MakeString("/"+te.getString(ref.Tail()).String()+".svg")),
			)
			return sxpf.MakeList(
				te.Make("figure"),
				sxpf.MakeList(
					te.Make("embed"),
					embedAttr,
				),
			)
		}
		a := te.getAttributes(args[0])
		a = a.Set("src", string(te.getString(ref.Tail().Car())))
		var sb strings.Builder
		te.flattenText(&sb, ref.Tail().Tail().Tail())
		if d := sb.String(); d != "" {
			a = a.Set("alt", d)
		}
		return sxpf.MakeList(te.Make("img"), te.transformAttribute(a))
	})
	te.bind(sz.NameSymEmbedBLOB, 3, func(args []sxpf.Object) sxpf.Object {
		a, syntax, data := te.getAttributes(args[0]), te.getString(args[1]), te.getString(args[2])
		summary, _ := a.Get(api.KeySummary)
		return te.transformBLOB(
			sxpf.MakeList(te.astSF.MustMake(sz.NameSymInline), sxpf.MakeString(summary)),
			syntax,
			data,
		)
	})

	te.bind(sz.NameSymCite, 2, func(args []sxpf.Object) sxpf.Object {
		result := sxpf.Nil()
		if key := te.getString(args[1]); key != "" {
			if len(args) > 2 {
				result = sxpf.MakeList(args[2:]...).Cons(sxpf.MakeString(", "))
			}
			result = result.Cons(key)
		}
		if a := te.getAttributes(args[0]); len(a) > 0 {
			result = result.Cons(te.transformAttribute(a))
		}
		if result == nil {
			return nil
		}
		return result.Cons(te.symSpan)
	})

	te.bind(sz.NameSymMark, 3, func(args []sxpf.Object) sxpf.Object {
		result := sxpf.MakeList(args[3:]...)
		if !te.tr.noLinks {
			if fragment := te.getString(args[2]); fragment != "" {
				a := attrs.Attributes{"id": fragment.String() + te.tr.unique}
				return result.Cons(te.transformAttribute(a)).Cons(te.symA)
			}
		}
		return result.Cons(te.symSpan)
	})

	te.bind(sz.NameSymEndnote, 1, func(args []sxpf.Object) sxpf.Object {
		attrPlist := sxpf.Nil()
		if a := te.getAttributes(args[0]); len(a) > 0 {
			if attrs := te.transformAttribute(a); attrs != nil {
				attrPlist = attrs.Tail()
			}
		}

		text, isCell := sxpf.GetCell(args[1])
		if !isCell {
			return sxpf.Nil()
		}
		te.tr.endnotes = append(te.tr.endnotes, endnoteInfo{noteAST: text, noteHx: nil, attrs: attrPlist})
		noteNum := strconv.Itoa(len(te.tr.endnotes))
		noteID := te.tr.unique + noteNum
		hrefAttr := sxpf.Nil().Cons(sxpf.Cons(te.Make("role"), sxpf.MakeString("doc-noteref"))).
			Cons(sxpf.Cons(te.Make("href"), sxpf.MakeString("#fn:"+noteID))).
			Cons(sxpf.Cons(te.tr.symClass, sxpf.MakeString("zs-noteref"))).
			Cons(te.symAttr)
		href := sxpf.Nil().Cons(sxpf.MakeString(noteNum)).Cons(hrefAttr).Cons(te.symA)
		supAttr := sxpf.Nil().Cons(sxpf.Cons(te.Make("id"), sxpf.MakeString("fnref:"+noteID))).Cons(te.symAttr)
		return sxpf.Nil().Cons(href).Cons(supAttr).Cons(te.Make("sup"))
	})

	te.bind(sz.NameSymFormatDelete, 1, te.makeFormatFn("del"))
	te.bind(sz.NameSymFormatEmph, 1, te.makeFormatFn("em"))
	te.bind(sz.NameSymFormatInsert, 1, te.makeFormatFn("ins"))
	te.bind(sz.NameSymFormatQuote, 1, te.transformQuote)
	te.bind(sz.NameSymFormatSpan, 1, te.makeFormatFn("span"))
	te.bind(sz.NameSymFormatStrong, 1, te.makeFormatFn("strong"))
	te.bind(sz.NameSymFormatSub, 1, te.makeFormatFn("sub"))
	te.bind(sz.NameSymFormatSuper, 1, te.makeFormatFn("sup"))

	te.bind(sz.NameSymLiteralComment, 1, func(args []sxpf.Object) sxpf.Object {
		if te.getAttributes(args[0]).HasDefault() {
			if len(args) > 1 {
				if s := te.getString(args[1]); s != "" {
					return sxpf.Nil().Cons(s).Cons(te.Make(sxhtml.NameSymInlineComment))
				}
			}
		}
		return sxpf.Nil()
	})
	te.bind(sz.NameSymLiteralHTML, 2, te.transformHTML)
	kbdSym := te.Make("kbd")
	te.bind(sz.NameSymLiteralInput, 2, func(args []sxpf.Object) sxpf.Object {
		return te.transformLiteral(args, nil, kbdSym)
	})
	codeSym := te.Make("code")
	te.bind(sz.NameSymLiteralMath, 2, func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0]).AddClass("zs-math")
		return te.transformLiteral(args, a, codeSym)
	})
	sampSym := te.Make("samp")
	te.bind(sz.NameSymLiteralOutput, 2, func(args []sxpf.Object) sxpf.Object {
		return te.transformLiteral(args, nil, sampSym)
	})
	te.bind(sz.NameSymLiteralProg, 2, func(args []sxpf.Object) sxpf.Object {
		return te.transformLiteral(args, nil, codeSym)
	})

	te.bind(sz.NameSymLiteralZettel, 0, func([]sxpf.Object) sxpf.Object { return sxpf.Nil() })
}

func (te *TransformEnv) makeFormatFn(tag string) transformFn {
	sym := te.Make(tag)
	return func(args []sxpf.Object) sxpf.Object {
		a := te.getAttributes(args[0])
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		res := sxpf.MakeList(args[1:]...)
		if len(a) > 0 {
			res = res.Cons(te.transformAttribute(a))
		}
		return res.Cons(sym)
	}
}
func (te *TransformEnv) transformQuote(args []sxpf.Object) sxpf.Object {
	const langAttr = "lang"
	a := te.getAttributes(args[0])
	langVal, found := a.Get(langAttr)
	if found {
		a = a.Remove(langAttr)
	}
	if val, found2 := a.Get(""); found2 {
		a = a.Remove("").AddClass(val)
	}
	res := sxpf.MakeList(args[1:]...)
	if len(a) > 0 {
		res = res.Cons(te.transformAttribute(a))
	}
	res = res.Cons(te.Make("q"))
	if found {
		res = sxpf.Nil().Cons(res).Cons(te.transformAttribute(attrs.Attributes{}.Set(langAttr, langVal))).Cons(te.symSpan)
	}
	return res
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func (te *TransformEnv) transformLiteral(args []sxpf.Object, a attrs.Attributes, sym *sxpf.Symbol) sxpf.Object {
	if a == nil {
		a = te.getAttributes(args[0])
	}
	a = setProgLang(a)
	literal := te.getString(args[1]).String()
	if a.HasDefault() {
		a = a.RemoveDefault()
		literal = visibleReplacer.Replace(literal)
	}
	res := sxpf.Nil().Cons(sxpf.MakeString(literal))
	if len(a) > 0 {
		res = res.Cons(te.transformAttribute(a))
	}
	return res.Cons(sym)
}

func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}

func (te *TransformEnv) transformHTML(args []sxpf.Object) sxpf.Object {
	if s := te.getString(args[1]); s != "" && IsSafe(s.String()) {
		return sxpf.Nil().Cons(s).Cons(te.symNoEscape)
	}
	return nil
}

func (te *TransformEnv) transformBLOB(description *sxpf.Cell, syntax, data sxpf.String) sxpf.Object {
	if data == "" {
		return sxpf.Nil()
	}
	switch syntax {
	case "":
		return sxpf.Nil()
	case api.ValueSyntaxSVG:
		return sxpf.Nil().Cons(sxpf.Nil().Cons(data).Cons(te.symNoEscape)).Cons(te.symP)
	default:
		imgAttr := sxpf.Nil().Cons(sxpf.Cons(te.Make("src"), sxpf.MakeString("data:image/"+syntax.String()+";base64,"+data.String())))
		var sb strings.Builder
		te.flattenText(&sb, description)
		if d := sb.String(); d != "" {
			imgAttr = imgAttr.Cons(sxpf.Cons(te.Make("alt"), sxpf.MakeString(d)))
		}
		return sxpf.Nil().Cons(sxpf.Nil().Cons(imgAttr.Cons(te.symAttr)).Cons(te.Make("img"))).Cons(te.symP)
	}
}

func (te *TransformEnv) flattenText(sb *strings.Builder, lst *sxpf.Cell) {
	for elem := lst; elem != nil; elem = elem.Tail() {
		switch obj := elem.Car().(type) {
		case sxpf.String:
			sb.WriteString(obj.String())
		case *sxpf.Cell:
			te.flattenText(sb, obj)
		}
	}
}

type transformFn func([]sxpf.Object) sxpf.Object

func (te *TransformEnv) bind(name string, minArity int, fn transformFn) {
	te.astEnv.Bind(te.astSF.MustMake(name), eval.BuiltinA(func(args []sxpf.Object) (sxpf.Object, error) {
		if nArgs := len(args); nArgs < minArity {
			return sxpf.Nil(), fmt.Errorf("not enough arguments (%d) for form %v (%d)", nArgs, name, minArity)
		}
		res := fn(args)
		return res, te.err
	}))
}

func (te *TransformEnv) Rebind(name string, fn func([]sxpf.Object, eval.Callable) sxpf.Object) {
	sym := te.astSF.MustMake(name)
	obj, found := te.astEnv.Lookup(sym)
	if !found {
		panic(sym.String())
	}
	preFn, ok := eval.GetCallable(obj)
	if !ok {
		panic(sym.String())
	}
	te.astEnv.Bind(sym, eval.BuiltinA(func(args []sxpf.Object) (sxpf.Object, error) {
		res := fn(args, preFn)
		return res, te.err
	}))
}

func (te *TransformEnv) Make(name string) *sxpf.Symbol { return te.tr.Make(name) }
func (te *TransformEnv) getSymbol(val sxpf.Object) *sxpf.Symbol {
	if te.err != nil {
		return nil
	}
	if sym, ok := sxpf.GetSymbol(val); ok {
		return sym
	}
	te.err = fmt.Errorf("%v/%T is not a symbol", val, val)
	return nil
}
func (te *TransformEnv) getString(val sxpf.Object) sxpf.String {
	if te.err != nil {
		return ""
	}
	if s, ok := sxpf.GetString(val); ok {
		return s
	}
	te.err = fmt.Errorf("%v/%T is not a string", val, val)
	return ""
}
func (te *TransformEnv) getInt64(val sxpf.Object) int64 {
	if te.err != nil {
		return -1017
	}
	if num, ok := sxpf.GetNumber(val); ok {
		return int64(num.(sxpf.Int64))
	}
	te.err = fmt.Errorf("%v/%T is not a number", val, val)
	return -1017
}
func (te *TransformEnv) getList(val sxpf.Object) *sxpf.Cell {
	if te.err == nil {
		if res, isCell := sxpf.GetCell(val); isCell {
			return res
		}
		te.err = fmt.Errorf("%v/%T is not a list", val, val)
	}
	return nil
}
func (te *TransformEnv) getAttributes(args sxpf.Object) attrs.Attributes {
	return sz.GetAttributes(te.getList(args))
}

func (te *TransformEnv) transformLink(a attrs.Attributes, refValue sxpf.String, inline []sxpf.Object) sxpf.Object {
	result := sxpf.MakeList(inline...)
	if len(inline) == 0 {
		result = sxpf.Nil().Cons(refValue)
	}
	if te.tr.noLinks {
		return result.Cons(te.symSpan)
	}
	return result.Cons(te.transformAttribute(a)).Cons(te.symA)
}

func (te *TransformEnv) transformAttribute(a attrs.Attributes) *sxpf.Cell {
	return te.tr.TransformAttrbute(a)
}

func (te *TransformEnv) transformMeta(a attrs.Attributes) *sxpf.Cell {
	return te.tr.TransformMeta(a)
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
