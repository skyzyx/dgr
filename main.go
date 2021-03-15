package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/awalterschulze/gographviz"
	"github.com/twmb/algoimpl/go/graph"
	dag "github.com/twmb/algoimpl/go/graph"
)

// I have no idea how this graph stuff works. I've picked up a little here and
// there, but I don't understand it well enough to know how to implement it
// myself. I have an interest in learning, but I'd need someone to walk me
// though the algorithm(s) first.
//
// https://xkcd.wtf/1988/

func main() {
	unit := flag.String("node", "", "The node in the sorted list from which to begin.")
	help := flag.Bool("help", false, "Show help.")
	flag.Parse()

	if *help {
		fmt.Println(`Takes a DOT-formatted 'digraph' where dependencies are expressed as
'dependent -> dependency', determines the order in which they need to occur, and
sorts the dependencies by earlier-to-later resolution.

With the -node parameter, you can specify one of the nodes in the graph, and
this will display the list of dependencies beginning with the node you
specified. The intended use-case is to assume that node's dependencies are met,
so just perform the work for that node and everything which depends on it.`)
		fmt.Println("")
		fmt.Println("  -node string The unit from which to begin.")
		fmt.Println("")
		os.Exit(0)
	}

	// Obtain data from stdin.
	reader := bufio.NewReader(os.Stdin)

	// Read the data from the handle as a []byte.
	input, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the data.
	graphAst, err := gographviz.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	// Analyze the data.
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		log.Fatal(err)
	}

	// Instantiate a more generic DAG processor.
	g := dag.New(dag.Directed)
	nodes := make(map[string]dag.Node)

	// Make a mapping from strings to a node.
	for i := range graph.Nodes.Nodes {
		node := graph.Nodes.Nodes[i]
		nodes[node.Name] = g.MakeNode()
	}

	// Make references back to the string values
	for key, node := range nodes {
		*node.Value = key
	}

	// Connect the elements
	for i := range graph.Edges.Edges {
		edge := graph.Edges.Edges[i]

		err := g.MakeEdge(nodes[edge.Src], nodes[edge.Dst])
		if err != nil {
			log.Fatal(err)
		}
	}

	// Sort by dependency; earlier to later.
	sorted := g.TopologicalSort()
	sorted = reverse(sorted)
	order := []string{}

	// The input nodes in DOT format are wrapped in quotation marks. Strip them.
	for i := range sorted {
		v := (*sorted[i].Value).(string)
		order = append(order, v[1:len(v)-1])
	}

	// Find the index of the node (if it exists). Else return -1.
	indexOf := float64(find(order, *unit))

	// Never go below zero.
	index := int(math.Max(0, indexOf))

	// Return the list starting with the indexOf (or zero).
	order = order[index:]

	// Output the list.
	for i := range order {
		fmt.Println(order[i])
	}
}

func reverse(nodes []graph.Node) []graph.Node {
	for i := 0; i < len(nodes)/2; i++ {
		j := len(nodes) - i - 1
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}

	return nodes
}

func find(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}

	return -1
}
