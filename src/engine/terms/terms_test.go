// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package terms

import (
	// "fmt"
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
	env := make(Bindings)
	ok := Match(t4, t3, env)
	if ok == false {
		t.Errorf("TestMatch1 failed\n")
	}
}

func TestMatch2(t *testing.T) {
	// check that a variable is not bound to two different terms
	t3 := Compound{Functor: "parent", Args: []Term{Atom("joe"), Atom("sally")}}
	t4 := Compound{Functor: "parent", Args: []Term{Variable{Name: "X"}, Variable{Name: "X"}}}
	env := make(Bindings)
	ok := Match(t4, t3, env)
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
	env := Bindings{"X": Atom("joe"), "Y": Atom("sally")}
	t5 := Substitute(t1, env)
	if !Equal(t2, t5) {
		t.Errorf("TestSubstitute failed\n")
	}
}

func TestVariableChain1(t *testing.T) {
	t1 := Compound{Functor: "person", Args: []Term{Variable{Name: "X"}}}
	t2 := Compound{Functor: "person", Args: []Term{Atom("joe")}}
	env := Bindings{"X": Variable{Name: "Y"}, "Y": Atom("joe")}
	t5 := Substitute(t1, env)
	if !Equal(t2, t5) {
		t.Errorf("TestVariableChain1 failed\n")
	}
}

func TestVariableChain2(t *testing.T) {
	t1 := Compound{Functor: "person", Args: []Term{Variable{Name: "X"}}}
	t2 := Compound{Functor: "person", Args: []Term{Variable{Name: "Y"}}}
	env := Bindings{"X": Variable{Name: "Y"}}
	t5 := Substitute(t1, env)
	if !Equal(t2, t5) {
		t.Errorf("TestVariableChain2 failed\n")
	}
}
