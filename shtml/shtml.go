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

	"codeberg.org/t73fde/sxpf"
	"codeberg.org/t73fde/sxpf/eval"
	"zettelstore.de/c/api"
	"zettelstore.de/c/attrs"
	"zettelstore.de/c/sexpr"
	"zettelstore.de/c/text"
)

// Transformer will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Transformer struct {
	sf            sxpf.SymbolFactory
	headingOffset int64
	unique        string
	noLinks       bool // true iff output must not include links
	symAt         *sxpf.Symbol
	symMeta       *sxpf.Symbol
}

// NewTransformer creates a new transformer object.
func NewTransformer(headingOffset int) *Transformer {
	sf := sxpf.MakeMappedFactory()
	return &Transformer{
		sf:            sf,
		headingOffset: int64(headingOffset),
		symAt:         sf.Make("@"),
		symMeta:       sf.Make("meta"),
	}
}

// Make a new HTML symbol.
func (tr *Transformer) Make(s string) *sxpf.Symbol { return tr.sf.Make(s) }

// TransformAttrbute transforms the given attributes into a HTML s-expression.
func (tr *Transformer) TransformAttrbute(a attrs.Attributes) *sxpf.List {
	if len(a) == 0 {
		return sxpf.Nil()
	}
	plist := sxpf.Nil()
	keys := a.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		plist = plist.Cons(sxpf.MakePair(tr.Make(key), sxpf.MakeString(a[key])))
	}
	return plist.Cons(tr.symAt)
}

// TransformMeta creates a HTML meta s-expression
func (tr *Transformer) TransformMeta(a attrs.Attributes) *sxpf.List {
	return sxpf.Nil().Cons(tr.TransformAttrbute(a)).Cons(tr.symMeta)
}

// Transform an AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) Transform(lst *sxpf.List) (*sxpf.List, error) {
	return tr.TransformInline(lst, false, false)
}

// TransformInline an inlines AST s-expression into a list of HTML s-expressions.
func (tr *Transformer) TransformInline(lst *sxpf.List, noFootnotes, noLinks bool) (*sxpf.List, error) {
	astSF := sxpf.FindSymbolFactory(lst)
	if astSF == nil {
		return nil, nil
	}
	if astSF == tr.sf {
		panic("Invalid AST SymbolFactory")
	}
	te := transformEnv{
		tr:          tr,
		astSF:       astSF,
		eenv:        sxpf.MakeRootEnvironment(),
		err:         nil,
		textEnc:     text.NewEncoder(astSF),
		noFootnotes: noFootnotes,
		noLinks:     noLinks,
	}
	te.initialize()

	sexpr.BindOther(te.eenv, astSF)

	val, err := eval.Eval(te.eenv, lst)
	res, ok := val.(*sxpf.List)
	if !ok {
		panic("Result is not a list")
	}
	return res, err
}

type transformEnv struct {
	tr          *Transformer
	astSF       sxpf.SymbolFactory
	eenv        sxpf.Environment
	err         error
	textEnc     *text.Encoder
	noFootnotes bool
	noLinks     bool
	symAt       *sxpf.Symbol
	symMeta     *sxpf.Symbol
	symA        *sxpf.Symbol
	symSpan     *sxpf.Symbol
}

func (te *transformEnv) initialize() {
	te.symAt = te.make("@")
	te.symMeta = te.make("meta")
	te.symA = te.make("a")
	te.symSpan = te.make("span")

	te.bindMetadata()
	te.bindBlocks()
	te.bindInlines()
}

func (te *transformEnv) bindMetadata() {
	te.bind(sexpr.NameSymMeta, 0, func(args *sxpf.List) sxpf.Value {
		return te.evaluateList(args)
	})
	te.bind(sexpr.NameSymTypeZettelmarkup, 2, func(args *sxpf.List) sxpf.Value {
		a := make(attrs.Attributes, 2).
			Set("name", te.getString(args).String()).
			Set("content", te.textEnc.Encode(te.getList(args.Tail())))
		return te.transformMeta(a)
	})
	metaString := func(args *sxpf.List) sxpf.Value {
		a := make(attrs.Attributes, 2).
			Set("name", te.getString(args).String()).
			Set("content", te.getString(args.Tail()).String())
		return te.transformMeta(a)
	}
	te.bind(sexpr.NameSymTypeCredential, 2, metaString)
	te.bind(sexpr.NameSymTypeEmpty, 2, metaString)
	te.bind(sexpr.NameSymTypeID, 2, metaString)
	te.bind(sexpr.NameSymTypeNumber, 2, metaString)
	te.bind(sexpr.NameSymTypeString, 2, metaString)
	te.bind(sexpr.NameSymTypeTimestamp, 2, metaString)
	te.bind(sexpr.NameSymTypeURL, 2, metaString)
	te.bind(sexpr.NameSymTypeWord, 2, metaString)
	metaSet := func(args *sxpf.List) sxpf.Value {
		var sb strings.Builder
		for elem := te.getList(args.Tail()); elem != nil; elem = elem.Tail() {
			sb.WriteByte(' ')
			sb.WriteString(te.getString(elem).String())
		}
		s := sb.String()
		if len(s) > 0 {
			s = s[1:]
		}
		a := make(attrs.Attributes, 2).
			Set("name", te.getString(args).String()).
			Set("content", s)
		return te.transformMeta(a)
	}
	te.bind(sexpr.NameSymTypeIDSet, 2, metaSet)
	te.bind(sexpr.NameSymTypeTagSet, 2, metaSet)
	te.bind(sexpr.NameSymTypeWordSet, 2, metaSet)
}

