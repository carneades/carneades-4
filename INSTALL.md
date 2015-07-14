
These are instructions for building and running Carneades-4.  

## Prerequisites

A binary distribution for Microsoft Windows (64 bit) is available on the [releases](https://github.com/carneades/carneades-4/releases) page. No further software needs to be installed to use this binary distribution.

For building the system from the source files, the following are required:

- Version 1.4x or newer of the [Go programming language](http://golang.org/) compiler suite.
- [Git](http://git-scm.com/).

Set the `GOPATH` environment variable to a directory for Go packages, e.g.

    $ mkdir ~/go
    $ typeset -x GOPATH=~/go

## Building and Running Carneades from Source

Use the `go` tool to get, build and install the Carneades
executable from Github:

    $ go get github.com/carneades/carneades-4/internal/cmd/carneades
    
The `carneades` executable should now be installed in

    $GOPATH/bin/carneades

You can execute the program using this full path. Alternatively, add `$GOPATH/bin` to your `PATH` environment.
You should then be able to execute the command directly, as in

    $ carneades eval -o tandem.graphml tandem.yml
    
For further information about how to use the `carneades` command, type

    $ carneades help

Example argument graphs, in YAML format, can be found in the `$GOPATH/src/github.com/carneades/carneades-4/examples/AGs` directory.

## Building and Running the Carneades ICCMA Entry

This version of Carneades has been entered in the [ICCMA](http://argumentationcompetition.org/index.html)
competition.

To build and install the Carneades ICCMA entry, use the go tool:

    $ go get github.com/carneades/carneades-4/internal/cmd/carneades-iccma

The carneades-iccma executable should now be installed in

    $GOPATH/bin/carneades-iccma

You can execute the program using this full path. If you have added
$GOPATH/bin to your PATH environment, you can then execute the command
directly, as in

    $ carneades-iccma -p EE-GR -f `$GOPATH/src/github.com/carneades/carneades-4/internal/test/AFs/bas1.tgf

See the [ICCMA supplementary notes](http://argumentationcompetition.org/2015/iccma15notes_v3.pdf) for further instructions about the flags and parameters, which are the same for all ICCMA entries.

Example abstract argumentation frameworks can be found in the ``$GOPATH/src/github.com/carneades/carneades-4/examples/AFs` directory.

Dung abstract argumentation frameworks can also be evaluated and visualized using the `carneades` command. For instructions, execute:

    $ carneades help dung

	


