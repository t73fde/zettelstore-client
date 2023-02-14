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
	"strconv"
	"strings"

	"codeberg.org/t73fde/sxpf"
	"codeberg.org/t73fde/sxpf/eval"
	"zettelstore.de/c/attrs"
	"zettelstore.de/c/sexpr"
	"zettelstore.de/c/text"
)

// Transformer will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Transformer struct {
	sf            sxpf.SymbolFactory
	headingOffset int64
}

// NewTransformer creates a new transformer object.
func NewTransformer(headingOffset int) *Transformer {
	return &Transformer{
		sf:            sxpf.MakeMappedFactory(),
		headingOffset: int64(headingOffset),
	}
}

// Transform an AST s-expression into a HTML s-expression.
func (tr *Transformer) Transform(lst *sxpf.List) (*sxpf.List, error) {
	astSF := sxpf.FindSymbolFactory(lst)
	if astSF == nil {
		return nil, nil
	}
	if astSF == tr.sf {
		panic("Invalid AST SymbolFactory")
	}
	te := transformEnv{
		tr:      tr,
		astSF:   astSF,
		eenv:    sxpf.MakeRootEnvironment(),
		err:     nil,
		textEnc: text.NewEncoder(astSF),
	}
	te.initialize()

	sexpr.BindOther(te.eenv, astSF)
	val, err := sexpr.Evaluate(te.eenv, lst)
	res, ok := val.(*sxpf.List)
	if !ok {
		panic("Result is not a list")
	}
	return res, err
}

type transformEnv struct {
	tr      *Transformer
	astSF   sxpf.SymbolFactory
	eenv    sxpf.Environment
	err     error
	textEnc *text.Encoder
	symAt   *sxpf.Symbol
	symMeta *sxpf.Symbol
}

func (te *transformEnv) initialize() {
	htmlSF := te.tr.sf
	te.symAt = htmlSF.Make("@")
	te.symMeta = htmlSF.Make("meta")

	te.bind(sexpr.NameSymText, 1, func(args *sxpf.List) sxpf.Value { return te.getString(args) })
	te.bind(sexpr.NameSymSpace, 0, func(args *sxpf.List) sxpf.Value {
		if args.IsNil() {
			return sxpf.MakeString(" ")
		}
		return te.getString(args)
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

	// Block nodes
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
		a := sexpr.GetAttributes(te.getList(argAttr))
		argFragment := argAttr.Tail().Tail()
		if fragment := te.getString(argFragment).String(); fragment != "" {
			a = a.Set("id", fragment)
		}

		result := te.evaluateList(argFragment.Tail())
		if len(a) > 0 {
			result = result.Cons(te.transformAttrbute(a))
		}
		return result.Cons(te.make("h" + level))
	})
}

func (te *transformEnv) bind(name string, minArity int, fn func(*sxpf.List) sxpf.Value) {
	te.eenv.Bind(te.astSF.Make(name), eval.MakeSpecial(name, func(env sxpf.Environment, args *sxpf.List) (sxpf.Value, error) {
		if arity := args.Length(); arity < minArity {
			return sxpf.Nil(), fmt.Errorf("not enough arguments (%d) for form %v (%d)", arity, name, minArity)
		}
		return fn(args), te.err
	}))
}

// func (te *transformEnv) evaluate(val sxpf.Value) sxpf.Value {
// 	if te.err == nil {
// 		res, err := sexpr.Evaluate(te.eenv, val)
// 		if err == nil {
// 			return res
// 		}
// 		te.err = err
// 	}
// 	return sxpf.Nil()
// }

func (te *transformEnv) evaluateList(lst *sxpf.List) *sxpf.List {
	if te.err == nil {
		res, err := sexpr.EvaluateList(te.eenv, lst)
		if err == nil {
			return res
		}
		te.err = err
	}
	return sxpf.Nil()
}

func (te *transformEnv) make(name string) *sxpf.Symbol { return te.tr.sf.Make(name) }
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

func (te *transformEnv) transformAttrbute(a attrs.Attributes) *sxpf.List {
	if len(a) == 0 {
		return sxpf.Nil()
	}
	plist := sxpf.Nil()
	keys := a.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		key := keys[i]
		plist = plist.Cons(sxpf.MakePair(te.make(key), sxpf.MakeString(a[key])))
	}
	return plist.Cons(te.symAt)
}

func (te *transformEnv) transformMeta(a attrs.Attributes) *sxpf.List {
	return sxpf.Nil().Cons(te.transformAttrbute(a)).Cons(te.symMeta)
}
