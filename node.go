//  node.go -- rx parse tree nodes

//  #%#% perhaps this should be broken into multiple files.  but only perhaps.

package rx

import (
	"fmt"
	"math/rand"
)

//  Node is the "parent class" of all parse tree node subtypes.
//  The four proper subtypes are MatchNode, ConcatNode, ReplNode, and AltNode.
//  Epsilon and Accept make special AltNode and MatchNode forms respectively.
//
//  Every Node subtype implements the following pointer receiver methods:
//  Data()		return pointer to NodeData
//  Children()		return list of children for tree walking
//  MinLen()            return minimum length matched (0 if nullable)
//  MaxLen()		return maximum length matched (-1 for infinity)
//  Example(buf,n)	append random synthesized example of max repl n to buf
//  SetNFL()            set Nullable, FirstPos, LastPos; clear FollowPos
//  SetFollow()		set FollowPos
//  SetNFL and SetFollow assume that child values have been previously set.
//
type Node interface {
	Data() *NodeData
	Children() []Node
	MinLen() int
	MaxLen() int
	Example([]byte, int) []byte
	SetNFL()
	SetFollow([]*MatchNode)
}

//  NodeData is included (anonymously) in every Node subtype.
type NodeData struct {
	Nullable  bool    // can this subtree match empty string?
	FirstPos  *BitSet // possible initial nodes ("positions")
	LastPos   *BitSet // possible final nodes ("positions")
	FollowPos *BitSet // positions that can follow in NFA
}

var nildata = NodeData{} // convenient for initilization

// IsEpsilon returns true for an Epsilon node
func IsEpsilon(d Node) bool {
	anode, ok := d.(*AltNode)
	return ok && len(anode.Alts) == 0
}

//  VisitFunc is a type for visiting tree nodes and passing an arbitrary value.
type VisitFunc func(d Node)

//  Walk calls pre and post for every node, before and after visiting children.
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

//---------------------------------------------------------------------------

//  MatchNode is a leaf node that matches exactly one char from a given set.
type MatchNode struct {
	cset *BitSet // the characters that will match
	Posn uint    // integer "position" desgnator of leaf
	NodeData
}

//  MatchAny creates a MatchNode for a given set of characters.
func MatchAny(cs *BitSet) Node {
	return &MatchNode{cs, 0, nildata}
}

//  Accept returns a special MatchNode with an empty cset.
func Accept() Node {
	return &MatchNode{&BitSet{}, 0, nildata}
}

// IsAccept returns true for an Accept node
func IsAccept(d Node) bool {
	mnode, ok := d.(*MatchNode)
	return ok && mnode.cset.IsEmpty()
}

//  MatchNode.Data returns a pointer to the embedded NodeData struct.
func (d *MatchNode) Data() *NodeData { return &d.NodeData }

//  MatchNode.Children returns an empty list
func (d *MatchNode) Children() []Node {
	return barren
}

var barren = make([]Node, 0, 0) // empty list of children

//  MatchNode.MinLen always returns 1.
func (d *MatchNode) MinLen() int { return 1 }

//  MatchNode.MaxLen always returns 1.
func (d *MatchNode) MaxLen() int { return 1 }

//  MatchNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *MatchNode) SetNFL() {
	d.Nullable = false
	d.FirstPos = &BitSet{}
	d.LastPos = &BitSet{}
	d.FollowPos = &BitSet{}
	d.FirstPos.Set(d.Posn)
	d.LastPos.Set(d.Posn)
}

//  MatchNode.SetFollow has nothing to do.
func (d *MatchNode) SetFollow(pmap []*MatchNode) {
}

//  MatchNode.Example appends a single randomly chosen matching character.
func (d *MatchNode) Example(s []byte, n int) []byte {
	//#%#% assumes cset is not empty
	return append(s, d.cset.RandChar())
}

//  MatchNode.string returns a singleton character or a bracketed expression.
func (d *MatchNode) String() string {
	if d.cset.IsEmpty() {
		return "#" // special "accept" node
	}
	s := d.cset.Bracketed()
	if len(s) == 3 {
		return s[1:2] // abbreviate set of one char
	} else {
		return s
	}
}

//---------------------------------------------------------------------------

//  ConcatNode matches the concatenation of two subpatterns.
type ConcatNode struct {
	l Node
	r Node
	NodeData
}

//  ConcatNode.Data returns a pointer to the embedded NodeData struct.
func (d *ConcatNode) Data() *NodeData { return &d.NodeData }

//  ConcatNode.Children returns a list of the two child nodes
func (d *ConcatNode) Children() []Node {
	children := []Node{d.l, d.r}
	return children
}

//  ConcatNode.MinLen sums the min lengths of its subpatterns.
func (d *ConcatNode) MinLen() int {
	return d.l.MinLen() + d.r.MinLen()
}

//  ConcatNode.MaxLen sums the max lengths of its subpatterns.
//  A value of -1 means that the length is unbounded.
func (d *ConcatNode) MaxLen() int {
	llen := d.l.MaxLen()
	rlen := d.r.MaxLen()
	if llen < 0 || rlen < 0 { // if unbounded
		return -1
	} else {
		return llen + rlen
	}
}

