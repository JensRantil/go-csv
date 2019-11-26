// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/JensRantil/go-csv/interfaces"
)

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

func testReaderQuick(t *testing.T, quoting QuoteMode) {
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

func TestEmptyLastField(t *testing.T) {
	in := `"Rob","Pike",
Ken,Thompson,ken
`
	r := csv.NewReader(strings.NewReader(in))

	var out [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
		}
		out = append(out, record)
	}

	expected := [][]string{
		{"Rob", "Pike", ""},
		{"Ken", "Thompson", "ken"},
	}

	if !reflect.DeepEqual(out, expected) {
		t.Errorf("Output differed from expected.\nout=%s\nexpected=%s", out, expected)
	}
}

// A reader that will source an infinitely repeating pattern of bytes.
type infiniteReader struct {
	RepeatingPattern []byte
	position         int
}

func (r *infiniteReader) Read(p []byte) (n int, err error) {
	j := 0
	for j < len(p) {
		nToCopy := min(len(p)-j, len(r.RepeatingPattern)-r.position)
		copy(p[j:(j+nToCopy)], r.RepeatingPattern[r.position:(r.position+nToCopy)])

		r.position += nToCopy
		r.position %= len(r.RepeatingPattern)

		j += nToCopy
	}
	return len(p), nil
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func TestInfiniteReader(t *testing.T) {
	testString := "this is a line\n"
	r := infiniteReader{RepeatingPattern: []byte(testString)}
	s := bufio.NewScanner(&r)
	for i := 0; i < 100000; i++ {
		if !s.Scan() {
			t.Fatal("Scan() returned false from infinite stream. Iteration:", i)
		}
		if ts, expected := s.Text(), testString[0:len(testString)-1]; expected != ts {
			t.Fatal("Incorrect string:", []byte(ts), "Expected:", []byte(expected))
		}
	}
	if err := s.Err(); err != nil {
		t.Error("unexpected error:", err)
	}
}

const testString = "peter,sweden\n"

func BenchmarkReadingCSV(b *testing.B) {
	r := infiniteReader{RepeatingPattern: []byte(testString)}
	csvr := NewReader(&r)
	benchmark(b, csvr)
}

func BenchmarkGolangCSV(b *testing.B) {
	r := infiniteReader{RepeatingPattern: []byte(testString)}
	csvr := csv.NewReader(&r)
	benchmark(b, csvr)
}

func benchmark(b *testing.B, csvr interfaces.Reader) {
	for i := 0; i < b.N; i++ {
		r, err := csvr.Read()
		if err != nil {
			b.Fatal("Unexpected error:", err)
		}
		if len(r) != 2 || r[0] != "peter" || r[1] != "sweden" {
			b.Fatalf("Unexpected row of len=%d: %s", len(r), r)
		}
	}
}

func TestReadingWithComments(t *testing.T) {
	t.Parallel()

	b := new(bytes.Buffer)
	b.WriteString("#-,-,-\n   #aa\na,b,c\n	#aa#aaaa\nd,e,f\n")
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
