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

func checkLabeling(l caes.Labelling, stats map[string]*caes.Statement, expectedLbls map[string]caes.Label) error {
	errStr := ""
	for _, stat := range stats {
		calcLbl := l[stat]
		expLbl, found := expectedLbls[stat.Id]
		if found {
			if expLbl != calcLbl {
				if errStr == "" {
					errStr = fmt.Sprintf(" statement: %s, expected Label: %v, calculated Label: %v \n", stat.Id, expLbl, calcLbl)
				} else {
					errStr = errStr + fmt.Sprintf(" statement: %s, expected Label: %v, calculated Label: %v \n", stat.Id, expLbl, calcLbl)
				}
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
