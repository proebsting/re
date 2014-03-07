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
	Bits big.Int
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
	return (b1.Bits.Cmp(&b2.Bits) == 0)
}

//  BitSet.Set sets one bit in a BitSet.
func (b *BitSet) Set(bit uint) *BitSet {
	b.Bits.SetBit(&b.Bits, int(bit), 1)
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

var bigone = big.NewInt(1)

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
	b3.Bits.Xor(&b1.Bits, &allChars.Bits)
	return b3
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

//  BitSet.RandChar returns a single randomly chosen BitSet element.
//  Printable characters are preferred to unprintables.
//  #%#%#% This code is very inefficient.
func (b BitSet) RandChar() byte {
	n := 0                   // number of characters considered
	l := b.lowbit()          // lowest eligible char
	h := b.Bits.BitLen() - 1 // highest eligible char
	if h < 0 {
		return '?' //#%#%#% ERROR cset was empty
	}
	c := byte(h)  // current working choice
	if c < 0x7F { // if initial ch is not unwanted DEL,
		n = 1 // count it as found
	}
	for h--; h >= l; h-- { // check lower valued characters
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
	l := b.lowbit()
	h := b.Bits.BitLen()
	for i := l; i <= h; i++ { // for all chars up to highest
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
	l := b.lowbit()
	h := b.Bits.BitLen()
	s := make([]byte, 0)
	s = append(s, '[')
	for i := l; i <= h; i++ { // for all chars up to highest
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
