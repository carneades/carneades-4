// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package chr

import (
	"fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	// "math/big"
	"testing"
)

/*
var b0 = big.NewInt(0)
var b1 = big.NewInt(1)
var b2 = big.NewInt(2)
var b3 = big.NewInt(3)
var b4 = big.NewInt(4)
var b5 = big.NewInt(5)
var b6 = big.NewInt(6)
var b7 = big.NewInt(7)
var b8 = big.NewInt(8)
var b9 = big.NewInt(9)
var b10 = big.NewInt(10)

func Test_History01(t *testing.T) {
	ruleHis := [][]*big.Int{{b0, b1}, {b2, b3, b4}, {b5, b6, b7}, {b8, b9, b10}}
	chrs := []*big.Int{b5, b6, b7}
	ok := pCHRsInHistory(chrs, ruleHis)
	if !ok {
		t.Error("(01) Chr not found in history\n")
		return
	}
	chrs = []*big.Int{b5, b6, b8}
	ok = pCHRsInHistory(chrs, ruleHis)
	if ok {
		t.Error("(02) Chr found in history\n")
		return
	}
	chrs = []*big.Int{b8, b9, b10}
	ok = pCHRsInHistory(chrs, ruleHis)
	if !ok {
		t.Error("(03) Chr not found in history\n")
		return
	}
	chrs = []*big.Int{b1, b0}
	ok = pCHRsInHistory(chrs, ruleHis)
	if !ok {
		t.Error("(04) Chr not found in history\n")
		return
	}
	chrs = []*big.Int{b0, b1, b2}
	ok = pCHRsInHistory(chrs, ruleHis)
	if ok {
		t.Error("(05) Chr found in history\n")
		return
	}

}

func tAtt(t *testing.T, store string, head string, result string) bool {

	term1, ok := ReadString(store)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan store in add/read store test \"%s\" failed, term1: %s", store, term1.String()))
		return false
	}
	term2, ok := ReadString(head)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan head in add/read store test \"%s\" failed, term2: %s", head, term2.String()))
		return false
	}
	term3, ok := ReadString(result)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan result add/read store test \"%s\" failed, term3: %s", result, term3.String()))
		return false
	}
	InitStore()
	switch term1.Type() {
	case ListType:
		fmt.Printf(" store [")
		for _, g := range term1.(List) {
			if g.Type() == CompoundType {
				addConstraintToStore(g.(Compound))
				fmt.Printf("%s, ", g)
			} else {
				fmt.Printf(" no CHR predicate: %s \n", g)
			}
		}
		fmt.Printf("]\n")
	case CompoundType:
		addConstraintToStore(term1.(Compound))
		fmt.Printf("store [%s]\n", term1)
	default:
		fmt.Printf(" no CHR predicate or list: %s \n", term1)
	}

	if term2.Type() != CompoundType {
		fmt.Printf(" head must be a predicate, not %s", term2)
	}
	att := cList{}
	fmt.Printf(" Head: %s ", term2)
	if term2.(Compound).Prio == 0 {
		att = readProperConstraintsFromCHR_Store(term2.(Compound), nil)
	} else {
		att = readProperConstraintsFromBI_Store(term2.(Compound), nil)
	}
	if term3.Type() != ListType {
		fmt.Printf(" result is not a list %s \n", term3)
	}
	cl := term3.(List)

	if len(cl) != len(att) {
		fmt.Printf("\n length of result not OK (exspected)%d != (computed)%d\n", len(cl), len(att))
		fmt.Printf(" Select: [")
		for _, a := range att {
			fmt.Printf("%s ,", a)
		}
		fmt.Printf("]\n")
		return false

	}
	fmt.Printf(" Select: [")
	for i, a := range att {
		if !Equal(a, cl[i]) {
			fmt.Printf(" term %d (%s) is not equal result %d (%s) \n", i, a, i, cl[i])
			return false
		}
		fmt.Printf("%s ,", a)
	}

	for i, c := range cl {
		if !Equal(c, att[i]) {
			fmt.Printf(" term %d (%s) is not equal result %d (%s) \n", i, att[i], i, c)
			return false
		}
	}

	fmt.Printf("]\n")
	return true

}

func TestCHR01(t *testing.T) {
	ok := tAtt(t, "[p(a, b), p(b, a), p(a,a), p(b, b)]", "p(A,a)", "[p(a, b), p(b, a), p(a,a), p(b, b)]")
	if ok != true {
		t.Errorf("TestStore01 failed\n")
	}
}

func TestCHR02(t *testing.T) {
	ok := tAtt(t, "[p(a, b), p(b, a), p(a,a), p(b, b)]", "p(a,a)", "[p(a, b), p(a,a)]")
	if ok != true {
		t.Errorf("TestStore02 failed\n")
	}
}

func TestCHR03(t *testing.T) {
	ok := tAtt(t, "[p(a, b), p(b, a), p(a,a), p(b, b)]", "p(b,a)", "[p(b, a), p(b, b)]")
	if ok != true {
		t.Errorf("TestStore03 failed\n")
	}
}

func TestCHR04(t *testing.T) {
	ok := tAtt(t, "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]",
		"p(true,a)", "[p(true, b),p(false, a)]")
	if ok != true {
		t.Errorf("TestStore04 failed\n")
	}
}

func TestCHR05(t *testing.T) {
	ok := tAtt(t, "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]",
		"p(162,a)", "[p(7,a),p(34, b)]")
	if ok != true {
		t.Errorf("TestStore05 failed\n")
	}
}

func TestCHR06(t *testing.T) {
	ok := tAtt(t, "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]",
		"p(736.6,a)", "[p(2.0,4.0),p(17.3,b)]")
	if ok != true {
		t.Errorf("TestStore06 failed\n")
	}
}

func TestCHR07(t *testing.T) {
	ok := tAtt(t, "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]",
		"p(\"OK\",a)", "[p(\"Hallo\", a),p(\"Welt\",b)]")
	if ok != true {
		t.Errorf("TestStoreß7 failed\n")
	}
}

func TestCHR08(t *testing.T) {
	ok := tAtt(t, "[p(q(2.0),4.0),p(q(\"Hallo\"), a),p(r(true), b),p(r(7),a),p(s(false), a),p(s(34), b),p(t(),b),p(t(),b)]",
		"p(r(77),a)", "[p(r(true), b),p(r(7),a)]")
	if ok != true {
		t.Errorf("TestStore08 failed\n")
	}
}

func TestCHR09(t *testing.T) {
	ok := tAtt(t, "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]",
		"p(B,a)", "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]")
	if ok != true {
		t.Errorf("TestStore09 failed\n")
	}
}

func TestCHR10(t *testing.T) {
	ok := tAtt(t, "[2.0+4.0, \"Hallo\"+a, true == b, 7 *a, false != a, 34>= b, 17.3 < b ,\"Welt\"< b]",
		"A+4", "[2.0+4.0, \"Hallo\"+a]")
	if ok != true {
		t.Errorf("TestStore10 failed\n")
	}
}

func TestCHR11(t *testing.T) {
	ok := tAtt(t, "[2.0+4.0, \"Hallo\"+a, true == b, 7 *a, false != a, 34>= b, 17.3 < b ,\"Welt\"< b]",
		"3.0+A", "[2.0+4.0]")
	if ok != true {
		t.Errorf("TestStore11 failed\n")
	}
}

func TestCHR12(t *testing.T) {
	ok := tAtt(t, "[2.0+4.0, \"Hallo\"+a, true == b, 7 *a, false != a, 34>= b, 17.3 < b ,\"Welt\"< b]",
		"\"Welt\"+x", "[\"Hallo\"+a]")
	if ok != true {
		t.Errorf("TestStore12 failed\n")
	}
}
*/
/*
func TestCHRnn(t *testing.T) {
	ok := tAtt(t, "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]",
	"p(true,a)", "[p(2.0,4.0),p(\"Hallo\", a),p(true, b),p(7,a),p(false, a),p(34, b),p(17.3,b),p(\"Welt\",b)]")
	if ok != true {
		t.Errorf("TestStorenn failed\n")
	}
}
*/
/*
func TestCHRxx(t *testing.T) {
	ok := tAtt(t, "[p(a, b), p(b, a), q(a,a), p(b, b)]", "p(c,A)", "[]")
	if ok != true {
		t.Errorf("TestStore02 failed\n")
	}
}
*/
//type chrRule struct {
//	name     string
//	id       int
//	delHead  cList // removed constraints
//	keepHead cList // kept constraint
//	guard    cList // built-in constraint
//	body     List  // add CHR and built-in constraint
//}

