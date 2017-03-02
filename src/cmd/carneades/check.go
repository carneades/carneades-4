package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/aif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/caf"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/lkif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

const helpCheck = `
usage: carneades check [-f input-format] [input-file]

Checks an argument graph for syntactic and semantic errors
and prints out error messages to the standard output.

If no input-file is specified, input is read from stdin. 

The -f flag ("from") specifies the format of the input file: Currently 
yaml, aif, agxml and lkif are supported. (default: yaml)

The "yaml" format is the native format of Carneades 4.x.  YAML is a schemeless
data interchange format. See http://yaml.org/ for general information about
YAML.  For examples of Carneades argument graphs in this format, see

    https://github.com/carneades/carneades-4/examples/AGs/YAML

The "aif" format is the JSON serialization of the Argument Interchange 
Format (AIF).  For further information about AIF, see:

    http://www.argumentinterchange.org/developers
    http://www.arg-tech.org/index.php/projects/

The "agxml" format is the XML schema used by the
annotated corpus of argumentative microtexts Github project.
For further information about the agxml XML schema, see:

    https://github.com/peldszus/arg-microtext

The "caf" format is the Carneades Argument Format (CAF), the native format of 
Carneades 3. For further information about the CAF and Carneades 3, see::

	https://github.com/carneades/carneades-3/blob/master/schemas/CAF.rnc
	https://github.com/carneades/carneades-3
	
The "lkif" format is the Legal Knowledge Interchange Format
(LKIF) XML schema. Only the first argument graph in the LKIF file
is translated. For more information about LKIF, see:
Gordon, T. F. The Legal Knowledge Interchange Format (LKIF).
Deliverable 4.1, European Commission, 2008.
http://www.tfgordon.de/publications/files/GordonLKIF2008.pdf

LKIF is the native format of Carneades 2, a desktop argument mapping
tool with a graphical user interface. Also known as the Carneades Editor.
https://github.com/carneades/carneades-2
`

func checkCmd() {
	eval := flag.NewFlagSet("eval", flag.ContinueOnError)
	fromFlag := eval.String("f", "yaml", "the format of the source file")

	var inFile *os.File
	var outFile *os.File = os.Stderr
	var err error

	if err := eval.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	if !contains(inputFormats, *fromFlag) {
		log.Fatal(fmt.Errorf("unsupported input format: %s\n", *fromFlag))
		return
	}

	switch eval.NArg() {
	case 0:
		inFile = os.Stdin
	case 1:
		inFile, err = os.Open(eval.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(fmt.Errorf("incorrect number of arguments after the command flags; should be 0, to read from stdin, or 1, naming the input file\n"))
		return
	}

	var ag *caes.ArgGraph

	switch *fromFlag {
	case "yaml":
		ag, err = yaml.Import(inFile)
		inFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	case "agxml":
		ag, err = agxml.Import(inFile)
		inFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	case "aif":
		ag, err = aif.Import(inFile)
		inFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	case "lkif":
		ag, err = lkif.Import(inFile)
		inFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	case "caf":
		ag, err = caf.Import(inFile)
		inFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	default:
		log.Fatal(fmt.Errorf("unknown or unsupported input format: %s\n", *fromFlag))
		return
	}

	// Validate the argument graph
	problems := validation.Validate(ag)

	// Print out any problems found to standard out
	for _, p := range problems {
		if p.Expression == "" {
			fmt.Fprintf(outFile, "%s: %s: %s\n", p.Category, p.Id, p.Description)
		} else {
			fmt.Fprintf(outFile, "%s: %s: %s: %s\n", p.Category, p.Id, p.Description, p.Expression)
		}
	}
}
