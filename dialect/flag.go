// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

// Helpers that makes it easy to build CSV dialects.  This API is currently in
// alpha. Feel free to discuss it on
// https://github.com/jensrantil/go-csv/issues.
package dialect

import (
	"flag"
	csv "github.com/JensRantil/go-csv"
	"strings"
	"unicode/utf8"
  "errors"
)

type DialectBuilder struct {
  quoteCharString *string
  escapeCharString *string
  delimiterCharString *string
  flagSet *flag.FlagSet
}

// Construct a CSV Dialect from command line using the `flag` package. This is
// three steps: First, call this function and store the handler. Optionally
// register other flags. Call `flag.Parse()`. A dialect can then be constructed
// by calling `DialectBuilder.Dialect()`.
func FromCommandLine() *DialectBuilder {
  return FromFlagSet(flag.CommandLine)
}

// Constructs a CSV Dialect from a specific flagset. Essentially the same as
// `FromCommandLine()`, except it supports a custom FlagSet. See
// `FromCommandLine()` for a description on how to use this one.
func FromFlagSet(f *flag.FlagSet) *DialectBuilder {
  p := DialectBuilder{}
  p.delimiterCharString = f.String("fields-terminated-by", "\t", "character to terminate fields by")
  p.quoteCharString = f.String("fields-optionally-enclosed-by", "\"", "character to enclose fields with when needed")
  p.escapeCharString = f.String("fields-escaped-by", "\\", "character to escape special characters with")
  p.flagSet = f
  return &p
}

// Construct a Dialect from a FlagSet. Make sure to parse the FlagSet before
// calling this.
func (p *DialectBuilder) Dialect() (*csv.Dialect, error) {
  if !p.flagSet.Parsed() {
    // Sure, could call flagSet.Parse() here. However, we don't know if the
    // user would like to parse something else than argv. Therefor, letting the
    // user decide.
    return nil, errors.New("FlagSet has not been parsed before calling this function.")
  }

  // `FlagSet`s don't have a rune type. Using string instead, but that adds
  // some manual error checking.
	if utf8.RuneCountInString(*p.quoteCharString) > 1 {
		return nil, errors.New("-fields-optionally-enclosed-by can't be more than one character.")
	}
	if utf8.RuneCountInString(*p.escapeCharString) > 1 {
		return nil, errors.New("-fields-escaped-by can't be more than one character.")
	}
	if utf8.RuneCountInString(*p.quoteCharString) < 1 {
		return nil, errors.New("-fields-optionally-enclosed-by can't be an empty string.")
	}
	if utf8.RuneCountInString(*p.escapeCharString) < 1 {
		return nil, errors.New("-fields-escaped-by can't be an empty string.")
	}

	quoteChar, _, _ := strings.NewReader(*p.quoteCharString).ReadRune()
	escapeChar, _, _ := strings.NewReader(*p.escapeCharString).ReadRune()
	delimiterChar, _, _ := strings.NewReader(*p.delimiterCharString).ReadRune()
	dialect := csv.Dialect{
		Delimiter:   delimiterChar,
		QuoteChar:   quoteChar,
		EscapeChar:  escapeChar,
		DoubleQuote: csv.NoDoubleQuote,
	}

  return &dialect, nil
}
