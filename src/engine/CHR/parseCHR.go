// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// parse Constraint Handling Rules

package chr

import (
	// "errors"
	"fmt"
	. "github.com/carneades/carneades-4/src/engine/terms"
	"os"
	sc "text/scanner"
	// "math/big"
	// "strconv"
	"strings"
)

type parseType int

const (
	ParseCHR parseType = iota
	ParseBI
	ParseGoal     // CHR and Built-In
	ParseRuleGoal // Chr, Built-In and Variable
)

type cList []*Compound

func CHRerr(format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, format, a)
}

func toClist(l Term) (cList, bool) {
	cl := cList{}
	if l.Type() != ListType {
		return cl, false
	}
	for _, t1 := range l.(List) {
		if t1.Type() != CompoundType {
			return cl, false
		}
		t2 := t1.(Compound)
		t2.EMap = &EnvMap{}
		cl = append(cl, &t2)
	}
	return cl, true
}

// parse CHR-rules and goals from string src
// CHR-rules:
//
// [<rulename>] '@' <keep-heads> '==>' [<guards> '|'] <body> '.'
// [<rulename>] '@' <keep-heads> '/' <del-heads> '<=>' [<guards> '|'] <body>'.'
// [<rulename>] '@' <del-heads> '<=>' [<guards> '|'] <body>'.'
//
// goals
// <predicates> '.'
func ParseStringCHRRulesGoals(src string) (ok bool) {
	// src is the input that we want to tokenize.
	// var s sc.Scanner
	var s sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))

	s.Error = Err

	ok = parseRules(&s)
	return
}

func parseRules(s *sc.Scanner) (ok bool) {
	var t Term
	InitStore()

	nameNr := 1
	tok := s.Scan()

	pTraceHeadln(4, 4, " parse rule tok: ", Tok2str(tok))
	if tok == sc.EOF {
		Err(s, " Empty input")
		return false
	}

	for tok != sc.EOF {
		tok1 := s.Peek()
		pTraceHeadln(1, 4, " in loop parse rule tok: ", Tok2str(tok), ", tok1: [", Tok2str(tok1), "]")
		switch tok {
		case sc.Ident:
			t, tok, ok = Factor_name(s.TokenText(), s, s.Scan())
			if !ok {
				return ok
			}
			if tok == '@' {
				tok, ok = parseKeepHead(s, s.Scan(), t.String())
			} else {
				tok, ok = parseKeepHead1(s, tok, fmt.Sprintf("(%d)", nameNr), t)
				nameNr++
			}
		default:
			Err(s, fmt.Sprintf("Missing a rule-name or a predicate-name at the beginning (not \"%v\")", Tok2str(tok)))
			return false
		}

	}
	return true
}

func parseKeepHead(s *sc.Scanner, tok rune, name string) (rune, bool) {

	// ParseGoal - it is not clear, a goal-list or a head-list
	pTraceHeadln(4, 4, " parse Keep Head:", name, " tok: ", Tok2str(tok))

	if tok != sc.Ident {
		Err(s, fmt.Sprintf("Missing predicate-name in rule: %s (not \"%v\")", name, Tok2str(tok)))
		return tok, false
	}
	t, tok, ok := Factor_name(s.TokenText(), s, s.Scan())
	if !ok {
		return tok, ok
	}

	return parseKeepHead1(s, tok, name, t)
}

