//  cset.go -- character set a la Icon, but only 128 bits
//
//  based on code from the Go Playground
//  http://play.golang.org/p/NpHns5EBnQ as of 11-Feb-2014
//
//  #%#% Some of these funcs are unused and therefore untested.

package rx

import (
	"math/big"
	"math/rand"
)

//  Cset is a simple bit-mapped representation of a set of characters.
type Cset struct {
	bits big.Int
	//#%#% maybe more later: n, low, high???
}

//  CharSet makes a Cset from a string of member characters.
func CharSet(s string) *Cset { // make new cset from string
	cs := new(Cset)
	for _, ch := range s {
		cs.Set(uint(ch))
	}
	return cs
}

//  Cset.Set sets one bit in a Cset.
func (b *Cset) Set(bit uint) *Cset {
	b.bits.SetBit(&b.bits, int(bit), 1)
	return b
}

//  Cset.Test returns true if the specified Cset bit is set.
func (b *Cset) Test(bit uint) bool {
	return b.bits.Bit(int(bit)) == 1
}

//  Cset.IsEmpty returns true if no bits are set in the Cset.
func (b *Cset) IsEmpty() bool {
	return b.bits.BitLen() == 0
}

// data structure used (and initialized) by Cset.Compl
// (was global, but had probs with nondeterministic init() call order)
const lowMatch = 0x01  // smallest ch matched: SOH or ^A
const highMatch = 0x7F // largest ch matched: DEL or RUBOUT
var allChars *Cset     // set of all chars

//  Cset.Compl produces a new Cset that is the complement of its inputs
//  with respect to the universe of matchable characters \x01-\x7F.
func (b1 *Cset) Compl() *Cset {
	if allChars == nil {
		allChars = new(Cset)
		for i := lowMatch; i <= highMatch; i++ {
			allChars.Set(uint(i))
		}
	}
	b3 := new(Cset)
	b3.bits.Xor(&b1.bits, &allChars.bits)
	return b3
}

//  Cset.Or produces a new Cset that is the union of its inputs.
func (b1 *Cset) Or(b2 *Cset) *Cset {
	b3 := new(Cset)
	b3.bits.Or(&b1.bits, &b2.bits)
	return b3
}

//  Cset.And produces a new Cset that is the intersection of its inputs.
func (b1 *Cset) And(b2 *Cset) *Cset {
	b3 := new(Cset)
	b3.bits.And(&b1.bits, &b2.bits)
	return b3
}

//  Cset.Choose returns a single randomly chosen Cset element.
//  Printable characters are preferred to unprintables.
//  #%#%#% This code is very inefficient.
func (b Cset) Choose() byte {
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

//  Cset.Members() returns a string listing the characters in the set.
func (b Cset) Members() string {
	s := make([]byte, 0)
	h := b.bits.BitLen()
	for i := 0; i <= h; i++ { // for all chars up to highest
		if b.Test(uint(i)) { // if char is included
			s = append(s, byte(i))
		}
	}
	return string(s)
}

//  Cset.String() returns a string representation in brackets,
//  using ranges if appropriate and escaping (only) unprintables.
func (b Cset) String() string {
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
