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
*/
/*
func tAtt(t *testing.T, store string, head string, result string) bool {

	term1, ok := ParseString(store)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan store in add/read store test \"%s\" failed, term1: %s", store, term1.String()))
		return false
	}
	term2, ok := ParseString(head)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan head in add/read store test \"%s\" failed, term2: %s", head, term2.String()))
		return false
	}
	term3, ok := ParseString(result)
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

func tAddStringChrRule(t *testing.T, name, keep, del, guard, body string) bool {

	keepList, ok := ParseCHRString(keep)
	if !ok || keepList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan KEEP-Head in rule %s failed(%v): %s\n", name, ok, keepList))
		return false
	}
	cKeepList, ok := toClist(keepList)
	if !ok {
		t.Errorf(fmt.Sprintf("Convert Keep-Head in rule %s failed: %s\n", name, keepList))
		return false
	}

	delList, ok := ParseCHRString(del)
	if !ok || delList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan DEl-Head in rule %s failed: %s\n", name, delList))
		return false
	}
	cDelList, ok := toClist(delList)
	if !ok {
		t.Errorf(fmt.Sprintf("Convert DEl-Head in rule %s failed: %s\n", name, delList))
		return false
	}

	guardList, ok := ParseBIString(guard)
	if !ok || guardList.Type() != ListType {
		t.Errorf(fmt.Sprintf("Scan GUARD in rule %s failed: %s (%v)\n", name, guardList, ok))
		return false
	}
	cGuardList, ok := toClist(guardList)
	if !ok {
		t.Errorf(fmt.Sprintf("Convert GUARD in rule %s failed: %s\n", name, guardList))
		return false
	}

	bodyList, ok := ParseRuleGoalString(body)
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

func tAddStringGoals(t *testing.T, goals string) bool {
	goalList, ok := ParseGoalString(goals)
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
	ok := tAddStringChrRule(t, "prime01", "prime(N)", "", "N>2", "prime(N-1)")

	if ok != true {
		t.Errorf("TestCHRRule01 failed, add Rule 01\n")
	}
	ok = tAddStringChrRule(t, "prime02", "prime(A)", "prime(B)", "B > A, B mod A == 0", "true")
	if ok != true {
		t.Errorf("TestCHRRule01 failed, add Rule 02\n")
	}
	ok = tAddStringGoals(t, "prime(100)")
	if ok != true {
		t.Errorf("TestCHRRule01 failed, add Goals\n")
	}
	CHRtrace = 0
	CHRsolver()
	CHRtrace = 1
	printCHRStore()
}

func checkResult(t *testing.T, chr, bi string) {
	if chr != chr2string() {
		t.Error(fmt.Sprintf(" exspected chr result: '%s' \n !=computed chr result: '%s'", chr, chr2string()))
	}
	if bi != bi2string() {
		t.Error(fmt.Sprintf(" exspected BI result: '%s' \n !=computed BI result: '%s'", bi, bi2string()))
	}
}

func TestCHRRule01a(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	prime01 @ prime(N) ==> N>2 | prime(N-1).
	prime02 @ prime(A) | prime(B) <=> B > A, B mod A == 0 | true.
	prime(20).`)
	if !ok {
		t.Error("TestCHRRule01a fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[prime(19), prime(17), prime(13), prime(11), prime(7), prime(5), prime(3), prime(2)]", "[]")
	CHRtrace = 1
	printCHRStore()

}

func TestCHRRule02(t *testing.T) {
	InitStore()
	ok := tAddStringChrRule(t, "gcd01", "", "gcd(0)", "", "true")

	if ok != true {
		t.Errorf("TestCHRRule02 failed, add Rule 01\n")
	}
	ok = tAddStringChrRule(t, "gcd02", "gcd(N)", "gcd(M)", "N <= M, L := M mod N", "gcd(L)")
	if ok != true {
		t.Errorf("TestCHRRule02 failed, add Rule 02\n")
	}
	ok = tAddStringGoals(t, "gcd(94017), gcd(1155),gcd(2035)")
	if ok != true {
		t.Errorf("TestCHRRule02 failed, add Goals\n")
	}
	CHRtrace = 0
	CHRsolver()
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule02a(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	gcd01@ gcd(0) <=> true .
	// logarithmic complexity
	gcd02@ gcd(N) \ gcd(M) <=> N <= M, L := M mod N | gcd(L).
	gcd(94017), gcd(1155),gcd(2035).`)
	if !ok {
		t.Error("TestCHRRule02a fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[gcd(11)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule02b(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	gcd01@ gcd(0) <=> true .
	// linear complexity
	gcd02@ gcd(N) \ gcd(M) <=> 0<N, N=<M | gcd(M-N).
	gcd(94017), gcd(1155),gcd(2035).`)
	if !ok {
		t.Error("TestCHRRule02b fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[gcd(11)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule02c(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	gcd01@ gcd(0) <=> true .
	// linear complexity
	gcd02@ gcd(N) \ gcd(M) <=> 0<N, N=<M, L is M - N | gcd(L).
	gcd(94017), gcd(1155),gcd(2035).`)
	if !ok {
		t.Error("TestCHRRule02c fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[gcd(11)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

/*
func TestCHRRule02d(t *testing.T) {
	CHRtrace = 3
	ok := ParseStringCHRRulesGoals(`
	gcd01@ gcd(0) <=> true .
	// linear complexity
	gcd02@ gcd(N) \ gcd(M) <=> 0<N, N=<M | L := M - N, gcd(L).
	gcd(94017), gcd(1155),gcd(2035).`)
	if !ok {
		t.Error("TestCHRRule02d fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[gcd(11)]", "[]")
	CHRtrace = 1
	printCHRStore()
}
*/

