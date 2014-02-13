//  node.go -- rx parse tree nodes

package rx

import (
	"fmt"
	"math/rand"
)

type Node interface { // "parent class" of Node subtypes
	MinLen() int                // minimum length matched (0 if nullable)
	MaxLen() int                // maximum length matched (-1 for infinity)
	Example([]byte, int) []byte // append random synthesized example to s
	// (the int arg gives max replication)
}

// MatchNode matches exactly one character from a predefined set

type MatchNode struct {
	cset *Cset // the characters that will match
}

func (d MatchNode) MinLen() int { return 1 }

func (d MatchNode) MaxLen() int { return 1 }

func (d MatchNode) Example(s []byte, n int) []byte {
	//#%#% assumes cset is not empty
	return append(s, d.cset.Choose())
}

func (d MatchNode) String() string {
	s := d.cset.String()
	if len(s) == 3 {
		return s[1:2] // abbreviate set of one char
	} else {
		return s
	}
}

// ConcatNode matches the concatenation of other patterns

type ConcatNode struct {
	Parts []Node
}

func (d ConcatNode) MinLen() int {
	n := 0
	for _, e := range d.Parts {
		n += e.MinLen()
	}
	return n
}

func (d ConcatNode) MaxLen() int {
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

func (d ConcatNode) Example(s []byte, n int) []byte {
	for _, e := range d.Parts {
		s = e.Example(s, n)
	}
	return s
}

func (d ConcatNode) String() string {
	b := make([]byte, 0)
	b = append(b, '(')
	for _, v := range d.Parts {
		b = append(b, fmt.Sprint(v)...)
	}
	b = append(b, ')')
	return string(b)
}

func Concatenate(d Node, e Node) Node { // return smart concatenation
	if d == nil {
		return e
	}
	if e == nil {
		return d
	}
	lcat, ok := d.(ConcatNode)
	if ok {
		lcat.Parts = append(lcat.Parts, e)
		return lcat
	} else {
		a := make([]Node, 0)
		return ConcatNode{append(a, d, e)}
	}
}

func Epsilon() Node { // return empty concat usable as epsilon
	return ConcatNode{Parts: make([]Node, 0)}
}

//  AltNode represents two or more choices in a pattern: ab|pq|xy.

type AltNode struct {
	Alts []Node
}

func (d AltNode) MinLen() int {
	n := 0
	for i, e := range d.Alts {
		emin := e.MinLen()
		if i == 0 || emin < n {
			n = emin
		}
	}
	return n
}

func (d AltNode) MaxLen() int {
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

func (d AltNode) Example(s []byte, n int) []byte {
	e := d.Alts[rand.Intn(len(d.Alts))]
	return e.Example(s, n)
}

func (d AltNode) String() string {
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

func Alternate(d Node, e Node) Node {
	rside, ok := e.(AltNode)
	if ok {
		rside.Alts = append(rside.Alts, d)
		return rside
	} else {
		a := make([]Node, 0)
		return AltNode{append(a, e, d)}
	}
}

//  ReplNode represents controlled (or not) replication: e? or e+ or e*.

type ReplNode struct {
	Min   int // minimum number of occurrences (0 or 1)
	Max   int // maximum (a positive limit, or -1 meaning infinity)
	Child Node
}

func (d ReplNode) MinLen() int {
	return d.Min * d.Child.MinLen()
}

func (d ReplNode) MaxLen() int {
	n := d.Child.MaxLen()
	if n == 0 || d.Max == 0 { // if only matches empty string
		return 0
	} else if n < 0 || d.Max < 0 { // if unbounded
		return -1
	} else {
		return d.Max * n // calculable maximum length
	}
}

func (d ReplNode) Example(s []byte, n int) []byte {
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
		s = d.Child.Example(s, n)
	}
	return s
}

func (d ReplNode) String() string {
	if d.Max >= 0 {
		return fmt.Sprintf("%s{%d%d}", d.Child, d.Min, d.Max)
	} else {
		return fmt.Sprintf("%s{%d*}", d.Child, d.Min)
	}
}
