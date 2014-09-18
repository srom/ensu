package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/google/cayley/quad/nquads"
)

type sortedMap struct {
	m map[string]int
	s []string
}

func (sm *sortedMap) Len() int {
	return len(sm.m)
}

func (sm *sortedMap) Less(i, j int) bool {
	return sm.m[sm.s[i]] > sm.m[sm.s[j]]
}

func (sm *sortedMap) Swap(i, j int) {
	sm.s[i], sm.s[j] = sm.s[j], sm.s[i]
}

func sortMap(m map[string]int) []string {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]string, len(m))
	i := 0
	for key, _ := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

func main() {
	max := 2000
	pol := false
	pattern := "<http://rdf.basekb.com/ns/tv."

	// Reading types.nt.gz stream from stdin.
	gzipReader, err := gzip.NewReader(os.Stdin)
	if err != nil {
		panic(err)
	}

	// Write unique entities to stdout.
	entityWriter := bufio.NewWriter(os.Stdout)

	defer func() {
		// Cleanup.
		gzipReader.Close()
		entityWriter.Flush()
	}()

	politicians := make(map[string]struct{})
	f, err := os.Open("map_id_freebase_id.csv")
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		var placeholder struct{}
		politicians["<"+record[1]+">"] = placeholder
	}
	f.Close()

	// Keep track of seen entities.
	seenEntities := make(map[string]int)

	scanner := bufio.NewScanner(gzipReader)
	for scanner.Scan() {
		t := scanner.Text() // Read line.
		quad, err := nquads.Parse(t)
		if err != nil {
			panic(err)
		}

		if strings.Contains(quad.Object, pattern) {
			if v, seen := seenEntities[quad.Subject]; !seen {
				// Add entity to set.
				seenEntities[quad.Subject] = 1

				if _, ok := politicians[quad.Subject]; ok && pol {
					// Write entity.
					fmt.Fprintf(entityWriter, "%s,%s\n", quad.Object, quad.Subject)
				}
				//fmt.Fprintf(entityWriter, "%s\n", quad.Subject)
			} else {
				if _, ok := politicians[quad.Subject]; !ok {
					// Increment.
					seenEntities[quad.Subject] = v + 1
				}
				// // Increment.
				// seenEntities[quad.Subject] = v + 1
			}
		}
	}

	keys := sortMap(seenEntities)
	for idx, key := range keys {
		if idx+1 > max {
			break
		}
		// Write most common entities.
		fmt.Fprintf(entityWriter, "%s\n", key)
	}
}
