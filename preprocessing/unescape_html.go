package main

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"github.com/kennygrant/sanitize"
)

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush() // Don't forget to flush at the end.

	re := regexp.MustCompile(`\(([^()]*)\)`)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text() // Read line.

		// XML from theyworkforyou.com are not valid.
		// Unescape some weird characters.
		t = html.UnescapeString(t)

		if strings.Contains(t, "<p ") {
			// In paragraphs, remove all trace of HTML.
			t = sanitize.HTML(t)

			// Re-escape a few characters which shouldn't
			// be escpaed in an XML file.
			t = html.EscapeString(t)

			// Remove parentheses.
			t = re.ReplaceAllString(t, "$1")

			// Trim spaces, re-add <p> tags and output the line.
			fmt.Fprintf(w, "<p>%s</p>\n", strings.TrimSpace(t))

			continue // Continue to next element
		}

		// Not a pragraph. Output the line.
		fmt.Fprintf(w, "%s\n", t)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
