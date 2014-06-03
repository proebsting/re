/*
	rxquest.go -- regular expression selector

	usage:  rxquest [options] exprfile

	-g n	set size of expression group for example generation
	-i n	initialize random seed for reproducible results
	-m n    indicate which mode to run in
	-s n    indicate size of partition for batch algorithm default is 20

	Questions reads a set of regular expressions and conducts a
	dialogue with the user to choose one.
	mode 1 accepts examples and counter examples as input by user then uses
	    algorithm in mode 3
	mode 2 partitions candidate set into goups of size s, asks the best question
	    from each partition and records the answers then keeps only the regexs
	    that match the sequence of question, answer pairs
	mode 3 selects a random sample of size s and selects the best questions based
	    on the number of DFA's in the candidate set accept the word
	mode 0 selects a random sample of size g, refines this subset to one expression
	    and repeats
	[more description...]
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"rx"
	"time"
)

type RegEx struct { // one regexp for JSON output
	Index int    // index number
	Rexpr string // regular expression
}

// global options from command line
var group *int    // number of exprs to consider together
var seed *int     // random number seed
var mode *int     //mode in which to run
var numElem *int  //number of elements in partition
var args []string // positional arguments

//delimiter for accepting more than one example
const delim = '\n'

// main control
func main() {

	filename := cmdline() // process command line

	maxEx := *numElem * 4
	exprs, trees := load(filename)            // load expressions
	qns := make([]string, 0, maxEx)           //create list of question words
	answers := new(rx.BitSet)                 //create list of answers
	dfaList := make([]*rx.DFA, 0, len(trees)) //list of candidate DFAs
	ind := make([]int, 0, len(dfaList))       //index array to track id of each regex

	a := 0
	init := 0

	if *mode == 1 { //run with examples

		//ask for examples
		fmt.Println("Enter some examples: ")

		//read examples
		r := bufio.NewReader(os.Stdin)
		line, err := r.ReadString(delim)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for line != string('\n') {
			line = line[:len(line)-1]
			qns = append(qns, line)
			answers.Set(a)
			a++
			line, err = r.ReadString(delim)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		//ask for counter examples
		fmt.Println("\nEnter some counter-examples: ")
		rr := bufio.NewReader(os.Stdin)

		//read counter examples
		line, err = rr.ReadString(delim)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for line != string('\n') {
			line = line[:len(line)-1]
			qns = append(qns, line)
			line, err = rr.ReadString(delim)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		}

		fmt.Printf("Processing ...\n")

		//build the DFAs of all candidates
		for i := 0; i < len(trees); i++ {
			aDfa := rx.BuildDFA(trees[i])
			dfaList = append(dfaList, aDfa)
			//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
		}

		fmt.Printf("\n")

		//refine the candidate set based on examples and counter examples
		dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind)
		init++

		fmt.Printf("Length of dfaList: %d\n", len(dfaList))
		for i := 0; i < len(ind); i++ {
			fmt.Printf("rx { %d } %s\n", ind[i], exprs[i].Expr)
		}

	} else if *mode == 2 { //partition

		numPerGroup := *numElem

		//build DFAs of all candidates
		for i := 0; i < len(trees); i++ {
			aDfa := rx.BuildDFA(trees[i])
			dfaList = append(dfaList, aDfa)
			//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
		}

		times := 0
		for len(dfaList) > 50 {

			permArray := rand.Perm(len(dfaList)) //randomly permute candidate set
			qns = make([]string, 0, maxEx)       //initialize new set of questions
			answers = new(rx.BitSet)             //initialize new set of answers

			//ask best question in each partition. Return questions and answers
			answers, qns = askBulk(answers, qns, permArray, dfaList, exprs, numPerGroup)

			fmt.Println("Here length of dfaList =", len(dfaList))

			//refine the candidate set based on questions ans answers
			dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind)
			init++

			fmt.Printf("Length of dfaList: %d\n", len(dfaList))

			times++
		}

	} else if *mode == 3 { //ask all DFAs

		numPerGroup := *numElem

		//build DFA's of all candidates
		for i := 0; i < len(trees); i++ {
			aDfa := rx.BuildDFA(trees[i])
			dfaList = append(dfaList, aDfa)
			fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
		}

		times := 0
		for len(dfaList) > 50 {
			//fmt.Println("iteration = ", times)
			//default partition is 20 per group
			permArray := rand.Perm(len(dfaList)) //randomly permute candidate set
			qns = make([]string, 0, maxEx)       //initialize new set of questions
			answers = new(rx.BitSet)             //initialize new set of answers

			//ask best question based on number of DFAs that accept question word
			answers, qns = askBulk2(answers, qns, permArray, dfaList, exprs, numPerGroup)

			fmt.Println("Here length of dfaList =", len(dfaList))
			//fmt.Printf("\n")

			//refine candidate set based on questions and answers
			dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind)
			init++

			fmt.Printf("Length of dfaList: %d\n", len(dfaList))

			times++
		}

	} else { //mode = 0, run basic algorithm
		//build DFA's of all candidates
		for i := 0; i < len(trees); i++ {
			aDfa := rx.BuildDFA(trees[i])
			dfaList = append(dfaList, aDfa)
			//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
		}

	}

	param := 1
	value := *numElem
	tempLen := 0
	done := false
	numPerGroup := 5

	for len(dfaList) > param && !done {

		subjlist := make([]int, 0, *group)
		subjtrees := make([]rx.Node, 0, *group)
		expressions := make([]*RegEx, 0, len(dfaList))

		if init == 0 { //if this is the first iteration, mode = 0

			//pick a random subset
			for i := 0; i < *group && i < len(dfaList); i++ {
				j := rand.Intn(len(dfaList)) // naive; can duplicate
				subjlist = append(subjlist, j)
				subjtrees = append(subjtrees, dfaList[j].Tree)
				expressions = append(expressions, &RegEx{j, exprs[j].Expr})
				fmt.Printf("rx { %d } %s\n", j, exprs[j].Expr)
			}
			init++
		} else if len(dfaList) < value { //if the number of dfa's is small enough, run qns on all remaining

			for i := 0; i < len(dfaList); i++ {
				//j := rand.Intn(len(dfaList)) // naive; can duplicate
				subjlist = append(subjlist, i)
				subjtrees = append(subjtrees, dfaList[i].Tree)
				expressions = append(expressions, &RegEx{ind[i], exprs[i].Expr})
				fmt.Printf("rx { %d } %s\n", ind[i], exprs[i].Expr)
			}

		} else { //if number of dfa's is too large, do nothing
		}

		qns = make([]string, 0, maxEx)       //initialize new set of questions
		answers = new(rx.BitSet)             //initialize new set of answers
		permArray := rand.Perm(len(dfaList)) //randomly permute candidate DFAs
		fmt.Println("len dfa list = ", len(dfaList))

		//ask best question based on number of candidates that accept question word
		answers, qns = askBulk2(answers, qns, permArray, dfaList, exprs, numPerGroup)

		fmt.Printf("\n")

		fmt.Println("before length of dfaList = ", len(dfaList))
		tempLen = len(dfaList) //store current length

		//refine the candidate set based on questions and answers
		dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind)

		fmt.Println("length of dfaList = ", len(dfaList))

		//if refinement did not change candidate list then we have equivalent reg exp
		if tempLen == len(dfaList) {
			done = true
			fmt.Println("Main: There are ", len(dfaList), " reg exs that match your query and they are equivalent:")
			for i := 0; i < len(dfaList); i++ {
				fmt.Printf("rx { %d } %s\n ", ind[i], exprs[i].Expr)
			}
		}

	}

	//if one remaining reg ex, print it
	if len(dfaList) == 1 {
		fmt.Printf("\nMain: The reg ex you are looking for is: rx { %d } %s\n\nMETA DATA: \n", ind[0], exprs[0].Expr)

		// print accumulated metadata
		for _, k := range rx.KeyList(exprs[0].Meta) {
			fmt.Printf("#%s: %v\n", k, exprs[0].Meta[k])
		}
		fmt.Printf("\n")
	}

	//if no remaining reg exs, no match in our library
	if len(dfaList) == 0 {
		fmt.Printf("Main: No reg ex in our library matches your query.\n")
	}

}

// validate and process command line arguments, returning input filename
func cmdline() string {

	group = flag.Int("g", 5, "grouping factor")
	seed = flag.Int("i", 0, "random seed")
	mode = flag.Int("m", 0, "run mode")
	numElem = flag.Int("s", 20, "number of elements per partition")
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		log.Fatal("single filename required")
	}
	if *group < 2 {
		log.Fatal("grouping factor must be 2 or greater")
	}
	if *seed == 0 {
		*seed = 1 + time.Now().Nanosecond()%9998
	}
	rand.Seed(int64(*seed))
	fmt.Printf("seed=%d\n", *seed)
	if *mode > 3 {
		log.Fatal("mode value must be 0, 1, 2 or 3")
	}

	if *numElem > 25 {
		log.Fatal("size must be 0-25")
	}
	fmt.Printf("mode=%d\n", *mode)
	fmt.Printf("size=%d\n", *numElem)
	return args[0] // return filename
}

// load expressions, returning list and augmented parse tree list
func load(filename string) ([]*rx.RegExParsed, []rx.Node) {
	errcount := 0                       // error count
	exprs := make([]*rx.RegExParsed, 0) // expression structure list
	trees := make([]rx.Node, 0)         // augmented parse tree list

	rx.LoadExpressions(filename, func(l *rx.RegExParsed) {
		if l.Tree != nil { // if a real expression
			augtree := rx.Augment(l.Tree, len(trees))
			trees = append(trees, augtree)
			exprs = append(exprs, l)
		} else if l.Err != nil { // if erroneous
			errcount++
		} // else is just a comment
	})
	fmt.Printf("%d expressions loaded\n", len(exprs))
	if errcount != 0 {
		fmt.Printf("(%d improper expressions ignored)\n", errcount)
	}
	return exprs, trees
}

//original recursive algorithm that attempts to eliminate half candidates with each question
func askQuestions(u []*rx.BitSet, h []rx.DFAexample, e []*RegEx, alive *rx.BitSet, track int, iter int, answers *rx.BitSet, qns []string, cand []*rx.DFA) ([]*rx.BitSet, []rx.DFAexample, *rx.BitSet, []string) {

	t := make([]*rx.BitSet, 0, len(u))
	k := make([]rx.DFAexample, 0, len(h))

	//none left alive, means there is no match in the corpus
	if alive.Count() == 0 {
		fmt.Println("No reg ex in our library matches your query.")
		fmt.Println("answers after iter = ", iter, " ", answers.String())
		return t, k, answers, qns
	}

	//if there was at least 1 yes answer and only 1 remaining expr alive, it must be accepted by the target
	if track >= 1 {
		if alive.Count() == 1 {
			if findQn(alive, h) != "" {
				answers.Set(iter)
				qns = append(qns, findQn(alive, h))
			}
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		}
	}

	//if there were 0 yes answers and there is 1 remaining, must ask again or return no match
	if track == 0 && alive.Count() == 1 {
		if len(u) == 0 {
			fmt.Println("No reg ex in our library matches your query.")
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		} else {
			goto askAgain
		}

	askAgain:
		//ask last question
		qns = append(qns, h[0].Example)
		fmt.Println("Does your language accept this word: ", h[0].Example)

		var input1 string
		_, e1 := fmt.Scan(&input1)
		if e1 != nil {
			fmt.Printf("err = %#v\n", e1)
			os.Exit(2)
		}

		if input1 == "Yes" || input1 == "y" || input1 == "Y" || input1 == "yes" {
			track++
			answers.Set(iter)
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		} else if input1 == "No" || input1 == "N" || input1 == "n" || input1 == "no" {
			fmt.Println("No reg ex in our library matches your query.")
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		} else {
			fmt.Println("That is not a valid expression, please try again.")
			goto askAgain
		}

	}

	//if 0 yes answers, and no more questions, no match found
	if track == 0 && len(u) == 0 {
		fmt.Println("No reg ex in our library matches your query.")
		return t, k, answers, qns
	}

	//if no more questions but more than 1 candidate, find the next question
	if len(u) == 0 && alive.Count() > 1 {
		qns = append(qns, findQn(alive, h))
		return t, k, answers, qns
	}

	//Iterate through u to find the rx.BitSet with numleft/2 bits set to 1
	i := 0
	n := float64(alive.Count() / 2)
	size := math.Floor(n)
	//if small enough set, take the cieling
	if alive.Count() < 4 && math.Mod(float64(alive.Count()), 2) == 1 {
		size = size + 1
	}
	cur := u[i].Count()
	c := float64(cur)
	hM := float64(math.Abs(c - size))
	nextQn := u[i]
	word := h[i].Example
	examInt := i

	//find the region that is the intersection of n/2 regexs (or closest to n/2)
	for i := 0; i < len(u); i++ {
		c = float64(u[i].Count())
		if math.Abs(c-size) == 0 {
			nextQn = u[i]
			word = h[i].Example
			examInt = i
			break
		} else if math.Abs(c-size) < hM {
			hM = math.Abs(c - size)
			nextQn = u[i]
			word = h[i].Example
			examInt = i
		} else { /*do nothing*/
		}
	}

	qns = append(qns, word)    //append question to list
	rate := rateQns(qns, cand) //rate the questions

	for i := 0; i < len(rate); i++ {
		fmt.Println(rate[i], " dfas accepted this word ", qns[i])
	}

	//ask if the best question word is accepted
	fmt.Println("Does your language accept this word: ", word)

