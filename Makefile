#  Makefile for rx programs

#D='(a|b)*abb' 'b(ab)*a'
D='\d+' '\d*[1-9]' '[1-9]\d*'

PKG = rx
PROGS = $(PKG)/rxd $(PKG)/rxq $(PKG)/rxr $(PKG)/rxx $(PKG)/rxg
GOBIN = $$GOPATH/bin

default: build test expt

fmt:
	go fmt *.go
	go fmt rxd/rxd.go
	go fmt rxq/rxq.go
	go fmt rxr/rxr.go
	go fmt rxx/rxx.go
	go fmt rxg/rxg.go

build:
	go install $(PROGS)

test:	build
	cd test; $(MAKE)

demo:	
	rxd $D >tmp.dot
	dot -Tpdf tmp.dot >tmp.pdf
	display tmp.pdf

#  partial build and test (list not complete)
rxg:	.FORCE; go install ${PKG}/rxg; cd test; runtest.sh *.rxg
.FORCE:

#  if expt.rx exists, run with rxr after standard build and test
expt:
	test -f expt.rx && $(GOBIN)/rxr expt.rx || :

clean:
	go clean -i $(PKG) $(PROGS)
	cd test; $(MAKE) clean
