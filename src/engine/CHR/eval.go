// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Constraint Handling Rules

package chr

import (
	// "fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	// "math/big"
	// "strconv"
	// "strings"
)

func Eval(t1 Term) Term {
	switch t1.Type() {
	case AtomType, BoolType, IntType, FloatType, StringType:
		return t1
	case CompoundType:

		args := []Term{}
		tArgs := []Type{}
		for _, a := range t1.(Compound).Args {
			a = Eval(a)
			args = append(args, a)
			tArgs = append(tArgs, a.Type())
		}
		t2 := t1.(Compound)
		t2.Args = args
		t1 = t2
		if t1.(Compound).Prio != 0 {
			an := len(args)
			switch an {
			case 1:
				return evalUnaryOperator(t1, args[0], tArgs[0])
			case 2:
				return evalBinaryOperator(t1, args[0], tArgs[0], args[1], tArgs[1])
			default:
				return evalN_aryOperator(t1, args, tArgs, an)
			}
		}
	}
	return t1
}

func evalUnaryOperator(t1, arg Term, typ Type) Term {
	switch t1.(Compound).Functor {
	case "+":
		return arg
	case "-":
		return evalUnaryMinus(t1, arg, typ)
	case "!", "¬":
		return evalNot(t1, arg, typ)
	case "^":
		return evalComp(t1, arg, typ)
	}
	return t1
}

func evalBinaryOperator(t1, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	switch t1.(Compound).Functor {
	case "*":
		return evalTimes(t1, a1, typ1, a2, typ2)
	case "/":
		return evalDivision(t1, a1, typ1, a2, typ2)
	case "div":
		return evalDiv(t1, a1, typ1, a2, typ2)
	case "%", "mod":
		return evalMod(t1, a1, typ1, a2, typ2)
	case "&":
		return evalBitAnd(t1, a1, typ1, a2, typ2)
	case "&^":
		return evalBitAndNot(t1, a1, typ1, a2, typ2)
	case "<<":
		return evalLeftShift(t1, a1, typ1, a2, typ2)
	case ">>":
		return evalRightShift(t1, a1, typ1, a2, typ2)
	case "+":
		return evalPlus(t1, a1, typ1, a2, typ2)
	case "-":
		return evalMinus(t1, a1, typ1, a2, typ2)
	case "^":
		return evalBitXOr(t1, a1, typ1, a2, typ2)
	case "or":
		return evalBitOr(t1, a1, typ1, a2, typ2)
	case "==":
		return evalEq(t1, a1, typ1, a2, typ2)
	case "!=":
		return evalNotEq(t1, a1, typ1, a2, typ2)
	case "<":
		return evalLess(t1, a1, typ1, a2, typ2)
	case "<=", "=<":
		return evalLessEq(t1, a1, typ1, a2, typ2)
	case ">":
		return evalGt(t1, a1, typ1, a2, typ2)
	case ">=":
		return evalGtEq(t1, a1, typ1, a2, typ2)
	case "&&":
		return evalLogAnd(t1, a1, typ1, a2, typ2)
	case "||":
		return evalLogOr(t1, a1, typ1, a2, typ2)
	}
	return t1
}

func evalN_aryOperator(t1 Term, args []Term, typs []Type, n int) Term {
	return t1
}

func evalUnaryMinus(t1 Term, a1 Term, typ1 Type) Term {
	// -a1
	switch typ1 {
	case IntType:
		return -a1.(Int)
	case FloatType:
		return -a1.(Float)
	}
	return t1
}

func evalNot(t1 Term, a1 Term, typ1 Type) Term {
	// !a1 or ¬a1
	if typ1 == BoolType {
		return !a1.(Bool)
	}
	if typ1 == CompoundType {
		a := a1.(Compound)
		n := len(a.Args)
		if n == 1 {
			switch a.Functor {
			case "!", "¬":
				return a.Args[0]
			}
			return t1
		}
		if n == 2 {
			c := Compound{}
			newc := false
			args := a.Args
			arg1 := args[0]
			arg2 := args[1]
			switch a.Functor {
			case "<":
				c = Compound{Functor: "<=", Args: []Term{arg2, arg1}}
				newc = true
			case "<=":
				c = Compound{Functor: "<", Args: []Term{arg2, arg1}}
				newc = true
			case "==":
				c = Eval(Term(Compound{Functor: "!=", Prio: 3, Args: args})).(Compound)
				newc = true
			case "!=":
				c = Eval(Term(Compound{Functor: "==", Prio: 3, Args: args})).(Compound)
				newc = true
			}
			if newc {
				c.Id = a.Id
				c.Prio = a.Prio
				c.IsActive = a.IsActive
				return c
			}
		}
	}
	return t1
}

