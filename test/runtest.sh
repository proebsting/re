#!/bin/sh
#
#  runtest [basename[.std]]... - run tests and compare results

RXR="../../../bin/rxr -Z"

# if no test files specified, run them all
if [ $# = 0 ]; then
    set - *.std
fi


# loop through the chosen tests
echo ""
FAILED=
for F in $*; do
    F=`basename $F .std`
    F=`basename $F .rx`
    rm -f $F.out
    echo "Testing $F"
    if $RXR $F.rx >$F.out && cmp $F.std $F.out; then
    	rm $F.out
    else
	diff -u $F.std $F.out | sed 18q
	echo ------------------------------------------------------------------
        FAILED="$FAILED $F"
    fi
done

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
