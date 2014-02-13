#  makefile for rx programs

PROG=rxr
BIN=$$GOPATH/bin/$(PROG)

default: build test expt

build:
	go install rx/$(PROG)

test:	build
	cd test; $(MAKE)

#  if expt.rx exists, run after standard build and test
expt:
	test -f expt.rx && $(BIN) expt.rx || :

clean:
	rm -f $(BIN)
	cd test; $(MAKE) clean
