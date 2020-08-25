package compiler

import (
	"bytes"
	. "parser"
)

type SymbolTable struct {
	Nodes    []Node
	Parent   *SymbolTable
	Flag     int
	Children []*SymbolTable
}

type Node struct {
	Identifier Token
	Type       Type
}

func (t *SymbolTable) Add(node Node) {
	t.Nodes = append(t.Nodes, node)
}

func (t *SymbolTable) Find(Ident Token) (Node, bool) {
	for _, node := range t.Nodes {
		// print(string(Ident.Buff), "\t", string(node.Identifier.Buff), "\n")
		if bytes.Compare(node.Identifier.Buff, Ident.Buff) == 0 {
			return node, true
		}
	}
	return Node{}, false
}

func (t *SymbolTable) NewChild() *SymbolTable {
	s := SymbolTable{Parent: t}
	t.Children = append(t.Children, &s)
	return &s
}

func (t *SymbolTable) FindInAll(Ident Token) (Node, bool) {
	// print("\n")
	node, ok := t.Find(Ident)
	if ok {
		return node, true
	}
	parent := t.Parent

	for parent != nil {
		node, ok := parent.Find(Ident)
		if ok {
			return node, true
		}
		parent = parent.Parent
	}
	// print("\n")
	return Node{}, false
}
