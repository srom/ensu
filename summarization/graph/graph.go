package graph

// A Node is defined by an ID, a list of edges
// and the distances to any other node.
type Node struct {
	ID       string
	Edges    []Edge
	Distance map[string]float32 // Distance (sum of weights) to a given node (by ID)
}

func NewNode(id string) *Node {
	return &Node{
		ID:       id,
		Edges:    make([]Edge, 0, 10),
		Distance: make(map[string]float32, 10),
	}
}

// Follow an edge from a starting node.
func NextNode(start *Node, edge Edge) *Node {
	if edge.A == start {
		return edge.B
	} else {
		return edge.A
	}
}

// An Edge link two nodes together with a weight.
type Edge struct {
	A      *Node
	B      *Node
	Weight float32
}

// Dijkstra updates the distances from the start node
// to *any* other node in the graph.
func Dijkstra(start *Node) {
	visited := make(map[*Node]struct{}, 10)
	(*start).Distance[(*start).ID] = float32(0)
	currentNode := start

	// Traverse all the graph and find the shortest paths.
	// End when all nodes have been marked as visited.
	for {
		// Check all the unvisited neighbours.
		for _, edge := range (*currentNode).Edges {
			nodePtr := NextNode(currentNode, edge)

			if _, ok := visited[nodePtr]; ok {
				continue // Node has been visited already.
			}

			// Update tentative.
			tentative := (*currentNode).Distance[(*start).ID] + edge.Weight
			d, ok := (*nodePtr).Distance[(*start).ID]
			if !ok || tentative < d {
				(*nodePtr).Distance[(*start).ID] = tentative
			}
		}

		// Mark the current node as visited.
		var foo struct{} // Placeholder.
		visited[currentNode] = foo

		// Select the neighbour with the smallest
		// distance as the next node.
		nextIndex := -1
		min := float32(-1)
		for idx, edge := range (*currentNode).Edges {
			nodePtr := NextNode(currentNode, edge)

			if _, ok := visited[nodePtr]; ok {
				continue // Node has been visited already.
			}
			if min == float32(-1) {
				nextIndex = idx
			} else if (*nodePtr).Distance[(*start).ID] < min {
				nextIndex = idx
			}
		}

		if nextIndex == -1 {
			// No more nodes to visit.
			break
		}

		// Select next node to consider.
		e := (*currentNode).Edges[nextIndex]
		currentNode = NextNode(currentNode, e)
	}
}

// Returns the number of shortest paths between the start node and the end node
// and the number of shortest paths passing through the middleman.
func ShortestPaths(start, end, middleman *Node) (int, int) {
	a, b := make(chan int), make(chan int)
	go followShortestPaths(start, end, middleman, false, 0, 0, a, b)
	return <-a, <-b
}

func followShortestPaths(
	start, end, middleman *Node,
	gtm bool, distance float32, depth int, a, b chan int) {

	if start == end {
		nbsp := 0
		gtmv := 0 // Go Through Middleman Value.
		if depth == 0 || distance == 0 {
			nbsp = 1
			if gtm {
				gtmv = 1
			}
		}
		a <- nbsp
		b <- gtmv
		return // Stop recursion.
	}

	nsp := 0   // Number of Shortest Paths
	nsptm := 0 // Number of Shorest Paths going Through the Middleman

	isMiddleman := gtm
	if end == middleman {
		isMiddleman = true
	}

	if depth == 0 {
		distance = (*end).Distance[(*start).ID]
	}

	// Continue exploration.
	buffer := len((*end).Edges)
	aa, bb := make(chan int, buffer), make(chan int, buffer)
	routines := 0 // Number of routines launched.
	for _, edge := range (*end).Edges {
		next := NextNode(end, edge)

		dist := distance - edge.Weight

		if dist < 0 {
			continue
		}

		// Follow next shortest paths.
		go followShortestPaths(start, next, middleman,
			isMiddleman, dist, depth+1, aa, bb)

		routines++
	}

	// Increment shortest paths.
	for i := 0; i < routines; i++ {
		nsp += <-aa
		nsptm += <-bb
	}

	// Send values to channels
	a <- nsp
	b <- nsptm
}

// Betweenness centrality quantifies the number of times a candidate node acts as
// a bridge along the shortest path between two other nodes.
//
// https://en.wikipedia.org/wiki/Centrality#Betweenness_centrality
//
// The parameter nbOfNodes is the total number of nodes in the graph.
// The candidate node will be ignored if it is in the list of nodes.
func BetweennessCentrality(candidate *Node, nodes []*Node, nbOfNodes int) float32 {
	res := float32(0)
	visited := make(map[*Node]struct{}, 10)
	for _, start := range nodes {
		// Ignore node if we've seen it already,
		// or if it is the candidate node.
		if _, ok := visited[start]; start == candidate || ok {
			continue
		}
		for _, end := range nodes {
			// Ignore node if it is the candidate,
			// or the starting node,
			// or a node we've seen already.
			if _, seen := visited[end]; end == candidate ||
				end == start || seen {
				continue
			}

			// Calculate the number of shortest paths between start and end,
			// and the number of these paths going through the candidate.
			sp, sptm := ShortestPaths(start, end, candidate)

			if sptm > 0 && sp > 0 {
				// Update the result.
				res += float32(sptm) / float32(sp)
			}
		}
		// Mark the start node as visited.
		var placeholder struct{}
		visited[start] = placeholder
	}
	// Normalize the result by the total number of pair of nodes
	// that exist in the graph excluding the candidate node.
	res /= float32(nbOfNodes) * (float32(nbOfNodes-1) / 2)
	return res
}
