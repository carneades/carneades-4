package test

import (
	// "fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	cjson "github.com/carneades/carneades-4/src/engine/caes/encoding/json"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	"os"
	"testing"
)

const jsonDir = "../../examples/AGs/JSON/"

// const yamlDir = "../../examples/AGs/YAML"

const jsonTmp = "/tmp/"

func ioJsonTest(t *testing.T, filename1 string, filename2 string) {
	// var ag  *caes.ArgGraph
	var ag2 *caes.ArgGraph
	var err error
	/*
		// Import a YAML-file from examples/AGs/YAML-Dir
		// ----------------------------------------------
		file, err := os.Open(yamlDir + filename1)
		check(t, err)
		ag, err = yaml.Import(file)
		file.Close()
		check(t, err)

		// file, err = os.Create(jsonTmp + "_a_" + filename1)
		// check(t, err)
		// yaml.ExportWithReferences(file, ag)

		l := ag.GroundedLabelling()
		// ------------------------

		//	file, err = os.Create(aifTmp + "_b_" + filename1)
		//	check(t, err)
		//	yaml.ExportWithReferences(file, ag)

		//	fmt.Printf(" ## Labeling yaml-Datei: \n")
		//	printLabeling(l)
		err = checkLabeling(l, ag.Statements)
		check(t, err)

		ag.ApplyLabelling(l)

		//	fmt.Printf(" ## ApplyLabeling")
		//	file, err = os.Create(aifTmp + "_c_" + filename1)
		//	check(t, err)
		//	yaml.ExportWithReferences(file, ag)

		// Export JSON-file to Temp-Dir
		// -----------------------------
		file, err = os.Create(jsonTmp + "_d_" + filename2)
		check(t, err)
		cjson.Export(file, ag)
		file.Close()
		file.Sync()
	*/
	// Import JSON-file from Temp-Dir
	// ------------------------------
	// file, err = os.Open(jsonTmp + filename2)
	// Import JSON-file from JSON-Dir
	// ------------------------------
	file, err := os.Open(jsonDir + filename2)
	check(t, err)
	ag2, err = cjson.Import(file)

	check(t, err)

	//	file, err = os.Create(aifTmp + "_e_" + filename1)
	//	check(t, err)
	//	yaml.ExportWithReferences(file, ag2)

	l2 := ag2.GroundedLabelling()
	// -----------------------
	//	fmt.Printf(" ## Labeling json-Datei: \n")
	//	printLabeling(l2)
	err = checkLabeling(l2, ag2.Statements)
	check(t, err)

	// Export YAML-file in Temp-Dir with the file-name: json2<name>.yml
	// ----------------------------------------------------------------
	//	file, err = os.Create(aifTmp + "_f_json2" + filename1)
	//	check(t, err)
	//	yaml.ExportWithReferences(file, ag2)

	file, err = os.Create(jsonTmp + "json2" + filename1)
	check(t, err)
	yaml.Export(file, ag2)

}

func TestJsonBachelor(t *testing.T) {
	ioJsonTest(t, "bachelor.yml", "Bachelor.json")
}

func TestJsonCaminada1(t *testing.T) {
	ioJsonTest(t, "caminada1.yml", "Caminada1.json")
}

func TestJsonDungAFs(t *testing.T) {
	ioJsonTest(t, "dung-AFs.yml", "Dung-AFs.json")
}

func TestJsonEvenLoop(t *testing.T) {
	ioJsonTest(t, "even-loop.yml", "Even-loop.json")
}

func TestJsonFalseDilemma(t *testing.T) {
	ioJsonTest(t, "false-dilemma.yml", "False-dilemma.json")
}

func TestJsonFrisan(t *testing.T) {
	ioJsonTest(t, "frisian.yml", "Frisian.json")
}

func TestJsonIndependentSupportLoop(t *testing.T) {
	ioJsonTest(t, "independent-support-loop.yml", "Independent-support-loop.json")
}

func TestJsonJogging(t *testing.T) {
	ioJsonTest(t, "jogging.yml", "Jogging.json")
}

func TestJsonLibrary(t *testing.T) {
	ioJsonTest(t, "library.yml", "Library.json")
}

func TestJsonMandatorySentences(t *testing.T) {
	ioJsonTest(t, "mandatory-sentences.yml", "Mandatory-sentences.json")
}

func TestJsonMcda1(t *testing.T) {
	ioJsonTest(t, "mcda1.yml", "Mcda1.json")
}

func TestJsonOddLoop(t *testing.T) {
	ioJsonTest(t, "odd-loop.yml", "Odd-loop.json")
}

func TestJsonParaconsistency(t *testing.T) {
	ioJsonTest(t, "paraconsistency.yml", "Paraconsistency.json")
}

func TestJsonPollockRedLight(t *testing.T) {
	ioJsonTest(t, "pollock-red-light.yml", "Pollock-red-light.json")
}

func TestJsonPorsche(t *testing.T) {
	ioJsonTest(t, "porsche.yml", "Porsche.json")
}

func TestJsonPrakkenSartorMurder(t *testing.T) {
	ioJsonTest(t, "prakken-sartor-murder.yml", "Prakken-sartor-murder.json")
}

func TestJsonReinstatement(t *testing.T) {
	ioJsonTest(t, "reinstatement.yml", "Reinstatement.json")
}

func TestJsonRiskyOperation(t *testing.T) {
	ioJsonTest(t, "risky-operation.yml", "Risky-operation.json")
}

func TestJsonSelfDefeat(t *testing.T) {
	ioJsonTest(t, "self-defeat.yml", "Self-defeat.json")
}

func TestJsonSherlock(t *testing.T) {
	ioJsonTest(t, "sherlock.yml", "Sherlock.json")
}

func TestJsonSnake(t *testing.T) {
	ioJsonTest(t, "snake.yml", "Snake.json")
}

func TestJsonSupportLoop(t *testing.T) {
	ioJsonTest(t, "support-loop.yml", "Support-loop.json")
}

func TestJsonTandem(t *testing.T) {
	ioJsonTest(t, "tandem.yml", "Tandem.json")
}

func TestJsonToulmin(t *testing.T) {
	ioJsonTest(t, "toulmin.yml", "Toulmin.json")
}

func TestJsonTrivial(t *testing.T) {
	ioJsonTest(t, "trivial.yml", "Trivial.json")
}

func TestJsonTweety(t *testing.T) {
	ioJsonTest(t, "tweety.yml", "Tweety.json")
}

func TestJsonUnreliableWitness(t *testing.T) {
	ioJsonTest(t, "unreliable-witness.yml", "Unreliable-witness.json")
}

func TestJsonVacation(t *testing.T) {
	ioJsonTest(t, "vacation.yml", "Vacation.json")
}
