//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore Client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

package meta

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"t73f.de/r/zsc/domain/id"
)

// DescriptionType is a description of a specific key type.
type DescriptionType struct {
	Name  string
	IsSet bool
}

// String returns the string representation of the given type
func (t DescriptionType) String() string { return t.Name }

var registeredTypes = make(map[string]*DescriptionType)

func registerType(name string, isSet bool) *DescriptionType {
	if _, ok := registeredTypes[name]; ok {
		panic("Type '" + name + "' already registered")
	}
	t := &DescriptionType{name, isSet}
	registeredTypes[name] = t
	return t
}

// Values of the metadata key/value types.
//
// See [Supported Key Types].
//
// [Supported Key Types]: https://zettelstore.de/manual/h/00001006030000
const (
	MetaCredential = "Credential"
	MetaEmpty      = "EString"
	MetaID         = "Identifier"
	MetaIDSet      = "IdentifierSet"
	MetaNumber     = "Number"
	MetaString     = "String"
	MetaTagSet     = "TagSet"
	MetaTimestamp  = "Timestamp"
	MetaURL        = "URL"
	MetaWord       = "Word"
)

// Supported key types.
var (
	TypeCredential = registerType(MetaCredential, false)
	TypeEmpty      = registerType(MetaEmpty, false)
	TypeID         = registerType(MetaID, false)
	TypeIDSet      = registerType(MetaIDSet, true)
	TypeNumber     = registerType(MetaNumber, false)
	TypeString     = registerType(MetaString, false)
	TypeTagSet     = registerType(MetaTagSet, true)
	TypeTimestamp  = registerType(MetaTimestamp, false)
	TypeURL        = registerType(MetaURL, false)
	TypeWord       = registerType(MetaWord, false)
)

// Type returns a type hint for the given key. If no type hint is specified,
// TypeUnknown is returned.
func (*Meta) Type(key string) *DescriptionType {
	return Type(key)
}

// Some constants for key suffixes that determine a type.
const (
	SuffixKeyRole = "-role"
	SuffixKeyURL  = "-url"
)

var (
	cachedTypedKeys = make(map[string]*DescriptionType)
	mxTypedKey      sync.RWMutex
	suffixTypes     = map[string]*DescriptionType{
		"-date":       TypeTimestamp,
		"-number":     TypeNumber,
		SuffixKeyRole: TypeWord,
		"-time":       TypeTimestamp,
		SuffixKeyURL:  TypeURL,
		"-zettel":     TypeID,
		"-zid":        TypeID,
		"-zids":       TypeIDSet,
	}
)

// Type returns a type hint for the given key. If no type hint is specified,
// TypeEmpty is returned.
func Type(key string) *DescriptionType {
	if k, ok := registeredKeys[key]; ok {
		return k.Type
	}
	mxTypedKey.RLock()
	k, found := cachedTypedKeys[key]
	mxTypedKey.RUnlock()
	if found {
		return k
	}

	for suffix, t := range suffixTypes {
		if strings.HasSuffix(key, suffix) {
			mxTypedKey.Lock()
			defer mxTypedKey.Unlock()
			// Double check to avoid races
			if _, found = cachedTypedKeys[key]; !found {
				cachedTypedKeys[key] = t
			}
			return t
		}
	}
	return TypeEmpty
}

// SetList stores the given string list value under the given key.
func (m *Meta) SetList(key string, values []string) {
	if key != KeyID {
		for i, val := range values {
			values[i] = string(Value(val).TrimSpace())
		}
		m.pairs[key] = Value(strings.Join(values, " "))
	}
}

// SetWord stores the given word under the given key.
func (m *Meta) SetWord(key, word string) {
	if slist := Value(word).AsList(); len(slist) > 0 {
		m.Set(key, Value(slist[0]))
	}
}

// SetNow stores the current timestamp under the given key.
func (m *Meta) SetNow(key string) {
	m.Set(key, Value(time.Now().Local().Format(id.TimestampLayout)))
}

// GetBool returns the boolean value of the given key.
func (m *Meta) GetBool(key string) bool {
	if val, ok := m.Get(key); ok {
		return val.AsBool()
	}
	return false
}

// GetList retrieves the string list value of a given key. The bool value
// signals, whether there was a value stored or not.
func (m *Meta) GetList(key string) ([]string, bool) {
	value, ok := m.Get(key)
	if !ok {
		return nil, false
	}
	return value.AsList(), true
}

// GetNumber retrieves the numeric value of a given key.
func (m *Meta) GetNumber(key string, def int64) int64 {
	if value, ok := m.Get(key); ok {
		if num, err := strconv.ParseInt(string(value), 10, 64); err == nil {
			return num
		}
	}
	return def
}
