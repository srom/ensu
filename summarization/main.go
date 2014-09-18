package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"graph"
	"io/ioutil"
	"net/http"
)

var QUERY_ENDPOINT string = "http://127.0.0.1:64210/api/v1/query/gremlin"

type NodeID struct {
	ID string `json:"id"`
}

type GraphResponse struct {
	Result []NodeID `json:"result"`
	Error  string   `json:"error"`
}

func QueryGraph(query string) (GraphResponse, error) {
	q := []byte(query)
	resp, err := http.Post(
		QUERY_ENDPOINT,
		"text/javascript",
		bytes.NewBuffer(q),
	)
	if err != nil {
		return GraphResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GraphResponse{}, err
	}
	var response GraphResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return GraphResponse{}, err
	}
	return response, nil
}

type Summary struct {
	Sentence1 string
	Sentence2 string
	Length    int
}

func CandidateSummaries(chain []string) []Summary {
	csu := make([]Summary, 0, 10)

	// Find candidate sentences.
	cs1 := CandidateSentences(chain[0], chain[1])
	cs2 := CandidateSentences(chain[1], chain[2])
	cs3 := CandidateSentences(chain...)

	// Remove cs3 sentences from cs1 and cs2.
	set := make(map[string]struct{}, 10)
	for _, sentence := range cs3 {
		var placeholder struct{}
		set[sentence] = placeholder
	}
	cs1b := make([]string, 0, len(cs1))
	for _, sentence := range cs1 {
		if _, seen := set[sentence]; !seen {
			cs1b = append(cs1b, sentence)
		}
	}
	cs2b := make([]string, 0, len(cs2))
	for _, sentence := range cs2 {
		if _, seen := set[sentence]; !seen {
			cs2b = append(cs2b, sentence)
		}
	}

	// Combine cs1 and cs2.
	for _, s1 := range cs1b {
		for _, s2 := range cs2b {
			csu = append(csu, Summary{s1, s2, 2})
		}
	}

	// Make cs3 summaries.
	for _, sentence := range cs3 {
		csu = append(csu, Summary{sentence, "", 1})
	}

	return csu
}

func main() {
	// Entity chain.
	chain := []string{
		"http://rdf.basekb.com/ns/m.0266lb",
		"http://rdf.basekb.com/ns/m.0_6t_z8",
		"http://rdf.basekb.com/ns/m.04qdjv",
	}

	// Retrieve candidate summaries
	csu := CandidateSummaries(chain)

	scores := make(map[Summary]float32, 10)

	for _, summary := range csu {
		// Create the graph
		nodes := CreateGraph(summary)

		// Calculate distance from a node to every other nodes.
		for _, node := range nodes {
			graph.Dijkstra(node)
		}

		// Calculate score.
		score := float32(0)
		entityToNode := make(map[string]*graph.Node, 10)
		for _, node := range nodes {
			entityToNode[(*node).ID] = node
		}
		for _, entity := range chain {
			node := entityToNode[entity]
			score += graph.BetweennessCentrality(node, nodes, len(nodes))
		}

		scores[summary] = score
	}

	// Find the best summary.
	var bestSummary Summary
	bestScore := float32(-1)
	for key, value := range scores {
		if value > bestScore {
			bestScore = value
			bestSummary = key
		}
	}
	fmt.Println(bestSummary)
}
