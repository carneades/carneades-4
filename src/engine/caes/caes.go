// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Carneades Argument Evaluation Structure (CAES)
// This version of CAES supports cyclic argument graphs,
// cumulative arguments and IBIS.

package caes

import (
	"reflect"
	"sort"
	"strings"
)

// The data types are sorted alphabetically

type Argument struct {
	Id          string
	Metadata    Metadata
	Scheme      *Scheme
	Premises    []Premise
	Conclusion  *Statement
	Undercutter *Statement
	Weight      float64 // for storing the evaluated weight
}

type ArgGraph struct {
	Metadata   Metadata
	Issues     []*Issue
	Statements []*Statement
	Arguments  []*Argument
	References map[string]Metadata // key -> metadata
}

// for sorting arguments by property order
type ByProperties struct {
	args  []*Argument
	order []PropertyOrder
}

type Criteria struct {
	HardConstraints []string                  // role names of hard constraints
	SoftConstraints map[string]SoftConstraint // role name to soft constraint
}

type Issue struct {
	Id        string
	Metadata  Metadata
	Positions []*Statement
	Standard  Standard
}

// And IssueScheme is list atomic formulas, which may
// contain schema variables.  Schema variables are denoted
// using Prolog's syntax for variables. Use "..." to indicate
// a variable number of positions, as in this example:
// {"buy(O1)", "...", "buy(On)"}
type IssueScheme []string

type Label int

const (
	Undecided Label = iota
	In
	Out
)

type Labelling map[*Statement]Label

// The keys of a Language map denote the predicate and its arity,
// using Prolog lexical conventions. The values are Go formatting
// strings, for displaying logical formulas in natural language.
// example: {"price/2": "The price of a %v is %v."}
type Language map[string]string

type Metadata map[string]interface{}

type Order int

const (
	Descending Order = iota
	Ascending
)

type Premise struct {
	Stmt *Statement
	Role string // e.g. major, minor
}

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

type Scheme struct {
	Id       string
	Metadata Metadata
	// Each parameter is a schema variable, using
	// Prolog syntax for variables, i.e. identifiers starting
	// a capital letter
	Variables   []string // declaration of schema variables
	Weight      WeighingFunction
	Valid       func(*Argument) bool
	Premises    map[string]string // role names to atomic formulas
	Assumptions map[string]string // CQ names to atomic formulas
	Exceptions  map[string]string // CQ names to atomic formulas
	// Deletions and Guards are extensions for implementing
	// schemes using Constrating Handling Rules (CHR)
	Deletions []string // list of role names of premises to delete
	Guards    []string // list of atomic formulas
	// Note that multiple conclusions are allowed, as in CHR
	Conclusions []string // list of atomic formulas or schema variables
}

type SoftConstraint struct {
	// Factor: relative weight of the constraint, in range of 0.00 to 1.00
	Factor float64
	// NormalizedValues: string to value in range of 0.0 to 1.0
	NormalizedValues map[string]float64
}

// Proof Standards
type Standard int

const (
	DV  Standard = iota // dialectical validity
	PE                  // preponderance of the evidence
	CCE                 // clear and convincing evidence
	BRD                 // beyond reasonable doubt
)

type Statement struct {
	Id       string // an atomic formula, using Prolog syntax
	Metadata Metadata
	Text     string // natural language
	Assumed  bool
	Issue    *Issue      // nil if not at issue
	Args     []*Argument // concluding with this statement
	Label    Label       // for storing the evaluated label
}

type Theory struct { // aka Knowledge Base
	Language          Language
	WeighingFunctions map[string]WeighingFunction
	ArgSchemes        map[string]*Scheme
	Assumptions       []string // list of atomic formula
	IssueSchemes      []*IssueScheme
}

type WeighingFunction func(*Argument, Labelling) float64 // [0.0,1.0]

func NewMetadata() Metadata {
	return make(map[string]interface{})
}

func NewIssue() *Issue {
	return &Issue{
		Metadata:  NewMetadata(),
		Positions: []*Statement{},
		Standard:  PE,
	}
}

