package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"

	"github.com/google/cayley/quad/nquads"
)

func main() {
	// Reading types.nt.gz stream from stdin.
	gzipReader, err := gzip.NewReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	// Writing RDF into labels_final.nt.gz
	rdfFile, err := os.OpenFile("TYPES_2.nt.gz", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		rdfFile, err = os.Create("TYPES_2.nt.gz")
		if err != nil {
			panic(err)
		}
	}
	rdfWriter := gzip.NewWriter(rdfFile)

	// Write unique entities to stdout.
	typeWriter := bufio.NewWriter(os.Stdout)

	defer func() {
		// Cleanup.
		gzipReader.Close()
		typeWriter.Flush()
		rdfWriter.Flush()
		rdfWriter.Close()
		rdfFile.Close()
	}()

	// put entities into a set in memory
	entities := make(map[string]struct{})
	f, err := os.Open("ENTITIES.txt")
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
		entity := s.Text() // Read entity.
		// Put entity in the set.
		var placeholder struct{}
		entities[entity] = placeholder
	}
	f.Close()

	// Keep track of seen entities.
	seenTypes := make(map[string]struct{})

	count := 0
	scanner := bufio.NewScanner(gzipReader)
	for scanner.Scan() {
		t := scanner.Text() // Read line.
		quad, err := nquads.Parse(t)
		if err != nil {
			panic(err)
		}

		count++
		if count%10000 == 0 {
			fmt.Fprintln(os.Stderr, count)
		}

		if _, ok := entities[quad.Subject]; ok {
			if _, seen := seenTypes[quad.Object]; !seen {
				// Add type to set.
				var placeholder struct{}
				seenTypes[quad.Object] = placeholder
				fmt.Fprintf(typeWriter, "%s\n", quad.Object)
			}
			fmt.Fprintln(rdfWriter, t)
		}
	}
}
