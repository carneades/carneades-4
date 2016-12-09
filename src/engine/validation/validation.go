// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Validator for Carneades Argument Evaluation Structures (CAES)
// The validator checks for syntactic and semantic errors in
// CAES source files, represented using YAML, and produces
// an error report.

package validation

import (
	"fmt"
	"io"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// The category of the problem
type Category int

const (
	IMPORT Category = iota
	STATEMENT
	ISSUE
	ARGUMENT
	ASSUMPTION
	EXPECTEDLABELING
)

// Problem represents an error in some Carneades source file
type Problem struct {
	Category    Category
	Description string
}

// Validate the statements of an argument graph
func validateStatements(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for k, _ := range ag.Statements {
		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{STATEMENT, fmt.Sprint("Key %s is not a term.", k)}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{STATEMENT, fmt.Sprint("Key %s is not a ground atomic formula.", k)}
				problems = append(problems, p)
			}
		}
	}
	return problems
}

// Validate the issues of an argument graph
func validateIssues(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for i1, v1 := range ag.Issues {
		// check that every statement is a position of at most one issue,
		// that is, that no position of an issue is also a position of
		// some other issue
		for _, s1 := range v1.Positions {
			for i2, v2 := range ag.Issues {
				if i2 != i1 { // if not the same issue
					for _, s2 := range v2.Positions {
						if s1 == s2 {
							// found s1 to be a position in both i1 and i2
							p := Problem{ISSUE, fmt.Sprintf("Statement %s is a position of two issues: %s and %s.", s1, i1, i2)}
							problems = append(problems, p)
						}
					}
				}
			}
		}
	}
	return problems
}

// Validate the arguments of an argument graph
func validateArguments(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for id, arg := range ag.Arguments {
		// check that number of parameters matches the number of variables in the scheme
		if len(arg.Parameters) != len(arg.Scheme.Variables) {
			p := Problem{ARGUMENT, fmt.Sprintf("Argument %s does not have the number of parameters declared in its scheme.", id)}
			problems = append(problems, p)
		}

		// check that the number of premises equals the sum of the number of
		// premises and assumptions of the argument's scheme
		if len(arg.Premises) != len(arg.Scheme.Premises)+len(arg.Scheme.Assumptions) {
			p := Problem{ARGUMENT, fmt.Sprintf("Argument %s does not have the number of premises declared in its scheme, including assumptions.", id)}
			problems = append(problems, p)
		} else {
			// Check whether the premises match the scheme
			for i, pr := range arg.Premises {
				t1, ok := terms.ReadString(pr.Stmt.Id)
				if !ok {
					p := Problem{ARGUMENT, fmt.Sprint("Premise %s of argument %s is not a term.", pr.Stmt.Id, id)}
					problems = append(problems, p)
				} else {
					if i < len(arg.Scheme.Premises)-1 {
						// the premise is not an assumption
						t2, _ := terms.ReadString(arg.Scheme.Premises[i])
						// Premises of schemes are checked elsewhere
						_, ok := terms.Match(t1, t2, nil)
						if !ok {
							p := Problem{ARGUMENT, fmt.Sprint("Premise %s of argument %s does not match its premise in the scheme.", pr.Stmt.Id, id)}
							problems = append(problems, p)
						}
					} else {
						// the premise is an assumption
						j := i - len(arg.Scheme.Premises)
						t2, _ := terms.ReadString(arg.Scheme.Assumptions[j])
						// Assumptions of schemes are checked elsewhere
						_, ok := terms.Match(t1, t2, nil)
						if !ok {
							p := Problem{ARGUMENT, fmt.Sprint("Premise %s of argument %s does not match its assumption in the scheme.", pr.Stmt.Id, id)}
							problems = append(problems, p)
						}
					}
				}
			}
		}
		// Check whether the conclusion matches some conclusion of the scheme
		t3, ok := terms.ReadString(arg.Conclusion.Id)
		if !ok {
			p := Problem{ARGUMENT, fmt.Sprint("Conclusion %s of argument %s is not a term.", arg.Conclusion.Id, id)}
			problems = append(problems, p)
		}
		found := false
		for _, s := range arg.Scheme.Conclusions {
			t4, _ := terms.ReadString(s)
			// conclusions of schemes validated elsewhere
			_, ok := terms.Match(t3, t4, nil)
			if ok {
				// matching conclusion found
				found = true
				break
			}
		}
		if !found {
			p := Problem{ARGUMENT, fmt.Sprint("Conclusion %s of argument %s does not match a conclusion of the argument's scheme.", arg.Conclusion.Id, id)}
			problems = append(problems, p)
		}
	}
	return problems
}

// Validate the assumptions of an argument graph
func validateAssumptions(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for k, _ := range ag.Assumptions {
		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{ASSUMPTION, fmt.Sprint("Assumption %s is not a term.", k)}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{ASSUMPTION, fmt.Sprint("Assumption %s is not a ground atomic formula.", k)}
				problems = append(problems, p)
			}
		}
		// Check that there is a statement for the assumption
		_, ok = ag.Statements[k]
		if !ok {
			p := Problem{ASSUMPTION, fmt.Sprint("Assumption %s is not declared to be a statement in the argument graph.", k)}
			problems = append(problems, p)
		}
	}
	return problems
}

// Validate the expected labeling of an argument graph
func validateExpectedLabeling(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for k, _ := range ag.ExpectedLabeling {
		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{EXPECTEDLABELING, fmt.Sprint("In the expected labeling, %s is not a term.", k)}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{EXPECTEDLABELING, fmt.Sprint("In the expected labeling, %s is not a ground atomic formula.", k)}
				problems = append(problems, p)
			}
		}
		// Check that there is a statement for the term
		_, ok = ag.Statements[k]
		if !ok {
			p := Problem{EXPECTEDLABELING, fmt.Sprint("In the expected labeling, %s is not declared to be a statement in the argument graph.", k)}
			problems = append(problems, p)
		}
	}
	return problems
}

// Validate a theory of an argument graph
func validateTheory(ag *caes.ArgGraph) []Problem {
	// START HERE
	return []Problem{}
}

// Validate a Carneades file, represented in YAML
func Validate(file io.Reader) []Problem {
	problems := []Problem{}

	// Validate that YAML file represents an argument graph
	ag, err := yaml.Import(file)
	// TO DO: implement yaml.Validate, which returns a list of all
	// problems found, rather than just the first found.
	if err != nil {
		problems = append(problems, Problem{IMPORT, err.Error()})
		return problems
	}

	problems = append(problems, validateStatements(ag)...)
	problems = append(problems, validateIssues(ag)...)
	problems = append(problems, validateArguments(ag)...)
	problems = append(problems, validateAssumptions(ag)...)
	problems = append(problems, validateExpectedLabeling(ag)...)
	problems = append(problems, validateTheory(ag)...)

	return problems
}
