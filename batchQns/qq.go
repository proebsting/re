/*
	questions.go -- regular expression selector

	usage:  questions [options] exprfile

	-g n	set size of expression group for example generation
	-i n	initialize random seed for reproducible results

	Questions reads a set of regular expressions and conducts a
	dialogue with the user to choose one...
	[more description...]
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"rx"
	"time"
	"math"
	"os"
	"bufio"
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


const delim = '\n'

// main control
func main() {
	maxEx := 10

	filename := cmdline()          // process command line
	exprs, trees := load(filename) // load expressions
	qns:= make([]string, 0, maxEx)
	answers := new(rx.BitSet)
	dfaList := make([]*rx.DFA, 0, len(trees))
	ind := make([]int, 0, len(dfaList))

	a := 0
	init:= 0

	if(*mode == 1){ //run with examples
		

	   //ask for examples and counter examples
	    fmt.Println("Enter some examples: ")

	    r := bufio.NewReader(os.Stdin)

   	    line, err := r.ReadString(delim)
  	    if err != nil {
      	 	 fmt.Println(err)
       		 os.Exit(1)
   	    }
   	   /*err = r.UnreadByte()
  	   if err != nil {
       		 fmt.Println(err)
       		 os.Exit(1)
   	   }*/

  	  for line != string('\n') {
		line = line[:len(line)-1]
		qns = append(qns, line)
	  	answers.Set(a);
	 	a++;
	     /* _,err = r.ReadByte()
	      if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
              }*/
	      line, err = r.ReadString(delim)
	      if err != nil {
                  fmt.Println(err)
                  os.Exit(1)
              }
	     /* err = r.UnreadByte()
              if err != nil {
                  fmt.Println(err)
                  os.Exit(1)
              }*/
         }

	 fmt.Println("\nEnter some counter-examples: ")
	 rr := bufio.NewReader(os.Stdin)

	 line, err = rr.ReadString(delim)
  	    if err != nil {
      	 	 fmt.Println(err)
       		 os.Exit(1)
   	    }
   	  /* err = rr.UnreadByte()
  	   if err != nil {
       		 fmt.Println(err)
       		 os.Exit(1)
   	   }*/

  	  for line != string('\n') {
		/*err = rr.UnreadByte()
                if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
                }*/
		line = line[:len(line)-1]
		qns = append(qns, line)
	  	//answers.Set(uint(a));
	 	//a++;
	     /* _,err = rr.ReadByte()
	      if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
              }*/
	      line, err = rr.ReadString(delim)
	      if err != nil {
                  fmt.Println(err)
                  os.Exit(1)
              }
	      
	}

fmt.Printf("Processing ...\n")
	//for i:= 0; i< len(qns); i++ {
  	 // fmt.Printf("%s\n", qns[i])
  //  }

