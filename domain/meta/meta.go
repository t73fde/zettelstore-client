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

// Package meta provides the zettel specific type 'meta'.
package meta

import (
	"iter"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/input"
	mymaps "t73f.de/r/zsc/maps"
	"t73f.de/r/zsc/strfun"
)

type keyUsage int

const (
	_             keyUsage = iota
	usageUser              // Key will be manipulated by the user
	usageComputed          // Key is computed by zettelstore
	usageProperty          // Key is computed and not stored by zettelstore
)

// DescriptionKey formally describes each supported metadata key.
type DescriptionKey struct {
	Name    string
	Type    *DescriptionType
	usage   keyUsage
	Inverse string
}

// IsComputed returns true, if metadata is computed and not set by the user.
func (kd *DescriptionKey) IsComputed() bool { return kd.usage >= usageComputed }

// IsProperty returns true, if metadata is a computed property.
func (kd *DescriptionKey) IsProperty() bool { return kd.usage >= usageProperty }

var registeredKeys = make(map[string]*DescriptionKey)

func registerKey(name string, t *DescriptionType, usage keyUsage, inverse string) {
	if _, ok := registeredKeys[name]; ok {
		panic("Key '" + name + "' already defined")
	}
	if inverse != "" {
		if t != TypeID && t != TypeIDSet {
			panic("Inversable key '" + name + "' is not identifier type, but " + t.String())
		}
		inv, ok := registeredKeys[inverse]
		if !ok {
			panic("Inverse Key '" + inverse + "' not found")
		}
		if !inv.IsComputed() {
			panic("Inverse Key '" + inverse + "' is not computed.")
		}
		if inv.Type != TypeIDSet {
			panic("Inverse Key '" + inverse + "' is not an identifier set, but " + inv.Type.String())
		}
	}
	registeredKeys[name] = &DescriptionKey{name, t, usage, inverse}
}

// IsComputed returns true, if key denotes a computed metadata key.
func IsComputed(name string) bool {
	if kd, ok := registeredKeys[name]; ok {
		return kd.IsComputed()
	}
	return false
}

// IsProperty returns true, if key denotes a property metadata value.
func IsProperty(name string) bool {
	if kd, ok := registeredKeys[name]; ok {
		return kd.IsProperty()
	}
	return false
}

// Inverse returns the name of the inverse key.
func Inverse(name string) string {
	if kd, ok := registeredKeys[name]; ok {
		return kd.Inverse
	}
	return ""
}

// GetDescription returns the key description object of the given key name.
func GetDescription(name string) DescriptionKey {
	if d, ok := registeredKeys[name]; ok {
		return *d
	}
	return DescriptionKey{Type: Type(name)}
}

// GetSortedKeyDescriptions delivers all metadata key descriptions as a slice, sorted by name.
func GetSortedKeyDescriptions() []*DescriptionKey {
	keys := mymaps.Keys(registeredKeys)
	result := make([]*DescriptionKey, 0, len(keys))
	for _, n := range keys {
		result = append(result, registeredKeys[n])
	}
	return result
}

// Key is the type of metadata keys.
type Key = string

// Predefined / supported metadata keys.
//
// See [Supported Metadata Keys].
//
// [Supported Metadata Keys]: https://zettelstore.de/manual/h/00001006020000
const (
	KeyID           = "id"
	KeyTitle        = "title"
	KeyRole         = "role"
	KeyTags         = "tags"
	KeySyntax       = "syntax"
	KeyAuthor       = "author"
	KeyBack         = "back"
	KeyBackward     = "backward"
	KeyBoxNumber    = "box-number"
	KeyCopyright    = "copyright"
	KeyCreated      = "created"
	KeyCredential   = "credential"
	KeyDead         = "dead"
	KeyExpire       = "expire"
	KeyFolge        = "folge"
	KeyFolgeRole    = "folge-role"
	KeyForward      = "forward"
	KeyLang         = "lang"
	KeyLicense      = "license"
	KeyModified     = "modified"
	KeyPrecursor    = "precursor"
	KeyPredecessor  = "predecessor"
	KeyPrequel      = "prequel"
	KeyPublished    = "published"
	KeyQuery        = "query"
	KeyReadOnly     = "read-only"
	KeySequel       = "sequel"
	KeySubordinates = "subordinates"
	KeySuccessors   = "successors"
	KeySummary      = "summary"
	KeySuperior     = "superior"
	KeyURL          = "url"
	KeyUselessFiles = "useless-files"
	KeyUserID       = "user-id"
	KeyUserRole     = "user-role"
	KeyVisibility   = "visibility"
)

