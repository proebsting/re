BUILDING THE RX CODE

Everything here is written in the Go programming language.
Download Go (and read its documentation) from golang.org.

Go requires that the environment variable GOPATH be set.
This source directory should be located two levels below that,
at $GOPATH/src/rx.

The command-level programs are found in the rx? subdirectories
and are documented in their comment headers.  To build then, run
	make build
which will produce executable binaries in $GOPATH/bin.

The default Makefile target is oriented towards development.
It runs the above build followed by a series of automated tests.
Finally it runs the "rxr" program on a local file "expt.rx"
if this file exists.