func NewStatement() *Statement {
	return &Statement{
		Metadata: NewMetadata(),
		Args:     []*Argument{},
	}
}

func DefaultValidityCheck(*Argument) bool {
	return true
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
			for _, p := range arg.Premises {
				if hc == p.Role && l[p.Stmt] == Out {
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
			weight = weight + (relativeWeight * sc.NormalizedValues[v])
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
		aip := ai.Metadata[p.Property]
		ajp := aj.Metadata[p.Property]
		if reflect.TypeOf(aip) != reflect.TypeOf(ajp) {
			// skip uncomparable values and try sorting by the next property
			continue
		}
		switch aip.(type) {
		case string:
			if aip.(string) == ajp.(string) {
				continue
			}
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
				if aip.(int) < ajp.(int) {
					return true
				}
			case Descending:
				if aip.(int) > ajp.(int) {
					return true
				}
			}
		case float64:
			if aip.(float64) == ajp.(float64) {
				continue
			}
			switch p.Order {
			case Ascending:
				if aip.(float64) < ajp.(float64) {
					return true
				}
			case Descending:
				if aip.(float64) > ajp.(float64) {
					return true
				}
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

// Orders arguments by the metadata properties of the schemes
// instantiated by the arguments. Can be used to model, e.g., Lex Superior
// and Lex Posterior.  If any premise of the argument is Out, the
// argument weights 0.0. If no premise is Out but
// the conclusion of the argument is not at issue, the argument weights 1.0.
// Otherwise all the arguments are the issue are ordered according to
// given PropertyOrder and assigned weights which respect this order.
// To do: Considering caching weights to improve efficiency, since
// currently the arguments are sorted multiple times, once for each
// argument being weighed. Problem: avoiding a memory leak when used the
// cache in a long running service
func SortingWeighingFunction(o []PropertyOrder) WeighingFunction {
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

		// groups is in an ordered list of sets of arguments,
		// representing a partial order. The groups are ordered
		// by increasing strength (ascending order)
		var groups [][]*Argument
		groups = make([][]*Argument, 0, len(args))
		group := []*Argument{}
		equalArgs := genEqualArgsFunction(o)
		for _, a := range args {
			if len(group) > 0 {
				if equalArgs(a, group[0]) {
					group = append(group, a)
				} else {
					// start a new group
					groups = append(groups, group)
					group = []*Argument{a}
				}
			} else {
				// first arg in the group
				group = append(group, a)
			}
		}

		// The weight of an argument depends on its place in the partial
		// order. All arguments in a group (equivalence class) have the
		// same weight. Arguments in the first group will have the weight
		// 1.0. All arguments have some weight greater than 0.0
		// If there are ten groups, arguments in the tenth group
		// will have the weight 0.1

		// Find arg in the partial order and returns its weight.

		n := float64(len(groups))
		var weight float64
		for i, group := range groups {
			weight = ((n - float64(i)) * 1.0) / n
			for _, a := range group {
				if arg == a {
					// found arg
					return weight
				}
			}
		}

		return 0.0 // The argument was not found in some group. Should not happen.
	}
}

var BasicSchemes = map[string]Scheme{
	"linked":     Scheme{Id: "linked", Weight: LinkedWeighingFunction, Valid: DefaultValidityCheck},
	"convergent": Scheme{Id: "convergent", Weight: ConvergentWeighingFunction, Valid: DefaultValidityCheck},
	"cumulative": Scheme{Id: "cumulative", Weight: CumulativeWeighingFunction, Valid: DefaultValidityCheck},
	"factorized": Scheme{Id: "factorized", Weight: FactorizedWeighingFunction, Valid: DefaultValidityCheck},
}

func NewArgument() *Argument {
	return &Argument{
		Metadata: NewMetadata(),
		Premises: []Premise{},
	}
}

func NewArgGraph() *ArgGraph {
	return &ArgGraph{
		Metadata:   NewMetadata(),
		Issues:     []*Issue{},
		Statements: []*Statement{},
		Arguments:  []*Argument{},
		References: make(map[string]Metadata),
	}
}

func (l Label) String() string {
	switch l {
	case In:
		return "in"
	case Out:
		return "out"
	default:
		return "undecided"
	}
}

func NewLabelling() Labelling {
	return Labelling(make(map[*Statement]Label))
}

//func (l Labelling) Get(stmt *Statement) Label {
//	//	v, found := l[stmt]
//	//	if found {
//	//		return v
//	//	} else {
//	//		return Undecided
//	//	}
//	return l[stmt]
//	// ToDo: replace calls to l.Get(s) with l[s] and then delete this method
//}

// Initialize a labelling by making all assumptions In
// other positions of each issue with an assumption Out,
// and unassumed statements without arguments Out.
func (l Labelling) init(ag *ArgGraph) {
	// first make all assumed statements In and all unsupported
	// statements out
	for _, s := range ag.Statements {
		if s.Assumed {
			l[s] = In
		} else if len(s.Args) == 0 {
			l[s] = Out
		}
	}
	// For each issue, if some position is In
	// make the undecided positions Out
	// The resulting issue may be inconsistent, with
	// multiple positions being In, if the assumptions are
	// inconsistent.
	for _, i := range ag.Issues {
		// is some position in?
		somePositionIn := false
		for _, p := range i.Positions {
			if l[p] == In {
				somePositionIn = true
				break
			}
		}
		if somePositionIn {
			for _, p := range i.Positions {
				if l[p] == Undecided {
					l[p] = Out
				}
			}
		}
	}
}

// Apply a labelling to an argument graph by setting
// the label property of each statement in the graph to
// its label in the labelling and by setting the weight
// of each argument in the graph to its evaluated weight
// in the labeling.
func (ag ArgGraph) ApplyLabelling(l Labelling) {
	for _, s := range ag.Statements {
		s.Label = l[s]
	}
	for _, arg := range ag.Arguments {
		arg.Weight = arg.GetWeight(l)
	}
}

// Returns In if the argument has been undercut, Out if the argument
// has no undercutter, the undercutter has no arguments,
// or attempts to undercut the argument it have failed, and Undecided otherwise
func (arg *Argument) Undercut(l Labelling) Label {
	if arg.Undercutter == nil {
		return Out // because there is no undercutter
	} else {
		return l[arg.Undercutter]
	}
}

// An argument is applicable if none of its premises are Undecided and
// its Undercut property is Out. Because arguments can be cumulative, arguments
// with Out premises can be applicable. Out premises affect the weight of an
// argument, not its applicability.
func (arg *Argument) Applicable(l Labelling) bool {
	if arg.Undercut(l) != Out {
		return false
	}
	for _, p := range arg.Premises {
		if l[p.Stmt] == Undecided {
			return false
		}
	}
	return true
}

// Returns the object of statements representing
// predicate-subject-object triples, or the empty string
// if the statement is not a triple.  Triples are assumed
// to be presented using Prolog syntax for atomic formulas:
// predicate(Subject, Object)
// To do: do a better job of checking that the statement
// has the required form.
func (s *Statement) Object() string {
	wff := s.Id
	v := strings.Split(wff, ",")
	if len(v) == 2 {
		str := v[len(v)-1]
		return strings.Trim(str, " )")
	} else {
		return ""
	}
}

func (arg *Argument) PropertyValue(p string, l Labelling) (string, bool) {
	for _, pr := range arg.Premises {
		if p == pr.Role {
			i := pr.Stmt.Issue
			for _, pos := range i.Positions {
				if l[pos] == In {
					return pos.Object(), true
				}
			}
		}
	}
	return "", false
}

// An issue is ready to be resolved if all the arguments of all its positions are
// either undercut or applicable
func (issue *Issue) ReadyToBeResolved(l Labelling) bool {
	for _, position := range issue.Positions {
		for _, arg := range position.Args {
			if !(arg.Undercut(l) == In || arg.Applicable(l)) {
				return false
			}
		}
	}
	return true
}

// Apply a proof standard to check whether w1 is strictly greater than
// w2, where w1 and w2 are argument weights
// Note: DV and PE are indistinguishable in this new model
func (std Standard) greater(w1, w2 float64) bool {
	alpha := 0.5
	beta := 0.3
	switch std {
	case DV, PE:
		return w1 > w2
	case CCE:
		return w1 > w2 && (w1-w2 > alpha)
	case BRD:
		return w1 > w2 && (w1-w2 > alpha) && w2 < beta
	default:
		return false
	}
}

// Apply the proof standard of an issue to each of its positions and update
// the labelling accordingly. After resolving the issue, at most
// one of its positions will be In and all the others will be Out.
// (No position will remain Undecided.) The issue is assumed to be ready to be
// resolved before this method is called.
func (issue *Issue) Resolve(l Labelling) {
	var maxArgWeight = make(map[*Statement]float64)
	for _, p := range issue.Positions {
		maxArgWeight[p] = 0.0
		for _, arg := range p.Args {
			w := arg.GetWeight(l)
			if w > maxArgWeight[p] {
				maxArgWeight[p] = w
			}
		}
	}
	var winner *Statement
PositionLoop:
	for _, p1 := range issue.Positions {
		if maxArgWeight[p1] == 0.0 {
			continue // the winner must be supported by at least one good argument
		}
		winner = p1 // assumption
		// look for another position which is at least as strong as p1
		for _, p2 := range issue.Positions {
			if p2 != p1 &&
				!issue.Standard.greater(maxArgWeight[p1], maxArgWeight[p2]) {
				winner = nil // found an alternative which is at least as good
				continue PositionLoop
			}
		}
		if winner != nil {
			break // winning position found
		}
	}
	// update the labels
	for _, p := range issue.Positions {
		if p == winner {
			l[p] = In
		} else {
			l[p] = Out
		}
	}
}

// A argument has 0.0 weight if it is undercut or inapplicable.
// Otherwise, if a scheme has been applied, it is the weight assigned by
// the evaluator of the scheme.  Otherwise it is the weight assigned
// by the default evaluator, LinkedArgument.
func (arg *Argument) GetWeight(l Labelling) float64 {
	if arg.Undercut(l) == In || !arg.Applicable(l) {
		return 0.0
	} else if arg.Scheme != nil {
		return arg.Scheme.Weight(arg, l)
	} else {
		// apply the default weighing function
		return LinkedWeighingFunction(arg, l)
	}
}

// A statement is supported if it is the conclusion of at least one
// argument with weight greater than 0.0.
func (stmt *Statement) Supported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if arg.GetWeight(l) > 0 {
			return true
		}
	}
	return false
}

// A statement is unsupported if it has no arguments or
// all of its arguments are applicable but none has weight greater than 0
func (stmt *Statement) Unsupported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if !arg.Applicable(l) || arg.GetWeight(l) > 0 {
			return false
		}
	}
	return true
}

