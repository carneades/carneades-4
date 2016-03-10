// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Constraint Handling Rules

package chr

import (
	"fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	"math/big"
	// "strconv"
	"strings"
)

var QueryVars Vars

var QueryStore List

var CHRstore store

var BuiltInStore store

type argCHR struct {
	atomArg  map[string]cList
	boolArg  cList
	intArg   cList
	floatArg cList
	strArg   cList
	compArg  map[string]cList
	listArg  cList
	varArg   cList
	noArg    cList
}

type store map[string]*argCHR

var nextRuleId int = 0

func InitStore() {
	chrCounter = big.NewInt(0)
	nextRuleId = 0
	CHRruleStore = []*chrRule{}
	CHRstore = store{}
	BuiltInStore = store{}
	QueryStore = List{}
	QueryVars = Vars{}
}

func NewArgCHR() *argCHR {
	return &argCHR{atomArg: map[string]cList{},
		boolArg: cList{}, intArg: cList{}, floatArg: cList{}, strArg: cList{},
		compArg: map[string]cList{}, listArg: cList{}, varArg: cList{}, noArg: cList{}}
}

func addGoal1(g *Compound, s store) {
	aArg, ok := s[g.Functor]
	if !ok {
		aArg = NewArgCHR()
		s[g.Functor] = aArg
	}
	args := g.Args
	if len(args) == 0 {
		aArg.noArg = append(aArg.noArg, g)
		return
	}
	arg0 := args[0]
	switch arg0.Type() {
	case AtomType:
		cl, ok := aArg.atomArg[string(arg0.(Atom))]
		if !ok {
			cl = cList{}
		}
		aArg.atomArg[string(arg0.(Atom))] = append(cl, g)
	case BoolType:
		aArg.boolArg = append(aArg.boolArg, g)
	case IntType:
		aArg.intArg = append(aArg.intArg, g)
	case FloatType:
		aArg.floatArg = append(aArg.floatArg, g)
	case StringType:
		aArg.strArg = append(aArg.strArg, g)
	case CompoundType:
		cl, ok := aArg.compArg[arg0.(Compound).Functor]
		if !ok {
			cl = cList{}
		}
		aArg.compArg[arg0.(Compound).Functor] = append(cl, g)
	case ListType:
		aArg.listArg = append(aArg.listArg, g)
	}
	aArg.varArg = append(aArg.varArg, g) // a veriable match to all types
}

var chrCounter *big.Int
var bigOne = big.NewInt(1)

func addConstraintToStore(g Compound) {
	// pTraceHeadln(3, 3, " a) Counter %v \n", chrCounter)
	g.Id = chrCounter
	chrCounter = new(big.Int).Add(chrCounter, bigOne)
	// pTraceHeadln(3, 3, " b) Counter++ %v , Id: %v \n", chrCounter, g.Id)
	if g.Prio == 0 {
		addGoal1(&g, CHRstore)
	} else {
		addGoal1(&g, BuiltInStore)
	}
}

func readProperConstraintsFromCHR_Store(t *Compound, env Bindings) cList {
	argAtt, ok := CHRstore[t.Functor]
	if ok {
		return readProperConstraintsFromStore(t, argAtt, env)
	}
	return cList{}
}

func readProperConstraintsFromBI_Store(t *Compound, env Bindings) cList {
	argAtt, ok := BuiltInStore[t.Functor]
	if ok {
		return readProperConstraintsFromStore(t, argAtt, env)
	}
	return cList{}
}

func readProperConstraintsFromStore(t *Compound, aAtt *argCHR, env Bindings) cList {
	args := t.Args
	l := len(args)
	if l == 0 {
		return aAtt.noArg
	}
	arg0 := args[0]
	argTyp := arg0.Type()
	for argTyp == VariableType {
		t2, ok := GetBinding(arg0.(Variable), env)
		if ok {
			arg0 = t2
			argTyp = arg0.Type()
		} else {
			break
		}
	}
	switch arg0.Type() {
	case AtomType:
		cl, ok := aAtt.atomArg[string(arg0.(Atom))]
		if ok {
			return cl
		}
	case BoolType:
		return aAtt.boolArg
	case IntType:
		return aAtt.intArg
	case FloatType:
		return aAtt.floatArg
	case StringType:
		return aAtt.strArg
	case CompoundType:
		cl, ok := aAtt.compArg[arg0.(Compound).Functor]
		if ok {
			return cl
		}
	case ListType:
		return aAtt.listArg
	case VariableType:
		return aAtt.varArg
	}
	return cList{}
}

