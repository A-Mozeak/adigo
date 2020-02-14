# ADIGo
A Go implementation of the Adjacency Descriptive Integer Graph data structure. The data structure uses *Integers* to *Describe* the set of nodes *Adjacent* to a given node in a graph.

## What is an ADI Graph?
Graphs are used throughout computer science to store and relate data. A good example is a Social Graph.

Two of the most common ways to implement a graph are the Adjacency List and the Adjacency Matrix. Adjacency Lists are compact, but handling them often requires linear traversal of linked-lists. Adjacency Matrices are easy to reason about and parallelizable, but often take up far too much space.

The Adjacency Descriptive Integer is an attempt to represent a graph in a compactly, while still allowing for simple reasoning about node connections and parallel operations.

ADI Graphs are a more specific implementation of a bitset graph representation, aiming to exploit the speed of processing register-size operations on 8-, 16-, 32-, and 64- bit machines.