// Returns the grounded labelling of an argument graph.
// The argument graph is not modified.
func (ag *ArgGraph) GroundedLabelling() Labelling {
	l := NewLabelling()
	l.init(ag)
	var changed bool
	for {
		changed = false // assumption
		// Try to label Undecided statements
		for _, stmt := range ag.Statements {
			if l[stmt] == Undecided {
				if stmt.Issue == nil {
					if stmt.Supported(l) {
						// make supported nonissues In
						l[stmt] = In
						changed = true
					} else if stmt.Unsupported(l) {
						// make unsupported nonissues Out
						l[stmt] = Out
						changed = true
					}
				} else if stmt.Issue.ReadyToBeResolved(l) {
					// Apply proof standards to label the positions of the issue
					stmt.Issue.Resolve(l)
					changed = true
				}
			}
		}
		// return if a fixpoint has been found
		if !changed {
			return l
		}
	}
}

// An argument graph is inconsistent if more than one position of some
// issue has been assumed true.
func (ag *ArgGraph) Inconsistent() bool {
	for _, issue := range ag.Issues {
		found := false
		for _, p := range issue.Positions {
			if p.Assumed {
				if found {
					// inconsistency, because a previous position
					// of the issue was found to be assumed true
					return false
				} else {
					found = true
				}
			}
		}
	}
	return false
}
