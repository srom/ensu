package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/google/cayley/quad"
)

func main() {
	// Writer of link between entities and sentences.
	rdfWriter := gzip.NewWriter(os.Stdout)

	defer func() {
		rdfWriter.Flush()
		rdfWriter.Close()
	}()

	// Read SENTENCES.csv
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

		// Speaker.
		quad_1 := quad.Quad{
			Subject:   "<" + record[0] + ">",
			Predicate: "<http://rdf.romainstrock.com/ns/sentence.speaker>",
			Object:    "<" + record[3] + ">",
		}
		fmt.Fprintln(rdfWriter, quad_1.NQuad())

		// Sentence text.
		text := record[6]
		quad_2 := quad.Quad{
			Subject:   "<" + record[0] + ">",
			Predicate: "<http://rdf.romainstrock.com/ns/sentence.text>",
			Object:    strconv.Quote(text),
		}
		fmt.Fprintln(rdfWriter, quad_2.NQuad())

		// Source URL.
		quad_3 := quad.Quad{
			Subject:   "<" + record[0] + ">",
			Predicate: "<http://rdf.romainstrock.com/ns/sentence.source_url>",
			Object:    "<" + record[5] + ">",
		}
		fmt.Fprintln(rdfWriter, quad_3.NQuad())
	}
}
