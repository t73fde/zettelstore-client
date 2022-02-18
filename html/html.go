//-----------------------------------------------------------------------------
// Copyright (c) 2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package html provides types, constants and function to work with HTML.
package html

import (
	"io"
	"strings"
)

const (
	htmlQuot     = "&quot;" // longer than "&#34;", but often requested in standards
	htmlAmp      = "&amp;"
	htmlLt       = "&lt;"
	htmlGt       = "&gt;"
	htmlNull     = "\uFFFD"
	htmlVisSpace = "\u2423"
)

var (
	htmlEscapes = []string{`&`, htmlAmp,
		`<`, htmlLt,
		`>`, htmlGt,
		`"`, htmlQuot,
		"\000", htmlNull,
	}
	htmlEscaper    = strings.NewReplacer(htmlEscapes...)
	htmlVisEscapes = append(htmlEscapes,
		" ", htmlVisSpace,
		"\u00a0", htmlVisSpace,
	)
	htmlVisEscaper = strings.NewReplacer(htmlVisEscapes...)
)

// Escape writes to w the escaped HTML equivalent of the given string.
func Escape(w io.Writer, s string) (int, error) { return htmlEscaper.WriteString(w, s) }

// EscapeVisible writes to w the escaped HTML equivalent of the given string.
// Each space is written as U-2423.
func EscapeVisible(w io.Writer, s string) (int, error) { return htmlVisEscaper.WriteString(w, s) }

var (
	escQuot = []byte(htmlQuot) // longer than "&#34;", but often requested in standards
	escAmp  = []byte(htmlAmp)
	escNull = []byte(htmlNull)
)

// AttributeEscape writes to w the escaped HTML equivalent of the given string to be used
// in attributes.
func AttributeEscape(w io.Writer, s string) {
	last := 0
	var html []byte
	lenS := len(s)
	for i := 0; i < lenS; i++ {
		switch s[i] {
		case '\000':
			html = escNull
		case '"':
			html = escQuot
		case '&':
			html = escAmp
		default:
			continue
		}
		io.WriteString(w, s[last:i])
		w.Write(html)
		last = i + 1
	}
	io.WriteString(w, s[last:])
}

var unsafeSnippets = []string{
	"<script", "</script",
	"<iframe", "</iframe",
}

// IsSave returns true if the given string does not contain unsafe HTML elements.
func IsSave(s string) bool {
	lower := strings.ToLower(s)
	for _, snippet := range unsafeSnippets {
		if strings.Contains(lower, snippet) {
			return false
		}
	}
	return true
}
