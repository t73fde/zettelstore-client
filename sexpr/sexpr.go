//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package sexpr allows to work with symbolic expressions, s-expression.
package sexpr

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Value is a generic value, the set of all possible values of a s-expression.
type Value interface {
	Equal(Value) bool
	Encode(io.Writer) (int, error)
	String() string
}

// ---------------------------------------------------------------------------

// Symbol is a value that identifies something.
type Symbol struct {
	val string
}

var (
	symbolMx  sync.Mutex // protects symbolMap
	symbolMap = map[string]*Symbol{}
)

// NewSymbol creates or reuses a symbol with the given string representation.
func NewSymbol(symVal string) *Symbol {
	if symVal == "" {
		return nil
	}
	v := strings.ToUpper(symVal)
	symbolMx.Lock()
	result, found := symbolMap[v]
	if !found {
		result = &Symbol{v}
		symbolMap[v] = result
	}
	symbolMx.Unlock()
	return result
}

// GetValue returns the string value of the symbol.
func (sym *Symbol) GetValue() string { return sym.val }

// Equal retruns true if the other value is equal to this one.
func (sym *Symbol) Equal(other Value) bool {
	if sym == nil || other == nil {
		return sym == other
	}
	if o, ok := other.(*Symbol); ok {
		return strings.EqualFold(sym.val, o.val)
	}
	return false
}

// Encode the symbol.
func (sym *Symbol) Encode(w io.Writer) (int, error) {
	return io.WriteString(w, sym.val)
}
func (sym *Symbol) String() string { return sym.val }

// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------

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
func (lst *List) GetValue() []Value { return lst.val }

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

// ---------------------------------------------------------------------------

func GetSymbol(args []Value, idx int) (*Symbol, error) {
	if idx < 0 && len(args) <= idx {
		return nil, fmt.Errorf("index %d out of bounds: %v", idx, args)
	}
	if val, ok := args[idx].(*Symbol); ok {
		return val, nil
	}
	return nil, fmt.Errorf("%v / %d is not a symbol", args[idx], idx)
}

func GetString(args []Value, idx int) (string, error) {
	if idx < 0 && len(args) <= idx {
		return "", fmt.Errorf("index %d out of bounds: %v", idx, args)

	}
	if val, ok := args[idx].(*String); ok {
		return val.GetValue(), nil
	}
	if val, ok := args[idx].(*Symbol); ok {
		return val.GetValue(), nil
	}
	return "", fmt.Errorf("%v / %d is not a string", args[idx], idx)
}

func GetList(args []Value, idx int) (*List, error) {
	if idx < 0 && len(args) <= idx {
		return nil, fmt.Errorf("index %d out of bounds: %v", idx, args)

	}
	if val, ok := args[idx].(*List); ok {
		return val, nil
	}
	return nil, fmt.Errorf("%v / %d is not a list", args[idx], idx)

}
