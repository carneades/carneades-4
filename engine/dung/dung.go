// Dung Abstract Argumentation Frameworks

// The algorithms for grounded and preferred semantics used here are from
// Modgil, S. & Caminada, M. Rahwan, I. & Simari, G. R. (Eds.)
// Proof Theories and Algorithms for Abstract Argumentation Frameworks
// Argumentation in Artificial Intelligence, Spinger, 2009, 105-129

package dung

import (
	"fmt"
	"github.com/mediocregopher/seq" // Clojure-like persistent data structures
	"reflect"
	"strings"
)

type Arg interface {
	fmt.Stringer
	seq.Setable
	Id() string // identifier
}

// Argumentation Framework
type AF struct {
	args []Arg         // the arguments
	atks map[Arg][]Arg // arguments attacking each key argument
}

func NewAF(args []Arg, atks map[Arg][]Arg) AF {
	return AF{args, atks}
}

func (af AF) String() string {
	a := []string{}
	for _, arg := range af.args {
		a = append(a, arg.Id())
	}
	d := []string{}
	for arg, attacks := range af.atks {
		attackStrings := []string{}
		for _, attack := range attacks {
			attackStrings = append(attackStrings, attack.Id())
		}
		d = append(d, fmt.Sprintf("%s: [%s]", arg.Id(),
			strings.Join(attackStrings, ",")))
	}
	return fmt.Sprintf("{args: [%s], attacks: {%s}}",
		strings.Join(a, ", "),
		strings.Join(d, ", "))
}

type ArgSet struct {
	members *seq.Set
}

func NewArgSet(args ...Arg) ArgSet {
	S := seq.NewSet()
	for _, arg := range args {
		S, _ = S.SetVal(arg)
	}
	return ArgSet{S}
}

func (e ArgSet) String() string {
	s := []string{}
	f, r, ok := e.members.FirstRest()
	for ok {
		arg := f.(Arg)
		s = append(s, arg.String())
		f, r, ok = r.FirstRest()
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ","))
}

func EqualArgSets(args1, args2 ArgSet) bool {
	S := args1.members.Union(args2.members)
	return S.Size() == args1.members.Size()
}

// EqualArgSetSlices: returns true iff for every ArgSet in l1 there is an
// equal ArgSet in l2
func EqualArgSetSlices(l1, l2 []ArgSet) bool {
	member := func(S1 ArgSet, l []ArgSet) bool {
		for _, S2 := range l {
			if EqualArgSets(S1, S2) {
				return true
			}
		}
		return false
	}
	for _, S1 := range l1 {
		if !member(S1, l2) {
			return false
		}
	}
	return true
}

