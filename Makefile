#  Makefile for rx programs

PKG = rx
PROGS = $(PKG)/rxd $(PKG)/rxq $(PKG)/rxr $(PKG)/rxx
GOBIN = $$GOPATH/bin

default: build test expt

fmt:
	go fmt *.go
	go fmt rxd/rxd.go
	go fmt rxq/rxq.go
	go fmt rxr/rxr.go
	go fmt rxx/rxx.go

build:
	go install $(PROGS)

test:	build
	cd test; $(MAKE)

demo:	
	rxd '(a|b)*abb' 'b(ab)*a' >tmp.dot
	dot -Tgif tmp.dot >tmp.gif
	display tmp.gif

#  if expt.rx exists, run with rxr after standard build and test
expt:
	test -f expt.rx && $(GOBIN)/rxr expt.rx || :

clean:
	go clean -i $(PKG) $(PROGS)
	cd test; $(MAKE) clean
