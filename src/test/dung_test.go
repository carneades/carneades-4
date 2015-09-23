// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package test

import (
	"github.com/carneades/carneades-4/internal/engine/dung"
	"github.com/carneades/carneades-4/internal/engine/dung/encoding/tgf"
	// "fmt"
	"log"
	"os"
	"testing"
)

const dungDir = "../../examples/AFs/TGF/"

const a1 = dung.Arg("1")
const a2 = dung.Arg("2")
const a3 = dung.Arg("3")

func TestUnattackedArg(t *testing.T) {
	af := dung.NewAF([]dung.Arg{a1},
		map[dung.Arg][]dung.Arg{})
	l := af.GroundedExtension()
	expected := true
	actual := l.Contains(a1)
	if actual != expected {
		t.Errorf("expected extension to contain 1")
	}
}

func TestSelfAttack(t *testing.T) {
	args := []dung.Arg{a1}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[a1] = []dung.Arg{a1}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	expected := false
	actual := l.Contains(a1)
	if actual != expected {
		t.Errorf("expected extension to not contain 1")
	}
}

func TestAttackedArg(t *testing.T) {
	args := []dung.Arg{a1, a2}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[a1] = []dung.Arg{a2}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	if l.Contains(a1) {
		t.Errorf("expected 1 to be out")
	}
	if !l.Contains(a2) {
		t.Errorf("expected 2 to be in")
	}
}

func TestReinstatement(t *testing.T) {
	args := []dung.Arg{a1, a2, a3}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[a1] = []dung.Arg{a2}
	atks[a2] = []dung.Arg{a3}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	if !l.Contains(a1) {
		t.Errorf("expected 1 to be in")
	}
	if l.Contains(a2) {
		t.Errorf("expected 2 to be out")
	}
	if !l.Contains(a3) {
		t.Errorf("expected 3 to be in")
	}
}

func TestOddLoop(t *testing.T) {
	args := []dung.Arg{a1, a2, a2}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[a1] = []dung.Arg{a2}
	atks[a2] = []dung.Arg{a2}
	atks[a2] = []dung.Arg{a1}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	if l.Contains(a1) {
		t.Errorf("expected 1 to be out")
	}
	if l.Contains(a2) {
		t.Errorf("expected 2 to be out")
	}
	if l.Contains(a2) {
		t.Errorf("expected 3 to be out")
	}
}

func TestEqualArgSets(t *testing.T) {
	args1 := dung.NewArgSet(a1, a2, a2)
	args2 := dung.NewArgSet(a2, a2, a1)
	expected := true
	actual := args1.Equals(args2)
	if expected != actual {
		t.Errorf("expected EqualArgSets(%s,%s)", args1, args2)
	}
}

func TestAf2Import(t *testing.T) {
	inFile, err := os.Open(dungDir + "reinstatement1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	args := []dung.Arg{a1, a2, a3}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[a2] = []dung.Arg{a1}
	atks[a3] = []dung.Arg{a2}
	expected := dung.NewAF(args, atks)
	if !af.Equals(expected) {
		t.Errorf("expected %s, not %s.\n", expected.String(), af.String())
	}
}

func TestAf2GroundedLabelling(t *testing.T) {
	inFile, err := os.Open(dungDir + "reinstatement1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	extension := af.GroundedExtension()
	expected := dung.NewArgSet(a3, a1)
	if !extension.Equals(expected) {
		t.Errorf("expected %s, not %s.\n", expected, extension)
	}
}

func TestEvenCycle1PreferredLabelling(t *testing.T) {
	inFile, err := os.Open(dungDir + "even_cycle1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	actual := af.PreferredExtensions()
	e1 := dung.NewArgSet(a1)
	e2 := dung.NewArgSet(a2)
	expected := []dung.ArgSet{e1, e2}
	//fmt.Printf("actual: %v\n", actual)
	//fmt.Printf("expected: %v\n", expected)
	if !dung.EqualArgSetSlices(actual, expected) {
		t.Errorf("expected %s, not %s.\n", expected, actual)
	}
}
