/*
Package adigo provides an API for building and manipulating compact and fast graphs, using Adjacency Descriptive Integers.

# ADI Graph

	This is the wrapper struct that manages the nodes of the graph. The graph is the host of its own Create, Read, Update, and Delete operations.

# ADI Node

	The ADI Node is the core of the implementation and provides methods to manage its own contents, labels, and edges.

A general-purpose node is provided through Box, but users of the package can implement their own nodes for specialized use-cases by implementing the ADINode interface.

A 0100 0000
B 1000 0000
C 0100 1000
D
E
F
G
H
*/
package adigo

import (
	"errors"
	"sync"
)

/*
------
ERRORS
------
*/

var (
	errGraphBoundsMismatch = errors.New("the number of nodes does not match the bounds of the ADIs")
	errLabelNotFound       = errors.New("label not found")
	errDeleted             = errors.New("node has been deleted")
)

/*
	-----
	TYPES
	-----
*/

// ADIGraph wraps a list of ADINodes, adding columns to itself as needed
// when the number of nodes outgrows the word size.
type ADIGraph struct {
	growthFactor int
	wordSize     int
	nodes        []ADINode
	labels       map[string]int
}

// Locator contains the column and offset used to identify a node in the graph.
type Locator struct {
	column int
	offset byte
}

/*
	-------
	METHODS
	-------
*/

// NewGraph constructs a new ADIGraph with the default word size of 8 bits.
func NewGraph() ADIGraph {
	graph := ADIGraph{}
	graph.wordSize = 8
	graph.labels = make(map[string]int)
	return graph
}

// lookup searches for a specific ADI node in the ADIGraph based on the given parameters.
// It checks if the bit at the specified offset in the ADI byte is set, and if so,
// retrieves the corresponding node from the graph and sends it to the provided channel.
// The method uses a WaitGroup to synchronize the completion of the operation.
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

// Neighbors returns a list of neighboring nodes for the given node in the ADIGraph.
// It concurrently looks up the neighboring nodes using the `lookup` method and returns the results.
// The returned list contains all the neighboring nodes found, excluding any nil values.
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

// Connect takes the label of a node in the graph and connects it to any number of other nodes by label.
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
func (g ADIGraph) GetByIndex(index int) (ADINode, error) {
	if index == -1 {
		return nil, errDeleted
	}
	return g.nodes[index], nil
}

// GetByLabel accepts a string identifier and returns the node labelled with that identifier.
func (g ADIGraph) GetByLabel(label string) (ADINode, error) {
	if index, ok := g.labels[label]; ok {
		return g.GetByIndex(index)
	}

	return nil, errLabelNotFound
}

// GetLocatorsByIndex accepts the index of a node and returns the integer column and bit offset required to locate the node within its neighbors' ADIs.
func (g ADIGraph) GetLocatorsByIndex(n int) (Locator, error) {
	if n == -1 {
		return Locator{0, 0}, errDeleted
	}
	column := n / g.wordSize
	offset := n % g.wordSize

	return Locator{column, 1 << byte(offset)}, nil
}

// GetLocatorsByLabel accepts the label of a node and returns the integer column and bit offset required to locate the node within its neighbors' ADIs.
// If the label is found, a nil error will also be returned. If the label is not found, the function will return
// zeroed locators and errLabelNotFound.
func (g ADIGraph) GetLocatorsByLabel(label string) (Locator, error) {
	if index, ok := g.labels[label]; ok {
		return g.GetLocatorsByIndex(index)
	}

	return Locator{0, 0}, errLabelNotFound
}

// DeleteByIndex accepts the integer index of a node within the graph and lazy deletes that node.
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

// DeleteByLabel accepts an identifier string and lazy deletes the node labeled with that identifier.
// If the label is found, returns a nil error. If not, returns errLabelNotFound.
func (g *ADIGraph) DeleteByLabel(label string) error {
	if index, ok := g.labels[label]; ok {
		g.DeleteByIndex(index)
		return nil
	}
	return errLabelNotFound
}

// BFS performs a breadth-first search on the ADIGraph starting from node A and checks if node B is reachable.
// It returns true if node B is reachable from node A, and false otherwise.
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
