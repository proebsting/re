//  bitset.go -- a datatype for representing sets of small integers
//
//  based originally on code from the Go Playground
//  http://play.golang.org/p/NpHns5EBnQ as of 11-Feb-2014

package rx

import (
	"fmt"
	"math/big"
)

//  A BitSet is a simple bit-mapped representation of a set of small uints.
//  No explicit constructor is needed; use new(BitSet) or &BitSet{}.
type BitSet struct {
	Bits big.Int
}

//  BitSet.Set sets one bit in a BitSet.
func (b *BitSet) Set(bit uint) *BitSet {
	b.Bits.SetBit(&b.Bits, int(bit), 1)
	return b
}

//  BitSet.Clear clears one bit in a BitSet.
func (b *BitSet) Clear(bit uint) *BitSet {
	b.Bits.SetBit(&b.Bits, int(bit), 0)
	return b
}

//  BitSet.Test returns true if the specified BitSet bit is set.
func (b *BitSet) Test(bit uint) bool {
	return b.Bits.Bit(int(bit)) == 1
}

//  BitSet.IsEmpty returns true if no bits are set in the BitSet.
func (b *BitSet) IsEmpty() bool {
	return b.Bits.BitLen() == 0
}

//  BitSet.Count returns the number of bits that are set.
func (b *BitSet) Count() int {
	n := 0
	l := b.lowbit()
	h := b.Bits.BitLen()
	for i := l; i <= h; i++ { // for all values up to highest
		if b.Test(uint(i)) { // if this value is included
			n++ // count it
		}
	}
	return n
}

//  BitSet.lowbit returns the number of the smallest bit set.
//  It returns 0 if the BitSet is empty.
func (b *BitSet) lowbit() int {
	// inspired by thoughts of HAKMEM...
	sub := (&big.Int{}).Sub(&b.Bits, bigone)
	xor := (&big.Int{}).Xor(&b.Bits, sub)
	add := (&big.Int{}).Add(bigone, xor)
	n := add.BitLen() - 2
	if n >= 0 {
		return n
	} else {
		return 0
	}
}

var bigone = big.NewInt(1) // static variable used in lowbit()

//  BitSet.Equals returns true if the argument set is identical to this one.
func (b1 *BitSet) Equals(b2 *BitSet) bool {
	return (b1.Bits.Cmp(&b2.Bits) == 0)
}

//  BitSet.Or produces a new BitSet that is the union of its inputs.
func (b1 *BitSet) Or(b2 *BitSet) *BitSet {
	b3 := new(BitSet)
	b3.Bits.Or(&b1.Bits, &b2.Bits)
	return b3
}

//  BitSet.And produces a new BitSet that is the intersection of its inputs.
func (b1 *BitSet) And(b2 *BitSet) *BitSet {
	b3 := new(BitSet)
	b3.Bits.And(&b1.Bits, &b2.Bits)
	return b3
}

//  BitSet.Members returns a slice containing the values found in the set.
//  This is the easiest way to iterate through the members of a bit set:
//	for _, i := range bset.Members() { ... }
func (b *BitSet) Members() []uint {
	m := make([]uint, 0)
	l := b.lowbit()
	h := b.Bits.BitLen()
	for i := l; i <= h; i++ { // for all values up to highest
		if b.Test(uint(i)) { // if this value is included
			m = append(m, uint(i))
		}
	}
	return m
}

//  BitSet.String() returns a set-notation representation of the bitset.
//  This is used automatically when printing a bitset with "%s" format.
func (b *BitSet) String() string {
	m := b.Members()
	s := make([]byte, 0, 4*len(m))
	s = append(s, '{')
	for _, i := range m {
		s = append(s, fmt.Sprintf(" %d", i)...)
	}
	return string(append(s, " }"...))
}
