#  Makefile for RX library and programs

PKG = rx
PG1 = $(PKG)/rxplor $(PKG)/rxpick $(PKG)/rxquest $(PKG)/rxcluster
PG2 = $(PKG)/rxtime $(PKG)/rxg $(PKG)/rxq $(PKG)/rxx
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

#  "make expt" runs "rxplor -a expt.rx" if expt.rx exists.
#  This allows adding a quick temporary test to the build process.
expt:
	test -f expt.rx && $(GOBIN)/rxplor -a expt.rx || :

#  "make bundle" combines all sources into a single file on standard output.
#  This requires the Kernighan and Pike "bundle" utility in the path.
bundle:
	@bundle *.go */*.go webapp/code/*go

#  "make fmt" formats all the source files to Go standards
#  This should be done before checking in any code.
#  If "go fmt" echoes a filename, it has modified that file.
fmt:
	go fmt *.go
	go fmt rxsys/*.go
	go fmt rxplor/rxplor.go
	go fmt rxpick/rxpick.go
	go fmt rxquest/rxquest.go
	go fmt rxcluster/rxcluster.go
	go fmt rxtime/rxtime.go
	go fmt rxg/rxg.go
	go fmt rxq/rxq.go
	go fmt rxx/rxx.go
	go fmt webapp/code/*.go

#  "make demo" displays a graph of the DFA of the exprs defined above.
demo:	
	rxplor -D @ -e $(DEMO)

#  "make serve" builds and runs the web app on localhost:8080.
#  The server runs until killed.
serve:
	goapp serve webapp

#  "make deploy" uploads the web app to appspot.com.
deploy:
	goapp deploy webapp


#  "make clean" removes the products of building and testing.
clean:
	go clean -i $(PKG) $(PROGS)
	cd test; $(MAKE) clean


.FORCE:		# target meaning "always run"
