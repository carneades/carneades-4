package test

import (
	// "fmt"
	// "github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/aif"
	//	"log"
	// "errors"
	"os"
	"testing"
)

func TestAIF1(t *testing.T) {
	// var ag *caes.ArgGraph
	file, _ := os.Open("../../examples/AIF/nodeset7.json")
	_, err := aif.Import(file)
	if err != nil {
		t.Errorf("AIF import failed.\n")
	}
}
