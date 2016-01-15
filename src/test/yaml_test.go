package test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	"os"
	"path"
	"testing"
)

func TestYaml(t *testing.T) {
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
			fmt.Printf(" =  =  =  =  =  =  Import %s =  =  =  =  =  = \n", fi.Name())
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

			err = checkLabeling(l, ag.Statements, ag.ExpectedLabeling)
			if err != nil {
				fmt.Printf(" check labeling fail: %s \n", err.Error())
			}
			// check(t, err)
			//	fmt.Printf("---------- Write ArgGraph 2 Yaml: %s ----------\n", filename1)
			//	yaml.Export(os.Stdout, ag)
			//	fmt.Printf("---------- End: Write ArgGraph 2 Yaml: %s ----------\n", filename1)

			filename2 := "Tmp" + fi.Name()

			fmt.Printf(" -  -  -  -  -  -  Start Export %s \n", yamlTmp+filename2)
			f, err := os.Create(yamlTmp + filename2)
			check(t, err)

			// ag.ApplyLabelling(l)
			// fmt.Printf(" Export-Assumptions: %v \n", ag.Assumptions)
			yaml.Export(f, ag)
			f.Close()
			defer os.Remove(f.Name())
			fmt.Printf(" -  -  -  -  -  -  Start Import %s\n", yamlTmp+filename2)
			file, err = os.Open(yamlTmp + filename2)
			check(t, err)
			ag, err = yaml.Import(file)
			file.Close()
			// fmt.Printf(" Import-Assumptions: %v \n", ag.Assumptions)
			fmt.Printf(" -  -  -  -  -  -  End Import %s \n", yamlTmp+filename2)
			check(t, err)
			// fmt.Printf("---------- WriteArgGraph 02  %s ----------\n", filename2)
			// yaml.ExportWithReferences(os.Stdout, ag)
			// fmt.Printf("---------- End: WriteArgGraph 02 %s ----------\n", filename2)

			// fmt.Printf(" -  -  -  -  -  -  -  Start Import-GroundLabeling \n")
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
	}
}