func evalComp(t1 Term, a1 Term, typ1 Type) Term {
	// ^a1
	if typ1 == IntType {
		return ^a1.(Int)
	}
	return t1
}

func evalTimes(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 * a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return a1.(Int) * a2.(Int)
		case FloatType:
			return Float(float64(a1.(Int)) * float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Float(float64(a1.(Float)) * float64(a2.(Int)))
		case FloatType:
			return a1.(Float) * a2.(Float)
		default:
			return t1
		}
	}
	return t1
}

func evalDivision(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// if a2 != 0 { a1 / a2 }
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			if a2.(Int) != 0 {
				return a1.(Int) / a2.(Int)
			}
		case FloatType:
			if a2.(Float) != 0.0 {
				return Float(float64(a1.(Int)) / float64(a2.(Float)))
			}
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			if a2.(Int) != 0 {
				return Float(float64(a1.(Float)) / float64(a2.(Int)))
			}
		case FloatType:
			if a2.(Float) != 0.0 {
				return a1.(Float) / a2.(Float)
			}
		default:
			return t1
		}
	}
	return t1
}

func evalDiv(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 / a2 for integer
	if typ1 != IntType || typ2 != IntType || a2.(Int) == 0 {
		return t1
	}
	return a1.(Int) / a2.(Int)
}

func evalMod(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 % a2 for integer
	if typ1 != IntType || typ2 != IntType || a2.(Int) == 0 {
		return t1
	}
	return a1.(Int) % a2.(Int)
}

func evalBitAnd(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 & a2 for integer
	if typ1 != IntType || typ2 != IntType {
		return t1
	}
	return a1.(Int) & a2.(Int)
}

func evalBitAndNot(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 &^ a2 for integer
	if typ1 != IntType || typ2 != IntType {
		return t1
	}
	return a1.(Int) &^ a2.(Int)
}

func evalLeftShift(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 << a2 for integer
	if typ1 != IntType || typ2 != IntType {
		return t1
	}
	return Int(uint(a1.(Int)) << uint(a2.(Int)))
}

func evalRightShift(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 >> a2 for integer
	if typ1 != IntType || typ2 != IntType {
		return t1
	}
	return Int(uint(a1.(Int)) >> uint(a2.(Int)))
}

func evalPlus(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return a1.(Int) + a2.(Int)
		case FloatType:
			return Float(float64(a1.(Int)) + float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Float(float64(a1.(Float)) + float64(a2.(Int)))
		case FloatType:
			return a1.(Float) + a2.(Float)
		default:
			return t1
		}
	case StringType:
		if typ2 == StringType {
			return a1.(String) + a2.(String)
		}
	}
	return t1
}

func evalMinus(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return a1.(Int) - a2.(Int)
		case FloatType:
			return Float(float64(a1.(Int)) - float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Float(float64(a1.(Float)) - float64(a2.(Int)))
		case FloatType:
			return a1.(Float) - a2.(Float)
		default:
			return t1
		}
	}
	return t1
}

func evalBitOr(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 or a2 == a1 | a2 for integer
	if typ1 != IntType || typ2 != IntType {
		return t1
	}
	return a1.(Int) | a2.(Int)
}

func evalBitXOr(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 ^ a2 for integer
	if typ1 != IntType || typ2 != IntType {
		return t1
	}
	return Int(uint(a1.(Int)) ^ uint(a2.(Int)))
}

func evalEq(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 == a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return Bool(a1.(Int) == a2.(Int))
		case FloatType:
			return Bool(float64(a1.(Int)) == float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Bool(float64(a1.(Float)) == float64(a2.(Int)))
		case FloatType:
			return Bool(a1.(Float) == a2.(Float))
		default:
			return t1
		}
	case StringType:
		if typ2 == StringType {
			return Bool(a1.(String) == a2.(String))
		}
	case BoolType:
		if typ2 == BoolType {
			return Bool(a1.(Bool) == a2.(Bool))
		}
	default:
		if Equal(a1, a2) {
			return Bool(true)
		}
	}
	return t1
}

