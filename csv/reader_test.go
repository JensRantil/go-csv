// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"bytes"
	oldcsv "encoding/csv"
	"io"
	"reflect"
	"testing"
)

func TestReaderInterface(t *testing.T) {
	t.Parallel()

	var iface CsvReader
	iface = NewReader(new(bytes.Buffer))
	iface = NewDialectReader(new(bytes.Buffer), Dialect{})
	iface = oldcsv.NewReader(new(bytes.Buffer))

	// To get rid of compile-time warning that this variable is not used.
	iface.Read()
}

func TestUnReader(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a,b,c\n")
	r := NewUnReader(b)
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
	b.WriteString("a b c\n")
	r := NewReader(b)

	err := testReadingSingleLine(t, r, []string{"a", "b", "c"})
	if err != nil && err != io.EOF {
		t.Error("Unexpected error:", err)
	}
}

func TestReadingTwoLines(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a b c\nd e f\n")
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
	b.WriteString("a \"b\" c\n")
	r := NewReader(b)

	err := testReadingSingleLine(t, r, []string{"a", "b", "c"})
	if err != nil && err != io.EOF {
		t.Error("Unexpected error:", err)
	}
}

func TestReadAll(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("a \"b\" c\nd e \"f\"\n")
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
