//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package zjson

type Meta map[string]MetaValue
type MetaValue struct {
	Type  string
	Key   string
	Value Value
}

func MakeMeta(val Value) Meta {
	obj := MakeObject(val)
	if len(obj) == 0 {
		return nil
	}
	result := make(Meta, len(obj))
	for k, v := range obj {
		mvObj := MakeObject(v)
		if len(mvObj) == 0 {
			continue
		}
		mv := makeMetaValue(mvObj)
		if mv.Type != "" {
			result[k] = mv
		}
	}
	return result
}
func makeMetaValue(mvObj Object) MetaValue {
	mv := MetaValue{}
	for n, val := range mvObj {
		if n == NameType {
			if t, ok := val.(string); ok {
				mv.Type = t
			}
		} else {
			mv.Key = n
			mv.Value = val
		}
	}
	return mv
}

func (m Meta) GetArray(key string) Array {
	if v, found := m[key]; found {
		return MakeArray(v.Value)
	}
	return nil
}

func (m Meta) GetString(key string) string {
	if v, found := m[key]; found {
		return MakeString(v.Value)
	}
	return ""
}
