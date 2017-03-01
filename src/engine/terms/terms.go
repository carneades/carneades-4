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

type Vars []Variable

type Term interface {
	OccurVars() Vars
	String() string
	Type() Type
}

type Atom string
type Bool bool
type Int int
type Float float64
type String string

type EnvMap map[int][]Bindings

type Compound struct {
	Functor           string
	Id                *big.Int
	Prio              int
	EMap              *EnvMap
	occurVars         Vars
	identifyOccurVars bool
	IsActive          bool
	Args              []Term
}

type List []Term

type Variable struct {
	Name  string
	index *big.Int
}

func NewVariable(name string) Variable {
	return Variable{Name: name, index: big.NewInt(0)}
}

func NewCompound(f string, args []Term) Term {
	c := Compound{Functor: f, Args: args}
	return Term(c)
}

func CopyCompound(c Compound) (c1 Compound) {
	c1 = Compound{
		Functor:           c.Functor,
		Id:                c.Id,
		Prio:              c.Prio,
		EMap:              c.EMap,
		occurVars:         c.occurVars,
		identifyOccurVars: c.identifyOccurVars,
		IsActive:          c.IsActive,
		Args:              []Term{}}
	args := []Term{}
	for _, a := range c.Args {
		args = append(args, a)
	}
	c1.Args = args
	return c1
}

type Bindings *BindEle
type BindEle struct {
	Var  Variable
	T    Term
	Next Bindings
}

func Normalize(s string) string {
	t, ok := ReadString(s)
	if ok {
		return t.String()
	} else {
		return s
	}
}

func AddBinding(v Variable, t Term, b Bindings) Bindings {
	// fmt.Printf(" Add Binding %s-%d == %s \n", v.String(), v.index, t.String())
	return &BindEle{Var: v, T: t, Next: b}
}

func GetBinding(v Variable, b Bindings) (t Term, ok bool) {
	// fmt.Printf(" GetBinding %s-%d %v \n", v.String(), v.index, b)
	name := v.Name
	id := v.index
	if id == nil {
		for b != nil {
			if b.Var.Name == name && b.Var.index == nil {
				// fmt.Printf(" Binding found %s \n", b.T.String())
				return b.T, true
			}
			b = b.Next
			// fmt.Printf(" NextBinding %s %v \n", name, b)
		}
	} else {

		for b != nil {
			if b.Var.Name == name && b.Var.index != nil && b.Var.index.Cmp(id) == 0 {
				// fmt.Printf(" Binding found %s \n", b.T.String())
				return b.T, true
			}
			b = b.Next
			// fmt.Printf(" NextBinding %s %v \n", name, b)
		}
	}
	// fmt.Printf(" Binding not found \n")
	return nil, false
}

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
	if t.Prio != 0 {
		prio := t.Prio
		f := t.Functor
		switch f {
		case "||", "&&", "in", "or", "div", "mod":
			f = " " + f + " "
		}
		switch t.Arity() {
		case 1:
			if t.Args[0].Type() == CompoundType {
				prio1 := t.Args[0].(Compound).Prio
				if prio1 == 0 {
					return f + t.Args[0].String()
				}
				if prio1 < prio {
					return f + "(" + t.Args[0].String() + ")"
				}
			}
			return f + t.Args[0].String()
		case 2:
			if t.Args[0].Type() == CompoundType {
				prio1 := t.Args[0].(Compound).Prio
				if t.Args[1].Type() == CompoundType {
					prio2 := t.Args[1].(Compound).Prio
					switch {
					case prio1 < prio && prio2 < prio:
						return "(" + t.Args[0].String() + ") " + f + " (" + t.Args[1].String() + ")"
					case prio1 < prio:
						return "(" + t.Args[0].String() + ") " + f + " " + t.Args[1].String()
					case prio2 < prio:
						return t.Args[0].String() + " " + f + " (" + t.Args[1].String() + ")"
					default:
						return t.Args[0].String() + f + t.Args[1].String()
					}
				} else {
					if prio1 < prio {
						return "(" + t.Args[0].String() + ") " + f + " " + t.Args[1].String()
					} else {
						return t.Args[0].String() + f + t.Args[1].String()
					}
				}
			} else if t.Args[1].Type() == CompoundType && t.Args[1].(Compound).Prio < prio {
				return t.Args[0].String() + " " + f + " (" + t.Args[1].String() + ")"
			}
			return t.Args[0].String() + f + t.Args[1].String()

		}
	}
	// Prio == 0
	if t.Functor == "|" {
		args := []string{}
		var oldarg Term = nil
		for _, arg := range t.Args {
			if oldarg != nil {
				args = append(args, oldarg.String())
			}
			oldarg = arg
		}
		return "[" + strings.Join(args, ",") + " | " + oldarg.String() + "]"
	}
	args := []string{}
	for _, arg := range t.Args {
		args = append(args, arg.String())
	}
	return t.Functor + "(" + strings.Join(args, ",") + ")"
}

