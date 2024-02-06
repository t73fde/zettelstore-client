//-----------------------------------------------------------------------------
// Copyright (c) 2020-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2020-present Detlef Stern
//-----------------------------------------------------------------------------

// Package attrs stores attributes of zettel parts.
package attrs

import (
	"strings"

	"zettelstore.de/client.fossil/maps"
)

// Attributes store additional information about some node types.
type Attributes map[string]string

// IsEmpty returns true if there are no attributes.
func (a Attributes) IsEmpty() bool { return len(a) == 0 }

// DefaultAttribute is the value of the key of the default attribute
const DefaultAttribute = "-"

// HasDefault returns true, if the default attribute "-" has been set.
func (a Attributes) HasDefault() bool {
	if a != nil {
		_, ok := a[DefaultAttribute]
		return ok
	}
	return false
}

// RemoveDefault removes the default attribute
func (a Attributes) RemoveDefault() Attributes {
	if a != nil {
		a.Remove(DefaultAttribute)
	}
	return a
}

// Keys returns the sorted list of keys.
func (a Attributes) Keys() []string { return maps.Keys(a) }

// Get returns the attribute value of the given key and a succes value.
func (a Attributes) Get(key string) (string, bool) {
	if a != nil {
		value, ok := a[key]
		return value, ok
	}
	return "", false
}

// Clone returns a duplicate of the attribute.
func (a Attributes) Clone() Attributes {
	if a == nil {
		return nil
	}
	attrs := make(map[string]string, len(a))
	for k, v := range a {
		attrs[k] = v
	}
	return attrs
}

// Set changes the attribute that a given key has now a given value.
func (a Attributes) Set(key, value string) Attributes {
	if a == nil {
		return map[string]string{key: value}
	}
	a[key] = value
	return a
}

// Remove the key from the attributes.
func (a Attributes) Remove(key string) Attributes {
	if a != nil {
		delete(a, key)
	}
	return a
}

// AddClass adds a value to the class attribute.
func (a Attributes) AddClass(class string) Attributes {
	if a == nil {
		return map[string]string{"class": class}
	}
	classes := a.GetClasses()
	for _, cls := range classes {
		if cls == class {
			return a
		}
	}
	classes = append(classes, class)
	a["class"] = strings.Join(classes, " ")
	return a
}

// GetClasses returns the class values as a string slice
func (a Attributes) GetClasses() []string {
	if a == nil {
		return nil
	}
	classes, ok := a["class"]
	if !ok {
		return nil
	}
	return strings.Fields(classes)
}

// HasClass returns true, if attributes contains the given class.
func (a Attributes) HasClass(s string) bool {
	if a == nil {
		return false
	}
	classes, found := a["class"]
	if !found {
		return false
	}
	return strings.Contains(" "+classes+" ", " "+s+" ")
}
