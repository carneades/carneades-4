// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Carneades Argument Evaluation Structure (CAES)
// This version of CAES supports cyclic argument graphs,
// cumulative arguments and IBIS.

package caes

// Proof Standards
type Standard int

const (
	DV  Standard = iota // dialectical validity
	PE                  // preponderance of the evidence
	CCE                 // clear and convincing evidence
	BRD                 // beyond reasonable doubt
)

type Metadata map[string]interface{}

type Issue struct {
	Id        string
	Metadata  Metadata
	Positions []*Statement
	Standard  Standard
}

type Statement struct {
	Id       string
	Metadata Metadata
	Text     string // natural language
	Assumed  bool
	Issue    *Issue      // nil if not at issue
	Args     []*Argument // concluding with this statement
	Label    Label       // for storing the evaluated label
}

type Scheme struct {
	Id       string
	Metadata Metadata
	Eval     func(*Argument, Labelling) float64 // [0.0,1.0]
	Valid    func(*Argument) bool
}

func DefaultValidityCheck(*Argument) bool {
	return true
}

func LinkedArgument(arg *Argument, l Labelling) float64 {
	for _, p := range arg.Premises {
		if l.Get(p.Stmt) != In {
			return 0.0
		}
	}
	return 1.0
}

func ConvergentArgument(arg *Argument, l Labelling) float64 {
	for _, p := range arg.Premises {
		if l.Get(p.Stmt) == In {
			return 1.0
		}
	}
	return 0.0
}

func CumulativeArgument(arg *Argument, l Labelling) float64 {
	n := len(arg.Premises)
	m := 0
	for _, p := range arg.Premises {
		if l.Get(p.Stmt) == In {
			m++
		}
	}
	return float64(m) / float64(n)
}

// Find the maximum number of premises of the arguments about positions
// of the given issue.
func maxPremises(issue *Issue) int {
	m := 0
	for _, p := range issue.Positions {
		for _, arg := range p.Args {
			n := len(arg.Premises)
			if n > m {
				m = n
			}
		}
	}
	return m
}

// A factorized argument, like a linked argument, has no weight unless all
// of its premises are labelled In. If all the premises are in, the weight
// of the argument depends on the number of its premises, compared to
// other arguments about the same issue. The greater the number of premises,
// relative to the other arguments, the greater the weight of the argument.
// See the jogging example for an illustration of its use.  Can be used
// to simulate HYPO-style case-based reasoning.
func FactorizedArgument(arg *Argument, l Labelling) float64 {
	n := maxPremises(arg.Conclusion.Issue)
	m := 0
	for _, p := range arg.Premises {
		switch l.Get(p.Stmt) {
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

var BasicSchemes = map[string]Scheme{
	"linked":     Scheme{Id: "linked", Eval: LinkedArgument, Valid: DefaultValidityCheck},
	"convergent": Scheme{Id: "convergent", Eval: ConvergentArgument, Valid: DefaultValidityCheck},
	"cumulative": Scheme{Id: "cumulative", Eval: CumulativeArgument, Valid: DefaultValidityCheck},
	"factorized": Scheme{Id: "factorized", Eval: FactorizedArgument, Valid: DefaultValidityCheck},
}

type Premise struct {
	Stmt *Statement
	Role string // e.g. major, minor
}

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

type Label int

const (
	Out Label = iota
	In
	Undecided
)

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

type Labelling map[*Statement]Label

func NewLabelling() Labelling {
	return Labelling(make(map[*Statement]Label))
}

func (l Labelling) Get(stmt *Statement) Label {
	v, found := l[stmt]
	if found {
		return v
	} else {
		return Undecided
	}
}

// Initialize a labelling by making all assumptions In
// other positions of each issue with an assumption Out,
// and unassumed statements without arguments Out.
// The argument graph is assumed to be consistent. If it
// is not consistent, the resulting labelling will depend
// on the order in which the assumptions are initialized.
func (l Labelling) init(ag *ArgGraph) {
	for _, s := range ag.Statements {
		if s.Assumed {
			l[s] = In
			if s.Issue != nil {
				for _, p := range s.Issue.Positions {
					if p != s {
						l[p] = Out
					}
				}
			}
		} else if len(s.Args) == 0 {
			l[s] = Out
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
		s.Label = l.Get(s)
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
		return l.Get(arg.Undercutter)
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
		if l.Get(p.Stmt) == Undecided {
			return false
		}
	}
	return true
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
		return w1 > w2 && (w2-w1 > alpha) && w2 < beta
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
		return arg.Scheme.Eval(arg, l)
	} else {
		return LinkedArgument(arg, l) // the default argument evaluator
	}
}

// An statement is supported if it is the conclusion of at least one
// argument with weight greater than 0.0.
func (stmt *Statement) Supported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if arg.GetWeight(l) > 0 {
			return true
		}
	}
	return false
}

// Returns the grounded labelling of an argument graph. The argument
// graph is assumed to be consistent. The argument graph is not modified.
func (ag *ArgGraph) GroundedLabelling() Labelling {
	l := NewLabelling()
	l.init(ag)
	var changed bool
	for {
		changed = false // assumption
		// Try to label Undecided statements
		for _, stmt := range ag.Statements {
			if l.Get(stmt) == Undecided {
				if stmt.Issue == nil {
					if stmt.Supported(l) {
						// Make supported nonissues In
						l[stmt] = In
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