// Supported keys.
func init() {
	registerKey(KeyID, TypeID, usageComputed, "")
	registerKey(KeyTitle, TypeEmpty, usageUser, "")
	registerKey(KeyRole, TypeWord, usageUser, "")
	registerKey(KeyTags, TypeTagSet, usageUser, "")
	registerKey(KeySyntax, TypeWord, usageUser, "")

	// Properties that are inverse keys
	registerKey(KeyFolge, TypeIDSet, usageProperty, "")
	registerKey(KeySequel, TypeIDSet, usageProperty, "")
	registerKey(KeySuccessors, TypeIDSet, usageProperty, "")
	registerKey(KeySubordinates, TypeIDSet, usageProperty, "")

	// Non-inverse keys
	registerKey(KeyAuthor, TypeString, usageUser, "")
	registerKey(KeyBack, TypeIDSet, usageProperty, "")
	registerKey(KeyBackward, TypeIDSet, usageProperty, "")
	registerKey(KeyBoxNumber, TypeNumber, usageProperty, "")
	registerKey(KeyCopyright, TypeString, usageUser, "")
	registerKey(KeyCreated, TypeTimestamp, usageComputed, "")
	registerKey(KeyCredential, TypeCredential, usageUser, "")
	registerKey(KeyDead, TypeIDSet, usageProperty, "")
	registerKey(KeyExpire, TypeTimestamp, usageUser, "")
	registerKey(KeyFolgeRole, TypeWord, usageUser, "")
	registerKey(KeyForward, TypeIDSet, usageProperty, "")
	registerKey(KeyLang, TypeWord, usageUser, "")
	registerKey(KeyLicense, TypeEmpty, usageUser, "")
	registerKey(KeyModified, TypeTimestamp, usageComputed, "")
	registerKey(KeyPrecursor, TypeIDSet, usageUser, KeyFolge)
	registerKey(KeyPredecessor, TypeID, usageUser, KeySuccessors)
	registerKey(KeyPrequel, TypeIDSet, usageUser, KeySequel)
	registerKey(KeyPublished, TypeTimestamp, usageProperty, "")
	registerKey(KeyQuery, TypeEmpty, usageUser, "")
	registerKey(KeyReadOnly, TypeWord, usageUser, "")
	registerKey(KeySummary, TypeString, usageUser, "")
	registerKey(KeySuperior, TypeIDSet, usageUser, KeySubordinates)
	registerKey(KeyURL, TypeURL, usageUser, "")
	registerKey(KeyUselessFiles, TypeString, usageProperty, "")
	registerKey(KeyUserID, TypeWord, usageUser, "")
	registerKey(KeyUserRole, TypeWord, usageUser, "")
	registerKey(KeyVisibility, TypeWord, usageUser, "")
}

// NewPrefix is the prefix for metadata keys in template zettel for creating new zettel.
const NewPrefix = "new-"

// Meta contains all meta-data of a zettel.
type Meta struct {
	Zid     id.Zid
	pairs   map[Key]Value
	YamlSep bool
}

// New creates a new chunk for storing metadata.
func New(zid id.Zid) *Meta {
	return &Meta{Zid: zid, pairs: make(map[Key]Value, 5)}
}

// NewWithData creates metadata object with given data.
func NewWithData(zid id.Zid, data map[string]string) *Meta {
	pairs := make(map[Key]Value, len(data))
	for k, v := range data {
		pairs[k] = Value(v)
	}
	return &Meta{Zid: zid, pairs: pairs}
}

// Length returns the number of bytes stored for the metadata.
func (m *Meta) Length() int {
	if m == nil {
		return 0
	}
	result := 6 // storage needed for Zid
	for k, v := range m.pairs {
		result += len(k) + len(v) + 1 // 1 because separator
	}
	return result
}

// Clone returns a new copy of the metadata.
func (m *Meta) Clone() *Meta {
	return &Meta{
		Zid:     m.Zid,
		pairs:   maps.Clone(m.pairs),
		YamlSep: m.YamlSep,
	}
}

// Map returns a copy of the meta data as a string map.
func (m *Meta) Map() map[string]string {
	pairs := make(map[string]string, len(m.pairs))
	for k, v := range m.pairs {
		pairs[k] = string(v)
	}
	return pairs
}

var reKey = regexp.MustCompile("^[0-9a-z][-0-9a-z]{0,254}$")

// KeyIsValid returns true, if the string is a valid metadata key.
func KeyIsValid(s string) bool { return reKey.MatchString(s) }

var firstKeys = []string{KeyTitle, KeyRole, KeyTags, KeySyntax}

// Set stores the given string value under the given key.
func (m *Meta) Set(key string, value Value) {
	if key != KeyID {
		m.pairs[key] = value.TrimSpace()
	}
}

