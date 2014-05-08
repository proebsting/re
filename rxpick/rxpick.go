/*
	rxpick.go - pick out regular expressions by ordinal number

	usage:  rxpick exprfile i ...

	Rxpick reads regular expressions from exprfile and prints
	those corresponding to the ordinals given as command arguments,
	with their metadata if any.

	spring 2014 / gmt
*/
package main

import (
	"fmt"
	"log"
	"os"
	"rx"
	"strconv"
)

func main() {
	// get command line options
	if len(os.Args) < 3 {
		log.Fatal("usage: rxpick exprfile i ...")
	}
	filename := os.Args[1]
	xset := &rx.BitSet{}
	for i := 2; i < len(os.Args); i++ {
		n, err := strconv.Atoi(os.Args[i])
		rx.CkErr(err)
		xset.Set(n)
	}

	// load expressions from file
	exprs := rx.LoadExpressions(filename, nil)

	// print desired entries
	for _, i := range xset.Members() {
		fmt.Printf("\n# { %d }\n", i)
		if i >= len(exprs) {
			fmt.Printf("# OUT OF RANGE\n")
			continue
		}
		for _, k := range rx.KeyList(exprs[i].Meta) {
			fmt.Printf("#%s: %s\n", k, exprs[i].Meta[k])
		}
		fmt.Println(exprs[i].Expr)
	}
}
