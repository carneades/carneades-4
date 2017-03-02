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
	"strconv"
	"strings"

	"github.com/carneades/carneades-4/src/engine/caes"
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
	LANGUAGE
	SCHEME  // Argument Scheme
	ISCHEME // Issue Scheme
)

func (c Category) String() string {
	switch c {
	case IMPORT:
		return "import"
	case STATEMENT:
		return "statement"
	case ISSUE:
		return "issue"
	case ARGUMENT:
		return "argument"
	case ASSUMPTION:
		return "assumption"
	case EXPECTEDLABELING:
		return "expected labeling"
	case LANGUAGE:
		return "language"
	case SCHEME:
		return "argument scheme"
	case ISCHEME:
		return "issue scheme"
	default:
		return ""
	}

}

// Problem represents an error in some Carneades source file
type Problem struct {
	Category    Category
	Id          string // id of the affected object, if available
	Description string // brief description of the problem, without referencing the category or object id
	Expression  string // the affected part of the object with the problem.
}

// Validate the statements of an argument graph
func validateStatements(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	for k, _ := range ag.Statements {
		// Check that the key is a term
		t, ok := terms.ReadString(k)
		if !ok {
			p := Problem{STATEMENT, "", "key not a term", k}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{STATEMENT, "", "key not a ground atomic formula", k}
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
							p := Problem{ISSUE, i1, "statement is a position of two issues", s1.Id}
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
		if caes.IsBasicScheme(arg.Scheme) {
			return problems
		}
		// check that number of parameters matches the number of variables in the scheme
		if len(arg.Parameters) != len(arg.Scheme.Variables) {
			p := Problem{ARGUMENT, id, "number of parameters not the same as declared in the scheme", ""}
			problems = append(problems, p)
		}

		// check that the number of premises equals the sum of the number of
		// premises and assumptions of the argument's scheme
		if len(arg.Premises) != len(arg.Scheme.Premises)+len(arg.Scheme.Assumptions) {
			p := Problem{ARGUMENT, id, "number of premises not the same as declared in the scheme, including assumptions", ""}
			problems = append(problems, p)
		} else {
			// Check whether the premises match the scheme
			for i, pr := range arg.Premises {
				t1, ok := terms.ReadString(pr.Stmt.Id)
				if !ok {
					p := Problem{ARGUMENT, id, "premise is not a term", pr.Stmt.Id}
					problems = append(problems, p)
				} else {
					if i < len(arg.Scheme.Premises)-1 {
						// the premise is not an assumption
						t2, _ := terms.ReadString(arg.Scheme.Premises[i])
						// Premises of schemes are checked elsewhere
						_, ok := terms.Match(t1, t2, nil)
						if !ok {
							p := Problem{ARGUMENT, id, "premise does not match the scheme", pr.Stmt.Id}
							problems = append(problems, p)
						}
					} else {
						// the premise is an assumption
						j := i - len(arg.Scheme.Premises)
						t2, _ := terms.ReadString(arg.Scheme.Assumptions[j])
						// Assumptions of schemes are checked elsewhere
						_, ok := terms.Match(t1, t2, nil)
						if !ok {
							p := Problem{ARGUMENT, id, "premise does not match its assumption in the scheme", pr.Stmt.Id}
							problems = append(problems, p)
						}
					}
				}
			}
		}
		// Check whether the conclusion matches some conclusion of the scheme
		t3, ok := terms.ReadString(arg.Conclusion.Id)
		if !ok {
			p := Problem{ARGUMENT, id, "conclusion is not a term.", arg.Conclusion.Id}
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
			p := Problem{ARGUMENT, id, "conclusion does not match the scheme.", arg.Conclusion.Id}
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
			p := Problem{ASSUMPTION, "", "not a term", k}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{ASSUMPTION, "", "not a ground atomic formula", k}
				problems = append(problems, p)
			}
		}
		// Check that there is a statement for the assumption
		//		_, ok = ag.Statements[k]
		//		if !ok {
		//			p := Problem{ASSUMPTION, id, "not declared to be a statement", k}
		//			problems = append(problems, p)
		//		}
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
			p := Problem{EXPECTEDLABELING, "", "not a term", k}
			problems = append(problems, p)
		} else {
			// Check that the term is a ground atomic formula
			var b terms.Bindings // empty environment
			if !terms.AtomicFormula(t) || !terms.Ground(t, b) {
				p := Problem{EXPECTEDLABELING, "", "not a ground atomic formula", k}
				problems = append(problems, p)
			}
		}
		// Check that there is a statement for the term
		_, ok = ag.Statements[k]
		if !ok {
			p := Problem{EXPECTEDLABELING, "", "not declared to be a statement", k}
			problems = append(problems, p)
		}
	}
	return problems
}