func (t List) String() string {
	var v Term = nil
	args := []string{}

	for _, arg := range t {
		if arg.Type() == CompoundType && arg.(Compound).Functor == "|" {
			v = arg.(Compound).Args[0]
		} else {
			args = append(args, arg.String())
		}
	}
	if v != nil {
		return "[" + strings.Join(args, ", ") + " | " + v.String() + "]"
	}
	return "[" + strings.Join(args, ", ") + "]"
}

func (v Variable) String() string {
	if v.index == nil || v.index.Cmp(big.NewInt(0)) == 0 {
		return v.Name
	} else {
		return v.Name + v.index.String()
	}
}

func (t Atom) OccurVars() Vars {
	return nil
}

func (t Bool) OccurVars() Vars {
	return nil
}

func (t Int) OccurVars() Vars {
	return nil
}

func (t Float) OccurVars() Vars {
	return nil
}

func (t String) OccurVars() Vars {
	return nil
}

func (t Compound) OccurVars() Vars {
	if t.identifyOccurVars {
		return t.occurVars
	}
	occur := Vars{}
	for _, t2 := range t.Args {
		occur = append(occur, t2.OccurVars()...)
	}
	t.occurVars = occur
	t.identifyOccurVars = true
	return t.occurVars
}

func (t List) OccurVars() Vars {
	occur := Vars{}
	for _, t2 := range t {
		occur = append(occur, t2.OccurVars()...)
	}
	return occur
}

func (t Variable) OccurVars() Vars {
	return Vars{t}
}

func (t Compound) Arity() int {
	return len(t.Args)
}

func Ground(t Term, b Bindings) bool {
	switch t.Type() {
	case CompoundType:
		if t.(Compound).Arity() > 0 {
			for _, t2 := range t.(Compound).Args {
				if !Ground(t2, b) {
					return false
				}
			}
		}
		return true
	case ListType:
		for _, t2 := range t.(List) {
			if !Ground(t2, b) {
				return false
			}
		}
		return true
	case VariableType:
		t2, ok := GetBinding(t.(Variable), b)
		if ok {
			return Ground(t2, b)
		} else {
			return false
		}
	default:
		return true
	}
}

func AtomicFormula(t Term) bool {
	switch t.Type() {
	case BoolType:
		return true
	case AtomType:
		return true
	case CompoundType:
		return true
	default:
		return false
	}
}

// stream of pointers to big integers for renaming variables
var Counter <-chan *big.Int

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
	Counter = c
}

