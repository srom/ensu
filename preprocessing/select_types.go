package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/cayley/quad/nquads"
)

func main() {
	// Reading .nt.gz stream from stdin.
	gzipReader, err := gzip.NewReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	// Writing RDF into types.nt.gz
	rdfFile, err := os.OpenFile("types_final_ter.nt.gz", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		rdfFile, err = os.Create("types_final_ter.nt.gz")
		if err != nil {
			panic(err)
		}
	}
	rdfWriter := gzip.NewWriter(rdfFile)

	// Write unique types to stdout.
	typeWriter := bufio.NewWriter(os.Stdout)

	defer func() {
		// Cleanup.
		gzipReader.Close()
		typeWriter.Flush()
		rdfWriter.Flush()
		rdfWriter.Close()
		rdfFile.Close()
	}()

	// topics to consider.
	topics := []string{
		"<http://rdf.basekb.com/ns/government.",
		"<http://rdf.basekb.com/ns/law.",
		"<http://rdf.basekb.com/ns/organization.",
		"<http://rdf.basekb.com/ns/education.",
		"<http://rdf.basekb.com/ns/business.",
		"<http://rdf.basekb.com/ns/people.",
		"<http://rdf.basekb.com/ns/religion.",
		"<http://rdf.basekb.com/ns/military.",
		"<http://rdf.basekb.com/ns/location.",
		"<http://rdf.basekb.com/ns/music.",
		"<http://rdf.basekb.com/ns/book.",
		"<http://rdf.basekb.com/ns/media_common.",
		"<http://rdf.basekb.com/ns/fictional_universe.",
		"<http://rdf.basekb.com/ns/sports.",
		"<http://rdf.basekb.com/ns/internet.",
		"<http://rdf.basekb.com/ns/film.",
		"<http://rdf.basekb.com/ns/tv.",
	}

	// put entities into a set in memory
	entities := make(map[string]struct{})
	f, err := os.Open("entities.txt")
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

	reEntity := regexp.MustCompile(`^.+ns/m\..+>$`)

	// Keep track of seen types.
	seenTypes := make(map[string]struct{})

	var writeCount int64
	var count int64

	// Reading RDF lines (N-triples).
	scanner := bufio.NewScanner(gzipReader)
	for scanner.Scan() {
		t := scanner.Text() // Read line.
		quad, err := nquads.Parse(t)
		if err != nil {
			panic(err)
		}

		count++
		fmt.Fprintln(os.Stderr, count)

		if reEntity.MatchString(quad.Subject) {
			if _, ok := entities[quad.Subject]; ok {
				// Entity found!
				valid := false
				for _, topicPattern := range topics {
					if strings.Contains(quad.Object, topicPattern) {
						valid = true
						break
					}
				}
				if !valid {
					continue
				}

				writeCount++

				// Append rdf row.
				fmt.Fprintln(rdfWriter, t)

				// Append new types to stdout.
				if _, seen := seenTypes[quad.Object]; !seen {
					var placeholder struct{}
					seenTypes[quad.Object] = placeholder
					fmt.Fprintln(typeWriter, quad.Object)
				}

				if writeCount%50 == 0 {
					typeWriter.Flush()
					rdfWriter.Flush()
				}
			}
		}
	}
}
