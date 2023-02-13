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
	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/sexpr"
)

// Transformer will transform a s-expression that encodes the zettel AST into an s-expression
// that represents HTML.
type Transformer struct {
	sf sxpf.SymbolFactory
}

// NewTransformer creates a new transformer object.
func NewTransformer() *Transformer {
	return &Transformer{
		sf: sxpf.MakeMappedFactory(),
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
	env := sxpf.MakeRootEnvironment()
	sexpr.BindOther(env, astSF)
	val, err := sexpr.Evaluate(env, lst)
	res, ok := val.(*sxpf.List)
	if !ok {
		panic("Result is not a list")
	}
	return res, err
}
