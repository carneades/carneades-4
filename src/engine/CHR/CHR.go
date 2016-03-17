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
var emptyBinding Bindings

func InitStore() {
	v := NewVariable("")
	emptyBinding = &BindEle{Var: v, T: nil, Next: nil}
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

func pTraceEnv(l int, e Bindings) {
	if e == nil {
		pTrace(l, "nil")
	} else {
		if e.Var.Name == "" {
			pTrace(l, "[\"\"=nil]")
		} else {
			if e.Next == nil || e.Next.Var.Name == "" {
				pTrace(l, "[", e.Var.Name, "=", e.T.String(), ", nil]")
			} else {
				pTrace(l, "[", e.Var.Name, "=", e.T.String(), ",...]")
			}
		}

	}
}

func pTraceEMap(l int, n int, h *Compound) {
	if CHRtrace >= l {
		for i := 0; i < n; i++ {
			fmt.Printf("      ")
		}
		fmt.Printf("*** head: %s [ ", h.String())
		env := h.EMap
		for i, e := range *env {
			fmt.Printf("[ %d ] =", i)
			for _, e1 := range e {
				pTraceEnv(l, e1)
			}
			fmt.Printf(" || ")
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

			if CHRtrace != 0 {
				pTraceHeadln(2, 1, "trial rule ", rule.name, "(ID: ", rule.id, ")")

				if pTraceRuleFired(rule) {
					pTraceHeadln(1, 1, "rule ", rule.name, " fired (id: ", rule.id, ")")
					ruleFound = true
					break
				}
				pTraceHeadln(2, 1, "rule ", rule.name, " NOT fired (id: ", rule.id, ")")
			} else {
				if pRuleFired(rule) {
					ruleFound = true
					break
				}
			}

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
		ok = unifyDelHead(rule, headList, 0, len_head, 0, nil)
		return ok
	}

	headList = rule.keepHead
	len_head = len(headList)
	if len_head == 0 {
		return false
	}

	ok = unifyKeepHead(rule, []*big.Int{}, headList, 0, len_head, 0, emptyBinding)
	return ok
}

func pTraceRuleFired(rule *chrRule) (ok bool) {
	headList := rule.delHead
	len_head := len(headList)
	if len_head != 0 {
		ok = traceUnifyDelHead(rule, headList, 0, len_head, 0, nil)
		return ok
	}

	headList = rule.keepHead
	len_head = len(headList)
	if len_head == 0 {
		return false
	}

	ok = traceUnifyKeepHead(rule, []*big.Int{}, headList, 0, len_head, 0, emptyBinding)
	return ok
}

//func attributedTerm(t Compound, env Bindings) cList {
//	return cList{}
//}

func unifyDelHead(r *chrRule, headList cList, it int, nt int, ienv int, env Bindings) (ok bool) {
	var env2 Bindings
	var mark bool
	head := headList[it]
	chrList := readProperConstraintsFromCHR_Store(head, env)
	len_chr := len(chrList)
	if len_chr == 0 {
		return false
	}
	// begin check the next head
	lastDelHead := it+1 == nt
	lastHead := false
	if lastDelHead {
		// last del head
		headList = r.keepHead
		nt = len(headList)
		if nt == 0 {
			lastHead = true
		}
	}
	// End next check next head, if lastDelHead the headList == r.keephead
	// check in head stored environment map
	ie := 0
	len_ie := 0
	senv, ok := (*head.EMap)[ienv]
	if ok {
		len_ie = len(senv)
		if lastHead {
			ie = len_ie
		} else {
			if lastDelHead {
				for ; ie < len_ie; ie++ {
					env2 = senv[ie]
					if env2 != nil {
						chr := chrList[ie]
						mark = markCHR(chr)
						if mark {
							ok = unifyKeepHead(r, nil, headList, 0, nt, ie, env2)
							if ok {
								return ok
							}
							unmarkDelCHR(chr)
						}
					}
				}
			} else { // not a last Del-Head
				for ; ie < len_ie; ie++ {
					env2 = senv[ie]
					if env2 != nil {
						chr := chrList[ie]
						mark = markCHR(chr)
						if mark {
							ok = unifyDelHead(r, headList, it+1, nt, ie, env2)
							if ok {
								// not unmarkDelCHR(chr), markt == deleted
								return ok
							}
							unmarkDelCHR(chr)
						}

					}
				} // for ; ie < len_ie; ie++
			} // ! lastDelHead
		} // ! lastHead
	} else {
		senv = []Bindings{}
		(*head.EMap)[ienv] = senv
	}
	// End check in head stored environment map
	// normal head-check, start at ie (not at 0 !!)
	if lastHead {
		for ok, ic := false, ie; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]
			env2, ok, mark = markCHRAndUnifyDelHead(r.id, head, chr, env)
			if ok {
				senv = append(senv, env2)
				ok = checkGuards(r, env2)
				if ok {
					(*head.EMap)[ienv] = senv
					return ok
				}
			} else {
				senv = append(senv, nil)
			}
			if mark {
				unmarkDelCHR(chr)
			}
		}
		(*head.EMap)[ienv] = senv
		return false
	}
	if lastDelHead {
		for ok, ic := false, ie; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]
			env2, ok, mark = markCHRAndUnifyDelHead(r.id, head, chr, env)
			if ok {
				senv = append(senv, env2)

				ok = unifyKeepHead(r, nil, headList, 0, nt, ic, env2)
				if ok {
					(*head.EMap)[ienv] = senv
					return ok
				}
			} else {
				senv = append(senv, nil)
			}
			if mark {
				unmarkDelCHR(chr)
			}
		}
		(*head.EMap)[ienv] = senv
		return false
	}

	for ok, ic := false, ie; !ok && ic < len_chr; ic++ {

		chr := chrList[ic]

		env2, ok, mark = markCHRAndUnifyDelHead(r.id, head, chr, env) // mark chr and Unify, if fail unmark chr
		if ok {
			senv = append(senv, env2)
			ok = unifyDelHead(r, headList, it+1, nt, ic, env2)
			if ok {
				// not unmarkDelCHR(chr), markt == deleted
				(*head.EMap)[ienv] = senv
				return ok
			}
		} else {
			senv = append(senv, nil)
		}
		if mark {
			unmarkDelCHR(chr)
		}
	}
	(*head.EMap)[ienv] = senv
	return false
}