func evalNotEq(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 != a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return Bool(a1.(Int) != a2.(Int))
		case FloatType:
			return Bool(float64(a1.(Int)) != float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Bool(float64(a1.(Float)) != float64(a2.(Int)))
		case FloatType:
			return Bool(a1.(Float) != a2.(Float))
		default:
			return t1
		}
	case StringType:
		if typ2 == StringType {
			return Bool(a1.(String) != a2.(String))
		}
	case BoolType:
		if typ2 == BoolType {
			return Bool(a1.(Bool) != a2.(Bool))
		}
	default:
		if Equal(a1, a2) {
			return Bool(false)
		}
	}
	return t1
}

func evalLess(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 < a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return Bool(a1.(Int) < a2.(Int))
		case FloatType:
			return Bool(float64(a1.(Int)) < float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Bool(float64(a1.(Float)) < float64(a2.(Int)))
		case FloatType:
			return Bool(a1.(Float) < a2.(Float))
		default:
			return t1
		}
	case StringType:
		if typ2 == StringType {
			return Bool(a1.(String) < a2.(String))
		}
	}
	return t1
}

func evalLessEq(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 <= a2 or a1 =< a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return Bool(a1.(Int) <= a2.(Int))
		case FloatType:
			return Bool(float64(a1.(Int)) <= float64(a2.(Float)))
		default:
			return t1
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Bool(float64(a1.(Float)) <= float64(a2.(Int)))
		case FloatType:
			return Bool(a1.(Float) <= a2.(Float))
		default:
			return t1
		}
	case StringType:
		if typ2 == StringType {
			return Bool(a1.(String) <= a2.(String))
		}
	}
	return t1
}

func evalGt(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 > a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return Bool(a1.(Int) > a2.(Int))
		case FloatType:
			return Bool(float64(a1.(Int)) > float64(a2.(Float)))
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Bool(float64(a1.(Float)) > float64(a2.(Int)))
		case FloatType:
			return Bool(a1.(Float) > a2.(Float))
		}
	case StringType:
		if typ2 == StringType {
			return Bool(a1.(String) > a2.(String))
		}
	}
	t := t1.(Compound)
	c := Compound{Functor: "<", Id: t.Id, Prio: t.Prio, IsActive: t.IsActive, Args: []Term{a2, a1}}
	return c
}

func evalGtEq(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 >= a2
	switch typ1 {
	case IntType:
		switch typ2 {
		case IntType:
			return Bool(a1.(Int) >= a2.(Int))
		case FloatType:
			return Bool(float64(a1.(Int)) >= float64(a2.(Float)))
		}
	case FloatType:
		switch typ2 {
		case IntType:
			return Bool(float64(a1.(Float)) >= float64(a2.(Int)))
		case FloatType:
			return Bool(a1.(Float) >= a2.(Float))
		}
	case StringType:
		if typ2 == StringType {
			return Bool(a1.(String) >= a2.(String))
		}
	}
	t := t1.(Compound)
	c := Compound{Functor: "<=", Id: t.Id, Prio: t.Prio, IsActive: t.IsActive, Args: []Term{a2, a1}}
	return c
}

func evalLogAnd(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 && a2
	if typ1 == BoolType && typ2 == BoolType {
		return a1.(Bool) && a2.(Bool)
	}
	if typ1 == BoolType {
		if a1.(Bool) {
			return a2
		} else {
			return Bool(false)
		}
	}
	if typ2 == BoolType {
		if a2.(Bool) {
			return a1
		} else {
			return Bool(false)
		}
	}
	if Equal(a1, Eval(Compound{Functor: "!", Prio: 6, Args: []Term{a2}})) {
		return Bool(false)
	}
	return t1
}

func evalLogOr(t1 Term, a1 Term, typ1 Type, a2 Term, typ2 Type) Term {
	// a1 || a2
	if typ1 == BoolType && typ2 == BoolType {
		return a1.(Bool) || a2.(Bool)
	}
	if typ1 == BoolType {
		if a1.(Bool) {
			return a1
		} else {
			return a2
		}
	}
	if typ2 == BoolType {
		if a2.(Bool) {
			return a2
		} else {
			return a1
		}
	}
	a3 := Eval(Compound{Functor: "!", Prio: 6, Args: []Term{a2}})
	// fmt.Printf(" A1: %s, A2: %s, Eval !A2: %s Equal(A1,!A2) %v\n", a1, a2, a3, Equal(a1, a3))
	if Equal(a1, a3) {
		return Bool(true)
	}
	return t1
}
