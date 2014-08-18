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
runtest() {	# runtest basename n command -- run one test and check output
    B=$1		# basename
    N=$2		# subtest number
    C=$3		# program command
    I=$B.rx		# input file
    O=$B.out${N%1}	# output file
    S=$B.std${N%1}	# standard file for comparison

    printf "%-16s %s\n" "$B.$N:" "$C"
    eval "$C" <$I >$O	# run the command
    if [ $? -eq 0 ] && cmp $S $O; then
	rm $O		# normal exit, files match, so remove output file
    else
	diff -u $S $O | sed 18q
	echo ------------------------------------------------------------------
	FAILED="$FAILED $B.$N"
    fi
}


# if no test files are specified, run them all
if [ $# = 0 ]; then
    set - `ls *.rx`			# add all .rx files as command args
fi

# loop through the chosen tests
PATH=$GOPATH/bin:$PATH
unset RX_COMPLEXITY
echo ""
FAILED=
for F in $*; do				# for each file on command line
    B=${F%.*}	# basename
    I=$B.rx	# input file
    N=0		# number of tests run
    exec <$I	# redirect stdin 
    while read LINE; do
	case "$LINE" in	
	    "#!"*)			# this is a test spec
		N=$(($N+1))		# count it
		CMD=${LINE#??}		# extract the command
		runtest $B $N "$CMD"	# run it as a test
		;;
	    ""|*)
		break;;
	esac
    done
    if [ $N = 0 ]; then			# if no explicit test named in file
	runtest $B 1 " rxplor -T"	# then run rxplor -T
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