type history [][]*big.Int

// var History []idSequence

var CurVarCounter *big.Int

type cList []*Compound

func (t cList) OccurVars() Vars {
	occur := Vars{}
	for _, t2 := range t {
		occur = append(occur, t2.OccurVars()...)
	}
	return occur
}

func (t cList) String() string {
	args := []string{}
	for _, arg := range t {
		args = append(args, arg.String())
	}
	return "[" + strings.Join(args, ", ") + "]"
}

func (t cList) Type() Type {
	return ListType
}

type chrRule struct {
	name     string
	id       int
	his      history
	delHead  cList // removed constraints
	keepHead cList // kept constraint
	guard    cList // built-in constraint
	body     List  // add CHR and built-in constraint
}

var CHRruleStore []*chrRule

var CHRtrace int

func pTraceHeadln(l, n int, s ...interface{}) {
	if CHRtrace >= l {
		for i := 0; i < n; i++ {
			fmt.Printf("      ")
		}
		fmt.Printf("*** ")
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
		fmt.Printf("\n")
	}
}

func pTraceHead(l, n int, s ...interface{}) {
	if CHRtrace >= l {
		for i := 0; i < n; i++ {
			fmt.Printf("      ")
		}
		fmt.Printf("*** ")
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
	}
}

func pTrace(l int, s ...interface{}) {
	if CHRtrace >= l {
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
	}
}

func pTraceln(l int, s ...interface{}) {
	if CHRtrace >= l {
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
		fmt.Printf("\n")
	}
}

func CHRsolver() {
	if CHRtrace != 0 {
		printCHRStore()
	}
	//	for ruleFound, i := true, 0; ruleFound && i < 1000; i++ {
	for ruleFound := true; ruleFound; {
		ruleFound = false
		for _, rule := range CHRruleStore {

			pTraceHeadln(2, 1, "trial rule ", rule.name, "(ID: ", rule.id, ")")

			if pRuleFired(rule) {
				pTraceHeadln(1, 1, "rule ", rule.name, " fired (id: ", rule.id, ")")
				ruleFound = true
				break
			}
			pTraceHeadln(2, 1, "rule ", rule.name, " NOT fired (id: ", rule.id, ")")
		}
		if ruleFound && CHRtrace != 0 {
			printCHRStore()
		}
	}
	if CHRtrace > 1 {
		printCHRStore()
	}
}

func pRuleFired(rule *chrRule) (ok bool) {
	headList := rule.delHead
	len_head := len(headList)
	if len_head != 0 {
		ok = unifyDelHead(rule, headList, 0, len_head, nil)
		return ok
	}

	headList = rule.keepHead
	len_head = len(headList)
	if len_head == 0 {
		return false
	}

	ok = unifyKeepHead(rule, []*big.Int{}, headList, 0, len_head, nil)
	return ok
}

//func attributedTerm(t Compound, env Bindings) cList {
//	return cList{}
//}

func unifyDelHead(r *chrRule, headList cList, it int, nt int, env Bindings) (ok bool) {
	var env2 Bindings
	var mark bool
	head := headList[it]
	chrList := readProperConstraintsFromCHR_Store(head, env)
	pTraceHeadln(3, 3, "     *** unify Del-Head %s with [", head)
	len_chr := len(chrList)
	if len_chr != 0 {
		// trace
		for _, c := range chrList {
			pTraceHeadln(3, 3, "%s, ", c)
		}
		pTraceHeadln(3, 3, "]\n")
		// trace
		for ok, ic := false, 0; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]

			env2, ok, mark = markCHRAndUnifyDelHead(r.id, head, chr, env) // mark chr and Unify, if fail unmark chr
			if ok {
				if it+1 < nt {
					ok = unifyDelHead(r, headList, it+1, nt, env2)
					if ok {
						// not unmarkDelCHR(chr), markt == deleted
						return ok
					}
				} else {
					// the last delHead-match was OK
					headList = r.keepHead
					nt = len(headList)
					if nt != 0 {
						ok = unifyKeepHead(r, nil, headList, 0, nt, env2)
						if ok {
							return ok
						}
					} else {
						// only delHead
						ok = checkGuards(r, env2)
						if ok {
							return ok
						}
					}
				} // if it+1 < nt
			}
			if mark {
				unmarkDelCHR(chr)
			}
			// mUnify was OK, but rule does not fire OR mUnify was not OK
			// env is the currend environment
			// try the next constrain for the constrain store
		}
		// no constrain from the constraint store match head
	}
	return false
}

