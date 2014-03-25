//  charset.go -- bit set extensions for use as sets of characters
//
//  These additional functions support the use of a BitSet as a set of chars.
//  No distinct type is defined, however -- it's still a BitSet.

package rx

import (
	"math/rand"
	"strconv"
)

//  CharSet makes a BitSet from a string of member characters.
func CharSet(s string) *BitSet {
	cs := new(BitSet)
	for _, ch := range s {
		cs.Set(uint(ch))
	}
	return cs
}

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

//  allChars is used and initialized by BitSet.CharCompl.
//  We can't use init() because of nondeterministic call order.
const lowMatch = 0x01  // smallest ch matched: SOH or ^A
const highMatch = 0x7F // largest ch matched: DEL or RUBOUT
var allChars *BitSet   // set of all chars

//  BitSet.RandChar returns a single randomly chosen BitSet element.
//  Printable characters are preferred to unprintables.
func (b *BitSet) RandChar() byte {
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
	// #%#%#% This code is not particularly efficient.
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

//  BitSet.Bracketed() returns a bracket-expression form of a character set,
//  using ranges if appropriate and escaping (only) unprintables.
func (b *BitSet) Bracketed() string {
	l := b.lowbit()
	h := b.Bits.BitLen()
	s := make([]byte, 0)
	s = append(s, '[')
	for i := l; i <= h; i++ { // for all chars up to highest
		if b.Test(uint(i)) { // if char is included
			s = append(s, cprotect(rune(byte(i)))...) // show char
			var j int
			for j = i + 1; b.Test(uint(j)); j++ {
				// count consecutive inclusions
			}
			if j-i > 3 { // if worth using [a-z] form
				i = j - 1
				s = append(s, '-')
				s = append(s, cprotect(rune(byte(i)))...)
			}
		}
	}
	return string(append(s, ']'))
}

//  cprotect returns its argument if printable, else a backslash form.
func cprotect(r rune) string {
	if strconv.IsPrint(r) {
		return string(r)
	} else {
		s := strconv.QuoteRune(r)
		return s[1 : len(s)-1]
	}
}
