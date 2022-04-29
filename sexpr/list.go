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

import (
	"bytes"
	"io"
)

// List is a sequence of values, including sub-lists.
type List struct {
	val []Value
}

// NewList creates a new list with the given values.
func NewList(lstVal ...Value) *List {
	for _, v := range lstVal {
		if v == nil {
			return nil
		}
	}
	return &List{lstVal}
}

// Append some more value to a list.
func (lst *List) Append(lstVal ...Value) {
	for _, v := range lstVal {
		if v == nil {
			return
		}
	}
	lst.val = append(lst.val, lstVal...)
}

// Extend the list by another
func (lst *List) Extend(o *List) {
	if o != nil {
		for _, v := range o.val {
			if v == nil {
				return
			}
		}
		lst.val = append(lst.val, o.val...)
	}
}

// GetValue returns the list value.
func (lst *List) GetValue() []Value {
	if lst == nil {
		return nil
	}
	return lst.val
}

// Equal retruns true if the other value is equal to this one.
func (lst *List) Equal(other Value) bool {
	if lst == nil || other == nil {
		return lst == other
	}
	o, ok := other.(*List)
	if !ok || len(lst.val) != len(o.val) {
		return false
	}
	for i, val := range lst.val {
		if !val.Equal(o.val[i]) {
			return false
		}
	}
	return true
}

var (
	space  = []byte{' '}
	lParen = []byte{'('}
	rParen = []byte{')'}
)

// Encode the list.
func (lst *List) Encode(w io.Writer) (int, error) {
	length, err := w.Write(lParen)
	if err != nil {
		return length, err
	}
	for i, val := range lst.val {
		if i > 0 {
			l, err2 := w.Write(space)
			length += l
			if err2 != nil {
				return length, err2
			}
		}
		l, err2 := val.Encode(w)
		length += l
		if err2 != nil {
			return length, err2
		}
	}
	l, err := w.Write(rParen)
	return length + l, err
}

func (lst *List) String() string {
	var buf bytes.Buffer
	if _, err := lst.Encode(&buf); err != nil {
		return err.Error()
	}
	return buf.String()
}
