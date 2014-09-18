package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"regexp"

	"github.com/google/cayley/quad/nquads"
)

func main() {
	// Reading labels.nt.gz stream from stdin.
	gzipReader, err := gzip.NewReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	// Writing RDF into labels_final.nt.gz
	rdfFile, err := os.OpenFile("LABELS_2.nt.gz", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		rdfFile, err = os.Create("LABELS_2.nt.gz")
		if err != nil {
			panic(err)
		}
	}
	rdfWriter := gzip.NewWriter(rdfFile)

	// Write unique aliases to stdout.
	aliasWriter := bufio.NewWriter(os.Stdout)

	defer func() {
		// Cleanup.
		gzipReader.Close()
		aliasWriter.Flush()
		rdfWriter.Flush()
		rdfWriter.Close()
		rdfFile.Close()
	}()

	// Keep track of seen aliases.
	seenAliases := make(map[string]struct{})

	reAlias := regexp.MustCompile(`^"(.+)"@en$`)
	rePredicate := regexp.MustCompile(
		`^(<http://www.w3.org/2000/01/rdf-schema#label>|` +
			`<http://rdf.basekb.com/ns/common.topic.alias>)$`)

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

	scanner := bufio.NewScanner(gzipReader)
	for scanner.Scan() {
		t := scanner.Text() // Read line.
		quad, err := nquads.Parse(t)
		if err != nil {
			panic(err)
		}

		if _, ok := entities[quad.Subject]; ok {
			// Entity found.
			if reAlias.MatchString(quad.Object) && rePredicate.MatchString(quad.Predicate) {
				obj := reAlias.FindStringSubmatch(quad.Object)[1]

				fmt.Fprintln(rdfWriter, t)

				if _, seen := seenAliases[obj]; !seen {
					var placeholder struct{}
					seenAliases[obj] = placeholder
					fmt.Fprintf(aliasWriter, "%s\n", obj)
				}
			}
		}
	}
}
