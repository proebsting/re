//  altnode.go -- parse tree node offering multiple alternative subtrees

package rx

import (
	"fmt"
	"math/rand"
)

//  An AltNode generalizes the two-child parse tree node of textbooks to
//  allow an arbitrary number of alternatives.  This makes it easier to
//  give multiple alternatives equal probability when generating examples
//  from the parse tree.  The degenerate form with zero children functions
//  as an Epsilon node.
type AltNode struct {
	Alts []Node
	NodeData
}

//  Epsilon returns a special AltNode exhibiting no alternatives.
func Epsilon() Node {
	return &AltNode{make([]Node, 0), nildata}
}

//  Node.IsEpsilon returns true for an Epsilon node.
func IsEpsilon(d Node) bool {
	anode, ok := d.(*AltNode)
	return ok && len(anode.Alts) == 0
}

//  AltNode.Data returns a pointer to the embedded NodeData struct.
func (d *AltNode) Data() *NodeData { return &d.NodeData }

//  AltNode.Children returns the list of alternative subtrees.
func (d *AltNode) Children() []Node {
	return d.Alts
}

//  AltNode.MinLen returns the smallest minimum of its subpatterns.
func (d *AltNode) MinLen() int {
	n := 0
	for i, e := range d.Alts {
		emin := e.MinLen()
		if i == 0 || emin < n {
			n = emin
		}
	}
	return n
}

//  AltNode.MaxLen returns the largest maxima of its subpatterns.
//  A value of -1 means that the length is unbounded.
func (d *AltNode) MaxLen() int {
	n := 0
	for _, e := range d.Alts {
		emax := e.MaxLen()
		if emax < 0 { // if unbounded
			return -1
		}
		if emax > n {
			n = emax
		}
	}
	return n
}

//  AltNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *AltNode) SetNFL() {
	d.Nullable = (len(d.Alts) == 0) // only if an Epsilon
	d.FirstPos = &BitSet{}
	d.LastPos = &BitSet{}
	for _, e := range d.Alts {
		d.Nullable = d.Nullable || e.Data().Nullable
		d.FirstPos = d.FirstPos.Or(e.Data().FirstPos)
		d.LastPos = d.LastPos.Or(e.Data().LastPos)
	}
}

//  AltNode.SetFollow has nothing to do.
func (d *AltNode) SetFollow(pmap []*MatchNode) {
}

//  AltNode.Example chooses one subpattern to generate an example.
func (d *AltNode) Example(s []byte, n int) []byte {
	if IsEpsilon(d) {
		return s // was an Epsilon
	} else {
		return d.Alts[rand.Intn(len(d.Alts))].Example(s, n)
	}
}

//  AltNode.String shows all subpatterns separated by | in parentheses.
func (d *AltNode) String() string {
	b := make([]byte, 0)
	b = append(b, '(')
	for k, v := range d.Alts {
		if k > 0 {
			b = append(b, '|')
		}
		b = append(b, fmt.Sprint(v)...)
	}
	b = append(b, ')')
	return string(b)
}

//  Alternate makes an AltNode, collapsing multiple alternatives.
func Alternate(d Node, e Node) Node {
	// if left is nil (not Epsilon), just return right
	// (this makes certain loops easier)
	if d == nil {
		return e
	}
	// if right is non-Epsilon AltNode, and left is not, combine
	altd, okd := d.(*AltNode)
	alte, oke := e.(*AltNode)
	if (oke && len(alte.Alts) > 0) && !(okd && len(altd.Alts) > 0) {
		// insert at left end for intuitive ordering
		alist := append(alte.Alts, nil)
		copy(alist[1:], alist[0:])
		alist[0] = d
		alte.Alts = alist
		return alte
	} else {
		return &AltNode{append(make([]Node, 0), d, e), nildata}
	}
}