//var CHRruleStore []*chrRule

//func CHRsolver() {..}

//func addConstraintToStore(g Compound)

func toClist(l Term) (cList, bool) {
	cl := cList{}
	if l.Type() != ListType {
		return cl, false
	}
	for _, t1 := range l.(List) {
		if t1.Type() != CompoundType {
			return cl, false
		}
		t2 := t1.(Compound)
		cl = append(cl, &t2)
	}
	return cl, true
}

func addStringChrRule(t *testing.T, name, del, keep, guard, body string) bool {

	delList, ok := ReadString(del)
	if !ok || delList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan DEl-Head in rule %s failed: %s\n", name, delList))
		return false
	}
	cDelList, ok := toClist(delList)
	if !ok {
		t.Errorf(fmt.Sprintf("Convert DEl-Head in rule %s failed: %s\n", name, delList))
		return false
	}

	keepList, ok := ReadString(keep)
	if !ok || keepList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan KEEP-Head in rule %s failed: %s\n", name, delList))
		return false
	}
	cKeepList, ok := toClist(keepList)
	if !ok {
		t.Errorf(fmt.Sprintf("Convert Keep-Head in rule %s failed: %s\n", name, keepList))
		return false
	}

	guardList, ok := ReadString(guard)
	if !ok || guardList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan GUARD in rule %s failed: %s\n", name, delList))
		return false
	}
	cGuardList, ok := toClist(guardList)
	if !ok {
		t.Errorf(fmt.Sprintf("Convert GUARD in rule %s failed: %s\n", name, guardList))
		return false
	}

	bodyList, ok := ReadString(body)
	if !ok || bodyList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan BODY in rule %s failed: %s\n", name, bodyList))
		return false
	}

	CHRruleStore = append(CHRruleStore, &chrRule{name: name, id: nextRuleId,
		delHead:  cDelList,
		keepHead: cKeepList,
		guard:    cGuardList,
		body:     bodyList.(List)})
	nextRuleId++
	return true

}

