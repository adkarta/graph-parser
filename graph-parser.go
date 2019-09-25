package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Node struct {
	ID           string
	Name         string
	Version      string
	Dependencies []*Node
}

func (n *Node) ChildExists(d *Node) (exists bool) {
	for _, c := range n.Dependencies {
		if c.ID == d.ID && c.Version == d.Version {
			exists = true
			break
		}
	}

	return
}

func (n *Node) AddDependency(d *Node) {
	if !n.ChildExists(d) {
		n.Dependencies = append(n.Dependencies, d)
	}
}

type Nodes struct {
	Nodes []*Node
}

func (n *Nodes) isExists(name string) (exists bool) {
	for _, v := range n.Nodes {
		if v.Name == name {
			exists = true
			break
		}
	}

	return
}

func (n *Nodes) RetrieveByName(name string) (node *Node) {
	for _, v := range n.Nodes {
		if v.Name == name {
			node = v
			break
		}
	}

	return
}

func (n *Nodes) RetrieveByNameAndVersion(name, version string) (node *Node) {
	for _, v := range n.Nodes {
		if v.Name == name && v.Version == version {
			node = v
			break
		}
	}

	return
}

func (n *Nodes) Store(name, version string) (node *Node) {
	if !n.isExists(name) {
		node = &Node{
			ID:           uuid.New().String(),
			Name:         name,
			Version:      version,
			Dependencies: []*Node{},
		}

		n.Nodes = append(n.Nodes, node)
	} else {
		node = n.RetrieveByName(name)
	}

	return
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	wd, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("unable to determine working directory: %s", err)

		log.SetOutput(os.Stderr)
		log.Fatalf("Error: %+v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "graph")

	cmd.Dir = wd
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("unable execute command: %s", err)

		log.SetOutput(os.Stderr)
		log.Fatalf("Error: %+v", err)
	}

	outStr := string(stdout.Bytes())
	//fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	a := strings.Split(outStr, "\n")

	var ns *Nodes
	ns = &Nodes{}

	root := ""
	rootVersion := "0.0.1"
	for _, v := range a {
		data := strings.Split(v, " ")
		if len(data) <= 1 {
			continue
		}

		var (
			parent *Node
			child  *Node
		)

		col1 := data[0]
		if col1 == root {
			parent = ns.RetrieveByNameAndVersion(root, rootVersion)
			if parent == nil {
				parent = &Node{
					Name:         root,
					Version:      rootVersion,
					Dependencies: []*Node{},
				}
				ns.Nodes = append(ns.Nodes, parent)
			}
		} else {
			col1Row := strings.Split(col1, "@")
			parentName := col1Row[0]
			parentVersion := col1Row[1]

			parent = ns.RetrieveByNameAndVersion(parentName, parentVersion)
			if parent == nil {
				parent = &Node{
					Name:         root,
					Version:      rootVersion,
					Dependencies: []*Node{},
				}
			}
		}

		col2 := data[1]
		childRaw := strings.Split(col2, "@")
		childName := childRaw[0]
		childVersion := childRaw[1]

		child = ns.RetrieveByNameAndVersion(childName, childVersion)
		if child == nil {
			child = &Node{
				Name:         childName,
				Version:      childVersion,
				Dependencies: []*Node{},
			}
			ns.Nodes = append(ns.Nodes, child)
		}
		parent.Dependencies = append(parent.Dependencies, child)

	}

	//for _, v := range ns.Nodes {
	//	fmt.Printf("Name: %s Version:%s Childs: %d\n", v.Name, v.Version, len(v.Dependencies))
	//}

	// Write to file

	f, err := os.Create("component.puml")
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)

	w.WriteString("@startuml\n")

	fmt.Printf(">>>> %d", len(ns.Nodes))
	for _, v := range ns.Nodes {
		//	fmt.Printf("Name: %s Version:%s Childs: %d\n", v.Name, v.Version, len(v.Dependencies))
		parentName := fmt.Sprintf("[%s:%s]", v.Name, v.Version)
		printChild(w, parentName, v.Dependencies)
	}

	w.WriteString("@endtuml")
	w.Flush()
}
