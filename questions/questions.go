package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"rx"
)

type RegEx struct {
	Index int
	Rexpr string
}

type DFAexample struct {
	State   uint
	RXSet   []uint
	Example string
}

type Jdata struct {
	Expressions []*RegEx
	Examples    []*DFAexample
}

func question(u []*rx.BitSet, h []*DFAexample, e []*RegEx, alive *rx.BitSet, track int) ([]*rx.BitSet, []*DFAexample) {

	t := make([]*rx.BitSet, 0, len(u))
	k := make([]*DFAexample, 0, len(h))

	if alive.Count() == 0 {
		fmt.Println("No reg ex in our library matches your query.")
		return t, k
	}

	if track == 1 {
		if alive.Count() == 1 {
			fmt.Println("The reg ex you are looking for is: ", e[alive.Lowbit()].Rexpr)
			return t, k
		}
	}
	if track == 0 && alive.Count() == 1 {
		fmt.Println(alive.Count())
		if len(u) == 0 {
			fmt.Println("No reg ex in our library matches your query.")
			return t, k
		} else {
			goto askAgain
		}

	askAgain:
		fmt.Println("Does your language accept this word: ", h[0].Example)
		var input1 string
		fmt.Scan(&input1)

		if input1 == "Yes" || input1 == "y" || input1 == "Y" || input1 == "yes" {
			track++
			fmt.Println("The reg ex you are looking for is: ", e[alive.Lowbit()].Rexpr)
			return t, k
		} else if input1 == "No" || input1 == "N" || input1 == "n" || input1 == "no" {
			fmt.Println("No reg ex in our library matches your query.")
			return t, k
		} else {
			fmt.Println("That is not a valid expression, please try again.")
			goto askAgain
		}

	}

	//Iterate through u to find the rx.BitSet with numleft/2 bits set to 1
	i := 0
	n := float64(alive.Count() / 2)
	size := math.Floor(n)
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

	fmt.Println("Does your language accept this word: ", word)

begin:
	var input string
	fmt.Scan(&input)

	if input == "Yes" || input == "y" || input == "Y" || input == "yes" {
		track++
		if nextQn.Count() == 1 {
			t = append(t, nextQn)
			fmt.Println("The reg ex you are looking for is: ", e[nextQn.Lowbit()].Rexpr)
			k = append(k, h[examInt])
			return t, k
		}
		alive = alive.And(nextQn)
		for i := 0; i < len(u); i++ {
			if !((nextQn.And(u[i])).IsEmpty()) {
				t = append(t, u[i])
				k = append(k, h[i])
			}
		}

		fmt.Println("alive =", alive.String())

	} else if input == "No" || input == "N" || input == "n" || input == "no" {

		for i := 0; i < len(u); i++ {
			if (nextQn.And(u[i])).IsEmpty() {
				t = append(t, u[i])
				k = append(k, h[i])
			}
		}

		bs := new(rx.BitSet)
		for i := 0; i < cap(e); i++ {
			if nextQn.Test(uint(i)) {

			} else {
				bs.Set(uint(i))
			}
		}
		alive = bs.And(alive)

		fmt.Println("alive = ", alive.String())

	} else {
		fmt.Println("That is not a valid expression, please try again.")
		fmt.Println("Does your language accept this word: ", word)
		goto begin
	}

	return question(t, k, e, alive, track)
}

