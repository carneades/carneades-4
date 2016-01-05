package test

import (
	// "fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	// "errors"
	"os"
	"path"
	"testing"
)

func TestGraphml(t *testing.T) {

	var ag *caes.ArgGraph
	var err error

	d, err := os.Open(yamlDir)
	check(t, err)
	files, err := d.Readdir(0)
	check(t, err)
	for _, fi := range files {
		file, err := os.Open(yamlDir + fi.Name())
		check(t, err)
		if path.Ext(file.Name()) == ".yml" {
			// skip non-YAML files
			ag, err = yaml.Import(file)
			file.Close()
			check(t, err)
			//	fmt.Printf("---------- WriteArgGraph %s ----------\n", filename1)
			//	yaml.ExportWithReferences(os.Stdout, ag)
			//	fmt.Printf("---------- End: WriteArgGraph %s ----------\n", filename1)
			l := ag.GroundedLabelling()
			ag.ApplyLabelling(l)
			//	fmt.Printf("---------- printLabeling %s ----------\n", filename1)
			//	printLabeling(l)
			//	fmt.Printf("---------- End: printLabeling %s ----------\n", filename1)
			file, err = os.Create(graphmlTmp + fi.Name())
			check(t, err)
			graphml.Export(file, ag)
			defer os.Remove(file.Name())
		}
	}
}
