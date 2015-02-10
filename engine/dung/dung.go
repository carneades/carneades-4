// Dung Abstract Argumentation Frameworks

package dung

import (
	"fmt"
	"reflect"
	"strings"
)

type Arg string

// Argumentation Framework
type AF struct {
	args   []Arg         // the arguments
	atks   map[Arg][]Arg // arguments attacking each key argument
	atkdby map[Arg][]Arg // arguments attacked by each argument
}

// Constructs an AF. The atkdby attribute is initialized to nil, since it
// is not needed for all semantics.  When needed use the
// attackedArgs() method.
func NewAF(args []Arg, atks map[Arg][]Arg) AF {
	return AF{args, atks, nil}
}

func (af *AF) String() string {
	args := []string{}
	for _, arg := range af.args {
		args = append(args, string(arg))
	}
	d := []string{}
	for arg, attacks := range af.atks {
		attackStrings := []string{}
		for _, attack := range attacks {
			attackStrings = append(attackStrings, string(attack))
		}
		d = append(d, fmt.Sprintf("%s: [%s]", arg,
			strings.Join(attackStrings, ",")))
	}
	return fmt.Sprintf("{args: [%s], attacks: {%s}}",
		strings.Join(args, ", "),
		strings.Join(d, ", "))
}

type ArgSet map[Arg]bool

func NewArgSet(args ...Arg) ArgSet {
	S := make(map[Arg]bool)
	for _, arg := range args {
		S[arg] = true
	}
	return S
}

func (args1 ArgSet) Copy() ArgSet {
	args2 := NewArgSet()
	for k, v := range args1 {
		args2[k] = v
	}
	return args2
}

// Add an argument to an ArgSet, nondestructively.
// Returns the input set if the arg was already a member.
func (args1 ArgSet) Add(arg Arg) ArgSet {
	_, found := args1[arg]
	if !found {
		args2 := args1.Copy()
		args2[arg] = true
		return args2
	}
	return args1
}

func (args ArgSet) Contains(arg Arg) bool {
	_, found := args[arg]
	return found
}

// Removes an element from an ArgSet, nondestructively.
// Returns the input set unchanged if the arg was not a member.
func (args1 ArgSet) Remove(arg Arg) ArgSet {
	if args1.Contains(arg) {
		args2 := args1.Copy()
		delete(args2, arg)
		return args2
	}
	return args1
}

func (args ArgSet) Size() int {
	return len(args)
}

func (args ArgSet) String() string {
	s := []string{}
	for arg, value := range args {
		if value == true {
			s = append(s, string(arg))
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ","))
}

func (args1 ArgSet) Equals(args2 ArgSet) bool {
	if args1.Size() != args2.Size() {
		return false
	}
	// every member of args1 is a member of args2
	for arg1, _ := range args1 {
		if !args2.Contains(arg1) {
			return false
		}
	}
	// every member of args2 is a member of args1
	for arg2, _ := range args2 {
		if !args1.Contains(arg2) {
			return false
		}
	}
	return true
}

func (args1 ArgSet) Subset(args2 ArgSet) bool {
	if args1.Size() <= 0 {
		return true
	}
	if args1.Size() > args2.Size() {
		return false
	}
	// every member of args1 is a member of args2
	for arg, _ := range args1 {
		if !args2.Contains(arg) {
			return false
		}
	}
	return true
}