func (te *transformEnv) bindBlocks() {
	te.bind(sexpr.NameSymBlock, 0, func(args *sxpf.List) sxpf.Value {
		return te.evaluateList(args)
	})
	te.bind(sexpr.NameSymPara, 0, func(args *sxpf.List) sxpf.Value {
		return te.evaluateList(args).Cons(te.make("p"))
	})
	te.bind(sexpr.NameSymHeading, 5, func(args *sxpf.List) sxpf.Value {
		nLevel := te.getInt64(args)
		if nLevel <= 0 {
			te.err = fmt.Errorf("%v is a negative level", nLevel)
			return sxpf.Nil()
		}
		level := strconv.FormatInt(nLevel+te.tr.headingOffset, 10)

		argAttr := args.Tail()
		a := te.getAttributes(argAttr)
		argFragment := argAttr.Tail().Tail()
		if fragment := te.getString(argFragment).String(); fragment != "" {
			a = a.Set("id", te.tr.unique+fragment)
		}

		result := te.evaluateList(argFragment.Tail())
		if len(a) > 0 {
			result = result.Cons(te.transformAttrbute(a))
		}
		return result.Cons(te.make("h" + level))
	})
	te.bind(sexpr.NameSymThematic, 0, func(args *sxpf.List) sxpf.Value {
		result := sxpf.Nil()
		if args != nil {
			if attrList := te.getList(args); attrList != nil {
				result = result.Cons(te.transformAttrbute(sexpr.GetAttributes(attrList)))
			}
		}
		return result.Cons(te.make("hr"))
	})
	te.bind(sexpr.NameSymListOrdered, 0, te.makeListFn("ol"))
	te.bind(sexpr.NameSymListUnordered, 0, te.makeListFn("ul"))
}

func (te *transformEnv) makeListFn(tag string) specialFn {
	sym := te.make(tag)
	return func(args *sxpf.List) sxpf.Value {
		result := sxpf.Nil().Cons(sym)
		last := result
		for elem := args; elem != nil; elem = elem.Tail() {
			item := sxpf.Nil().Cons(te.make("li"))
			if res, ok := te.evaluate(elem.Head()).(*sxpf.List); ok {
				item.ExtendBang(res)
			}
			last = last.AppendBang(item)
		}
		return result
	}
}

