package main

import (
	"bytes"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	"github.com/gopherjs/gopherjs/js"
)

func GraphMLExportToString(ag caes.ArgGraph) (string, error) {
	var b bytes.Buffer
	var err = graphml.Export(&b, ag)
	return b.String(), err
}

func YamlImportFromString(inString string) (*caes.ArgGraph, error) {
	return yaml.Import(bytes.NewBufferString(inString))
}

func YamlExportToString(ag *caes.ArgGraph) string {
	var b bytes.Buffer
	yaml.Export(&b, ag)
	return b.String()
}

func main() {
	js.Global.Set("carneades", map[string]interface{}{
		"eval": evalJs,
	})
}

func evalJs(input string, inputFormat string, outputFormat string) map[string]interface{} {
	var ag *caes.ArgGraph
	var err error
	var str string

	switch inputFormat {
	case "yaml":
		ag, err = YamlImportFromString(input)
	default:
		return map[string]interface{}{
			"err": "invalid input format",
		}
	}

	if err != nil {
		return map[string]interface{}{
			"err": "error trying to import.",
		}
	}

	switch outputFormat {
	case "yaml":
		str = YamlExportToString(ag)
		err = nil
	case "graphml":
		str, err = GraphMLExportToString(*ag)
	default:
		return map[string]interface{}{
			"err": "invalid output format",
		}
	}
	if err != nil {
		return map[string]interface{}{
			"err": "error trying to export.",
		}
	} else {
		return map[string]interface{}{
			"result": str,
		}
	}
}
