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

type idSequence [][]*big.Int

var History []idSequence

var CurVarCounter *big.Int

type chrRule struct {
	name  string
	id    int
	delHead List // removed constraints
	keepHead List // kept constraint
	guard List // built-in constraint
	body  List // add CHR and built-in constraint
}

var CHRruleStore []*chrRule

func CHRsolver () {
	var ruleFound true
	var rule *chrRule
	var env 
	for ruleFound {
		ruleFound = false
		for _, rule = range CHRruleStore {
			env, ok := UnifyHeads(rule)
			if ok {break}
		}
		if ok {
			env, ok = 
		}
		
		
	}
}

ProceRule


