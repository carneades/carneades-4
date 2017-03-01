// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Carneades Argument Evaluation Structure (CAES)
// This version of CAES supports cyclic argument graphs,
// cumulative arguments and IBIS.

package caes

// types and procedures for weighing arguments

import (
	"fmt"
	"reflect"
	"sort"
	"time"
)

// for sorting arguments by property order
type ByProperties struct {
	args  []*Argument
	order []PropertyOrder
}

type Criteria struct {
	HardConstraints []int                     // premise indices of hard constraints
	SoftConstraints map[string]SoftConstraint // role name to soft constraint
}

type Order int

const (
	Descending Order = iota
	Ascending
)

// PropertyOrder: orders the values of a property so
// that the highest-ranked values appear first when a
// sequence of values is sorted according to the order
// The Order field is ignored, if the order is specified
// explicitly, by providing an ordered slice of values
type PropertyOrder struct {
	Property string
	Order    Order    // implicit ordering (ascending or descending)
	Values   []string // explicit ordering, highest ranked values first
}

type SoftConstraint struct {
	// Factor: relative weight of the constraint, in range of 0.00 to 1.00
	Factor float64
	// NormalizedValues: string to value in range of 0.0 to 1.0
	NormalizedValues map[string]float64
}

var BasicWeighingFunctions = map[string]WeighingFunction{
	"linked":     LinkedWeighingFunction,
	"convergent": ConvergentWeighingFunction,
	"cumulative": CumulativeWeighingFunction,
	"factorized": FactorizedWeighingFunction,
}

// Note: the names of the basic schemes are the same
// as the corresponding basic weighing functions.
var BasicSchemes = map[string]*Scheme{
	"linked":     &Scheme{Id: "linked", Weight: LinkedWeighingFunction},
	"convergent": &Scheme{Id: "convergent", Weight: ConvergentWeighingFunction},
	"cumulative": &Scheme{Id: "cumulative", Weight: CumulativeWeighingFunction},
	"factorized": &Scheme{Id: "factorized", Weight: FactorizedWeighingFunction},
}

func IsBasicScheme(scheme *Scheme) bool {
	for _, s := range BasicSchemes {
		if scheme == s {
			return true
		}
	}
	return false
}

func LinkedWeighingFunction(arg *Argument, l Labelling) float64 {
	for _, p := range arg.Premises {
		if l[p.Stmt] != In {
			return 0.0
		}
	}
	return 1.0
}

func ConvergentWeighingFunction(arg *Argument, l Labelling) float64 {
	for _, p := range arg.Premises {
		if l[p.Stmt] == In {
			return 1.0
		}
	}
	return 0.0
}

func CumulativeWeighingFunction(arg *Argument, l Labelling) float64 {
	n := len(arg.Premises)
	m := 0
	for _, p := range arg.Premises {
		if l[p.Stmt] == In {
			m++
		}
	}
	return float64(m) / float64(n)
}

// Count the number of distinct premises for all arguments of an issue.
func premiseCount(issue *Issue) int {
	m := make(map[string]bool)
	for _, p := range issue.Positions {
		for _, arg := range p.Args {
			for _, pr := range arg.Premises {
				m[pr.Stmt.Text] = true
			}
		}
	}
	return len(m)
}

// A factorized argument, like a linked argument, has no weight unless all
// of its premises are labelled In. If all the premises are in, the weight
// of the argument depends on the number of its premises, compared to
// other arguments about the same issue. The greater the number of premises,
// relative to the other arguments, the greater the weight of the argument.
// See the jogging example for an illustration of its use.  Can be used
// to simulate HYPO-style case-based reasoning.
func FactorizedWeighingFunction(arg *Argument, l Labelling) float64 {
	n := premiseCount(arg.Conclusion.Issue)
	m := 0
	for _, p := range arg.Premises {
		switch l[p.Stmt] {
		case In:
			m++
		case Out:
			return 0.0
		default:
			continue
		}
	}
	return float64(m) / float64(n)
}

func ConstantWeighingFunction(w float64) WeighingFunction {
	return func(arg *Argument, l Labelling) float64 {
		return w
	}
}

func CriteriaWeighingFunction(cs *Criteria) WeighingFunction {
	return func(arg *Argument, l Labelling) float64 {
		// check the hard constraints
		for _, hc := range cs.HardConstraints {
			for i, p := range arg.Premises {
				if hc == i && l[p.Stmt] == Out {
					return 0.0 // KO Criteria
				}
			}
		}
		// All the hard constraints are satisfied.
		// Compute the weighted sum of the soft constraints

		// factorSum is the sum of the factors of all soft constraints.
		// Let f be the factor of some constraint.  The relative weight
		// of the constraint is f/factorSum .
		factorSum := 0.0
		for _, sc := range cs.SoftConstraints {
			factorSum += sc.Factor
		}

		weight := 0.0
		for property, sc := range cs.SoftConstraints {
			v, ok := arg.PropertyValue(property, l)
			if !ok {
				// the argument does have a premise for the specified property
				return 0.0
			}
			// If v is not one of the specified values of a soft constraint
			// the normalized value will be 0.0
			relativeWeight := sc.Factor / factorSum
			weight = weight + (relativeWeight * sc.NormalizedValues[v.String()])
		}
		return weight
	}
}

// Define the methods needed to make ByProperties match
// the sort.Interface interface.
func (s ByProperties) Len() int {
	return len(s.args)
}

func (s ByProperties) Swap(i, j int) {
	s.args[i], s.args[j] = s.args[j], s.args[i]
}