func main() {

	/*
	  {"Examples":[{"State":1,"RXset":[0],"Example":"one"},
	  {"State":2,"RXset":[0,1],"Example":"two"},
	  {"State":4,"RXset":[1],"Example":"four"},
	  {"State":5,"RXset":[2],"Example":"five"},
	  {"State":6,"RXset":[2,3],"Example":"six"},
	  {"State":7,"RXset":[2,3,4],"Example":"seven"},
	  {"State":8,"RXset":[3],"Example":"eight"},
	  {"State":9,"RXset":[2,3,5,6],"Example":"nine"},
	  {"State":10,"RXset":[2,3,5],"Example":"ten"},
	  {"State":11,"RXset":[3,6],"Example":"eleven"},
	  {"State":12,"RXset":[6],"Example":"twelve"},
	  {"State":13,"RXset":[5],"Example":"thirteen"},
	  {"State":14,"RXset":[3,5],"Example":"fourteen"},
	  {"State":15,"RXset":[6,8],"Example":"fifteen"},
	  {"State":16,"RXset":[9],"Example":"sixteen"},
	  {"State":17,"RXset":[8,9],"Example":"seventeen"},
	  {"State":18,"RXset":[7],"Example":"eighteen"},
	  {"State":19,"RXset":[8],"Example":"nineteen"}],
	  "Expressions":[{"Index":0, "Rexpr":"regex0"},
	  {"Index":1, "Rexpr":"regex1"},
	  {"Index":2, "Rexpr":"regex2"},
	  {"Index":3, "Rexpr":"regex3"},
	  {"Index":4, "Rexpr":"regex4"},
	  {"Index":5, "Rexpr":"regex5"},
	  {"Index":6, "Rexpr":"regex6"},
	  {"Index":7, "Rexpr":"regex7"},
	  {"Index":8, "Rexpr":"regex8"},
	  {"Index":9, "Rexpr":"regex9"}]}


	  {"Examples":[{"State":1,"RXset":[5],"Example":"one"},
	  {"State":2,"RXset":[4,5],"Example":"two"},
	  {"State":3,"RXset":[0,4,5],"Example":"three"},
	  {"State":4,"RXset":[4],"Example":"four"},
	  {"State":5,"RXset":[3],"Example":"five"},
	  {"State":6,"RXset":[3,4],"Example":"six"},
	  {"State":7,"RXset":[3,5],"Example":"seven"},
	  {"State":8,"RXset":[2],"Example":"eight"},
	  {"State":9,"RXset":[1,3,4],"Example":"nine"},
	  {"State":10,"RXset":[1,2,3,4],"Example":"ten"},
	  {"State":11,"RXset":[1,2],"Example":"eleven"},
	  {"State":12,"RXset":[1],"Example":"twelve"}],
	  "Expressions":[{"Index":0, "Rexpr":"regex0"},
	  {"Index":1, "Rexpr":"regex1"},
	  {"Index":2, "Rexpr":"regex2"},
	  {"Index":3, "Rexpr":"regex3"},
	  {"Index":4, "Rexpr":"regex4"},
	  {"Index":5, "Rexpr":"regex5"}]}
	*/

	var testExp Jdata

	dec := json.NewDecoder(os.Stdin)
	jerr := dec.Decode(&testExp)
	fmt.Printf("jerr = %#v\n", jerr)

	//Use Examples to make BitSets for each region
	start := make([]*rx.BitSet, 0, len(testExp.Examples))
	for it := 0; it < len(testExp.Examples); it++ {
		elem := new(rx.BitSet)
		for nt := 0; nt < len(testExp.Examples[it].RXSet); nt++ {
			elem.Set(uint(testExp.Examples[it].RXSet[nt]))
		}
		start = append(start, elem)
	}

	fmt.Println("Start")
	for it := 0; it < len(start); it++ {
		fmt.Println(start[it].String())
	}

	//Make a bit set for regular expressions that are still alive, initially set to all 1's
	alive := new(rx.BitSet)
	for it := 0; it < len(testExp.Expressions); it++ {
		alive.Set(uint(it))
	}
	// fmt.Println(alive.String());

	v := make([]*rx.BitSet, 0, len(testExp.Examples))
	v2 := make([]*DFAexample, 0, len(testExp.Examples))

	//run 20 questions
	v, v2 = question(start, testExp.Examples, testExp.Expressions, alive, 0)

	if len(v) < 0 {
		fmt.Println(v[0].String())
		fmt.Println(v2[0].State)
	}

}

func TestStdIn(i int) {
	fmt.Printf("%d\n", i)
	var jd Jdata
	dec := json.NewDecoder(os.Stdin)
	err := dec.Decode(&jd)
	fmt.Printf("err = %#v\n", err)
	for _, v := range jd.Expressions {
		fmt.Printf("%#v\n", v)
	}
	for _, v := range jd.Examples {
		fmt.Printf("%#v\n", v)
	}
}
