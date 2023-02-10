//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr

import (
	"codeberg.org/t73fde/sxpf"
	"codeberg.org/t73fde/sxpf/eval"
)

// BindOther bind all unbound symbols to a special function that just recursively
// traverses through the argument lists.
func BindOther(env sxpf.Environment, sf sxpf.SymbolFactory) {
	for _, sym := range sf.Symbols() {
		if _, found := env.Resolve(sym); !found {
			env.Bind(sym, eval.MakeSpecial(sym.String(), DoNothing))
		}
	}

}

// DoNothing just traverses though all (sub-) lists.
func DoNothing(env sxpf.Environment, args *sxpf.List) (sxpf.Value, error) {
	for elem := args; elem != nil; elem = elem.Tail() {
		if lst, ok := elem.Head().(*sxpf.List); ok {
			if _, err := eval.Eval(env, lst); err != nil {
				return sxpf.Nil(), err
			}
		}
	}
	return sxpf.Nil(), nil
}

// Evaluate a given value in a given environment.
func Evaluate(env sxpf.Environment, val sxpf.Value) (sxpf.Value, error) {
	for {
		switch v := val.(type) {
		case *sxpf.Symbol:
			res, found := env.Resolve(v)
			if !found {
				return sxpf.Nil(), eval.NotBoundError{Env: env, Sym: v}
			}
			return res, nil
		case *sxpf.List:
			if v.IsNil() {
				return sxpf.Nil(), nil // Nil() evaluates to itself
			}
			res, err := Evaluate(env, v.Head())
			if err != nil {
				return sxpf.Nil(), err
			}
			if fn, ok := res.(sxpf.Callable); ok {
				res, err = fn.Call(env, v.Tail())
			} else if lst, ok := res.(*sxpf.List); ok {
				res, err = EvaluateList(env, lst)
			}
			if err == nil || err != sxpf.ErrEvalAgain {
				return res, err
			}
			val = res
		default:
			return val, nil // All other values evaluate to themself
		}
	}
}

// EvaluateList will return a list of evaluated elements
func EvaluateList(env sxpf.Environment, lst *sxpf.List) (*sxpf.List, error) {
	temp := make([]sxpf.Value, 0, lst.Length())
	for elem := lst; elem != nil; elem = elem.Tail() {
		val, err := Evaluate(env, elem.Head())
		if err != nil {
			return sxpf.MakeList(temp...), err
		}
		temp = append(temp, val)
	}
	return sxpf.MakeList(temp...), nil
}

// func MakeString(val sxpf.Value) string {
// 	if strVal, ok := val.(*sxpf.String); ok {
// 		return strVal.GetValue()
// 	}
// 	return ""
// }

// // GetMetaContent returns the metadata and the content of a sexpr encoded zettel.
// func GetMetaContent(zettel sxpf.Value) (Meta, *sxpf.Pair) {
// 	if pair, ok := zettel.(*sxpf.Pair); ok {
// 		m := pair.GetFirst()
// 		if s := pair.GetSecond(); s != nil {
// 			if p, ok2 := s.(*sxpf.Pair); ok2 {
// 				if content, err := p.GetPair(); err == nil {
// 					return MakeMeta(m), content
// 				}
// 			}
// 		}
// 		return MakeMeta(m), nil
// 	}
// 	return nil, nil
// }

// type Meta map[string]MetaValue
// type MetaValue struct {
// 	Type  string
// 	Key   string
// 	Value sxpf.Value
// }

// func MakeMeta(val sxpf.Value) Meta {
// 	if result := doMakeMeta(val); len(result) > 0 {
// 		return result
// 	}
// 	return nil
// }
// func doMakeMeta(val sxpf.Value) Meta {
// 	result := make(map[string]MetaValue)
// 	for {
// 		if val == nil {
// 			return result
// 		}
// 		pair, ok := val.(*sxpf.Pair)
// 		if !ok {
// 			return result
// 		}
// 		if mv, ok2 := makeMetaValue(pair); ok2 {
// 			result[mv.Key] = mv
// 		}
// 		val = pair.GetSecond()
// 	}
// }
// func makeMetaValue(pair *sxpf.Pair) (MetaValue, bool) {
// 	var result MetaValue
// 	typePair, ok := pair.GetFirst().(*sxpf.Pair)
// 	if !ok {
// 		return result, false
// 	}
// 	typeVal, ok := typePair.GetFirst().(*sxpf.Symbol)
// 	if !ok {
// 		return result, false
// 	}
// 	keyPair, ok := typePair.GetSecond().(*sxpf.Pair)
// 	if !ok {
// 		return result, false
// 	}
// 	keyStr, ok := keyPair.GetFirst().(*sxpf.String)
// 	if !ok {
// 		return result, false
// 	}
// 	valPair, ok := keyPair.GetSecond().(*sxpf.Pair)
// 	if !ok {
// 		return result, false
// 	}
// 	result.Type = typeVal.GetValue()
// 	result.Key = keyStr.GetValue()
// 	result.Value = valPair.GetFirst()
// 	return result, true
// }

// func (m Meta) GetString(key string) string {
// 	if v, found := m[key]; found {
// 		return MakeString(v.Value)
// 	}
// 	return ""
// }

// func (m Meta) GetPair(key string) *sxpf.Pair {
// 	if mv, found := m[key]; found {
// 		if seq, ok := mv.Value.(*sxpf.Pair); ok && !seq.IsEmpty() {
// 			return seq
// 		}
// 	}
// 	return nil
// }