//  ConcatNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *ConcatNode) SetNFL() {
	l := d.l.Data()
	r := d.r.Data()
	d.Nullable = l.Nullable && r.Nullable
	if l.Nullable {
		d.FirstPos = l.FirstPos.Or(r.FirstPos)
	} else {
		d.FirstPos = l.FirstPos
	}
	if r.Nullable {
		d.LastPos = r.LastPos.Or(l.LastPos)
	} else {
		d.LastPos = r.LastPos
	}
}

//  ConcatNode.SetFollow registers FollowPos nodes due to concatenation.
func (d *ConcatNode) SetFollow(pmap []*MatchNode) {
	for _, i := range d.l.Data().LastPos.Members() {
		for _, f := range d.r.Data().FirstPos.Members() {
			pmap[i].Data().FollowPos.Set(uint(f))
		}
	}
}

//  ConcatNode.Example appends one example from each subpattern.
func (d *ConcatNode) Example(s []byte, n int) []byte {
	s = d.l.Example(s, n)
	s = d.r.Example(s, n)
	return s
}

//  ConcatNode.String appends a parenthesized concatenation of subpatterns.
func (d *ConcatNode) String() string {
	return fmt.Sprintf("(%s%s)", d.l, d.r)
}

//  Concatenate makes a ConcatNode, optimizing if either arg is an Epsilon
func Concatenate(d Node, e Node) Node {
	if d == nil || IsEpsilon(d) {
		return e
	}
	if e == nil || IsEpsilon(e) {
		return d
	}
	return &ConcatNode{d, e, nildata}
}

//---------------------------------------------------------------------------

//  AltNode represents two or more choices in a pattern: ab|pq|xy.
type AltNode struct {
	Alts []Node
	NodeData
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
	d.FollowPos = &BitSet{}
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

//  Epsilon returns a speial AltNode exhibiting no alternatives.
func Epsilon() Node {
	return &AltNode{make([]Node, 0), nildata}
}

//---------------------------------------------------------------------------

//  ReplNode represents controlled (or not) replication: e?, e+, e*, e{n,m}.
type ReplNode struct {
	Min   int  // minimum number of occurrences (0 or 1)
	Max   int  // maximum (a positive limit, or -1 meaning infinity)
	Child Node // subpattern being replicated
	NodeData
}

//  ReplNode.Data returns a pointer to the embedded NodeData struct.
func (d *ReplNode) Data() *NodeData { return &d.NodeData }

//  ReplNode.Children returns a list consisting of the one child.
func (d *ReplNode) Children() []Node {
	children := []Node{d.Child}
	return children
}

//  ReplNode.MinLen returns the minimum length after replication.
func (d *ReplNode) MinLen() int {
	return d.Min * d.Child.MinLen()
}

//  ReplNode.MaxLen returns the maximum length after replication.
//  A value of -1 means that the length is unbounded.
func (d *ReplNode) MaxLen() int {
	n := d.Child.MaxLen()
	if n == 0 || d.Max == 0 { // if only matches empty string
		return 0
	} else if n < 0 || d.Max < 0 { // if unbounded
		return -1
	} else {
		return d.Max * n // calculable maximum length
	}
}

//  ReplNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *ReplNode) SetNFL() {
	d.Nullable = d.Min == 0 || d.Child.Data().Nullable
	d.FirstPos = d.Child.Data().FirstPos
	d.LastPos = d.Child.Data().LastPos
	d.FollowPos = &BitSet{}
}

//  ReplNode.SetFollow registers FollowPos nodes.
func (d *ReplNode) SetFollow(pmap []*MatchNode) {
	if d.Max != 1 { // if just 1, self can't follow
		for _, i := range d.LastPos.Members() {
			for _, f := range d.FirstPos.Members() {
				pmap[i].Data().FollowPos.Set(uint(f))
			}
		}
	}
}

//  ReplNode.Example produces an example with maximum replication n.
func (d *ReplNode) Example(s []byte, n int) []byte {
	m := n // save original n for propagation to child
	// limit n to maximum allowed by the regexp
	if n > d.Max && d.Max >= 0 {
		n = d.Max
	}
	// choose desired replication count randomly within legal range
	if n > d.Min {
		n = d.Min + rand.Intn(n-d.Min+1)
	} else {
		n = d.Min
	}
	// and finally replicate
	for i := 0; i < n; i++ {
		s = d.Child.Example(s, m)
	}
	return s
}

//  ReplNode.String produces a string representation using a postfix
//  replication operator: e* or e+ or e? or e{n} or e{n,} or e{n,m}.
func (d *ReplNode) String() string {
	if d.Max < 0 {
		if d.Min == 0 {
			return fmt.Sprintf("%s*", d.Child)
		} else if d.Min == 1 {
			return fmt.Sprintf("%s+", d.Child)
		} else {
			return fmt.Sprintf("%s{%d,}", d.Child, d.Min)
		}
	} else if d.Max == d.Min {
		return fmt.Sprintf("%s{%d}", d.Child, d.Min)
	} else if d.Max == 1 && d.Min == 0 {
		return fmt.Sprintf("%s?", d.Child)
	} else {
		return fmt.Sprintf("%s{%d,%d}", d.Child, d.Min, d.Max)
	}
}
