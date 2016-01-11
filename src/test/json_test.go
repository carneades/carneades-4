package test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	cjson "github.com/carneades/carneades-4/src/engine/caes/encoding/json"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	"os"
	"path"
	"strings"
	"testing"
)

func TestJson(t *testing.T) {
	var ag *caes.ArgGraph
	var err error

	d, err := os.Open(yamlDir)
	check(t, err)
	files, err := d.Readdir(0)
	check(t, err)
	for _, fi := range files {
		// YML-IMPORT
		// ==========
		file, err := os.Open(yamlDir + fi.Name())
		check(t, err)
		if path.Ext(file.Name()) == ".yml" {
			// skip non-yml files
			fmt.Printf(" =  =  =  =  =  =  Import %s =  =  =  =  =  = \n", fi.Name())

			ag, err = yaml.Import(file)

			check(t, err)
			// YAML-Export with reference
			// ==========================

			file0, err := os.Create(yamlTmp + "ref" + fi.Name())
			check(t, err)
			yaml.ExportWithReferences(file0, ag)
			defer os.Remove(file0.Name())

			l := ag.GroundedLabelling()
			// -----------------------
			//	fmt.Printf(" ## Labeling json-Datei: \n")
			//	printLabeling(l2)
			err = checkLabeling(l, ag.Statements, ag.ExpectedLabeling)
			if err != nil {
				fmt.Printf(" yaml-Import: %v, Labeling fail %v \n", fi.Name(), err)
			}
			// check(t, err)

			// JSON-Export in Temp-Dir with the file-name: <name>.json
			// =========== -------------------------------------------

			jsonFilePaht := jsonTmp + strings.Replace(fi.Name(), ".yml", ".json", 1)
			file, err = os.Create(jsonFilePaht)
			check(t, err)
			cjson.Export(file, ag)
			file.Close()

			// JSON-Import
			// ===========
			fmt.Printf(" =  =  =  =  =  =  Import %s =  =  =  =  =  = \n", strings.Replace(fi.Name(), ".yml", ".json", 1))
			file2, err := os.Open(jsonFilePaht)
			check(t, err)
			ag2, err := cjson.Import(file2)
			check(t, err)
			defer os.Remove(file2.Name())
			// JSON_v02-Export
			// ===============
			fmt.Printf(" =  =  =  =  =  =  Export %s =  =  =  =  =  = \n", strings.Replace(fi.Name(), ".yml", "_v-02.json", 1))
			jsonFilePaht_v02 := jsonTmp + strings.Replace(fi.Name(), ".yml", "_v-02.json", 1)
			file02, err := os.Create(jsonFilePaht_v02)
			check(t, err)
			cjson.Export(file02, ag2)
			file.Close()
			defer os.Remove(file.Name())
			// YML-Export in Temp-Dir
			// ==========
			file3, err := os.Create(yamlTmp + fi.Name())
			check(t, err)
			yaml.ExportWithReferences(file3, ag2)
			defer os.Remove(file3.Name())
			// Check Labeling JSON
			// ===================
			l2 := ag2.GroundedLabelling()
			// -----------------------
			//	fmt.Printf(" ## Labeling json-Datei: \n")
			//	printLabeling(l2)
			err = checkLabeling(l2, ag2.Statements, ag.ExpectedLabeling)
			if err != nil {
				fmt.Printf(" json-Import: %v, Labeling fail %v \n", strings.Replace(fi.Name(), ".yml", ".json", 1), err)
			}
			// check(t, err)

		}
	}
}
