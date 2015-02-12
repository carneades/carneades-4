
# Instructions for downloading, building and installing the Go vesion
  of the Carneades argumentation system.

## Prerequisites

- The (Go programming language)[http://golang.org/] compiler and other
  tools, available for Linux, Mac OS, Windows, and some other operating
  systems.

- The (Git)[http://git-scm.com/] distributed version control system.

## GOPATH

Be sure that you have set the `GOPATH` environment variable, as
explained in <https://golang.org/doc/code.html#GOPATH>.

Add `$GOPATH/bin` to your `PATH` environment variable. 

## Downloading 

    $ go get github.com/carneades/carneades-go

## Installation

    $ cd install github.com/carneades/carneades-go

## Running

The Carneades commands (executable files) should have been installed
in the `bin` directory in your `GOPATH` environment variable. If you
have added the `$GOPATH/bin` to your `PATH` environment variable, as
suggested above, you can now execute the Carneades commands from the
command line, without typing their full path names.  For example:

    $ carneades-iccma