begin:
	var input string
	_, e1 := fmt.Scan(&input)
	if e1 != nil {
		fmt.Printf("err = %#v\n", e1)
		os.Exit(2)
	}

	//Refine the candidate set based on the answer
	if input == "Yes" || input == "y" || input == "Y" || input == "yes" {
		track++
		if nextQn.Count() == 1 {
			t = append(t, nextQn)
			k = append(k, h[examInt])
			answers.Set(iter)
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		}
		answers.Set(iter)
		alive = alive.And(nextQn)
		for i := 0; i < len(u); i++ {
			if !((nextQn.And(u[i])).IsEmpty()) && !(nextQn.Equals(u[i])) {
				t = append(t, u[i])
				k = append(k, h[i])
			}
		}

		fmt.Println("alive =", alive.String())

	} else if input == "No" || input == "N" || input == "n" || input == "no" {

		bs := new(rx.BitSet)
		for i := 0; i < len(e); i++ {
			if !(nextQn.Test(e[i].Index)) {
				bs.Set(e[i].Index)
			}
		}

		alive = bs.And(alive)
		for i := 0; i < len(u); i++ {
			if !((bs.And(u[i])).IsEmpty()) {
				t = append(t, u[i])
				k = append(k, h[i])
			}
		}

		fmt.Println("alive = ", alive.String())

	} else {
		fmt.Println("That is not a valid expression, please try again.")
		fmt.Println("Does your language accept this word: ", word)
		goto begin
	}
	fmt.Println("answers after iter = ", iter, " ", answers.String())
	iter++

	//recursive call on the new candidate set
	return askQuestions(t, k, e, alive, track, iter, answers, qns, cand)
}