func (te *transformEnv) bindInlines() {
	te.bind(sexpr.NameSymInline, 0, func(args *sxpf.List) sxpf.Value {
		return te.evaluateList(args)
	})
	te.bind(sexpr.NameSymText, 1, func(args *sxpf.List) sxpf.Value { return te.getString(args) })
	te.bind(sexpr.NameSymSpace, 0, func(args *sxpf.List) sxpf.Value {
		if args.IsNil() {
			return sxpf.MakeString(" ")
		}
		return te.getString(args)
	})
	te.bind(sexpr.NameSymSoft, 0, func(*sxpf.List) sxpf.Value { return sxpf.MakeString(" ") })
	brSym := te.make("br")
	te.bind(sexpr.NameSymHard, 0, func(*sxpf.List) sxpf.Value { return sxpf.Nil().Cons(brSym) })
	transformAsSpan := func(args *sxpf.List) sxpf.Value {
		if args.Length() > 2 {
			return te.evaluate(args.Tail().Tail().Cons(args.Head()).Cons(te.make(sexpr.NameSymFormatSpan)))
		}
		return nil
	}
	te.bind(sexpr.NameSymLinkInvalid, 2, transformAsSpan)
	transformHREF := func(args *sxpf.List) sxpf.Value {
		a := te.getAttributes(args)
		refValue := te.getString(args.Tail())
		return te.transformLink(a.Set("href", refValue.String()), refValue, args.Tail().Tail())
	}
	te.bind(sexpr.NameSymLinkZettel, 2, transformHREF)
	te.bind(sexpr.NameSymLinkSelf, 2, transformHREF)
	te.bind(sexpr.NameSymLinkFound, 2, transformHREF)
	te.bind(sexpr.NameSymLinkBroken, 2, func(args *sxpf.List) sxpf.Value {
		a := te.getAttributes(args)
		refValue := te.getString(args.Tail())
		return te.transformLink(a.AddClass("broken"), refValue, args.Tail().Tail())
	})
	te.bind(sexpr.NameSymLinkHosted, 2, transformHREF)
	te.bind(sexpr.NameSymLinkBased, 2, transformHREF)
	te.bind(sexpr.NameSymLinkQuery, 2, func(args *sxpf.List) sxpf.Value {
		a := te.getAttributes(args)
		refValue := te.getString(args.Tail())
		query := "?" + api.QueryKeyQuery + "=" + url.QueryEscape(refValue.String())
		return te.transformLink(a.Set("href", query), refValue, args.Tail().Tail())
	})
	te.bind(sexpr.NameSymLinkExternal, 2, func(args *sxpf.List) sxpf.Value {
		a := te.getAttributes(args)
		refValue := te.getString(args.Tail())
		return te.transformLink(a.Set("href", refValue.String()).AddClass("external"), refValue, args.Tail().Tail())
	})

	te.bind(sexpr.NameSymCite, 2, func(args *sxpf.List) sxpf.Value {
		result := sxpf.Nil()
		argKey := args.Tail()
		if key := te.getString(argKey); key != "" {
			if text := argKey.Tail(); text != nil {
				result = te.evaluateList(text).Cons(sxpf.MakeString(", "))
			}
			result = result.Cons(key)
		}
		if a := te.getAttributes(args); len(a) > 0 {
			result = result.Cons(te.transformAttrbute(a))
		}
		if result == nil {
			return nil
		}
		return result.Cons(te.symSpan)
	})

	te.bind(sexpr.NameSymMark, 3, func(args *sxpf.List) sxpf.Value {
		argFragment := args.Tail().Tail()
		result := te.evaluateList(argFragment.Tail())
		if !te.tr.noLinks {
			if fragment := te.getString(argFragment); fragment != "" {
				a := attrs.Attributes{"id": fragment.String() + te.tr.unique}
				return result.Cons(te.transformAttrbute(a)).Cons(te.make("a"))
			}
		}
		return result.Cons(te.symSpan)
	})

	te.bind(sexpr.NameSymFormatDelete, 1, te.makeFormatFn("del"))
	te.bind(sexpr.NameSymFormatEmph, 1, te.makeFormatFn("em"))
	te.bind(sexpr.NameSymFormatInsert, 1, te.makeFormatFn("ins"))
	te.bind(sexpr.NameSymFormatQuote, 1, te.transformQuote)
	te.bind(sexpr.NameSymFormatSpan, 1, te.makeFormatFn("span"))
	te.bind(sexpr.NameSymFormatStrong, 1, te.makeFormatFn("strong"))
	te.bind(sexpr.NameSymFormatSub, 1, te.makeFormatFn("sub"))
	te.bind(sexpr.NameSymFormatSuper, 1, te.makeFormatFn("sup"))

	te.bind(sexpr.NameSymLiteralComment, 1, func(args *sxpf.List) sxpf.Value {
		if te.getAttributes(args).HasDefault() {
			if s := te.getString(args.Tail()); s != "" {
				return sxpf.Nil().Cons(s).Cons(te.make("@@"))
			}
		}
		return nil
	})
	te.bind(sexpr.NameSymLiteralHTML, 2, te.transformHTML)
	kbdSym := te.make("kbd")
	te.bind(sexpr.NameSymLiteralInput, 2, func(args *sxpf.List) sxpf.Value {
		return te.transformLiteral(args, nil, kbdSym)
	})
	codeSym := te.make("code")
	te.bind(sexpr.NameSymLiteralMath, 2, func(args *sxpf.List) sxpf.Value {
		a := te.getAttributes(args).AddClass("zs-math")
		return te.transformLiteral(args, a, codeSym)
	})
	sampSym := te.make("samp")
	te.bind(sexpr.NameSymLiteralOutput, 2, func(args *sxpf.List) sxpf.Value {
		return te.transformLiteral(args, nil, sampSym)
	})
	te.bind(sexpr.NameSymLiteralProg, 2, func(args *sxpf.List) sxpf.Value {
		a := setProgLang(te.getAttributes(args))
		return te.transformLiteral(args, a, codeSym)
	})
}

