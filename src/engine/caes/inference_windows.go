// Copyright © 2016 The Carneades Authors
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
	"strconv"
	"strings"
	"time"

	"github.com/carneades/carneades-4/src/engine/terms"
	// "github.com/mndrix/golog/read"
	// "github.com/mndrix/golog/term"
)

// resource limits for Prolog processes
const (
	timeLimit  = 15     // Seconds
	stackLimit = "256m" // MB
)

const header = `
:- use_module(library(chr)).
:- use_module(library(http/json)).
:- use_module(library(http/json_convert)).
:- chr_constraint argument/2, go/0, '¬'/1.
:- json_object argument(scheme:text, values:list(text)).
:- initialization main.
:- op(900, fx, ¬).

terms_strings([],[]).
terms_strings([H|T],[SH|ST]) :-
  term_string(H,SH),
  terms_strings(T,ST).

argument(I,P) ==> 
  term_string(I,S),
  terms_strings(P,L),
  prolog_to_json(argument(S,L),J), 
  json_write(current_output, J), 
  nl. 

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
// by forward chaining from the goals. If the theory could not be
// translated or saved to a temporary file, an error is returned.
// If all goes well, nil is returned.
func writeCHR(t *Theory, assms map[string]bool, f *os.File) error {
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

			conclusions := []string{}
			// append the assumptions of the schemes
			// as additional conclusions
			for _, stmt := range s.Assumptions {
				conclusions = append(conclusions, stmt)
			}
			// also append the actual conclusions
			for _, stmt := range s.Conclusions {
				conclusions = append(conclusions, stmt)
			}
			writeTerms(conclusions)
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
		return errors.New("Could not write the constraint handling rules to a temporary file.")
	} else {
		return nil
	}
}

// makeIssue: match the patterns of an issue scheme against the
// statements of the argument graph.  If more than one statement
// matches, make them positions of an issue, creating the issue
// if one does not already exist and adding it to the argument graph.
// Every statement may be a position of at most one issue.  No statement
// is made a position of some issue if this constraint would be violated.
// If some pattern is not synatically correct and thus cannot be parsed,
// an error is returned and the argument graph is left unchanged.
// If all goes well, the argument graph is updated and nil is returned
func (ag *ArgGraph) makeIssue(issueScheme string, patterns []string) (err error) {

	// skip issue schemes with no patterns
	if len(patterns) == 0 {
		fmt.Fprintf(os.Stderr, "Issue scheme with no patterns: %v\n", issueScheme)
		return
	}
	// Try to match the first pattern with each statement
	// in the argument graph.
	pattern, ok := terms.ReadString(patterns[0])

	if !ok {
		fmt.Fprintf(os.Stderr, "Could not parse issue scheme pattern: %v\n", patterns[0])
		return
	}
	for wff1, stmt := range ag.Statements {
		term1, ok := terms.ReadString(wff1)
		if !ok {
			fmt.Fprintf(os.Stderr, "Statement key not a term: %v\n", wff1)
			continue
		}
		var bindings terms.Bindings
		bindings, ok = terms.Match(pattern, term1, bindings)
		if !ok {
			continue // terms do not match
		} else {
			candidates := []*Statement{stmt}

			// Check if the issue scheme defines an enumeration.
			isEnumeration := len(patterns) > 1 && patterns[1] == "..."

			// Create a copy of bindings with all variables with names ending
			// in integer indexes unbound
			var bindings2 terms.Bindings
			for env := bindings; env != nil; env = env.Next {
				v := env.Var
				t := env.T
				suffix := v.Name[len(v.Name)-1:]
				_, err := strconv.Atoi(suffix)
				if err != nil {
					// the variable does not end with an integer suffix
					// so keep its binding
					bindings2 = terms.AddBinding(v, t, bindings2)
				}
			}

			// For each matching statement, iterate over the statements
			// again to try to find other positions of the issue. Whether or
			// not a statement is a position depends on the remaining patterns
			// of the issue scheme and, in particular, whether or not the
			// issue scheme is an enumeration.

			for wff2, stmt2 := range ag.Statements {
				if wff2 == wff1 {
					// skip the matching statement found previously
					continue
				}
				term2, ok := terms.ReadString(wff2)
				if !ok {
					fmt.Fprintf(os.Stderr, "Statement key not a term: %v\n", wff2)
					continue
				}

				match := false
				if !isEnumeration {
					// try matching against each of the remaining patterns
					// and update the bindings and add the statement as
					// as candidate if any pattern matches
					for _, p := range patterns[1:] {
						pattern2, ok := terms.ReadString(p)
						if !ok {
							fmt.Fprintf(os.Stderr, "Could not parse issue scheme pattern: %v\n", p)
							continue
						}
						bindings, match = terms.Match(pattern2, term2, bindings)
						if match {
							break
						}
					}
				} else {
					// Use a fresh copy of bindings2 for enumeration issue patterns
					var b2copy terms.Bindings
					for env := bindings2; env != nil; env = env.Next {
						k := env.Var
						v := env.T
						b2copy = terms.AddBinding(k, v, b2copy)
					}
					b2copy, match = terms.Match(pattern, term2, b2copy)
				}
				if !match {
					continue // terms do not match
				} else {
					candidates = append(candidates, stmt2)
				}
			}

			// Check whether any of the candidates found are already at issue
			// and, if so, that they are all positions of the same issue.
			// No statement may be a position of more than one issue.
			// Add candidates which do not violate the single issue constraint
			// to the list of positions.
			var issue *Issue
			var positions = []*Statement{}
			for _, c := range candidates {
				if c.Issue != nil {
					if issue == nil {
						issue = c.Issue
					} else if c.Issue != issue {
						// found a conflict, due to statements being positions of different issues
						fmt.Fprintf(os.Stderr, "Statement matching an issue scheme is already a position of different issue: %v\n", c.Id)
						continue
					}
				}
				positions = append(positions, c)
			}

			// Do not make an issue when there are less than two positions
			if len(positions) < 2 {
				continue
			}

			// Add the statements which are not at issue to the existing
			// issue, if any, or create a new issue and add all the statements
			// found as positions of the new issue.
			if issue == nil {
				// create a new issue
				issue = NewIssue()
				issue.Positions = positions
				for _, pos := range positions {
					pos.Issue = issue
				}
				// generate an id for the new issue
				i := len(ag.Issues) + 1
				prefix := "i"
				id := prefix + strconv.Itoa(i)
				_, existing := ag.Issues[id]
				for existing {
					i++
					id = prefix + strconv.Itoa(i)
					_, existing = ag.Issues[id]
				}
				// add the new issue to the argument graph
				issue.Id = id
				ag.Issues[id] = issue
			} else {
				// add new positions to the existing issue
				for _, pos := range positions {
					if pos.Issue == nil {
						pos.Issue = issue
					}
				}
			}
		}
	}
	return err
}

// Infer: Translate a theory into CHR rules and use
// SWI Prolog to construct arguments and add them to the argument graph.
// Does not compute or update labels.  If the theory is synatically incorrect
// and thus cannot be parsed by the CHR inference engine, an error is returned
// and argument graph is left unchanged. If all goes well, the argument
// graph is updated and nil is returned.
func (ag *ArgGraph) Infer() error {
	// Translate the theory to CHR in SWI-Prolog and
	// write the output to a temporary file
	if ag.Theory == nil || ag.Theory.ArgSchemes == nil || len(ag.Theory.ArgSchemes) == 0 {
		return nil
	}
	f, err := ioutil.TempFile(os.TempDir(), "carneades")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// Create an index of the previous arguments constructed
	// to avoid constructing equivalent instanstiations of schemes
	// and to allow the inference engine to construct undercutters
	prevArgs := map[string]bool{}
	for _, a := range ag.Arguments {
		if a != nil {
			prevArgs["argument("+a.Scheme.Id+",["+strings.Join(a.Parameters, ",")+"])"] = true
		}
	}

	// The assumptions to be handled as goals by the CHR inference engine
	// consists of the assumptions of the argument graph plus assumptions
	// for each of the previous arguments
	assums := map[string]bool{}
	for k, v := range ag.Assumptions {
		assums[k] = v
	}
	for k, v := range prevArgs {
		assums[k] = v
	}
	err = writeCHR(ag.Theory, assums, f)
	if err != nil {
		return err
	}

	// Call SWI Prolog to evaluate the theory and write arguments
	// to standard out.  Handle SWI-Prolog errors.  Assure termination
	// within given limits (time, stack size, ...)
	cmd := exec.Command("C:\\Program Files (x86)\\swipl\\bin\\swipl.exe", "-s ", f.Name(), "-L"+stackLimit)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	//	stderr, err := cmd.StderrPipe()
	//	if err != nil {
	//		return err
	//	}

	//cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Read the output and construct CAES arguments by instantiating
	// schemes in the theory and adding statements and arguments to
	// the argument graph
	scanner := bufio.NewScanner(stdout)
	// Wrap the individual JSON objects in a JSON array
	bytes := []byte{}

	//	for scanner.Scan() {
	//		bytes = append(bytes, scanner.Bytes()...)
	//	}

	// Limit the runtime of the SWI Prolog command
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	finished := false
	timer := time.After(timeLimit * time.Second)
	for !finished {
		select {
		case <-timer:
			cmd.Process.Kill()
			finished = true
		case <-done:
			finished = true
		default:
			scanner.Scan()
			bytes = append(bytes, scanner.Bytes()...)
		}
	}
	re, err := regexp.Compile("}[[:space:]]*{")
	if err != nil {
		return err
	}
	bytes = re.ReplaceAll(bytes, []byte("},\n{"))
	bytes = []byte("[" + string(bytes) + "]")

	// Unmarshall the bytes
	var d []ArgDesc
	err = json.Unmarshal(bytes, &d)
	if err != nil {
		return err
	}

	for _, a := range d {
		// Check that an equivalent argument is not already in the graph
		if _, exists := prevArgs[a.Scheme+"("+strings.Join(a.Values, ",")+")"]; !exists {
			ag.InstantiateScheme(a.Scheme, a.Values)
			prevArgs["argument("+a.Scheme+",["+strings.Join(a.Values, ",")+"])"] = true
		}
	}

	// Use issue schemes of the theory to derive or update the issues
	// of the argument graph

	if ag.Theory.IssueSchemes != nil {
		for issue, patterns := range ag.Theory.IssueSchemes {
			err = ag.makeIssue(issue, *patterns)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Infer: %v\n", err)
				continue
			}
		}
	}

	return nil
}
