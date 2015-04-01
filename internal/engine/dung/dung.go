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
	args []Arg         // the arguments
	atks map[Arg][]Arg // arguments attacking each key argument
}

func (af *AF) Args() []Arg {
	return af.args
}

func (af *AF) Atks() map[Arg][]Arg {
	return af.atks
}

func NewAF(args []Arg, atks map[Arg][]Arg) AF {
	return AF{args, atks}
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

type Semantics int

const (
	Grounded Semantics = iota
	Complete
	Preferred
	Stable
)

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

// Find the first subset of the args of an AF, starting with the empty set,
// which satifies the given predicate.  The boolean value returned is false
// if no such ArgSet was found.
func (af *AF) Find(pred func(ArgSet) bool) (ArgSet, bool) {
	allOut := NewArgSet()
	var subsets func(int, ArgSet) (ArgSet, bool)
	subsets = func(i int, A ArgSet) (ArgSet, bool) {
		if i == len(af.args) {
			return A, pred(A)
		}
		if S1, ok1 := subsets(i+1, A); ok1 {
			return S1, true
		} else if S2, ok2 := subsets(i+1, A.Add(af.args[i])); ok2 {
			return S2, true
		} else {
			return nil, false
		}
	}
	return subsets(0, allOut)
}

func (af *AF) complete(L ArgSet) bool {
	// If L is not conflict free, return false
	for arg, _ := range L {
		for _, atk := range af.atks[arg] {
			if L.Contains(atk) {
				return false
			}
		}
	}

	// Does L defend against some attacker, atk, by containing an
	// attacker of atk?
	defendsAgainst := func(atk Arg) bool {
		for _, defender := range af.atks[atk] {
			if L.Contains(defender) {
				return true
			}
		}
		return false
	}

	// Is every attacker of arg defended against by L?
	defends := func(arg Arg) bool {
		for _, atk := range af.atks[arg] {
			if !defendsAgainst(atk) {
				return false
			}
		}
		return true
	}

	// F is the characteristic function
	F := func(S ArgSet) ArgSet {
		defendedArgs := []Arg{}
		for _, arg := range af.args {
			if defends(arg) {
				defendedArgs = append(defendedArgs, arg)
			}
		}
		return NewArgSet(defendedArgs...)
	}

	// return true if L is a fixpoint
	return L.Equals(F(L))
}

// An complete extension E is stable iff it attacks every argument not
// a member of E. The tested argument set is assumed to be a complete extension.
func (af *AF) stable(E ArgSet) bool {
	hasAttacker := func(arg Arg) bool {
		for _, atk := range af.atks[arg] {
			if E.Contains(atk) {
				return true
			}
		}
		return false
	}
	for _, arg := range af.args {
		if !E.Contains(arg) && !hasAttacker(arg) {
			return false
		}
	}
	return true
}

func (af *AF) CompleteExtensions() []ArgSet {
	extensions := []ArgSet{}
	af.Traverse(func(A ArgSet) {
		if af.complete(A) {
			extensions = append(extensions, A)
		}
	})
	return extensions
}

func (af *AF) PreferredExtensions() []ArgSet {
	candidates := []ArgSet{}

	subsetOfCandidate := func(L1 ArgSet) bool {
		for _, L2 := range candidates {
			if L1.Subset(L2) {
				return true
			}
		}
		return false
	}

	for _, CE := range af.CompleteExtensions() {
		if subsetOfCandidate(CE) {
			continue
		}
		s := []ArgSet{CE}
		// remove subsets of CE from the list of candidates
		for _, C := range candidates {
			if !C.Subset(CE) {
				s = append(s, C)
			}
		}
		candidates = s
	}

	return candidates
}

func (af *AF) StableExtensions() []ArgSet {
	extensions := []ArgSet{}
	for _, CE := range af.CompleteExtensions() {
		if af.stable(CE) {
			extensions = append(extensions, CE)
		}
	}
	return extensions
}

func (af *AF) CredulouslyInferred(s Semantics, arg Arg) bool {
	switch s {
	case Grounded:
		return af.GroundedExtension().Contains(arg)
	case Complete:
		pred := func(E ArgSet) bool {
			return af.complete(E) && E.Contains(arg)
		}
		if _, found := af.Find(pred); found {
			return true
		} else {
			return false
		}
	case Preferred:
		s := af.PreferredExtensions()
		for _, E := range s {
			if E.Contains(arg) {
				return true
			}
		}
		return false
	case Stable:
		pred := func(E ArgSet) bool {
			return af.complete(E) && af.stable(E) && E.Contains(arg)
		}
		if _, found := af.Find(pred); found {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func (af *AF) SkepticallyInferred(s Semantics, arg Arg) bool {
	memberOfAll := func(extensions []ArgSet) bool {
		for _, E := range extensions {
			if !E.Contains(arg) {
				return false
			}
		}
		return true
	}
	switch s {
	case Grounded:
		return af.GroundedExtension().Contains(arg)
	case Complete:
		return memberOfAll(af.CompleteExtensions())
	case Preferred:
		return memberOfAll(af.PreferredExtensions())
	case Stable:
		return memberOfAll(af.StableExtensions())
	default:
		return false
	}
}

// Returns an extension of the AF with the given semantics, if one exists.
// The boolean value returned is false if no extension exists for the chosen
// semantics.
func (af *AF) SomeExtension(s Semantics) (ArgSet, bool) {
	switch s {
	case Grounded:
		return af.GroundedExtension(), true
	case Complete:
		return af.Find(af.complete)
	case Preferred:
		extensions := af.PreferredExtensions()
		if len(extensions) > 0 {
			return extensions[0], true
		} else {
			return nil, false
		}
	case Stable:
		return af.Find(func(E ArgSet) bool {
			return af.complete(E) && af.stable(E)
		})
	default:
		return nil, false
	}
}
