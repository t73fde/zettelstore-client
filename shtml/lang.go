//-----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
//-----------------------------------------------------------------------------

package shtml

import (
	"strings"

	"t73f.de/r/zsc/api"
)

// LangStack is a stack to store the nesting of "lang" attribute values.
// It is used to generate typographically correct quotes.
type LangStack []string

// NewLangStack creates a new language stack.
func NewLangStack(lang string) LangStack {
	ls := make([]string, 1, 16)
	ls[0] = lang
	return ls
}

// Reset restores the language stack to its initial value.
func (ls *LangStack) Reset() {
	*ls = (*ls)[0:1]
}

// Push adds a new language value.
func (ls *LangStack) Push(lang string) {
	*ls = append(*ls, lang)
}

// Pop removes the topmost language value.
func (ls *LangStack) Pop() {
	*ls = (*ls)[0 : len(*ls)-1]
}

// Top returns the topmost language value.
func (ls *LangStack) Top() string {
	return (*ls)[len(*ls)-1]
}

// Dup duplicates the topmost language value.
func (ls *LangStack) Dup() {
	*ls = append(*ls, (*ls)[len(*ls)-1])
}

// QuoteInfo contains language specific data about quotes.
type QuoteInfo struct {
	primLeft, primRight string
	secLeft, secRight   string
	nbsp                bool
}

// GetPrimary returns the primary left and right quote entity.
func (qi *QuoteInfo) GetPrimary() (string, string) {
	return qi.primLeft, qi.primRight
}

// GetSecondary returns the secondary left and right quote entity.
func (qi *QuoteInfo) GetSecondary() (string, string) {
	return qi.secLeft, qi.secRight
}

// GetQuotes returns quotes based on a nesting level.
func (qi *QuoteInfo) GetQuotes(level uint) (string, string) {
	if level%2 == 0 {
		return qi.GetPrimary()
	}
	return qi.GetSecondary()
}

// GetNBSp returns true, if there must be a non-breaking space between the
// quote entities and the quoted text.
func (qi *QuoteInfo) GetNBSp() bool { return qi.nbsp }

var langQuotes = map[string]*QuoteInfo{
	"":              {"&quot;", "&quot;", "&quot;", "&quot;", false},
	api.ValueLangEN: {"&ldquo;", "&rdquo;", "&lsquo;", "&rsquo;", false},
	"de":            {"&bdquo;", "&ldquo;", "&sbquo;", "&lsquo;", false},
	"fr":            {"&laquo;", "&raquo;", "&lsaquo;", "&rsaquo;", true},
}

// GetQuoteInfo returns language specific data about quotes.
func GetQuoteInfo(lang string) *QuoteInfo {
	langFields := strings.FieldsFunc(lang, func(r rune) bool { return r == '-' || r == '_' })
	for len(langFields) > 0 {
		langSup := strings.Join(langFields, "-")
		quotes, ok := langQuotes[langSup]
		if ok {
			return quotes
		}
		langFields = langFields[0 : len(langFields)-1]
	}
	return langQuotes[""]
}
