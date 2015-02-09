// Dung Abstract Argumentation Frameworks

package dung

import (
	"fmt"
	"github.com/mediocregopher/seq" // Clojure-like persistent data structures
	"reflect"
	"strings"
)

// Argumentation Framework
type AF struct {
	args   []string            // the arguments
	atks   map[string][]string // arguments attacking each key argument
	atkdby map[string][]string // arguments attacked by each argument
}

// Constructs an AF. The atkdby attribute is initialized to nil, since it
// is not needed for all semantics.  When needed use the
// attackedArgs() method.
func NewAF(args []string, atks map[string][]string) AF {
	return AF{args, atks, nil}
}

func (af *AF) String() string {
	d := []string{}
	for arg, attacks := range af.atks {
		attackStrings := []string{}
		for _, attack := range attacks {
			attackStrings = append(attackStrings, attack)
		}
		d = append(d, fmt.Sprintf("%s: [%s]", arg,
			strings.Join(attackStrings, ",")))
	}
	return fmt.Sprintf("{args: [%s], attacks: {%s}}",
		strings.Join(af.args, ", "),
		strings.Join(d, ", "))
}

type ArgSet struct {
	members *seq.Set // Set[string]
}

// work around for bugs in the seq library
func size(S *seq.Set) uint64 {
	if S == nil {
		return 0
	} else {
		return S.Size() - 1
	}
}

func NewArgSet(args ...string) ArgSet {
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
		s = append(s, f.(string))
		f, r, ok = r.FirstRest()
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ","))
}

func EqualArgSets(args1, args2 ArgSet) bool {
	if size(args1.members) != size(args2.members) {
		return false
	}
	S := args1.members.SymDifference(args2.members)
	if size(S) == 0 {
		return true
	}
	return false
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
	S1 := NewArgSet(af1.args...)
	S2 := NewArgSet(af2.args...)
	return EqualArgSets(S1, S2) &&
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
	Labels map[string]Label
}

func NewLabelling() Labelling {
	return Labelling{make(map[string]Label)}
}

func (l Labelling) Get(arg string) Label {
	v, found := l.Labels[arg]
	if found {
		return v
	} else {
		return Undecided
	}
}

func (l Labelling) AsExtension() ArgSet {
	// fmt.Printf("l=%v\n", l)
	S := seq.NewSet()
	for arg, label := range l.Labels {
		if label == In {
			S, _ = S.SetVal(arg)
		}
	}
	return ArgSet{S}
}

func (af *AF) GroundedLabelling() Labelling {
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
	inArgs *seq.Set     // Set[string]
}

func newPLabelling() PLabelling {
	return PLabelling{seq.NewHashMap(), seq.NewSet()}
}

func (pl PLabelling) AsExtension() ArgSet {
	return ArgSet{pl.inArgs}
}