func traceUnifyDelHead(r *chrRule, headList cList, it int, nt int, ienv int, env Bindings) (ok bool) {
	var env2 Bindings
	var mark bool
	head := headList[it]
	chrList := readProperConstraintsFromCHR_Store(head, env)
	pTraceHead(3, 3, "unify Del-Head (", ienv, ") ", head, " with [")
	len_chr := len(chrList)
	if len_chr == 0 {
		return false
	}
	// begin trace
	first := true
	for _, c := range chrList {
		if first {
			pTrace(3, c)
			first = false
		} else {
			pTrace(3, ", ", c)
		}
	}
	pTraceln(3, "]")
	// end trace
	// begin check the next head
	lastDelHead := it+1 == nt
	lastHead := false
	if lastDelHead {
		// last del head
		headList = r.keepHead
		nt = len(headList)
		if nt == 0 {
			lastHead = true
		}
	}
	// End next check next head, if lastDelHead the headList == r.keephead
	// check in head stored environment map
	ie := 0
	len_ie := 0
	senv, ok := (*head.EMap)[ienv]
	if ok {
		pTraceEMap(3, 4, head)
		len_ie = len(senv)
		// trace
		pTraceHead(3, 3, "unify Del-Head (", ienv, ") ", head, " Env: [")
		first = true
		for _, e := range senv {
			if first {
				first = false
			} else {
				pTrace(3, ", ")
			}
			pTraceEnv(3, e)
		}
		pTraceln(3, "]")

		// End trace

		if lastHead {
			ie = len_ie
		} else {
			if lastDelHead {
				for ; ie < len_ie; ie++ {
					env2 = senv[ie]
					if env2 != nil {
						chr := chrList[ie]
						mark = markCHR(chr)
						if mark {
							ok = unifyKeepHead(r, nil, headList, 0, nt, ie, env2)
							if ok {
								return ok
							}
							traceUnmarkDelCHR(chr)
						}
					}
				}
			} else { // not a last Del-Head
				for ; ie < len_ie; ie++ {
					env2 = senv[ie]
					if env2 != nil {
						chr := chrList[ie]
						mark = markCHR(chr)
						if mark {
							ok = unifyDelHead(r, headList, it+1, nt, ie, env2)
							if ok {
								// not unmarkDelCHR(chr), markt == deleted
								return ok
							}
							traceUnmarkDelCHR(chr)
						}

					}
				} // for ; ie < len_ie; ie++
			} // ! lastDelHead
		} // ! lastHead
	} else {
		// head.EMap = &EnvMap{}
		pTraceEMap(3, 4, head)
		senv = []Bindings{}
		(*head.EMap)[ienv] = senv
	}
	// End check in head stored environment map
	// normal head-check, start at ie (not at 0 !!)
	pTraceHeadln(3, 3, "unify del-Head (", ienv, ") ", head, " from: ", ie, " < ", len_chr)
	if lastHead {
		for ok, ic := false, ie; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]
			env2, ok, mark = traceMarkCHRAndUnifyDelHead(r.id, head, chr, env)
			if ok {
				senv = append(senv, env2)
				// trace senv changes

				pTraceHead(3, 3, "New environment ", "Head: ", head.String(), ", Env: (", ienv, ") [", ic, "], =")
				pTraceEnv(3, env2)
				pTraceln(3, "")

				ok = traceCheckGuards(r, env2)
				if ok {
					(*head.EMap)[ienv] = senv
					pTraceEMap(3, 4, head)
					return ok
				}
			} else {
				senv = append(senv, nil)
			}
			if mark {
				traceUnmarkDelCHR(chr)
			}
		}
		(*head.EMap)[ienv] = senv
		pTraceEMap(3, 4, head)
		return false
	}
	if lastDelHead {
		for ok, ic := false, ie; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]
			env2, ok, mark = traceMarkCHRAndUnifyDelHead(r.id, head, chr, env)
			if ok {
				senv = append(senv, env2)
				// trace senv changes

				pTraceHead(3, 3, "New environment ", "Head: ", head.String(), ", Env: (", ienv, ") [", ic, "], =")
				pTraceEnv(3, env2)
				pTraceln(3, "")

				ok = unifyKeepHead(r, nil, headList, 0, nt, ic, env2)
				if ok {
					(*head.EMap)[ienv] = senv
					pTraceEMap(3, 4, head)
					return ok
				}
			} else {
				senv = append(senv, nil)
			}
			if mark {
				traceUnmarkDelCHR(chr)
			}
		}
		(*head.EMap)[ienv] = senv
		pTraceEMap(3, 4, head)
		return false
	}

	for ok, ic := false, ie; !ok && ic < len_chr; ic++ {

		chr := chrList[ic]

		env2, ok, mark = traceMarkCHRAndUnifyDelHead(r.id, head, chr, env) // mark chr and Unify, if fail unmark chr
		if ok {
			senv = append(senv, env2)
			// trace senv changes

			pTraceHead(3, 3, "New environment ", "Head: ", head.String(), ", Env: (", ienv, ") [", ic, "], =")
			pTraceEnv(3, env2)
			pTraceln(3, "")

			ok = unifyDelHead(r, headList, it+1, nt, ic, env2)
			if ok {
				// not unmarkDelCHR(chr), markt == deleted
				(*head.EMap)[ienv] = senv
				pTraceEMap(3, 4, head)
				return ok
			}
		} else {
			senv = append(senv, nil)
		}
		if mark {
			traceUnmarkDelCHR(chr)
		}
	}
	(*head.EMap)[ienv] = senv
	pTraceEMap(3, 4, head)
	return false
}

