package main

import (
	"./engine/dung"
	"./serialization/tgf"
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
const problems = "[DC-GR,DS-GR,EE-GR,SE-GR,DC-PR,DS-PR,EE-PR,SE-PR]"

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

	if (*problemFlag == "DC-GR" || *problemFlag == "DS-GR") && *argFlag != "" {
		E := af.GroundedExtension()
		if E.Contains(dung.Arg(arg)) {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "EE-GR" {
		E := af.GroundedExtension()
		fmt.Printf("[%s]\n", E)
	} else if *problemFlag == "SE-GR" {
		E := af.GroundedExtension()
		fmt.Printf("%s\n", E)
	} else if *problemFlag == "DC-PR" {
		if af.CredulouslyInferredPR(dung.Arg(arg)) {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "DS-PR" {
		if af.SkepticallyInferredPR(dung.Arg(arg)) {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "EE-PR" {
		extensions := af.PreferredExtensions()
		s := []string{}
		for _, E := range extensions {
			s = append(s, E.String())
		}
		fmt.Printf("[%s]\n", strings.Join(s, ","))
	} else if *problemFlag == "SE-PR" {
		extensions := af.PreferredExtensions()
		if len(extensions) > 0 {
			fmt.Printf("%s\n", extensions[0])
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "traverse" {
		af.Traverse(func(E dung.ArgSet) {
			fmt.Printf("%v\n", E)
		})
	} else {
		log.Fatal(fmt.Errorf("unsupported problem: %s\n", *problemFlag))
		return
	}
}