//find the expression that corresponds to the question word
func findExpr(nextQn *rx.BitSet, e []*RegEx) string {

	b := new(rx.BitSet)
	for i := 0; i < len(e); i++ {

		b.Set(e[i].Index)
		if b.Equals(nextQn) {
			return e[i].Rexpr
		}
		b.Clear(e[i].Index)
	}
	return ""
}

//find the question word that corresponds to the given region tag
func findQn(tag *rx.BitSet, e []rx.DFAexample) string {

	for i := 0; i < len(e); i++ {
		if (e[i].RXset).Equals(tag) {
			return e[i].Example
		}
	}
	return ""
}

//refine the candidate set based on the questions and answers provided
func refine(answers *rx.BitSet, qns []string, cand []*rx.DFA, expns []*rx.RegExParsed, dex []int) ([]*rx.DFA, []*rx.RegExParsed, []int) {

	remain := make([]*rx.DFA, 0, len(cand))
	express := make([]*rx.RegExParsed, 0, len(expns))
	index := make([]int, 0, len(expns))

	//iterate through all candidates
	for i := 0; i < len(cand); i++ {
		test := new(rx.BitSet)          //initialize new bitset
		for j := 0; j < len(qns); j++ { //ask all the questions
			if cand[i].Accepts(qns[j]) != nil {
				test.Set(j) //record answers
			}
		}
		if test.Equals(answers) { //if test matches answers, keep this DFA
			remain = append(remain, cand[i])
			express = append(express, expns[i])
			if len(dex) == 0 {
				index = append(index, i)
			} else {
				index = append(index, dex[i])
				if len(cand) < 20 {
					fmt.Printf("%d ", dex[i])
				}
			}

		}

	}
	fmt.Printf("\n")

	return remain, express, index
}

