// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

// A CSV implementation inspired by Python's CSV module. Supports custom CSV
// formats. Currently only writing CSV files is supported.
package csv

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// A helper interface for a general CSV writer. Adheres to encoding/csv Writer
// in the standard go library as well as the Writer implemented by this
// package.
type CsvWriter interface {
	// Currently no errors are possible.
	Error() error

	// Flush writes any buffered data to the underlying io.Writer.
	// To check if an error occurred during the Flush, call Error.
	Flush()

	// Writer writes a single CSV record to w along with any necessary quoting.
	// A record is a slice of strings with each string being one field.
	Write(record []string) error

	// WriteAll writes multiple CSV records to w using Write and then calls Flush.
	WriteAll(records [][]string) error
}

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

// bufio that supports putting stuff back into it.
type unReader struct {
	r *bufio.Reader
	b *bytes.Buffer
}

func NewUnReader(r io.Reader) *unReader {
	return &unReader{
		r: bufio.NewReader(r),
		b: new(bytes.Buffer),
	}
}

func (u *unReader) ReadRune() (rune, int, error) {
	if u.b.Len() > 0 {
		return u.b.ReadRune()
	} else {
		return u.r.ReadRune()
	}
}

func (u *unReader) UnreadRune(r rune) {
	// Poor man's prepend
	var tmpBuf bytes.Buffer
	tmpBuf.WriteRune(r)
	tmpBuf.ReadFrom(u.b)

	u.b = &tmpBuf
}

func (u *unReader) NextIsString(s string) (bool, error) {
	// Fill up bytes in buffer
	for u.b.Len() < len(s) {
		r, _, err := u.r.ReadRune()
		if err != nil {
			return false, err
		}
		u.b.WriteRune(r)
	}
	return strings.HasPrefix(u.b.String(), s), nil
}

type Reader struct {
	opts Dialect
	r    *unReader
}

func NewReader(r io.Reader) *Reader {
	opts := Dialect{}
	opts.setDefaults()
	return NewDialectReader(r, opts)
}

// Create a custom CSV writer.
func NewDialectReader(r io.Reader, opts Dialect) *Reader {
	opts.setDefaults()
	return &Reader{
		opts: opts,
		r:    NewUnReader(r),
	}
}

func (r *Reader) Read() ([]string, error) {
	// TODO: Possible optimization; store the maximum number of columns for
	// faster preallocation.
	record := make([]string, 0, 2)

	for {
		field, err := r.readField()
		record = append(record, field)
		if err != nil {
			return record, err
		}

		if nextIsLineTerminator, _ := r.nextIsLineTerminator(); nextIsLineTerminator {
			// Skipping so that next read call is good to go.
			err = r.skipLineTerminator()
			// Error is not expected since it should be in the Unreader buffer, but
			// might as well return it just in case.
			return record, err
		}
		nextIsDelimiter, err := r.nextIsDelimiter()
		if !nextIsDelimiter {
			// Herein lies the devil!
			return record, err
		} else {
			r.skipDelimiter()
		}
	}
}

func (r *Reader) readField() (string, error) {
	char, _, err := r.r.ReadRune()
	if err != nil {
		return "", err
	}

	// Let the next individual reader functions handle this.
	r.r.UnreadRune(char)

	if char == r.opts.QuoteChar {
		return r.readQuotedField()
	} else {
		return r.readUnquotedField()
	}
}

func (r *Reader) nextIsLineTerminator() (bool, error) {
	return r.r.NextIsString(r.opts.LineTerminator)
}

func (r *Reader) nextIsDelimiter() (bool, error) {
	return r.r.NextIsString(string(r.opts.Delimiter))
}

func (r *Reader) skipLineTerminator() error {
	for _ = range r.opts.LineTerminator {
		_, _, err := r.r.ReadRune()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) skipDelimiter() error {
	_, _, err := r.r.ReadRune()
	return err
}

func (r *Reader) readQuotedField() (string, error) {
	char, _, err := r.r.ReadRune()
	if err != nil {
		return "", err
	}
	if char != r.opts.QuoteChar {
		panic("Expected first character to be quote character.")
	}

	s := bytes.Buffer{}
	for {
		char, _, err := r.r.ReadRune()
		if err != nil {
			return s.String(), err
		}
		if char != r.opts.QuoteChar {
			s.WriteRune(char)
		} else {
			switch r.opts.DoubleQuote {
			case DoDoubleQuote:
				char, _, err = r.r.ReadRune()
				if err != nil {
					return s.String(), err
				}
				if char == r.opts.QuoteChar {
					s.WriteRune(char)
				} else {
					r.r.UnreadRune(char)
					return s.String(), nil
				}
			case NoDoubleQuote:
				if s.Len() == 0 {
					return s.String(), nil
				}
				lastRune, size := utf8.DecodeLastRuneInString(s.String())
				if lastRune == utf8.RuneError && size == 1 {
					panic("Field contained malformed rune.")
				}
				if lastRune == r.opts.EscapeChar {
					// Replace previous escape character.
					s.Truncate(s.Len() - utf8.RuneLen(char))
					s.WriteRune(char)
				} else {
					return s.String(), nil
				}
			default:
				panic("Unrecognized double quote mode.")
			}
		}
	}
}

func (r *Reader) readUnquotedField() (string, error) {
	// TODO: Use bytes.Buffer
	s := ""
	for {
		char, _, err := r.r.ReadRune()
		if err != nil || char == r.opts.Delimiter {
			// TODO Can a non quoted string be escaped? In that case, it should be
			// handled here. Should probably have a look at how Python's csv module
			// is handling this.

			// Putting it back for the outer loop to read separators. This makes more
			// compatible with readQuotedField().
			r.r.UnreadRune(char)

			return s, err
		} else {
			s = s + string(char)
		}
		if ok, _ := r.nextIsLineTerminator(); ok {
			return s, nil
		}
	}
}

type Writer struct {
	opts Dialect
	w    *bufio.Writer
}

// Create a writer that adheres to the Golang CSV writer.
func NewWriter(w io.Writer) Writer {
	opts := Dialect{}
	opts.setDefaults()
	return Writer{
		opts: opts,
		w:    bufio.NewWriter(w),
	}
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

func (w Writer) fieldNeedsQuote(field string) bool {
	switch w.opts.Quoting {
	case QuoteNone:
		return false
	case QuoteAll:
		return true
	case QuoteNonNumeric:
		return !isNumeric(field)
	case QuoteMinimal:
		// TODO: Can be improved by making a single search with trie.
		// See https://docs.python.org/2/library/csv.html#csv.QUOTE_MINIMAL for info on this.
		return strings.Contains(field, w.opts.LineTerminator) || strings.ContainsRune(field, w.opts.Delimiter) || strings.ContainsRune(field, w.opts.QuoteChar)
	default:
		panic("Unexpected quoting.")
	}
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
	default:
		panic("Unrecognized double quote type.")
	}
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
	} else {
		return w.writeString(field)
	}
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
