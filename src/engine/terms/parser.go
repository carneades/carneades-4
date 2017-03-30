// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Terms Parser

package terms

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	sc "text/scanner"
	// "go/scanner"
	// "go/token"
)

// Precedence  Operator
//     7 (coded as 0)  Variable, Function, Constant
//     6         unary operators +, -, !, ^, ¬ and in Go: *, &, <-
//     5         *, /, %, div, mod, &, &^, <<, >>
//     4        +, -, ^, or (the | will be used as list-operator, as in [a|B])
//     3        =, ==, !=, <, <=, >, >= and =< (only for Prolog-like)
//     2        &&
//     1        ||

const trace = false

func ReadString(src string) (result Term, ok bool) {
	// src is the input that we want to tokenize.
	var s sc.Scanner
	// var s *sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))
	s.Error = Err

	result, _, ok = readBIConstraint(&s)
	return
}

func Err(s *sc.Scanner, str string) {
	if str != "illegal char literal" {
		fmt.Fprintln(os.Stderr, "*** Parse Error before[", s.Pos(), "]:", str)
	}
}

func readBIConstraint(s *sc.Scanner) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> readBIConstraint : \n")
	}
	t, tok, ok = expression(s, s.Scan())
	if trace {
		fmt.Printf("<-- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	if tok == sc.EOF || !ok {
		return
	}

	if tok == ',' {
		t1 := List{t}
		for tok == ',' {
			t, tok, ok = expression(s, s.Scan())
			if trace {
				fmt.Printf("<-- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
			}
			if !ok {
				return t1, tok, false
			}
			t1 = append(t1, t)
		}
		t = t1
	}

	if tok != sc.EOF {
		// err(s, fmt.Sprintf("',' or EOF exspected, not '%c' = Code %d, %X, %u", tok, tok, tok, tok))
		return t, tok, false
	}
	return
}

// <expression> | <variable> ':=' <exspression>
func Assignexpr(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {

	if trace {
		fmt.Printf("--> assign expression: '%s'\n", Tok2str(tok1))
	}
	t, tok, ok = expression(s, tok1)
	if trace {
		fmt.Printf("<-- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	if !ok {
		return
	}
	for {
		op := ""
		// named op "is"
		if tok <= 0 {
			if tok == sc.Ident && s.TokenText() == "is" {
				op = "is"
			} else {
				return
			}
		}
		tok2 := s.Peek()
		if tok == ':' && tok2 == '=' {
			op = ":="
			tok = s.Scan()
		}
		if op == "" {
			return
		}
		if t.Type() != VariableType {
			Err(s, fmt.Sprintf(" A Variable, not %s, exspected on the left site of %s", t, op))
			return t, tok, false
		}
		t1 := t
		t, tok, ok = expression(s, s.Scan())
		if trace {
			fmt.Printf("<-- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
		if !ok {
			return t1, tok, ok
		}
		t = Compound{Functor: op, Args: []Term{t1, t}, Prio: 1}
		if trace {
			fmt.Printf("-<- assign-expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	}
}

// <and_expr> | <and_expr> '||' <and_expr>
func expression(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> expression: '%s'\n", Tok2str(tok1))
	}
	t, tok, ok = and_expr(s, tok1)
	if trace {
		fmt.Printf("<-- and_expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	if !ok {
		return
	}
	for {
		op := ""
		/* named op "or" , "or" is used as log. or for go "|" operator, because [ a| B]
		if tok <= 0 {
			if tok == sc.Ident && s.TokenText() == "or" {
				op = "or"
			} else {
				return
			}
		} */
		tok2 := s.Peek()
		if tok == '|' && tok2 == '|' {
			op = "||"
			tok = s.Scan()
		}
		if op == "" {
			return
		}
		t1 := t
		t, tok, ok = and_expr(s, s.Scan())
		if trace {
			fmt.Printf("<-- and_expr: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
		if !ok {
			return t1, tok, ok
		}
		t = Compound{Functor: op, Args: []Term{t1, t}, Prio: 1}
		if trace {
			fmt.Printf("-<- expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
	}
}

// <comp_expr> | <comp_expr> '&&' <comp_expr>
func and_expr(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> and_exp: '%s'\n", Tok2str(tok1))
	}
	t, tok, ok = comp_expr(s, tok1)
	if trace {
		fmt.Printf("<-- comp_expr: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	if !ok {
		return
	}
	for {
		op := ""
		/* named op "and"
		if tok <= 0 {
			if tok == sc.Ident && s.TokenText() == "and" {
				op = "and"
			} else {
				return
			}
		}
		*/
		tok2 := s.Peek()
		if tok == '&' && tok2 == '&' {
			op = "&&"
			tok = s.Scan()
		}
		if op == "" {
			return
		}
		t1 := t
		t, tok, ok = comp_expr(s, s.Scan())
		if trace {
			fmt.Printf("<-- comp_expr: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
		if !ok {
			return t1, tok, ok
		}
		t = Compound{Functor: op, Args: []Term{t1, t}, Prio: 2}
		if trace {
			fmt.Printf("-<- and_exp: term: %s tok: '%s' ok: %v\n", t.String(), Tok2str(tok), ok)
		}
	}
}

// <simple_expression> | <simple_expression> ['in','==','<=','>=','!=','=<','<','>'] <simple_expression>
func comp_expr(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> comp_expr: '%s'\n", Tok2str(tok1))
	}
	t, tok, ok = simple_expression(s, tok1)
	if trace {
		fmt.Printf("<-- simple_expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	op := ""
	// named operator
	if tok <= 0 {
		if tok == sc.Ident && s.TokenText() == "in" {
			op = "in"
		} else {
			return
		}
	} else {
		// sign operator
		tok2 := s.Peek()
		switch tok {
		case '=':
			if tok2 == '=' {
				op = "=="
				tok = s.Scan()
			} else {
				op = "="
			}
			// only for PROLOG
			if tok2 == '<' {
				op = "=<"
				tok = s.Scan()
			}
		case '<':
			if tok2 == '=' {
				op = "<="
				tok = s.Scan()
			} else {
				op = "<"
			}
		case '!':
			if tok2 == '=' {
				op = "!="
				tok = s.Scan()
			}
		case '>':
			if tok2 == '=' {
				op = ">="
				tok = s.Scan()
			} else {
				op = ">"
			}

		}
	}
	if op == "" {
		return
	}
	// compare expression with op
	t1 := t
	t, tok, ok = simple_expression(s, s.Scan())
	if trace {
		fmt.Printf("<-- simple_expression: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	if !ok {
		return t1, tok, ok
	}

	return Compound{Functor: op, Args: []Term{t1, t}, Prio: 3}, tok, ok
}

// <sterm> | <sterm> ['or','-','+','^'] <sterm>
func simple_expression(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> simple_expression : '%s'\n", Tok2str(tok1))
	}

	t, tok, ok = sterm(s, tok1)
	if trace {
		fmt.Printf("<-- sterm: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	for {
		op := ""
		if tok <= 0 {
			if tok == sc.Ident && s.TokenText() == "or" {
				op = "or"
			} else {
				return
			}
		}
		// tok2 := s.Peek()
		switch tok {
		case '-':
			op = "-"
		case '+':
			op = "+"
		/* in Go log. or, in Prolog: [a|B]
		case '|':
			if tok2 == '|' {
				return
			}
			op = "|"
		*/
		case '^':
			op = "^"
		}
		if op == "" {
			return
		}

		t1 := t
		t, tok, ok = sterm(s, s.Scan())
		if trace {
			fmt.Printf("<-- rec. sterm: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
		if !ok {
			return
		}
		t = Compound{Functor: op, Args: []Term{t1, t}, Prio: 4}
		if trace {
			fmt.Printf("-<- simple_expression: term: %s tok: '%s' ok: %v\n", t.String(), Tok2str(tok), ok)
		}
	}
}

// <unary_factor> | <unary_factor> ['div','mod','*','/','%','&','&^','<<','>>'] <unary_factor>
func sterm(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> sterm : '%s'\n", Tok2str(tok1))
	}
	t, tok, ok = unary_factor(s, tok1)
	if trace {
		fmt.Printf("<-- unary_factor: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
	}
	for {
		op := ""
		// named operator
		if tok <= 0 {
			if tok == sc.Ident {
				switch s.TokenText() {
				case "div":
					op = "div"
				case "mod":
					op = "mod"
				default:
					return
				}
			} else {
				return
			}
		} else {
			// sign operator
			tok2 := s.Peek()
			switch tok {
			case '*':
				op = "*"
			case '/':
				op = "/"
			case '%':
				op = "%"
			case '&':
				if tok2 == '&' {
					return
				}
				if tok2 == '^' {
					op = "&^"
					tok = s.Scan()
				} else {
					op = "&"
				}
			case '<':
				if tok2 == '<' {
					op = "<<"
					tok = s.Scan()
				}
			case '>':
				if tok2 == '>' {
					op = ">>"
					tok = s.Scan()
				}
			}
		}
		if op == "" {
			return
		}
		// factor with op
		t1 := t
		t, tok, ok = unary_factor(s, s.Scan())
		if trace {
			fmt.Printf("<-- unary_factor: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
		}
		if !ok {
			return t1, tok, ok
		}
		t = Compound{Functor: op, Args: []Term{t1, t}, Prio: 5}
		if trace {
			fmt.Printf("<- sterm: term: %s tok: '%s' ok: %v\n", t.String(), Tok2str(tok), ok)
		}
	}
}

// ['+','-','!','^','¬'] <unary_factor> | <factor>
func unary_factor(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> unary_factor : '%s'\n", Tok2str(tok1))
	}
	// op +, -, !, ^ and in GO: *, &, <-

	unaryop := ""
	tok2 := s.Peek()
	switch tok1 {
	case '+':
		return unary_factor(s, s.Scan())
	case '-':
		if tok2 == '-' {
			s.Scan()
			return unary_factor(s, s.Scan())
		} else {
			unaryop = "-"
		}
	case '!':
		if tok2 == '!' {
			s.Scan()
			return unary_factor(s, s.Scan())
		} else {
			unaryop = "!"
		}
	case '^':
		unaryop = "^"
	case '¬':
		unaryop = "¬"
	}
	if unaryop == "" {
		return factor(s, tok1)
	}

	t, tok, ok = unary_factor(s, s.Scan())
	if trace {
		fmt.Printf("--> unary_factor : '%s'\n", Tok2str(tok1))
	}
	if !ok {
		return
	}
	return Compound{Functor: unaryop, Args: []Term{t}, Prio: 6}, tok, ok
}

// '[' ']' | '[' <expression> [',' <expression>]0..n ['|' <variable>]0..1 ']' |
// '(' <expression> ')' | <factor-name> | <int> | <float> | <char> | <string> | <raw-string>
func factor(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> factor : '%s'\n", Tok2str(tok1))
	}
	tok = tok1
	ok = true
	switch tok1 {
	case '[': //
		list := List{}
		if s.Peek() == ']' {
			s.Scan()
			return list, s.Scan(), true
		}
		tok = ','
		pos := s.Pos()
		for tok == ',' {
			t, tok, ok = expression(s, s.Scan())
			if trace {
				fmt.Printf("<-- expression in [ factor: term: %s tok: '%s' ok: %v \n", t.String(), Tok2str(tok), ok)
			}
			if !ok {
				t = list
				return
			}
			list = append(list, t)
		}
		t = list
		// [ expr | variable ]
		if tok == '|' {
			tok = s.Scan()
			if tok == sc.Ident {
				n := s.TokenText()
				c := n[0]
				ok = true
				if c < 'A' || c > 'Z' {
					Err(s, fmt.Sprintf("expected variable in [-list after '|' not '%s'", n))
					ok = false
				}
				v := Variable{Name: n, index: big.NewInt(0)}
				tok = s.Scan()
				t = Compound{Functor: "|", Args: List{v}, Prio: 6}
				t = append(list, t)
				if tok == ']' {
					return t, s.Scan(), ok
				} else {
					Err(s, fmt.Sprintf("missing closed ']' after '[ ... | %s", n))
					return t, tok, false
				}
			} else {
				Err(s, fmt.Sprintf("expected variable in [-list after '|' not '%s'", Tok2str(tok)))
				return t, tok, false

			}
		}
		if tok != ']' {
			Err(s, fmt.Sprintf("missing closed ']' for the open '[' at position %s", pos))
			return t, tok, false
		}
		return t, s.Scan(), true
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
			Err(s, fmt.Sprintf("missing closed ')' for the open '(' at position %s", pos))
			return t, tok, false
		}
		tok = s.Scan()
	case sc.Ident:
		t, tok, ok = Factor_name(s.TokenText(), s, s.Scan())
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
		t, tok, ok = Factor_name(s.TokenText(), s, s.Scan())
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
		Err(s, "EOF missing term")
		return Atom("nil"), tok, false
	default:
		Err(s, fmt.Sprintf("unexpected character '%c', expect <variable>, <name>, <constant>, '(' or '['", tok))
		t = Atom("nil")
		return t, tok, false
	}
	return t, tok, true
}

// <bi_0 name> | <name>'('')' | <name> '(' <expression> [',' <expression>]0..n ')'
func Factor_name(name string, s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> factor_name : %s, '%s'\n", name, Tok2str(tok1))
	}
	t = Atom("nil")
	ok = true
	tok = tok1
	if tok != '(' {
		return bi_0(name, tok)
	}
	// to do: distinguish between CHR- and Built In-Constraint
	args := []Term{}
	if s.Peek() == ')' {
		s.Scan()

		return Compound{Functor: name, Args: []Term{}}, s.Scan(), true
	}
	tok = ','
	pos := s.Pos()
	for tok == ',' {
		t, tok, ok = expression(s, s.Scan())
		if !ok {
			t = Compound{Functor: name, Args: args}
			return
		}
		args = append(args, t)
	}
	t = Compound{Functor: name, Args: args}

	if tok != ')' {
		Err(s, fmt.Sprintf("missing closed ')' for the open '(' at position %s", pos))
		return t, tok, false
	}
	return t, s.Scan(), true
}

func sInt(s *sc.Scanner) (Term, rune, bool) {
	if trace {
		fmt.Printf("--> sInt : '%s'\n", s.TokenText())
	}
	var (
		i   int
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &i)
	if err == nil {
		return Int(i), s.Scan(), true
	}
	return Int(i), s.Scan(), false
}

func sFloat(s *sc.Scanner) (Term, rune, bool) {
	if trace {
		fmt.Printf("--> sFloat : '%s'\n", s.TokenText())
	}
	var (
		f   float64
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &f)
	if err == nil {
		return Float(f), s.Scan(), true
	}
	return Float(f), s.Scan(), false
}

func sString(s *sc.Scanner) (Term, rune, bool) {
	if trace {
		fmt.Printf("--> sString : '%s'\n", s.TokenText())
	}
	var (
		str string
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &str)
	if err == nil {
		return String(str), s.Scan(), true
	}
	return String(str), s.Scan(), false
}

func sChar(s *sc.Scanner) (Term, rune, bool) {
	if trace {
		fmt.Printf("--> sChar : %s\n", s.TokenText())
	}

	return Atom(fmt.Sprintf("%s", s.TokenText())), s.Scan(), true
}

func Tok2str(tok rune) string {
	if tok > 0 {
		return string(tok)
	}

	switch tok {
	case sc.Ident:
		return "Ident"
	case sc.Int:
		return "Int"
	case sc.Float:
		return "Float"
	case sc.Char:
		return "Char"
	case sc.String:
		return "String"
	case sc.RawString:
		return "RawString"
	case sc.Comment:
		return "Comment"
	case sc.EOF:
		return "EOF"
	}
	return "??"
}

func bi_0(n string, tk rune) (t Term, tok rune, ok bool) {
	if trace {
		fmt.Printf("--> bi_0 : '%s'\n", n)
	}
	tok = tk
	ok = true
	switch n {
	case "true":
		t = Bool(true)
		return
	case "false":
		t = Bool(false)
		return
	}

	c := n[0]
	if c >= 'A' && c <= 'Z' {
		t = Variable{Name: n, index: big.NewInt(0)}
		return
	} else {
		t = Atom(n)
		return
	}
}
