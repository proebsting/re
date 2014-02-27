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
//  Walk(pre,post)	walk subtree, calling VisitFunc functions pre & post
//  MinLen()            return minimum length matched (0 if nullable)
//  MaxLen()		return maximum length matched (-1 for infinity)
//  Example(buf,n)	append random synthesized example of max repl n to buf
//  SetNFL()            set Nullable, FirstPos, LastPos; clear FollowPos
//  SetFollow()		set FollowPos
//  SetNFL and SetFollow assume that child values have been previously set.
//
type Node interface {
	Data() *NodeData
	Walk(pre VisitFunc, post VisitFunc)
	MinLen() int
	MaxLen() int
	Example([]byte, int) []byte
	SetNFL()
	SetFollow()
}

//  NodeData is included (anonymously) in every Node subtype.
type NodeData struct {
	Nullable  bool          // can this subtree match an empty string?
	FirstPos  map[Node]bool // set of possible initial nodes ("positions")
	LastPos   map[Node]bool // set of possible final nodes ("positions")
	FollowPos map[Node]bool // set of positions that can follow in NFA
}

var nildata = NodeData{} // convenient for initilization

//  VisitFunc is a type for visiting tree nodes and passing an arbitrary value.
type VisitFunc func(d Node)

//  Growset adds (or replaces) all elements from addl into base.
//  With a nil (or initially empty) base this effects a copy.
func growset(base map[Node]bool, addl map[Node]bool) map[Node]bool {
	if base == nil {
		base = make(map[Node]bool, len(addl))
	}
	for k, v := range addl {
		base[k] = v
	}
	return base
}

// IsEpsilon returns true for an Epsilon node
func IsEpsilon(d Node) bool {
	anode, ok := d.(*AltNode)
	return ok && len(anode.Alts) == 0
}

//---------------------------------------------------------------------------

//  MatchNode is a leaf node that matches exactly one char from a given set.
type MatchNode struct {
	cset *Cset // the characters that will match
	Posn int   // integer "position" desgnator of leaf
	NodeData
}

//  Match creates a MatchNode for a given set of characters.
func Match(cs *Cset) Node {
	return &MatchNode{cs, 0, nildata}
}

//  Accept returns a special MatchNode with an empty cset.
func Accept() Node {
	return &MatchNode{nil, 0, nildata}
}

//  MatchNode.Data returns a pointer to the embedded NodeData struct.
func (d *MatchNode) Data() *NodeData { return &d.NodeData }

//  MatchNode.Walk visits nodes in a subtree, calling functions pre and/or post,
//  if non-nil, before and after walking this node's children.
func (d *MatchNode) Walk(pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d)
	}
	// no children
	if post != nil {
		post(d)
	}
}

//  MatchNode.MinLen always returns 1.
func (d *MatchNode) MinLen() int { return 1 }

//  MatchNode.MaxLen always returns 1.
func (d *MatchNode) MaxLen() int { return 1 }

//  MatchNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *MatchNode) SetNFL() {
	d.Nullable = false
	d.FirstPos = make(map[Node]bool)
	d.LastPos = make(map[Node]bool)
	d.FollowPos = make(map[Node]bool)
	d.FirstPos[d] = true
	d.LastPos[d] = true
}

//  MatchNode.SetFollow, applied bottom up, computes followpos sets.
func (d *MatchNode) SetFollow() {
}

//  MatchNode.Example appends a single randomly chosen matching character.
func (d *MatchNode) Example(s []byte, n int) []byte {
	//#%#% assumes cset is not empty
	return append(s, d.cset.Choose())
}

//  MatchNode.string returns a singleton character or a bracketed expression.
func (d *MatchNode) String() string {
	if d.cset == nil {
		return "#" // special "accept" node
	}
	s := d.cset.String()
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

//  ConcatNode.Walk visits nodes in a subtree, calling functions pre and/or post,
//  if non-nil, before and after walking this node's children.
func (d *ConcatNode) Walk(pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d)
	}
	d.l.Walk(pre, post)
	d.r.Walk(pre, post)
	if post != nil {
		post(d)
	}
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
	d.FirstPos = growset(nil, l.FirstPos)
	if l.Nullable {
		growset(d.FirstPos, r.FirstPos)
	}
	d.LastPos = growset(nil, r.LastPos)
	if r.Nullable {
		growset(d.LastPos, l.LastPos)
	}
}

//  ConcatNode.SetFollow, applied bottom up, computes followpos sets.
func (d *ConcatNode) SetFollow() {
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

//  AltNode.Walk visits nodes in a subtree, calling functions pre and/or post,
//  if non-nil, before and after walking this node's children.
func (d *AltNode) Walk(pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d)
	}
	for _, c := range d.Alts {
		c.Walk(pre, post)
	}
	if post != nil {
		post(d)
	}
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
	d.FirstPos = make(map[Node]bool)
	d.LastPos = make(map[Node]bool)
	d.FollowPos = make(map[Node]bool)
	for _, e := range d.Alts {
		d.Nullable = d.Nullable || e.Data().Nullable
		growset(d.FirstPos, e.Data().FirstPos)
		growset(d.LastPos, e.Data().LastPos)
	}
}

//  AltNode.SetFollow, applied bottom up, computes followpos sets.
func (d *AltNode) SetFollow() {
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

//  ReplNode.Walk visits nodes in a subtree, calling functions pre and/or post,
//  if non-nil, before and after walking this node's children.
func (d *ReplNode) Walk(pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d)
	}
	d.Child.Walk(pre, post)
	if post != nil {
		post(d)
	}
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
	d.FirstPos = growset(nil, d.Child.Data().FirstPos)
	d.LastPos = growset(nil, d.Child.Data().LastPos)
	d.FollowPos = make(map[Node]bool)
}

//  ReplNode.SetFollow, applied bottom up, computes followpos sets.
func (d *ReplNode) SetFollow() {
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
