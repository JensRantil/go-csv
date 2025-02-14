// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"bufio"
	"io"
	"strings"
)

// A Writer writes records to a CSV encoded file.
//
// Can be created by calling either NewWriter or using NewDialectWriter.
type Writer struct {
	opts Dialect
	w    *bufio.Writer
}

// Create a writer that conforms to RFC 4180 and behaves identical as a
// encoding/csv.Reader.
//
// See `Default*` constants for default dialect used.
func NewWriter(w io.Writer) Writer {
	return NewDialectWriter(w, Dialect{})
}

// Create a custom CSV writer.
func NewDialectWriter(w io.Writer, opts Dialect) Writer {
	opts.setDefaults()
	return Writer{
		opts: opts,
		w:    bufio.NewWriter(w),
	}
}

// Error reports any error that has occurred during a previous Write or Flush.
func (w Writer) Error() error {
	_, err := w.w.Write(nil)
	return err
}

// Flush writes any buffered data to the underlying io.Writer.
// To check if an error occurred during the Flush, call Error.
func (w Writer) Flush() {
	w.w.Flush()
}

// Helper function that ditches the first return value of w.w.WriteString().
// Simplifies code.
func (w Writer) writeString(s string) error {
	_, err := w.w.WriteString(s)
	return err
}

func (w Writer) writeDelimiter() error {
	return w.writeRune(w.opts.Delimiter)
}

func (w Writer) fieldNeedsQuote(field string) bool {
	switch w.opts.Quoting {
	case QuoteNone:
		return false
	case QuoteAll:
		return true
	case QuoteNonNumeric:
		return !isNumeric(field)
	case QuoteNonNumericNonEmpty:
		return !(isNumeric(field) || isEmpty(field))
	case QuoteMinimal:
		// TODO: Can be improved by making a single search with trie.
		// See https://docs.python.org/2/library/csv.html#csv.QUOTE_MINIMAL for info on this.
		return strings.Contains(field, w.opts.LineTerminator) || strings.ContainsRune(field, w.opts.Delimiter) || strings.ContainsRune(field, w.opts.QuoteChar)
	}
	panic("Unexpected quoting.")
}

func (w Writer) writeRune(r rune) error {
	_, err := w.w.WriteRune(r)
	return err
}

func (w Writer) writeEscapeChar(r rune) error {
	switch w.opts.DoubleQuote {
	case DoDoubleQuote:
		return w.writeRune(r)
	case NoDoubleQuote:
		return w.writeRune(w.opts.EscapeChar)
	}
	panic("Unrecognized double quote type.")
}

func (w Writer) writeQuotedRune(r rune) error {
	switch r {
	case w.opts.EscapeChar:
		if err := w.writeEscapeChar(r); err != nil {
			return err
		}
	case w.opts.QuoteChar:
		if err := w.writeEscapeChar(r); err != nil {
			return err
		}
	}
	return w.writeRune(r)
}

func (w Writer) writeQuoted(field string) error {
	if err := w.writeRune(w.opts.QuoteChar); err != nil {
		return err
	}
	for _, r := range field {
		if err := w.writeQuotedRune(r); err != nil {
			return err
		}
	}
	return w.writeRune(w.opts.QuoteChar)
}

func (w Writer) writeField(field string) error {
	if w.fieldNeedsQuote(field) {
		return w.writeQuoted(field)
	}
	return w.writeString(field)
}

func (w Writer) writeNewline() error {
	return w.writeString(w.opts.LineTerminator)
}

// Writer writes a single CSV record to w along with any necessary quoting.
// A record is a slice of strings with each string being one field.
func (w Writer) Write(record []string) (err error) {
	for n, field := range record {
		if n > 0 {
			if err = w.writeDelimiter(); err != nil {
				return
			}
		}
		if err = w.writeField(field); err != nil {
			return
		}
	}
	err = w.writeNewline()
	return
}

// WriteAll writes multiple CSV records to w using Write and then calls Flush.
func (w Writer) WriteAll(records [][]string) (err error) {
	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return w.w.Flush()
}
