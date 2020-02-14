// An ADI Graph describes a mathematical graph using sets of adjacency descriptive integers.
// Adjacency descriptive integers are essentially bitsets that have an "on" bit (1) if there is
// a connection between two nodes and an "off" bit (0) if there isn't. It can be thought of as a
// compressed adjacency matrix.
//
// The benefit of implementing a graph this way is that you can easily perform parallelized operations
// on the nodes and edges, while avoiding the common overhead of having a large, sparsely-connected matrix.
// As a simple example, checking whether or not two nodes are connected is an O(1) operation.
//
// ADI Graph
//
// This is the wrapper struct that manages the nodes of the graph. The graph is the host of CRUD operations on nodes.
//
// Currently, lookup by index is O(1), but lookup by label is O(n). Adding a map to this wrapper,
// from labels to indices, will allow O(1) lookups by label.
//
// ADI Node
//
// The ADI Node is the core of the implementation and provides methods to manage its own contents, labels, and edges.
//
// A general-purpose node is provided through Box, but users of the package can implement their own nodes
// for specialized use-cases by simply implementing the interface.
package main

import (
	"errors"
	"fmt"
	"sync"
)

var (
	errGraphBoundsMismatch = errors.New("the number of nodes does not match the bounds of the ADIs")
	errLabelNotFound       = errors.New("label not found")
	errDeleted             = errors.New("node has been deleted")
)

// ADIGraph wraps a list of ADINodes, adding columns to itself as needed
// when the number of nodes outgrows the word size.
type ADIGraph struct {
	growthFactor int
	wordSize     int
	nodes        []ADINode
	labels       map[string]int
}

// NewGraph constructs a new ADIGraph with the default word size of 8 bits.
func NewGraph() ADIGraph {
	graph := ADIGraph{}
	graph.wordSize = 8
	graph.labels = make(map[string]int)
	return graph
}

// Locator contains the column and offset used to identify a node in the graph.
type Locator struct {
	column int
	offset byte
}

// Neighbors accepts a node and returns the neighbors of the node in the graph.
func (g ADIGraph) Neighbors(n ADINode) []ADINode {
	gn := g.nodes
	results := make([]ADINode, len(gn))
	adichan := make(chan ADINode, len(gn))
	adis := n.Edges()
	var wg sync.WaitGroup

	for i := 0; i < len(adis); i++ {
		for j := 0; j < g.wordSize; j++ {
			wg.Add(1)
			go g.lookup(i, adis[i], j, adichan, &wg)
		}
	}
	wg.Wait()
	close(adichan)

	idx := 0
	for node := range adichan {
		results[idx] = node
		idx++
	}

	return results[:idx]
}

func (g ADIGraph) lookup(col int, adi byte, offset int, ch chan ADINode, wg *sync.WaitGroup) {
	checkbit := 1 << byte(offset)
	if adi&byte(checkbit) != 0 {
		index := (col * g.wordSize) + offset
		node, err := g.GetByIndex(index)
		if err == nil {
			ch <- node
		}
	}
	wg.Done()
}

// Connect takes a node in the graph and connects it to any number of other nodes.
func (g ADIGraph) Connect(label string, neighbors ...string) error {
	item, err := g.GetByLabel(label)
	if err != nil {
		return errLabelNotFound
	}
	locations := []Locator{}
	for _, v := range neighbors {
		loc, err := g.GetLocatorsByLabel(v)
		if err == nil {
			locations = append(locations, loc)
		}
	}
	for _, v := range locations {
		item.AddEdges(v)
	}
	return nil
}

// Size returns the length of the array of nodes in the receiving ADIGraph.
func (g ADIGraph) Size() int {
	return len(g.nodes)
}

// AddNode accepts an ADINode and adds it to the receiving ADIGraph. It returns a nil error
// unless the node fails to be added.
//
// If adding the node causes the size of the graph to be larger than the graph's word size,
// the graph will grow a column to locate the new node within the existing nodes' lists of ADIs.
func (g *ADIGraph) AddNode(n ADINode) error {
	if len(g.nodes) > (g.wordSize-1) && (len(g.nodes)%g.wordSize) == 0 {
		g.labels[n.Label()] = len(g.nodes)
		g.nodes = append(g.nodes, n)
		g.Grow()
		return nil
	}

	// Map the node's index to its label before appending to the node list.
	g.labels[n.Label()] = len(g.nodes)
	g.nodes = append(g.nodes, n)
	return nil
}

// Grow adds a new column to the ADI list of every node in the graph. This way, the graph is able to
// address more nodes than the word size would normally allow.
func (g *ADIGraph) Grow() {
	for _, node := range g.nodes {
		node.AddColumn()
	}
	g.growthFactor++
}

// GetByIndex accepts an integer index and returns the node at that index within the graph.
// Might be useful to return a pointer to the node.
func (g ADIGraph) GetByIndex(index int) (ADINode, error) {
	if index == -1 {
		return nil, errDeleted
	}
	return g.nodes[index], nil
}

// GetByLabel accepts a string identifier and returns the node labelled with that identifier.
// Might be useful to return a pointer to the node.
func (g ADIGraph) GetByLabel(label string) (ADINode, error) {
	if index, ok := g.labels[label]; ok {
		return g.GetByIndex(index)
	}

	return nil, errLabelNotFound
}

