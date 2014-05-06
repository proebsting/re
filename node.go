//  node.go -- parse tree node supertype
//
//  It's not clear that this is the best way to accomplish "variant records",
//  nor is it clear that this is not.

package rx

import (
	"fmt"
	"io"
)

//  Node is the "parent class" of all parse tree node subtypes.
//  All subtype fields must also be exportable for use with package gob.
//
//  The four proper subtypes are MatchNode, ConcatNode, ReplNode, and AltNode.
//  Epsilon and Accept are special AltNode and MatchNode forms respectively.
//
//  A parse tree is referenced by its root node.  There is nothing special
//  about the root node, and every subtree is also a valid parse tree.
//  However, the First/Last/Follow sets only make sense in the context of
//  the root node with respect to which they were computed.
type Node interface {
	Data() *NodeData            // return pointer to common data
	Children() []Node           // return children for tree walking
	MinLen() int                //  return min len matched (0 if nullable)
	MaxLen() int                //  return max len matched (-1 for infinity)
	Example([]byte, int) []byte //  append random synthesized example of max repl n to buf
	SetNFL()                    //  set Nullable, FirstPos, LastPos
	SetFollow([]*MatchNode)     //  set FollowPos
	String() string             // return string for printing
}

//  NodeData is included (anonymously) in every Node subtype.
//  The FollowPos sets represent arcs of the NFA implementing the parse tree.
type NodeData struct {
	Nullable  bool    // can this subtree match empty string?
	FirstPos  *BitSet // possible initial nodes ("positions")
	LastPos   *BitSet // possible final nodes ("positions")
	FollowPos *BitSet // positions that can follow in NFA
}

var nildata = NodeData{} // convenient for initialization

//  Specimen generates a synthetic example from a parse tree
func Specimen(tree Node, maxrepl int) string {
	return string(tree.Example(make([]byte, 0), maxrepl))
}

//  VisitFunc is a function to be executed at each node when walking a tree.
type VisitFunc func(d Node)

//  Walk calls pre and post for every node, before and after visiting children.
//  One of these can be nil to accomplish preorder or postorder traversal.
func Walk(tree Node, pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(tree)
	}
	for _, c := range tree.Children() {
		Walk(c, pre, post)
	}
	if post != nil {
		post(tree)
	}
}

//  treenodes prints details of the parse tree.
func (dfa *DFA) ShowTree(f io.Writer, tree Node, label string) {
	ShowLabel(f, label)
	indent := ""
	Walk(tree, func(d Node) {
		indent = indent + "   "
		a := d.Data()
		c := "F"
		if a.Nullable {
			c = "T"
		}
		fmt.Fprintf(f, "%s{%s, ", indent[3:], c)
		for _, k := range a.FirstPos.Members() {
			fmt.Fprint(f, dfa.Leaves[k])
		}
		fmt.Fprint(f, ", ")
		for _, k := range a.LastPos.Members() {
			fmt.Fprint(f, dfa.Leaves[k])
		}
		fmt.Fprintln(f, "}  ", d)
	}, func(d Node) {
		indent = indent[3:]
	})
}
