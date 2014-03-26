//  bit_test.go -- unit tests for the BitSet functions
//
//  These tests are run by the "go test" command.

package rx

import (
	"fmt"
	"testing"
)

//  TestBits exercises BitSet operations and checks the results.
func TestBits(t *testing.T) {
	fmt.Println("bit_test.go: TestBits")
	bs := &BitSet{}
	ck(t, "e00", true, bs.IsEmpty())
	ck(t, "n00", 0, bs.Count())
	ck(t, "l00", 0, bs.lowbit())
	ck(t, "s00", "{ }", bs)
	ck(t, "s0", "{ 0 }", bs.Set(0))
	ck(t, "s2", "{ 0 2 }", bs.Set(2))
	ck(t, "c0", "{ 2 }", bs.Clear(0))
	ck(t, "l2", 2, bs.lowbit())
	ck(t, "s3a", "{ 2 3 }", bs.Set(3))
	ck(t, "s3b", "{ 2 3 }", bs.Set(3)) // no harm setting twice
	ck(t, "s5", "{ 2 3 5 }", bs.Set(5))
	ck(t, "n5", 3, bs.Count())
	ck(t, "c3", "{ 2 5 }", bs.Clear(3))
	ck(t, "t2", true, bs.Test(2))
	ck(t, "t3", false, bs.Test(3))
	ck(t, "e25", false, bs.IsEmpty())
	bs25 := (&BitSet{}).Set(2).Set(5)
	ck(t, "bs25", "{ 2 5 }", bs25)
	bs58 := (&BitSet{}).Set(8).Set(5)
	ck(t, "bs58", "{ 5 8 }", bs58)
	ck(t, "eq25", true, bs25.Equals(bs))
	ck(t, "!eq0", false, bs25.Equals(&BitSet{}))
	ck(t, "eq58", false, bs58.Equals(bs25))
	ck(t, "and", "{ 5 }", bs25.And(bs58))
	ck(t, "or", "{ 2 5 8 }", bs25.Or(bs58))
	cset := CharSet("abcdeiouvw")
	ck(t, "cset", "{ 97 98 99 100 101 105 111 117 118 119 }", cset)
	ck(t, "brac", "[a-eiouvw]", cset.Bracketed())
}

//  ck validates expected versus actual values (after string conversion).
func ck(t *testing.T, label string, expected interface{}, actual interface{}) {
	ex := fmt.Sprint(expected)
	ac := fmt.Sprint(actual)
	if ex != ac {
		t.Error(label, ": expected \"", ex, "\", got \"", ac, "\"")
	}
}
