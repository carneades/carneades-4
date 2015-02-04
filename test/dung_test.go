package test

import (
	"../engine/dung"
	"../serialization/tgf"
	"log"
	"os"
	"testing"
)

func TestUnattackedArg(t *testing.T) {
	af := dung.NewAF([]dung.Arg{tgf.Arg("1")},
		map[dung.Arg][]dung.Arg{})
	l := af.GroundedLabelling()
	expected := dung.In
	actual := l.Get(tgf.Arg("1"))
	if actual != expected {
		t.Errorf("expected %s, not %s.", expected, actual)
	}
}

func TestSelfAttack(t *testing.T) {
	args := []dung.Arg{tgf.Arg("1")}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[tgf.Arg("1")] = []dung.Arg{tgf.Arg("1")}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	expected := dung.Undecided
	actual := l.Get(tgf.Arg("1"))
	if actual != expected {
		t.Errorf("expected %s, not %s.", expected, actual)
	}
}

func TestAttackedArg(t *testing.T) {
	args := []dung.Arg{tgf.Arg("1"), tgf.Arg("2")}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[tgf.Arg("1")] = []dung.Arg{tgf.Arg("2")}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	if v := l.Get(tgf.Arg("1")); v != dung.Out {
		t.Errorf("expected out, not %s.", v)
	}
	if v := l.Get(tgf.Arg("2")); v != dung.In {
		t.Errorf("expected in, not %s.", v)
	}
}

func TestReinstatement(t *testing.T) {
	args := []dung.Arg{tgf.Arg("1"), tgf.Arg("2"), tgf.Arg("3")}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[tgf.Arg("1")] = []dung.Arg{tgf.Arg("2")}
	atks[tgf.Arg("2")] = []dung.Arg{tgf.Arg("3")}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	if v := l.Get(tgf.Arg("1")); v != dung.In {
		t.Errorf("expected in, not %s.", v)
	}
	if v := l.Get(tgf.Arg("2")); v != dung.Out {
		t.Errorf("expected out, not %s.", v)
	}
	if v := l.Get(tgf.Arg("3")); v != dung.In {
		t.Errorf("expected in, not %s.", v)
	}
}

func TestOddLoop(t *testing.T) {
	args := []dung.Arg{tgf.Arg("1"), tgf.Arg("2"), tgf.Arg("3")}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[tgf.Arg("1")] = []dung.Arg{tgf.Arg("2")}
	atks[tgf.Arg("2")] = []dung.Arg{tgf.Arg("3")}
	atks[tgf.Arg("3")] = []dung.Arg{tgf.Arg("1")}
	af := dung.NewAF(args, atks)
	l := af.GroundedLabelling()
	if v := l.Get(tgf.Arg("1")); v != dung.Undecided {
		t.Errorf("expected undecided, not %s.", v)
	}
	if v := l.Get(tgf.Arg("2")); v != dung.Undecided {
		t.Errorf("expected undecided, not %s.", v)
	}
	if v := l.Get(tgf.Arg("3")); v != dung.Undecided {
		t.Errorf("expected undecided, not %s.", v)
	}
}

func TestEqualArgSets(t *testing.T) {
	args1 := dung.NewArgSet(tgf.Arg("1"), tgf.Arg("2"), tgf.Arg("3"))
	args2 := dung.NewArgSet(tgf.Arg("3"), tgf.Arg("2"), tgf.Arg("1"))
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
	args := []dung.Arg{tgf.Arg("1"), tgf.Arg("2"), tgf.Arg("3")}
	atks := make(map[dung.Arg][]dung.Arg)
	atks[tgf.Arg("2")] = []dung.Arg{tgf.Arg("1")}
	atks[tgf.Arg("3")] = []dung.Arg{tgf.Arg("2")}
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
	expected := dung.NewArgSet(tgf.Arg("3"), tgf.Arg("1"))
	if !dung.EqualArgSets(extension, expected) {
		t.Errorf("expected %s, not %s.\n", expected, extension)
	}
}

func TestAf5PreferredLabelling(t *testing.T) {
	inFile, err := os.Open("AFs/even_cycle1.tgf")
	if err != nil {
		log.Fatal(err)
	}
	af, err := tgf.Import(inFile)
	l := af.PreferredLabellings()
	actual := []dung.ArgSet{}
	for _, labelling := range l {
		actual = append(actual, labelling.AsExtension())
	}
	e1 := dung.NewArgSet(tgf.Arg("1"))
	e2 := dung.NewArgSet(tgf.Arg("2"))
	expected := []dung.ArgSet{e1, e2}
	if !dung.EqualArgSetSlices(actual, expected) {
		t.Errorf("expected %s, not %s.\n", expected, actual)
	}
}
