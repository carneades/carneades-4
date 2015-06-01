
These are instructions for building and running the Carneades entry to
the [ICCMA](http://argumentationcompetition.org/index.html)
competition.

## Prerequisites

- The [Go programming language](http://golang.org/) compiler suite
- [Git](http://git-scm.com/)

## Building the carneades-iccma executable

Set the GOPATH environment variable to a directory for Go packages, e.g.

    $ mkdir ~/go
	$ typeset -x GOPATH=~/go

Use the go tool to get, build and install the carneades-iccma
executable from Github:

    $ go get github.com/carneades/carneades-4/internal/cmd/carneades-iccma

(Or use the "build" script included in the distribution, which runs
this command.)

## Running the carneades-iccma executable

The carneades-iccma executable should now be installed in

    $GOPATH/bin/carneades-iccma

You can execute the program using this full path. Alternatively, add
$GOPATH/bin to your PATH environment. You can then execute the command
directly, as in

    $ carneades-iccma -p EE-GR -f ...


	









