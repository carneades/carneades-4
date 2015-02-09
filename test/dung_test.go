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
	l := af.GroundedLabelling()
	expected := dung.In
	actual := l.Get("1")
	if actual != expected {
		t.Errorf("expected %s, not %s.", expected, actual)
	}
}

func TestSelfAttack(t *testing.T) {
	args := []string{"1"}
	atks := make(map[string][]string)
	atks["1"] = []string{"1"}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	expected := dung.Undecided
	actual := l.Get("1")
	if actual != expected {
		t.Errorf("expected %s, not %s.", expected, actual)
	}
}

func TestAttackedArg(t *testing.T) {
	args := []string{"1", "2"}
	atks := make(map[string][]string)
	atks["1"] = []string{"2"}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	if v := l.Get("1"); v != dung.Out {
		t.Errorf("expected out, not %s.", v)
	}
	if v := l.Get("2"); v != dung.In {
		t.Errorf("expected in, not %s.", v)
	}
}

func TestReinstatement(t *testing.T) {
	args := []string{"1", "2", "3"}
	atks := make(map[string][]string)
	atks["1"] = []string{"2"}
	atks["2"] = []string{"3"}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	if v := l.Get("1"); v != dung.In {
		t.Errorf("expected in, not %s.", v)
	}
	if v := l.Get("2"); v != dung.Out {
		t.Errorf("expected out, not %s.", v)
	}
	if v := l.Get("3"); v != dung.In {
		t.Errorf("expected in, not %s.", v)
	}
}

func TestOddLoop(t *testing.T) {
	args := []string{"1", "2", "3"}
	atks := make(map[string][]string)
	atks["1"] = []string{"2"}
	atks["2"] = []string{"3"}
	atks["3"] = []string{"1"}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	if v := l.Get("1"); v != dung.Undecided {
		t.Errorf("expected undecided, not %s.", v)
	}
	if v := l.Get("2"); v != dung.Undecided {
		t.Errorf("expected undecided, not %s.", v)
	}
	if v := l.Get("3"); v != dung.Undecided {
		t.Errorf("expected undecided, not %s.", v)
	}
}

func TestEqualArgSets(t *testing.T) {
	args1 := dung.NewArgSet("1", "2", "3")
	args2 := dung.NewArgSet("3", "2", "1")
	expected := true
	actual := dung.EqualArgSets(args1, args2)
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
	if !dung.EqualAFs(af, expected) {
		t.Errorf("expected %s, not %s.\n", expected.String(), af.String())
	}
}

func TestAf2GroundedLabelling(t *testing.T) {
	inFile, err := os.Open("AFs/reinstatement1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	extension := af.GroundedLabelling().AsExtension()
	expected := dung.NewArgSet("3", "1")
	if !dung.EqualArgSets(extension, expected) {
		t.Errorf("expected %s, not %s.\n", expected, extension)
	}
}

func TestEvenCycle1PreferredLabelling(t *testing.T) {
	inFile, err := os.Open("AFs/even_cycle1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	l := af.PreferredLabellings3()
	actual := []dung.ArgSet{}
	for _, labelling := range l {
		actual = append(actual, labelling.AsExtension())
	}
	e1 := dung.NewArgSet("1")
	e2 := dung.NewArgSet("2")
	expected := []dung.ArgSet{e1, e2}
	//fmt.Printf("actual: %v\n", actual)
	//fmt.Printf("expected: %v\n", expected)
	if !dung.EqualArgSetSlices(actual, expected) {
		t.Errorf("expected %s, not %s.\n", expected, actual)
	}
}
