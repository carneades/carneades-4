// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package test

import (
	"github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	// "log"
	"fmt"
	"os"
	"testing"
)

// The Tandem example
// Source: Baroni, P., Caminada, M., and Giacomin, M. An introduction to
// argumentation semantics. The Knowledge Engineering Review 00, 0 (2004), 1-24.

func TestTandem(t *testing.T) {
	var jw = caes.Statement{
		Text:    "John wants to ride on the tandem.",
		Assumed: true}
	var mw = caes.Statement{
		Text:    "Mary wants to ride on the tandem.",
		Assumed: true}
	var sw = caes.Statement{
		Text:    "Suzy wants to ride on the tandem.",
		Assumed: true}
	var a1, a2, a3, a4, a5, a6 caes.Argument
	var jt = caes.Statement{
		Text: "John is riding on the tandem.",
		Args: []*caes.Argument{&a1}}
	var mt = caes.Statement{
		Text: "Mary is riding on the tandem.",
		Args: []*caes.Argument{&a2}}
	var st = caes.Statement{
		Text: "Suzy is riding on the tandem.",
		Args: []*caes.Argument{&a3}}
	var i1 caes.Issue
	var jmt = caes.Statement{
		Text:  "John and Mary are riding on the tandem.",
		Issue: &i1,
		Args:  []*caes.Argument{&a4}}
	var jst = caes.Statement{
		Text:  "John and Suzy are riding on the tandem.",
		Issue: &i1,
		Args:  []*caes.Argument{&a5}}
	var mst = caes.Statement{
		Text:  "Mary and Suzy are riding on the tandem.",
		Issue: &i1,
		Args:  []*caes.Argument{&a6}}
	i1 = caes.Issue{
		Standard:  caes.PE,
		Positions: []*caes.Statement{&jmt, &jst, &mst}}
	a1 = caes.Argument{Conclusion: &jt, Premises: []caes.Premise{caes.Premise{Stmt: &jw}}}
	a2 = caes.Argument{Conclusion: &mt, Premises: []caes.Premise{caes.Premise{Stmt: &mw}}}
	a3 = caes.Argument{Conclusion: &st, Premises: []caes.Premise{caes.Premise{Stmt: &sw}}}
	a4 = caes.Argument{
		Conclusion: &jmt,
		Premises:   []caes.Premise{caes.Premise{Stmt: &jt}, caes.Premise{Stmt: &mt}}}
	a5 = caes.Argument{
		Conclusion: &jst,
		Premises:   []caes.Premise{caes.Premise{Stmt: &jt}, caes.Premise{Stmt: &st}}}
	a6 = caes.Argument{
		Conclusion: &mst,
		Premises:   []caes.Premise{caes.Premise{Stmt: &mt}, caes.Premise{Stmt: &st}}}
	var ag = caes.ArgGraph{
		Issues:     []*caes.Issue{&i1},
		Statements: []*caes.Statement{&jw, &mw, &sw, &jt, &mt, &st, &jmt, &jst, &mst},
		Arguments:  []*caes.Argument{&a1, &a2, &a3, &a4, &a5, &a6}}
	l := ag.GroundedLabelling()
	expected := l[&jmt] == caes.Out &&
		l[&jst] == caes.Out &&
		l[&mst] == caes.Out
	if !expected {
		t.Errorf("TestTandem failed\n")
	}
}

// The following examples are from:
// Prakken, H. An abstract framework for argumentation with structured arguments.
// Argument & Computation 1, (2010), 93-124.

