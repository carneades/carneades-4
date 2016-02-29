// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package terms

import (
	"fmt"
	"testing"
)

func TestMatch1(t *testing.T) {
	//	checkErr := func(e error) {
	//		if e != nil {
	//			t.Errorf(e.Error())
	//		}
	//	}
	t1 := Atom("joe")
	t2 := Atom("sally")
	t3 := Compound{Functor: "parent", Args: []Term{t1, t2}}
	t4 := Compound{Functor: "parent", Args: []Term{Variable{Name: "X"}, Variable{Name: "Y"}}}
	_, ok := Match(t4, t3, nil)
	if ok == false {
		t.Errorf("TestMatch1 failed\n")
	}
}

func TestMatch2(t *testing.T) {
	// check that a variable is not bound to two different terms
	t3 := Compound{Functor: "parent", Args: []Term{Atom("joe"), Atom("sally")}}
	t4 := Compound{Functor: "parent", Args: []Term{Variable{Name: "X"}, Variable{Name: "X"}}}
	_, ok := Match(t4, t3, nil)
	if ok == true {
		t.Errorf("TestMatch2 failed\n")
	}
}

func TestEqual(t *testing.T) {
	t1 := Compound{Functor: "parent", Args: []Term{Atom("joe"), Atom("sally")}}
	t2 := Compound{Functor: "parent", Args: []Term{Atom("joe"), Atom("sally")}}
	if !Equal(t1, t2) {
		t.Errorf("TestEqual failed\n")
	}
}

func TestSubstitute(t *testing.T) {
	t1 := Compound{Functor: "parent", Args: []Term{Variable{Name: "X"}, Variable{Name: "Y"}}}
	t2 := Compound{Functor: "parent", Args: []Term{Atom("joe"), Atom("sally")}}
	env := AddBinding(Variable{Name: "Y"}, Atom("sally"), nil)
	env = AddBinding(Variable{Name: "X"}, Atom("joe"), env)
	t5 := Substitute(t1, env)
	if !Equal(t2, t5) {
		t.Errorf("TestSubstitute failed\n")
	}
}

func TestVariableChain1(t *testing.T) {
	t1 := Compound{Functor: "person", Args: []Term{Variable{Name: "X"}}}
	t2 := Compound{Functor: "person", Args: []Term{Atom("joe")}}
	env := AddBinding(Variable{Name: "Y"}, Atom("joe"), nil)
	env = AddBinding(Variable{Name: "X"}, Variable{Name: "Y"}, env)
	t5 := Substitute(t1, env)
	if !Equal(t2, t5) {
		t.Errorf("TestVariableChain1 failed\n")
	}
}

func TestVariableChain2(t *testing.T) {
	t1 := Compound{Functor: "person", Args: []Term{Variable{Name: "X"}}}
	t2 := Compound{Functor: "person", Args: []Term{Variable{Name: "Y"}}}
	env := AddBinding(Variable{Name: "X"}, Variable{Name: "Y"}, nil)
	t5 := Substitute(t1, env)
	if !Equal(t2, t5) {
		t.Errorf("TestVariableChain2 failed\n")
	}
}

func tunify(t *testing.T, str1, str2 string) bool {

	term1, ok := ReadString(str1)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan str1 in unify \"%s\" failed, term1: %s", str1, term1.String()))
		return false
	}

	term2, ok := ReadString(str2)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan str2 in unify \"%s\" failed, term2: %s", str2, term2.String()))
	}
	fmt.Printf("  Unitfy  %s  \n       mit  %s \n", term1.String(), term2.String())
	ok, env := Unify(term1, term2, nil)
	fmt.Printf("---Binding---\n")
	for ; env != nil; env = env.Next {
		fmt.Printf(" %s == %s \n", env.Var, env.T)
	}
	return ok
}

func TestUnify1(t *testing.T) {
	//	checkErr := func(e error) {
	//		if e != nil {
	//			t.Errorf(e.Error())
	//		}
	//	}
	t1 := Atom("joe")
	t2 := Atom("sally")
	t3 := Compound{Functor: "parent", Args: []Term{t1, t2}}
	t4 := Compound{Functor: "parent", Args: []Term{Variable{Name: "X"}, Variable{Name: "Y"}}}
	ok, _ := Unify(t4, t3, nil)
	if ok == false {
		t.Errorf("TestUnify1 failed\n")
	}
}

func TestUnify2(t *testing.T) {
	// check that a variable is not bound to two different terms
	t3 := Compound{Functor: "parent", Args: []Term{Atom("joe"), Atom("sally")}}
	t4 := Compound{Functor: "parent", Args: []Term{Variable{Name: "X"}, Variable{Name: "X"}}}
	ok, _ := Unify(t4, t3, nil)
	if ok == true {
		t.Errorf("TestUnify2 failed\n")
	}
}

func TestUnify3(t *testing.T) {
	ok := tunify(t, "[color(X), color(Y), mix(X,Y,Z), color(Z)]", "[color(blue), color(yellow), mix(blue,yellow,green), color(C)]")
	if !ok {
		t.Errorf("TestUnify3 failed\n")
	}
}

func TestUnify4(t *testing.T) {
	ok := tunify(t, "[p(X, f(a)), f(a, Z), f(g(a,D), h(c))]",
		"[p(a, f(X)), f(Y, b), f(g(a,b), E) ]")
	if !ok {
		t.Errorf("TestUnify4 failed\n")
	}
}

func TestUnify5(t *testing.T) {
	ok := tunify(t, "g(A,A)",
		"g(X,f(X))")
	if ok {
		t.Errorf("TestUnify5 failed\n")
	}
}

func TestUnify6(t *testing.T) {
	ok := tunify(t, "g(A, A)",
		"g(X, p(E,f(X,a)))")
	if ok {
		t.Errorf("TestUnify6 failed\n")
	}
}

func TestUnify7(t *testing.T) {
	ok := tunify(t, "p(X,X)",
		"p(Y,f(Y))")
	if ok {
		t.Errorf("TestUnify7 failed\n")
	}
}

func TestUnify8(t *testing.T) {
	ok := tunify(t, "f(g(a,D),E, E)",
		"f(g(a,b),h(Y),Y)")
	if ok {
		t.Errorf("TestUnify8 failed\n")
	}
}

func TestUnify9(t *testing.T) {
	ok := tunify(t, "f(D, g(a,D))",
		"f(X, X)")
	if ok {
		t.Errorf("TestUnify9 failed\n")
	}
}

func TestUnify10(t *testing.T) {
	ok := tunify(t,
		"p(f(A,A),g(B,B),B)",
		"p(f(X,Y),g(Y,Z),X)")
	if !ok {
		t.Errorf("TestUnify10 failed\n")
	}
}

func TestUnify11(t *testing.T) {
	ok := tunify(t,
		"p(f(A,A),g(B,B),B)",
		"p(f(X,Y),g(Y,Z),h(X))")
	if ok {
		t.Errorf("TestUnify11 failed\n")
	}
}
