// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package chr

import (
	"fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	"testing"
)

func tAtt(t *testing.T, str1 string, result string) bool {

	term1, ok := ReadString(str1)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan str1 in test eval \"%s\" failed, term1: %s", str1, term1.String()))
		return false
	}
	term2, ok := ReadString(result)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan result in test eval \"%s\" failed, term2: %s", result, term2.String()))
		return false
	}
	InitStore()
	switch term1.Type() {
	case ListType:
		fmt.Printf(" store [")
		for _, g := range term1.(List) {
			if g.Type() == CompoundType {
				addGoal(g.(Compound))
				fmt.Printf("%s, ", g)
			} else {
				fmt.Printf(" no CHR predicate: %s \n", g)
			}
		}
		fmt.Printf("]\n")
	case CompoundType:
		addGoal(term1.(Compound))
		fmt.Printf("store [%s]\n", term1)
	default:
		fmt.Printf(" no CHR predicate or list: %s \n", term1)
	}

	switch term2.Type() {
	case ListType:
		for _, g := range term2.(List) {
			if g.Type() == CompoundType {
				fmt.Printf(" head: %s store:[", g)
				if g.(Compound).Prio == 0 {
					for _, s := range attributedTermCHR(g.(Compound), nil) {
						fmt.Printf("%s, ", s)
					}
				} else {
					for _, s := range attributedTermBI(g.(Compound), nil) {
						fmt.Printf("%s, ", s)
					}
				}
				fmt.Printf("]\n")
			} else {
				fmt.Printf(" no head predicate: %s \n", g)
			}
		}
	case CompoundType:
		if term2.(Compound).Prio == 0 {
			for _, s := range attributedTermCHR(term2.(Compound), nil) {
				fmt.Printf("%s, ", s)
			}
		} else {
			for _, s := range attributedTermBI(term2.(Compound), nil) {
				fmt.Printf("%s, ", s)
			}
		}
		fmt.Printf("]\n")
	default:
		fmt.Printf(" no head predicate or list: %s \n", term2)
	}

	return true

}

func TestCHR01(t *testing.T) {
	ok := tAtt(t, "[p(a, b), p(b, a), p(a,a), p(b, b)]", "[p(A,a), p(a,A), p(b,A)]")
	if ok != true {
		t.Errorf("TestStore01 failed\n")
	}
}

func TestCHR02(t *testing.T) {
	ok := tAtt(t, "[p(a, b), p(b, a), q(a,a), p(b, b)]", "[p(A,a), p(a,A), p(b,A), q(b,b), q(a,3)]")
	if ok != true {
		t.Errorf("TestStore02 failed\n")
	}
}