func addGoals(t *testing.T, goals string) bool {
	goalList, ok := ReadString(goals)
	if !ok || goalList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan GOAL-List failed: %s\n", goalList))
		return false
	}
	for _, g := range goalList.(List) {
		if g.Type() == CompoundType {
			addConstraintToStore(g.(Compound))
		} else {
			t.Errorf(fmt.Sprintf(" GOAL is not a predicate: %s\n", g))
			return false
		}

	}
	return true
}

func TestCHRRule01(t *testing.T) {
	InitStore()
	ok := addStringChrRule(t, "prime01", "[]", "[prime(N)]", "[N>2]", "[prime(N-1)]")

	if ok != true {
		t.Errorf("TestCHRRule01 failed, add Rule 01\n")
	}
	ok = addStringChrRule(t, "prime02", "[prime(B)]", "[prime(A)]", "[B > A, B mod A == 0]", "[true]")
	if ok != true {
		t.Errorf("TestCHRRule01 failed, add Rule 02\n")
	}
	ok = addGoals(t, "[prime(100)]")
	if ok != true {
		t.Errorf("TestCHRRule01 failed, add Goals\n")
	}

	CHRsolver()
	printCHRStore()
}

//var CHRstore store

//var BuiltInStore store

//type argCHR struct {
//	atomArg  map[string]cList
//	boolArg  cList
//	intArg   cList
//	floatArg cList
//	strArg   cList
//	compArg  map[string]cList
//	listArg  cList
//	varArg   cList
//	noArg    cList
//}

//type store map[string]*argCHR