func EqualAFs(af1, af2 AF) bool {
	S1 := seq.NewSet()
	for _, arg1 := range af1.args {
		S1, _ = S1.SetVal(arg1)
	}
	S2 := seq.NewSet()
	for _, arg2 := range af2.args {
		S2, _ = S2.SetVal(arg2)
	}
	S3 := S1.Union(S2)
	return S3.Size() == S1.Size() &&
		reflect.DeepEqual(af1.atks, af2.atks)
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

type Labelling struct {
	Labels map[Arg]Label
}

func NewLabelling() Labelling {
	return Labelling{make(map[Arg]Label)}
}

func (l Labelling) Get(arg Arg) Label {
	v, found := l.Labels[arg]
	if found {
		return v
	} else {
		return Undecided
	}
}

func (l Labelling) AsExtension() ArgSet {
	S := seq.NewSet()
	for arg, label := range l.Labels {
		if label == In {
			S, _ = S.SetVal(arg)
		}
	}
	return ArgSet{S}
}

func (af AF) GroundedLabelling() Labelling {
	l := NewLabelling()
	var changed bool
	for {
		changed = false
		// Label an argument in if all its attackers are out
		// or out if some attacker is in
		for _, arg := range af.args {
			_, found := l.Labels[arg]
			if found {
				continue
			}
			atks := af.atks[arg]
			allOut := true // assumption
			for _, atk := range atks {
				switch l.Get(atk) {
				case In:
					allOut = false
					l.Labels[arg] = Out // since an attacker is in
					changed = true
				case Out:
					continue
				case Undecided:
					allOut = false
				}
			}
			if allOut == true {
				l.Labels[arg] = In
				changed = true
			}
		}
		if changed == false {
			return l
		}
	}
}

// Persistent, immutable labelling
type PLabelling struct {
	labels *seq.HashMap // argId  -> Label
	inArgs *seq.Set     // set-of(Arg)
}

func newPLabelling() PLabelling {
	return PLabelling{seq.NewHashMap(), seq.NewSet()}
}

func (pl PLabelling) set(arg Arg, newLabel Label) PLabelling {
	currentLabel := pl.lookup(arg)

	if currentLabel == newLabel {
		return pl
	}

	newLabels, _ := pl.labels.Set(arg.Id(), newLabel)
	if currentLabel == In {
		newInArgs, _ := pl.inArgs.DelVal(arg)
		return PLabelling{newLabels, newInArgs}
	} else if newLabel == In {
		newInArgs, _ := pl.inArgs.SetVal(arg)
		return PLabelling{newLabels, newInArgs}
	} else {
		return PLabelling{newLabels, pl.inArgs}
	}
}

func (pl PLabelling) lookup(arg Arg) Label {
	l, ok := pl.labels.Get(arg.Id())
	if ok {
		return l.(Label)
	} else {
		return Undecided
	}
}

func (af AF) PreferredLabellings() []Labelling {

	candidatePLabellings := []PLabelling{}
	closed := []ArgSet{} // Argument sets already visited

	visited := func(x ArgSet) bool {
		for _, y := range closed {
			if EqualArgSets(x, y) {
				return true
			}
		}
		closed = append(closed, x)
		return false
	}

	// allIn is a labelling with all args in af labelled In
	allIn := newPLabelling()
	for _, arg := range af.args {
		allIn = allIn.set(arg, In)
	}

	toLabelling := func(pl PLabelling) Labelling {
		l := NewLabelling()
		for _, arg := range af.args {
			l.Labels[arg] = pl.lookup(arg)
		}
		return l
	}

	// attackedBy: the arguments attacked by each argument in the AF
	attackedBy := make(map[Arg][]Arg)
	for _, arg := range af.args {
		attackedBy[arg] = []Arg{} // initialize to an empty slice
	}
	for arg, s := range af.atks {
		for _, attacker := range s {
			attackedBy[attacker] = append(attackedBy[attacker], arg)
		}
	}

	// subset returns true if in(L1) is a subset of in(L2)
	subset := func(L1, L2 PLabelling) bool {
		if L1.inArgs.Size() <= 0 {
			return true
		}
		if L1.inArgs.Size() > L2.inArgs.Size() {
			return false
		}
		f, r, ok := L1.inArgs.FirstRest()
		for ok {
			arg1 := f.(Arg)
			_, member := L2.inArgs.GetVal(arg1)
			if !member {
				return false
			}
			f, r, ok = r.FirstRest()
		}
		return true
	}

	subsetOfCandidate := func(L PLabelling) bool {
		for _, L2 := range candidatePLabellings {
			if subset(L, L2) {
				return true
			}
		}
		return false
	}

	// x is legally IN iff x is labelled IN and every y
	// that attacks x (yRx) is labelled OUT
	legallyIn := func(x Arg, L PLabelling) bool {
		if L.lookup(x) != In {
			return false
		}
		for _, y := range af.atks[x] {
			if L.lookup(y) != Out {
				return false
			}
		}
		return true
	}

	// an argument is illegally In if it is In but
	// not legally In
	illegallyIn := func(arg Arg, L PLabelling) bool {
		return L.lookup(arg) == In && !legallyIn(arg, L)
	}

	illegallyInArgs := func(L PLabelling) seq.Seq {
		f := func(arg interface{}) bool {
			return illegallyIn(arg.(Arg), L)
		}
		return seq.LFilter(f, L.inArgs)
	}

	// x is legally OUT iff x is labelled OUT and there is at least
	// one y that attacks x and y is labelled IN

	legallyOut := func(x Arg, L PLabelling) bool {
		if L.lookup(x) != Out {
			return false
		}
		for _, y := range af.atks[x] {
			if L.lookup(y) == In {
				return true
			}
		}
		return false
	}

	// an argument is illegally Out if it is Out
	// but not legally out.
	illegallyOut := func(arg Arg, L PLabelling) bool {
		return L.lookup(arg) == Out && !legallyOut(arg, L)
	}

	// noIllegalInArg: returns true if no argument in the labelling
	// is illegally In
	noIllegalInArg := func(L PLabelling) bool {
		f := func(x interface{}) bool {
			return illegallyIn(x.(Arg), L)
		}
		_, found := seq.Any(f, L.inArgs)
		return !found
	}

	// An illegally In argument x in L is also super-illegally In
	// iff it is attacked by a y that is *legally* In in L,
	// or Undecided in L. The input argument is assumed to be illegally In.

	superIllegallyIn := func(arg Arg, L PLabelling) bool {
		for _, atk := range af.atks[arg] {
			switch L.lookup(atk) {
			case In:
				if legallyIn(atk, L) {
					return true
				}
			case Out:
				continue
			case Undecided:
				return true
			}
		}
		return false
	}

	// filters as a sequence of illegally In args in L to return
	// the super illegally In arguments s.  The input sequence s is assumed
	// to consist of only illegally In arguments.
	superIllegallyInArgs := func(s seq.Seq, L PLabelling) seq.Seq {
		f := func(arg interface{}) bool {
			return superIllegallyIn(arg.(Arg), L)
		}
		return seq.LFilter(f, s)
	}

	transitionStep := func(L PLabelling, x Arg) PLabelling {
		L2 := L.set(x, Out)
		if illegallyOut(x, L2) {
			L2 = L2.set(x, Undecided)
		}
		for _, z := range attackedBy[x] {
			if illegallyOut(z, L2) {
				L2 = L2.set(z, Undecided)
			}
		}
		return L2
	}

	var findLabellings func(PLabelling)

	findLabellings = func(L PLabelling) {
		if visited(ArgSet{L.inArgs}) {
			return
		}
		// fmt.Printf("L=%v\n", L.inArgs)
		if subsetOfCandidate(L) {
			// fmt.Printf("subset\n")
			return
		}

		if noIllegalInArg(L) {
			for i, L3 := range candidatePLabellings {
				// if L3’s In arguments are a subset of L's
				//  In arguments then remove L3 from the candidate labellings
				// fmt.Printf("removing: %v\n", L3.inArgs)
				if subset(L3, L) {
					candidatePLabellings =
						append(candidatePLabellings[:i],
							candidatePLabellings[i+1:]...)
				}
			}
			// add L as a new candidate
			// fmt.Printf("candidate: %v\n", L.inArgs)
			candidatePLabellings = append(candidatePLabellings, L)
			return
		}

		// else backtrack

		iiArgs := illegallyInArgs(L)
		// fmt.Printf("illegals: %v\n", iiArgs)
		siiArgs := superIllegallyInArgs(iiArgs, L)
		// fmt.Printf("super illegals: %v\n", siiArgs)
		f := func(arg interface{}) bool {
			findLabellings(transitionStep(L, arg.(Arg)))
			return true
		}

		if first, _, ok := siiArgs.FirstRest(); ok {
			// fmt.Printf("s\n")
			findLabellings(transitionStep(L, first.(Arg)))
		} else {
			// fmt.Printf("i\n")
			seq.All(f, iiArgs)
		}
	}

	findLabellings(allIn)

	labellings := []Labelling{}
	for _, candidate := range candidatePLabellings {
		labellings = append(labellings, toLabelling(candidate))
	}
	// fmt.Printf("\n")
	return labellings
}

// Checks whether an argument, arg, is credulous inferred in an argumentation
// framework, af, using preferred semantics.
func (af AF) CredulouslyInferredPR(arg Arg) bool {
	s := af.PreferredLabellings()
	for _, l := range s {
		if l.Get(arg) == In {
			return true
		}
	}
	return false
}

// Checks whether an argument, arg, is skeptically inferred in an
// Argumentation framework, af, using preferred semantics.
func (af AF) SkepticallyInferredPR(arg Arg) bool {
	s := af.PreferredLabellings()
	for _, l := range s {
		if l.Get(arg) != In {
			return false
		}
	}
	return true
}
