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

// String is a string value without any restrictions.
type String struct {
	val string
}

// NewString creates a new string with the given value.
func NewString(strVal string) *String { return &String{strVal} }

// GetValue returns the string value.
func (str *String) GetValue() string { return str.val }

// Equal retruns true if the other value is equal to this one.
func (str *String) Equal(other Value) bool {
	if str == nil || other == nil {
		return str == other
	}
	if o, ok := other.(*String); ok {
		return str.val == o.val
	}
	return false
}

var (
	quote        = []byte{'"'}
	encBackslash = []byte{'\\', '\\'}
	encQuote     = []byte{'\\', '"'}
	encNewline   = []byte{'\\', 'n'}
	encTab       = []byte{'\\', 't'}
	encCr        = []byte{'\\', 'r'}
	encUnicode   = []byte{'\\', 'u', '0', '0', '0', '0'}
	encHex       = []byte("0123456789ABCDEF")
)

// Encode the string value
func (str *String) Encode(w io.Writer) (int, error) {
	length, err := w.Write(quote)
	if err != nil {
		return length, err
	}
	last := 0
	for i, ch := range str.val {
		var b []byte
		switch ch {
		case '\t':
			b = encTab
		case '\r':
			b = encCr
		case '\n':
			b = encNewline
		case '"':
			b = encQuote
		case '\\':
			b = encBackslash
		default:
			if ch >= ' ' {
				continue
			}
			b = encUnicode
			b[2] = '0'
			b[3] = '0'
			b[4] = encHex[ch>>4]
			b[5] = encHex[ch&0xF]
		}
		l, err2 := io.WriteString(w, str.val[last:i])
		length += l
		if err2 != nil {
			return length, err2
		}
		l, err2 = w.Write(b)
		length += l
		if err2 != nil {
			return length, err2
		}
		last = i + 1
	}
	l, err := io.WriteString(w, str.val[last:])
	length += l
	if err != nil {
		return length, err
	}
	l, err = w.Write(quote)
	return length + l, err
}

func (str *String) String() string {
	var buf bytes.Buffer
	if _, err := str.Encode(&buf); err != nil {
		return err.Error()
	}
	return buf.String()
}
