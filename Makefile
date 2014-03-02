#  Makefile for rx programs

PKG = rx
PROGS = $(PKG)/rxq $(PKG)/rxr
GOBIN = $$GOPATH/bin

default: build test expt

build:
	go install $(PROGS)

test:	build
	cd test; $(MAKE)

#  if expt.rx exists, run with rxr after standard build and test
expt:
	test -f expt.rx && $(GOBIN)/rxr expt.rx || :

clean:
	go clean -i $(PKG) $(PROGS)
	cd test; $(MAKE) clean
