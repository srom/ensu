package graph

import "testing"

func TestGraph(t *testing.T) {
	/*
		Let's consider the following graph:

			                      +--+
		               +--1.0-----+N4|
		               |          ++-+
		               |           |
		               |           |
		             +-++          |
		  +----------+N1+----1.0----             +--+
		  |          +-++            +----1.0----+N5|
		  |          | |             |           ++-+
		 0.5         | |             |            |
		  |          | +----1.0----+-++-----1.0---+
		  |          |             |N3|
		  |          +----1.0------++-+----+
		+-++                        |      |          +--+
		|N0|                        |     1.0         |N7|
		+-++                        |      |          +--+
		  |                        0.25    |
		  |                         |     ++-+
		  |            +--+         |     |N6|
		  +-----1.0----+N2+---------+     +--+
		               +--+

		(Made with http://asciiflow.com/)
	*/

	// Create Nodes.
	nodes := []*Node{
		NewNode("N0"),
		NewNode("N1"),
		NewNode("N2"),
		NewNode("N3"),
		NewNode("N4"),
		NewNode("N5"),
		NewNode("N6"),
		NewNode("N7"),
	}

	// Create Edges.
	e01 := Edge{nodes[0], nodes[1], float32(0.5)}
	e02 := Edge{nodes[0], nodes[2], float32(1)}
	e13 := Edge{nodes[1], nodes[3], float32(1)}
	e13b := Edge{nodes[1], nodes[3], float32(1)}
	e14 := Edge{nodes[1], nodes[4], float32(1)}
	e14b := Edge{nodes[1], nodes[4], float32(1)}
	e23 := Edge{nodes[2], nodes[3], float32(0.25)}
	e35 := Edge{nodes[3], nodes[5], float32(1)}
	e35b := Edge{nodes[3], nodes[5], float32(1)}
	e36 := Edge{nodes[3], nodes[6], float32(1)}
	e45 := Edge{nodes[4], nodes[5], float32(1)}

	// Append edges to nodes.
	(*nodes[0]).Edges = append((*nodes[0]).Edges, e01, e02)
	(*nodes[1]).Edges = append((*nodes[1]).Edges, e01, e13, e13b, e14, e14b)
	(*nodes[2]).Edges = append((*nodes[2]).Edges, e23, e02)
	(*nodes[3]).Edges = append((*nodes[3]).Edges, e35, e35b, e36, e13, e13b, e23)
	(*nodes[4]).Edges = append((*nodes[4]).Edges, e14, e14b, e45)
	(*nodes[5]).Edges = append((*nodes[5]).Edges, e35, e35b, e45)
	(*nodes[6]).Edges = append((*nodes[6]).Edges, e36)

	// Calculate distance from a node to every other nodes.
	for _, node := range nodes {
		Dijkstra(node)
	}

	//
	// Test ShortestPaths function.
	//
	type spTestCase struct {
		From      *Node
		To        *Node
		Middleman *Node
		ExpectedG int
		ExpectedX int
	}

	// Test a few cases.
	testCases := []spTestCase{
		spTestCase{nodes[0], nodes[3], nodes[2], 1, 1},
		spTestCase{nodes[0], nodes[6], nodes[3], 1, 1},
		spTestCase{nodes[0], nodes[7], nodes[5], 0, 0},
		spTestCase{nodes[7], nodes[5], nodes[3], 0, 0},
		spTestCase{nodes[0], nodes[4], nodes[1], 2, 2},
		spTestCase{nodes[4], nodes[3], nodes[5], 6, 2},
		spTestCase{nodes[1], nodes[1], nodes[5], 1, 0},
		spTestCase{nodes[1], nodes[1], nodes[1], 1, 0},
	}

	for idx, test := range testCases {
		g, x := ShortestPaths(test.From, test.To, test.Middleman)
		expectedG := test.ExpectedG
		expectedX := test.ExpectedX
		if g != expectedG || x != expectedX {
			t.Errorf(
				"Test %d: Unexpected values for g and x:\n"+
					"Expected: g=%d, x=%d\nActual:   g=%d, x=%d",
				idx, expectedG, expectedX, g, x)
		}
	}

	//
	// Test BetweennessCentrality function.
	//
	type bcTestCase struct {
		Candidate     *Node
		ExpectedValue float32
	}

	cases := []bcTestCase{
		bcTestCase{nodes[3], float32(0.29761907)},
		bcTestCase{nodes[1], float32(0.15476191)},
		bcTestCase{nodes[7], float32(0)},
	}

	for idx, test := range cases {
		bc := BetweennessCentrality(test.Candidate, nodes, len(nodes))
		if bc != test.ExpectedValue {
			t.Errorf(
				"Test %d: Unexpected value:\n"+
					"Expected: bc=%.10f\nActual:   bc=%.10f",
				idx, test.ExpectedValue, bc)
		}
	}
}