func (v Variable) Rename() Variable {
	return Variable{Name: v.Name, index: <-Counter}
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
			//		if t1.(Compound).Prio != 3 && t2.(Compound).Prio != 3 { return false }
			// 	return EqualCompare(t1.(Compound).Functor, )
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

/*func copyBindings(env Bindings) Bindings {
	result := make(Bindings)
	for v, t := range env {
		result[v] = t
	}
	return result
} */

// Match updates the bindings only if the match
// is successful, in which case true is returned.
// One way match, not unification:  variables
// in t1 are bound to terms in t2.
//func Match(t1, t2 Term, env Bindings) (ok bool) {
//	ok, _ = Match1(t1, t2, env)
//	return ok
//}

func Match(t1, t2 Term, env Bindings) (env2 Bindings, ok bool) {
	if t1.Type() != VariableType && t1.Type() != t2.Type() {
		return env, false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return env, Equal(t1, t2)
	case CompoundType:
		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			return env, false
		}
		env2 := env
		for i, _ := range t1.(Compound).Args {
			env2, ok = Match(t1.(Compound).Args[i], t2.(Compound).Args[i], env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*		for v, t := range env2 {
				env[v] = t
			} */
		return env, true
	case ListType:
		if len(t1.(List)) != len(t2.(List)) {
			return env, false
		}
		env2 := env
		for i, _ := range t1.(List) {
			env2, ok = Match(t1.(List)[i], t2.(List)[i], env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*	for v, t := range env2 {
			env[v] = t
		} */
		return env, true
	case VariableType:
		t3, ok := GetBinding(t1.(Variable), env)
		if !ok { // variable was not yet bound in env
			env = AddBinding(t1.(Variable), t2, env)
			return env, true
		} else {
			// return true only if the two instances of the variable
			// would be bound to the same term
			if Equal(t2, t3) {
				return env, true
			} else {
				return env, false
			}
		}
	default:
		return env, false
	}
}

// Unify head-term with goal-term
func Unify(head, goal Term, env Bindings) (env2 Bindings, ok bool) {
	return Unify1(head, goal, true /* head vars of a rule */, Vars{}, env)
}

func Unify1(t1, t2 Term, renaming bool, visited Vars, env Bindings) (env2 Bindings, ok bool) {
	isHead := renaming // if t1 is a head-term
	t1Type := t1.Type()
	if t1Type == VariableType {
		renaming = false
		for t1Type == VariableType {
			visited = append(visited, t1.(Variable))
			t3, ok := GetBinding(t1.(Variable), env)
			if ok {
				isHead = false
				t1 = t3
				t1Type = t1.Type()
			} else {
				break
			}

		}

	}
	t2Type := t2.Type()
	for t2Type == VariableType {
		visited = append(visited, t2.(Variable))
		t3, ok := GetBinding(t2.(Variable), env)
		if ok {
			t2 = t3
			t2Type = t2.Type()
		} else {
			break
		}
	}
	if t1Type == VariableType {
		if t2Type == VariableType {
			if t1.(Variable).Name == t2.(Variable).Name &&
				(t1.(Variable).index.Cmp(t2.(Variable).index) == 0 ||
					(t1.(Variable).index == nil && t2.(Variable).index == nil)) {
				// Var == Var
				return env, true
			} else {
				if isHead {
					// Var1 != Var2 , no occur-check
					env2 = AddBinding(t1.(Variable), t2, env)
					return env2, true
				} else {
					return env2, false
				}
			}
		}
		if checkOccur(visited, t2, env) {
			return nil, false
		}
		env2 = AddBinding(t1.(Variable), t2, env)
		return env2, true
	}
	if t2Type == VariableType {
		return env2, false
		//		if checkOccur(visited, t1, env) {
		//			return nil, false
		//		}
		//		// to do: if renaming { rename vars in t1 }
		//		if renaming { renameVars(t1)}
		//		env2 = AddBinding(t2.(Variable), t1, env)
		//		return env2, true
	}
	/*
		if t1Type == CompoundType && t1.(Compound).Functor == "|" && t2Type == ListType {
			args := t1.(Compound).Args
			arg0 := args[0]
			arg1 := args[1]
			l0 := len(arg0.(List))
			t2List := t2.(List)
			l2 := len(t2List)
			if arg0.Type() == ListType && l2 >= l0 {
				for i, ele := range arg0.(List) {
					// Unify1(t1, t2 Term, renaming bool, visited Vars, env Bindings) (env2 Bindings, ok bool)
					env2, ok = Unify1(ele, t2List[i], renaming, visited, env)
					if !ok {
						return env, false
					}
					env = env2
				}
				if l2 == l0 {
					env2, ok = Unify1(arg1, List{}, renaming, visited, env)
				} else {
					env2, ok = Unify1(arg1, t2List[len(arg0.(List)):], renaming, visited, env)
				}
				if !ok {
					return env, false
				}
				env = env2
			} else {
				return env, false
			}
		}
	*/
	if t1Type != t2Type {
		return env, false
	}
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return env, Equal(t1, t2)
	case CompoundType:
		if t1.(Compound).Functor != t2.(Compound).Functor ||
			t1.(Compound).Arity() != t2.(Compound).Arity() {
			return env, false
		}
		env2 := env
		for i, _ := range t1.(Compound).Args {
			env2, ok = Unify1(t1.(Compound).Args[i], t2.(Compound).Args[i], renaming, visited, env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*		for v, t := range env2 {
				env[v] = t
			} */
		return env, true
	case ListType:
		lent1 := len(t1.(List))
		lent2 := len(t2.(List))
		if lent1 == 0 {
			if lent2 == 0 {
				return env, true
			} else {
				return env, false
			}
		}
		lent1m1 := lent1 - 1
		last := t1.(List)[lent1m1]
		if last.Type() == CompoundType && last.(Compound).Functor == "|" {
			if lent2 < lent1m1 {
				return env, false
			}
			env2 := env
			for i := 0; i < lent1m1; i++ {
				env2, ok = Unify1(t1.(List)[i], t2.(List)[i], renaming, visited, env2)
				if !ok {
					return env, false
				}
			}
			v := last.(Compound).Args[0]
			if lent2 == lent1m1 {
				env2, ok = Unify1(v, List{}, renaming, visited, env2)
			} else {
				env2, ok = Unify1(v, t2.(List)[lent1m1:], renaming, visited, env2)
			}
			if !ok {
				return env, false
			}
			env = env2
			return env, true
		}
		if lent1 != lent2 {
			return env, false
		}
		env2 := env
		// for i, _ := range t1.(List) {
		for i := 0; i < lent1; i++ {
			env2, ok = Unify1(t1.(List)[i], t2.(List)[i], renaming, visited, env2)
			if !ok {
				return env, false
			}
		}
		// update env with the new bindings
		env = env2
		/*	for v, t := range env2 {
			env[v] = t
		} */
		return env, true
	default:
		return env, false
	}
}

func checkOccur(v Vars, t Term, env Bindings) bool {

	for _, termv := range t.OccurVars() {
		for _, visitv := range v {
			if termv.Name == visitv.Name && termv.index.Cmp(visitv.index) == 0 {
				return true
			}
		}
		t2, ok := GetBinding(termv, env)
		if ok {
			for _, termv := range t2.OccurVars() {
				for _, visitv := range v {
					if termv.Name == visitv.Name && termv.index.Cmp(visitv.index) == 0 {
						return true
					}
				}
			}
		}

	}
	return false
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
		return Compound{Functor: t.(Compound).Functor, Id: t.(Compound).Id,
			Prio: t.(Compound).Prio, Args: args}
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l = append(l, Substitute(t2, env))
		}
		return List(l)
	case VariableType:
		result := t
		visited[t.(Variable)] = true
		t2, ok := GetBinding(t.(Variable), env)
		for ok == true {
			result = t2
			if t2.Type() == VariableType && !visited[t2.(Variable)] {
				t2, ok = GetBinding(t2.(Variable), env)
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

// Substitute: replace variables in the term t with
// their bindings in the Build-In environment (BIVarEqTerm),
// if they are bound.
// Follows variable chains, so that if a variable
// is bound to a variable, the second variable is also
// substituted if it is bound in env, recursively.
func SubstituteBiEnv(t Term, biEnv Bindings) (Term, bool) {
	ok := false
	visited := map[Variable]bool{}

	switch t.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t, ok
	case CompoundType:
		args := []Term{}
		for _, t2 := range t.(Compound).Args {
			a, ok2 := SubstituteBiEnv(t2, biEnv)
			ok = ok || ok2
			args = append(args, a)
		}
		return Compound{Functor: t.(Compound).Functor, Id: t.(Compound).Id,
			Prio: t.(Compound).Prio, Args: args}, ok
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l1, ok2 := SubstituteBiEnv(t2, biEnv)
			ok = ok || ok2
			l = append(l, l1)
		}
		return List(l), ok
	case VariableType:

		t2, ok2 := GetBinding(t.(Variable), biEnv)
		ok = ok || ok2

		for ok2 == true {
			t = t2
			if t2.Type() == VariableType && !visited[t2.(Variable)] {
				t2, ok2 = GetBinding(t2.(Variable), biEnv)
				continue
			} else {
				break
			}
		}
		return t, ok
	default:
		return t, ok
	}
}

// Substitute: replace variables in the term t with
// their bindings in the env or in the Build-In environment (BIVarEqTerm),
// if they are bound. If their are not bound rename
// the body-varaible of the rule (very late renaming).
// Follows variable chains, so that if a variable
// is bound to a variable, the second variable is also
// substituted if it is bound in env, recursively.
func RenameAndSubstitute(t Term, idx *big.Int, env Bindings) Term {
	visited := map[Variable]bool{}

	switch t.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t
	case CompoundType:
		args := []Term{}
		for _, t2 := range t.(Compound).Args {
			args = append(args, RenameAndSubstitute(t2, idx, env))
		}
		return Compound{Functor: t.(Compound).Functor, Id: t.(Compound).Id,
			Prio: t.(Compound).Prio, Args: args}
	case ListType:
		l := []Term{}
		for _, t2 := range t.(List) {
			l = append(l, RenameAndSubstitute(t2, idx, env))
		}
		return List(l)
	case VariableType:

		t2, ok := GetBinding(t.(Variable), env)
		if !ok {
			// very late variable renaming
			t = Variable{Name: t.(Variable).Name, index: idx}
			visited[t.(Variable)] = true
			t2, ok = GetBinding(t.(Variable), env)
			if !ok {
				return t
			}
		}
		for ok == true {
			t = t2
			if t2.Type() == VariableType && !visited[t2.(Variable)] {
				t2, ok = GetBinding(t.(Variable), env)
			} else {
				break
			}
		}
		return t
	default:
		return t
	}
}
