// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"testing/quick"
)

func TestUnReader(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a,b,c\n")
	r := newUnreader(b)
	if ru, _, _ := r.ReadRune(); ru != 'a' {
		t.Error("Unexpected char:", ru, "Expected:", 'a')
	}
	if ok, _ := r.NextIsString(",b,c"); !ok {
		t.Error("Unexpected next string.")
	}
	r.UnreadRune('d')
	if ok, _ := r.NextIsString("d,b,c"); !ok {
		t.Error("Unreading failed.")
	}
	if ok, _ := r.NextIsString("b,c"); ok {
		t.Error("Unexpected next string.")
	}
	if ru, _, _ := r.ReadRune(); ru != 'd' {
		t.Error("Unexpected char:", ru, "Expected:", 'd')
	}
}

func testReadingSingleLine(t *testing.T, r *Reader, expected []string) error {
	record, err := r.Read()
	if c := len(record); c != len(expected) {
		t.Fatal("Wrong number of fields:", c, "Expected:", len(expected))
	}
	if !reflect.DeepEqual(record, expected) {
		t.Error("Incorrect records.")
		t.Error(record)
		t.Error(expected)
	}
	return err
}

func TestReadingSingleFieldLine(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a\n")
	r := NewReader(b)

	err := testReadingSingleLine(t, r, []string{"a"})
	if err != nil && err != io.EOF {
		t.Error("Unexpected error:", err)
	}
}

func TestReadingSingleLine(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a,b,c\n")
	r := NewReader(b)

	err := testReadingSingleLine(t, r, []string{"a", "b", "c"})
	if err != nil && err != io.EOF {
		t.Error("Unexpected error:", err)
	}
}

func TestReadingTwoLines(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a,b,c\nd,e,f\n")
	r := NewReader(b)
	err := testReadingSingleLine(t, r, []string{"a", "b", "c"})
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	err = testReadingSingleLine(t, r, []string{"d", "e", "f"})
	if err != nil && err != io.EOF {
		t.Error("Expected EOF, but got:", err)
	}
}

func TestReadingBasicCommaDelimitedFile(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("\"b\"\n")
	r := NewReader(b)

	err := testReadingSingleLine(t, r, []string{"b"})
	if err != nil && err != io.EOF {
		t.Error("Unexpected error:", err)
	}
}

func TestReadingCommaDelimitedFile(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a,\"b\",c\n")
	r := NewReader(b)

	err := testReadingSingleLine(t, r, []string{"a", "b", "c"})
	if err != nil && err != io.EOF {
		t.Error("Unexpected error:", err)
	}
}

func TestReadAll(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a,\"b\",c\nd,e,\"f\"\n")
	r := NewReader(b)

	data, err := r.ReadAll()
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	equals := reflect.DeepEqual(data, [][]string{
		{
			"a",
			"b",
			"c",
		},
		{
			"d",
			"e",
			"f",
		},
	})
	if !equals {
		t.Error("Unexpected output:", data)
	}
}

func testReaderQuick(t *testing.T, quoting int) {
	f := func(records [][]string, doubleQuote bool, escapeChar, del, quoteChar rune, lt string) bool {
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
		b := new(bytes.Buffer)
		w := NewDialectWriter(b, dialect)
		w.WriteAll(records)

		r := NewDialectReader(b, dialect)
		data, err := r.ReadAll()
		if err != nil {
			t.Error("Error when reading CSV:", err)
			return false
		}

		equal := reflect.DeepEqual(records, data)
		if !equal {
			t.Error("Not equal:", records, data)
		}
		return equal
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Test writing to and then reading from using various CSV dialects.
func TestReaderQuick(t *testing.T) {
	t.Parallel()

	testWriterQuick(t, QuoteAll)
	testWriterQuick(t, QuoteMinimal)
	testWriterQuick(t, QuoteNonNumeric)
}