func markCHR(chr *Compound) bool {
	if chr.IsActive {
		return false
	}
	chr.IsActive = true
	return true
}

func traceMarkCHRAndUnifyDelHead(id int, head, chr *Compound, env Bindings) (env2 Bindings, ok bool, m bool) {
	// mark and unmark chr
	if chr.IsActive {
		return env, false, false
	}
	// pTraceHeadln(3, 3, "     *** mark del %v, ID: %v\n", chr, chr.Id)
	chr.IsActive = true
	env2, ok = Unify(*head, *chr, env)
	pTraceHead(3, 3, "Unify head ", head, " with mark CHR ", chr, " (Id: ", chr.Id, ") is ", ok, " (Binding: ")
	pTraceEnv(3, env2)
	pTraceln(3, ")")
	return env2, ok, true
}

func markCHRAndUnifyDelHead(id int, head, chr *Compound, env Bindings) (env2 Bindings, ok bool, m bool) {
	// mark and unmark chr
	if chr.IsActive {
		return env, false, false
	}
	chr.IsActive = true
	env2, ok = Unify(*head, *chr, env)
	return env2, ok, true
}

func unmarkDelCHR(chr *Compound) {
	chr.IsActive = false
	return
}

func traceUnmarkDelCHR(chr *Compound) {
	chr.IsActive = false
	pTraceHeadln(3, 3, "unmark del ", chr, ", ID: ", chr.Id)
	return
}

