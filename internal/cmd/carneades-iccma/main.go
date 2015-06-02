// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"github.com/carneades/carneades-go/internal/engine/dung"
	"github.com/carneades/carneades-go/internal/engine/dung/encoding/graphml"
	"github.com/carneades/carneades-go/internal/engine/dung/encoding/tgf"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const name = "Carneades ICCMA"
const version = "v1.0"
const author = "Tom Gordon (thomas.gordon@fokus.fraunhofer.de)"
const formats = "[tgf]"
const problems = "[DC-GR,DS-GR,EE-GR,SE-GR,DC-PR,DS-PR,EE-PR,SE-PR,DC-CO,DS-CO,EE-CO,SE-CO,DC-ST,DS-ST,EE-ST,SE-ST]"

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s %s\n%s\n", name, version, author)
		return
	}

	formatsFlag := flag.Bool("formats", false, "print supported formats")
	problemsFlag := flag.Bool("problems", false, "print supported problems")
	problemFlag := flag.String("p", "DC-GR", "the problem to solve")
	fileFlag := flag.String("f", "", "the source file for the AF")
	formatFlag := flag.String("fo", "tgf", "the format of the source file")
	argFlag := flag.String("a", "", "the id of the argument to check")
	outputFlag := flag.String("o", "", "the name of a directory to create to output GraphML")

	flag.Parse()

	arg := *argFlag

	checkArgFlag := func() {
		if arg == "" {
			log.Fatal(fmt.Errorf("no -a flag"))
		}
	}

	if *formatsFlag {
		fmt.Printf("%s\n", formats)
		return
	}

	if *problemsFlag {
		fmt.Printf("%s\n", problems)
		return
	}

	if *formatFlag != "tgf" {
		log.Fatal(fmt.Errorf("unsupported format: %s\n", *formatFlag))
		return
	}

	var inFile *os.File
	var err error
	var af dung.AF
	var extensions []dung.ArgSet

	if *fileFlag == "" {
		log.Fatal(fmt.Errorf("no file flag (-f)"))
		return
	}

	if *formatFlag == "tgf" {
		inFile, err = os.Open(*fileFlag)
		if err != nil {
			log.Fatal(err)
		}
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

	// Grounded Semantics
	if *problemFlag == "DC-GR" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Grounded, dung.Arg(arg)))
	} else if *problemFlag == "DS-GR" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Grounded, dung.Arg(arg)))
	} else if *problemFlag == "EE-GR" {
		E := af.GroundedExtension()
		extensions = []dung.ArgSet{E}
		fmt.Printf("[%s]\n", E)
	} else if *problemFlag == "SE-GR" {
		E := af.GroundedExtension()
		extensions = []dung.ArgSet{E}
		fmt.Printf("%s\n", E)

		// Preferred Semantics
	} else if *problemFlag == "DC-PR" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Preferred, dung.Arg(arg)))
	} else if *problemFlag == "DS-PR" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Preferred, dung.Arg(arg)))
	} else if *problemFlag == "EE-PR" {
		extensions = af.PreferredExtensions()
		printExtensions(extensions)
	} else if *problemFlag == "SE-PR" {
		E, ok := af.SomeExtension(dung.Preferred)
		printExtension(E, ok)

		// Complete Semantics
	} else if *problemFlag == "DC-CO" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Complete, dung.Arg(arg)))
	} else if *problemFlag == "DS-CO" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Complete, dung.Arg(arg)))
	} else if *problemFlag == "EE-CO" {
		extensions = af.CompleteExtensions()
		printExtensions(extensions)
	} else if *problemFlag == "SE-CO" {
		E, ok := af.SomeExtension(dung.Complete)
		printExtension(E, ok)

		// Stable Semantics
	} else if *problemFlag == "DC-ST" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Stable, dung.Arg(arg)))
	} else if *problemFlag == "DS-ST" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Stable, dung.Arg(arg)))
	} else if *problemFlag == "EE-ST" {
		extensions = af.StableExtensions()
		printExtensions(extensions)
	} else if *problemFlag == "SE-ST" {
		E, ok := af.SomeExtension(dung.Stable)
		printExtension(E, ok)
	} else if *problemFlag == "traverse" {
		af.Traverse(func(E dung.ArgSet) {
			fmt.Printf("%v\n", E)
		})
	} else {
		log.Fatal(fmt.Errorf("unsupported problem: %s\n", *problemFlag))
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