func validateLanguage(l caes.Language) []Problem {
	isPred := func(s string) bool {
		t, ok := terms.ReadString(s)
		return ok && t.Type() == terms.AtomType
	}
	// count the number of verbs in a format string
	verbCount := func(s string) int {
		n := 0
		for _, c := range s {
			if c == '%' {
				n++
			}
		}
		return n
	}
	problems := []Problem{}
	for k, v := range l {
		// check that the key has the form predicate/arity
		l := strings.Split(k, "/")
		if len(l) != 2 {
			p := Problem{LANGUAGE, "", "does not have the form predicate/arity", k}
			problems = append(problems, p)
		} else if !isPred(l[0]) {
			p := Problem{LANGUAGE, k, "not a predicate symbol", l[0]}
			problems = append(problems, p)
		} else {
			var n int
			_, err := fmt.Sscanf(l[1], "%d", &n)
			if err != nil {
				p := Problem{LANGUAGE, k, "non-integer arity", l[1]}
				problems = append(problems, p)
			} else if verbCount(v) != n {
				p := Problem{LANGUAGE, k, "format string has incorrect number of placeholders (verbs)", l[1]}
				problems = append(problems, p)
			}
		}
	}
	return problems
}

// Validate that each string in a list represents a logical variable
func validateVariables(s *caes.Scheme) []Problem {
	l := s.Variables
	problems := []Problem{}
	for _, v := range l {
		t, ok := terms.ReadString(v)
		if !ok || t.Type() != terms.VariableType {
			p := Problem{SCHEME, s.Id, "not a variable", v}
			problems = append(problems, p)
		}
	}
	return problems
}

// Validate an argumentation scheme s against a lanuage l
func validateScheme(s *caes.Scheme, l caes.Language) []Problem {
	// Checks if s is declared as a variable in the scheme.
	declaredVariable := func(s2 string) bool {
		for _, v := range s.Variables {
			if s2 == v {
				return true
			}
		}
		return false
	}

	problems := validateVariables(s)

	validateAtom := func(atm string, kind string) {
		t, ok := terms.ReadString(atm)
		if !ok {
			p := Problem{SCHEME, s.Id, "not a term", atm}
			problems = append(problems, p)
		} else {
			var key string
			var varOrBool bool = false
			switch t.Type() {
			case terms.BoolType:
				varOrBool = true
			case terms.AtomType:
				key = t.String() + "/" + "0"
			case terms.CompoundType:
				key = t.(terms.Compound).Functor + "/" + strconv.Itoa(len(t.(terms.Compound).Args))
			case terms.VariableType:
				varOrBool = true
				if kind != "conclusion" {
					p := Problem{SCHEME, s.Id, fmt.Sprintf("%s may not be a variable", kind), atm}
					problems = append(problems, p)
				}
			default:
				p := Problem{SCHEME, s.Id, "not an atomic formula", atm}
				problems = append(problems, p)
			}
			// Check that the predicate of the atom, with the given arity, has been declared in the language

			if key != "¬/1" && !varOrBool {
				// negation operator, variables and booleans need not be declared
				_, ok := l[key]
				if !ok {
					p := Problem{SCHEME, s.Id, "predicate not declared in the language", key}
					problems = append(problems, p)
				}
			}
			// Check that all variables in the atom have been declared in the scheme
			vars := t.OccurVars()
			for _, v := range vars {
				if !declaredVariable(v.Name) {
					p := Problem{SCHEME, s.Id, "variable not declared in the scheme", v.Name}
					problems = append(problems, p)
				}
			}
		}
	}

	for _, atm := range s.Premises {
		validateAtom(atm, "premise")
	}
	for _, atm := range s.Assumptions {
		validateAtom(atm, "assumption")
	}
	for _, atm := range s.Exceptions {
		validateAtom(atm, "exception")
	}
	for _, atm := range s.Deletions {
		validateAtom(atm, "deletion")
	}
	for _, atm := range s.Guards {
		validateAtom(atm, "guard")
	}
	for _, atm := range s.Conclusions {
		validateAtom(atm, "conclusion")
	}
	return problems
}

