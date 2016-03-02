// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

package chr

import (
	"fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	"testing"
)

func teval(t *testing.T, str1 string, result string) bool {

	term1, ok := ReadString(str1)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan str1 in test eval \"%s\" failed, term1: %s", str1, term1.String()))
		return false
	}
	term2, ok := ReadString(result)
	if !ok {
		t.Errorf(fmt.Sprintf("Scan result in test eval \"%s\" failed, term2: %s", result, term2.String()))
		return false
	}
	term3 := Eval(term1)
	if Equal(term3, term2) {
		fmt.Printf(" '%s' eval to:'%s' == '%s'\n", term1.String(), term3.String(), term2.String())
		return true
	}
	fmt.Printf(" '%s' eval to:'%s' NOT= '%s'\n", term1.String(), term3.String(), term2)

	return false

}

func TestEval01(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "+++9387", "9387")
	if ok != true {
		t.Errorf("TestEval01 failed\n")
	}
}

func TestEval02(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "-(-726)", "726")
	if ok != true {
		t.Errorf("TestEval02 failed\n")
	}
}

//func TestEval02a(t *testing.T) {
//	// check that a variable is not bound to two different terms
//	ok := teval(t, "---726)", "-726")
//	if ok != true {
//		t.Errorf("TestEval02a failed\n")
//	}
//}

func TestEval03(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "!!!true", "false")
	if ok != true {
		t.Errorf("TestEval03 failed\n")
	}
}
func TestEval04(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "^-1", "0")
	if ok != true {
		t.Errorf("TestEval04 failed\n")
	}
}
func TestEval05(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "¬¬¬true", "false")
	if ok != true {
		t.Errorf("TestEval05 failed\n")
	}
}
func TestEval06(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "7*8", "56")
	if ok != true {
		t.Errorf("TestEval06 failed\n")
	}
}
func TestEval07(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "5*8.2", "41.0")
	if ok != true {
		t.Errorf("TestEval07 failed\n")
	}
}
func TestEval08(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "7.1*8", "56.8")
	if ok != true {
		t.Errorf("TestEval08 failed\n")
	}
}

func TestEval09(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "21/7", "3")
	if ok != true {
		t.Errorf("TestEval09 failed\n")
	}
}
func TestEval10(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "21.0/7", "3.0")
	if ok != true {
		t.Errorf("TestEval10 failed\n")
	}
}
func TestEval11(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "25/7", "3")
	if ok != true {
		t.Errorf("TestEval11 failed\n")
	}
}
func TestEval12(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "25%7", "4")
	if ok != true {
		t.Errorf("TestEval12 failed\n")
	}
}

func TestEval13(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "25 div 7", "3")
	if ok != true {
		t.Errorf("TestEval13 failed\n")
	}
}
func TestEval14(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "25 mod 7", "4")
	if ok != true {
		t.Errorf("TestEval14 failed\n")
	}
}
func TestEval15(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "30 & 21", "20")
	if ok != true {
		t.Errorf("TestEval15 failed\n")
	}
}

func TestEval16(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "30 &^ 21", "10")
	if ok != true {
		t.Errorf("TestEval16 failed\n")
	}
}

func TestEval17(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "4 << 2", "16")
	if ok != true {
		t.Errorf("TestEval17 failed\n")
	}
}
func TestEval18(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "16 >> 2", "4")
	if ok != true {
		t.Errorf("TestEval18 failed\n")
	}
}

func TestEval19(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "3+8762", "8765")
	if ok != true {
		t.Errorf("TestEval19 failed\n")
	}
}
func TestEval20(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "3.0+8762", "8765.0")
	if ok != true {
		t.Errorf("TestEval20 failed\n")
	}
}

func TestEval21(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "3+87.62", "90.62")
	if ok != true {
		t.Errorf("TestEval21 failed\n")
	}
}

func TestEval22(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "3+87.62", "90.62")
	if ok != true {
		t.Errorf("TestEval22 failed\n")
	}
}

//func TestEval22a(t *testing.T) {
//	// check that a variable is not bound to two different terms
//	ok := teval(t, "\"Hallo\" + \" Welt\"", "\"Hallo Welt\"")
//	if ok != true {
//		t.Errorf("TestEval22a failed\n")
//	}
//}

func TestEval23(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "87-3", "84")
	if ok != true {
		t.Errorf("TestEval23 failed\n")
	}
}

func TestEval24(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "87.2-3", "84.2")
	if ok != true {
		t.Errorf("TestEval24 failed\n")
	}
}

func TestEval25(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "87.4-3.2", "84.2")
	if ok != true {
		t.Errorf("TestEval25 failed\n")
	}
}

func TestEval26(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "14 ^ 20", "26")
	if ok != true {
		t.Errorf("TestEval26 failed\n")
	}
}

func TestEval27(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "10 or 20", "30")
	if ok != true {
		t.Errorf("TestEval27 failed\n")
	}
}

func TestEval28(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "27.8 == 27.8", "true")
	if ok != true {
		t.Errorf("TestEval28 failed\n")
	}
}

func TestEval29(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "5*8 == 6*8-7", "false")
	if ok != true {
		t.Errorf("TestEval29 failed\n")
	}
}

func TestEval30(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "false == (7 > 9)", "true")
	if ok != true {
		t.Errorf("TestEval30 failed\n")
	}
}

func TestEval31(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "21 != 4*4", "true")
	if ok != true {
		t.Errorf("TestEval31 failed\n")
	}
}

func TestEval32(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "16 != 4*4", "false")
	if ok != true {
		t.Errorf("TestEval32 failed\n")
	}
}

func TestEval33(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "234 < 238", "true")
	if ok != true {
		t.Errorf("TestEval33 failed\n")
	}
}

func TestEval34(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "23.4 < 23.8", "true")
	if ok != true {
		t.Errorf("TestEval34 failed\n")
	}
}

func TestEval35(t *testing.T) {
	// check that a variable is not bound to two different terms
	ok := teval(t, "3>4 && 7<= 6 || 3<4 && 7>=6 ", "true")
	if ok != true {
		t.Errorf("TestEval35 failed\n")
	}
}