func parseKeepHead1(s *sc.Scanner, tok rune, name string, t Term) (tok1 rune, ok bool) {

	if t.Type() != CompoundType {
		Err(s, fmt.Sprintf("Missing a predicate in rule %s (not %s)", name, t.String()))
		return tok, false
	}
	keepList := List{t}

	for tok == ',' {
		tok = s.Scan()
		if tok != sc.Ident {
			Err(s, fmt.Sprintf("Missing predicate-name in rule %s (not \"%v\")", name, Tok2str(tok)))
			return tok, false
		}
		t, tok, ok = Factor_name(s.TokenText(), s, s.Scan())
		if !ok {
			return tok, ok
		}
		keepList = append(keepList, t)
	}

	if tok == '.' {
		// Goals-List
		cGoalList, ok := prove2Clist(ParseGoal, name, keepList)
		if !ok {
			return tok, false
		}
		for _, g := range cGoalList {
			addRefConstraintToStore(g)
		}
		return s.Scan(), true
	}

	// keep- or del-head
	cKeepList, ok := prove2Clist(ParseCHR, name, keepList)
	if !ok {
		return tok, false
	}
	var delList Term
	switch tok {
	case '\\', '|':
		delList, tok, ok = parseDelHead(s, s.Scan())

		cDelList, ok := prove2Clist(ParseCHR, name, delList)
		if !ok {
			return tok, false
		}
		if tok != '<' {
			Err(s, fmt.Sprintf(" '<' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		tok = s.Scan()
		if tok != '=' {
			Err(s, fmt.Sprintf(" '=' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		tok = s.Scan()
		if tok != '>' {
			Err(s, fmt.Sprintf(" '>' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		return parseGuardHead(s, s.Scan(), name, cKeepList, cDelList)

	case '<':
		tok = s.Scan()
		if tok != '=' {
			Err(s, fmt.Sprintf(" '=' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		tok = s.Scan()
		if tok != '>' {
			Err(s, fmt.Sprintf(" '>' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		// the scaned keep-list is the del-list
		return parseGuardHead(s, s.Scan(), name, nil, cKeepList)
	case '=':
		tok = s.Scan()
		if tok != '=' {
			Err(s, fmt.Sprintf(" '=' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		tok = s.Scan()
		if tok != '>' {
			Err(s, fmt.Sprintf(" '>' in '<=>' excpected, not: %s", Tok2str(tok)))
			return tok, false
		}
		return parseGuardHead(s, s.Scan(), name, cKeepList, nil)
	default:
		Err(s, fmt.Sprintf(" unexcpected token: %s, excpect in head-rule '\\', '<=>' or '==>'", Tok2str(tok)))
	}

	return tok, false
}

func parseDelHead(s *sc.Scanner, tok rune) (delList List, tok1 rune, ok bool) {
	var t Term
	pTraceHeadln(1, 4, "parse Del-Head tok[", Tok2str(tok))
	delList = List{}
	if tok != sc.Ident {
		Err(s, fmt.Sprintf("Missing predicate-name (not \"%v\")", Tok2str(tok)))
		return delList, tok, false
	}
	t, tok1, ok = Factor_name(s.TokenText(), s, s.Scan())
	if !ok {
		return
	}
	delList = append(delList, t)
	for tok1 == ',' {
		tok = s.Scan()
		if tok != sc.Ident {
			Err(s, fmt.Sprintf("Missing predicate-name (not \"%v\")", Tok2str(tok)))
			return delList, tok, false
		}
		t, tok1, ok = Factor_name(s.TokenText(), s, s.Scan())
		if !ok {
			return
		}
		delList = append(delList, t)
	}
	return
}

func parseGuardHead(s *sc.Scanner, tok rune, name string, cKeepList, cDelList cList) (tok1 rune, ok bool) {
	// ParseRuleGoal - it is no clear, if it a guard or body
	bodyList, tok, ok := parseConstraints1(ParseRuleGoal, s, tok)
	cGuardList := cList{}
	if tok == '.' {
		tok1 = s.Scan()

	}
	if tok == '|' {
		cGuardList, ok = prove2Clist(ParseBI, name, bodyList)
		tok = s.Scan()
		bodyList, tok, ok = parseConstraints1(ParseRuleGoal, s, tok)
		if tok != '.' {
			Err(s, fmt.Sprintf(" After rule %s a '.' exspected, not a %s", name, Tok2str(tok)))
		}
		tok = s.Scan()

	}

	CHRruleStore = append(CHRruleStore, &chrRule{name: name, id: nextRuleId,
		delHead:  cDelList,
		keepHead: cKeepList,
		guard:    cGuardList,
		body:     bodyList.(List)})
	nextRuleId++

	return tok, true
}

/*	t, tok, ok := assignexpr(s, tok)
	if trace {
		fmt.Printf("--> factor : '%s'\n", Tok2str(tok))
	}
	tok1 := tok
	ok = true
	switch tok1 {

	case '(':
		pos := s.Pos()
		t, tok, ok = expression(s, s.Scan())
		if trace {
			fmt.Printf("<-- expression in ( factor: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
		if !ok {
			return
		}
		if tok != ')' {
			err(s, fmt.Sprintf("missing closed ')' for the open '(' at position %s", pos))
			return false
		}
		tok = s.Scan()
	case sc.Ident:
		t, tok, ok = factor_name(s.TokenText(), s, s.Scan())
		if trace {
			fmt.Printf("<-- factor_name: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	case sc.Int:
		t, tok, ok = sInt(s)
		if trace {
			fmt.Printf("<-- sInt: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	case sc.Float:
		t, tok, ok = sFloat(s)
		if trace {
			fmt.Printf("<-- sFloat: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	case sc.Char:
		// t, tok, ok = sChar(s)
		t, tok, ok = factor_name(s.TokenText(), s, s.Scan())
		if trace {
			fmt.Printf("<-- sChar: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	case sc.String:
		t, tok, ok = sString(s)
		if trace {
			fmt.Printf("<-- sString: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	case sc.RawString:
		t, tok, ok = sString(s)
		if trace {
			fmt.Printf("<-- sRawString: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	// case sc.Comment:
	case sc.EOF:
		err(s, "EOF missing term")
		return false
	default:
		err(s, fmt.Sprintf("unexpected character '%c', expect <variable>, <name>, <constant>, '(' or '['", tok))
		t = Atom("nil")
		return false
	}
	return true
}
*/

func addChrRule(name string, keepList, delList, guardList, bodyList Term) bool {

	cKeepList, ok := prove2Clist(ParseCHR, name, keepList)
	if !ok {
		return ok
	}
	//		return errors.New(fmt.SprintTok2str("Convert Keep-Head in rule %s failed: %s\n", name, keepList))

	//	if delList.Type() != ListType {
	//		return errors.New(fmt.Sprintf("DEl-Head in rule %s must be a List, not:  %s\n", name, delList))
	//	}
	cDelList, ok := prove2Clist(ParseCHR, name, delList)
	if !ok {
		return ok
	}
	// 		return errors.New(fmt.Sprintf("Convert DEl-Head in rule %s failed: %s\n", name, delList))

	//	if guardList.Type() != ListType {
	//		return errors.New(fmt.Sprintf("GUARD in rule %s must be a List, not:  %s (%v)\n", name, guardList, ok))
	//	}
	cGuardList, ok := prove2Clist(ParseBI, name, guardList)
	if !ok {
		return ok
	}
	//		return errors.New(fmt.Sprintf("Convert GUARD in rule %s failed: %s\n", name, guardList))

	// bodyList, err = prove2BodyList(bodyList)

	//	if bodyList.Type() != ListType {
	//		return errors.New(fmt.Sprintf("BODY in rule %s must be a List, not:  %s\n", name, bodyList))
	//	}

	CHRruleStore = append(CHRruleStore, &chrRule{name: name, id: nextRuleId,
		delHead:  cDelList,
		keepHead: cKeepList,
		guard:    cGuardList,
		body:     bodyList.(List)})
	nextRuleId++
	return true
}

func addStringGoals(goals string) bool {
	goalList, ok := ParseGoalString(goals)
	if !ok || goalList.Type() != ListType {
		CHRerr("Scan GOAL-List failed: %s\n", goalList)
		return false
	}
	for _, g := range goalList.(List) {
		if g.Type() == CompoundType {
			addConstraintToStore(g.(Compound))
		} else {
			CHRerr(" GOAL is not a predicate: %s\n", g)
			return false
		}

	}
	return true
}

func ParseCHRString(src string) (result Term, ok bool) {
	// src is the input that we want to tokenize.
	var s sc.Scanner
	// var s *sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))
	s.Error = Err

	result, _, ok = parseConstraints(ParseCHR, &s)
	return
}

func ParseBIString(src string) (result Term, ok bool) {
	// src is the input that we want to tokenize.
	var s sc.Scanner
	// var s *sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))
	s.Error = Err

	result, _, ok = parseConstraints(ParseBI, &s)
	return
}

func ParseGoalString(src string) (result Term, ok bool) {
	// src is the input that we want to tokenize.
	var s sc.Scanner
	// var s *sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))
	s.Error = Err

	result, _, ok = parseConstraints(ParseGoal, &s)
	return
}

func ParseRuleGoalString(src string) (result Term, ok bool) {
	// src is the input that we want to tokenize.
	var s sc.Scanner
	// var s *sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))
	s.Error = Err

	result, _, ok = parseConstraints(ParseRuleGoal, &s)
	return
}

func prove2Clist(ty parseType, name string, t Term) (cl cList, ok bool) {
	// ty == ParseCHR, ParseBI, ParseGoal-CHR and Built-In,
	// no: ParseRuleGoal-Chr, Built-In and Variable
	cl = cList{}
	switch t.Type() {
	case AtomType:
		switch ty {
		case ParseCHR:
			CHRerr(" unexpected atom ", t, " in head of rule ", name)
			return cl, false
		case ParseBI:
			CHRerr(" unexpected atom ", t, " in guard of rule ", name)
			return cl, false
		case ParseGoal:
			CHRerr(" unexpected atom ", t, " in goal-list ")
			return cl, false
		}
	case BoolType:
		switch ty {
		case ParseCHR:
			CHRerr(" unexpected boolean ", t, " in head of rule ", name)
			return cl, false
		case ParseBI:
			CHRerr(" unexpected boolean ", t, " in guard of rule ", name)
			return cl, false
		case ParseGoal:
			CHRerr(" unexpected boolean ", t, " in goal-list ")
			return cl, false
		}
	case IntType:
		switch ty {
		case ParseCHR:
			CHRerr(" unexpected integer ", t, " in head of rule ", name)
			return cl, false
		case ParseBI:
			CHRerr(" unexpected integer ", t, " in guard of rule ", name)
			return cl, false
		case ParseGoal:
			CHRerr(" unexpected integer ", t, " in goal-list ")
			return cl, false
		}
	case FloatType:
		switch ty {
		case ParseCHR:
			CHRerr(" unexpected float-number ", t, " in head of rule ", name)
			return cl, false
		case ParseBI:
			CHRerr(" unexpected float-number ", t, " in guard of rule ", name)
			return cl, false
		case ParseGoal:
			CHRerr(" unexpected float-number ", t, " in goal-list ")
			return cl, false
		}
	case StringType:
		switch ty {
		case ParseCHR:
			CHRerr(" unexpected string ", t, " in head of rule ", name)
			return cl, false
		case ParseBI:
			CHRerr(" unexpected string ", t, " in guard of rule ", name)
			return cl, false
		case ParseGoal:
			CHRerr(" unexpected string ", t, " in goal-list ")
			return cl, false
		}
	case CompoundType:
		comp := t.(Compound)
		switch ty {
		case ParseCHR: // CHR, no Build-In
			if comp.Prio != 0 {
				CHRerr(" unexpected Build-In predicate ", t, " in head of rule ", name)
				return cl, false
			}
			comp.EMap = &EnvMap{}
			cl = append(cl, &comp)
			return cl, true
		case ParseBI: // only Build-In
			if comp.Prio == 0 {
				CHRerr(" unexpected CHR predicate ", t, " in guard of rule ", name)
				return cl, false
			}
			cl = append(cl, &comp)
			return cl, true
		case ParseGoal: // both
			cl = append(cl, &comp)
			return cl, true
		}
	case ListType:

		for _, t1 := range t.(List) {
			if t1.Type() != CompoundType {
				return prove2Clist(ty, name, t1)
			}
			t2 := t1.(Compound)
			if ty == ParseCHR {
				t2.EMap = &EnvMap{}
			}
			cl = append(cl, &t2)
		}
		return cl, true

	case VariableType:
		switch ty {
		case ParseCHR:
			CHRerr(" unexpected variable ", t, " in head of rule ", name)
			return cl, false
		case ParseBI:
			CHRerr(" unexpected variable ", t, " in guard of rule ", name)
			return cl, false
		case ParseGoal:
			CHRerr(" unexpected variable ", t, " in goal-list ")
			return cl, false
		}
	}
	return nil, false
}

func parseConstraints(ty parseType, s *sc.Scanner) (t Term, tok rune, ok bool) {
	pTraceHeadln(3, 4, " parse constraints ")
	return parseConstraints1(ty, s, s.Scan())
}

func parseConstraints1(ty parseType, s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	pTraceHeadln(3, 4, " parse constraints ", Tok2str(tok1))
	tok = tok1
	if tok == sc.EOF {
		return List{}, tok, true
	}

	t, tok, ok = Assignexpr(s, tok)
	if !ok {
		return
	}
	switch ty {
	case ParseCHR:
		if t.Type() != CompoundType || t.(Compound).Prio != 0 {
			Err(s, fmt.Sprintf(" Not a CHR-predicate: %s ", t))
		}
	case ParseBI:
		if t.Type() != CompoundType || t.(Compound).Prio == 0 {
			Err(s, fmt.Sprintf(" Not a Built-in constraint: %s ", t))
		}
	case ParseGoal:
		if t.Type() != CompoundType {
			Err(s, fmt.Sprintf(" Not a CHR-predicate, a predicate or a build-in function: %s ", t))
		}
	case ParseRuleGoal:
		if t.Type() != CompoundType && t.Type() != VariableType && t.Type() != BoolType {
			Err(s, fmt.Sprintf(" Not a CHR-predicate, a predicatea, a build-in function or variable: %s ", t))
		}
	}

	pTraceHeadln(4, 4, "<-- assing-expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)

	if tok == ',' {
		t1 := List{t}
		for tok == ',' {
			t, tok, ok = Assignexpr(s, s.Scan())
			if !ok {
				return t1, tok, false
			}
			switch ty {
			case ParseCHR:
				if t.Type() != CompoundType || t.(Compound).Prio != 0 {
					Err(s, fmt.Sprintf(" Not a CHR-predicate: %s ", t))
				}
			case ParseBI:
				if t.Type() != CompoundType || t.(Compound).Prio == 0 {
					Err(s, fmt.Sprintf(" Not a Built-in constraint: %s ", t))
				}
			case ParseGoal:
				if t.Type() != CompoundType {
					Err(s, fmt.Sprintf(" Not a CHR-predicate, a predicate or a build-in function: %s ", t))
				}
			case ParseRuleGoal:
				if t.Type() != CompoundType && t.Type() != VariableType && t.Type() != BoolType {
					Err(s, fmt.Sprintf(" Not a CHR-predicate, a predicatea, a build-in function or variable: %s ", t))
				}
			}

			pTraceHeadln(4, 4, "<-- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)

			t1 = append(t1, t)
		}
		t = t1
	} else {
		t = List{t}
		//	if tok == '.' {
		//		tok = s.Scan()
		//	}
		//	if tok != sc.EOF {
		//		err(s, fmt.Sprintf("',' or EOF exspected, not '%c' = Code %d, %X, %u", tok, tok, tok, tok))
	}
	return
}

func parseBIConstraint(s *sc.Scanner) (t Term, tok rune, ok bool) {

	pTraceHeadln(4, 4, "--> readBIConstraint : ")

	tok = s.Scan()
	if tok == sc.EOF {
		return List{}, tok, true
	}

	t, tok, ok = Assignexpr(s, tok)

	pTraceHeadln(4, 4, "<-- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)

	if t.Type() != CompoundType || t.(Compound).Prio == 0 {
		Err(s, fmt.Sprintf(" Not a Built-in constraint: %s ", t))
	}

	if tok == sc.EOF || !ok {
		return
	}
	if tok == ',' {
		t1 := List{t}
		for tok == ',' {
			t, tok, ok = Assignexpr(s, s.Scan())

			pTraceHeadln(4, 4, "<-- expression: term: %s tok: '%s' ok: %v ", t.String(), Tok2str(tok), ok)

			if t.Type() != CompoundType || t.(Compound).Prio == 0 {
				Err(s, fmt.Sprintf(" Not a Built-in constraint: %s ", t))
			}
			if !ok {
				return t1, tok, false
			}
			t1 = append(t1, t)
		}
		t = t1
	} else {
		t = List{t}
	}
	//	if tok == '.' {
	//		tok = s.Scan()
	//	}
	//	if tok != sc.EOF {
	//		err(s, fmt.Sprintf("',' or EOF exspected, not '%c' = Code %d, %X, %u", tok, tok, tok, tok))
	//	}
	return
}
