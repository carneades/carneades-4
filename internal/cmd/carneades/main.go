// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
)

const name = "The Carneades Argumentation System"
const version = "4.1"
const source = "https://github.com/carneades/carneades-4"
const blog = "https://carneades.github.io/"

const help = `
Carneades is a tool for evaluating and visualizing argument graphs.

Usage: carneades command [arguments]

The commands are:

eval - evaluate a structured argument graph
dung - compute extensions of a Dung abstract argumentation framework
help - displays instructions

Execute "carneades help [command]" for further information.
`

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s\nversion: %s\nsource: %s\nblog: %s\nTry 'carneades help' for instructions.\n", name, version, source, blog)
	} else {
		switch os.Args[1] {
		case "eval":
			evalCmd()
		case "dung":
			dungCmd()
		case "server":
			webServerCmd()
		default:
			if len(os.Args) == 2 {
				fmt.Printf("%s\n", help)
			} else {
				switch os.Args[2] {
				case "eval":
					fmt.Printf("%s\n", helpEval)
				case "dung":
					fmt.Printf("%s\n", helpDung)
				default:
					fmt.Printf("%s\n", help)
				}
			}
		}
	}
}
