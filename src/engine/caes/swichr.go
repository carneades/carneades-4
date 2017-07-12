// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// An implementation of the CAES Rulebase interface using the
// SWI Prolog implementation of Constraint Handling Rules

package caes

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	// "log"
	"math"
	"os"
	"strings"
	"time"
)

const DEBUG = false

const header = `
:- use_module(library(chr)).
:- chr_constraint argument/2, go/0, '¬'/1.
:- initialization main.
:- op(900, fx, ¬).

write_constraint_store :-
	find_chr_constraint(C),
	writeln(C),
	fail.
write_constraint_store.
	
main :-
  assumptions,
  write_constraint_store,
  halt(0).
`

const PREFIX = "ß"

type SWIRulebase struct {
	// constraint specifiers and rules in SWI Prolog syntax
	constraints []string
	rules       []string
}

func MakeSWIRulebase(l Language) *SWIRulebase {
	cs := []string{}
	for term, _ := range l {
		cs = append(cs, term)
	}
	return &SWIRulebase{constraints: cs}
}

func (rb *SWIRulebase) AddRule(name string, keeps []string, deletes []string, guards []string, body []string) error {
	var rule string
	f := func(terms []string) string {
		return strings.Join(terms, ",")
	}
	if len(keeps) > 0 && len(deletes) == 0 {
		// propagation rule
		if len(guards) > 0 {
			rule = fmt.Sprintf("%s @ %s ==> %s | %s.", name, f(keeps), f(guards), f(body))
		} else {
			rule = fmt.Sprintf("%s @ %s ==> %s.", name, f(keeps), f(body))
		}
	} else if len(keeps) == 0 && len(deletes) > 0 {
		// simplification rule
		if len(guards) > 0 {
			rule = fmt.Sprintf("%s @ %s <=> %s | %s.", name, f(deletes), f(guards), f(body))
		} else {
			rule = fmt.Sprintf("%s @ %s <=> %s.", name, f(deletes), f(body))
		}
	} else if len(keeps) > 0 && len(deletes) > 0 {
		// simpagation rule
		if len(guards) > 0 {
			rule = fmt.Sprintf("%s @ %s \\ %s <=> %s | %s.", name, f(keeps), f(deletes), f(guards), f(body))
		} else {
			rule = fmt.Sprintf("%s @ %s \\ %s <=> %s.", name, f(keeps), f(deletes), f(body))
		}
	} else {
		// syntactically incorrect, so ignore and return an error
		return errors.New(fmt.Sprintf("In the rule named %q both keeps and deletes are empty.", name))
	}
	rb.rules = append(rb.rules, rule)
	return nil
}

// Translate a theory into a SWIRulebase
func TheoryToSWIRulebase(t *Theory) *SWIRulebase {
	// log.Println("TheoryToSWIRulebase") // DEBUG
	rb := MakeSWIRulebase(t.Language)
	for _, s := range t.ArgSchemes {
		// If the scheme has no conclusions, skip the scheme
		// and assume it only defines a weighing function but no rule
		if len(s.Conclusions) > 0 {
			// A "go" term is added to CHR rules for
			// argument schemes with no premises, since CHR requires
			// rules to have at least one term in the head.
			var premises []string
			if len(s.Premises) > 0 {
				premises = s.Premises
			} else if len(s.Deletions) == 0 {
				premises = []string{"go"}
			}
			argTerm := fmt.Sprintf("argument(%s,[%s])", s.Id, strings.Join(s.Variables, ","))
			conclusions := append(s.Conclusions, s.Assumptions...)
			conclusions = append(conclusions, argTerm)
			// Errors raised by AddRule are ignored. The rule is just skipped.
			// Note that the body of the rules includes the assumptions
			// and conclusions of the scheme.
			rb.AddRule(s.Id, premises, s.Deletions, s.Guards, conclusions)
		}
	}
	return rb
}