func traceMarkCHRAndUnifyKeepHead(id int, head, chr *Compound, env Bindings) (env2 Bindings, ok bool, m bool) {
	// mark and unmark chr

	if chr.IsActive {
		return env, false, false
	}
	// pTraceHeadln(3, 3, "mark keep ",chr,", ID: ",chr.Id )
	chr.IsActive = true
	env2, ok = Unify(*head, *chr, env)
	pTraceHead(3, 3, "Unify head ", head, " with mark CHR ", chr, " (Id: ", chr.Id, ") is ", ok, " (Binding: ")
	pTraceEnv(3, env2)
	pTraceln(3, ")")
	return env2, ok, true
}

func markCHRAndUnifyKeepHead(id int, head, chr *Compound, env Bindings) (env2 Bindings, ok bool, m bool) {
	// mark and unmark chr

	if chr.IsActive {
		return env, false, false
	}
	chr.IsActive = true
	env2, ok = Unify(*head, *chr, env)
	return env2, ok, true
}

func traceUnmarkKeepCHR(chr *Compound) {
	chr.IsActive = false
	pTraceHeadln(3, 3, "unmark keep ", chr, ", ID: ", chr.Id)
	return
}

func unmarkKeepCHR(chr *Compound) {
	chr.IsActive = false
	return
}

func unifyKeepHead(r *chrRule, his []*big.Int, headList cList, it int, nt int, ienv int, env Bindings) (ok bool) {
	var env2 Bindings
	var mark bool
	head := headList[it]
	chrList := readProperConstraintsFromCHR_Store(head, env)
	len_chr := len(chrList)
	if len_chr == 0 {
		return false
	}
	// begin check the next head
	lastKeepHead := it+1 == nt
	// End next check next head
	// check in head stored environment map
	ie := 0
	len_ie := 0
	senv, ok := (*head.EMap)[ienv]
	if ok {
		len_ie = len(senv)
		if lastKeepHead {
			ie = len_ie
		} else {
			for ; ie < len_ie; ie++ {
				env2 = senv[ie]
				if env2 != nil {
					chr := chrList[ie]
					mark = markCHR(chr)
					if mark {
						ok = unifyKeepHead(r, nil, headList, it+1, nt, ie, env2)
						if ok {
							unmarkKeepCHR(chr)
							return ok
						}
						unmarkKeepCHR(chr)
					}
				}
			}

		} // ! lastHead
	} else { // if !ok
		// head.EMap = &EnvMap{}
		senv = []Bindings{}
		(*head.EMap)[ienv] = senv
	}
	// End check in head stored environment map
	// normal head-check, start at ie (not at 0 !!)
	if lastKeepHead {
		for ok, ic := false, ie; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]
			env2, ok, mark = markCHRAndUnifyKeepHead(r.id, head, chr, env)
			if ok {
				senv = append(senv, env2)
				ok = checkGuards(r, env2)
				if ok {
					unmarkKeepCHR(chr)
					(*head.EMap)[ienv] = senv
					return ok
				}
			} else {
				senv = append(senv, nil)
			}
			if mark {
				unmarkKeepCHR(chr)
			}
		}
		(*head.EMap)[ienv] = senv
		return false
	}

	for ok, ic := false, ie; !ok && ic < len_chr; ic++ {

		chr := chrList[ic]

		env2, ok, mark = markCHRAndUnifyKeepHead(r.id, head, chr, env) // mark chr and Unify, if fail unmark chr
		if ok {
			senv = append(senv, env2)

			ok = unifyKeepHead(r, nil, headList, it+1, nt, ic, env2)
			if ok {
				unmarkKeepCHR(chr)
				(*head.EMap)[ienv] = senv
				return ok
			}
		} else {
			senv = append(senv, nil)
		}
		if mark {
			unmarkDelCHR(chr)
		}
	}
	(*head.EMap)[ienv] = senv
	return false
}

