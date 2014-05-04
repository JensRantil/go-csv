// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

// A CSV implementation inspired by Python's CSV module. Supports custom CSV
// formats.
package csv

import (
	"unicode"
)

// Values Dialect.Quoting can take.
const (
	QuoteDefault    = iota // See DefaultQuoting.
	QuoteAll        = iota // Quotes around every field.
	QuoteMinimal    = iota // Quotes when needed.
	QuoteNonNumeric = iota // Quotes around non-numeric fields.

	// Never quote. Use with care. Could make things unparsable.
	QuoteNone = iota
)

// Values Dialect.DoubleQuote can take.
const (
	DoubleQuoteDefault = iota // See DefaultDoubleQuote.
	DoDoubleQuote      = iota // Escape using double escape characters.
	NoDoubleQuote      = iota // Escape using escape character.
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

// A Dialect specifies the format of a CSV file. This structure is used by a
// Reader or Writer to know how to operate on the file they are
// reading/writing.
type Dialect struct {
	// The delimiter that separates each field from another. Defaults to
	// DefaultDelimiter.
	Delimiter rune
	// What quoting mode to use. Defaults to DefaultQuoting.
	Quoting int
	// How to escape quotes. Defaults to DefaultDoubleQuote.
	DoubleQuote int
	// Character to use for escaping. Only used if DoubleQuote==NoDoubleQuote.
	// Defaults to DefaultEscapeChar.
	EscapeChar rune
	// Character to use as quotation mark around quoted fields. Defaults to
	// DefaultQuoteChar.
	QuoteChar rune
	// String that separates each record in a CSV file. Defaults to
	// DefaultLineTerminator.
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