// Translate a SWIRulebase and goals to CHR in SWI-Prolog and
// write the output to the given file. If the rulebase or goals could not be
// translated or saved to a temporary file, an error is returned.
// If all goes well, nil is returned.
func writeCHR(t *SWIRulebase, goals []string, f *os.File) error {
	// Write each term of a slice of terms on a separate line
	// indented by four spaces and separated by commas.
	// Write nothing after the last term, not even white space.
	var err error

	_, err = f.WriteString(header)

	// Translate the constraints of the SWIRulebase to constraint declarations
	if len(t.constraints) > 0 {
		_, err = f.WriteString("\n\n:- chr_constraint ")
		n := len(t.constraints) - 1
		i := 0
		for _, term := range t.constraints {
			_, err = f.WriteString(term)
			if i < n {
				_, err = f.WriteString(",\n   ")
			} else {
				_, err = f.WriteString(".\n\n")
			}
			i++
		}
	}

	// Write the rules of the SWIRulebase
	for _, r := range t.rules {
		f.WriteString(r + "\n")
	}

	// Translate the goals into a Prolog rule, where each
	// goal is a literal of the body of the rule and the
	// head of the rule is assumptions/0.

	_, err = f.WriteString("\n\nassumptions :- \n  ")
	n := len(goals)
	for i := 0; i < n; i++ {
		_, err = f.WriteString("  " + goals[i])
		if i < n-1 {
			_, err = f.WriteString(",\n")
		} else {
			_, err = f.WriteString(".\n\n")
		}
	}

	if err != nil {
		return errors.New("Could not write the constraint handling rules to a temporary file.")
	} else {
		return nil
	}
}

// Infer: Apply an SWIRulebase to a list of goals.  The
// maximum amount of time alloted the SWI Prolog process depends
// on the maximum number of rules (schemes) which may be applied, max.
// Return true if the goals are successfully solved, and false if the goals fail.
// Return a list of the terms in the constraint store, if the goals
// succeeded. Returns an error if the rulebase could not be applied to the goals.
func (rb *SWIRulebase) Infer(goals []string, max int) (bool, []string, error) {

	// Assume that a rule or scheme application takes about 0.00015 seconds.
	// If there is no limit to the number of rule applications (i.e. max==0),
	// then limit the maximum amount of time alloted to the SWI Prolog to 15 seconds
	// anyway, to assure termination. Otherwise limit the time to
	// max * 0.00015 seconds, rounded up to the next second.

	secsPerRule := 0.00015
	timeLimit := 15 // seconds
	if max > 0 {
		timeLimit = int(math.Ceil(float64(max) * secsPerRule))
	}

	f, err := ioutil.TempFile(os.TempDir(), "swirulebase")
	if err != nil {
		return false, nil, err
	}

	defer f.Close()
	if !DEBUG {
		defer os.Remove(f.Name())
	}

	err = writeCHR(rb, goals, f)
	if err != nil {
		return false, nil, err
	}

	cmd := MakeSWIPrologCmd(f)
	if err != nil {
		return false, nil, err
	}
	stdout, err := cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		return false, nil, err
	}

	// Limit the runtime of the SWI Prolog command
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	finished := false
	timer := time.After(time.Duration(timeLimit) * time.Second)

	scanner := bufio.NewScanner(stdout)
	store := []string{}

	for !finished {
		select {
		case <-timer:
			cmd.Process.Kill()
			finished = true
		case <-done:
			finished = true
		default:
			for scanner.Scan() {
				store = append(store, scanner.Text())
			}
		}
	}

	// Read the lines of the output file
	// Each line represents a term of the constraint store,
	// in Prolog syntax.

	// Return false if the constraint store includes "fail".
	for _, t := range store {
		if t == "fail" {
			return false, []string{}, nil
		}
	}
	// If the constraint store does not include "fail", return true
	// along with a slice with the terms in the store.
	return true, store, nil
}
