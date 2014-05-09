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
		cs.Set(int(ch))
	}
	return cs
}

//  BitSet.CharCompl produces a new BitSet that is the complement of its inputs
//  with respect to the universe of matchable characters \x01-\x7F.
func (b1 *BitSet) CharCompl() *BitSet {
	if allChars == nil {
		allChars = new(BitSet)
		for i := lowMatch; i <= highMatch; i++ {
			allChars.Set(i)
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
	n := 0           // number of characters considered
	l := b.LowBit()  // lowest eligible char
	h := b.HighBit() // highest eligible char
	if h < 0 {
		return '?' //#%#%#% ERROR cset was empty
	}
	c := byte(h)  // current working choice
	if c < 0x7F { // if initial ch is not unwanted DEL,
		n = 1 // count it as found
	}
	// #%#%#% This code is not particularly efficient.
	for h--; h >= l; h-- { // check lower valued characters
		if b.Test(h) { // if eligible to be chosen
			n++ // adjust n for unbiased odds
			if rand.Intn(n) == 0 {
				c = byte(h) // replace tentative choice
			}
		}
		if h <= ' ' {
			// now entering the unprintables -- bail out
			break
		}
	}
	return c // return surviving choice
}

//  BitSet.Bracketed() returns a bracket-expression form of a character set,
//  using ranges if appropriate and escaping (only) unprintables.
//  ] and - are specially placed at beginning and end respectively.
func (b *BitSet) Bracketed() string {
	l := b.LowBit()
	h := b.HighBit()
	open := "["
	close := "]"
	s := make([]byte, 0)
	for i := l; i <= h; i++ { // for all chars up to highest
		if b.Test(i) { // if char is included
			if i == ']' {
				open = "[]" // move ']' to front
				continue
			}
			if i == '-' {
				close = "-]" // defer '-' to end
				continue
			}
			s = append(s, cprotect(rune(byte(i)))...) // show char
			var j int
			for j = i + 1; b.Test(j); j++ {
				// count consecutive inclusions
			}
			if j-i > 3 { // if worth using [a-z] form
				if (j-1) == '-' || (j-1) == ']' {
					// don't end range with '-' or ']'
					j--
				}
				i = j - 1
				s = append(s, '-')
				s = append(s, cprotect(rune(byte(i)))...)
			}
		}
	}
	return open + string(s) + close
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
