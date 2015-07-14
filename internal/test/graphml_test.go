package test

import (
	// "fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	// "github.com/carneades/carneades-4/internal/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	//	"log"
	// "errors"
	"os"
	"testing"
)

const aifDir = "../../examples/AGs/YAML/"
const aifTmp = "/tmp"

func gmlcheck(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
		t.Skip("Skip Test")
	}
}

func TestIOGmlTandem(t *testing.T) {
	ioGmlTest(t, "tandem.yml", "TempTandem.graphml")
}

func TestIOGmlBachelor(t *testing.T) {
	ioGmlTest(t, "bachelor.yml", "TempBachelor.graphml")
}

func TestIOGmlFrisan(t *testing.T) {
	ioGmlTest(t, "frisian.yml", "TempFrisian.graphml")
}

func TestIOGmlJogging(t *testing.T) {
	ioGmlTest(t, "jogging.yml", "TempJogging.graphml")
}

func TestIOGmlSherlock(t *testing.T) {
	ioGmlTest(t, "sherlock.yml", "TempSherlock.graphml")
}

func TestIOGmlVacation(t *testing.T) {
	ioGmlTest(t, "vacation.yml", "TempVacation.graphml")
}

func ioGmlTest(t *testing.T, filename1 string, filename2 string) {

	var ag *caes.ArgGraph
	var err error

	file, err := os.Open(aifDir + filename1)
	gmlcheck(t, err)
	ag, err = yaml.Import(file)
	file.Close()
	gmlcheck(t, err)
	//	fmt.Printf("---------- WriteArgGraph %s ----------\n", filename1)
	//	yaml.ExportWithReferences(os.Stdout, ag)
	//	fmt.Printf("---------- End: WriteArgGraph %s ----------\n", filename1)
	l := ag.GroundedLabelling()
	ag.ApplyLabelling(l)
	//	fmt.Printf("---------- printLabeling %s ----------\n", filename1)
	//	printLabeling(l)
	//	fmt.Printf("---------- End: printLabeling %s ----------\n", filename1)
	//file, err = os.Create(aifTmp + filename2)
	//gmlcheck(t, err)
	//graphml.Export(file, *ag)

}
