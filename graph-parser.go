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
	Dependencies []Node
}

func (n *Node) ChildExists(d Node) (exists bool) {
	for _, c := range n.Dependencies {
		if c.ID == d.ID && c.Version == d.Version {
			exists = true
			break
		}
	}

	return
}

func (n *Node) AddDependency(d Node) {
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

func (n *Nodes) RetrieveByName(name string) (node *Node, err error) {
	for _, v := range n.Nodes {
		if v.Name == name {
			return v, nil
		}
	}

	return &Node{}, fmt.Errorf("not found")
}

func (n *Nodes) RetrieveByNameAndVersion(name, version string) (node *Node, err error) {
	for _, v := range n.Nodes {
		if v.Name == name && v.Version == version {
			return v, nil
		}
	}

	return &Node{}, fmt.Errorf("not found")
}

func (n *Nodes) Store(name, version string) (node *Node, err error) {
	if !n.isExists(name) {
		node = &Node{
			ID:           uuid.New().String(),
			Name:         name,
			Version:      version,
			Dependencies: []Node{},
		}

		n.Nodes = append(n.Nodes, node)
	} else {
		node, err = n.RetrieveByName(name)
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

	var ns Nodes
	ns = Nodes{}

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
			parent, err = ns.RetrieveByNameAndVersion(root, rootVersion)
			if err != nil {
				parent = &Node{
					Name:         root,
					Version:      rootVersion,
					Dependencies: []Node{},
				}
				ns.Nodes = append(ns.Nodes, parent)
			}
		} else {
			col1Row := strings.Split(col1, "@")
			parentName := col1Row[0]
			parentVersion := col1Row[1]

			parent, err = ns.RetrieveByNameAndVersion(parentName, parentVersion)
			if err != nil {
				parent = &Node{
					Name:         parentName,
					Version:      parentVersion,
					Dependencies: []Node{},
				}
				ns.Nodes = append(ns.Nodes, parent)
			}
		}

		col2 := data[1]
		childRaw := strings.Split(col2, "@")
		childName := childRaw[0]
		childVersion := childRaw[1]

		child, err = ns.RetrieveByNameAndVersion(childName, childVersion)
		if err != nil {
			child = &Node{
				Name:         childName,
				Version:      childVersion,
				Dependencies: []Node{},
			}
			ns.Nodes = append(ns.Nodes, child)
		}
		parent.Dependencies = append(parent.Dependencies, *child)

	}

	//fmt.Printf(">>>> %d", len(ns.Nodes))
	//for _, v := range ns.Nodes {
	//	fmt.Printf("Name: %s Version:%s Childs: %d\n", v.Name, v.Version, len(v.Dependencies))
	//}
	//// Write to file

	f, err := os.Create("component.puml")
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)

	w.WriteString("@startuml\n")

	for _, x := range ns.Nodes {
		parent := fmt.Sprintf("[%s:%s]", x.Name, x.Version)
		//fmt.Println(">>> ", parent)
		print(w, parent, x.Dependencies)
	}

	w.WriteString("@endtuml")
	w.Flush()
}

func print(w *bufio.Writer, parent string, n []Node) {
	for _, x := range n {
		child := fmt.Sprintf("[%s:%s]", x.Name, x.Version)
		fmt.Println(parent, "-->", child)
		//w.WriteString(fmt.Sprintf("%s --> %s \n", parent, child))
		if len(x.Dependencies) > 0 {
			print(w, child, x.Dependencies)
		}
	}
}
