// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Inference of arguments (aka argument construction or generation)
// using the SWI Prolog implementation of Constraint Handling Rules (CHR)

package caes

import (
	"bufio"
	// "bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"
)

// resource limits for Prolog processes
const (
	timeLimit  = 15     // seconds
	stackLimit = "256m" // MB
)

const header = `
:- use_module(library(chr)).
:- use_module(library(http/json)).
:- use_module(library(http/json_convert)).
:- chr_constraint argument/2, go/0.
:- json_object argument(scheme:text, values:list(text)).
:- initialization main.

terms_strings([],[]).
terms_strings([H|T],[SH|ST]) :-
  term_string(H,SH),
  terms_strings(T,ST).

argument(I,P) <=> 
  term_string(I,S),
  terms_strings(P,L),
  prolog_to_json(argument(S,L),J), 
  json_write(current_output, J), 
  nl | 
  true.

main :-
  assumptions,
  halt(0).
`

// ArgDesc: Structure describing an argument instantiating an
// argument scheme
type ArgDesc struct {
	Scheme string   // id of the scheme
	Values []string // values instantiating the variables of the scheme
}

// Translate the theory and assumptions to CHR in SWI-Prolog and
// write the output to the given file. The assumptions are
// translated into CHR "goals", to which the CHR rules will be applied,
// by forward chaining from the goals.
func writeCHR(t *Theory, assms map[string]bool, f *os.File) (bool, error) {
	// Write each term of a slice of terms on a separate line
	// indented by four spaces and separated by commas.
	// Write nothing after the last term, not even white space.
	var err error

	writeTerms := func(v []string) {
		n := len(v) - 1
		i := 0
		for _, term := range v {
			_, err = f.WriteString("  " + term)
			if i < n {
				_, err = f.WriteString(",\n")
			}
			i++
		}
	}

	_, err = f.WriteString(header)

	// Translate the language of the theory to constraint declarations
	if len(t.Language) > 0 {
		_, err = f.WriteString("\n\n:- chr_constraint ")
		n := len(t.Language) - 1
		i := 0
		for term, _ := range t.Language {
			_, err = f.WriteString(term)
			if i < n {
				_, err = f.WriteString(",\n   ")
			} else {
				_, err = f.WriteString(".\n\n")
			}
			i++
		}
	}

	// Translate the argumentation schemes of the theory to CHR rules
	for id, s := range t.ArgSchemes {
		// If the scheme has no conclusions, skip the scheme
		// and assume it only defines a weighing function but no rule
		if len(s.Conclusions) > 0 {
			// Partition the premises into ones to keep and ones
			// to delete
			keep := []string{}
			remove := []string{}
			for k, term := range s.Premises {
				member := false
				for _, d := range s.Deletions {
					if k == d {
						member = true
						break
					}
				}
				if member {
					remove = append(remove, term)
				} else {
					keep = append(keep, term)
				}
			}

			// write the rule
			// write the rule id
			_, err = f.WriteString(id + " @\n")
			// write the heads to keep
			if len(keep) == 0 {
				// A "go" term is added to CHR rules for
				// argument schemes with no premises, since CHR requires
				// rules to have at least one term in the head.
				_, err = f.WriteString("  go\n")
			} else {
				writeTerms(keep)
				_, err = f.WriteString("\n")
			}
			// write the heads to delete
			if len(remove) > 0 {
				_, err = f.WriteString("\\ \n")
				writeTerms(remove)
				_, err = f.WriteString("\n")
				_, err = f.WriteString("<=>\n")
			} else {
				_, err = f.WriteString("==>\n")
			}

			// write the guards
			if len(s.Guards) > 0 {
				writeTerms(s.Guards)
				_, err = f.WriteString("|\n")
			}
			// write the argument
			_, err = f.WriteString("  argument(" + id + ",[")
			for i := 0; i < len(s.Variables); i++ {
				if i < len(s.Variables)-1 {
					_, err = f.WriteString(s.Variables[i] + ",")
				} else {
					_, err = f.WriteString(s.Variables[i])
				}
			}
			_, err = f.WriteString("]),\n")

			// write the conclusions
			writeTerms(s.Conclusions)
			_, err = f.WriteString(".\n\n")

		}
	}

	// Translate the assumptions into a Prolog rule, where each
	// assumption is a literal of the body of the rule and the
	// head of the rule is assumptions/0.
	l := []string{}
	// Convert the ag.Assumptions map to a slice of the assumed wffs
	for wff, b := range assms {
		if b == true {
			l = append(l, wff)
		}
	}
	_, err = f.WriteString("\n\nassumptions :- \n  ")
	// go is a dummy assumption for handling arguments
	// with no premises.
	_, err = f.WriteString("go")
	if len(l) > 0 {
		_, err = f.WriteString(",\n")
		n := len(l)
		for i := 0; i < n; i++ {
			_, err = f.WriteString("  " + l[i])
			if i < n-1 {
				_, err = f.WriteString(",\n  ")
			}
		}
	}
	_, err = f.WriteString(".\n\n")

	if err != nil {
		return false, errors.New("Could not write the constraint handling rules to a temporary file.")
	} else {
		return true, nil
	}
}

