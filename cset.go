//  cset.go -- character set a la Icon, but only 128 bits
//
//  based on code from the Go Playground
//  http://play.golang.org/p/NpHns5EBnQ as of 11-Feb-2014

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

//  LowMatch is the smallest character matched by "."
const LowMatch = 0x01 // SOH or ^A
//  HighMatch is the largest character matched by "."
const HighMatch = 0x7F // DEL or RUBOUT
//  AllChars is the set of all matchable characters \x01 - \x7F.
var AllChars *Cset

func init() {
	AllChars = new(Cset)
	for i := LowMatch; i <= HighMatch; i++ {
		AllChars.Set(uint(i))
	}
}

//  NewCset makes a Cset from a string of member characters.
func NewCset(s string) *Cset { // make new cset from string
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

//  Cset.Compl produces a new Cset that is the complement of its inputs
//  with respect to the universe of matchable characters \x01-\x7F.
func (b1 *Cset) Compl() *Cset {
	b3 := new(Cset)
	b3.bits.Xor(&b1.bits, &AllChars.bits)
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
//  #%#%#% It is very inefficient.
func (b Cset) Choose() byte {
	h := b.bits.BitLen() - 1
	if h < 0 {
		return '#' //#%#%#% ERROR cset was empty
	}
	c := byte(h)
	n := 1
	for i := 0; i < h; i++ {
		if b.Test(uint(i)) {
			n++
			if rand.Intn(n) == 0 {
				c = byte(i)
			}
		}
	}
	return c
}

//  Cset.Choose() returns a string representation in brackets,
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
