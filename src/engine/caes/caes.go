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
	"fmt"
	"github.com/carneades/carneades-4/src/engine/terms"
	"os"
	"strconv"
)

// The data types are sorted alphabetically

type Argument struct {
	Id          string
	Metadata    Metadata
	Scheme      *Scheme
	Parameters  []string // the values of the scheme variables, in the same order
	Premises    []Premise
	Conclusion  *Statement
	Undercutter *Statement
	Weight      float64 // for storing the evaluated weight
}

type ArgGraph struct {
	Metadata         Metadata
	Issues           map[string]*Issue // id to *Issue
	Statements       map[string]*Statement
	Arguments        map[string]*Argument
	References       map[string]Metadata // key -> metadata
	Theory           *Theory
	Assumptions      map[string]bool  // keys are atomic formulas or statement keys
	ExpectedLabeling map[string]Label // for testing
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

type Premise struct {
	Stmt *Statement
	Role string // e.g. major, minor
}

type Scheme struct {
	Id       string
	Metadata Metadata
	// Each parameter is a schema variable, using
	// Prolog syntax for variables, i.e. identifiers starting
	// a capital letter
	Variables   []string // declaration of schema variables
	Weight      WeighingFunction
	Premises    map[string]string // role names to atomic formulas
	Assumptions map[string]string // role names to atomic formulas
	Exceptions  map[string]string // role names to atomic formulas
	// Deletions and Guards are extensions for implementing
	// schemes using Constrating Handling Rules (CHR)
	Deletions []string // list of role names of premises to delete
	Guards    []string // list of atomic formulas
	// Note that multiple conclusions are allowed, as in CHR
	Conclusions []string // list of atomic formulas or schema variables
}

// Proof Standards
type Standard int

const (
	PE  Standard = iota // dialectical validity
	CCE                 // clear and convincing evidence
	BRD                 // beyond reasonable doubt
)

type Statement struct {
	Id       string // an atomic formula, using Prolog syntax
	Metadata Metadata
	Text     string      // natural language
	Issue    *Issue      // nil if not at issue
	Args     []*Argument // concluding with this statement
	Label    Label       // for storing the evaluated label
}

type Theory struct { // aka Knowledge Base
	Language          Language
	WeighingFunctions map[string]WeighingFunction
	ArgSchemes        map[string]*Scheme
	IssueSchemes      map[string]*IssueScheme
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

func NewArgument() *Argument {
	return &Argument{
		Metadata:   NewMetadata(),
		Premises:   []Premise{},
		Parameters: []string{},
	}
}

func NewTheory() *Theory {
	return &Theory{
		Language:          make(map[string]string),
		WeighingFunctions: make(map[string]WeighingFunction),
		ArgSchemes:        make(map[string]*Scheme),
		IssueSchemes:      make(map[string]*IssueScheme),
	}
}

func NewArgGraph() *ArgGraph {
	return &ArgGraph{
		Metadata:         NewMetadata(),
		Issues:           map[string]*Issue{},
		Statements:       map[string]*Statement{},
		Arguments:        map[string]*Argument{},
		References:       make(map[string]Metadata),
		Assumptions:      map[string]bool{},
		Theory:           NewTheory(),
		ExpectedLabeling: map[string]Label{},
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

// Normalize the string representation of a term
func normalize(s string) string {
	t, ok := terms.ReadString(s)
	if ok {
		return t.String()
	} else {
		return s
	}
}

// Initialize a labelling by making all assumptions In
// other positions of each issue with an assumption Out,
// and unassumed statements without arguments Out.
func (l Labelling) init(ag *ArgGraph) {

	// Normalize the keys of the assumptions table
	m := map[string]bool{}
	for k, v := range ag.Assumptions {
		m[normalize(k)] = v
	}
	ag.Assumptions = m

	// Normalize the keys of the statements table
	m2 := map[string]*Statement{}
	for k, v := range ag.Statements {
		m2[normalize(k)] = v
	}
	ag.Statements = m2

	// Make all assumed statements In and all unsupported
	// statements out
	for _, s := range ag.Statements {
		if ag.Assumptions[normalize(s.Id)] {
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

func (arg *Argument) PropertyValue(p string, l Labelling) (result terms.Term, ok bool) {
	for _, pr := range arg.Premises {
		term1, ok := terms.ReadString(pr.Stmt.Id)
		if ok {
			p2, ok := terms.Predicate(term1)
			if ok && p2 == p {
				if l[pr.Stmt] == In {
					return terms.Object(term1)
				} else {
					i := pr.Stmt.Issue
					if i != nil {
						for _, pos := range i.Positions {
							if l[pos] == In {
								term2, ok := terms.ReadString(pos.Id)
								if ok {
									return terms.Object(term2)
								} else {
									fmt.Fprintf(os.Stderr, "Could not parse: %s\n", pos.Id)
									return result, false
								}
							}
						}
					}
				}
			}
		}
	}
	return result, false
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
// Note: PE are indistinguishable in this new model
func (std Standard) greater(w1, w2 float64) bool {
	alpha := 0.5
	beta := 0.3
	switch std {
	case PE:
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
			if ag.Assumptions[normalize(p.Id)] {
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

// Apply a language to a term to construct a string,
// usually to represent the term in natural language.
func (l Language) Apply(term1 terms.Term) string {
	functor, ok := terms.Functor(term1)
	if !ok {
		return ""
	}
	arity := terms.Arity(term1)

	if arity == 0 {
		return l[functor]
	}

	if arity > 0 {

		args := []interface{}{}
		for _, arg := range term1.(terms.Compound).Args {
			args = append(args, arg.String())
		}

		// Use Sprintf to instantiate the template of the functor in the
		// language spec
		spec := functor + "/" + strconv.Itoa(arity)
		template, ok := l[spec]
		if !ok {
			fmt.Fprintf(os.Stderr, "Functor not defined in the language: %v\n", spec)
			return term1.String()
		}

		return fmt.Sprintf(template, args...)
	}

	return ""
}

func (ag *ArgGraph) InstantiateScheme(id string, parameters []string) {
	genArgId := func() string {
		prefix := "a"
		// Assume exisiting arguments have been given ids using the
		// system. Thus initializing i using the number of existing
		// arguments is likely to produce an unused id.
		i := len(ag.Arguments) + 1
		// increment i until no argument id with this index is already in use
		for _, ok := ag.Arguments[prefix+strconv.Itoa(i)]; ok; _, ok = ag.Arguments[prefix+strconv.Itoa(i)] {
			i++
		}
		return prefix + strconv.Itoa(i)
	}

	if ag.Theory != nil {
		scheme, ok := ag.Theory.ArgSchemes[id]
		if !ok {
			scheme, ok = BasicSchemes[id]
		}
		if ok {
			// bind each schema variable to its value
			if len(scheme.Variables) != len(parameters) {
				fmt.Fprintf(os.Stderr, "Scheme formal (%v) and actual parameters (%v) do not match: %v\n", scheme.Variables, parameters)
				return
			}

			var bindings terms.Bindings
			for i, varName := range scheme.Variables {
				// v := terms.NewVariable(varName)
				t, ok := terms.ReadString(parameters[i])
				if ok {
					bindings = terms.AddBinding(terms.NewVariable(varName), t, bindings)
				} else {
					fmt.Fprintf(os.Stderr, "Could not parse parameter: %v\n", parameters[i])
				}
			}

			// construct the premises and conclusions,
			// adding new statements to the graph
			premises := []Premise{}
			conclusions := []*Statement{}

			addPremises := func(m map[string]string, assumptions bool) {
				for role, p := range m {
					term1, ok := terms.ReadString(p)
					if ok {
						term2 := terms.Substitute(term1, bindings)
						// Leave argument(S,P) premises implicit; Enthymeme!
						pred, ok := terms.Predicate(term2)
						if ok && pred == "argument" {
							continue
						}
						stmt, ok := ag.Statements[term2.String()]
						if !ok {
							s := Statement{Id: term2.String(),
								Text: ag.Theory.Language.Apply(term2)}
							ag.Statements[term2.String()] = &s
							stmt = &s

						}
						if assumptions {
							ag.Assumptions[term2.String()] = true
						}
						premises = append(premises, Premise{Role: role, Stmt: stmt})
					} else {
						fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", p)
					}
				}
			}

			addPremises(scheme.Premises, false)
			// add the assumptions as additional premises
			addPremises(scheme.Assumptions, true)

			// instantiate the conclusions of the scheme
			for _, c := range scheme.Conclusions {
				term1, ok := terms.ReadString(c)
				if ok {
					term2 := terms.Substitute(term1, bindings)
					stmt, ok := ag.Statements[term2.String()]
					if !ok {
						s := Statement{Id: term2.String(),
							Text: ag.Theory.Language.Apply(term2)}
						ag.Statements[term2.String()] = &s
						stmt = &s
					}
					conclusions = append(conclusions, stmt)
				} else {
					fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", c)
				}
			}

			// construct an argument for each conclusion and add
			// the argument to the graph

			var uc Statement // the undercutter
			for _, c := range conclusions {
				argId := genArgId()

				// Construct the undercutter statement and
				// add it to the statements of the graph
				ucid := "¬applicable(" + argId + ")"
				uc = Statement{Id: ucid,
					Text: argId + " is not applicable."}
				ag.Statements[normalize(ucid)] = &uc

				// Construct the argument and add it to the graph
				arg := Argument{Id: argId,
					Scheme:      scheme,
					Parameters:  parameters,
					Premises:    premises,
					Undercutter: &uc,
					Conclusion:  c}
				ag.Arguments[argId] = &arg
				c.Args = append(c.Args, &arg)
			}

			// instantiate the exceptions of the scheme
			exceptions := []*Statement{}
			for _, e := range scheme.Exceptions {
				term1, ok := terms.ReadString(e)
				if ok {
					term2 := terms.Substitute(term1, bindings)
					stmt, ok := ag.Statements[term2.String()]
					if !ok {
						s := Statement{Id: term2.String(),
							Text: ag.Theory.Language.Apply(term2)}
						ag.Statements[term2.String()] = &s
						stmt = &s
					}
					exceptions = append(exceptions, stmt)
				} else {
					fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", e)
				}
			}

			// construct an undercutting argument for each exception
			// and add it to the argument graph
			for _, e := range exceptions {
				argId := genArgId()

				// Construct an undercutter statement (for the undercutter of
				// undercutter!) and add it to the statements of the graph

				ucid := "¬applicable(" + argId + ")"
				uc2 := Statement{Id: ucid,
					Text: argId + " is not applicable."}
				ag.Statements[normalize(ucid)] = &uc2

				// Construct the argument and add it to the graph
				arg := Argument{Id: argId,
					Premises:    []Premise{Premise{Stmt: e}},
					Undercutter: &uc2,
					Conclusion:  &uc}
				ag.Arguments[argId] = &arg
				uc.Args = append(uc.Args, &arg)
			}
		} else {
			fmt.Fprintf(os.Stderr, "No scheme with this id: %v\n", id)
		}
	}
}