func markCHRAndUnifyDelHead(id int, head, chr *Compound, env Bindings) (env2 Bindings, ok bool, m bool) {
	// mark and unmark chr
	if chr.IsActive {
		return env, false, false
	}
	// pTraceHeadln(3, 3, "     *** mark del %v, ID: %v\n", chr, chr.Id)
	chr.IsActive = true
	env2, ok = Unify(*head, *chr, env)
	pTraceHeadln(3, 3, "     *** Unify head %s with mark CHR %s (Id: %v) is %v (Binding: %v)\n", head, chr, chr.Id, ok, env2)
	return env2, ok, true
}

func unmarkDelCHR(chr *Compound) {
	chr.IsActive = false
	pTraceHeadln(3, 3, "     *** unmark del %v, ID: %v\n", chr, chr.Id)
	return
}

func markCHRAndUnifyKeepHead(id int, head, chr *Compound, env Bindings) (env2 Bindings, ok bool, m bool) {
	// mark and unmark chr

	if chr.IsActive {
		return env, false, false
	}
	// pTraceHeadln(3, 3, "     *** mark keep %v, ID: %v\n", chr, chr.Id)
	chr.IsActive = true
	env2, ok = Unify(*head, *chr, env)
	pTraceHeadln(3, 3, "     *** Unify head %s with mark CHR %s (Id: %v) is %v (Binding: %v)\n", head, chr, chr.Id, ok, env2)
	return env2, ok, true
}

func unmarkKeepCHR(chr *Compound) {
	chr.IsActive = false
	pTraceHeadln(3, 3, "     *** unmark keep %v, ID: %v\n", chr, chr.Id)
	return
}

func unifyKeepHead(r *chrRule, his []*big.Int, headList cList, it int, nt int, env Bindings) (ok bool) {
	var env2 Bindings
	var mark bool
	head := headList[it]
	chrList := readProperConstraintsFromCHR_Store(head, env)
	pTraceHeadln(3, 3, "     *** unify keep-Head %s with [", head)
	len_chr := len(chrList)
	if len_chr != 0 {
		// trace
		for _, c := range chrList {
			pTraceHeadln(3, 3, "%s, ", c)
		}
		pTraceHeadln(3, 3, "]\n")
		// trace

		for ok, ic := false, 0; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]

			env2, ok, mark = markCHRAndUnifyKeepHead(r.id, head, chr, env) // mark chr and Unify, if fail unmark chr
			if ok {
				if it+1 < nt {
					if his == nil {
						// rule with delHead
						ok = unifyKeepHead(r, nil, headList, it+1, nt, env2)
					} else {
						ok = unifyKeepHead(r, append(his, chr.Id), headList, it+1, nt, env2)
					}

					if ok {
						unmarkKeepCHR(chr)
						return ok
					}
				} else {
					// the last keepHead-match was OK
					// check history
					if his == nil {
						ok = checkGuards(r, env2)
						if ok {
							unmarkKeepCHR(chr)
							return ok
						}
					} else {
						// pTraceHeadln(3, 3, "     *** id von %s Args: %v ID: %v \n", chr.Functor, chr.Args, chr.Id)
						his2 := append(his, chr.Id)
						if !pCHRsInHistory(his2, r.his) {
							ok = checkGuards(r, env2)
							if ok {
								r.his = append(r.his, his2)
								unmarkKeepCHR(chr)
								return ok
							}
						} else {
							ok = false
						}
					}
				} // if it+1 < nt
			}
			// mUnify was OK, but rule does not fire OR mUnify was not OK
			// env is the currend environment
			// try the next constrain of the constrain store
			if mark {
				unmarkKeepCHR(chr)
			}
		}
		// no constrain from the constraint store match head
	}
	return false
}

