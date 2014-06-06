//  repnode.go -- parse tree node for replication of a subtree

package rx

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
)

//  A ReplNode represents controlled (or not) replication: e?, e+, e*, e{m,n}.
//  It is a generalization of the parse tree "*" node of textbooks.
//  In an augmented parse tree, m and n cannot exceed 1.
type ReplNode struct {
	Min   int  // minimum number of occurrences (0 or 1)
	Max   int  // maximum (a positive limit, or -1 meaning infinity)
	Child Node // subpattern being replicated
	NodeData
}

//  ReplNode.Children returns a list consisting of the one child.
func (d *ReplNode) Children() []Node {
	return []Node{d.Child}
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
}

//  ReplNode.SetFollow registers FollowPos nodes.
func (d *ReplNode) SetFollow(pmap []*MatchNode) {
	if d.Max != 1 { // if just 1, self can't follow
		for _, i := range d.LastPos.Members() {
			for _, f := range d.FirstPos.Members() {
				pmap[i].Data().FollowPos.Set(f)
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
//  replication operator: e* or e+ or e? or e{n} or e{n,} or e{m,n}.
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

//  replfix returns a replacement subtree if counting is needed, e.g. a{3}.
//  The original subtree is returned if it is okay.
func replfix(d Node) Node {
	r, ok := d.(*ReplNode)
	if !ok {
		return d // nothing to do, not a replication node
	}
	if r.Min < 2 && r.Max < 2 {
		return r // nothing to do, return as is
	}

	// We need to split this node into a concatenation of two or more
	// deep copies, each with a modified ReplNode at the top.
	// Do this by bundling the subtree into a gob and then decoding
	// as many times as needed.
	gob.Register(&MatchNode{}) // register concrete types
	gob.Register(&ConcatNode{})
	gob.Register(&AltNode{})
	gob.Register(&ReplNode{})
	wbuf := new(bytes.Buffer)             // writable output buffer
	enc := gob.NewEncoder(wbuf)           // create encoder
	CkErr(enc.Encode(&r.Child))           // encode the tree
	rbuf := bytes.NewReader(wbuf.Bytes()) // make resettable input buffer

	// Final result will be a concatenation of a left side of duplicate
	// nodes followed by a final node handling leftovers.
	var lside Node = nil                                   // l side = empty
	rside := &ReplNode{r.Min, r.Max, degob(rbuf), nildata} // r side = copy
	for rside.Min > 1 || (rside.Min > 0 && rside.Min < rside.Max) {
		lside = Concatenate(lside, degob(rbuf))
		rside.Min--
		if rside.Max > 0 {
			rside.Max--
		}
	}
	for rside.Max > 1 {
		optr := ReplNode{0, 1, degob(rbuf), nildata}
		lside = Concatenate(lside, &optr)
		rside.Max--
	}
	if rside.Min == 1 && rside.Max == 1 { // if max==min originally
		return Concatenate(lside, rside.Child) // simpler case
	} else {
		return Concatenate(lside, rside) // e.g. regexp{5,*}
	}
}

//  degob converts a gob into a new copy of a subtree.
func degob(buf *bytes.Reader) Node {
	var tree Node
	buf.Seek(0, 0)
	dec := gob.NewDecoder(buf)
	CkErr(dec.Decode(&tree))
	return tree
}