//used when candidate set partitioned. Asks the best question from each partition and records answers
func askBulk(answers *rx.BitSet, qns []string, perm []int, cand []*rx.DFA, expns []*rx.RegExParsed, numPerGrp int) (*rx.BitSet, []string) {

	numGroups := math.Floor(float64(len(cand) / numPerGrp))
	fmt.Println("num groups = ", numGroups)
	word := ""

	//modTest := math.Mod(float64(len(cand)), numGroups)

	//fmt.Println("testMod = ", modTest)

	subjlist := make([]int, 0, *group)
	subjtrees := make([]rx.Node, 0, *group)
	expressions := make([]*RegEx, 0, len(cand))

	for it := 0; it < int(numGroups/2); it++ {

		subjlist = make([]int, 0, *group)
		subjtrees = make([]rx.Node, 0, *group)
		expressions = make([]*RegEx, 0, len(cand))

		//create the grouping
		for j := it * int(numPerGrp); j < (it+1)*numPerGrp; j++ {
			subjlist = append(subjlist, perm[j])
			subjtrees = append(subjtrees, cand[perm[j]].Tree)
			expressions = append(expressions, &RegEx{perm[j], expns[perm[j]].Expr})
			fmt.Printf("rx { %d } %s\n", perm[j], expns[perm[j]].Expr)
		}

		//construct multiDFA
		fmt.Printf("Constructing multiDFA it = %d\n", it)
		dfa := rx.MultiDFA(subjtrees)
		//dfa = dfa.Minimize()
		h := dfa.Synthesize()
		//for _, ex := range h {
		//fmt.Printf("eg %s %s\n", ex.RXset, ex.Example)
		//}

		u := make([]*rx.BitSet, 0, len(h))

		//possible questions
		for i := 0; i < len(h); i++ {
			u = append(u, h[i].RXset)
		}

		//candidate set
		alive := new(rx.BitSet)
		for i := 0; i < len(expressions); i++ {
			alive.Set(expressions[i].Index)
		}

		//fmt.Println("alive = ", alive.String());

		//find best question and ask it
		i := 0
		n := float64(alive.Count() / 2)
		size := math.Floor(n)
		if alive.Count() < 4 && math.Mod(float64(alive.Count()), 2) == 1 {
			size = size + 1
		}
		cur := u[i].Count()
		c := float64(cur)
		hM := float64(math.Abs(c - size))
		word = h[i].Example

		for i := 0; i < len(u); i++ {
			c = float64(u[i].Count())
			if math.Abs(c-size) == 0 {
				word = h[i].Example
				break
			} else if math.Abs(c-size) < hM {
				hM = math.Abs(c - size)
				word = h[i].Example
			} else { /*do nothing*/
			}
		}
		fmt.Println("before qns append")

		qns = append(qns, word)
	}
	//rate questions and print rating
	rate := rateQns(qns, cand)

	for i := 0; i < len(rate); i++ {
		fmt.Println(rate[i], " dfas accepted this word ", qns[i])
	}

	for i := 0; i < len(qns); i++ {

	begin:
		//ask questions
		fmt.Println("Does your language accept this word (): ", qns[i])

		var input string
		_, e1 := fmt.Scan(&input)
		if e1 != nil {
			fmt.Printf("err = %#v\n", e1)
			os.Exit(2)
		}

		if input == "Yes" || input == "y" || input == "Y" || input == "yes" {
			answers.Set(i)

		} else if input == "No" || input == "N" || input == "n" || input == "no" {

		} else {
			fmt.Println("That is not a valid expression, please try again.")
			fmt.Println("Does your language accept this word: ", word)
			goto begin
		}

	}
	fmt.Println("answers = ", answers.String())

	return answers, qns

}

