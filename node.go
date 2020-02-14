package adigo

// ADINode describes the interface that a node should have to be compatible with the graph.
// See the implementation of Box for descriptions of the interface methods.
type ADINode interface {
	Label() string
	Contents() interface{}
	SetLabel(string)
	SetContents(interface{})
	Edges() []byte
	AddColumn()
	AddEdges(...Locator)
	RemoveEdges(...Locator)
	HasEdges(bool, ...Locator) bool
	Deleted() bool
	Delete()
}

// Box is a basic ADI Node that takes any interface as its contents.
type Box struct {
	label    string
	contents interface{}
	adis     []byte
	deleted  bool
}

// Label returns the string identifying a given node.
func (s Box) Label() string {
	return s.label
}

// Contents returns the interface contained in the node.
func (s Box) Contents() interface{} {
	return s.contents
}

// SetLabel accepts a string and labels the node with it.
func (s *Box) SetLabel(l string) {
	s.label = l
}

// SetContents accepts any interface and inserts it into the node's contents.
func (s *Box) SetContents(stuff interface{}) {
	s.contents = stuff
}

// Edges returns the list of ADIs representing the edges that connect the node to other nodes.
func (s Box) Edges() []byte {
	return s.adis
}

// AddColumn adds a new ADI to the node's list of ADIs. Used to connect nodes that are outside of the
// bitset size.
func (s *Box) AddColumn() {
	s.adis = append(s.adis, 0)
}

// AddEdges accepts the set of locators from a node and connects it with the receiver node.
func (s *Box) AddEdges(locators ...Locator) {
	for _, v := range locators {
		s.adis[v.column] |= v.offset
	}
}

// RemoveEdges accepts the set of locators from a node and disconnects it from the receiver node.
// If delete is set to true, shifts the bits over to reference the proper nodes.
func (s *Box) RemoveEdges(locators ...Locator) {
	for _, v := range locators {
		s.adis[v.column] &= ^v.offset
	}
}

// HasEdges accepts the set of locators from a node and a strict flag.
//
// If strict is true, the function will return true only if all of the specified edges are connected. If strict
// is false, the function will return true if any of the specified edges are connected.
func (s *Box) HasEdges(strict bool, locators ...Locator) bool {
	if strict {
		for _, v := range locators {
			if s.adis[v.column]&v.offset != v.offset {
				return false
			}
		}
		return true
	}
	for _, v := range locators {
		if s.adis[v.column]&v.offset > 0 {
			return true
		}
	}
	return false
}

// Deleted returns whether or not the box has been lazy-deleted.
func (s Box) Deleted() bool {
	return s.deleted
}

// Delete removes the node from the graph.
func (s *Box) Delete() {
	s.deleted = true
}

// Could rewrite this so that it only checks for words of length word_size.
// Then externally call the function for each column.
