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
		labeling := af.GroundedLabelling()
		if labeling.Get(arg) == dung.In {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "EE-GR" {
		labeling := af.GroundedLabelling()
		fmt.Printf("[%s]\n", labeling.AsExtension().String())
	} else if *problemFlag == "SE-GR" {
		labeling := af.GroundedLabelling()
		fmt.Printf("%s\n", labeling.AsExtension().String())
	} else if *problemFlag == "DC-PR" {
		if af.CredulouslyInferredPR(arg) {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "DS-PR" {
		if af.SkepticallyInferredPR(arg) {
			fmt.Printf("YES\n")
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "EE-PR" {
		labellings := af.PreferredLabellings()
		s := []string{}
		for _, l := range labellings {
			s = append(s, l.AsExtension().String())
		}
		fmt.Printf("[%s]\n", strings.Join(s, ","))
	} else if *problemFlag == "EE-PR3" {
		labellings := af.PreferredLabellings3()
		s := []string{}
		for _, l := range labellings {
			s = append(s, l.AsExtension().String())
		}
		fmt.Printf("[%s]\n", strings.Join(s, ","))
	} else if *problemFlag == "SE-PR" {
		labellings := af.PreferredLabellings3()
		if len(labellings) > 0 {
			fmt.Printf("%s\n", labellings[0].AsExtension().String())
		} else {
			fmt.Printf("NO\n")
		}
	} else if *problemFlag == "traverse" {
		af.TraverseLabellings(func(L dung.PLabelling) {
			fmt.Printf("%v\n", L.AsExtension())
		})
	} else {
		log.Fatal(fmt.Errorf("unsupported problem: %s\n", *problemFlag))
		return
	}
}
