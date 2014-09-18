package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/google/cayley/quad"
)

func main() {
	// Writer of link between entities and sentences.
	rdfWriter := gzip.NewWriter(os.Stdout)

	defer func() {
		rdfWriter.Flush()
		rdfWriter.Close()
	}()

	// Read ASSIGNMENT.csv
	r := csv.NewReader(os.Stdin)
	line := 0
	for {
		record, err := r.Read()
		line++
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error line %d: %v\n", line, err)
			continue
		}

		if line%10000 == 0 {
			fmt.Fprintf(os.Stderr, "Line %d\n", line)
		}

		// Link to entity.
		quad := quad.Quad{
			Subject:   "<" + record[1] + ">",
			Predicate: "<http://rdf.romainstrock.com/ns/sentence.is_about>",
			Object:    "<" + record[3] + ">",
		}
		fmt.Fprintln(rdfWriter, quad.NQuad())
	}
}