// SetNonEmpty stores the given value under the given key, if the value is non-empty.
// An empty value will delete the previous association.
func (m *Meta) SetNonEmpty(key string, value Value) {
	if value == "" {
		delete(m.pairs, key) // TODO: key != KeyID
	} else {
		m.Set(key, value.TrimSpace())
	}
}

// TrimSpace removes all leading and remaining space from value
func (val Value) TrimSpace() Value {
	return Value(strings.TrimFunc(string(val), input.IsSpace))
}

// Get retrieves the string value of a given key. The bool value signals,
// whether there was a value stored or not.
func (m *Meta) Get(key string) (Value, bool) {
	if m == nil {
		return "", false
	}
	if key == KeyID {
		return Value(m.Zid.String()), true
	}
	value, ok := m.pairs[key]
	return value, ok
}

// GetDefault retrieves the string value of the given key. If no value was
// stored, the given default value is returned.
func (m *Meta) GetDefault(key string, def Value) Value {
	if value, found := m.Get(key); found {
		return value
	}
	return def
}

// GetTitle returns the title of the metadata. It is the only key that has a
// defined default value: the string representation of the zettel identifier.
func (m *Meta) GetTitle() string {
	if title, found := m.Get(KeyTitle); found {
		return string(title)
	}
	return m.Zid.String()
}

// All returns an iterator over all key/value pairs, except the zettel identifier
// and computed values.
func (m *Meta) All() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		m.firstKeys()(yield)
		m.restKeys(notComputedKey)(yield)
	}
}

// Computed returns an iterator over all key/value pairs, except the zettel identifier.
func (m *Meta) Computed() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		m.firstKeys()(yield)
		m.restKeys(anyKey)(yield)
	}
}

// Rest returns an iterator over all key/value pairs, except the zettel identifier,
// the main keys, and computed values.
func (m *Meta) Rest() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		m.restKeys(notComputedKey)(yield)
	}
}

// ComputedRest returns an iterator over all key/value pairs, except the zettel identifier,
// and the main keys.
func (m *Meta) ComputedRest() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		m.restKeys(anyKey)(yield)
	}
}

func (m *Meta) firstKeys() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		for _, key := range firstKeys {
			if val, ok := m.pairs[key]; ok {
				if !yield(key, val) {
					return
				}
			}
		}
	}
}

func (m *Meta) restKeys(addKeyPred func(Key) bool) iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		keys := slices.Sorted(maps.Keys(m.pairs))
		for _, key := range keys {
			if !slices.Contains(firstKeys, key) && addKeyPred(key) {
				if !yield(key, m.pairs[key]) {
					return
				}
			}
		}
	}
}

func notComputedKey(key string) bool { return !IsComputed(key) }
func anyKey(string) bool             { return true }

// Delete removes a key from the data.
func (m *Meta) Delete(key string) {
	if key != KeyID {
		delete(m.pairs, key)
	}
}

// Equal compares to metas for equality.
func (m *Meta) Equal(o *Meta, allowComputed bool) bool {
	if m == nil && o == nil {
		return true
	}
	if m == nil || o == nil || m.Zid != o.Zid {
		return false
	}
	tested := make(strfun.Set, len(m.pairs))
	for k, v := range m.pairs {
		tested.Set(k)
		if !equalValue(k, v, o, allowComputed) {
			return false
		}
	}
	for k, v := range o.pairs {
		if !tested.Has(k) && !equalValue(k, v, m, allowComputed) {
			return false
		}
	}
	return true
}

func equalValue(key string, val Value, other *Meta, allowComputed bool) bool {
	if allowComputed || !IsComputed(key) {
		if valO, found := other.pairs[key]; !found || val != valO {
			return false
		}
	}
	return true
}

// Sanitize all metadata keys and values, so that they can be written safely into a file.
func (m *Meta) Sanitize() {
	if m == nil {
		return
	}
	for k, v := range m.pairs {
		m.pairs[RemoveNonGraphic(k)] = Value(RemoveNonGraphic(string(v)))
	}
}

// RemoveNonGraphic changes the given string not to include non-graphical characters.
// It is needed to sanitize meta data.
func RemoveNonGraphic(s string) string {
	if s == "" {
		return ""
	}
	pos := 0
	var sb strings.Builder
	for pos < len(s) {
		nextPos := strings.IndexFunc(s[pos:], func(r rune) bool { return !unicode.IsGraphic(r) })
		if nextPos < 0 {
			break
		}
		sb.WriteString(s[pos:nextPos])
		sb.WriteByte(' ')
		_, size := utf8.DecodeRuneInString(s[nextPos:])
		pos = nextPos + size
	}
	if pos == 0 {
		return strings.TrimSpace(s)
	}
	sb.WriteString(s[pos:])
	return strings.TrimSpace(sb.String())
}
