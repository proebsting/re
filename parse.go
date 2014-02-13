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

	var lside, rside Node		// left and right side subtrees
	curr = Epsilon()		// initialize empty parse tree
	oprstack = make([]byte, 0)	// initialize empty operator stack
	exprstack = make([]Node, 0)	// initialize empty expresion stack

	for len(rexpr) > 0 {		// for every character in regexp
		// invariant: curr holds the parse tree completed so far
		// invariant: rexpr holds remaining unprocessed characters

		// dispatch based on next character
		thischar := rexpr[0]		// get character
		rexpr = rexpr[1:]		// trim from string
		switch thischar {		// dispatch

		case '?', '*', '+':
			// ignore:  if seen here, any of these characters
			// just replicates an empty string to no effect
			continue

		case '|', '(':
			// swap of context: clear current expression after
			// pushing it on a stack (and push opr on another)
			oprstack = append(oprstack, thischar)
			exprstack = append(exprstack, curr)
			curr = Epsilon()
			continue	// don't check/allow postfix replication

		case ')':
			// close parenthesis: gather parts together
			curr = popAlts(curr)	// first handle alternation
			j := len(oprstack) - 1	// pop the opening paren
			if j >= 0 && oprstack[j] == '(' {
				lside = exprstack[j]		// predecessor
				exprstack = exprstack[0:j]	// pop stack
				oprstack = oprstack[0:j]	// pop opr
				rside, rexpr = replicate(curr, rexpr) 
				curr = Concatenate(lside, rside)
				continue
			}
			// #%#%#% ERROR: no preceding '('!  #%#% need to handle
			rside = Epsilon()

		case '[':
			// bracket expression
			var cset *Cset
			cset, rexpr = bracketx(rexpr)
			rside = MatchNode{cset}

		case '.':
			// wild character
			// #%#%#% for now treat as [ -~] (ascii printables only)
			cset, _ := bracketx(" -~]")
			rside = MatchNode{cset}

		case '\\':
			// escaped character
			if len(rexpr) > 0 {
				cset := NewCset(string(Escape(rexpr[0])))
				rside = MatchNode{cset}
				rexpr = rexpr[1:]
			} else {
				//#%#%#% ERROR, \ at end
				rside = Epsilon()
			}
		default:
			// single literal character
			rside = MatchNode{NewCset(string(thischar))}
		}

		// common code for handling postfix replication
		rside, rexpr = replicate(rside, rexpr)
		curr = Concatenate(curr, rside)
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
//  always return resulting node and remaining string
func replicate(d Node, p string) (Node, string) {
	if len(p) == 0 {
	    return d, p
	}
	switch p[0] {
	case '?':
		return ReplNode{Min: 0, Max: 1, Child: d}, p[1:]
	case '*':
		return ReplNode{Min: 0, Max: -1, Child: d}, p[1:]
	case '+':
		return ReplNode{Min: 1, Max: -1, Child: d}, p[1:]
	default:
		return d, p
	}
}

func unneeded() { fmt.Println(00) } //#%#% just to ensure a fmt reference
