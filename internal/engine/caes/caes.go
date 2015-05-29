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
	Value    Label       // for storing evaluation results
}

type Scheme struct {
	Metadata Metadata
	Eval     func(*Argument, Labelling) float64 // [0.0,1.0]
	Valid    func(*Argument) bool
}

type Premise struct {
	Stmt *Statement
	Role string // e.g. major, minor
}

type Argument struct {
	Id         string
	Metadata   Metadata
	Scheme     *Scheme
	Premises   []Premise
	Conclusion *Statement
	NotAppStmt *Statement
	Value      float64 // for storing the evaluation results
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

// Returns In if the argument has been undercut, Out if the argument is
// not at issue or attempts to undercut it have failed, and Undecided otherwise
func (arg *Argument) Undercut(l Labelling) Label {
	if arg.NotAppStmt == nil {
		return Out // applicability not at issue
	} else {
		return l.Get(arg.NotAppStmt)
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
			w := arg.Weight(l)
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

// The default argument evaluator. Handles the argument as
// noncumulative. Returns 1 if all premises are In.
// Returns 0 otherwise. Does not check whether the
// argument has been undercut.
func eval(arg *Argument, l Labelling) float64 {
	for _, p := range arg.Premises {
		if l[p.Stmt] != In {
			return 0.0
		}
	}
	return 1
}

// A argument has 0.0 weight if it is undercut or inapplicable.
// Otherwise it has the weight assigned by the evaluator of its
// scheme, or the weight assigned by the default evaluator, if it has
// no scheme.
func (arg *Argument) Weight(l Labelling) float64 {
	if arg.Undercut(l) == In || !arg.Applicable(l) {
		return 0.0
	} else if arg.Scheme != nil {
		return arg.Scheme.Eval(arg, l)
	} else {
		return eval(arg, l)
	}
}

// An statement is supported if it is the conclusion of at least one
// argument with weight greater than 0.0.
func (stmt *Statement) Supported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if arg.Weight(l) > 0 {
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
