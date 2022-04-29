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
	"io"
	"strings"
	"sync"
)

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
