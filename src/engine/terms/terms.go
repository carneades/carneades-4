// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Logical Terms

package terms

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Type int

const (
	AtomType Type = iota
	BoolType
	IntType
	FloatType
	StringType
	CompoundType
	ListType
	VariableType
)

type Term interface {
	String() string
	Type() Type
}

type Atom string
type Bool bool
type Int int
type Float float64
type String string

type Compound struct {
	Functor string
	Args    []Term
}

type List []Term

type Variable struct {
	Name  string
	index *big.Int
}

func NewVariable(name string) Variable {
	return Variable{Name: name, index: big.NewInt(0)}
}

// This mutable representation of variable bindings
// should be suitable and sufficient implementing Constraint
// Handling Rules, but is presumably not adequate
// for implementing a Prolog-style inference engine,
// which requires a way to backtrack to a previous
// set of bindings
type Bindings map[string]Term // variables are represented as strings

func (t Atom) Type() Type {
	return AtomType
}

func (t Bool) Type() Type {
	return BoolType
}

func (t Int) Type() Type {
	return IntType
}

func (t Float) Type() Type {
	return FloatType
}

func (t String) Type() Type {
	return StringType
}

func (t Compound) Type() Type {
	return CompoundType
}

func (t List) Type() Type {
	return ListType
}

func (t Variable) Type() Type {
	return VariableType
}

func (t Atom) String() string {
	return string(t)
}

func (t Bool) String() string {
	if t {
		return "true"
	} else {
		return "false"
	}
}

func (t Int) String() string {
	return strconv.Itoa(int(t))
}

func (t Float) String() string {
	return fmt.Sprintf("%f", t)
}

func (t String) String() string {
	return string(t)
}

func (t Compound) String() string {
	args := []string{}
	for _, arg := range t.Args {
		args = append(args, arg.String())
	}
	return t.Functor + "(" + strings.Join(args, ",") + ")"
}

func (t List) String() string {
	args := []string{}
	for _, arg := range t {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ",") + "]"
}

func (v Variable) String() string {
	if v.index == nil || v.index.Cmp(big.NewInt(0)) == 0 {
		return v.Name
	} else {
		return v.Name + v.index.String()
	}
}

func (t Compound) Arity() int {
	return len(t.Args)
}

// stream of pointers to big integers for renaming variables
var counter <-chan *big.Int

func init() {
	c := make(chan *big.Int)
	i := big.NewInt(1)
	one := big.NewInt(1)
	go func() {
		for {
			c <- i
			i = new(big.Int).Add(i, one)
		}
	}()
	counter = c
}

func (v Variable) Rename() Variable {
	return Variable{Name: v.Name, index: <-counter}
}

func Equal(t1, t2 Term) bool {
	if t1.Type() != t2.Type() {
		return false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t1 == t2
	case CompoundType:
		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			return false
		}
		for i, _ := range t1.(Compound).Args {
			if !Equal(t1.(Compound).Args[i], t2.(Compound).Args[i]) {
				return false
			}
		}
		return true
	case ListType:
		if len(t1.(List)) != len(t2.(List)) {
			return false
		}
		for i, _ := range t1.(List) {
			if !Equal(t1.(List)[i], t2.(List)[i]) {
				return false
			}
		}
		return true
	case VariableType:
		if t1.(Variable).Name == t2.(Variable).Name &&
			t1.(Variable).index == t2.(Variable).index {
			return true
		}
		return false
	default:
		return false
	}
}

func copyBindings(env Bindings) Bindings {
	result := make(Bindings)
	for v, t := range env {
		result[v] = t
	}
	return result
}

// Match updates the bindings only if the match
// is successful, in which case true is returned.
// One way match, not unification:  variables
// in t1 are bound to terms in t2.
func Match(t1, t2 Term, env Bindings) (ok bool) {
	if t1.Type() != VariableType && t1.Type() != t2.Type() {
		return false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return Equal(t1, t2)
	case CompoundType:
		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			return false
		}
		env2 := copyBindings(env)
		for i, _ := range t1.(Compound).Args {
			ok := Match(t1.(Compound).Args[i], t2.(Compound).Args[i], env2)
			if !ok {
				return false
			}
		}
		// update env with the new bindings
		for v, t := range env2 {
			env[v] = t
		}
		return true
	case ListType:
		if len(t1.(List)) != len(t2.(List)) {
			return false
		}
		env2 := copyBindings(env)
		for i, _ := range t1.(List) {
			ok := Match(t1.(List)[i], t2.(List)[i], env2)
			if !ok {
				return false
			}
		}
		// update env with the new bindings
		for v, t := range env2 {
			env[v] = t
		}
		return true
	case VariableType:
		t3, ok := env[t1.String()]
		if !ok { // variable was not yet bound in env
			env[t1.String()] = t2
			return true
		} else {
			// return true only if the two instances of the variable
			// would be bound to the same term
			if Equal(t2, t3) {
				return true
			} else {
				return false
			}
		}
	default:
		return false
	}
}

func Arity(t Term) int {
	if t.Type() != CompoundType {
		return 0
	}
	return t.(Compound).Arity()
}

func isTriple(t Term) bool {
	return Arity(t) == 2
}

func Functor(t Term) (result string, ok bool) {
	switch t.Type() {
	case AtomType:
		return t.String(), true
	case CompoundType:
		return t.(Compound).Functor, true
	default:
		return result, false
	}
}

// Predicate is a synonym for Functor
func Predicate(t Term) (string, bool) {
	return Functor(t)
}

func Subject(t Term) (result Term, ok bool) {
	if isTriple(t) {
		return t.(Compound).Args[0], true
	}
	return result, false
}

func Object(t Term) (result Term, ok bool) {
	if isTriple(t) {
		return t.(Compound).Args[1], true
	}
	return result, false
}

// Substitute: replace variables in the term t with
// their bindings in the env, if they are bound.
// Follows variable chains, so that if a variable
// is bound to a variable, the second variable is also
// substituted if it is bound in env, recursively.
func Substitute(t Term, env Bindings) Term {
	visited := map[Variable]bool{}

	switch t.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t
	case CompoundType:
		args := []Term{}
		for _, t2 := range t.(Compound).Args {
			args = append(args, Substitute(t2, env))
		}
		return Compound{Functor: t.(Compound).Functor, Args: args}
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l = append(l, Substitute(t2, env))
		}
		return List(l)
	case VariableType:
		result := t
		visited[t.(Variable)] = true
		t2, ok := env[t.String()]
		for ok == true {
			result = t2
			if t2.Type() == VariableType && !visited[t2.(Variable)] {
				t2, ok = env[t2.String()]
				continue
			} else {
				break
			}
		}
		return result
	default:
		return t
	}
}
