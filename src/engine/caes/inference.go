// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Inference of arguments (aka argument construction or generation)
// using the SWI Prolog implementation of Constraint Handling Rules (CHR)

package caes

import (
	"errors"
	"io/ioutil"
	"os"
)

const header = `
:- use_module(library(chr)).
:- use_module(library(http/json)).
:- use_module(library(http/json_convert)).
:- chr_constraint argument/2.
:- json_object argument(scheme:text, parameters:list(text)).
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
			_, err = f.WriteString("    " + term)
			if i < n {
				_, err = f.WriteString(",\n")
			}
			i++
		}
	}

	_, err = f.WriteString(header)

	// Translate the language of the theory to constraint declarations
	_, err = f.WriteString(":- chr_constraint ")
	n := len(t.Language) - 1
	i := 0
	for term, _ := range t.Language {
		_, err = f.WriteString(term)
		if i == n {
			_, err = f.WriteString(",\n   ")
		} else {
			_, err = f.WriteString(".\n\n")
		}
		i++
	}

	// Translate the argumentation schemes of the theory to CHR rules
	for id, s := range t.ArgSchemes {
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
		_, err = f.WriteString(id + "@\n")
		// write the heads to keep
		if len(keep) == 0 {
			_, err = f.WriteString("    true\n")
		} else {
			writeTerms(keep)
			_, err = f.WriteString("\n")
		}
		// write the heads to delete
		if len(remove) > 0 {
			_, err = f.WriteString("\\ \n")
			writeTerms(remove)
			_, err = f.WriteString("\n")
		}
		_, err = f.WriteString("<=>\n")
		// write the guards
		if len(s.Guards) > 0 {
			writeTerms(s.Guards)
			_, err = f.WriteString("|\n")
		}
		// write the argument
		_, err = f.WriteString("    argument(" + id + ",[")
		for i := 0; i < len(s.Variables); i++ {
			if i < len(s.Variables)-1 {
				_, err = f.WriteString(s.Variables[i] + ",")
			} else {
				_, err = f.WriteString(s.Variables[i])
			}
		}
		_, err = f.WriteString("]),\n")

		// write the conclusions
		if len(s.Conclusions) > 0 {
			_, err = f.WriteString(",\n")
			writeTerms(s.Conclusions)
			_, err = f.WriteString(".\n\n")
		} else {
			_, err = f.WriteString(".\n\n")
		}
	}

	// Translate the issue schemes of the theory to CHR rules
	// START HERE

	// Translate the assumptions into a Prolog rule, where each
	// assumption is a literal of the body of the rule and the
	// head of the rule is assumptions/0.

	if err != nil {
		return false, errors.New("Could not write the constraint handling rules to a temporary file.")
	} else {
		return true, nil
	}
}

// Infer: Translate a theory into CHR rules and use
// SWI Prolog to construct arguments and add them to the argument graph.
// Does not compute or update labels.
func (ag ArgGraph) Infer() {
	// Translate the theory to CHR in SWI-Prolog and
	// write the output to a temporary file

	f, err := ioutil.TempFile(os.TempDir(), "carneades")
	defer f.Close()
	defer os.Remove(f.Name())
	writeCHR(ag.Theory, ag.Assumptions, f)

	// Call SWI Prolog to evaluate the theory and write arguments
	// to standard out.  Handle SWI-Prolog errors.  Assure termination
	// within given limits (time, stack size, ...)

	// Read the output and construct CAES arguments by instantiating
	// schemes in the theory and adding statements and arguments to
	// the argument graph

	// Clean up, deleting any temporary files
}