func pCHRsInHistory(chrs []*big.Int, his history) (ok bool) {
	if his == nil || len(his) == 0 {
		return false
	}
	if chrs == nil || len(chrs) == 0 {
		return false
	}
	pTraceHeadln(3, 3, "     *** In History: chrs %v and his %vexist\n", chrs, his)
	lc := len(chrs)
	found := false
	// for i, h := range his {
	for _, h := range his {
		if len(h) != lc {
			// pTraceHeadln(3, 3, " In History: len of %d (len: %d) not == %d\n", i, len(h), lc)
			continue
		}
		// for j, c := range chrs {
		for _, c := range chrs {
			found = false
			// for k, h1 := range h {
			for _, h1 := range h {
				if h1 == nil {
					pTraceHeadln(3, 3, "    !!! In History h1 == nil \n")
				}
				if c == nil {
					pTraceHeadln(3, 3, "    !!! In History c == nil \n")
				}
				if h1 != nil && c != nil && h1.Cmp(c) == 0 {
					// pTraceHeadln(3, 3, " In History %v \n", h1)
					// pTraceHeadln(3, 3, " In History Nr: %d, idx %d == idx/chr %d \n", i, k, j)
					found = true
					break
				}
			}
			if !found {
				// pTraceHeadln(3, 3, " In History Nr: %d, Chr / idx: %d not found\n", i, j)
				break
			}
		}
		if found {
			// pTraceHeadln(3, 3, " History found \n")
			break
		}
	}
	return found
}

func checkGuards(r *chrRule, env Bindings) (ok bool) {
	for _, g := range r.guard {
		env2, ok := checkGuard(g, env)
		if !ok {
			return false
		}
		env = env2
	}
	if fireRule(r, env) {
		return true
	}
	// dt do setFail
	return true
}

func checkGuard(g *Compound, env Bindings) (env2 Bindings, ok bool) {
	pTraceHeadln(3, 3, "     *** check guard: %s, ", g)
	g1 := Substitute(*g, env).(Compound)
	pTraceHeadln(3, 3, "subst: %s, ", g1)
	if g.Functor == ":=" || g1.Functor == "is" || g1.Functor == "=" {
		if !pVar(g1.Args[0]) {
			return env, false
		}
		a := Eval(g1.Args[1])
		env2 = AddBinding(g1.Args[0].(Variable), a, env)
		return env2, true
	}

	t1 := Eval(g1)
	pTraceHeadln(3, 3, "eval: %s \n", t1)
	switch t1.Type() {
	case BoolType:
		if t1.(Bool) {
			return env, true
		}
		return env, false
	case CompoundType:
		t2 := t1.(Compound)
		biChrList := readProperConstraintsFromBI_Store(&t2, nil)
		len_chr := len(biChrList)
		if len_chr == 0 {
			return env, false
		}
		for _, chr := range biChrList {
			if Equal(t1, chr) {
				return env, true
			}
		}
		// to do for the operators(@): ==, !=, <, <=, >, >=, =<
		// symmetry: x @ y --> y @ x
		// transitivity: x @ y && y @ z --> x @ z
		//
		// case AtomType, IntType, FloatType, StringType:
		//	case ListType:
		//	case VariableType:
	}
	return env, false
}

func pVar(t Term) bool {
	if t.Type() == VariableType {
		return true
	}
	return false
}

func fireRule(rule *chrRule, env Bindings) bool {
	goals := Substitute(rule.body, env)
	goals = Eval(goals)
	pTraceHeadln(3, 3, "     *** Add Goals: %v \n", goals)
	if goals.Type() == ListType {
		for _, g := range goals.(List) {
			if g.Type() == CompoundType {
				addConstraintToStore(g.(Compound))
			} else {
				if g.Type() == BoolType && !g.(Bool) {
					return false
				}
			}
		}
	}
	return true
}

func printCHRStore() {
	first := true
	for _, aChr := range CHRstore {
		for _, con := range aChr.varArg {
			if !con.IsActive {
				if first {
					pTraceHead(1, 0, "CHR-Store: [", con.String())
					first = false
				} else {
					pTrace(1, ", ", con.String())
				}
			}
		}
	}
	if first {
		pTraceHeadln(1, 0, "CHR-Store: []")
	} else {
		pTraceln(1, "]")
	}

	first = true
	for _, aChr := range BuiltInStore {
		for _, con := range aChr.varArg {
			if !con.IsActive {
				if first {
					pTraceHead(1, 0, "Built-In Store: [", con.String())
					first = false
				} else {
					pTrace(1, ", ", con.String())
				}
			}
		}
		if first {
			pTraceHeadln(1, 0, "Built-In Store: []")
		} else {
			pTraceln(1, "]")
		}
	}
}
