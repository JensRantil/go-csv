package csv_test

import (
	"bytes"
	"fmt"
	"github.com/JensRantil/go-csv"
)

func Example_readingWriting() {
	buf := bytes.Buffer{}

	writer := csv.NewWriter(&buf)
	writer.Write([]string{"Hello", "World", "!"})
	writer.Flush()

	reader := csv.NewReader(&buf)
	columns, err := reader.Read()
	if err != nil {
		panic(err)
	}

	for _, s := range columns {
		fmt.Println(s)
	}

	// Output:
	// Hello
	// World
	// !
}
