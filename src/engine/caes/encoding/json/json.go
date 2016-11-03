// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import and export argument graphs in JSON format.
// Implemented using the YAML importer and exporter, since
// JSON files are also YAML files. This also makes
// it easier to maintain this package, since any modifications
// to the YAML translators are inherited automatically.

package json

import (
	"bytes"
	"io"

	"github.com/bronze1man/go-yaml2json"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

func Export(f io.Writer, ag *caes.ArgGraph) {
	var b bytes.Buffer
	yaml.Export(&b, ag)
	j, err := yaml2json.Convert(b.Bytes())
	if err != nil {
		return
	}
	io.Copy(f, bytes.NewReader(j))
}

// Import an argument graph represented in JSON or YAML.
func Import(inFile io.Reader) (*caes.ArgGraph, error) {
	return yaml.Import(inFile)
}
