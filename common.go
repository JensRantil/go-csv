// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

// A CSV implementation inspired by Python's CSV module. Supports custom CSV
// formats. Currently only writing CSV files is supported.
package csv

import (
	"unicode"
)

// Values Dialect.Quoting can take.
const (
	QuoteDefault    = iota // See DefaultQuoting.
	QuoteAll        = iota
	QuoteMinimal    = iota
	QuoteNonNumeric = iota
	QuoteNone       = iota
)

// Values Dialect.DoubleQuote can take.
const (
	DoubleQuoteDefault = iota // See DefaultDoubleQuote.
	DoDoubleQuote      = iota
	NoDoubleQuote      = iota
)

// Default dialect.
const (
	DefaultDelimiter      = ' '
	DefaultQuoting        = QuoteMinimal
	DefaultDoubleQuote    = DoDoubleQuote
	DefaultEscapeChar     = '\\'
	DefaultQuoteChar      = '"'
	DefaultLineTerminator = "\n"
)

type Dialect struct {
	Delimiter      rune
	Quoting        int
	DoubleQuote    int
	EscapeChar     rune
	QuoteChar      rune
	LineTerminator string
}

func (wo *Dialect) setDefaults() {
	if wo.Delimiter == 0 {
		wo.Delimiter = DefaultDelimiter
	}
	if wo.Quoting == QuoteDefault {
		wo.Quoting = DefaultQuoting
	}
	if wo.LineTerminator == "" {
		wo.LineTerminator = DefaultLineTerminator
	}
	if wo.DoubleQuote == DoubleQuoteDefault {
		wo.DoubleQuote = DefaultDoubleQuote
	}
	if wo.QuoteChar == 0 {
		wo.QuoteChar = DefaultQuoteChar
	}
	if wo.EscapeChar == 0 {
		wo.EscapeChar = DefaultEscapeChar
	}
}

func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
