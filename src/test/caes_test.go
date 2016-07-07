// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package test

import (
	// "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	// "log"
	"fmt"
	"os"
	"path"
	"testing"
)

const examples = "../../examples/AGs/YAML/"

func TestCAES(t *testing.T) {
	checkErr := func(e error) {
		if e != nil {
			t.Errorf(e.Error())
		}
	}
	d, err := os.Open(examples)
	checkErr(err)
	files, err := d.Readdir(0)
	checkErr(err)
	success := true
	for _, fi := range files {
		f, err := os.Open(examples + fi.Name())
		checkErr(err)
		if path.Ext(f.Name()) == ".yml" {
			// skip non-YAML files
			ag, err := yaml.Import(f)
			checkErr(err)
			err = ag.Infer()
			if err != nil {
				fmt.Printf("Infer error: %v\n", err)
				checkErr(err)
			}
			l := ag.GroundedLabelling()
			for id, expected := range ag.ExpectedLabeling {
				s, ok := ag.Statements[id]
				if !ok {
					success = false
					fmt.Printf("file: %v; expected a statement for %v\n",
						fi.Name(), id)
				} else {
					actual := l[s]
					if expected != actual {
						success = false
						fmt.Printf("file: %v; statement: %v; expected: %v; actual: %v\n",
							fi.Name(), id, expected, actual)
					}
				}
			}
		}
	}
	if success == false {
		t.Errorf("TestCAES failed\n")
	}
}