func TestBachelor(t *testing.T) {
	var i1 caes.Issue
	var a1, a2 caes.Argument
	var bachelor = caes.Statement{
		Text:  "Fred is a bachelor.",
		Issue: &i1,
		Args:  []*caes.Argument{&a1}}
	var married = caes.Statement{
		Text:  "Fred is married.",
		Issue: &i1,
		Args:  []*caes.Argument{&a2}}
	var wearsRing = caes.Statement{Text: "Fred wears a ring.", Assumed: true}
	var partyAnimal = caes.Statement{Text: "Fred is a pary animal.", Assumed: true}
	i1 = caes.Issue{
		Standard:  caes.PE,
		Positions: []*caes.Statement{&married, &bachelor}}
	a1 = caes.Argument{
		Conclusion: &bachelor,
		Premises:   []caes.Premise{caes.Premise{Stmt: &partyAnimal}}}
	a2 = caes.Argument{
		Conclusion: &married,
		Premises:   []caes.Premise{caes.Premise{Stmt: &wearsRing}}}
	var ag = caes.ArgGraph{
		Issues:     []*caes.Issue{&i1},
		Statements: []*caes.Statement{&bachelor, &married, &wearsRing, &partyAnimal},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	expected := l[&married] == caes.Out && l[&bachelor] == caes.Out
	if !expected {
		t.Errorf("TestBachelor failed\n")
		fmt.Printf("label(married)=%v\n", l[&married])
		fmt.Printf("label(bachelor)=%v\n", l[&bachelor])
		fmt.Printf("label(partyAnimal)=%v\n", l[&partyAnimal])
		fmt.Printf("label(wearsRing)=%v\n", l[&wearsRing])
	}
}

// The Frisian example, ibid., page 11

func TestFrisian(t *testing.T) {
	var a1, a2 caes.Argument
	var dutch = caes.Statement{
		Text: "Wiebe is Dutch.",
		Args: []*caes.Argument{&a1}}
	var tall = caes.Statement{
		Text: "Wiebe is tall.",
		Args: []*caes.Argument{&a2}}
	var frisian = caes.Statement{Text: "Wiebe is Frisian.", Assumed: true}
	a1 = caes.Argument{
		Conclusion: &dutch,
		Premises:   []caes.Premise{caes.Premise{Stmt: &frisian}}}
	a2 = caes.Argument{
		Conclusion: &tall,
		Premises:   []caes.Premise{caes.Premise{Stmt: &dutch}}}
	var ag = caes.ArgGraph{
		Statements: []*caes.Statement{&frisian, &dutch, &tall},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	expected := l[&tall] == caes.In && l[&dutch] == caes.In
	if !expected {
		t.Errorf("TestFrisian failed\n")
		fmt.Printf("label(dutch)=%v\n", l[&dutch])
		fmt.Printf("label(tall)=%v\n", l[&tall])
	}
}

// The next example is from "Relating Carneades with abstract argumentation
// via the ASPIC+ framework for structured argumentation", by Bas Gijzel and
// Henry Prakken. It is the example they use to illustrate the inability of
// Carneades to handle cycles. But, interestingly, in this formulation of the
// problem there are no cycles in the argument graph. Indeed, there are no
// arguments as well.

func TestVacation(t *testing.T) {
	var i1 caes.Issue
	var italy = caes.Statement{
		Text:  "Let's go to Italy.",
		Issue: &i1}
	var greece = caes.Statement{
		Text:  "Let's go to Greece.",
		Issue: &i1}
	i1 = caes.Issue{
		Standard:  caes.PE,
		Positions: []*caes.Statement{&greece, &italy}}
	var ag = caes.ArgGraph{
		Issues:     []*caes.Issue{&i1},
		Statements: []*caes.Statement{&greece, &italy}}
	l := ag.GroundedLabelling()
	expected := l[&greece] == caes.Out && l[&italy] == caes.Out
	if !expected {
		t.Errorf("TestVacation failed\n")
		fmt.Printf("label(greece)=%v\n", l[&greece])
		fmt.Printf("label(italy)=%v\n", l[&italy])
	}
}

// The library example, ibid., page 17

//func TestLibrary(t *testing.T) {
//	var a1, a2, a3 caes.Argument
//	var i1 caes.Issue
//	var snores = caes.Statement{
//		Text:    "The person is snoring in the library.",
//		Assumed: true}
//	var professor = caes.Statement{
//		Text:    "The person is a professor.",
//		Assumed: true}
//	var misbehaves = caes.Statement{
//		Text: "The person is misbehaving.",
//		Args: []*caes.Argument{&a1}}
//	var accessDenied = caes.Statement{
//		Text:  "The person is denied access to the library.",
//		Issue: &i1,
//		Args:  []*caes.Argument{&a2}}
//	var accessNotDenied = caes.Statement{
//		Text:  "The person is not denied access to the library.",
//		Issue: &i1,
//		Args:  []*caes.Argument{&a3}}
//	i1 = caes.Issue{
//		Standard:  caes.PE,
//		Positions: []*caes.Statement{&accessDenied, &accessNotDenied}}
//	a1 = caes.Argument{
//		Conclusion: &misbehaves,
//		Premises:   []caes.Premise{caes.Premise{Stmt: &snores}}}
//	a2 = caes.Argument{
//		Scheme: &caes.Scheme{
//			Eval: func(arg *caes.Argument, l caes.Labelling) float64 {
//				return 0.5
//			}},
//		Conclusion: &accessDenied,
//		Premises:   []caes.Premise{caes.Premise{Stmt: &misbehaves}}}
//	a3 = caes.Argument{
//		Scheme: &caes.Scheme{
//			Eval: func(arg *caes.Argument, l caes.Labelling) float64 {
//				return 0.6
//			}},
//		Conclusion: &accessNotDenied,
//		Premises:   []caes.Premise{caes.Premise{Stmt: &professor}}}
//	var ag = caes.ArgGraph{
//		Issues: []*caes.Issue{&i1},
//		Statements: []*caes.Statement{&snores, &professor, &misbehaves,
//			&accessDenied, &accessNotDenied},
//		Arguments: []*caes.Argument{&a1, &a2, &a3}}
//	l := ag.GroundedLabelling()
//	expected := l[&accessDenied] == caes.Out && l[&accessNotDenied] == caes.In
//	if !expected {
//		t.Errorf("TestLibrary failed\n")
//		fmt.Printf("label(accessDenied)=%v\n", l[&accessDenied])
//		fmt.Printf("label(accessNotDenied)=%v\n", l[&accessNotDenied])
//		fmt.Printf("label(misbehaves)=%v\n", l[&misbehaves])
//		fmt.Printf("label(snores)=%v\n", l[&snores])
//		fmt.Printf("label(professor)=%v\n", l[&professor])
//	}
//}

// Serial self defeat example, ibid., page 18
func TestSelfDefeat(t *testing.T) {
	var a1, a2 caes.Argument
	var i1 caes.Issue
	var P = caes.Statement{
		Text:    "Witness John says he is unreliable.",
		Assumed: true}
	var Q = caes.Statement{
		Text:  "Witness John is unreliable.",
		Issue: &i1,
		Args:  []*caes.Argument{&a1}}
	var R = caes.Statement{
		Text:  "Witness John is reliable.",
		Issue: &i1}
	var a1Invalid = caes.Statement{
		Text: "Argument a1 is invalid.",
		Args: []*caes.Argument{&a2}}
	a1 = caes.Argument{
		Conclusion:  &Q,
		Premises:    []caes.Premise{caes.Premise{Stmt: &P}},
		Undercutter: &a1Invalid}
	i1 = caes.Issue{
		Standard:  caes.PE,
		Positions: []*caes.Statement{&Q, &R}}
	a2 = caes.Argument{
		Conclusion: &a1Invalid,
		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
	var ag = caes.ArgGraph{
		Issues:     []*caes.Issue{&i1},
		Statements: []*caes.Statement{&P, &Q, &R, &a1Invalid},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	expected := l[&Q] == caes.Out && l[&R] == caes.Out && l[&a1Invalid] == caes.Out
	if !expected {
		t.Errorf("TestSelfDefeat failed\n")
		fmt.Printf("label(unreliable(John))=%v\n", l[&Q])
		fmt.Printf("label(reliable(John))=%v\n", l[&R])
		fmt.Printf("label(invalid(a1))=%v\n", l[&a1Invalid])
	}
}

func TestEvenLoop(t *testing.T) {
	var a1, a2 caes.Argument
	var P = caes.Statement{Id: "P", Text: "P"}
	var Q = caes.Statement{Id: "Q", Text: "Q"}
	a1 = caes.Argument{
		Conclusion: &Q,
		Premises:   []caes.Premise{caes.Premise{Stmt: &P}}}
	a2 = caes.Argument{
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
	var ag = caes.ArgGraph{
		Statements: []*caes.Statement{&P, &Q},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	expected := l[&P] == caes.Out && l[&Q] == caes.Out

	if !expected {
		t.Errorf("TestEvenLoop failed\n")
		fmt.Printf("label(P)=%v\n", l[&P])
		fmt.Printf("label(Q)=%v\n", l[&Q])
	}
}

func TestOddLoop2(t *testing.T) {
	var a1, a2, a3 caes.Argument
	var P = caes.Statement{Text: "P"}
	var Q = caes.Statement{Text: "Q"}
	var R = caes.Statement{Text: "R"}
	a1 = caes.Argument{
		Conclusion: &Q,
		Premises:   []caes.Premise{caes.Premise{Stmt: &P}}}
	a2 = caes.Argument{
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &R}}}
	a3 = caes.Argument{
		Conclusion: &R,
		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
	var ag = caes.ArgGraph{
		Statements: []*caes.Statement{&P, &Q, &R},
		Arguments:  []*caes.Argument{&a1, &a2, &a3}}
	l := ag.GroundedLabelling()
	expected := l[&P] == caes.Out && l[&Q] == caes.Out && l[&R] == caes.Out
	if !expected {
		t.Errorf("TestOddLoop2 failed\n")
		fmt.Printf("label(P)=%v\n", l[&P])
		fmt.Printf("label(Q)=%v\n", l[&Q])
		fmt.Printf("label(R)=%v\n", l[&R])
	}
}

func TestSelfDefeat2(t *testing.T) {
	var a1, a2 caes.Argument
	var i1 caes.Issue
	var P = caes.Statement{Text: "P", Issue: &i1, Args: []*caes.Argument{&a2}}
	var notP = caes.Statement{Text: "not P", Issue: &i1, Args: []*caes.Argument{&a1}}
	a1 = caes.Argument{
		Conclusion: &notP,
		Premises:   []caes.Premise{caes.Premise{Stmt: &P}}}
	a2 = caes.Argument{
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &notP}}}
	i1 = caes.Issue{Positions: []*caes.Statement{&P, &notP}}
	var ag = caes.ArgGraph{
		Issues:     []*caes.Issue{&i1},
		Statements: []*caes.Statement{&P, &notP},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	expected := l[&P] == caes.Out && l[&notP] == caes.Out
	if !expected {
		t.Errorf("TestSelfDefeat2 failed\n")
		fmt.Printf("label(P)=%v\n", l[&P])
		fmt.Printf("label(not P)=%v\n", l[&notP])
	}
}

// The arguments in TestSelfDefeat2 are irrelevant. We get the
// same labelling without them.
func TestSelfDefeat3(t *testing.T) {
	var i1 caes.Issue
	var P = caes.Statement{Text: "P", Issue: &i1}
	var notP = caes.Statement{Text: "not P", Issue: &i1}
	i1 = caes.Issue{Positions: []*caes.Statement{&P, &notP}}
	var ag = caes.ArgGraph{
		Issues:     []*caes.Issue{&i1},
		Statements: []*caes.Statement{&P, &notP},
		Arguments:  []*caes.Argument{}}
	l := ag.GroundedLabelling()
	expected := l[&P] == caes.Out && l[&notP] == caes.Out
	if !expected {
		t.Errorf("TestSelfDefeat3 failed\n")
		fmt.Printf("label(P)=%v\n", l[&P])
		fmt.Printf("label(not P)=%v\n", l[&notP])
	}
}

//func TestReinstatement2(t *testing.T) {
//	var a1, a2, a3 caes.Argument
//	var i1 caes.Issue
//	var P = caes.Statement{
//		Text:  "P",
//		Issue: &i1,
//		Args:  []*caes.Argument{&a1, &a3}}
//	var notP = caes.Statement{
//		Text:  "not P",
//		Issue: &i1,
//		Args:  []*caes.Argument{&a2}}
//	var Q = caes.Statement{
//		Text:    "Q",
//		Assumed: true}
//	var R = caes.Statement{
//		Text:    "R",
//		Assumed: true}
//	var S = caes.Statement{
//		Text:    "S",
//		Assumed: true}
//	i1 = caes.Issue{
//		Standard:  caes.PE,
//		Positions: []*caes.Statement{&P, &notP}}
//	a1 = caes.Argument{
//		Scheme: &caes.Scheme{
//			Eval: func(arg *caes.Argument, l caes.Labelling) float64 {
//				return 0.4
//			}},
//		Conclusion: &P,
//		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
//	a2 = caes.Argument{
//		Scheme: &caes.Scheme{
//			Eval: func(arg *caes.Argument, l caes.Labelling) float64 {
//				return 0.5
//			}},
//		Conclusion: &notP,
//		Premises:   []caes.Premise{caes.Premise{Stmt: &R}}}
//	a3 = caes.Argument{
//		Scheme: &caes.Scheme{
//			Eval: func(arg *caes.Argument, l caes.Labelling) float64 {
//				return 0.6
//			}},
//		Conclusion: &P,
//		Premises:   []caes.Premise{caes.Premise{Stmt: &S}}}
//	var ag = caes.ArgGraph{
//		Issues:     []*caes.Issue{&i1},
//		Statements: []*caes.Statement{&P, &notP, &Q, &R, &S},
//		Arguments:  []*caes.Argument{&a1, &a2, &a3}}
//	l := ag.GroundedLabelling()
//	expected := l[&P] == caes.In && l[&notP] == caes.Out
//	if !expected {
//		t.Errorf("TestReinstatement2 failed\n")
//		fmt.Printf("label(P)=%v\n", l[&P])
//		fmt.Printf("label(not P)=%v\n", l[&notP])
//	}
//}

func TestSupportLoop(t *testing.T) {
	var a1, a2 caes.Argument
	var P = caes.Statement{Text: "P", Args: []*caes.Argument{&a1}}
	var Q = caes.Statement{Text: "Q", Args: []*caes.Argument{&a2}}
	a1 = caes.Argument{
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
	a2 = caes.Argument{
		Conclusion: &Q,
		Premises:   []caes.Premise{caes.Premise{Stmt: &P}}}
	var ag = caes.ArgGraph{
		Statements: []*caes.Statement{&P, &Q},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	expected := l[&P] == caes.Out && l[&Q] == caes.Out
	if !expected {
		t.Errorf("TestSupportLoop failed\n")
		fmt.Printf("label(P)=%v\n", l[&P])
		fmt.Printf("label(Q)=%v\n", l[&Q])
	}
}

func TestIndependentSupportLoop(t *testing.T) {
	var a1, a2, a3 caes.Argument
	var P = caes.Statement{Text: "P", Args: []*caes.Argument{&a1, &a3}}
	var Q = caes.Statement{Text: "Q", Args: []*caes.Argument{&a2}}
	var R = caes.Statement{Text: "R", Assumed: true}
	a1 = caes.Argument{
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
	a2 = caes.Argument{
		Conclusion: &Q,
		Premises:   []caes.Premise{caes.Premise{Stmt: &P}}}
	a3 = caes.Argument{
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &R}}}
	var ag = caes.ArgGraph{
		Statements: []*caes.Statement{&P, &Q, &R},
		Arguments:  []*caes.Argument{&a1, &a2, &a3}}
	l := ag.GroundedLabelling()
	expected := l[&P] == caes.In && l[&Q] == caes.In && l[&R] == caes.In
	if !expected {
		t.Errorf("TestIndependentSupportLoop failed\n")
		fmt.Printf("label(P)=%v\n", l[&P])
		fmt.Printf("label(Q)=%v\n", l[&Q])
		fmt.Printf("label(R)=%v\n", l[&R])
	}
}

func TestApplyLabelling(t *testing.T) {
	// same AG as in the even-loop test
	var a1, a2 caes.Argument
	var P = caes.Statement{Id: "P", Text: "P"}
	var Q = caes.Statement{Id: "Q", Text: "Q"}
	a1 = caes.Argument{
		Id:         "a1",
		Conclusion: &Q,
		Premises:   []caes.Premise{caes.Premise{Stmt: &P}}}
	a2 = caes.Argument{
		Id:         "a2",
		Conclusion: &P,
		Premises:   []caes.Premise{caes.Premise{Stmt: &Q}}}
	var ag = caes.ArgGraph{
		Statements: []*caes.Statement{&P, &Q},
		Arguments:  []*caes.Argument{&a1, &a2}}
	l := ag.GroundedLabelling()
	fmt.Printf("labelling=%v\n", l)
	ag.ApplyLabelling(l)
	expected := l[&P] == P.Label && l[&Q] == Q.Label
	yaml.Export(os.Stdout, &ag)

	if !expected {
		t.Errorf("TestApplyLabelling failed\n")
		fmt.Printf("l[P]=%v; P.Label=%v\n", l[&P], P.Label)
		fmt.Printf("l[Q]=%v; Q.Label=%v\n", l[&Q], Q.Label)
	}
}
