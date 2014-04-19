// A CSV implementation inspired by Python's CSV module. Supports custom CSV
// formats. Currently only writing CSV files is supported.
package csv;

import (
  "bufio"
  "io"
  "strings"
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

const (
  QuoteDefault = iota  // See DefaultQuoting.
  QuoteAll = iota
  QuoteMinimal = iota
  QuoteNonNumeric = iota
  QuoteNone = iota
)

const (
  DoubleQuoteDefault = iota  // See DefaultDoubleQuote.
  DoDoubleQuote = iota
  NoDoubleQuote = iota
)

// Default dialect
const (
  DefaultDelimiter = " "
  DefaultQuoting = QuoteMinimal
  DefaultDoubleQuote = DoDoubleQuote
  DefaultEscapeChar = '\\'
  DefaultQuoteChar = '"'
  DefaultLineTerminator = "\n"
)

type Dialect struct {
  Delimiter string
  Quoting int
  DoubleQuote int
  EscapeChar rune
  QuoteChar rune
  LineTerminator string
}

func (wo *Dialect) setDefaults() {
  if wo.Delimiter == "" {
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

type Writer struct {
  opts Dialect
  w *bufio.Writer
}

// Create a writer that adheres to the Golang CSV writer.
func NewWriter(w io.Writer) Writer {
  opts := Dialect{}
  opts.setDefaults()
  return Writer{
    opts: opts,
    w: bufio.NewWriter(w),
  }
}

// Create a custom CSV writer.
func NewDialectWriter(w io.Writer, opts Dialect) Writer {
  opts.setDefaults()
  return Writer{
    opts: opts,
    w: bufio.NewWriter(w),
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
  return w.writeString(w.opts.Delimiter)
}

func isDigit(s rune) bool {
  switch s {
  case '0':
    return true
  case '1':
    return true
  case '2':
    return true
  case '3':
    return true
  case '4':
    return true
  case '5':
    return true
  case '6':
    return true
  case '7':
    return true
  case '8':
    return true
  case '9':
    return true
  default:
    return false
  }
}

func isNumeric(s string) bool {
  if len(s) == 0 {
    return false
  }
  for _, r := range s {
    if !isDigit(r) {
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
    return strings.Contains(field, w.opts.LineTerminator) || strings.Contains(field, w.opts.Delimiter) || strings.ContainsRune(field, w.opts.QuoteChar)
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
