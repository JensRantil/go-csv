package dialect_test

import (
	"flag"
	csv "github.com/JensRantil/go-csv"
	"github.com/JensRantil/go-csv/dialect"
	"os"
)

func Example_flag() {
	builder := dialect.FromCommandLine()

	flag.Parse()

	dialect, err := builder.Dialect()
	if err != nil {
		panic(err)
	}

	reader := csv.NewDialectWriter(os.Stdout, *dialect)
	reader.Write([]string{"Hello", "World"})
	reader.Flush()

	// Output:
	// Hello	World
}

func Example_flagSet() {
	fset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	builder := dialect.FromFlagSet(fset)

	fset.Parse([]string{})

	dialect, err := builder.Dialect()
	if err != nil {
		panic(err)
	}

	reader := csv.NewDialectWriter(os.Stdout, *dialect)
	reader.Write([]string{"Hello", "World"})
	reader.Flush()

	// Output:
	// Hello	World
}
