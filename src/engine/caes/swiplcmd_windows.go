// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// SWI Prolog Command

package caes

import (
	"log"
	"os"
	"os/exec"
)

// resource limits for Prolog processes
const (
	stackLimit = "256m" // MB
)

// Create an SWI Prolog command for evaluating a given file.
// Handle SWI-Prolog errors.  Assure termination
// within given limits (time, stack size, ...)

func MakeSWIPrologCmd(f *os.File) *exec.Cmd {
	swipl := ""

	if _, err := os.Stat("C:\\Program Files\\swipl\\bin\\swipl.exe"); !os.IsNotExist(err) {
		swipl = "C:\\Program Files\\swipl\\bin\\swipl.exe"
	}

	if _, err := os.Stat("C:\\Program Files (x86)\\swipl\\bin\\swipl.exe"); !os.IsNotExist(err) {
		swipl = "C:\\Program Files (x86)\\swipl\\bin\\swipl.exe"
	}

	log.Printf("PROLOG-Datei: %s", f.Name())
	cmd := exec.Command(swipl, "-s ", f.Name(), "-L"+stackLimit)
	return cmd
}