// GetLocatorsByIndex accepts the index of a node and returns the integer column and bit offset required
// to locate the node within its neighbors' ADIs.
func (g ADIGraph) GetLocatorsByIndex(n int) (Locator, error) {
	if n == -1 {
		return Locator{0, 0}, errDeleted
	}
	column := n / g.wordSize
	offset := n % g.wordSize

	return Locator{column, 1 << byte(offset)}, nil
}

// GetLocatorsByLabel accepts the label of a node and returns the integer column and bit offset required
// to locate the node within its neighbors' ADIs.
//
// If the label is found, a nil error will also be returned. If the label is not found, the function will return
// zeroed locators and errLabelNotFound.
func (g ADIGraph) GetLocatorsByLabel(label string) (Locator, error) {
	if index, ok := g.labels[label]; ok {
		return g.GetLocatorsByIndex(index)
	}

	return Locator{0, 0}, errLabelNotFound
}

// DeleteByIndex accepts the integer index of a node within the graph and deletes that node.
//
// If the index provided is within the graph, a nil error is returned. If it is not in the graph,
// an errOutOfBounds is returned.
func (g *ADIGraph) DeleteByIndex(index int) error {
	// Get the node's label and flag it in the labels map.
	n, _ := g.GetByIndex(index)
	name := n.Label()
	g.labels[name] = -1

	// Lazy delete.
	n.Delete()

	// If there are any edges connecting to the node, delete them.
	gn := g.nodes
	locs, _ := g.GetLocatorsByIndex(index)
	for _, n := range gn {
		n.RemoveEdges(locs)
	}

	return nil
}

// DeleteByLabel accepts an identifier string and deletes the node labeled with that identifier.
//
// If the label is found, returns a nil error. If not, returns errLabelNotFound.
func (g *ADIGraph) DeleteByLabel(label string) error {
	if index, ok := g.labels[label]; ok {
		g.DeleteByIndex(index)
		return nil
	}
	return errLabelNotFound
}

// BFS from A to B
//	1. Get node A's ADI, if B is in it, return found.
//	2. Else append the ADIs of each neighbor to a list.
//	3. Map isConnected(B) to each node in the list.
func (g ADIGraph) BFS(a, b ADINode) bool {
	// Get the locators for B
	bLoc, _ := g.GetLocatorsByLabel(b.Label())
	if a.HasEdges(false, bLoc) {
		return true
	}

	var queue []ADINode
	// Else add A's neighbors to the queue.
	for _, n := range g.Neighbors(a) {
		queue = append(queue, n)
	}

	for len(queue) > 0 {
		for _, v := range queue {
			fmt.Printf("%s, ", v.Label())
		}
		fmt.Println("")
		if queue[0].HasEdges(false, bLoc) {
			return true
		}
		for _, item := range g.Neighbors(queue[0]) {
			queue = append(queue, item)
		}
		queue = queue[1:]
	}

	return false
}

/*
	The graph should altogether look like:
		Node	|	ADIs
		"one"   | [0]: 12, [1]: 34
		"two"   | [0]: 54, [1]: 96

	The second array of ADIs is only allocated when the amount of Nodes is at capacity.

	Should use slices because the growability is built-in and I'm not trying to reinvent the wheel.

	Growth algorithm:
		- Array starts with no nodes.
		- If the length of the array mod the word size == 0, for every node add one more column to the list of ADIs.
		- Add one to the growth factor.

	Growth Factor:
		- Measures the amount of times the word size has been passed, and the graph thus grown in both directions.
		- GF can come in handy when paralellizing operations and also maybe in copying/resizing the graph.
		- GF == len(nodes % wordSize) == len(ADIs % wordSize) == len(ADIs[0] & wordSize)
		- GF also helps construct IDs for nodes. If word size == 8, the 9th position will need 2 words to hold its ID, or do (1 << (GF * wordSize) + (node.cardinality % wordSize))


	Might be useful to have some counters in the graph.
*/

/* Synchronous Neighbors behavior.
// Neighbors accepts a node and returns the neighbors of the node in the graph.
func (g ADIGraph) Neighbors(n ADINode) []ADINode {
	var indices []int
	var results []ADINode
	gn := g.nodes
	adis := n.Edges()
	// For each ADI, calculate the indices.
	// I can speed this up by parallelizing the check across each possible bit.
	// Instead of running biterate, I could go Lookup(column, offset).
	// If s.adis[column]&offset > 0, return gn[column*word_size] + o.
	// Each goroutine can return an ADINode.
	for col := 0; col < len(adis); col++ {
		offs := g.bIterate(adis[col])
		for _, o := range offs {
			index := (col * g.wordSize) + o
			indices = append(indices, index)
		}
	}

	for _, i := range indices {
		results = append(results, gn[i])
	}
	// Add the indices to an array.
	// Read from the graph at those indices.

	return results
}

// Helper function that gets the individual edges from an ADI.
func (g ADIGraph) bIterate(b byte) []int {
	var offs []int
	for i := byte(0); i < byte(g.wordSize); i++ {
		check := 1 << i
		if b&byte(check) != 0 {
			offs = append(offs, int(i))
		}
	}
	return offs
}
*/