func (te *transformEnv) makeFormatFn(tag string) specialFn {
	sym := te.make(tag)
	return func(args *sxpf.List) sxpf.Value {
		a := te.getAttributes(args)
		if val, found := a.Get(""); found {
			a = a.Remove("").AddClass(val)
		}
		res := te.evaluateList(args.Tail())
		if len(a) > 0 {
			res = res.Cons(te.transformAttrbute(a))
		}
		return res.Cons(sym)
	}
}
func (te *transformEnv) transformQuote(args *sxpf.List) sxpf.Value {
	const langAttr = "lang"
	a := te.getAttributes(args)
	langVal, found := a.Get(langAttr)
	if found {
		a = a.Remove(langAttr)
	}
	if val, found := a.Get(""); found {
		a = a.Remove("").AddClass(val)
	}
	res := te.evaluateList(args.Tail())
	if len(a) > 0 {
		res = res.Cons(te.transformAttrbute(a))
	}
	res = res.Cons(te.make("q"))
	if found {
		res = sxpf.Nil().Cons(res).Cons(te.transformAttrbute(attrs.Attributes{}.Set(langAttr, langVal))).Cons(te.make("span"))
	}
	return res
}

var visibleReplacer = strings.NewReplacer(" ", "\u2423")

func (te *transformEnv) transformLiteral(args *sxpf.List, a attrs.Attributes, sym *sxpf.Symbol) sxpf.Value {
	if a == nil {
		a = te.getAttributes(args)
	}
	literal := te.getString(args.Tail()).String()
	if a.HasDefault() {
		a = a.RemoveDefault()
		literal = visibleReplacer.Replace(literal)
	}
	res := sxpf.Nil().Cons(sxpf.MakeString(literal))
	if len(a) > 0 {
		res = res.Cons(te.transformAttrbute(a))
	}
	return res.Cons(sym)
}
func setProgLang(a attrs.Attributes) attrs.Attributes {
	if val, found := a.Get(""); found {
		a = a.AddClass("language-" + val).Remove("")
	}
	return a
}
func (te *transformEnv) transformHTML(args *sxpf.List) sxpf.Value {
	if s := te.getString(args.Tail()); s != "" && IsSafe(s.String()) {
		return s
	}
	return nil
}

type specialFn func(*sxpf.List) sxpf.Value

func (te *transformEnv) bind(name string, minArity int, fn specialFn) {
	te.eenv.Bind(te.astSF.Make(name), eval.MakeSpecial(name, func(env sxpf.Environment, args *sxpf.List) (sxpf.Value, error) {
		if arity := args.Length(); arity < minArity {
			return sxpf.Nil(), fmt.Errorf("not enough arguments (%d) for form %v (%d)", arity, name, minArity)
		}
		return fn(args), te.err
	}))
}

func (te *transformEnv) evaluate(val sxpf.Value) sxpf.Value {
	if te.err == nil {
		res, err := eval.Eval(te.eenv, val)
		if err == nil {
			return res
		}
		te.err = err
	}
	return sxpf.Nil()
}

func (te *transformEnv) evaluateList(lst *sxpf.List) *sxpf.List {
	if te.err == nil {
		res, _, err := eval.EvalList(te.eenv, lst)
		if err == nil {
			return res
		}
		te.err = err
	}
	return sxpf.Nil()
}

func (te *transformEnv) make(name string) *sxpf.Symbol { return te.tr.Make(name) }
func (te *transformEnv) getString(lst *sxpf.List) sxpf.String {
	if te.err != nil {
		return ""
	}
	val := lst.Head()
	if s, ok := val.(sxpf.String); ok {
		return s
	}
	te.err = fmt.Errorf("%v/%T is not a string", val, val)
	return ""
}
func (te *transformEnv) getInt64(lst *sxpf.List) int64 {
	if te.err != nil {
		return -1017
	}
	val := lst.Head()
	if num, ok := val.(*sxpf.Number); ok {
		return num.GetInt64()
	}
	te.err = fmt.Errorf("%v/%T is not a number", val, val)
	return -1017
}
func (te *transformEnv) getList(lst *sxpf.List) *sxpf.List {
	if te.err == nil {
		val := lst.Head()
		if res, ok := val.(*sxpf.List); ok {
			return res
		}
		te.err = fmt.Errorf("%v/%T is not a list", val, val)
	}
	return sxpf.Nil()
}
func (te *transformEnv) getAttributes(args *sxpf.List) attrs.Attributes {
	return sexpr.GetAttributes(te.getList(args))
}

func (te *transformEnv) transformLink(a attrs.Attributes, refValue sxpf.String, inline *sxpf.List) sxpf.Value {
	var result *sxpf.List
	if inline.IsNil() {
		result = sxpf.Nil().Cons(refValue)
	} else {
		result = te.evaluateList(inline)
	}
	if te.tr.noLinks {
		return result.Cons(te.symSpan)
	}
	return result.Cons(te.transformAttrbute(a)).Cons(te.make("a"))
}

func (te *transformEnv) transformAttrbute(a attrs.Attributes) *sxpf.List {
	return te.tr.TransformAttrbute(a)
}

func (te *transformEnv) transformMeta(a attrs.Attributes) *sxpf.List {
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
