package main

import (
	"../../engine/dung"
	"../../serialization/tgf"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const name = "Carneades"
const version = "v0.2"
const author = "Tom Gordon (thomas.gordon@fokus.fraunhofer.de)"
const formats = "[tgf]"
const problems = "[DC-GR,DS-GR,EE-GR,SE-GR,DC-PR,DS-PR,EE-PR,SE-PR,DC-CO,DS-CO,EE-CO,SE-CO,DC-ST,DS-ST,EE-ST,SE-ST]"

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s %s\n%s\n", name, version, author)
	}

	formatsFlag := flag.Bool("formats", false, "print supported formats")
	problemsFlag := flag.Bool("problems", false, "print supported problems")
	problemFlag := flag.String("p", "DC-GR", "the problem to solve")
	fileFlag := flag.String("f", "", "the source file for the AF")
	formatFlag := flag.String("fo", "tgf", "the format of the source file")
	argFlag := flag.String("a", "", "the id of the argument to check")

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
	} else if *problemFlag == "EE-GR" || *problemFlag == "SE-GR" {
		E := af.GroundedExtension()
		fmt.Printf("[%s]\n", E)

		// Preferred Semantics
	} else if *problemFlag == "DC-PR" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Preferred, dung.Arg(arg)))
	} else if *problemFlag == "DS-PR" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Preferred, dung.Arg(arg)))
	} else if *problemFlag == "EE-PR" {
		printExtensions(af.PreferredExtensions())
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
		printExtensions(af.CompleteExtensions())
	} else if *problemFlag == "SE-CO" {
		E, ok := af.SomeExtension(dung.Preferred)
		printExtension(E, ok)

		// Stable Semantics
	} else if *problemFlag == "DC-ST" {
		checkArgFlag()
		printBool(af.CredulouslyInferred(dung.Stable, dung.Arg(arg)))
	} else if *problemFlag == "DS-ST" {
		checkArgFlag()
		printBool(af.SkepticallyInferred(dung.Stable, dung.Arg(arg)))
	} else if *problemFlag == "EE-ST" {
		printExtensions(af.StableExtensions())
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
}
