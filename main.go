package main

import (
	adigo "adigo/internal"
	"fmt"
)

func main() {
	// Create a new graph
	graph := adigo.NewGraph()

	// Add 10 nodes to the graph
	for i := 0; i < 10; i++ {
		node := adigo.NewBox(fmt.Sprintf("Node %d", i))
		graph.AddNode(node)
	}

	// Connect nodes in the graph
	graph.Connect("Node 0", "Node 1", "Node 2")
	graph.Connect("Node 1", "Node 3", "Node 4")
	graph.Connect("Node 2", "Node 5", "Node 6")
	graph.Connect("Node 3", "Node 7", "Node 8")
	graph.Connect("Node 4", "Node 9")

	// Get the start and end nodes for BFS
	startNode, _ := graph.GetByLabel("Node 0")
	endNode, _ := graph.GetByLabel("Node 9")

	// Run BFS
	result := graph.BFS(startNode, endNode)

	// Print the result
	if result {
		fmt.Println("Node 9 is reachable from Node 0")
	} else {
		fmt.Println("Node 9 is not reachable from Node 0")
	}
}
