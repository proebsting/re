//  node.go -- rx parse tree nodes

package rx

import (
	"fmt"
	"math/rand"
)

//  Node is the "parent class" of all parse tree node subtypes.
//  Every Node subtype includes an anonymous NodeData node and
//  implements the following pointer receiver methods:
//
//  Data()		return pointer to NodeData
//  Walk(v,pre,post)	walk subtree, calling VisitFunc functions pre & post
//  MinLen()            return minimum length matched (0 if nullable)
//  MaxLen()		return maximum length matched (-1 for infinity)
//  Example(buf,n)	append random synthesized example of max repl n to buf
//  SetNFL()            set Nullable, FirstPos, LastPos attributes
//
type Node interface {
	Data() *NodeData
	Walk(v interface{}, pre VisitFunc, post VisitFunc)
	MinLen() int
	MaxLen() int
	Example([]byte, int) []byte
	SetNFL()
}

//  NodeData is included (anonymously) in every Node subtype.
//  #%#% FirstPos / LastPos should really be sets, not appended lists
type NodeData struct {
	Nullable  bool   // can this subtree match an empty string?
	FirstPos  []Node // set of legal first characters
	LastPos   []Node // set of legal last characters
	FollowPos []Node // positions that can follow in DFA
}

var nildata = NodeData{} // convenient for initilization

//  VisitFunc is a type for visiting tree nodes and passing an arbitrary value.
type VisitFunc func(d Node, v interface{})

//  Epsilon returns an empty concatenation that matches an empty string.
func Epsilon() Node {
	return &ConcatNode{Parts: make([]Node, 0)}
}

//---------------------------------------------------------------------------

//  MatchNode matches exactly one character from a predefined set.
type MatchNode struct {
	cset *Cset // the characters that will match
	NodeData
}

//  MatchNode.Data returns a pointer to the embedded NodeData struct.
func (d *MatchNode) Data() *NodeData { return &d.NodeData }

//  MatchNode.Walk visits nodes in a subtree, calling functions pre and/or post,
//  if non-nil, before and after walking this node's children.
func (d *MatchNode) Walk(v interface{}, pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d, v)
	}
	// no children
	if post != nil {
		post(d, v)
	}
}

//  MatchNode.MinLen always returns 1.
func (d *MatchNode) MinLen() int { return 1 }

//  MatchNode.MaxLen always returns 1.
func (d *MatchNode) MaxLen() int { return 1 }

//  MatchNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *MatchNode) SetNFL() {
	d.Nullable = false
	d.FirstPos = []Node{d}
	d.LastPos = []Node{d}
}

//  MatchNode.Example appends a single randomly chosen matching character.
func (d *MatchNode) Example(s []byte, n int) []byte {
	//#%#% assumes cset is not empty
	return append(s, d.cset.Choose())
}

//  MatchNode.string returns a singleton character or a bracketed expression.
func (d *MatchNode) String() string {
	s := d.cset.String()
	if len(s) == 3 {
		return s[1:2] // abbreviate set of one char
	} else {
		return s
	}
}

//---------------------------------------------------------------------------

//  ConcatNode matches a concatenation of subpatterns.
type ConcatNode struct {
	Parts []Node
	NodeData
}

//  ConcatNode.Data returns a pointer to the embedded NodeData struct.
func (d *ConcatNode) Data() *NodeData { return &d.NodeData }

//  ConcatNode.Walk visits nodes in a subtree, calling functions pre and/or post,
//  if non-nil, before and after walking this node's children.
func (d *ConcatNode) Walk(v interface{}, pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d, v)
	}
	for _, c := range d.Parts {
		c.Walk(v, pre, post)
	}
	if post != nil {
		post(d, v)
	}
}

//  ConcatNode.MinLen sums the min lengths of its subpatterns.
func (d *ConcatNode) MinLen() int {
	n := 0
	for _, e := range d.Parts {
		n += e.MinLen()
	}
	return n
}

