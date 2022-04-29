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
	"errors"
	"unicode"

	"zettelstore.de/c/input"
)

// Error values
var (
	ErrEOF = errors.New("unexpected eof")
)

func ReadString(src string) (Value, error) {
	return ReadBytes([]byte(src))
}

func ReadBytes(src []byte) (Value, error) {
	inp := input.NewInput(src)
	return ReadValue(inp)
}

func ReadValue(inp *input.Input) (Value, error) {
	skipSpace(inp)
	return readValue(inp)
}

func skipSpace(inp *input.Input) {
	for unicode.IsSpace(inp.Ch) {
		inp.Next()
	}
}

func readValue(inp *input.Input) (Value, error) {
	switch inp.Ch {
	case input.EOS:
		return nil, ErrEOF
	case '(': // List
		return readList(inp)
	case '"': // String
		return readString(inp)
	default: // Must be symbol
		return readSymbol(inp)
	}
}

func readSymbol(inp *input.Input) (Value, error) {
	var buf bytes.Buffer
	buf.WriteRune(inp.Ch)
	for {
		inp.Next()
		switch inp.Ch {
		case input.EOS, '(', ')', '"':
			return NewSymbol(buf.String()), nil
		}
		if unicode.In(inp.Ch, unicode.Space, unicode.C) {
			return NewSymbol(buf.String()), nil
		}
		buf.WriteRune(inp.Ch)
	}
}

func readString(inp *input.Input) (Value, error) {
	var buf bytes.Buffer
	for {
		inp.Next()
		switch inp.Ch {
		case input.EOS:
			return nil, ErrEOF
		case '"':
			inp.Next() // skip '"'
			return NewString(buf.String()), nil
		}
		buf.WriteRune(inp.Ch)
	}
}

func readList(inp *input.Input) (Value, error) {
	inp.Next() // Skip '('
	elems := []Value{}
	for {
		skipSpace(inp)
		switch inp.Ch {
		case input.EOS:
			return nil, ErrEOF
		case ')':
			inp.Next() // Skip ')'
			return NewList(elems...), nil
		}
		val, err := readValue(inp)
		if err != nil {
			return nil, err
		}
		elems = append(elems, val)
	}
}