// Run a command with a time limit
func runCmd(cmd *exec.Cmd) {
	// Run the command in its own process group, so that each
	// process can be interrupted separately.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	if err != nil {
		// wait or timeout
		donec := make(chan error, 1)
		go func() {
			donec <- cmd.Wait()
		}()
		select {
		case <-time.After(timeLimit * time.Second):
			cmd.Process.Kill()
		case <-donec:
		}
	}
}

// Infer: Translate a theory into CHR rules and use
// SWI Prolog to construct arguments and add them to the argument graph.
// Does not compute or update labels.
func (ag *ArgGraph) Infer() (bool, error) {
	// Translate the theory to CHR in SWI-Prolog and
	// write the output to a temporary file
	if ag.Theory == nil || ag.Theory.ArgSchemes == nil || len(ag.Theory.ArgSchemes) == 0 {
		return true, nil
	}
	f, err := ioutil.TempFile(os.TempDir(), "carneades")
	if err != nil {
		return false, err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	writeCHR(ag.Theory, ag.Assumptions, f)

	// Call SWI Prolog to evaluate the theory and write arguments
	// to standard out.  Handle SWI-Prolog errors.  Assure termination
	// within given limits (time, stack size, ...)
	cmd := exec.Command("swipl", "-s ", f.Name(), "-L"+stackLimit)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	//	stderr, err := cmd.StderrPipe()
	//	if err != nil {
	//		return false, err
	//	}
	runCmd(cmd)

	//// Debugging code:
	//	buf1 := new(bytes.Buffer)
	//	buf1.ReadFrom(stdout)
	//	fmt.Printf("output:\n%v\n", buf1.String())

	// Read the output and construct CAES arguments by instantiating
	// schemes in the theory and adding statements and arguments to
	// the argument graph
	scanner := bufio.NewScanner(stdout)
	// Wrap the individual JSON objects in a JSON array
	bytes := []byte{}
	for scanner.Scan() {
		bytes = append(bytes, scanner.Bytes()...)
	}
	re, err := regexp.Compile("}[[:space:]]*{")
	if err != nil {
		return false, err
	}
	bytes = re.ReplaceAll(bytes, []byte("},\n{"))
	bytes = []byte("[" + string(bytes) + "]")
	fmt.Printf("output:\n%v\n", string(bytes))
	var d []ArgDesc
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return false, err
	}
	for _, a := range d {
		ag.InstantiateScheme(a.Scheme, a.Values)
	}

	// Use issue schemes of the theory to derive and update the issues
	// of the argument graph

	// START HERE

	return true, nil
}
