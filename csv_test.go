// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"bytes"
	oldcsv "encoding/csv"
	"testing"
	"testing/quick"
)

func TestInterfaceAdherence(t *testing.T) {
	t.Parallel()

	var iface CsvWriter
	iface = NewWriter(new(bytes.Buffer))
	iface = NewDialectWriter(new(bytes.Buffer), Dialect{})
	iface = oldcsv.NewWriter(new(bytes.Buffer))

	// To get rid of compile-time warning that this variable is not used.
	iface.Flush()
}

// Execute a quicktest for a specific quoting.
func testQuick(t *testing.T, quoting int) {
	f := func(records [][]string, doubleQuote bool, escapeChar, quoteChar rune, del, lt string) bool {
		b1 := new(bytes.Buffer)
		dialect := Dialect{
			Quoting:        quoting,
			EscapeChar:     escapeChar,
			QuoteChar:      quoteChar,
			Delimiter:      del,
			LineTerminator: lt,
		}
		if doubleQuote {
			dialect.DoubleQuote = DoDoubleQuote
		} else {
			dialect.DoubleQuote = NoDoubleQuote
		}
		w := NewDialectWriter(b1, dialect)
		for _, record := range records {
			w.Write(record)
		}
		w.Flush()

		b2 := new(bytes.Buffer)
		w = NewDialectWriter(b2, dialect)
		w.WriteAll(records)

		return bytes.Equal(b1.Bytes(), b2.Bytes())
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Run quicktest using various quoting types
func TestQuick(t *testing.T) {
	t.Parallel()

	testQuick(t, QuoteAll)
	testQuick(t, QuoteNone)
	testQuick(t, QuoteMinimal)
	testQuick(t, QuoteNonNumeric)
}

func TestBasic(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	w := NewWriter(b)
	w.Write([]string{
		"a",
		"b",
		"c",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "a b c\n" {
		t.Error("Unexpected output:", s)
	}

	w.Write([]string{
		"d",
		"e",
		"f",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "a b c\nd e f\n" {
		t.Error("Unexpected output:", s)
	}
}

func TestMinimalQuoting(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	w := NewWriter(b)

	if w.opts.Quoting != QuoteMinimal {
		t.Fatal("Unexpected quoting.")
	}
	if s := "b c"; !w.fieldNeedsQuote(s) {
		t.Error("Expected field to need quoting:", s)
	}

	w.Write([]string{
		"a",
		"b c",
		"d",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "a \"b c\" d\n" {
		t.Error("Unexpected output:", s)
	}
}

func TestIsNumeric(t *testing.T) {
	t.Parallel()

	notNumeric := []string{
		"",
		" ",
		"a",
		"1a",
		"a1",
	}
	numeric := []string{
		"1",
		"11",
		"123456789",
	}
	for _, item := range numeric {
		if !isNumeric(item) {
			t.Error("Should be numeric:", item)
		}
	}
	for _, item := range notNumeric {
		if isNumeric(item) {
			t.Error("Should not be numeric:", item)
		}
	}
}

func TestNumericQuoting(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	dialect := Dialect{
		Quoting: QuoteNonNumeric,
	}
	w := NewDialectWriter(b, dialect)
	w.Write([]string{
		"a",
		"112",
		"b c",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "\"a\" 112 \"b c\"\n" {
		t.Error("Unexpected output:", s)
	}
}

func TestEscaping(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	w := NewWriter(b)
	w.Write([]string{
		"a",
		"\"",
		"b c",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "a \"\"\"\" \"b c\"\n" {
		t.Error("Unexpected output:", s)
	}

	b.Reset()
	dialect := Dialect{
		DoubleQuote: NoDoubleQuote,
	}
	w = NewDialectWriter(b, dialect)
	w.Write([]string{
		"a",
		"\"",
		"b c",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "a \"\\\"\" \"b c\"\n" {
		t.Error("Unexpected output:", s)
	}
}

func TestNewLineRecord(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	w := NewWriter(b)
	w.Write([]string{
		"a",
		"he\nllo",
		"b c",
	})
	w.Flush()
	if s := string(b.Bytes()); s != "a \"he\nllo\" \"b c\"\n" {
		t.Error("Unexpected output:", s)
	}
}
