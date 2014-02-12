//  parse.go -- rx parse tree construction

package rx

import (
	"fmt"
)


//  Parse parses a regular expression and returns a return tree of Nodes.

var rexpr []byte // remaining characters to parse
var postfix byte // postfix replication char, if any: * + ?

var curr Node        // current tree under construction

// oprstack and exprstack are stacks that move in synchrony
var oprstack []byte  // operators associated with pushed expressions
var exprstack []Node // stack of pushed expressions

func Parse(rexpr string) Node {
	var lside, rside Node
	i := 0				// initialize character position
	curr = Epsilon()		// initialize empty parse tree
	oprstack = make([]byte, 0)	// initialize empty operator stack
	exprstack = make([]Node, 0)	// initialize empty expresion stack

	for i < len(rexpr) {		// for every character in regexp

		// dispatch based on next character
		thischar := rexpr[i]
		switch thischar {

		case '?', '*', '+':
			// ignore:  if seen here, any of these characters
			// just replicates an empty string to no effect
			i++
			continue

		case '|', '(':
			// swap of context: clear current expression after
			// pushing it on a stack (and push opr on another)
			oprstack = append(oprstack, thischar)
			exprstack = append(exprstack, curr)
			curr = Epsilon()
			i++
			continue	// don't check/allow postfix replication

		case ')':
			// close parenthesis: gather parts together
			curr = popAlts(curr)	// first handle alternation
			j := len(oprstack) - 1	// pop the opening paren
			if j >= 0 && oprstack[j] == '(' {
				lside = exprstack[j]		// predecessor
				exprstack = exprstack[0:j]	// pop stack
				oprstack = oprstack[0:j]	// pop opr
				rside, incr := replicate(curr, rexpr[i+1:]) 
				curr = Concatenate(lside, rside)
				i += incr + 1
				continue
			}
			// #%#%#% ERROR! need to signal somehow
			rside = Epsilon()

		case '[':
			// bracket expression
			cset, snew := bracketx(rexpr[i:])
			rside = MatchNode{cset}
			i += len(rexpr[i:]) - len(snew) - 1

		case '.':
			// wild character
			// #%#%#% for now treat as [ -~] (ascii printables only)
			cset, _ := bracketx("[ -~]")
			rside = MatchNode{cset}

		case '\\':
			// escaped character
			if i+1 < len(rexpr) {
				i++
				cset := NewCset(string(Escape(rexpr[i])))
				rside = MatchNode{cset}
			} else {
				//#%#%#% ERROR, \ at end
				rside = Epsilon()
			}
		default:
			// single literal character
			rside = MatchNode{NewCset(rexpr[i:i+1])}
		}

		// common code for handling postfix replication
		rside, incr := replicate(rside, rexpr[i+1:])
		curr = Concatenate(curr, rside)
		i += incr + 1
	}

	curr = popAlts(curr)	// check unpopped alternatives at end of string
	//#%#%#% should then check that stack is now empty (no unclosed parens)
	return curr
}

//  pops all consecutive alternatives from the operator/expression stacks
//  and returns the resulting subtree
func popAlts(d Node) Node {
	for j := len(oprstack) - 1; j >= 0 && oprstack[j] == '|'; j-- {
		d = Alternate(exprstack[j], d)
		oprstack = oprstack[0:j]
		exprstack = exprstack[0:j]
	}
	return d
}

//  wrap a replication node around a subtree if followed by posfix ?, *, +
//  always return resulting node and increment (1 or 0) for string pointer
func replicate(d Node, p string) (Node, int) {
	if len(p) == 0 {
	    return d, 0
	}
	switch p[0] {
	case '?':
		return ReplNode{Min: 0, Max: 1, Child: d}, 1
	case '*':
		return ReplNode{Min: 0, Max: -1, Child: d}, 1
	case '+':
		return ReplNode{Min: 1, Max: -1, Child: d}, 1
	default:
		return d, 0
	}
}

func unneeded() { fmt.Println(00) } //#%#% just to ensure a fmt reference