func traceUnifyKeepHead(r *chrRule, his []*big.Int, headList cList, it int, nt int, ienv int, env Bindings) (ok bool) {
	var env2 Bindings
	var mark bool
	head := headList[it]
	chrList := readProperConstraintsFromCHR_Store(head, env)
	pTraceHead(3, 3, "unify keep-Head (", ienv, ") ", head, " with [")
	len_chr := len(chrList)
	if len_chr == 0 {
		pTraceln(3, "] - empty chr")
		return false
	}
	// begin trace
	first := true
	for _, c := range chrList {
		if first {
			pTrace(3, c)
			first = false
		} else {
			pTrace(3, ", ", c)
		}
	}
	pTraceln(3, "]")
	// end trace
	// begin check the next head

	lastKeepHead := it+1 == nt
	pTraceHeadln(3, 4, " last keep head = ", lastKeepHead)

	// End next check next head
	// check in head stored environment map
	ie := 0
	len_ie := 0
	senv, ok := (*head.EMap)[ienv]
	if !ok {
		pTraceHeadln(3, 4, " !!! head: ", head, " with no Emap[ ", ienv, " ]")
	}
	if ok {
		pTraceEMap(3, 4, head)
		len_ie = len(senv)
		pTraceHeadln(3, 4, " len env (", ienv, ") = ", len_ie)
		if lastKeepHead {
			ie = len_ie
			pTraceHeadln(3, 4, " ie == len_ie == ", ie, " = ", len_ie)
		} else {
			// trace
			pTraceHead(3, 3, "unify Keep-Head (", ienv, ") ", head, " Env: [")
			first = true
			for _, e := range senv {
				if first {
					first = false
				} else {
					pTrace(3, ", ")
				}
				pTraceEnv(3, e)
			}
			pTraceln(3, "]")

			// End trace
			for ; ie < len_ie; ie++ {
				env2 = senv[ie]
				if env2 != nil {
					chr := chrList[ie]
					mark = markCHR(chr)
					pTraceHeadln(3, 4, " mark keep chr:", chr.String(), " = ", mark)
					if mark {
						ok = unifyKeepHead(r, nil, headList, it+1, nt, ie, env2)
						if ok {
							traceUnmarkKeepCHR(chr)
							return ok
						}
						traceUnmarkKeepCHR(chr)
					}
				}
			}

		} // ! lastHead
	} else { // if !ok
		// head.EMap = &EnvMap{}
		senv = []Bindings{}
		(*head.EMap)[ienv] = senv
	}
	// End check in head stored environment map
	// normal head-check, start at ie (not at 0 !!)
	pTraceHeadln(3, 3, "unify keep-Head ", head, " from: ", ie, " < ", len_chr)
	if lastKeepHead {
		for ok, ic := false, ie; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]
			env2, ok, mark = traceMarkCHRAndUnifyKeepHead(r.id, head, chr, env)
			if ok {
				senv = append(senv, env2)
				// trace senv changes

				pTraceHead(3, 3, "New environment ", "Head: ", head.String(), ", Env (", ienv, ") [", ic, "], =")
				pTraceEnv(3, env2)
				pTraceln(3, "")

				ok = traceCheckGuards(r, env2)
				if ok {
					traceUnmarkKeepCHR(chr)
					(*head.EMap)[ienv] = senv
					pTraceEMap(3, 4, head)
					return ok
				}
			} else {
				senv = append(senv, nil)
			}
			if mark {
				traceUnmarkKeepCHR(chr)
			}
		}
		(*head.EMap)[ienv] = senv
		pTraceEMap(3, 4, head)
		return false
	}

	for ok, ic := false, ie; !ok && ic < len_chr; ic++ {

		chr := chrList[ic]

		env2, ok, mark = traceMarkCHRAndUnifyKeepHead(r.id, head, chr, env) // mark chr and Unify, if fail unmark chr
		if ok {
			senv = append(senv, env2)
			// trace senv changes

			pTraceHead(3, 3, "New environment ", "Head: ", head.String(), ", Env: (", ienv, ") [", ic, "], =")
			pTraceEnv(3, env2)
			pTraceln(3, "")

			ok = unifyKeepHead(r, nil, headList, it+1, nt, ic, env2)
			if ok {
				traceUnmarkKeepCHR(chr)
				(*head.EMap)[ienv] = senv
				pTraceEMap(3, 4, head)
				return ok
			}
		} else {
			senv = append(senv, nil)
		}
		if mark {
			unmarkDelCHR(chr)
		}
	}
	(*head.EMap)[ienv] = senv
	pTraceEMap(3, 4, head)
	return false
}

