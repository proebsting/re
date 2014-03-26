//  matnode.go -- parse tree node for matching any one character of a set

package rx

//  MatchNode is a leaf node that matches exactly one char from a given set.
//  It generalizes the textbook leaf node that matches a particular char.
//  A special MatchNode with an empty set represents an acceptance marker.
type MatchNode struct {
	Cset    *BitSet // the characters that will match
	Posn    uint    // integer "position" designator of leaf
	RxIndex uint    // which RE does this Accept node belong to?
	NodeData
}

//  MatchAny creates a MatchNode for a given set of characters.
func MatchAny(cs *BitSet) Node {
	return &MatchNode{cs, 0, 0, nildata}
}

//  Accept returns a special MatchNode with an empty cset.
func Accept(rxindex uint) Node {
	return &MatchNode{&BitSet{}, 0, rxindex, nildata}
}

//  IsAccept returns true for an Accept node.
func IsAccept(d Node) bool {
	mnode, ok := d.(*MatchNode)
	return ok && mnode.Cset.IsEmpty()
}

//  MatchNode.Data returns a pointer to the embedded NodeData struct.
func (d *MatchNode) Data() *NodeData { return &d.NodeData }

//  MatchNode.Children returns an empty list.
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
	d.FirstPos.Set(d.Posn)
	d.LastPos.Set(d.Posn)
}

//  MatchNode.SetFollow has nothing to do.
func (d *MatchNode) SetFollow(pmap []*MatchNode) {
}

//  MatchNode.Example appends a single randomly chosen matching character.
func (d *MatchNode) Example(s []byte, n int) []byte {
	if IsAccept(d) {
		return s // don't alter if Accept node
	} else {
		// assumes cset is not empty
		return append(s, d.Cset.RandChar())
	}
}

//  MatchNode.string returns a singleton character or a bracketed expression.
func (d *MatchNode) String() string {
	if d.Cset.IsEmpty() {
		return "#" // special "accept" node
	}
	s := d.Cset.Bracketed()
	if len(s) == 3 {
		return s[1:2] // abbreviate set of one char
	} else {
		return s
	}
}