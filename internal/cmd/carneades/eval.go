package main

import (
	"flag"
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/dotout"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	"log"
	"os"
)

const helpEval = `
usage: carneades eval [-f input-format] [-t output-format] [-o output-file] [input-file]

Evaluates an argument graph and prints it in the selected output format.

If no input-file is specified, input is read from stdin. 

The -f flag ("from") specifies the format of the input file: Currently 
yaml and agxml are supported.  (default: yaml)

For further information about the agxml format, see:

	https://github.com/peldszus/arg-microtexts

The -t flag ("to") specifies the output format of the evaluated argument
graph. Currently graphml, dot, and yaml are supported. (default: graphml)	

The -o flag specifies the output file name. If the -o flag is not used, 
output goes to stdout.
`

var inputFormats = []string{"yaml", "agxml"}
var outputFormats = []string{"graphml", "yaml", "dot"}

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
	default:
		log.Fatal(fmt.Errorf("unknown or unsupported input format: %s\n", *fromFlag))
		return
	}

	// evaluate the argument graph, using grounded semantics
	// and update the labels of the statements in the argument graph
	l := ag.GroundedLabelling()
	ag.ApplyLabelling(l)

	switch *toFlag {
	case "yaml":
		yaml.Export(outFile, ag)
		outFile.Close()
	case "graphml":
		err = graphml.Export(outFile, *ag)
		outFile.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	case "dot":
		err = dotout.Export(outFile, *ag)
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
