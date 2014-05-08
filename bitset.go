//  bitset.go -- a datatype for representing sets of small integers
//
//  Most of these operations are functional in concept,
//  but a few modify the receiver: Set, Clear, OrWith, AndWith.
//
//  based originally on code from the Go Playground
//  http://play.golang.org/p/NpHns5EBnQ as of 11-Feb-2014

package rx

import (
	"fmt"
	"math/big"
)

//  A BitSet is a simple bit-mapped representation of a set of small ints.
//  No explicit constructor is needed; use new(BitSet) or &BitSet{}.
type BitSet struct {
	Bits big.Int
}

//  BitSet.Set sets one bit in a BitSet.
func (b *BitSet) Set(bit int) *BitSet {
	b.Bits.SetBit(&b.Bits, bit, 1)
	return b
}

//  BitSet.Clear clears one bit in a BitSet.
func (b *BitSet) Clear(bit int) *BitSet {
	b.Bits.SetBit(&b.Bits, bit, 0)
	return b
}

//  BitSet.Test returns true if the specified BitSet bit is set.
func (b *BitSet) Test(bit int) bool {
	return b.Bits.Bit(bit) == 1
}

//  BitSet.IsEmpty returns true if no bits are set in the BitSet.
func (b *BitSet) IsEmpty() bool {
	return b.Bits.BitLen() == 0
}

//  BitSet.Count returns the number of bits that are set.
func (b *BitSet) Count() int {
	n := 0
	l := b.Lowbit()
	h := b.Highbit()
	for i := l; i <= h; i++ { // for all values up to highest
		if b.Test(i) { // if this value is included
			n++ // count it
		}
	}
	return n
}

//  BitSet.Lowbit returns the number of the smallest bit set.
//  It returns 0 if the BitSet is empty.
func (b *BitSet) Lowbit() int {
	// inspired by thoughts of HAKMEM...
	bigTemp.Sub(&b.Bits, bigOne)
	bigTemp.Xor(&b.Bits, bigTemp)
	bigTemp.Add(bigOne, bigTemp)
	n := bigTemp.BitLen() - 2
	if n >= 0 {
		return n
	} else {
		return 0
	}
}

//  BitSet.Highbit returns the number of the highest bit set.
//  It returns -1 if the BitSet is empty.
func (b *BitSet) Highbit() int {
	return b.Bits.BitLen() - 1
}

var bigOne = big.NewInt(1)  // static constant used in Lowbit()
var bigTemp = big.NewInt(0) // static temporary used in Lowbit()

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

//  BitSet.OrWith accomplishes an OR-in-place, eliminating a memory allocation.
func (b1 *BitSet) OrWith(b2 *BitSet) *BitSet {
	b1.Bits.Or(&b1.Bits, &b2.Bits)
	return b1
}

//  BitSet.And produces a new BitSet that is the intersection of its inputs.
func (b1 *BitSet) And(b2 *BitSet) *BitSet {
	b3 := new(BitSet)
	b3.Bits.And(&b1.Bits, &b2.Bits)
	return b3
}

//  BitSet.AndWith accomplishes an And-in-place, eliminating a memory allocn.
func (b1 *BitSet) AndWith(b2 *BitSet) *BitSet {
	b1.Bits.And(&b1.Bits, &b2.Bits)
	return b1
}

//  BitSet.Key returns an unprintable string usable as a map key.
//  (Neither a BitSet nor the underlying big.Int is a legal key type.)
func (b *BitSet) Key() string {
	if b == nil {
		return ""
	} else {
		return string(b.Bits.Bytes())
	}
}

//  BitSet.Members returns a slice containing the values found in the set.
//  This is the easiest way to iterate through the members of a bit set:
//	for _, i := range bset.Members() { ... }
func (b *BitSet) Members() []int {
	m := make([]int, 0, 0) // initial capacity 0 is faster than h-l+1
	l := b.Lowbit()
	h := b.Highbit()
	for i := l; i <= h; i++ { // for all values up to highest
		if b.Test(i) { // if this value is included
			m = append(m, i)
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
