#  Makefile for rxr tests

default:  test

test:
	runtest.sh
	rxr -Z 18-basic.rx >18-basic.out && cmp 18-basic.stdA 18-basic.out
	:
	:

clean:
	rm -f *.tmp *.out
