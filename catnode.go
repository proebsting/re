//  catnode.go -- parse tree node for concatenation of two subtrees

package rx

import (
	"fmt"
)

//  ConcatNode is a parse tree node for concatenating two subpatterns.
//  Unlike an AltNode it is *not* generalized to other than two children.
type ConcatNode struct {
	L Node
	R Node
	NodeData
}

//  ConcatNode.Children returns a list of the two child nodes.
func (d *ConcatNode) Children() []Node {
	return []Node{d.L, d.R}
}

//  ConcatNode.MinLen sums the min lengths of its subpatterns.
func (d *ConcatNode) MinLen() int {
	return d.L.MinLen() + d.R.MinLen()
}

//  ConcatNode.MaxLen sums the max lengths of its subpatterns.
//  A value of -1 means that the length is unbounded.
func (d *ConcatNode) MaxLen() int {
	llen := d.L.MaxLen()
	rlen := d.R.MaxLen()
	if llen < 0 || rlen < 0 { // if unbounded
		return -1
	} else {
		return llen + rlen
	}
}

//  ConcatNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *ConcatNode) SetNFL() {
	L := d.L.Data()
	R := d.R.Data()
	d.Nullable = L.Nullable && R.Nullable
	if L.Nullable {
		d.FirstPos = L.FirstPos.Or(R.FirstPos)
	} else {
		d.FirstPos = L.FirstPos
	}
	if R.Nullable {
		d.LastPos = R.LastPos.Or(L.LastPos)
	} else {
		d.LastPos = R.LastPos
	}
}

//  ConcatNode.SetFollow registers FollowPos nodes due to concatenation.
func (d *ConcatNode) SetFollow(pmap []*MatchNode) {
	for _, i := range d.L.Data().LastPos.Members() {
		for _, f := range d.R.Data().FirstPos.Members() {
			pmap[i].Data().FollowPos.Set(f)
		}
	}
}

//  ConcatNode.Example appends one example from each subpattern.
func (d *ConcatNode) Example(s []byte, n int) []byte {
	s = d.L.Example(s, n)
	s = d.R.Example(s, n)
	return s
}

//  ConcatNode.String appends a parenthesized concatenation of subpatterns.
func (d *ConcatNode) String() string {
	return fmt.Sprintf("(%s%s)", d.L, d.R)
}

//  Concatenate makes a ConcatNode, optimizing if either arg is an Epsilon.
func Concatenate(d Node, e Node) Node {
	if d == nil || IsEpsilon(d) {
		return e
	}
	if e == nil || IsEpsilon(e) {
		return d
	}
	return &ConcatNode{d, e, nildata}
}
