// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Inference of arguments (aka argument construction or generation)
// using Constraint Handling Rules (CHR)

package caes

import (
	// "fmt"
	"strings"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// maximum number of rule (scheme) applications when deriving arguments
const MAXRULEAPPS = 100000

// ArgDesc: Structure describing an argument instantiating an
// argument scheme. Represented in Prolog as argument(Scheme,Values)
type ArgDesc struct {
	Scheme string   // id of the scheme
	Values []string // values instantiating the variables of the scheme
}

// Returns true with an ArgDesc in the second result if the
// string s matches argument(S,P)
func termToArgDesc(s string) (bool, *ArgDesc) {
	if s == "" {
		return false, nil
	}
	t1, ok := terms.ReadString(s)
	if !ok {
		return false, nil
	}
	S := terms.NewVariable("S")
	P := terms.NewVariable("P")
	t2 := terms.NewCompound("argument", []terms.Term{S, P})
	var bindings terms.Bindings
	bindings, ok = terms.Match(t2, t1, bindings)
	if !ok {
		return false, nil
	}
	scheme, _ := terms.GetBinding(S, bindings)
	parms, _ := terms.GetBinding(P, bindings)

	// convert the parms from a list of terms to a list of strings
	l := []string{}
	if parms.Type() == terms.ListType {
		for _, t := range parms.(terms.List) {
			l = append(l, t.String())
		}
	}

	return true, &ArgDesc{scheme.String(), l}
}

// Infer: Translate a theory into CHR rules and use
// SWI Prolog to construct arguments and add them to the argument graph.
// Does not compute or update labels.  If the theory is syntactically incorrect
// and thus cannot be parsed by the CHR inference engine, an error is returned
// and argument graph is left unchanged. If all goes well, the argument
// graph is updated and nil is returned.
func (ag *ArgGraph) Infer() error {
	if len(ag.Theory.ArgSchemes) != 0 {
		// rb := TheoryToSWIRulebase(ag.Theory)
		rb := TheoryToRuleStore(ag.Theory)

		// Create an index of the previous arguments constructed
		// to avoid constructing equivalent instanstiations of schemes
		// and to allow the inference engine to construct undercutters
		prevArgs := map[string]bool{}
		for _, a := range ag.Arguments {
			if a != nil {
				prevArgs["argument("+a.Scheme.Id+",["+strings.Join(a.Parameters, ",")+"])"] = true
			}
		}

		// The goals are initialized with a "go" goal, for use by schemes
		// with no premises, which are translated into CHR rules with a dummy "go"
		// term in their heads.
		goals := []string{"go"}

		// The actual goals in the query with the CHR inference
		// engine consist of the union of the assumptions of the argument graph
		// and the assumptions for each of the previous arguments
		for _, k := range ag.Assumptions {
			goals = append(goals, k)
		}
		for k, _ := range prevArgs {
			goals = append(goals, k)
		}

		success, store, err := rb.Infer(goals, MAXRULEAPPS)
		if err != nil {
			return err
		}
		if !success {
			return nil
		}

		// fmt.Printf("store:\n")
		for _, s := range store {
			// fmt.Printf("   %s\n", s)
			// If the term does not represent an argument already in the graph
			// then, if the term does represent an argument, use it to
			// add an argument to the argument graph by instantiating the
			// argumentation scheme applied.
			if _, exists := prevArgs[s]; !exists {
				isArg, a := termToArgDesc(s)
				if isArg {
					ag.InstantiateScheme(a.Scheme, a.Values)
					prevArgs[s] = true
				}
			}
		}
	}
	// Use issue schemes of the theory to derive or update the issues
	// of the argument graph

	if ag.Theory.IssueSchemes != nil {
		for issue, patterns := range ag.Theory.IssueSchemes {
			err := ag.makeIssue(issue, *patterns)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
