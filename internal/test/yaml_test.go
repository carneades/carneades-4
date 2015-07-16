package test

import (
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	//	"log"
	"errors"
	"os"
	"testing"
)

const yamlDir = "../../examples/AGs/YAML/"
const yamlTmp = "/tmp/"

func check(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
		t.Skip("Skip Test")
	}
}

func TestIOTandem(t *testing.T) {
	ioTest(t, "tandem.yml", "TempTandem.yml")
}

func TestIOBachelor(t *testing.T) {
	ioTest(t, "bachelor.yml", "TempBachelor.yml")
}

func TestIOFrisan(t *testing.T) {
	ioTest(t, "frisian.yml", "TempFrisian.yml")
}

func TestIOJogging(t *testing.T) {
	ioTest(t, "jogging.yml", "TempJogging.yml")
}

func TestIOSherlock(t *testing.T) {
	ioTest(t, "sherlock.yml", "TempSherlock.yml")
}

func TestIOVacation(t *testing.T) {
	ioTest(t, "vacation.yml", "TempVacation.yml")
}

func TestIOEvenLoop(t *testing.T) {
	ioTest(t, "even-loop.yml", "TempEvenLoop.yml")
}

func TestIOSelfDefeat(t *testing.T) {
	ioTest(t, "self-defeat.yml", "TempSelfDefeat.yml")
}

func TestIOOddLoop(t *testing.T) {
	ioTest(t, "odd-loop.yml", "TempOddLoop.yml")
}

func TestIOUnreliableWitness(t *testing.T) {
	ioTest(t, "unreliable-witness.yml", "TempUnreliableWitness.yml")
}

func checkLabeling(l caes.Labelling, stats []*caes.Statement) error {
	errStr := ""
	for _, stat := range stats {
		lbl := l[stat]
		if stat.Label != lbl {
			if errStr == "" {
				errStr = fmt.Sprintf(" statement: %s, expected Label: %v, calculated Label: %v \n", stat.Id, stat.Label, lbl)
			} else {
				errStr = fmt.Sprintf("%s statement: %s, expected Label: %v, calculated Label: %v \n", errStr, stat.Id, stat.Label, lbl)
			}
		}
	}
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}

func printLabeling(l caes.Labelling) {

	for ref_stat, lbl := range l {
		fmt.Printf(" statement: %s    Label: %v\n", ref_stat.Id, lbl)
	}

}

func ioTest(t *testing.T, filename1 string, filename2 string) {

	var ag *caes.ArgGraph
	var err error
	file, err := os.Open(yamlDir + filename1)
	check(t, err)
	ag, err = yaml.Import(file)

	check(t, err)
	// fmt.Printf("---------- WriteArgGraph %s ----------\n", filename1)
	// yaml.ExportWithReferences(os.Stdout, ag)
	// fmt.Printf("---------- End: WriteArgGraph %s ----------\n", filename1)
	l := ag.GroundedLabelling()
	// fmt.Printf("---------- printLabeling %s ----------\n", filename1)
	// printLabeling(l)
	// fmt.Printf("---------- End: printLabeling %s ----------\n", filename1)

	err = checkLabeling(l, ag.Statements)
	check(t, err)
	//	fmt.Printf("---------- Write ArgGraph 2 Yaml: %s ----------\n", filename1)
	//	yaml.Export(os.Stdout, ag)
	//	fmt.Printf("---------- End: Write ArgGraph 2 Yaml: %s ----------\n", filename1)

	f, err := os.Create(yamlTmp + filename2)
	check(t, err)
	yaml.Export(f, ag)
	file, err = os.Open(yamlTmp + filename2)
	check(t, err)
	ag, err = yaml.Import(file)
	check(t, err)
	// fmt.Printf("---------- WriteArgGraph 02  %s ----------\n", filename2)
	// yaml.ExportWithReferences(os.Stdout, ag)
	// fmt.Printf("---------- End: WriteArgGraph 02 %s ----------\n", filename2)
	l = ag.GroundedLabelling()
	// fmt.Printf("---------- printLabeling %s ----------\n", filename2)
	// printLabeling(l)
	// fmt.Printf("---------- End: printLabeling %s ----------\n", filename2)
	err = checkLabeling(l, ag.Statements)
	check(t, err)
	// fmt.Printf("---------- Write ArgGraph 2 Yaml: %s ----------\n", filename2)
	// yaml.Export(os.Stdout, ag)
	// fmt.Printf("---------- End: Write ArgGraph 2 Yaml: %s ----------\n", filename2)

}
