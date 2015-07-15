package main

import (
	"flag"
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/aif"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/lkif"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	"log"
	"os"
)

const helpEval = `
usage: carneades eval [-f input-format] [-t output-format] [-o output-file] [input-file]

Evaluates an argument graph and prints it in the selected output format.

If no input-file is specified, input is read from stdin. 

The -f flag ("from") specifies the format of the input file: Currently 
yaml, aif, agxml and lkif are supported. (default: yaml)

The "yaml" format is the native format of Carneades 4.x.  YAML is a schemeless
data interchange format. See http://yaml.org/ for general information about
YAML.  For examples of Carneades argument graphs in this format, see

    https://github.com/carneades/carneades-4/tree/master/internal/test/AGs

The "aif" format is the JSON serialization of the Argument Interchange 
Format (AIF).  For further information about AIF, see:

    http://www.argumentinterchange.org/developers
    http://www.arg-tech.org/index.php/projects/

The "agxml" format is the XML schema used by the
annotated corpus of argumentative microtexts Github project.
For further information about the agxml XML schema, see:

    https://github.com/peldszus/arg-microtext

The "lkif" format is the Legal Knowledge Interchange Format
(LKIF) XML schema. Only the first argument graph in the LKIF file
is translated. For more information about LKIF, see:
Gordon, T. F. The Legal Knowledge Interchange Format (LKIF).
Deliverable 4.1, European Commission, 2008.
http://www.tfgordon.de/publications/files/GordonLKIF2008.pdf

LKIF is the native format of Carneades 2, a desktop argument mapping
tool with a graphical user interface. Also known as the Carneades Editor.
https://github.com/carneades/carneades-2

The -t flag ("to") specifies the output format of the evaluated argument
graph. Currently graphml and yaml are supported. (default: graphml)	

GraphML is an XML schema for directed graphs.  Graphml is supported by 
several graph editors and visualizations tools.  For further information see:

    http://graphml.graphdrawing.org/

The graphml produced by Carneades is intended for use with the yEd
graph editor. For more information about yEd, see:

    https://www.yworks.com/en/products/yfiles/yed/

The -o flag specifies the output file name. If the -o flag is not used, 
output goes to stdout.
`

var inputFormats = []string{"yaml", "aif", "agxml", "lkif"}
var outputFormats = []string{"graphml", "yaml"}

func contains(l []string, s1 string) bool {
	for _, s2 := range l {
		if s1 == s2 {
			return true
		}
	}
	return false
}

func evalCmd() {
	eval := flag.NewFlagSet("eval", flag.ContinueOnError)
	fromFlag := eval.String("f", "yaml", "the format of the source file")
	toFlag := eval.String("t", "graphml", "the format of the output file")
	outFileFlag := eval.String("o", "", "the filename of the output file")

	var inFile *os.File
	var outFile *os.File
	var err error

	if err := eval.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	if !contains(inputFormats, *fromFlag) {
		log.Fatal(fmt.Errorf("unsupported input format: %s\n", *fromFlag))
		return
	}
	if !contains(outputFormats, *toFlag) {
		log.Fatal(fmt.Errorf("unsupported output format: %s\n", *toFlag))
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
	if *outFileFlag == "" {
		outFile = os.Stdout
	} else {
		outFile, err = os.Create(*outFileFlag)
		if err != nil {
			log.Fatal(fmt.Errorf("%s\n", err))
			return
		}
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
	default:
		log.Fatal(fmt.Errorf("unknown or unsupported input format: %s\n", *fromFlag))
		return
	}

	// evaluate the argument graph, using grounded semantics
	// and update the labels of the statements in the argument graph
	l := ag.GroundedLabelling()
	fmt.Printf("labelling=%v\n", l)
	ag.ApplyLabelling(l)

	switch *toFlag {
	case "yaml":
		yaml.Export(outFile, ag)
		outFile.Close()
	case "graphml":
		err = graphml.Export(outFile, ag)
		outFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	default:
		log.Fatal(fmt.Errorf("unknown or unsupported output format: %s\n", *toFlag))
		return
	}
}
