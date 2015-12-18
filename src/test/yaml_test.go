package test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	"os"
	"testing"
)

func TestIOPorsche2(t *testing.T) {
	ioTest(t, "Porsche2.yml", "TempPorsche2.yml")
}

func TestIOBachelor(t *testing.T) {
	ioTest(t, "bachelor.yml", "TempBachelor.yml")
}

func TestIOCaminada1(t *testing.T) {
	ioTest(t, "caminada1.yml", "TempCaminada1.yml")
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

func TestIOTandem(t *testing.T) {
	ioTest(t, "tandem.yml", "TempTandem.yml")
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

func ioTest(t *testing.T, filename1 string, filename2 string) {

	var ag *caes.ArgGraph
	var err error
	fmt.Printf(" ------------------- Start Import %s -----------------\n", filename1)
	file, err := os.Open(yamlDir + filename1)
	check(t, err)
	ag, err = yaml.Import(file)

	check(t, err)
	// fmt.Printf("---------- WriteArgGraph %s ----------\n", filename1)
	// yaml.ExportWithReferences(os.Stdout, ag)
	// fmt.Printf("---------- End: WriteArgGraph %s ----------\n", filename1)
	fmt.Printf(" -   -  -  -  -  -   Start Groundlabeling & Checklabeling \n")
	l := ag.GroundedLabelling()
	// fmt.Printf("---------- printLabeling %s ----------\n", filename1)
	// printLabeling(l)
	// fmt.Printf("---------- End: printLabeling %s ----------\n", filename1)

	err = checkLabeling(l, ag.Statements)
	if err != nil {
		fmt.Printf(" check labeling fail: %s \n", err.Error())
	}
	// check(t, err)
	//	fmt.Printf("---------- Write ArgGraph 2 Yaml: %s ----------\n", filename1)
	//	yaml.Export(os.Stdout, ag)
	//	fmt.Printf("---------- End: Write ArgGraph 2 Yaml: %s ----------\n", filename1)

	fmt.Printf(" -  -  -  -  -  -  Start Export %s \n", yamlTmp+filename2)
	f, err := os.Create(yamlTmp + filename2)
	check(t, err)
	ag.ApplyLabelling(l)

	// fmt.Printf(" Export-Assumptions: %v \n", ag.Assumptions)
	yaml.Export(f, ag)

	fmt.Printf(" -  -  -  -  -  -  Start Import %s\n", yamlTmp+filename2)
	file, err = os.Open(yamlTmp + filename2)
	check(t, err)
	ag, err = yaml.Import(file)
	// fmt.Printf(" Import-Assumptions: %v \n", ag.Assumptions)
	check(t, err)
	// fmt.Printf("---------- WriteArgGraph 02  %s ----------\n", filename2)
	// yaml.ExportWithReferences(os.Stdout, ag)
	// fmt.Printf("---------- End: WriteArgGraph 02 %s ----------\n", filename2)

	fmt.Printf(" -  -  -  -  -  -  -  Start Import-GroundLabeling \n")
	// l = ag.GroundedLabelling()
	// fmt.Printf("---------- printLabeling %s ----------\n", filename2)
	// printLabeling(l)
	// fmt.Printf("---------- End: printLabeling %s ----------\n", filename2)
	// err = checkLabeling(l, ag.Statements)
	// check(t, err)
	// fmt.Printf("---------- Write ArgGraph 2 Yaml: %s ----------\n", filename2)
	// yaml.Export(os.Stdout, ag)
	// fmt.Printf("---------- End: Write ArgGraph 2 Yaml: %s ----------\n", filename2)

}
