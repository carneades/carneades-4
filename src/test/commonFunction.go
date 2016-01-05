package test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	//	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	"errors"
	//	"os"
	"testing"
)

const yamlDir = "../../examples/AGs/YAML/"
const yamlTmp = "/tmp/"
const jsonDir = "../../examples/AGs/AIF/"
const jsonTmp = "/tmp/"
const graphmlTmp = "/tmp/"

func check(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
		t.Skip("Skip Test")
	}
}

func checkLabeling(l caes.Labelling, stats map[string]*caes.Statement) error {
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
