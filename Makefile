#  Makefile for RX library and programs

PKG = rx
PG1 = $(PKG)/rxd $(PKG)/rxg $(PKG)/rxq $(PKG)/rxr $(PKG)/rxx $(PKG)/rxplor
PG2 = $(PKG)/rxcluster $(PKG)/rxcombos $(PKG)/rxpick $(PKG)/questions
PROGS = $(PG1) $(PG2)
GOBIN = $$GOPATH/bin

DEMO='-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d{1,3})?'


#  The default is to rebuild, run all tests, run expt if present.
default: build test expt

#  "make build" compiles all programs (and the library).
build:  .FORCE
	go install $(PROGS)

#  "make test" runs unit tests and the shell-based tests
test:	.FORCE
	go test
	cd test; $(MAKE)

#  "make expt" runs "rxr expt.rx" if expt.rx exists.
#  This allows adding a quick temporary test to the build process.
expt:
	test -f expt.rx && $(GOBIN)/rxr expt.rx || :

#  "make bundle" combines all sources into a single file on standard output.
#  This requires the Kernighan and Pike "bundle" utility in the path.
bundle:
	@bundle *.go */*.go

#  "make fmt" formats all the source files to Go standards
#  This should be done before checking in any code.
#  If "go fmt" echoes a filename, it has modified that file.
fmt:
	go fmt *.go
	go fmt rsys/*.go
	go fmt rxcluster/rxcluster.go
	go fmt rxcombos/rxcombos.go
	go fmt rxd/rxd.go
	go fmt rxg/rxg.go
	go fmt rxpick/rxpick.go
	go fmt rxplor/rxplor.go
	go fmt rxq/rxq.go
	go fmt rxr/rxr.go
	go fmt rxx/rxx.go
	go fmt questions/questions.go
	go fmt webapp/code/*.go

#  "make demo" displays a graph of the DFA of the exprs defined above.
demo:	
	rxplor -D - -e $(DEMO)

#  "make clean" removes the products of building and testing.
clean:
	go clean -i $(PKG) $(PROGS)
	cd test; $(MAKE) clean


.FORCE:		# target meaning "always run"