//  ConcatNode.MaxLen sums the max lengths of its subpatterns.
//  A value of -1 means that the length is unbounded.
func (d *ConcatNode) MaxLen() int {
	n := 0
	for _, e := range d.Parts {
		l := e.MaxLen()
		if l < 0 { // if unbounded
			return l
		}
		n += l
	}
	return n
}

//  ConcatNode.SetNFL sets the Nullable, FirstPos, LastPos fields.
func (d *ConcatNode) SetNFL() {
	d.Nullable = true
	d.FirstPos = []Node{}
	d.LastPos = []Node{}
	for _, e := range d.Parts {
		e.SetNFL()
		if d.Nullable { // if nullable so far...
			d.FirstPos = append(d.FirstPos, e.Data().FirstPos...)
		}
		if e.Data().Nullable {
			d.LastPos = append(d.LastPos, e.Data().LastPos...)
		} else {
			d.LastPos = e.Data().LastPos
		}
		d.Nullable = d.Nullable && e.Data().Nullable
	}
}

//  ConcatNode.Example appends one example from each subpattern.
func (d *ConcatNode) Example(s []byte, n int) []byte {
	for _, e := range d.Parts {
		s = e.Example(s, n)
	}
	return s
}

//  ConcatNode.String appends a parenthesized concatenation of subpatterns.
func (d *ConcatNode) String() string {
	b := make([]byte, 0)
	b = append(b, '(')
	for _, v := range d.Parts {
		b = append(b, fmt.Sprint(v)...)
	}
	b = append(b, ')')
	return string(b)
}

//  Concatenate makes a ConcatNode, collapsing multiple concatenations.
func Concatenate(d Node, e Node) Node { // return smart concatenation
	if d == nil {
		return e
	}
	if e == nil {
		return d
	}
	lcat, ok := d.(*ConcatNode)
	if ok {
		lcat.Parts = append(lcat.Parts, e)
		return lcat
	} else {
		a := make([]Node, 0)
		return &ConcatNode{append(a, d, e), nildata}
	}
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
func (d *AltNode) Walk(v interface{}, pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d, v)
	}
	for _, c := range d.Alts {
		c.Walk(v, pre, post)
	}
	if post != nil {
		post(d, v)
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
	d.Nullable = false
	d.FirstPos = []Node{}
	d.LastPos = []Node{}
	for _, e := range d.Alts {
		e.SetNFL()
		d.Nullable = d.Nullable || e.Data().Nullable
		d.FirstPos = append(d.FirstPos, e.Data().FirstPos...)
		d.LastPos = append(d.LastPos, e.Data().LastPos...)
	}
}

//  AltNode.Example chooses one subpattern to generate an example.
func (d *AltNode) Example(s []byte, n int) []byte {
	e := d.Alts[rand.Intn(len(d.Alts))]
	return e.Example(s, n)
}

//  AltNode.String shows all subpatterns separated by | in parentheses.
func (d *AltNode) String() string {
	b := make([]byte, 0)
	b = append(b, '(')
	n := len(d.Alts) - 1
	for i := range d.Alts {
		if i > 0 {
			b = append(b, '|')
		}
		b = append(b, fmt.Sprint(d.Alts[n-i])...)
	}
	b = append(b, ')')
	return string(b)
}

//  Alternate makes an AltNode, collapsing multiple alternatives.
func Alternate(d Node, e Node) Node {
	rside, ok := e.(*AltNode)
	if ok {
		rside.Alts = append(rside.Alts, d)
		return rside
	} else {
		a := make([]Node, 0)
		return &AltNode{append(a, e, d), nildata}
	}
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
func (d *ReplNode) Walk(v interface{}, pre VisitFunc, post VisitFunc) {
	if pre != nil {
		pre(d, v)
	}
	d.Child.Walk(v, pre, post)
	if post != nil {
		post(d, v)
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
	d.Child.SetNFL()
	d.Nullable = d.Min == 0 || d.Child.Data().Nullable
	d.FirstPos = d.Child.Data().FirstPos
	d.LastPos = d.Child.Data().LastPos
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
