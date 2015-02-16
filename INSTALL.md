
There are instructions for building and running the Carneades entry to
the [ICCMA](http://argumentationcompetition.org/index.html)
competition.

## Prerequisites

- The [Go programming language](http://golang.org/) compiler
- [Git](http://git-scm.com/)

## Building the carneades-iccma executable

Set the GOPATH environment variable to a directory for Go packages, e.g.

    $ mkdir ~/go
	$ typeset -x GOPATH=~/go

Use the go tool to get the carneades-go package from Github:

    $ go get github.com/carneades/carneades-go

Use the go tool to build and install the carneades-iccma executable:
 
    $ go install github.com/carneades/carneades-go/src/cmd/carneades-iccma
	
## Running the carneades-iccma executable

The carneades-iccma executable should now be installed in

    $GOPATH/bin/carneades-iccma

You can execute the program using this full path. Altneratively, add
$GOPATH/bin to your PATH environment. You can then execute the command
directly, as in

    $ carneades-iccma -p EE-GR -f ...


	









