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
	"os"
	"strconv"

	"github.com/carneades/carneades-4/src/engine/terms"
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
	Assumptions      []string         // atomic formulas or statement keys
	assums           map[string]bool  // map representation of the assumptions
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
	Roles       []string // Roles[i] is the role name of Premises[i]
	Premises    []string // list of atomic formulas
	Assumptions []string // list of atomic formulas
	Exceptions  []string // list of atomic formulas
	// Deletions and Guards are extensions for implementing
	// schemes using Constrating Handling Rules (CHR)
	Deletions []string // list of atomic formulas
	Guards    []string // list of atomic formulas
	// Note that multiple conclusions are allowed, as in CHR
	Conclusions []string // list of atomic formulas or schema variables
}

// Proof Standards
type Standard int

const (
	PE  Standard = iota // preponderance of the evidence
	CCE                 // clear and convincing evidence
	BRD                 // beyond reasonable doubt
)

type Statement struct {
	Id       string // a ground atomic formula, using Prolog syntax
	Metadata Metadata
	Text     string      // natural language
	Issue    *Issue      // nil if not at issue
	Args     []*Argument // concluding with this statement
	Label    Label       // for storing the evaluated label
}

// A Rulebase is a set of constraint handling rules.
// https://en.wikipedia.org/wiki/Constraint_Handling_Rules
// The strings in the methods of the Rulebase interface
// denote terms, represented using Prolog syntax.
type Rulebase interface {
	// Infer returns true if the goals succeed and false if they fail.
	// The list of strings returned represents the constraint store
	Infer(goals []string, max int) (bool, []string, error)
	// AddRule adds a constraint handling rule to the rulebase
	AddRule(name string, keep []string, delete []string, guard []string, body []string) error
}

type Theory struct { // aka Knowledge Base
	Language          Language
	WeighingFunctions map[string]WeighingFunction
	ArgSchemes        []*Scheme
	IssueSchemes      map[string]*IssueScheme
	schemeIndex       map[string]*Scheme
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
		ArgSchemes:        []*Scheme{},
		IssueSchemes:      make(map[string]*IssueScheme),
		schemeIndex:       make(map[string]*Scheme),
	}
}

func (t *Theory) InitSchemeIndex() {
	for _, s := range t.ArgSchemes {
		t.schemeIndex[s.Id] = s
	}
}

func NewArgGraph() *ArgGraph {
	return &ArgGraph{
		Metadata:         NewMetadata(),
		Issues:           map[string]*Issue{},
		Statements:       map[string]*Statement{},
		Arguments:        map[string]*Argument{},
		References:       make(map[string]Metadata),
		Assumptions:      []string{},
		assums:           make(map[string]bool),
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

func SliceToMap(s []string) map[string]bool {
	result := map[string]bool{}
	for _, k := range s {
		result[k] = true
	}
	return result
}

// Initialize a labelling by making all assumptions In
// other positions of each issue with an assumption Out,
// and unassumed statements without arguments Out.
func (l Labelling) init(ag *ArgGraph) {
	// Normalize the assumptions
	s := []string{}
	for _, k := range ag.Assumptions {
		s = append(s, terms.Normalize(k))
	}
	ag.Assumptions = s

	// Map representation of the assumptions
	// to make membership test efficient
	ag.assums = SliceToMap(ag.Assumptions)

	// Normalize the keys of the statements table
	m2 := map[string]*Statement{}
	for k, v := range ag.Statements {
		m2[terms.Normalize(k)] = v
	}

	ag.Statements = m2

	// Make all assumed statements In
	for _, s := range ag.Statements {
		if ag.assums[terms.Normalize(s.Id)] {
			l[s] = In
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

// An argument is applicable if its Undercut property is In, or the undercutter
// is Out and none of the premises are Undecided. Because arguments can be
// cumulative, arguments with Out premises can be applicable. Out premises or
// In undercutters affect the weight of an argument, not its applicability.
// The labels of the premises are irrelevant for applicability if the
// undercutter is In.
func (arg *Argument) Applicable(l Labelling) bool {
	if arg.Undercut(l) == In {
		return true
	}
	if arg.Undercut(l) == Undecided {
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
// undercut or applicable
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
// applicable argument with weight greater than 0.0.
func (stmt *Statement) Supported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if arg.Applicable(l) && arg.GetWeight(l) > 0 {
			return true
		}
	}
	return false
}

// A statement is unsupported if it has no arguments or
// all of its arguments are applicable and weigh 0.0 or less
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
				if stmt.Unsupported(l) {
					// make unsupported statements Out
					l[stmt] = Out
					changed = true
				} else if stmt.Issue == nil && stmt.Supported(l) {
					// make supported nonissues In
					l[stmt] = In
					changed = true
				} else if stmt.Issue != nil && stmt.Issue.ReadyToBeResolved(l) {
					// Apply proof standards to label the positions of issues
					// ready to be resolved
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
			if ag.assums[terms.Normalize(p.Id)] {
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
		return l[functor+"/0"]
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
		scheme, ok := ag.Theory.schemeIndex[id]
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

			addPremises := func(l []string, assumptions bool) {
				for i, p := range l {
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
							ag.AddAssumption(term2.String())
						}
						role := ""
						if i < len(scheme.Roles) {
							role = scheme.Roles[i]
						}
						premises = append(premises, Premise{Role: role, Stmt: stmt})
					} else {
						fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", p)
					}
				}
			}

			addPremises(scheme.Premises, false)
			addPremises(scheme.Deletions, false)
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
			// the argument to the graph.  All these arguments
			// share the same undercutter.

			if len(conclusions) > 0 {
				var uc Statement    // the undercutter
				argId := genArgId() // pseudo-argument id, representing the set of all arguments constructed by the scheme
				// To Do: this is a hack.  Clean this up by allowing arguments
				// to have multiple conclusions. This is a fairly big change, requiring
				// modifications to the import and export translators and the YAML and JSON representation of arguments.

				// Construct the undercutter statement and
				// add it to the statements of the graph
				ucid := "undercut(" + argId + ")"
				uc = Statement{Id: ucid,
					Text: argId + " is undercut."}
				ag.Statements[terms.Normalize(ucid)] = &uc
				for _, c := range conclusions {
					// Construct an argument for each conclusion and add it to the graph
					argId := genArgId()
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

					ucid := "undercut(" + argId + ")"
					uc2 := Statement{Id: ucid,
						Text: argId + " is undercut."}
					ag.Statements[terms.Normalize(ucid)] = &uc2

					// Construct the argument and add it to the graph
					arg := Argument{Id: argId,
						Premises:    []Premise{Premise{Stmt: e}},
						Undercutter: &uc2,
						Conclusion:  &uc}
					ag.Arguments[argId] = &arg
					uc.Args = append(uc.Args, &arg)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "No scheme with this id: %v\n", id)
		}
	}
}

// Add a statement id to the assumptions of an argument graph.  Assumes a
// statement with this id has already been declared in the graph
func (ag *ArgGraph) AddAssumption(stmt_id string) {
	ag.assums[stmt_id] = true
	ag.Assumptions = append(ag.Assumptions, stmt_id)
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

	// skip issue schemes with less than two patterns
	if len(patterns) < 2 {
		fmt.Fprintf(os.Stderr, "Issue scheme with less than two patterns: %v\n", issueScheme)
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
			isEnumeration := len(patterns) == 3 && patterns[1] == "..."

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
