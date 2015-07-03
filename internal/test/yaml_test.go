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

func check(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
		t.Skip("Skip Test")
	}
}

func TestIOTandem(t *testing.T) {
	ioTest(t, "AGs/tandem.yml", "AGs/TempTandem.yml")
}

func TestIOBachelor(t *testing.T) {
	ioTest(t, "AGs/bachelor.yml", "AGs/TempBachelor.yml")
}

func TestIOFrisan(t *testing.T) {
	ioTest(t, "AGs/frisian.yml", "AGs/TempFrisian.yml")
}

func TestIOJogging(t *testing.T) {
	ioTest(t, "AGs/jogging.yml", "AGs/TempJogging.yml")
}

func TestIOSherlock(t *testing.T) {
	ioTest(t, "AGs/sherlock.yml", "AGs/TempSherlock.yml")
}

func TestIOVacation(t *testing.T) {
	ioTest(t, "AGs/vacation.yml", "AGs/TempVacation.yml")
}

func checkLabeling(l caes.Labelling, stats []*caes.Statement) error {
	errStr := ""
	for _, stat := range stats {
		lbl := l.Get(stat)
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
	file, err := os.Open(filename1)
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

	f, err := os.Create(filename2)
	check(t, err)
	yaml.Export(f, ag)

	file, err = os.Open(filename2)
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
	//	fmt.Printf("---------- Write ArgGraph 2 Yaml: %s ----------\n", filename2)
	//	yaml.Export(os.Stdout, ag)
	//	fmt.Printf("---------- End: Write ArgGraph 2 Yaml: %s ----------\n", filename2)

}