//determines best question by the number of candidates that accept the question word.
func askBulk2(answers *rx.BitSet, qns []string, perm []int, cand []*rx.DFA, expns []*rx.RegExParsed, numPerGrp int) (*rx.BitSet, []string) {

	numGroups := math.Floor(float64(len(cand) / numPerGrp))
	word := ""

	if len(cand) < numPerGrp {
		numPerGrp = len(cand)
	}

	fmt.Println("length cand = ", len(cand))

	subjlist := make([]int, 0, *group)
	subjtrees := make([]rx.Node, 0, *group)
	expressions := make([]*RegEx, 0, len(cand))
	numKeep := make([]int, 0, int(numGroups))
	overlap := make([]*rx.BitSet, 0, int(numGroups))

	it := 0

	subjlist = make([]int, 0, *group)
	subjtrees = make([]rx.Node, 0, *group)
	expressions = make([]*RegEx, 0, len(cand))

	//create the grouping
	for j := it * int(numPerGrp); j < (it+1)*numPerGrp; j++ {
		subjlist = append(subjlist, perm[j])
		subjtrees = append(subjtrees, cand[perm[j]].Tree)
		expressions = append(expressions, &RegEx{perm[j], expns[perm[j]].Expr})
		fmt.Printf("rx { %d } %s \n", perm[j], expns[perm[j]].Expr)
	}

	//construct multi DFA
	fmt.Printf("Constructing multiDFA\n")
	dfa := rx.MultiDFA(subjtrees)
	//dfa = dfa.Minimize()
	h := dfa.Synthesize()
	for _, ex := range h {
		fmt.Printf("eg %s %s\n", ex.RXset, ex.Example)
	}

	u := make([]*rx.BitSet, 0, len(h))

	for i := 0; i < len(h); i++ {
		u = append(u, h[i].RXset)
	}

	//candidates
	alive := new(rx.BitSet)
	for i := 0; i < len(expressions); i++ {
		alive.Set(expressions[i].Index)
	}

	total := 0
	max := 0
	diff := float64(len(cand))
	half := math.Floor(float64(len(cand) / 2))
	maxWord := ""
	over := new(rx.BitSet)
	prevWord := make([]string, 0, len(cand))

	//iterate through all example words generated
	for j := 0; j < len(h); j++ {
		total = 0
		over = new(rx.BitSet) //tracks which candidates accept the question word
		//iterate through all candidates
		for i := 0; i < len(cand); i++ {
			//if the candidate accepts the example word, increment total and update bitset
			if cand[i].Accepts(h[j].Example) != nil {
				total++
				over.Set(i)
			}
		}
		//order question words according to the number of candidates that accept them
		if math.Abs(float64(total)-half) < float64(diff) {
			diff = math.Abs(float64(total) - half)
			max = total
			maxWord = h[j].Example
			prevWord = append(prevWord, maxWord)
			overlap = append(overlap, over)
		}

	}

	fmt.Println("max = ", max, " max word: ", maxWord)

	qns = append(qns, maxWord)
	numKeep = append(numKeep, int(max))

	fmt.Println("alive =", len(cand), ", max eliminate = ", max, ", maxWord =", maxWord)
	fmt.Printf("\n")

	answers = new(rx.BitSet)
	maybe := 0

	for i := 0; i < len(qns); i++ {
		//if answer is maybe, ask the next question in the list
	begin:
		if maybe > 0 {
			if len(prevWord) == 1 {
				qns[i] = h[maybe].Example
				fmt.Println("Does your language accept this word: ", h[maybe].Example)
			} else {
				qns[i] = prevWord[len(prevWord)-maybe-1]
				fmt.Printf("Does your language accept this word: %s, max over = %s\n", prevWord[len(prevWord)-maybe-1], overlap[len(overlap)-maybe-1].String())
			}
		} else {
			fmt.Println("Does your language accept this word: ", qns[i])
		}

		var input string
		_, e1 := fmt.Scan(&input)
		if e1 != nil {
			fmt.Printf("err = %#v\n", e1)
			os.Exit(2)
		}

		if input == "Yes" || input == "y" || input == "Y" || input == "yes" {

			answers.Set(i)

		} else if input == "No" || input == "N" || input == "n" || input == "no" {

		} else if input == "maybe" || input == "m" || input == "M" {
			maybe++
			goto begin
		} else {
			fmt.Println("That is not a valid expression, please try again.")
			fmt.Println("Does your language accept this word: ", word)
			goto begin
		}

	}
	fmt.Println("answers = ", answers.String())

	for i := 0; i < len(qns); i++ {
		fmt.Println(qns[i])
	}

	return answers, qns

}

//rates the question word based on how many candidates accept it
func rateQns(qns []string, cand []*rx.DFA) []int {

	rate := make([]int, 0, len(qns))
	total := 0

	for j := 0; j < len(qns); j++ {
		total = 0
		for i := 0; i < len(cand); i++ {
			if cand[i].Accepts(qns[j]) != nil {
				total++
			}
		}
		rate = append(rate, total)
	}

	return rate
}