/*	   //ask for examples and counter-examples
	   fmt.Println("Enter an example: ")
	
	   var inputEx string
	   _, e1 := fmt.Scan(&inputEx);
	   if (e1 != nil ){
		fmt.Printf("err = %#v\n",e1)
		os.Exit(2)
	   }
	   qns = append(qns, inputEx)
	   answers.Set(uint(a));
	   a++;

	   fmt.Println("\nEnter a counter-example: ")
	
	   var inputCntrEx string
	   _, e2 := fmt.Scan(&inputCntrEx);
	   if (e2 != nil ){
		fmt.Printf("err = %#v\n",e1)
		os.Exit(2)
	   }
	   qns = append(qns, inputCntrEx)
	   //answers.Set(uint(a));
	   a++;
*/
	
	   for i:=0;i<len(trees); i++{
		aDfa := rx.BuildDFA(trees[i])
		dfaList = append(dfaList, aDfa)
		//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
	   }


	   fmt.Println("\nanswers = ", answers.String())
	   //fmt.Println("Questions:")
	   //for i:=0; i<len(qns); i++{
	//	fmt.Printf(" %s\n", qns[i])
	   //}


           fmt.Printf("\n")
	   //refine the candidate set based on questions and answers
	   dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind);
	   init++;	
	
	   fmt.Printf("Length of dfaList: %d, length of index = %d\n", len(dfaList), len(ind))

	   for i:=0;i<len(ind); i++ {
		fmt.Printf("rx { %d } %s\n", ind[i], exprs[i].Expr)
	   }
	//os.Exit(2)
  	}else if( *mode == 2){//partition

	numPerGroup := *numElem

	 for i:=0;i<len(trees); i++{
		aDfa := rx.BuildDFA(trees[i])
		dfaList = append(dfaList, aDfa)
		//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
	   }

	times:=0
	for len(dfaList) > 50 {
	
	   if times >0 {
		numPerGroup = 12
	    }

	   fmt.Println("iteration = ", times)

	   permArray := rand.Perm(len(dfaList))
	   //for i:=0; i<len(permArray); i++{
		//fmt.Println(permArray[i])
	   //}

	
	   //default partition is ~20 per group
	   //numGroups := math.Floor(float64(len(dfaList)/numPerGroup))
	   //fmt.Println("num groups = ", numGroups)
	

	   qns = make([]string, 0, maxEx)
	   answers = new(rx.BitSet)

	   answers, qns = askBulk( answers, qns, permArray, dfaList, exprs, numPerGroup);


	   fmt.Println("Here length of dfaList =", len(dfaList))

	   fmt.Printf("\n")

	  //os.Exit(2)

	   //qnsHalf:= make([]string, 0, maxEx/2)
	   //answersHalf := new(rx.BitSet)

	   //for i:=0; i<len(qns);i++{
	//	fmt.Printf("%s\n", qns[i])
	   //}

	   fmt.Println("exit after answers = ", answers.String())	
	   //fmt.Println(" answersHalf = ", answersHalf.String())	
	   fmt.Println("length of dfaList =", len(dfaList))

	   dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind);
	   init++;	
	
	   fmt.Printf("Length of dfaList: %d\n", len(dfaList))

	   times++
	}

	//fmt.Println("repeat?\n")
		//iter++;
	/*var input1 string
	_, e1 := fmt.Scan(&input1);
	if (e1 != nil ){
		fmt.Printf("err = %#v\n",e1)
		os.Exit(2)
	}
	if(input1 == "y"){
		goto begin
	}else{
		os.Exit(2)
	
	}*/


	}else if(*mode == 3){

	numPerGroup := *numElem

	 for i:=0;i<len(trees); i++{
		aDfa := rx.BuildDFA(trees[i])
		dfaList = append(dfaList, aDfa)
		//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
	   }

	times:=0
	for len(dfaList) > 50 {
	
	   //if times >0 {
		//numPerGroup = 12
	    //}

	   fmt.Println("iteration = ", times)

	   permArray := rand.Perm(len(dfaList))
	   //for i:=0; i<len(permArray); i++{
		//fmt.Println(permArray[i])
	   //}

	
	   //default partition is ~20 per group
	   //numGroups := math.Floor(float64(len(dfaList)/numPerGroup))
	   //fmt.Println("num groups = ", numGroups)
	

	   qns = make([]string, 0, maxEx)
	   answers = new(rx.BitSet)

	   answers, qns = askBulk2( answers, qns, permArray, dfaList, exprs, numPerGroup);

	   fmt.Println("Here length of dfaList =", len(dfaList))

	   fmt.Printf("\n")

	   //qnsHalf:= make([]string, 0, maxEx/2)
	   //answersHalf := new(rx.BitSet)

	  // for i:=0; i<len(qns);i++{
		//fmt.Printf("%s\n", qns[i])
	  // }

	   fmt.Println("exit after answers = ", answers.String())	
	   //fmt.Println(" answersHalf = ", answersHalf.String())	
	   fmt.Println("length of dfaList =", len(dfaList))

	   dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind);
	   init++;	
	
	   fmt.Printf("Length of dfaList: %d\n", len(dfaList))
	   //os.Exit(2)
	   times++
	}

	//fmt.Println("repeat?\n")
		//iter++;
	/*var input1 string

	_, e1 := fmt.Scan(&input1);
	if (e1 != nil ){

		fmt.Printf("err = %#v\n",e1)

		os.Exit(2)

	}

	if(input1 == "y"){
		goto begin

	}else{

		os.Exit(2)

	

	}*/


        }else{ //mode = 0, run as normal

	
	   for i:=0;i<len(trees); i++{
		aDfa := rx.BuildDFA(trees[i])
		dfaList = append(dfaList, aDfa)
		//fmt.Printf("rx { %d } %s\n", i, exprs[i].Expr)
	   }


	}	

	/*for i := 0; i < *group && i < len(dfaList); i++ {
		j := rand.Intn(len(dfaList)) // naive; can duplicate
		subjlist = append(subjlist, j)
		subjtrees = append(subjtrees, dfaList[j].Tree)
		expressions = append(expressions, &RegEx{j, exprs[j].Expr})
		fmt.Printf("rx { %d } %s\n", j, exprs[j].Expr)
	}*/
	
	param := 1
	value := *numElem
	tempLen := 0
	done := false
	numPerGroup := 5

	for (len(dfaList) > param && !done ){

	subjlist := make([]int, 0, *group)
	subjtrees := make([]rx.Node, 0, *group)
	expressions := make([]*RegEx, 0, len(dfaList))
	
	if(init==0){ //if this is the first iteration
	   for i := 0; i < *group && i < len(dfaList); i++ {
		j := rand.Intn(len(dfaList)) // naive; can duplicate
		subjlist = append(subjlist, j)
		subjtrees = append(subjtrees, dfaList[j].Tree)
		expressions = append(expressions, &RegEx{j, exprs[j].Expr})
		fmt.Printf("rx { %d } %s\n", j, exprs[j].Expr)
	   }
	   init++
	}else if (len(dfaList) < value){ //if the number of dfa's is small enough, run qns on all remaining

	   for i := 0; i < len(dfaList); i++ {
		//j := rand.Intn(len(dfaList)) // naive; can duplicate
		subjlist = append(subjlist, i)
		subjtrees = append(subjtrees, dfaList[i].Tree)
		expressions = append(expressions, &RegEx{ind[i], exprs[i].Expr})
		fmt.Printf("rx { %d } %s\n", ind[i], exprs[i].Expr)
	   }

	}else{//if number of dfa's is too large, pick random subset

	   /*for i := 0; i < *group && i < len(dfaList); i++ {
		j := rand.Intn(len(dfaList)) // naive; can duplicate
		subjlist = append(subjlist, j)
		subjtrees = append(subjtrees, dfaList[j].Tree)
		expressions = append(expressions, &RegEx{ind[j], exprs[j].Expr})
		fmt.Printf("rx { %d } %s\n", ind[j], exprs[j].Expr)
	   }*/
	fmt.Printf("\n")
	}
//-------------------------------------------------------------------------------------------
	//  construct a DFA and generate distinguishing examples
//	fmt.Printf("Constructing multiDFA\n")
//	dfa := rx.MultiDFA(subjtrees)
	//dfa = dfa.Minimize()
//	examples := dfa.Synthesize()
//	for _, ex := range examples {
//		fmt.Printf("eg %s %s\n", ex.RXset, ex.Example)
//	}
//	fmt.Printf("size examples = %d\n", len(examples))

	//fmt.Println("test accepts ", dfaList[0].Accepts(examples[0].Example))
	
	/*for i:=0; i<len(expressions); i++ {
		fmt.Println(expressions[i].Index, " , ", expressions[i].Rexpr)
		
	}*/
	//create the list of all possible questions
//	start := make([]*rx.BitSet, 0, len(examples));	

//	for i:=0;i<len(examples); i++{
//		start = append(start, examples[i].RXset)
//	}

	/*for it:=0; it<len(start); it++{
		fmt.Println(start[it].String());
   	}*/

	//Make a bit set for regular expressions that are still alive, initially set to all 1's
//  	 alive := new(rx.BitSet)
//  	 for i:=0; i<len(expressions); i++ { 
//    		 alive.Set(uint(expressions[i].Index));
//  	 }

//	fmt.Println("alive = ", alive.String());
//------------------------------------------------------------------------------------------------------------
	
	//v1 := make([]*rx.BitSet, 0, len(examples));
   	//v2 := make([]rx.DFAexample, 0, len(examples));
	qns = make([]string, 0, maxEx)
	answers = new(rx.BitSet)
	
	//answers contains the sequence of answers and qns contains the questions that were asked
	//v1, v2, answers, qns = askQuestions(start, examples, expressions, alive, 0, 0, answers, qns, dfaList );
	fmt.Println("len dfa list = ", len(dfaList))

	 permArray := rand.Perm(len(dfaList))

	answers, qns = askBulk2( answers, qns, permArray, dfaList, exprs, numPerGroup);
	
	//for i:=0;i<len(qns);i++{
		//fmt.Println(qns[i])
	//}
	fmt.Printf("\n")

	fmt.Println("before length of dfaList = ",len(dfaList) )
	tempLen =len(dfaList)

	//refine the candidate set based on questions and answers
	dfaList, exprs, ind = refine(answers, qns, dfaList, exprs, ind);

	//for i :=0; i<len(ind); i++ {
	//	fmt.Println(ind[i])
	//}
	//os.Exit(2)

	//fmt.Println("length of dfaList = ",len(dfaList), "len exprs = ", len(exprs), "len ind", len(ind) )
	fmt.Println("length of dfaList = ",len(dfaList));	
	
	//for i:=0; i<len(ind); i++{
		//fmt.Println("ind ", i, " = ", ind[i])
	   // fmt.Printf("rx { %d } %s\n", ind[i], exprs[i].Expr)
	//}

	//if refinement did not change candidate list then we have equivalent reg exp
	if(tempLen == len(dfaList)){
		done = true
		fmt.Println("Main: There are ", len(dfaList), " reg exs that match your query and they are equivalent:")
		for i:=0; i<len(dfaList); i++ {
			fmt.Printf("rx { %d } %s\n ",ind[i], exprs[i].Expr)
         	}
	}

	//if(len(v1)<0){ 
   	//    fmt.Println(v1[0].String());
    	//    fmt.Println(v2[0].State);
  	// }    
       }

	//if one remaining reg ex, print it
	if(len(dfaList)==1){
		fmt.Printf("\nMain: The reg ex you are looking for is: rx { %d } %s\n\nMETA DATA: \n",ind[0], exprs[0].Expr)

		// print accumulated metadata
		for _, k := range rx.KeyList(exprs[0].Meta) {
			fmt.Printf("#%s: %v\n", k, exprs[0].Meta[k])
		}
		fmt.Printf("\n")
	}

	//if no remaining reg exs, no match in our library
	if(len(dfaList)==0){
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
	if(*mode > 3){
		log.Fatal("mode value must be 0, 1, 2 or 3")
	}

	if(*numElem > 25){
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

func askQuestions(u []*rx.BitSet, h []rx.DFAexample, e []*RegEx, alive *rx.BitSet, track int, iter int, answers *rx.BitSet, qns []string, cand []*rx.DFA) ([]*rx.BitSet, []rx.DFAexample, *rx.BitSet, []string) {

	t := make([]*rx.BitSet, 0, len(u))
	k := make([]rx.DFAexample, 0, len(h))

	if alive.Count() == 0 {
		fmt.Println("No reg ex in our library matches your query.")
		fmt.Println("answers after iter = ", iter, " ", answers.String())
		return t, k, answers, qns
	}

	if (track >= 1) {
		if alive.Count() == 1 {
			//fmt.Println("The reg ex you are looking for is: ", findExpr(alive,e))
			if( findQn(alive,h) != ""){				
				answers.Set(iter);
				qns = append(qns, findQn(alive,h))
			}
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		}
	}
	if (track == 0 && alive.Count() == 1) {
		//fmt.Println(alive.Count())
		if (len(u) == 0 ){
			fmt.Println("No reg ex in our library matches your query.")
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		} else {
			goto askAgain
		}

		 

	askAgain:
		qns = append(qns, h[0].Example)
		fmt.Println("Does your language accept this word: ", h[0].Example)
		//iter++;
		var input1 string
		_, e1 := fmt.Scan(&input1);
		if (e1 != nil ){
			fmt.Printf("err = %#v\n",e1)
			os.Exit(2)
		}

		if input1 == "Yes" || input1 == "y" || input1 == "Y" || input1 == "yes" {
			track++
			answers.Set(iter);
			//fmt.Println("1The reg ex you are looking for is: ", findExpr(alive,e))
			//qns = append(qns, findQn(alive,h))
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

	if(track == 0 && len(u) == 0){
  	   fmt.Println("No reg ex in our library matches your query.");
		return t, k, answers, qns
         }

	if(len(u) == 0 && alive.Count() > 1 ){
		//fmt.Println("There are ", alive.Count(), " reg exs that match your query and they are equivalent:")
		//for i:=0; i<cap(e)-1; i++ {
			//if((alive.Test(uint(i)))){
		    	 // fmt.Println( e[i].Rexpr)
			//}
         	//}
		qns = append(qns, findQn(alive,h))
		return t, k, answers, qns
     	}

	//Iterate through u to find the rx.BitSet with numleft/2 bits set to 1
	i := 0
	n := float64(alive.Count() / 2)
	size := math.Floor(n)
	if(alive.Count() < 4 && math.Mod(float64(alive.Count()), 2) == 1){
	    size = size+1;
        }
	cur := u[i].Count()
	c := float64(cur)
	hM := float64(math.Abs(c - size))
	nextQn := u[i]
	word := h[i].Example
	examInt := i

	for i := 0; i < len(u); i++ {
		c = float64(u[i].Count())
		//fmt.Println("c = ", c, "i = ", i);
		if math.Abs(c-size) == 0 {
			nextQn = u[i]
			word = h[i].Example
			examInt = i
			break
		} else if math.Abs(c-size) < hM {
			//fmt.Println("hM = ", hM);
			hM = math.Abs(c - size)
			nextQn = u[i]
			word = h[i].Example
			examInt = i
		} else { /*do nothing*/
		}
	}
	qns = append(qns, word)

	rate := rateQns(qns, cand)

	for i:=0; i<len(rate);i++{
		fmt.Println( rate[i], " dfas accepted this word ", qns[i]);
	   }

	fmt.Println("Does your language accept this word: ", word)

begin:
	var input string
	_, e1 := fmt.Scan(&input);
  	 if (e1 != nil ){
		fmt.Printf("err = %#v\n",e1)
		os.Exit(2)
  	 }
	//fmt.Println(nextQn.String());

	if input == "Yes" || input == "y" || input == "Y" || input == "yes" {
		track++
		if nextQn.Count() == 1 {
			t = append(t, nextQn)
			//fmt.Println("2The reg ex you are looking for is: ", findExpr(nextQn,e))
			k = append(k, h[examInt])
			answers.Set(iter);
			//qns = append(qns, findQn(alive,h))
			fmt.Println("answers after iter = ", iter, " ", answers.String())
			return t, k, answers, qns
		}
		answers.Set(iter);
		alive = alive.And(nextQn)
		for i := 0; i < len(u); i++ {
			if ( !((nextQn.And(u[i])).IsEmpty()) && !(nextQn.Equals(u[i])) ) {
				t = append(t, u[i])
				k = append(k, h[i])
			}
		}

		fmt.Println("alive =", alive.String())

	} else if input == "No" || input == "N" || input == "n" || input == "no" {

		/*for i := 0; i < len(u); i++ {
			if (nextQn.And(u[i])).IsEmpty() {
				t = append(t, u[i])
				k = append(k, h[i])
			}
		}*/

		bs := new(rx.BitSet)
		for i := 0; i < len(e); i++ {
			if ( !(nextQn.Test(e[i].Index)) ){
				bs.Set(e[i].Index)
			}// else {
				//bs.Set(uint(i))
			//}
		}

		//fmt.Println(bs.String())

		alive = bs.And(alive)
		for i:=0; i<len(u); i++ {
	  	  if(!((bs.And(u[i])).IsEmpty())){
			t = append(t, u[i]);
			k = append(k,h[i]);
	   	   }
      	        }

		fmt.Println("alive = ", alive.String())

	} else {
		fmt.Println("That is not a valid expression, please try again.")
		fmt.Println("Does your language accept this word: ", word)
		goto begin
	}
	fmt.Println("answers after iter = ", iter, " ", answers.String())
	iter++;
	
	return askQuestions(t, k, e, alive, track, iter, answers, qns, cand)
}

func findExpr(nextQn *rx.BitSet, e []*RegEx) (string) {

	b:= new(rx.BitSet)
	for i:=0;i<len(e); i++{
		
		b.Set(e[i].Index);
		//fmt.Println(b.String())
		//fmt.Println(nextQn.String())
		if(b.Equals(nextQn)){
			//fmt.Println(e[i].Rexpr)
			return e[i].Rexpr;
		}
		b.Clear(e[i].Index);
	}
	return ""
}

func findQn(tag *rx.BitSet, e []rx.DFAexample) (string) {

	//b:= new(rx.BitSet)
	for i:=0;i<len(e); i++{
		
		//b.Set(uint(e[i].Index));
		//fmt.Println(b.String())
		//fmt.Println(nextQn.String())
		if((e[i].RXset).Equals(tag)){
			//fmt.Println(e[i].Rexpr)
			return e[i].Example;
		}
		//b.Clear(uint(e[i].Index));
	}
	return ""
}

func refine(answers *rx.BitSet, qns []string, cand []*rx.DFA, expns []*rx.RegExParsed, dex []int )([]*rx.DFA , []*rx.RegExParsed, []int){

	//test := new(rx.BitSet)
	remain := make([]*rx.DFA,0,len(cand))
	express := make([]*rx.RegExParsed,0,len(expns))
	index := make([]int, 0, len(expns))

	//fmt.Println("answers = ", answers.String())
	//fmt.Println("CHECK---------------------------")
	//if( cand[355].Accepts(qns[0]) != nil ) {
	//	fmt.Println(expns[355].Expr)		
	//}




	for i:=0; i<len(cand); i++{
		test := new(rx.BitSet)
		for j:=0; j<len(qns); j++{
			//fmt.Println("qns = ", qns[j])		
			if( cand[i].Accepts(qns[j]) != nil ) {
				test.Set(j)		
			}
		}
		//fmt.Println("test = ", test.String())
		if( test.Equals(answers) ){
			remain = append(remain, cand[i])
			express = append(express, expns[i])
			if(len(dex) == 0){
				index = append(index, i)
			}else{
				index = append(index, dex[i])
				if(len(cand)< 20){
					fmt.Printf("%d ", dex[i])
				}
			}
			//fmt.Println("i = ", i)
		}	

	}
	fmt.Printf("\n")

	//for i:= 0; i< len(index); i++ {
  	 // fmt.Printf("%d", index[i])
      // }

	//os.Exit(2)

	//fmt.Println("length of remain = ", len(remain))
	return remain, express, index
}

func askBulk(  answers *rx.BitSet, qns []string, perm []int,  cand []*rx.DFA, expns []*rx.RegExParsed, numPerGrp int )(*rx.BitSet, []string){

	numGroups := math.Floor(float64(len(cand)/numPerGrp))
	fmt.Println("num groups = ", numGroups)
	word := ""

	modTest := math.Mod(float64(len(cand)), numGroups)

	fmt.Println("testMod = ", modTest)

	subjlist := make([]int, 0, *group)
	subjtrees := make([]rx.Node, 0, *group)
	expressions := make([]*RegEx, 0, len(cand))

	

	for it:=0; it< int(numGroups/2); it++{

		subjlist = make([]int, 0, *group)
		subjtrees = make([]rx.Node, 0, *group)
		expressions = make([]*RegEx, 0, len(cand))
	
		
		//create the grouping
		for j:=it*int(numPerGrp); j< (it+1)*numPerGrp; j++{
			subjlist = append(subjlist, perm[j])
			subjtrees = append(subjtrees, cand[perm[j]].Tree)
			expressions = append(expressions, &RegEx{perm[j], expns[perm[j]].Expr})
			fmt.Printf("rx { %d } %s\n", perm[j], expns[perm[j]].Expr)
		}

		//os.Exit(2)

		/*fmt.Println("continue? ")

		var input string
		_, e1 := fmt.Scan(&input);
  		 if (e1 != nil ){
			fmt.Printf("err = %#v\n",e1)
			os.Exit(2)
  		 }

		if(input == "n"){
			os.Exit(2)
		}*/

		fmt.Printf("Constructing multiDFA it = %d\n", it)
		dfa := rx.MultiDFA(subjtrees)
		//dfa = dfa.Minimize()
		h := dfa.Synthesize()
		//for _, ex := range h {
			//fmt.Printf("eg %s %s\n", ex.RXset, ex.Example)
		//}
		//fmt.Printf("size examples = %d\n", len(h))

		

		u := make([]*rx.BitSet, 0, len(h));	

		for i:=0;i<len(h); i++{
			u = append(u, h[i].RXset)
		}


		 alive := new(rx.BitSet)
  		 for i:=0; i<len(expressions); i++ { 
    			 alive.Set(expressions[i].Index);
  		 }

		//fmt.Println("alive = ", alive.String());

		/*total:=0
		max:=0
		maxWord := ""	
		for j:=0; j<len(h); j++{
		    total=0
		    for i:=0;i<len(cand); i++{
			if( cand[i].Accepts(h[j].Example) != nil ){
				total++
			}
		    }
		   if(total > max){
			max = total
			maxWord = h[j].Example
		   }
		   fmt.Println("total: ", total, " word: ", h[j].Example)
		}
		fmt.Println("max = ", max, " max word: ", maxWord)
		os.Exit(2)*/
	
		//find best question and ask it
		i := 0
		n := float64(alive.Count() / 2)
		size := math.Floor(n)
		if(alive.Count() < 4 && math.Mod(float64(alive.Count()), 2) == 1){
		    size = size+1;
      	 	 }
		cur := u[i].Count()
		c := float64(cur)
		hM := float64(math.Abs(c - size))
		//nextQn := u[i]
		word = h[i].Example
		//examInt := i

		for i := 0; i < len(u); i++ {
			c = float64(u[i].Count())
			//fmt.Println("c = ", c, "i = ", i);
			if math.Abs(c-size) == 0 {
				//nextQn = u[i]
				word = h[i].Example
				//examInt = i
				break
			} else if math.Abs(c-size) < hM {
				//fmt.Println("hM = ", hM);
				hM = math.Abs(c - size)
				//nextQn = u[i]
				word = h[i].Example
				//examInt = i
			} else { /*do nothing*/
			}
		}
		fmt.Println("before qns append")
		
		

		qns = append(qns, word)
		//fmt.Println("total= ", total)
		

	}

	/*for i:=0; i<len(qns);i++{
		fmt.Printf("%s\n", qns[i])
	}*/
	//os.Exit(2)

	rate := rateQns(qns, cand);

	   for i:=0; i<len(rate);i++{
		fmt.Println( rate[i], " dfas accepted this word ", qns[i]);
	   }

	//os.Exit(2)

	     for i:=0; i<len(qns);i++{
		
		begin:

		fmt.Println("Does your language accept this word (): ", qns[i])

		var input string
		_, e1 := fmt.Scan(&input);
  		 if (e1 != nil ){
			fmt.Printf("err = %#v\n",e1)
			os.Exit(2)
  		 }

		if input == "Yes" || input == "y" || input == "Y" || input == "yes" {
			answers.Set(i)

		} else if input == "No" || input == "N" || input == "n" || input == "no" {

		}else{
			fmt.Println("That is not a valid expression, please try again.")
			fmt.Println("Does your language accept this word: ", word)
			goto begin
		}

	    }
		fmt.Println("answers = ", answers.String())	
		
	//os.Exit(2)

	return answers, qns

}

func askBulk2(  answers *rx.BitSet, qns []string, perm []int,  cand []*rx.DFA, expns []*rx.RegExParsed, numPerGrp int )(*rx.BitSet, []string){

	numGroups := math.Floor(float64(len(cand)/numPerGrp))
	//fmt.Println("num groups = ", numGroups)
	word := ""

	if(len(cand)<numPerGrp){
		numPerGrp = len(cand);
	}

	//modTest := math.Mod(float64(len(cand)), numGroups)

	//fmt.Println("testMod = ", modTest)
	fmt.Println("length cand = ", len(cand))

	subjlist := make([]int, 0, *group)
	subjtrees := make([]rx.Node, 0, *group)
	expressions := make([]*RegEx, 0, len(cand))
	numKeep := make([] int, 0, int(numGroups))
	overlap := make([] *rx.BitSet, 0, int(numGroups))
	
	it:=0
	//for it=0; it< int(numGroups); it++{

		subjlist = make([]int, 0, *group)
		subjtrees = make([]rx.Node, 0, *group)
		expressions = make([]*RegEx, 0, len(cand))
		
		
		//create the grouping
		for j:=it*int(numPerGrp); j< (it+1)*numPerGrp; j++{
			subjlist = append(subjlist, perm[j])
			subjtrees = append(subjtrees, cand[perm[j]].Tree)
			expressions = append(expressions, &RegEx{perm[j], expns[perm[j]].Expr})
			fmt.Printf("rx { %d } %s \n", perm[j], expns[perm[j]].Expr)
		}

		//os.Exit(2)

		/*fmt.Println("continue? ")

		var input string
		_, e1 := fmt.Scan(&input);
  		 if (e1 != nil ){

			fmt.Printf("err = %#v\n",e1)
			os.Exit(2)
  		 }

		if(input == "n"){

			os.Exit(2)
		}*/

		fmt.Printf("Constructing multiDFA\n")
		dfa := rx.MultiDFA(subjtrees)
		//dfa = dfa.Minimize()
		h := dfa.Synthesize()
		for _, ex := range h {
			fmt.Printf("eg %s %s\n", ex.RXset, ex.Example)
		}
		//fmt.Printf("size examples = %d\n", len(h))

		

		u := make([]*rx.BitSet, 0, len(h));
			

		for i:=0;i<len(h); i++{
			u = append(u, h[i].RXset)
		}


		 alive := new(rx.BitSet)
  		 for i:=0; i<len(expressions); i++ { 
    			 alive.Set(expressions[i].Index);
  		 }

		//fmt.Println("alive = ", alive.Count());

		total:=0
		max :=0
		diff := float64(len(cand))
		half := math.Floor(float64(len(cand)/2));
		maxWord := ""	
		over := new(rx.BitSet)
		//maxOver := new(rx.BitSet)
		prevWord := make([]string,0,len(cand))

		for j:=0; j<len(h); j++{
		    total=0
		    over = new(rx.BitSet)
		    for i:=0;i<len(cand); i++{
			if( cand[i].Accepts(h[j].Example) != nil ){
				total++
				over.Set(i)
			}
		    }
		   if(math.Abs(float64(total)-half) < float64(diff)){
			diff = math.Abs(float64(total)-half)
			max = total
			maxWord = h[j].Example
			prevWord = append(prevWord, maxWord);
			overlap = append(overlap, over)
			fmt.Println("total: ", total, " word: ", maxWord, "max over =", over.String())
		   }
		  // fmt.Println("total: ", total, " word: ", h[j].Example)
		     //fmt.Println("max: ", max, " word: ", maxWord)
		}

		
		fmt.Println("max = ", max, " max word: ", maxWord)
		//fmt.Println("length of cand = ", len(cand),  "half = ", half)

		for i :=0; i<len(prevWord); i++ {
			fmt.Println(prevWord[i])
		}
		//os.Exit(2)
	
		//find best question and ask it
		/*i := 0
		n := float64(alive.Count() / 2)
		size := math.Floor(n)
		if(alive.Count() < 4 && math.Mod(float64(alive.Count()), 2) == 1){
		    size = size+1;
      	 	 }
		cur := u[i].Count()
		c := float64(cur)
		hM := float64(math.Abs(c - size))
		//nextQn := u[i]
		word = h[i].Example
		//examInt := i

		for i := 0; i < len(u); i++ {
			c = float64(u[i].Count())
			//fmt.Println("c = ", c, "i = ", i);
			if math.Abs(c-size) == 0 {
				//nextQn = u[i]
				word = h[i].Example
				//examInt = i
				break
			} else if math.Abs(c-size) < hM {
				//fmt.Println("hM = ", hM);
				hM = math.Abs(c - size)
				//nextQn = u[i]
				word = h[i].Example
				//examInt = i
			} else { /*do nothing*/
			//}
		//}
		//fmt.Println("before qns append")
		
		

		qns = append(qns, maxWord)
		numKeep = append(numKeep, int(max))
		//overlap = append(overlap, maxOver)
		//answers = new(rx.BitSet)
		fmt.Println("currently alive =" , len(cand), ", max to eliminate = ", max, ", maxWord =", maxWord, ", max over =", overlap[len(overlap)-1].String())
		fmt.Printf("\n")
	//os.Exit(2)	

	//}

	//os.Exit(2)
	//fmt.Printf("overlap 0: %d , overlap 1: %d \n", overlap[0].Count(), overlap[1].Count())
	//bOverlap := overlap[0].And(overlap[2])
	//fmt.Printf("intersection: %d \n", bOverlap.Count())

	//for i:=0; i<len(qns);i++{
		//fmt.Printf("%d %s %d\n", numKeep[i], qns[i], (overlap[0].And(overlap[i])).Count())

	//}
	 answers = new(rx.BitSet)
	 maybe := 0;

	     for i:=0; i<len(qns);i++{
		
		begin:

		if(maybe > 0 ){
			if(len(prevWord) == 1){
				qns[i] = h[maybe].Example
				fmt.Println("Does your language accept this word: ", h[maybe].Example)
			}else{
			    qns[i] = prevWord[len(prevWord)-maybe -1]
			    fmt.Printf("Here Does your language accept this word: %s, max over = %s\n", prevWord[len(prevWord)-maybe -1], overlap[len(overlap)-maybe -1].String())
			    //fmt.Println("max over = ", overlap[len(prevWord)-maybe -1]);
			}
		}else{
			fmt.Println("Does your language accept this word: ", qns[i])
		}

		

		var input string
		_, e1 := fmt.Scan(&input);
  		 if (e1 != nil ){
			fmt.Printf("err = %#v\n",e1)
			os.Exit(2)
  		 }

		if input == "Yes" || input == "y" || input == "Y" || input == "yes" {

			answers.Set(i)

		} else if input == "No" || input == "N" || input == "n" || input == "no" {

		}else if(input == "maybe" || input == "m" || input == "M"){
			maybe++;
			goto begin
		}else{
			fmt.Println("That is not a valid expression, please try again.")
			fmt.Println("Does your language accept this word: ", word)
			goto begin
		}

	    }
		fmt.Println("answers = ", answers.String())	
		
	//os.Exit(2)

	for i:=0; i<len(qns); i++ {
		fmt.Println(qns[i]);
	}
	//os.Exit(2)
	return answers, qns

}

func rateQns(qns []string, cand []*rx.DFA) []int {

	rate := make([]int,0,len(qns))
	total := 0
	
	for j:=0; j<len(qns); j++ {
		total=0
		for i:=0;i<len(cand); i++{
			if( cand[i].Accepts(qns[j]) != nil ){
				total++
			}
		}
		rate = append(rate, total)
		//fmt.Println()
	}

	return rate
}