// Validate the argumentation schemes of a theory
func validateSchemes(theory *caes.Theory) []Problem {
	problems := []Problem{}
	ids := map[string]bool{}
	for _, s := range theory.ArgSchemes {
		if ids[s.Id] {
			p := Problem{SCHEME, s.Id, "duplicate scheme id", ""}
			problems = append(problems, p)
		} else {
			ids[s.Id] = true
		}
		problems = append(problems, validateScheme(s, theory.Language)...)
	}
	return problems
}

func validateIssueSchemes(theory *caes.Theory) []Problem {
	problems := []Problem{}
	l := theory.Language

	validatePattern := func(sid string, t terms.Term) {
		// allow patterns to be variables
		if t.Type() == terms.VariableType {
			return
		}
		var key string
		switch t.Type() {
		case terms.AtomType:
			key = t.String() + "/" + "0"
		case terms.CompoundType:
			key = t.(terms.Compound).Functor + "/" + strconv.Itoa(len(t.(terms.Compound).Args))
		default:
			p := Problem{ISCHEME, sid, "pattern is not an atomic formula", t.String()}
			problems = append(problems, p)
		}

		// Check that the predicate of the atom, with the given arity, has been declared in the language
		if key == "¬/1" {
			// Negation is built-in and doesn't need to be declared
			return
		} else {
			_, ok := l[key]
			if !ok {

				p := Problem{ISCHEME, sid, "predicate of pattern not declared in the language", key}
				problems = append(problems, p)
			}
		}
	}

	for sid, is := range theory.IssueSchemes {
		s := *is
		if len(s) < 2 {
			p := Problem{ISCHEME, sid, "fewer than two patterns", ""}
			problems = append(problems, p)
		}
		// Check that each string in the list of the scheme represents an atom, or
		// has three elements, where the first an last are atoms and the
		// second is "...".  If it is an atom, also check that its predicate
		// is defined in the language.
		if len(s) == 3 && s[1] == "..." {
			for _, i := range []int{0, 2} {
				t, ok := terms.ReadString(s[i])
				if !ok {
					p := Problem{ISCHEME, sid, "pattern is not a term", s[i]}
					problems = append(problems, p)
				} else {
					validatePattern(sid, t)
				}
			}
		} else {
			for i, _ := range s {
				t, ok := terms.ReadString(s[i])
				if !ok {
					p := Problem{ISCHEME, sid, "pattern is not a term", s[i]}
					problems = append(problems, p)
				} else {
					validatePattern(sid, t)
				}
			}
		}
	}
	return problems
}

// Validate a theory of an argument graph
func validateTheory(ag *caes.ArgGraph) []Problem {
	problems := validateLanguage(ag.Theory.Language)
	schemeProblems := validateSchemes(ag.Theory)
	if len(schemeProblems) > 0 {
		problems = append(problems, schemeProblems...)
	}
	issueSchemeProblems := validateIssueSchemes(ag.Theory)
	if len(issueSchemeProblems) > 0 {
		problems = append(problems, issueSchemeProblems...)
	}
	return problems
}

// Validate an argument graph
func Validate(ag *caes.ArgGraph) []Problem {
	problems := []Problem{}
	problems = append(problems, validateStatements(ag)...)
	problems = append(problems, validateIssues(ag)...)
	problems = append(problems, validateArguments(ag)...)
	problems = append(problems, validateAssumptions(ag)...)
	problems = append(problems, validateExpectedLabeling(ag)...)
	problems = append(problems, validateTheory(ag)...)

	return problems
}
