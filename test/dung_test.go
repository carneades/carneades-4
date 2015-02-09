package test

import (
	"../engine/dung"
	"../serialization/tgf"
	// "fmt"
	"log"
	"os"
	"testing"
)

func TestUnattackedArg(t *testing.T) {
	af := dung.NewAF([]string{"1"},
		map[string][]string{})
	l := af.GroundedExtension()
	expected := true
	actual := l.Contains("1")
	if actual != expected {
		t.Errorf("expected extension to contain 1")
	}
}

func TestSelfAttack(t *testing.T) {
	args := []string{"1"}
	atks := make(map[string][]string)
	atks["1"] = []string{"1"}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	expected := false
	actual := l.Contains("1")
	if actual != expected {
		t.Errorf("expected extension to not contain 1")
	}
}

func TestAttackedArg(t *testing.T) {
	args := []string{"1", "2"}
	atks := make(map[string][]string)
	atks["1"] = []string{"2"}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	if l.Contains("1") {
		t.Errorf("expected 1 to be out")
	}
	if !l.Contains("2") {
		t.Errorf("expected 2 to be in")
	}
}

func TestReinstatement(t *testing.T) {
	args := []string{"1", "2", "3"}
	atks := make(map[string][]string)
	atks["1"] = []string{"2"}
	atks["2"] = []string{"3"}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	if !l.Contains("1") {
		t.Errorf("expected 1 to be in")
	}
	if l.Contains("2") {
		t.Errorf("expected 2 to be out")
	}
	if !l.Contains("3") {
		t.Errorf("expected 3 to be in")
	}
}

func TestOddLoop(t *testing.T) {
	args := []string{"1", "2", "3"}
	atks := make(map[string][]string)
	atks["1"] = []string{"2"}
	atks["2"] = []string{"3"}
	atks["3"] = []string{"1"}
	af := dung.NewAF(args, atks)
	l := af.GroundedExtension()
	if l.Contains("1") {
		t.Errorf("expected 1 to be out")
	}
	if l.Contains("2") {
		t.Errorf("expected 2 to be out")
	}
	if l.Contains("3") {
		t.Errorf("expected 3 to be out")
	}
}

func TestEqualArgSets(t *testing.T) {
	args1 := dung.NewArgSet("1", "2", "3")
	args2 := dung.NewArgSet("3", "2", "1")
	expected := true
	actual := args1.Equals(args2)
	if expected != actual {
		t.Errorf("expected EqualArgSets(%s,%s)", args1, args2)
	}
}

func TestAf2Import(t *testing.T) {
	inFile, err := os.Open("AFs/reinstatement1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	args := []string{"1", "2", "3"}
	atks := make(map[string][]string)
	atks["2"] = []string{"1"}
	atks["3"] = []string{"2"}
	expected := dung.NewAF(args, atks)
	if !af.Equals(expected) {
		t.Errorf("expected %s, not %s.\n", expected.String(), af.String())
	}
}

func TestAf2GroundedLabelling(t *testing.T) {
	inFile, err := os.Open("AFs/reinstatement1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	extension := af.GroundedExtension()
	expected := dung.NewArgSet("3", "1")
	if !extension.Equals(expected) {
		t.Errorf("expected %s, not %s.\n", expected, extension)
	}
}

func TestEvenCycle1PreferredLabelling(t *testing.T) {
	inFile, err := os.Open("AFs/even_cycle1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	actual := af.PreferredExtensions()
	e1 := dung.NewArgSet("1")
	e2 := dung.NewArgSet("2")
	expected := []dung.ArgSet{e1, e2}
	//fmt.Printf("actual: %v\n", actual)
	//fmt.Printf("expected: %v\n", expected)
	if !dung.EqualArgSetSlices(actual, expected) {
		t.Errorf("expected %s, not %s.\n", expected, actual)
	}
}