func (s ByProperties) Less(i, j int) bool {
	ai := s.args[i]
	aj := s.args[j]

	// indexOf: returns the index of a string s in a list l
	// or the length of l if s is not in l.
	indexOf := func(s string, l []string) int {
		for i, v := range l {
			if s == v {
				return i
			}
		}
		return len(l) + 1
	}

	for _, p := range s.order {
		aip := ai.Scheme.Metadata[p.Property]
		ajp := aj.Scheme.Metadata[p.Property]
		if reflect.TypeOf(aip) != reflect.TypeOf(ajp) {
			// skip uncomparable values and try sorting by the next property
			continue
		}

		switch aip.(type) {
		case string:
			if aip.(string) == ajp.(string) {
				continue
			}
			// try interpreting the strings as times
			ti, err1 := time.Parse("2006-01-02", aip.(string))
			tj, err2 := time.Parse("2006-01-02", ajp.(string))
			if err1 == nil && err2 == nil {
				// both strings represent times in the supported format
				switch p.Order {
				case Ascending:
					return ti.Before(tj)
				case Descending:
					return ti.After(tj)
				}
			}

			if err1 == nil || err2 == nil {
				// only one of the values represents time
				continue
			}
			// neither string represents time
			switch {
			case len(p.Values) > 0:
				if indexOf(aip.(string), p.Values) < indexOf(ajp.(string), p.Values) {
					return true
				} else {
					continue
				}
			case aip.(string) < ajp.(string):
				return true
			default:
				continue
			}
		case int:
			if aip.(int) == ajp.(int) {
				continue
			}
			switch p.Order {
			case Ascending:
				return aip.(int) < ajp.(int)
			case Descending:
				return aip.(int) > ajp.(int)
			}
		case float64:
			if aip.(float64) == ajp.(float64) {
				continue
			}
			switch p.Order {
			case Ascending:
				return aip.(float64) < ajp.(float64)
			case Descending:
				return aip.(float64) > ajp.(float64)
			}
		default:
			continue
		}
	}
	return false
}

func genEqualArgsFunction(o []PropertyOrder) func(*Argument, *Argument) bool {
	return func(a1, a2 *Argument) bool {
		for _, p := range o {
			a1 := a1.Metadata[p.Property]
			a2 := a2.Metadata[p.Property]
			if reflect.TypeOf(a1) != reflect.TypeOf(a2) {
				// skip uncomparable values and try sorting by the next property
				continue
			}
			switch a1.(type) {
			case string:
				return a1.(string) == a2.(string)
			case int:
				return a1.(int) == a2.(int)
			case float64:
				return a1.(float64) == a2.(float64)
			default:
				continue
			}
		}
		return false
	}
}

// Debugging code
func printGroup(g []*Argument) {
	fmt.Printf("group[")
	for _, a := range g {
		fmt.Printf("%v ", a.Id)
	}
	fmt.Printf("]\n")
}

// PreferenceWeighingFunction: Orders arguments by the metadata properties of
// the schemes instantiated by the arguments. Can be used to model, e.g.,
// Lex Superior and Lex Posterior. Can also be used to simulate value-based
// argumentation frameworks, using value properties and an ordering
// of values. If any premise of the argument is Out, the
// argument weights 0.0. If no premise is Out but
// the conclusion of the argument is not at issue, the argument weights 1.0.
// Otherwise all the arguments are the issue are ordered according to
// given PropertyOrder and assigned weights which respect this order.
// To do: Considering caching weights to improve efficiency, since
// currently the arguments are sorted multiple times, once for each
// argument being weighed. Problem: avoiding a memory leak when used the
// cache in a long running service
func PreferenceWeighingFunction(o []PropertyOrder) WeighingFunction {
	return func(arg *Argument, l Labelling) float64 {
		c := arg.Conclusion
		issue := c.Issue
		w := LinkedWeighingFunction(arg, l)
		if issue == nil || w == 0.0 {
			return w
		}
		// collect the arguments for all positions of the issue
		args := []*Argument{}
		for _, p := range issue.Positions {
			for _, a := range p.Args {
				args = append(args, a)
			}
		}

		// Sort the arguments, so that the weakest arguments
		// appear first in the args list (ascending order)
		sort.Sort(ByProperties{args: args, order: o})
		//		fmt.Printf("sorted args:\n")
		//		for _, arg := range args {
		//			fmt.Printf("\t%v\n", arg.Id)
		//		}

		// groups is in an ordered list of sets of arguments,
		// representing a partial order. The groups are ordered
		// by increasing strength (ascending order)
		var groups [][]*Argument
		group := []*Argument{}
		equalArgs := genEqualArgsFunction(o)
		for _, a := range args {
			if len(group) == 0 {
				// first arg in the group
				group = append(group, a)
			} else {
				if equalArgs(a, group[0]) {
					group = append(group, a)
					printGroup(group)
				} else {
					// start a new group
					groups = append(groups, group)
					group = []*Argument{a}
				}
			}
		}
		groups = append(groups, group)

		// Debugging, to see if the groups have been formed correctly
		//		for i, g := range groups {
		//			fmt.Printf("Group %v. ", i)
		//			printGroup(g)
		//		}

		// The weight of an argument depends on its place in the partial
		// order. All arguments in a group (equivalence class) have the
		// same weight. Arguments in the last group will have the weight
		// 1.0. All arguments have some weight greater than 0.0
		// If there are ten groups, arguments in the first group
		// will have the weight 0.1

		// Find arg in the partial order and returns its weight.

		n := float64(len(groups))
		var weight float64
		for i, group := range groups {
			weight = ((float64(i) + 1.0) * 1.0) / n
			for _, a := range group {
				if arg.Id == a.Id {
					// found arg
					return weight
				}
			}
		}

		return 0.0 // The argument was not found in some group. Should not happen.
	}
}
