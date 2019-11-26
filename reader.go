// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"bufio"
	"bytes"
	"io"
	"unicode/utf8"
)

// A Reader reads records from a CSV-encoded file.
//
// Can be created by calling either NewReader or using NewDialectReader.
type Reader struct {
	opts                    Dialect
	r                       *bufio.Reader
	tmpBuf                  bytes.Buffer
	optimizedDelimiter      []byte
	optimizedLineTerminator []byte
}

// Creates a reader that conforms to RFC 4180 and behaves identical as a
// encoding/csv.Reader.
//
// See `Default*` constants for default dialect used.
func NewReader(r io.Reader) *Reader {
	return NewDialectReader(r, Dialect{})
}

// Create a custom CSV reader.
func NewDialectReader(r io.Reader, opts Dialect) *Reader {
	opts.setDefaults()
	return &Reader{
		opts:                    opts,
		r:                       bufio.NewReader(r),
		optimizedDelimiter:      []byte(string(opts.Delimiter)),
		optimizedLineTerminator: []byte(opts.LineTerminator),
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

	// Required by Go 1.0 to compile. Unreachable code.
	return allRows, nil
}

// Read reads one record from r. The record is a slice of strings with each
// string representing one field.
func (r *Reader) Read() ([]string, error) {
	// TODO: Possible optimization; store the maximum number of columns for
	// faster preallocation.
	record := make([]string, 0, 2)

	if err := r.skipComments(); err != nil {
		return record, err
	}

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

	// Required by Go 1.0 to compile. Unreachable code.
	return record, nil
}

func (r *Reader) readField() (string, error) {
	if islt, err := r.nextIsLineTerminator(); islt || err != nil {
		return "", err
	}

	char, _, err := r.r.ReadRune()
	if err != nil {
		return "", err
	}

	// Let the next individual reader functions handle this.
	r.r.UnreadRune()

	if char == r.opts.QuoteChar {
		return r.readQuotedField()
	}
	return r.readUnquotedField()
}

func (r *Reader) nextIsLineTerminator() (bool, error) {
	return r.nextIsBytes(r.optimizedLineTerminator)
}

func (r *Reader) nextIsDelimiter() (bool, error) {
	return r.nextIsBytes(r.optimizedDelimiter)
}

func (r *Reader) nextIsBytes(bs []byte) (bool, error) {
	n := len(bs)
	nextBytes, err := r.r.Peek(n)
	return bytes.Equal(nextBytes, bs), err
}

func (r *Reader) skipLineTerminator() error {
	_, err := r.r.Discard(len(r.optimizedLineTerminator))
	return err
}

func (r *Reader) skipComments() error {
	var n = 1
	var isComment bool
	for {
		nextBytes, err := r.r.Peek(n)
		if err != nil {
			return err
		}

		switch rune(nextBytes[n-1]) {
		case ' ', '\t': //skip space
			if !isComment {
				n += 1
				continue
			} else {
				return nil
			}

		case r.opts.Comment:
			_, err := r.r.Discard(n)
			if err != nil {
				return err
			}
			n = 1
			isComment = true

		default:
			if !isComment {
				return nil
			} else if nextIsLineTerminator, _ := r.nextIsLineTerminator(); nextIsLineTerminator {
				err = r.skipLineTerminator()
				if err != nil {
					return err
				}
				isComment = false
			} else if _, err := r.r.Discard(n); err != nil {
				return err
			}
			n = 1 //after discard or skip LineTermintator, reset n
		}
	}
	//skip until LineTerminator
	return nil
}

func (r *Reader) skipDelimiter() error {
	_, err := r.r.Discard(len(r.optimizedDelimiter))
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

	s := &r.tmpBuf
	defer r.tmpBuf.Reset() // TODO: Not using defer here is faster.
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
					r.r.UnreadRune()
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

	// Required by Go 1.0 to compile. Unreachable code.
	return s.String(), nil
}

func (r *Reader) readUnquotedField() (string, error) {
	// TODO: Use bytes.Buffer
	s := &r.tmpBuf
	defer r.tmpBuf.Reset() // TODO: Not using defer here is faster.
	for {
		char, _, err := r.r.ReadRune()
		if err != nil || char == r.opts.Delimiter {
			// TODO Can a non quoted string be escaped? In that case, it should be
			// handled here. Should probably have a look at how Python's csv module
			// is handling this.

			// Putting it back for the outer loop to read separators. This makes more
			// compatible with readQuotedField().
			r.r.UnreadRune()

			return s.String(), err
		} else {
			s.WriteRune(char)
		}
		if ok, _ := r.nextIsLineTerminator(); ok {
			return s.String(), nil
		}
	}

	// Required by Go 1.0 to compile. Unreachable code.
	return s.String(), nil
}
