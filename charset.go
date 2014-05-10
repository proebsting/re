//  charset.go -- bit set extensions for use as sets of characters
//
//  These additional functions support the use of a BitSet as a set of chars.
//  No distinct type is defined, however -- it's still a BitSet.

package rx

import (
	"math/rand"
	"strconv"
)

//  Predefined global character sets

var SpaceSet *BitSet = CharSet("\t\n\v\f\r ")
var DigitSet *BitSet = CharSet("0123456789")
var UpperSet *BitSet = CharSet("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var LowerSet *BitSet = CharSet("abcdefghijklmnopqrstuvwxyz")
var LetterSet *BitSet = UpperSet.Or(LowerSet)
var WordSet *BitSet = LetterSet.Or(DigitSet).Set('_')
var PrintSet *BitSet = CharRange(' ', '~')
var AllChars *BitSet = CharRange('\x01', '\x7F') // matched by "."
var NonDigit *BitSet = DigitSet.CharCompl()
var NonSpace *BitSet = SpaceSet.CharCompl()
var NonWord *BitSet = WordSet.CharCompl()

//  CharSet makes a BitSet from a string of member characters.
func CharSet(s string) *BitSet {
	cs := new(BitSet)
	for _, ch := range s {
		cs.Set(int(ch))
	}
	return cs
}

//  CharRange makes a BitSet from a range of characters.
func CharRange(low int, high int) *BitSet {
	cs := new(BitSet)
	for ; low <= high; low++ {
		cs.Set(low)
	}
	return cs
}

//  BitSet.CharCompl produces a new BitSet that is the complement of its inputs
//  with respect to the universe of matchable characters \x01-\x7F.
func (b1 *BitSet) CharCompl() *BitSet {
	b3 := new(BitSet)
	b3.Bits.Xor(&b1.Bits, &AllChars.Bits)
	return b3
}

//  BitSet.RandChar returns a single randomly chosen BitSet element.
//  A printable character is chosen if possible.
//  Selection is not unbiased if the BitSet is non-contiguous.
func (b *BitSet) RandChar() byte {
	low := b.LowBit()            // lowest eligible char
	high := b.HighBit()          // highest eligible char
	if low < ' ' || high > '~' { // if range extends beyond printables
		b2 := b.And(PrintSet) // check reduced range
		if !b2.IsEmpty() {    // and use that if non-empty
			low = b2.LowBit()
			high = b2.HighBit()
		}
	}
	//  pick a random char between low and high inclusive.
	//  if it's part of the set, we're done.
	c := low + rand.Intn(high-low+1)
	if b.Test(c) {
		return byte(c)
	}
	//  otherwise, pick a direction and start walking
	d := rand.Intn(2) // 0 or 1
	d = 1 - 2*d       // 1 or -1
	c = c + d         // first step
	for !b.Test(c) {
		c = c + d
	}
	return byte(c)
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
