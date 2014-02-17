//  cset.go -- character set a la Icon
//
//  based on code from the Go Playground
//  http://play.golang.org/p/NpHns5EBnQ as of 11-Feb-2014

package rx

import (
	"math/big"
	"math/rand"
)

type Cset struct {
	bits big.Int
	//#%#% maybe more later: n, low, high???
}

func NewCset(s string) *Cset { // make new cset from string
	cs := new(Cset)
	for _, ch := range s {
		cs.Set(uint(ch))
	}
	return cs
}

func (b *Cset) Set(bit uint) *Cset {
	b.bits.SetBit(&b.bits, int(bit), 1)
	return b
}

func (b *Cset) Test(bit uint) bool {
	return b.bits.Bit(int(bit)) == 1
}

func (b1 *Cset) Or(b2 *Cset) *Cset {
	b3 := new(Cset)
	b3.bits.Or(&b1.bits, &b2.bits)
	return b3
}

func (b1 *Cset) And(b2 *Cset) *Cset {
	b3 := new(Cset)
	b3.bits.And(&b1.bits, &b2.bits)
	return b3
}

func (b Cset) Choose() byte {
	//#%#%#% really inefficient!
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

//  return string representation, escaping (only) unprintables
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
