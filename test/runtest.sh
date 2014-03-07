#!/bin/sh
#
#  runtest [basename[.std]]... - run tests and compare results
#
#  The utility tested is determined by the input extension found.
#	.rx or .rxr -- run "rxr -Q -R"
#	.rxv -- run "rxr -R"
#	.rxq -- run "rxq pattern", using first line as the query pattern
#	.rxx -- run "rxx basename.rxx basename.rcx"
#	.rxd -- run "rxd `<basename.rxd`"

BIN=$GOPATH/bin

# if no test files are specified, run them all
if [ $# = 0 ]; then
    set - *.std
fi

# loop through the chosen tests
echo ""
FAILED=
for F in $*; do
    F=${F%.std}
    F=${F%.rx*}
    I="`echo $F.rx*`"
    EXTN="${I#*.}"
    echo "Testing $F.$EXTN"
    case ".$EXTN" in 
	.*.*)
	    echo 1>&2 "multiple input files: $I"
	    FAILED="$FAILED $I"
	    continue;;
	.rx|.rxr)
	    $BIN/rxr -Q -R $I >$F.out;;
	.rxv)
	    $BIN/rxr -R $I >$F.out;;
	.rxq)
	    REXPR=`sed 1q $I`
	    cat $I | sed 1d | rxq "$REXPR" - >$F.out;;
	.rxx)
	    $BIN/rxx $I ${I%.rxx}.rcx >$F.out;;
	.rxd)
	    $BIN/rxd `cat $I` >$F.out;;
	.*)
	    echo 1>&2 "unrecognized extension: $I"
	    FAILED="$FAILED $I"
	    continue;;
    esac
    if [ $? -eq 0 ] && cmp $F.std $F.out; then
    	rm $F.out
    else
	diff -u $F.std $F.out | sed 18q
	echo ------------------------------------------------------------------
        FAILED="$FAILED $F"
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
