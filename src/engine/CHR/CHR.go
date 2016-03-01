// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Constraint Handling Rules

package chr

import (
	. "github.com/carneades/carneades-4/src/engine/terms"
	// "fmt"
	"math/big"
	// "strconv"
	// "strings"
)

var QueryVars Vars

var QuerySore List

var CHRstore List

var BuiltInStore List

type history [][]*big.Int

// var History []idSequence

var CurVarCounter *big.Int

type chrRule struct {
	name     string
	id       int
	his      history
	delHead  List // removed constraints
	keepHead List // kept constraint
	guard    List // built-in constraint
	body     List // add CHR and built-in constraint
}

var CHRruleStore []*chrRule

func CHRsolver() {
	for ruleFound := true; ruleFound; {
		ruleFound = false
		for _, rule := range CHRruleStore {
			if ruleFired(rule) {
				ruleFound = true
				break
			}
		}
	}
}

func ruleFired(rule *chrRule) (ok bool) {
	headList := rule.delHead
	len_head := len(headList)
	if len_head != 0 {
		_, ok = unifyDelHead(rule, headList, 0, len_head, nil)
		return ok
	}

	headList = rule.keepHead
	len_head = len(headList)
	if len_head == 0 {
		return false
	}

	_, ok = unifyKeepHead(rule, rule.his, headList, 0, len_head, nil)
	return ok
}

func attributedTerm(t Term) []Term {
	return []Term{}
}

func unifyDelHead(r *chrRule, headList List, it int, nt int, env Bindings) (env2 Bindings, ok bool) {
	head := headList[it]
	chrList := attributedTerm(head)
	len_chr := len(chrList)
	if len_chr != 0 {
		for ok, ic := false, 0; !ok && ic < len_chr; ic++ {
			chr := chrList[ic]

			env2, ok = mUnify(head, chr, env) // mark chr and Unify, if fail unmark chr
			if ok {
				if it+1 < nt {
					env2, ok = unifyDelHead(r, headList, it+1, nt, env2)
					if ok {
						return env2, ok
					}
				} else {
					// the last delHead-match was OK
					headList = r.keepHead
					nt = len(headList)
					if nt != 0 {
						env2, ok = unifyKeepHead(r, r.his, headList, 0, nt, env2)
						if ok {
							return env2, ok
						}
					} else {
						// only delHead
						_, ok := checkGuards(r, env2)
						if ok {
							return env2, ok
						}

					}

				} // if it+1 < nt
			}
			// mUnify was OK, but rule does not fire OR mUnify was not OK
			// env is the currend environment
			// try the next constrain for the constrain store
		}
		// no constrain from the constraint store match head
	}
	return env, false
}

func mUnify(head, chr Term, env Bindings) (env2 Bindings, ok bool) {
	// mark and unmark chr
	return Unify(head, chr, env)
}

func unifyKeepHead(r *chrRule, his history, headList List, it int, nt int, env Bindings) (env2 Bindings, ok bool) {
	return nil, true
}

func checkGuards(rule *chrRule, env Bindings) (env2 Bindings, ok bool) {
	return env, true
}

func fireRule(rule *chrRule, env Bindings) (ok bool) {
	return true
}
