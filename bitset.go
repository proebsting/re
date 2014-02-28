//  bitset.go -- a bit set with some cset properties in the ASCII range
//
//  based originally on code from the Go Playground
//  http://play.golang.org/p/NpHns5EBnQ as of 11-Feb-2014
//
//  #%#% Some of these functions are unused and therefore untested.

package rx

import (
	"fmt"
	"math/big"
	"math/rand"
)

//  BitSet is a simple bit-mapped representation of a set of characters.
type BitSet struct {
	bits big.Int
	//#%#% maybe more later: n, low, high???
}

//  CharSet makes a BitSet from a string of member characters.
func CharSet(s string) *BitSet {
	cs := new(BitSet)
	for _, ch := range s {
		cs.Set(uint(ch))
	}
	return cs
}

//  BitSet.Equals returns true if the argument cset is identical to this one.
func (b1 *BitSet) Equals(b2 *BitSet) bool {
	return (b1.bits.Cmp(&b2.bits) == 0)
}

//  BitSet.Set sets one bit in a BitSet.
func (b *BitSet) Set(bit uint) *BitSet {
	b.bits.SetBit(&b.bits, int(bit), 1)
	return b
}

//  BitSet.Test returns true if the specified BitSet bit is set.
func (b *BitSet) Test(bit uint) bool {
	return b.bits.Bit(int(bit)) == 1
}

//  BitSet.IsEmpty returns true if no bits are set in the BitSet.
func (b *BitSet) IsEmpty() bool {
	return b.bits.BitLen() == 0
}

// data structure used (and initialized) by BitSet.CharCompl
// (was global, but had probs with nondeterministic init() call order)
const lowMatch = 0x01  // smallest ch matched: SOH or ^A
const highMatch = 0x7F // largest ch matched: DEL or RUBOUT
var allChars *BitSet   // set of all chars

//  BitSet.CharCompl produces a new BitSet that is the complement of its inputs
//  with respect to the universe of matchable characters \x01-\x7F.
func (b1 *BitSet) CharCompl() *BitSet {
	if allChars == nil {
		allChars = new(BitSet)
		for i := lowMatch; i <= highMatch; i++ {
			allChars.Set(uint(i))
		}
	}
	b3 := new(BitSet)
	b3.bits.Xor(&b1.bits, &allChars.bits)
	return b3
}

//  BitSet.Or produces a new BitSet that is the union of its inputs.
func (b1 *BitSet) Or(b2 *BitSet) *BitSet {
	b3 := new(BitSet)
	b3.bits.Or(&b1.bits, &b2.bits)
	return b3
}

//  BitSet.And produces a new BitSet that is the intersection of its inputs.
func (b1 *BitSet) And(b2 *BitSet) *BitSet {
	b3 := new(BitSet)
	b3.bits.And(&b1.bits, &b2.bits)
	return b3
}

//  BitSet.RandChar returns a single randomly chosen BitSet element.
//  Printable characters are preferred to unprintables.
//  #%#%#% This code is very inefficient.
func (b BitSet) RandChar() byte {
	n := 0                   // number of characters considered
	h := b.bits.BitLen() - 1 // highest eligible char
	if h < 0 {
		return '?' //#%#%#% ERROR cset was empty
	}
	c := byte(h)  // current working choice
	if c < 0x7F { // if initial ch is not unwanted DEL,
		n = 1 // count it as found
	}
	for h--; h > 0; h-- { // check lower valued characters
		if b.Test(uint(h)) { // if eligible to be chosen
			n++ // adjust n for unbiased odds
			if rand.Intn(n) == 0 {
				c = byte(h) // replace tentative choice
			}
		}
		if h <= ' ' && n >= 5 {
			// now entering the unprintables --
			// bail out if 5 choices seen earlier
			break
		}
	}
	return c // return surviving choice
}

//  BitSet.Members() returns a slice containing the values found in the set.
func (b BitSet) Members() []uint16 {
	m := make([]uint16, 0)
	h := b.bits.BitLen()
	//#%#% should go low to high instead of 0 to high
	for i := 0; i <= h; i++ { // for all chars up to highest
		if b.Test(uint(i)) { // if char is included
			m = append(m, uint16(i))
		}
	}
	return m
}

//  BitSet.String() returns a set-notation representation of the bitset.
func (b BitSet) String() string {
	m := b.Members()
	s := make([]byte, 0, 4*len(m))
	s = append(s, '{')
	for _, i := range m {
		s = append(s, fmt.Sprintf(" %d", i)...)
	}
	return string(append(s, " }"...))
}

//  BitSet.Bracketed() returns a bracket-expression form of a character set,
//  using ranges if appropriate and escaping (only) unprintables.
func (b BitSet) Bracketed() string {
	h := b.bits.BitLen()
	s := make([]byte, 0)
	s = append(s, '[')
	for i := 0; i <= h; i++ { // for all chars up to highest
		if b.Test(uint(i)) { // if char is included
			s = append(s, Cprotect(rune(byte(i)))...) // show char
			var j int
			for j = i + 1; b.Test(uint(j)); j++ {
				// count consecutive inclusions
			}
			if j-i > 3 { // if worth using [a-z] form
				i = j - 1
				s = append(s, '-')
				s = append(s, Cprotect(rune(byte(i)))...)
			}
		}
	}
	return string(append(s, ']'))
}
