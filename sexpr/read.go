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
	"strconv"
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
	return parseValue(inp)
}

func skipSpace(inp *input.Input) {
	for unicode.IsSpace(inp.Ch) {
		inp.Next()
	}
}

func parseValue(inp *input.Input) (Value, error) {
	switch inp.Ch {
	case input.EOS:
		return nil, ErrEOF
	case '(': // List
		return parseList(inp)
	case '"': // String
		return parseString(inp)
	default: // Must be symbol
		return parseSymbol(inp)
	}
}

func parseSymbol(inp *input.Input) (Value, error) {
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

func parseString(inp *input.Input) (Value, error) {
	var buf bytes.Buffer
	for {
		inp.Next()
		switch inp.Ch {
		case input.EOS:
			return nil, ErrEOF
		case '"':
			inp.Next() // skip '"'
			return NewString(buf.String()), nil
		case '\\':
			inp.Next()
			switch inp.Ch {
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case 'n':
				buf.WriteByte('\n')
			case 'x':
				parseRune(inp, &buf, 2)
			case 'u':
				parseRune(inp, &buf, 4)
			case 'U':
				parseRune(inp, &buf, 6)
			default:
				buf.WriteRune(inp.Ch)
			}
		default:
			buf.WriteRune(inp.Ch)
		}
	}
}
func parseRune(inp *input.Input, buf *bytes.Buffer, numDigits int) {
	endPos := inp.Pos + numDigits
	if len(inp.Src) <= endPos {
		buf.WriteRune(inp.Ch)
		return
	}
	n, err := strconv.ParseInt(string(inp.Src[inp.Pos+1:endPos+1]), 16, 4*numDigits)
	if err != nil {
		buf.WriteRune(inp.Ch)
		return
	}
	buf.WriteRune(rune(n))

	switch numDigits {
	case 6:
		inp.Next()
		inp.Next()
		fallthrough
	case 4:
		inp.Next()
		inp.Next()
		fallthrough
	case 2:
		inp.Next()
		inp.Next()
	default:
		panic(numDigits)
	}
}

func parseList(inp *input.Input) (Value, error) {
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
		val, err := parseValue(inp)
		if err != nil {
			return nil, err
		}
		elems = append(elems, val)
	}
}
