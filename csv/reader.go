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
	"unicode/utf8"
)

// A helper interface for a general CSV reader. Adheres to encoding/csv Reader
// in the standard Go library as well as the Reader implemented by this
// package.
type CsvReader interface {
	// Read reads one record from r. The record is a slice of strings with each
	// string representing one field.
	Read() (record []string, err error)

	// ReadAll reads all the remaining records from r. Each record is a slice of
	// fields. A successful call returns err == nil, not err == EOF. Because
	// ReadAll is defined to read until EOF, it does not treat end of file as an
	// error to be reported.
	ReadAll() (records [][]string, err error)
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

// ReadAll reads all the remaining records from r. Each record is a slice of
// fields. A successful call returns err == nil, not err == EOF. Because
// ReadAll is defined to read until EOF, it does not treat end of file as an
// error to be reported.
func (r *Reader) ReadAll() ([][]string, error) {
	allRows := make([][]string, 0, 1)
	for {
		fields, err := r.Read()
		if err == io.EOF {
			return allRows, nil
		}
		if err != nil {
			return nil, err
		}
		allRows = append(allRows, fields)
	}
}

// Read reads one record from r. The record is a slice of strings with each
// string representing one field.
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
	s := bytes.Buffer{}
	for {
		char, _, err := r.r.ReadRune()
		if err != nil || char == r.opts.Delimiter {
			// TODO Can a non quoted string be escaped? In that case, it should be
			// handled here. Should probably have a look at how Python's csv module
			// is handling this.

			// Putting it back for the outer loop to read separators. This makes more
			// compatible with readQuotedField().
			r.r.UnreadRune(char)

			return s.String(), err
		} else {
			s.WriteRune(char)
		}
		if ok, _ := r.nextIsLineTerminator(); ok {
			return s.String(), nil
		}
	}
}
