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

	// Writing RDF into labels.nt.gz
	rdfFile, err := os.OpenFile("labels.nt.gz", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		rdfFile, err = os.Create("labels.nt.gz")
		if err != nil {
			panic(err)
		}
	}
	rdfWriter := gzip.NewWriter(rdfFile)

	// Writing entities into stdout.
	entityWriter := bufio.NewWriter(os.Stdout)

	var writeCount int64

	defer func() {
		// Cleanup
		gzipReader.Close()
		rdfWriter.Flush()
		rdfWriter.Close()
		rdfFile.Close()
		entityWriter.Flush()
	}()

	reAlias := regexp.MustCompile(`^"(.+)"@en$`)
	reEntity := regexp.MustCompile(`^.+ns/m\..+>$`)

	// put aliases into memory
	aliases := make(map[string]struct{})
	f, err := os.Open("aliases.txt")
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
		alias := strings.ToLower(s.Text()) // Read alias.
		// Set in go:
		// programmers.stackexchange.com/questions/177428/sets-data-structure-in-golang
		var placeholder struct{}
		aliases[alias] = placeholder
	}
	f.Close()

	// Keep track of seen entities.
	seenEntity := make(map[string]struct{})

	// Reading RDF lines (N-triples).
	scanner := bufio.NewScanner(gzipReader)
	for scanner.Scan() {
		t := scanner.Text() // Read line.
		quad, err := nquads.Parse(t)
		if err != nil {
			panic(err)
		}

		if reAlias.MatchString(quad.Object) && reEntity.MatchString(quad.Subject) {
			// Object, in english, refering to an entity.
			obj := reAlias.FindStringSubmatch(quad.Object)[1]

			if _, ok := aliases[strings.ToLower(obj)]; ok {
				// It's a match! Save alias, rdf row and entity.
				writeCount++
				if _, seen := seenEntity[quad.Subject]; !seen {
					// Add entity to set.
					var placeholder struct{}
					seenEntity[quad.Subject] = placeholder
					// Write entity.
					fmt.Fprintf(entityWriter, "%s\n", quad.Subject)
				}
				fmt.Fprintln(rdfWriter, t)

				if writeCount%50 == 0 {
					entityWriter.Flush()
					rdfWriter.Flush()
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
