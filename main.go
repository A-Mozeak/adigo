package main

import "fmt"

func main() {

	graph := NewGraph()
	for _, name := range people {
		graph.AddNode(&Box{name, "", []byte{0, 0}, false})
	}

	homiesA := []string{"andre", "addy", "arjun"}
	homiesB := []string{"jasmin", "cody", "farnam"}
	homiesC := []string{"bhanu", "punit", "nikhil", "sherene"}
	homiesD := []string{"nar", "noor"}

	graph.Connect("alex", homiesA...)
	graph.Connect("bhanu", homiesB...)
	graph.Connect("arjun", homiesC...)
	graph.Connect("michael", homiesD...)

	alex, _ := graph.GetByLabel("alex")
	farnam, _ := graph.GetByLabel("farnam")
	noor, _ := graph.GetByLabel("noor")

	fmt.Println(graph.BFS(alex, farnam))
	fmt.Println(graph.BFS(alex, noor))
	// It looks like the graph is having trouble reaching the elements in the last set.
	// Perhaps not enough ADIs in there.
}

var people = []string{
	"alex",
	"andre",
	"addy",
	"arjun",
	"bhanu",
	"punit",
	"angelica",
	"michael",
	"evins",
	"kt",
	"sherene",
	"jasmin",
	"nar",
	"noor",
	"cody",
	"farnam",
	"rain",
	"stephanie",
	"nikhil",
}