func pCHRsInHistory(chrs []*big.Int, his history) (ok bool) {
	if his == nil || len(his) == 0 {
		return false
	}
	if chrs == nil || len(chrs) == 0 {
		return false
	}
	// pTraceHeadln(3, 3, "     *** In History: chrs %v and his %vexist\n", chrs, his)
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
					pTraceHeadln(3, 3, "!!! In History h1 == nil")
				}
				if c == nil {
					pTraceHeadln(3, 3, "!!! In History c == nil")
				}
				if h1 != nil && c != nil && h1.Cmp(c) == 0 {
					// pTraceHeadln(3, 3, " In History %v \n", h1)
					// pTraceHeadln(3, 3, " In History Nr: %d, idx %d == idx/chr %d \n", i, k, j)
					found = true
					break
				}
			}
			if !found {
				// pTraceHeadln(3, 3, "In History Nr: %d, Chr / idx: %d not found\n", i, j)
				break
			}
		}
		if found {
			break
		}
	}
	pTraceHeadln(3, 3, "CHR in history: ", found)
	return found
}

func traceCheckGuards(r *chrRule, env Bindings) (ok bool) {
	for _, g := range r.guard {
		env2, ok := traceCheckGuard(g, env)
		if !ok {
			return false
		}
		env = env2
	}
	if traceFireRule(r, env) {
		return true
	}
	// dt do setFail
	return true
}

func traceCheckGuard(g *Compound, env Bindings) (env2 Bindings, ok bool) {
	pTraceHead(3, 3, "check guard: ", g.String())
	g1 := Substitute(*g, env).(Compound)
	pTrace(3, ", subst: ", g1)
	if g.Functor == ":=" || g1.Functor == "is" || g1.Functor == "=" {
		if !pVar(g1.Args[0]) {
			return env, false
		}
		a := Eval(g1.Args[1])
		env2 = AddBinding(g1.Args[0].(Variable), a, env)
		return env2, true
	}

	t1 := Eval(g1)
	pTraceln(3, ", eval: ", t1)
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

	g1 := Substitute(*g, env).(Compound)

	if g.Functor == ":=" || g1.Functor == "is" || g1.Functor == "=" {
		if !pVar(g1.Args[0]) {
			return env, false
		}
		a := Eval(g1.Args[1])
		env2 = AddBinding(g1.Args[0].(Variable), a, env)
		return env2, true
	}
	t1 := Eval(g1)
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

func traceFireRule(rule *chrRule, env Bindings) bool {
	goals := Substitute(rule.body, env)
	goals = Eval(goals)
	pTraceHeadln(3, 3, "Add Goals: ", goals)
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

func fireRule(rule *chrRule, env Bindings) bool {
	goals := Substitute(rule.body, env)
	goals = Eval(goals)
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
