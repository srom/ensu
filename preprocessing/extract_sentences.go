package main

import (
	"encoding/csv"
	"fmt"
	"html"
	"os"

	"github.com/srom/xmlstream"
)

type Speech struct {
	SpeechID    string   `xml:"id,attr"`
	Paragraphs  []string `xml:"p"`
	NoSpeaker   bool     `xml:"nospeaker,attr"`
	SpeakerName string   `xml:"speakername,attr"`
	SpeakerID   string   `xml:"speakerid,attr"`
	URL         string   `xml:"url,attr"`
	Error       string   `xml:"error,attr"`
}

func (Speech) TagName() string {
	return "speech"
}

type TagHandler struct {
	count int
	W     *csv.Writer
}

// TagHandler implements xmlStream.Handler
func (th *TagHandler) HandleTag(tag interface{}) {
	switch el := tag.(type) {
	case *Speech:
		(*th).count++
		speech := *el

		if !speech.NoSpeaker && speech.Error == "" {
			// Concatenate the paragraphs together.
			text := ""
			for idx, t := range speech.Paragraphs {
				text += t
				if idx != len(speech.Paragraphs)-1 {
					text += "\\n"
				}
			}

			text = html.UnescapeString(text)

			(*th).W.Write([]string{
				speech.SpeechID,    // Speech ID
				speech.SpeakerID,   // Speaker ID
				speech.SpeakerName, // Speaker name
				speech.URL,         // Source
				text,               // Textual data
			})
			if (*th).count%50 == 0 {
				// Flush from time to time.
				(*th).W.Flush()
			}
		}

	default:
		panic("Can't handle this tag.")
	}
}

func main() {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush() // Don't forget to flush at the end.

	// Parse Person file.
	handler := TagHandler{W: w}
	if err := xmlstream.Parse(os.Stdin, &handler, 0, Speech{}); err != nil {
		panic(fmt.Sprint("Error parsing the file:", err))
	}
}