// EqualArgSetSlices: returns true iff for every ArgSet in l1 there is an
// equal ArgSet in l2
func EqualArgSetSlices(l1, l2 []ArgSet) bool {
	member := func(S1 ArgSet, l []ArgSet) bool {
		for _, S2 := range l {
			if S1.Equals(S2) {
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

func (af1 AF) Equals(af2 AF) bool {
	S1 := NewArgSet(af1.args...)
	S2 := NewArgSet(af2.args...)
	return S1.Equals(S2) &&
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

type Labelling map[Arg]Label

func NewLabelling() Labelling {
	return Labelling(make(map[Arg]Label))
}

func (l Labelling) Get(arg Arg) Label {
	v, found := l[arg]
	if found {
		return v
	} else {
		return Undecided
	}
}

func (l Labelling) AsExtension() ArgSet {
	S := make(map[Arg]bool)
	for arg, label := range l {
		if label == In {
			S[arg] = true
		}
	}
	return ArgSet(S)
}

func (af *AF) GroundedExtension() ArgSet {
	l := NewLabelling()
	var changed bool
	for {
		changed = false
		// Label an argument in if all its attackers are out
		// or out if some attacker is in
		for _, arg := range af.args {
			_, found := l[arg]
			if found {
				continue
			}
			atks := af.atks[arg]
			allOut := true // assumption
			for _, atk := range atks {
				switch l.Get(atk) {
				case In:
					allOut = false
					l[arg] = Out // since an attacker is in
					changed = true
				case Out:
					continue
				case Undecided:
					allOut = false
				}
			}
			if allOut == true {
				l[arg] = In
				changed = true
			}
		}
		if changed == false {
			return l.AsExtension()
		}
	}
}

// Traverse subsets of the args of an AF, starting with the empty set.
// Visit each subset exactly once.
func (af *AF) Traverse(f func(L ArgSet)) {
	allOut := NewArgSet()
	// count := 1

	var subsets func(int, ArgSet)
	subsets = func(i int, L ArgSet) {
		if i == len(af.args) {
			// fmt.Printf("%v. ", count)
			// count++
			f(L)
			return
		}
		subsets(i+1, L)
		subsets(i+1, L.Add(af.args[i]))
	}
	subsets(0, allOut)
}

// The arguments attacked by each argument in the AF
func (af *AF) attackedArgs() {
	attackedBy := make(map[Arg][]Arg)
	for _, arg := range af.args {
		attackedBy[arg] = []Arg{} // initialize to an empty slice
	}
	for arg, s := range af.atks {
		for _, attacker := range s {
			attackedBy[attacker] = append(attackedBy[attacker], arg)
		}
	}
	af.atkdby = attackedBy
}

func (af *AF) complete(L ArgSet) bool {
	// fmt.Printf("L=%v\n", L.inArgs)
	// Is atk a member of L?
	conflict := func(arg, atk Arg) bool {
		if L.Contains(atk) {
			// fmt.Printf("%v conflicts with %v:\n", arg, atk)
			return true
		}
		return false
	}
	// Defended against atk by some member of L?
	defended := func(arg, atk Arg) bool {
		for _, defender := range af.atks[atk] {
			if L.Contains(defender) {
				// fmt.Printf("%v defended against %v by %v\n", arg, atk, defender)
				return true
			}
		}
		// fmt.Printf("not defended against: %v", atk)
		return false
	}
	for arg, _ := range L {
		for _, atk := range af.atks[arg] {
			// fmt.Printf("atk=%v\n", atk)
			if conflict(arg, atk) || !defended(arg, atk) {
				return false
			}
		}
	}
	return true
}

func (af *AF) PreferredExtensions() []ArgSet {
	af.attackedArgs()

	candidates := []ArgSet{}

	subsetOfCandidate := func(L1 ArgSet) bool {
		for _, L2 := range candidates {
			if L1.Subset(L2) {
				return true
			}
		}
		return false
	}

	af.Traverse(func(L ArgSet) {

		if subsetOfCandidate(L) {
			// fmt.Printf("subset\n")
			return
		}

		if af.complete(L) {
			// fmt.Printf("candidate: %v\n", L.inArgs)
			// Add L as a new candidate and remove each labelling from the
			// candidates which is a subset of L.
			s := []ArgSet{L}
			for _, L3 := range candidates {
				if !L3.Subset(L) {
					s = append(s, L3)
				}
			}
			candidates = s
		}
	})

	return candidates
}

// Checks whether an argument, arg, is credulous inferred in an argumentation
// framework, af, using preferred semantics.
func (af *AF) CredulouslyInferredPR(arg Arg) bool {
	s := af.PreferredExtensions()
	for _, args := range s {
		if args.Contains(arg) {
			return true
		}
	}
	return false
}

// Checks whether an argument, arg, is skeptically inferred in an
// Argumentation framework, af, using preferred semantics.
func (af *AF) SkepticallyInferredPR(arg Arg) bool {
	s := af.PreferredExtensions()
	for _, args := range s {
		if !args.Contains(arg) {
			return false
		}
	}
	return true
}
