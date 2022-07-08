//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package sexpr

import "codeberg.org/t73fde/sxpf"

func MakeString(val sxpf.Value) string {
	if strVal, ok := val.(*sxpf.String); ok {
		return strVal.GetValue()
	}
	return ""
}

// GetMetaContent returns the metadata and the content of a sexpr encoded zettel.
func GetMetaContent(zettel sxpf.Value) (Meta, sxpf.Value) {
	if pair, ok := zettel.(*sxpf.Pair); ok {
		m := pair.GetFirst()
		if s := pair.GetSecond(); s != nil {
			if p, ok := s.(*sxpf.Pair); ok {
				return MakeMeta(m), p.GetFirst()
			}
		}
		return MakeMeta(m), nil
	}
	return nil, nil
}

type Meta map[string]MetaValue
type MetaValue struct {
	Type  string
	Key   string
	Value sxpf.Value
}

func MakeMeta(val sxpf.Value) Meta {
	if result := doMakeMeta(val); len(result) > 0 {
		return result
	}
	return nil
}
func doMakeMeta(val sxpf.Value) Meta {
	result := make(map[string]MetaValue)
	for {
		if val == nil {
			return result
		}
		pair, ok := val.(*sxpf.Pair)
		if !ok {
			return result
		}
		if mv, ok := makeMetaValue(pair); ok {
			result[mv.Key] = mv
		}
		val = pair.GetSecond()
	}
}
func makeMetaValue(pair *sxpf.Pair) (MetaValue, bool) {
	var result MetaValue
	typePair, ok := pair.GetFirst().(*sxpf.Pair)
	if !ok {
		return result, false
	}
	typeVal, ok := typePair.GetFirst().(*sxpf.Symbol)
	if !ok {
		return result, false
	}
	keyPair, ok := typePair.GetSecond().(*sxpf.Pair)
	if !ok {
		return result, false
	}
	keyStr, ok := keyPair.GetFirst().(*sxpf.String)
	if !ok {
		return result, false
	}
	valPair, ok := keyPair.GetSecond().(*sxpf.Pair)
	if !ok {
		return result, false
	}
	result.Type = typeVal.GetValue()
	result.Key = keyStr.GetValue()
	result.Value = valPair.GetFirst()
	return result, true
}

func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return MakeString(v.Value)
	}
	return ""
}

func (m Meta) GetPair(key string) *sxpf.Pair {
	if mv, found := m[key]; found {
		if seq, ok := mv.Value.(*sxpf.Pair); ok && !seq.IsEmpty() {
			return seq
		}
	}
	return nil
}
