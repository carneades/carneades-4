package test

import (
	// "fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	// "errors"
	"os"
	"testing"
)

const aifDir = "../../examples/AGs/YAML/"
const aifTmp = "/tmp/"

func gmlcheck(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
		t.Skip("Skip Test")
	}
}

func TestIOGmlBachelor(t *testing.T) {
	ioGmlTest(t, "bachelor.yml", "TempBachelor.graphml")
}

func TestIOGmlCaminada1(t *testing.T) {
	ioGmlTest(t, "caminada1.yml", "TempCaminada1.graphml")
}

func TestIOGmlDungAFs(t *testing.T) {
	ioGmlTest(t, "dung-AFs.yml", "TempDung-AFs.graphml")
}

func TestIOGmlEvenLoop(t *testing.T) {
	ioGmlTest(t, "even-loop.yml", "TempEven-loop.graphml")
}

func TestIOGmlFalseDilemma(t *testing.T) {
	ioGmlTest(t, "false-dilemma.yml", "TempFalse-dilemma.graphml")
}

func TestIOGmlFrisan(t *testing.T) {
	ioGmlTest(t, "frisian.yml", "TempFrisian.graphml")
}

func TestIOGmlIndependentSupportLoop(t *testing.T) {
	ioGmlTest(t, "independent-support-loop.yml", "TempIndependent-support-loop.graphml")
}

func TestIOGmlJogging(t *testing.T) {
	ioGmlTest(t, "jogging.yml", "TempJogging.graphml")
}

func TestIOGmlLibrary(t *testing.T) {
	ioGmlTest(t, "library.yml", "TempLibrary.graphml")
}

func TestIOGmlMandatorySentences(t *testing.T) {
	ioGmlTest(t, "mandatory-sentences.yml", "TempMandatory-sentences.graphml")
}

func TestIOGmlMcda1(t *testing.T) {
	ioGmlTest(t, "mcda1.yml", "TempMcda1.graphml")
}

func TestIOGmlOddLoop(t *testing.T) {
	ioGmlTest(t, "odd-loop.yml", "TempOdd-loop.graphml")
}

func TestIOGmlParaconsistency(t *testing.T) {
	ioGmlTest(t, "paraconsistency.yml", "TempParaconsistency.graphml")
}

func TestIOGmlPollockRedLight(t *testing.T) {
	ioGmlTest(t, "pollock-red-light.yml", "TempPollock-red-light.graphml")
}

func TestIOGmlPorsche(t *testing.T) {
	ioGmlTest(t, "porsche.yml", "TempPorsche.graphml")
}

func TestIOGmlPrakkenSartorMurder(t *testing.T) {
	ioGmlTest(t, "prakken-sartor-murder.yml", "TempPrakken-sartor-murder.graphml")
}

func TestIOGmlReinstatement(t *testing.T) {
	ioGmlTest(t, "reinstatement.yml", "TempReinstatement.graphml")
}

func TestIOGmlRiskyOperation(t *testing.T) {
	ioGmlTest(t, "risky-operation.yml", "TempRisky-operation.graphml")
}

func TestIOGmlSelfDefeat(t *testing.T) {
	ioGmlTest(t, "self-defeat.yml", "TempSelf-defeat.graphml")
}

func TestIOGmlSherlock(t *testing.T) {
	ioGmlTest(t, "sherlock.yml", "TempSherlock.graphml")
}

func TestIOGmlSnake(t *testing.T) {
	ioGmlTest(t, "snake.yml", "TempSnake.graphml")
}

func TestIOGmlSupportLoop(t *testing.T) {
	ioGmlTest(t, "support-loop.yml", "TempSupport-loop.graphml")
}

func TestIOGmlTandem(t *testing.T) {
	ioGmlTest(t, "tandem.yml", "TempTandem.graphml")
}

func TestIOGmlToulmin(t *testing.T) {
	ioGmlTest(t, "toulmin.yml", "TempToulmin.graphml")
}

func TestIOGmlTrivial(t *testing.T) {
	ioGmlTest(t, "trivial.yml", "TempTrivial.graphml")
}

func TestIOGmlTweety(t *testing.T) {
	ioGmlTest(t, "tweety.yml", "TempTweety.graphml")
}

func TestIOGmlUnreliableWitness(t *testing.T) {
	ioGmlTest(t, "unreliable-witness.yml", "TempUnreliable-witness.graphml")
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
	file, err = os.Create(aifTmp + filename2)
	gmlcheck(t, err)
	graphml.Export(file, ag)

}
