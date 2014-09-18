package main

import (
	"fmt"
	"os"

	"github.com/srom/xmlstream"
)

// Member is an object which can be unmarshalled from all-members.xml.
//
// Example:
// <member
//     id="uk.org.publicwhip/member/1656"
//     house="commons"
//     title="" firstname="Emily" lastname="Thornberry"
//     constituency="Islington South &amp; Finsbury" party="Lab"
//     fromdate="2005-05-05" todate="9999-12-31"
//     fromwhy="general_election" towhy="still_in_office"
// />
type Member struct {
	Id           string `xml:"id,attr"`
	Title        string `xml:"title,attr"`
	FirstName    string `xml:"firstname,attr"`
	LastName     string `xml:"lastname,attr"`
	House        string `xml:"house,attr"`
	Party        string `xml:"party,attr"`
	Constituency string `xml:"constituency,attr"`
	FromDate     string `xml:"fromdate,attr"`
	ToDate       string `xml:"todate,attr"`
	FromWhy      string `xml:"fromwhy,attr"`
	ToWhy        string `xml:"towhy,attr"`
}

func (Member) TagName() string {
	return "member"
}

// Person is an object which can be unmarshalled from people.xml.
//
// See Person's godoc for an example.
type Office struct {
	Id      string `xml:"id,attr"`
	Current string `xml:"current,attr"`
}

// Person is an object which can be unmarshalled from people.xml.
//
// Example:
// <person id="uk.org.publicwhip/person/10002" latestname="Gerry Adams">
//     <office id="uk.org.publicwhip/member/1403"/>
//     <office id="uk.org.publicwhip/member/40071" current="yes"/>
//	</person>
type Person struct {
	Id         string   `xml:"id,attr"`
	LatestName string   `xml:"latestname,attr"`
	Offices    []Office `xml:"office"`
}

func (Person) TagName() string {
	return "person"
}

// Lord is an object which can be unmarshalled from peers.xml.
//
// Example:
// <lord
// 	id="uk.org.publicwhip/lord/100001"
// 	house="lords"
// 	forenames="Morys"
// 	title="Lord" lordname="Aberdare" lordofname=""
// 	peeragetype="HD" affiliation="Con"
// 	fromdate="1957" todate="2005-01-23" towhy="died"
// />
type Lord struct {
	Id          string `xml:"id,attr"`
	House       string `xml:"house,attr"`
	County      string `xml:"county"`
	PeerageType string `xml:"peeragetype,attr"`
	FromDate    string `xml:"fromdate,attr"`
	ToDate      string `xml:"todate,attr"`
	Party       string `xml:"affiliation,attr"`
}

func (Lord) TagName() string {
	return "lord"
}

// Minister is an object which can be unmarshalled from ministers.xml.
//
// Example:
// <moffice
//   id="uk.org.publicwhip/moffice/10000"
//   name="David Cameron" matchid="uk.org.publicwhip/member/40665"
//   dept="" position="Prime Minister" fromdate="2010-05-11"
//   todate="9999-12-31" source="manual"
//  />
type Minister struct {
	Id       string `xml:"id,attr"`
	MathId   string `xml:"matchid,attr"`
	Name     string `xml:"name,attr"`
	FromDate string `xml:"fromdate,attr"`
	ToDate   string `xml:"todate,attr"`
	Position string `xml:position`
	Dept     string `xml:"dept"`
}

func (Minister) TagName() string {
	return "moffice"
}

type TagHandler struct {
}

// TagHandler implements xmlStream.Handler
func (th *TagHandler) HandleTag(tag interface{}) {
	switch el := tag.(type) {
	case *Person:
		person := *el

		fmt.Printf("Person: %#v\n", person)

		// // Find entity ID on freebase.
		// //
		// entityID := "<http://rdf.freebase.com/ns/m.11vjz1y>" // placeholder

		// quads := make([]quad.Quad, 0, 10) // List of quads

		// quads = append(quads, quad.Quad{
		// 	Subject:   fmt.sprintf("<%s>", entityID),
		// 	Predicate: "</publicwhip/person/id>",
		// 	Object:    person.Id,
		// })

		// // Map the current office id of this person to the person object.
		// for _, office := range person.Offices {
		// 	quads = append(quads, quad.Quad{
		// 		Subject:   fmt.sprintf("<%s>", entityID),
		// 		Predicate: "</publicwhip/office/id>",
		// 		Object:    office.Id,
		// 	})
		// }

		// Write quads to file.
		//

		// case *Member:
		// 	member := *el
		// 	if person, ok := (*th).currentIdToPersonId[member.Id]; ok {
		// 		fmt.Println(person.Id, member.Id)
		// 	}

		// case *Minister:
		// 	minister := *el
		// 	if person, ok := (*th).currentIdToPersonId[minister.Id]; ok {
		// 		fmt.Println(person.Id, minister.Id)
		// 	}

		// case *Lord:
		// 	lord := *el
		// 	if person, ok := (*th).currentIdToPersonId[lord.Id]; ok {
		// 		fmt.Println(person.Id, lord.Id)
		// 	}

		// default:
		// 	panic("This handler expects a tag of type *Member, *Person, *Minister or *Lord.")
	}
}

func main() {
	// Open Person file.
	personFile, err := os.Open("people.xml")
	if err != nil {
		panic(fmt.Sprint("Error opening the file:", err))
	}
	defer personFile.Close()

	// Parse Person file.
	handler := TagHandler{}
	err = xmlstream.Parse(personFile, &handler, 0, Person{})
	if err != nil {
		panic(fmt.Sprint("Error parsing the file:", err))
	}
}
