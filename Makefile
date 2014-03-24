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
	go test
	cd test; $(MAKE)

demo:	
	rxd $D >tmp.dot
	dot -Tpdf tmp.dot >tmp.pdf
	display tmp.pdf

#  partial build and test (imperfect but useful)
rxd:	.FORCE; go install ${PKG}/rxd && cd test && runtest.sh *.rxd
rxg:	.FORCE; go install ${PKG}/rxg && cd test && runtest.sh *.rxg
rxq:	.FORCE; go install ${PKG}/rxq && cd test && runtest.sh *.rxq
rxr:	.FORCE; go install ${PKG}/rxr && cd test && runtest.sh *.rx
rxv:	.FORCE; go install ${PKG}/rxr && cd test && runtest.sh *.rxv
rxx:	.FORCE; go install ${PKG}/rxx && cd test && runtest.sh *.rxx
.FORCE:

#  if expt.rx exists, run with rxr after standard build and test
expt:
	test -f expt.rx && $(GOBIN)/rxr expt.rx || :

clean:
	go clean -i $(PKG) $(PROGS)
	cd test; $(MAKE) clean
