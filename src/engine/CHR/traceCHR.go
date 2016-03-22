// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Constraint Handling Rules

package chr

import (
	"fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	// "math/big"
	// "strconv"
	// "strings"
)

var CHRtrace int

// ---------------
// trace functions
// ---------------

func pTraceHeadln(l, n int, s ...interface{}) {
	if CHRtrace >= l {
		for i := 0; i < n; i++ {
			fmt.Printf("      ")
		}
		fmt.Printf("*** ")
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
		fmt.Printf("\n")
	}
}

func pTraceHead(l, n int, s ...interface{}) {
	if CHRtrace >= l {
		for i := 0; i < n; i++ {
			fmt.Printf("      ")
		}
		fmt.Printf("*** ")
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
	}
}

func pTrace(l int, s ...interface{}) {
	if CHRtrace >= l {
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
	}
}

func pTraceln(l int, s ...interface{}) {
	if CHRtrace >= l {
		for _, s1 := range s {
			fmt.Printf("%v", s1)
		}
		fmt.Printf("\n")
	}
}

func pTraceEnv(l int, e Bindings) {
	if e == nil {
		pTrace(l, "nil")
	} else {
		if e.Var.Name == "" {
			pTrace(l, "[\"\"=nil]")
		} else {
			if e.Next == nil || e.Next.Var.Name == "" {
				pTrace(l, "[", e.Var.Name, "=", e.T.String(), ", nil]")
			} else {
				pTrace(l, "[", e.Var.Name, "=", e.T.String(), ",...]")
			}
		}

	}
}

func pTraceEMap(l int, n int, h *Compound) {
	if CHRtrace >= l {
		for i := 0; i < n; i++ {
			fmt.Printf("      ")
		}
		fmt.Printf("*** head: %s [ ", h.String())
		env := h.EMap
		for i, e := range *env {
			fmt.Printf("[ %d ] =", i)
			for _, e1 := range e {
				pTraceEnv(l, e1)
			}
			fmt.Printf(" || ")
		}
		fmt.Printf("\n")
	}
}

func printCHRStore() {
	first := true
	for _, aChr := range CHRstore {
		for _, con := range aChr.varArg {
			if !con.IsActive {
				if first {
					pTraceHead(1, 0, "CHR-Store: [", con.String())
					first = false
				} else {
					pTrace(1, ", ", con.String())
				}
			}
		}
	}
	if first {
		pTraceHeadln(1, 0, "CHR-Store: []")
	} else {
		pTraceln(1, "]")
	}

	first = true
	for _, aChr := range BuiltInStore {
		for _, con := range aChr.varArg {
			if !con.IsActive {
				if first {
					pTraceHead(1, 0, "Built-In Store: [", con.String())
					first = false
				} else {
					pTrace(1, ", ", con.String())
				}
			}
		}
	}
	if first {
		pTraceHeadln(1, 0, "Built-In Store: []")
	} else {
		pTraceln(1, "]")
	}

}

func chr2string() (str string) {
	first := true
	for _, aChr := range CHRstore {
		for _, con := range aChr.varArg {
			if !con.IsActive {
				if first {
					str = "[" + con.String()
					first = false
				} else {
					str = str + ", " + con.String()
				}
			}
		}
	}
	if first {
		str = "[]"
	} else {
		str = str + "]"
	}
	return
}

func bi2string() (str string) {
	first := true
	for _, aChr := range BuiltInStore {
		for _, con := range aChr.varArg {
			if !con.IsActive {
				if first {
					str = "[" + con.String()
					first = false
				} else {
					str = str + ", " + con.String()
				}
			}
		}
	}
	if first {
		str = "[]"
	} else {
		str = str + "]"
	}
	return
}
