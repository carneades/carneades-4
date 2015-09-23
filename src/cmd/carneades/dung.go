// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"github.com/carneades/carneades-4/src/engine/dung"
	"github.com/carneades/carneades-4/src/engine/dung/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/dung/encoding/tgf"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const helpDung = `
usage: carneades dung [-f input-format] [-p problem] [-s semantics] [-a argument] [-o output-directory] [input-file]

Evaluates a Dung abstract argumentation framework and prints its extensions
to stdout and, optionally, to a directory of graphml files for visualizing the extensions.

If no input-file is specified, input is read from stdin. 

The -f flag ("from") specifies the format of the input file. Currently only 
the Trivial Graph Format (tgf) is supported.  (default: tgf)

The -p flag specifies the decision problem to be solved, which must be one
of DC, DS, EE or SE, where

- DC: Decide whether the given argument is credulously inferred.
- DS: Decide whether the given argument is skeptically inferred.
- EE: Enumerate all extensions of the framework.
- SE: Show one extension of the framework, if any exist, or print "NO" if the
      framework has no extensions.

The default problem is EE, enumerating all extensions.

The -s flag specifies the Dung semantics to use, which must be one of
GR, SO, PR, ST, where

- GR: Grounded semantics
- CO: Complete semantics
- PR: Preferred semantics
- ST: Stable semantics

The default is GR, grounded semantics.

The -a flag specifies the argument to check when solving DC and DS problems.
It should be the id of the argument in the input file.  default: none.

If the -o flag is specified, graphml files are written to the given directory.
The yEd Graphml editor can be used to view the evaluated argumentation framework.
Existing directories will not be overwritten or modified. A file for each 
extension of the argumentation framework will be created.
`

const formats = "[tgf]"
const problems = "[DC-GR,DS-GR,EE-GR,SE-GR,DC-PR,DS-PR,EE-PR,SE-PR,DC-CO,DS-CO,EE-CO,SE-CO,DC-ST,DS-ST,EE-ST,SE-ST]"

func dungCmd() {
	dungFlags := flag.NewFlagSet("dung", flag.ContinueOnError)
	problemFlag := dungFlags.String("p", "EE", "the problem to solve")
	semanticsFlag := dungFlags.String("s", "GR", "the semantics to use")
	formatFlag := dungFlags.String("f", "tgf", "the format of the source file")
	argFlag := dungFlags.String("a", "", "the id of the argument to check")
	outputFlag := dungFlags.String("o", "", "the name of a new directory to create for outputting GraphML files")

	if err := dungFlags.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	arg := *argFlag

	checkArgFlag := func() {
		if arg == "" {
			log.Fatal(fmt.Errorf("no -a flag"))
		}
	}

	if *formatFlag != "tgf" {
		log.Fatal(fmt.Errorf("unsupported format: %s\n", *formatFlag))
		return
	}

	var inFile *os.File
	var err error
	var af dung.AF
	var extensions []dung.ArgSet

	switch dungFlags.NArg() {
	case 0:
		inFile = os.Stdin
	case 1:
		inFile, err = os.Open(dungFlags.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(fmt.Errorf("incorrect number of arguments after the command flags; should be 0, to read from stdin, or 1, naming the input file\n"))
		return
	}

	if *formatFlag == "tgf" {
		af, err = tgf.Import(inFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	printExtensions := func(extensions []dung.ArgSet) {
		s := []string{}
		for _, E := range extensions {
			s = append(s, E.String())
		}
		fmt.Printf("[%s]\n", strings.Join(s, ","))
	}

	printExtension := func(E dung.ArgSet, exists bool) {
		if exists {
			fmt.Printf("%s\n", E)
		} else {
			fmt.Printf("NO\n")
		}
	}

	printBool := func(b bool) {
		if b {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	}

	problem := *problemFlag + "-" + *semanticsFlag

	// Grounded Semantics
	if problem == "DC-GR" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Grounded, dung.Arg(arg)))
	} else if problem == "DS-GR" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Grounded, dung.Arg(arg)))
	} else if problem == "EE-GR" {
		E := af.GroundedExtension()
		extensions = []dung.ArgSet{E}
		fmt.Printf("[%s]\n", E)
	} else if problem == "SE-GR" {
		E := af.GroundedExtension()
		extensions = []dung.ArgSet{E}
		fmt.Printf("%s\n", E)

		// Preferred Semantics
	} else if problem == "DC-PR" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Preferred, dung.Arg(arg)))
	} else if problem == "DS-PR" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Preferred, dung.Arg(arg)))
	} else if problem == "EE-PR" {
		extensions = af.PreferredExtensions()
		printExtensions(extensions)
	} else if problem == "SE-PR" {
		E, ok := af.SomeExtension(dung.Preferred)
		printExtension(E, ok)

		// Complete Semantics
	} else if problem == "DC-CO" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Complete, dung.Arg(arg)))
	} else if problem == "DS-CO" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Complete, dung.Arg(arg)))
	} else if problem == "EE-CO" {
		extensions = af.CompleteExtensions()
		printExtensions(extensions)
	} else if problem == "SE-CO" {
		E, ok := af.SomeExtension(dung.Complete)
		printExtension(E, ok)

		// Stable Semantics
	} else if problem == "DC-ST" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Stable, dung.Arg(arg)))
	} else if problem == "DS-ST" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Stable, dung.Arg(arg)))
	} else if problem == "EE-ST" {
		extensions = af.StableExtensions()
		printExtensions(extensions)
	} else if problem == "SE-ST" {
		E, ok := af.SomeExtension(dung.Stable)
		printExtension(E, ok)
	} else if problem == "traverse" {
		af.Traverse(func(E dung.ArgSet) {
			fmt.Printf("%v\n", E)
		})
	} else {
		log.Fatal(fmt.Errorf("unsupported problem: %s\n", problem))
		return
	}
	if *outputFlag != "" && extensions != nil {
		if _, err := os.Stat(*outputFlag); err == nil {
			log.Fatal(fmt.Errorf("The output directory, %s, should not already exist\n", *outputFlag))
			return
		}
		if err = os.MkdirAll(*outputFlag, 0755); err != nil {
			log.Fatal(fmt.Errorf("%s\n", err))
			return
		}
		for i, ext := range extensions {
			filename := "e" + fmt.Sprintf("%d", i) + ".graphml"
			f, err := os.Create(filepath.Join(*outputFlag, filename))
			if err != nil {
				log.Fatal(fmt.Errorf("%s\n", err))
				return
			}
			graphml.Export(f, af, ext)
			f.Close()
		}
	}
}
