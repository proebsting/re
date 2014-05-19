#!/bin/sh
#
#  runtest [file.rx...] - run tests and validate outputs
#
#  For each .rx file, this script runs rxplor (or another program) and
#  compares the output with the corresponding .std file.
#
#  The command "rxplor -T" is run by default.  If the .rx file begins 
#  with a #! line, the command from that line is run instead.
#  If the file begins with multiple #! lines, the subsequent commands
#  are executed and their outputs compared against .std2, .std3, etc.


# function definition
runtest() {	 # runtest basename n command -- run one test and check output
    B=$1
    N=$2
    C=$3
    I=$B.rx
    O=$B.out${N%1}
    S=$B.std${N%1}

    printf "%-16s %s\n" "$B.$N:" "$C"
    eval "$C" <$I >$O
    if [ $? -eq 0 ] && cmp $S $O; then
    	rm $O
    else
	diff -u $S $O | sed 18q
	echo ------------------------------------------------------------------
	FAILED="$FAILED $B.$N"
    fi
}


# if no test files are specified, run them all
if [ $# = 0 ]; then
    set - `ls *.rx`
fi

# loop through the chosen tests
PATH=$GOPATH/bin:$PATH
echo ""
FAILED=
for F in $*; do
    B=${F%.*}
    I=$B.rx
    N=0
    exec <$I
    while read LINE; do
    	case "$LINE" in
	    "#!"*)
		N=$(($N+1))
		CMD=${LINE#??}
		runtest $B $N "$CMD"
		;;
	    ""|*)
		break;;
	esac
    done
    if [ $N = 0 ]; then
	runtest $B 1 " rxplor -T"
    fi
done

#  summarize the results
echo ""
if [ "$FAILED" = "" ]; then
    echo "All tests passed."
    echo ""
    exit 0
else
    echo "Tests failed: $FAILED"
    echo ""
    exit 1
fi