func TestCHRRule03(t *testing.T) {
	InitStore()
	ok := tAddStringChrRule(t, "gcd01", "", "gcd(0)", "", "true")

	if ok != true {
		t.Errorf("TestCHRRule03 failed, add Rule 01\n")
	}
	ok = tAddStringChrRule(t, "gcd02", "gcd(N)", "gcd(M)", "N <= M", "gcd(M mod N)")
	if ok != true {
		t.Errorf("TestCHRRule03 failed, add Rule 02\n")
	}
	ok = tAddStringGoals(t, "gcd(94017), gcd(1155),gcd(2035)")
	if ok != true {
		t.Errorf("TestCHRRule03 failed, add Goals\n")
	}
	CHRtrace = 0
	CHRsolver()
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule03a(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	gcd01@ gcd(0) <=> true .
	gcd02@ gcd(N) \ gcd(M) <=> N <= M | gcd(M mod N).
	gcd(94017), gcd(1155),gcd(2035).`)
	if !ok {
		t.Error("TestCHRRule03a fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[gcd(11)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule04(t *testing.T) {
	InitStore()
	ok := tAddStringChrRule(t, "fib01", "upto(A)", "", "", "fib(0,1), fib(1,1)")

	if ok != true {
		t.Errorf("TestCHRRule04 failed, add Rule 01\n")
	}
	ok = tAddStringChrRule(t, "fib02", "upto(Max), fib(N1,M1),fib(N2,M2)", "", "Max > N2, N2 == N1+1", "fib(N2+1,M1+M2)")
	if ok != true {
		t.Errorf("TestCHRRule04 failed, add Rule 02\n")
	}
	ok = tAddStringGoals(t, "upto(100)")
	if ok != true {
		t.Errorf("TestCHRRule04 failed, add Goals\n")
	}
	CHRtrace = 0
	CHRsolver()
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule04a(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	fib01@ upto(A) ==> fib(0,1), fib(1,1).
	fib02@ upto(Max), fib(N1,M1), fib(N2,M2) ==> Max > N2, N2 == N1+1 | fib(N2+1,M1+M2).
	upto(10).`)
	if !ok {
		t.Error("TestCHRRule04a fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[upto(10), fib(0,1), fib(1,1), fib(2,2), fib(3,3), fib(4,5), fib(5,8), fib(6,13), fib(7,21), fib(8,34), fib(9,55), fib(10,89)]", "[]")

	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule05(t *testing.T) {
	tstr := "TestCHRRule05"
	InitStore()
	ok := tAddStringChrRule(t, "leq reflexivity", "", "leq(X,X)", "", "true")

	if ok != true {
		t.Errorf(fmt.Sprintf("%s failed, add Rule 01\n", tstr))
	}

	ok = tAddStringChrRule(t, "leq antisymmetry", "", "leq(X,Y),leq(Y,X)", "", "X==Y")

	if ok != true {
		t.Errorf(fmt.Sprintf("%s failed, add Rule 02\n", tstr))
	}

	ok = tAddStringChrRule(t, "leq idempotence", "leq(X,Y)", "leq(X,Y)", "", "true")

	if ok != true {
		t.Errorf(fmt.Sprintf("%s failed, add Rule 03\n", tstr))
	}

	ok = tAddStringChrRule(t, "leq transitivity", "leq(X,Y),leq(Y,Z)", "", "", "leq(X,Z)")

	if ok != true {
		t.Errorf(fmt.Sprintf("%s failed, add Rule 04\n", tstr))
	}

	ok = tAddStringGoals(t, "leq(a,b), leq(b,c), leq(c,a)")
	if ok != true {
		t.Errorf(fmt.Sprintf("%s failed, add Goals\n", tstr))
	}
	CHRtrace = 0
	CHRsolver()
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule05a(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	leq_reflexivity  @ leq(X,X) <=> true.
	leq_antisymmetry @ leq(X,Y), leq(Y,X) <=> X==Y.
	leq_idempotence  @ leq(X,Y)\ leq(X,Y) <=> true.
	leq_transitivity @ leq(X,Y), leq(Y,Z) ==> leq(X,Z).
	leq(A,B), leq(B,C), leq(C,A).`)
	if !ok {
		t.Error("TestCHRRule05a fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[]", "[A==C, B==C]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule06(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	data1 @ data() ==> edge(berlin, 230, wolfsburg), edge(hannover, 89, wolfsburg), edge(hannover, 108, bielefeld), edge(bielefeld, 194, köln).
	data2 @ data() ==> edge(berlin,259, jena), edge(jena,55, erfurt), edge(erfurt,205,giessen), edge(giessen,158,köln), edge(köln, 85, aachen).
	source @ source(V) ==> dist(V, 0).
	del @ dist(V,D1) \ dist(V, D2) <=> D1 <= D2 | true.
	dist_plus1 @ dist(V,D1), edge(V, D2, V2) ==> dist(V2, D1+D2).
	dist_plus2 @ dist(V,D1), edge(V2, D2, V) ==> dist(V2, D1+D2).
	del_data @ edge(X,Y,Z) <=> true.
	data(), source(berlin).`)
	if !ok {
		t.Error("TestCHRRule06 fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[source(berlin), dist(berlin,0), dist(wolfsburg,230), dist(jena,259), dist(erfurt,314), dist(giessen,519), dist(hannover,319), dist(bielefeld,427), dist(köln,621), dist(aachen,706)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule07(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
	
	data1 @ data() ==> edge(berlin, 230, wolfsburg), edge(hannover, 89, wolfsburg), edge(hannover, 108, bielefeld), edge(bielefeld, 194, köln).
	data2 @ data() ==> edge(berlin,259, jena), edge(jena,55, erfurt), edge(erfurt,205,giessen), edge(giessen,158,köln), edge(köln, 85, aachen).
	source @ source(V) ==> dist(V, [V], 0).
	del @ dist(V, L, D1) \ dist(V, M, D2) <=> D1 <= D2 | true.
	dist_plus1 @ dist(V, L, D1), edge(V, D2, V2) ==> dist(V2,[V2|L], D1+D2).
	dist_plus2 @ dist(V, L, D1), edge(V2, D2, V) ==> dist(V2,[V2|L], D1+D2).
	del_data @ edge(X,Y,Z) <=> true.
	data(), source(berlin).`)
	if !ok {
		t.Error("TestCHRRule07 fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[source(berlin), dist(berlin,[berlin],0), dist(wolfsburg,[wolfsburg, berlin],230), dist(jena,[jena, berlin],259), dist(erfurt,[erfurt, jena, berlin],314), dist(giessen,[giessen, erfurt, jena, berlin],519), dist(hannover,[hannover, wolfsburg, berlin],319), dist(bielefeld,[bielefeld, hannover, wolfsburg, berlin],427), dist(köln,[köln, bielefeld, hannover, wolfsburg, berlin],621), dist(aachen,[aachen, köln, bielefeld, hannover, wolfsburg, berlin],706)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule08(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`

	data1 @ data() ==> edge(berlin, 230, wolfsburg), edge(hannover, 89, wolfsburg), edge(hannover, 108, bielefeld), edge(bielefeld, 194, köln).
	data2 @ data() ==> edge(berlin,259, jena), edge(jena,55, erfurt), edge(erfurt,205,giessen), edge(giessen,158,köln), edge(köln, 85, aachen).
	source @ source(V) ==> dist([V], 0).
	del @ dist([V|L], D1) \ dist([V|M], D2) <=> D1 <= D2 | true.
	dist_plus_a@ dist([V|L], D1), edge(V, D2, V2) ==> dist([V2, V|L], D1+D2).
	dist_plus_b@ dist([V|L], D1), edge(V2, D2, V) ==> dist([V2, V|L], D1+D2).
	del_data @ edge(X,Y,Z) <=> true.
	data(), source(berlin).`)
	if !ok {
		t.Error("TestCHRRule08 fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[source(berlin), dist([berlin],0), dist([wolfsburg, berlin],230), dist([jena, berlin],259), dist([erfurt, jena, berlin],314), dist([giessen, erfurt, jena, berlin],519), dist([hannover, wolfsburg, berlin],319), dist([bielefeld, hannover, wolfsburg, berlin],427), dist([köln, bielefeld, hannover, wolfsburg, berlin],621), dist([aachen, köln, bielefeld, hannover, wolfsburg, berlin],706)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

func TestCHRRule09(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`

	data1 @ data() ==> edge(berlin, 230, wolfsburg), edge(hannover, 89, wolfsburg), edge(hannover, 108, bielefeld), edge(bielefeld, 194, köln).
	data2 @ data() ==> edge(berlin,259, jena), edge(jena,55, erfurt), edge(erfurt,205,giessen), edge(giessen,158,köln), edge(köln, 85, aachen).
	source @ source(V) ==> dist([V], 0).
	del @ dist([V|L], D1) \ dist([V|M], D2) <=> D1 <= D2 | true.
	dist_plus_a@ dist([V|L], D1), edge(V, D2, V2) ==> dist([V2, V|L], D1+D2).
	dist_plus_b@ dist([V|L], D1), edge(V2, D2, V) ==> dist([V2, V|L], D1+D2).
	del_data @ edge(X,Y,Z) <=> true.
	data(), source(berlin).`)
	if !ok {
		t.Error("TestCHRRule09 fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[source(berlin), dist([berlin],0), dist([wolfsburg, berlin],230), dist([jena, berlin],259), dist([erfurt, jena, berlin],314), dist([giessen, erfurt, jena, berlin],519), dist([hannover, wolfsburg, berlin],319), dist([bielefeld, hannover, wolfsburg, berlin],427), dist([köln, bielefeld, hannover, wolfsburg, berlin],621), dist([aachen, köln, bielefeld, hannover, wolfsburg, berlin],706)]", "[]")
	CHRtrace = 1
	printCHRStore()
}

/*
func TestCHRRule10(t *testing.T) {
	CHRtrace = 0
	ok := ParseStringCHRRulesGoals(`
// make domains for variables
item(V,W) <=> in(item(V,W,count(0)),[0,1]).

// generate auxiliary constraints etc.
knapsack(WL) <=> retractall(memlistitems(X)), retractall(memtotalv(X)), asserta(memlistitems([])),
    asserta(memtotalv(0)), wlimit(WL), ptotalw(0), ptotalv(0), listitems([]), labelk(), labelmem().

// if an item can't be added to the set, constrain its domain to be [0]
wlimit(WL), ptotalw(TW) \ in(item(V,W, count(0)),[0,1]) <=> TW2 is TW+W, TW2 > WL | in(item(V,W,count(0)),[0]).

// if an item is labeled to be added to the set, update the partial total weight
wlimit(WL) \ in(item(V,W,count(0)),[1]), ptotalw(TW) <=> TW2 is TW+W, TW2 =< WL |
    ptotalw(TW2), in(item(V,W,count(1)),[1]).

// labeling
labelk() \ in(item(V,W,count(0)), [0,1]) <=> member(P,[0,1]), in(item(V,W,count(0)),[P]).

// calculating total value of the solution
in(item(V,W,count(1)),[1]), ptotalv(TV) <=> TV2 is TV+V, in(item(V,W,count(2)),[1]), ptotalv(TV2).

// listing items included in the solution
in(item(V,W,count(2)),D), listitems(L) <=> in(item(V,W,count(3)),D), listitems([item(V,W) | L]).

// if the current solution has a value greater than the previous ones, update the greatest value
// (and the corresponding set), print intermediate result, then backtrack
labelmem() \ listitems(L), ptotalv(TV) <=>
    memtotalv(TV1), TV > TV1, memlistitems(L1) |
        retract(memtotalv(TV1)), asserta(memtotalv(TV)), retract(memlistitems(L1)), asserta(memlistitems(L)),
//        print('best set '), print(L), nl,
//        print('with value '), print(TV), nl, nl,
        false.

// ...otherwise just backtrack
labelmem() \ ptotalv(TV) <=> memtotalv(TV1), TV =< TV1 | false.
item(1,1), item(1,2), item(2,4), item(3,4), item(4,3), knapsack(10).`)
	if !ok {
		t.Error("TestCHRRule09 fails, Error in parse string")
	}

	CHRsolver()

	checkResult(t, "[item(1, 1), item(1, 2), item(3, 4), item(4, 3)]", "[]")
	CHRtrace = 1
	printCHRStore()
}
*/
