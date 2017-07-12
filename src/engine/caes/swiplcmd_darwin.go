// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// SWI Prolog Command

package caes

import (
	"os"
	"os/exec"
	"syscall"
)

// resource limits for Prolog processes
const (
	stackLimit = "256m" // MB
)

// Create an SWI Prolog command for evaluating a given file.
// Handle SWI-Prolog errors.  Assure termination
// within given limits (time, stack size, ...)

func MakeSWIPrologCmd(f *os.File) *exec.Cmd {
	cmd := exec.Command("swipl", "-s ", f.Name(), "-L"+stackLimit)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}