func (pl PLabelling) set(arg string, newLabel Label) PLabelling {
	currentLabel := pl.lookup(arg)

	if currentLabel == newLabel {
		return pl
	}

	newLabels, _ := pl.labels.Set(arg, newLabel)
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

func (pl PLabelling) lookup(arg string) Label {
	l, ok := pl.labels.Get(arg)
	if ok {
		return l.(Label)
	} else {
		return Undecided
	}
}

func (pl PLabelling) toLabelling(af *AF) Labelling {
	l := NewLabelling()
	for _, arg := range af.args {
		l.Labels[arg] = pl.lookup(arg)
	}
	return l
}

// subset returns true if in(L1) is a subset of in(L2)
func (L1 PLabelling) subset(L2 PLabelling) bool {
	if size(L1.inArgs) <= 0 {
		return true
	}
	if size(L1.inArgs) > size(L2.inArgs) {
		return false
	}
	f, r, ok := L1.inArgs.FirstRest()
	for ok {
		_, member := L2.inArgs.GetVal(f.(string))
		if !member {
			return false
		}
		f, r, ok = r.FirstRest()
	}
	return true
}

// x is legally IN iff x is labelled IN and every y
// that attacks x (yRx) is labelled OUT
func (af *AF) legallyIn(x string, L PLabelling) bool {
	if L.lookup(x) != In {
		return false
	}
	for _, y := range af.atks[x] {
		if L.lookup(y) != Out {
			return false
		}
		//if L.lookup(y) == In {
		//	return false
		//}
	}
	return true
}

// an argument is illegally In if it is In but
// not legally In
func (af *AF) illegallyIn(arg string, L PLabelling) bool {
	return L.lookup(arg) == In && !af.legallyIn(arg, L)
}

func (af *AF) illegallyInArgs(L PLabelling) seq.Seq {
	f := func(arg interface{}) bool {
		return af.illegallyIn(arg.(string), L)
	}
	return seq.LFilter(f, L.inArgs)
}

// x is legally OUT iff x is labelled OUT and there is at least
// one y that attacks x and y is labelled IN
func (af *AF) legallyOut(x string, L PLabelling) bool {
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
func (af *AF) illegallyOut(arg string, L PLabelling) bool {
	return L.lookup(arg) == Out && !af.legallyOut(arg, L)
}

// noIllegalInArg: returns true if no argument in the labelling
// is illegally In
func (af *AF) noIllegalInArg(L PLabelling) bool {
	f := func(x interface{}) bool {
		return af.illegallyIn(x.(string), L)
	}
	_, found := seq.Any(f, L.inArgs)
	return !found
}

// An illegally In argument x in L is also super-illegally In
// iff it is attacked by a y that is *legally* In in L,
// or Undecided in L. The input argument is assumed to be illegally In.
func (af *AF) superIllegallyIn(arg string, L PLabelling) bool {
	for _, atk := range af.atks[arg] {
		switch L.lookup(atk) {
		case In:
			if af.legallyIn(atk, L) {
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
// the super illegally In arguments s. The input sequence s is assumed
// to consist of only illegally In arguments.
func (af *AF) superIllegallyInArgs(s seq.Seq, L PLabelling) seq.Seq {
	f := func(arg interface{}) bool {
		return af.superIllegallyIn(arg.(string), L)
	}
	return seq.LFilter(f, s)
}

// Traverse labellings, starting with all arguments labelled out.
// Visit each possible labelling only once
func (af *AF) TraverseLabellings(f func(L PLabelling)) {
	allOut := af.setAllLabelsTo(Out)
	// count := 1

	var subsets func(int, PLabelling)
	subsets = func(i int, L PLabelling) {
		if i == len(af.args) {
			// fmt.Printf("%v. ", count)
			// count++
			f(L)
			return
		}
		subsets(i+1, L)
		subsets(i+1, L.set(af.args[i], In))
	}
	subsets(0, allOut)
}

func (af *AF) setAllLabelsTo(l Label) PLabelling {
	pl := newPLabelling()
	for _, arg := range af.args {
		pl = pl.set(arg, l)
	}
	return pl
}

// The arguments attacked by each argument in the AF
func (af *AF) attackedArgs() {
	attackedBy := make(map[string][]string)
	for _, arg := range af.args {
		attackedBy[arg] = []string{} // initialize to an empty slice
	}
	for arg, s := range af.atks {
		for _, attacker := range s {
			attackedBy[attacker] = append(attackedBy[attacker], arg)
		}
	}
	af.atkdby = attackedBy
}

func (af *AF) transitionStep(L PLabelling, x string) PLabelling {
	L2 := L.set(x, Out)
	if af.illegallyOut(x, L2) {
		L2 = L2.set(x, Undecided)
	}
	for _, z := range af.atkdby[x] {
		if af.illegallyOut(z, L2) {
			L2 = L2.set(z, Undecided)
		}
	}
	return L2
}

func (af *AF) toLabellings(pls []PLabelling) []Labelling {
	labellings := []Labelling{}
	for _, candidate := range pls {
		labellings = append(labellings, candidate.toLabelling(af))
	}
	return labellings
}

// An implemention of the algorithm for preferred semantics from
// Modgil, S. & Caminada, M. Rahwan, I. & Simari, G. R. (Eds.)
// Proof Theories and Algorithms for Abstract Argumentation Frameworks
// Argumentation in Artificial Intelligence, Spinger, 2009, 105-129
func (af *AF) PreferredLabellings() []Labelling {

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
	allIn := af.setAllLabelsTo(In)
	af.attackedArgs()

	subsetOfCandidate := func(L PLabelling) bool {
		for _, L2 := range candidatePLabellings {
			if L.subset(L2) {
				return true
			}
		}
		return false
	}

	var findLabellings func(PLabelling)

	findLabellings = func(L PLabelling) {
		// fmt.Printf("L=%v\n", L.inArgs)
		if visited(ArgSet{L.inArgs}) {
			// fmt.Printf("visited\n")
			return
		}
		if subsetOfCandidate(L) {
			// fmt.Printf("subset\n")
			return
		}

		if af.noIllegalInArg(L) {
			// Add L as a new candidate and remove each labelling from the
			// candidates which is a subset of L.
			s := []PLabelling{L}
			for _, L3 := range candidatePLabellings {
				if !L3.subset(L) {
					s = append(s, L3)
				}
			}
			candidatePLabellings = s
			return
		}

		// else backtrack

		iiArgs := af.illegallyInArgs(L)
		siiArgs := af.superIllegallyInArgs(iiArgs, L)

		if first, _, ok := siiArgs.FirstRest(); ok {
			// try a super illegally In argument, if one exists
			findLabellings(af.transitionStep(L, first.(string)))
		} else {
			// try each illegally In argument
			for f, r, ok := iiArgs.FirstRest(); ok; {
				findLabellings(af.transitionStep(L, f.(string)))
				f, r, ok = r.FirstRest()
			}
		}
	}

	findLabellings(allIn)
	return af.toLabellings(candidatePLabellings)
}

func (af *AF) complete(L PLabelling) bool {
	// fmt.Printf("L=%v\n", L.inArgs)
	// Is atk a member of L?
	conflict := func(arg, atk string) bool {
		_, isMember := L.inArgs.GetVal(atk)
		if isMember {
			// fmt.Printf("%v conflicts with %v:\n", arg, atk)
			return true
		}
		return false
	}
	// Defended against atk by some member of L?
	defended := func(arg, atk string) bool {
		for _, defender := range af.atks[atk] {
			_, isMember1 := L.inArgs.GetVal(defender)
			if isMember1 {
				// fmt.Printf("%v defended against %v by %v\n", arg, atk, defender)
				return true
			}
		}
		// fmt.Printf("not defended against: %v", atk)
		return false
	}
	for f, r, ok := L.inArgs.FirstRest(); ok; f, r, ok = r.FirstRest() {
		arg := f.(string)
		for _, atk := range af.atks[arg] {
			// fmt.Printf("atk=%v\n", atk)
			if conflict(arg, atk) || !defended(arg, atk) {
				return false
			}
		}
	}
	return true
}

func (af *AF) PreferredLabellings3() []Labelling {
	af.attackedArgs()

	candidatePLabellings := []PLabelling{}

	subsetOfCandidate := func(L PLabelling) bool {
		for _, L2 := range candidatePLabellings {
			if L.subset(L2) {
				return true
			}
		}
		return false
	}

	af.TraverseLabellings(func(L PLabelling) {

		if subsetOfCandidate(L) {
			// fmt.Printf("subset\n")
			return
		}

		if af.complete(L) {
			// fmt.Printf("candidate: %v\n", L.inArgs)
			// Add L as a new candidate and remove each labelling from the
			// candidates which is a subset of L.
			s := []PLabelling{L}
			for _, L3 := range candidatePLabellings {
				if !L3.subset(L) {
					s = append(s, L3)
				}
			}
			candidatePLabellings = s
		}
	})

	return af.toLabellings(candidatePLabellings)
}

// Checks whether an argument, arg, is credulous inferred in an argumentation
// framework, af, using preferred semantics.
func (af *AF) CredulouslyInferredPR(arg string) bool {
	s := af.PreferredLabellings3()
	for _, l := range s {
		if l.Get(arg) == In {
			return true
		}
	}
	return false
}

// Checks whether an argument, arg, is skeptically inferred in an
// Argumentation framework, af, using preferred semantics.
func (af *AF) SkepticallyInferredPR(arg string) bool {
	s := af.PreferredLabellings3()
	for _, l := range s {
		if l.Get(arg) != In {
			return false
		}
	}
	return true
}
